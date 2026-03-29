package bluetooth

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

type DeviceType int

const (
	TypeUnknown DeviceType = iota
	TypeAudio
	TypeInput
	TypePhone
	TypeComputer
	TypeVideo
	TypePeripheral
)

type Device struct {
	MAC        string
	Name       string
	Alias      string
	Paired     bool
	Trusted    bool
	Connected  bool
	Blocked    bool
	Icon       string
	DeviceType DeviceType
}

func (d Device) DisplayName() string {
	if d.Alias != "" {
		return d.Alias
	}
	if d.Name != "" {
		return d.Name
	}
	return d.MAC
}

func (d Device) TypeIcon() string {
	switch d.DeviceType {
	case TypeAudio:
		return "🎧"
	case TypeInput:
		return "⌨️"
	case TypePhone:
		return "📱"
	case TypeComputer:
		return "💻"
	case TypeVideo:
		return "📺"
	case TypePeripheral:
		return "🖱️"
	default:
		return "📟"
	}
}

func parseDeviceType(icon string) DeviceType {
	switch {
	case strings.HasPrefix(icon, "audio"):
		return TypeAudio
	case strings.HasPrefix(icon, "input"):
		return TypeInput
	case strings.HasPrefix(icon, "phone"):
		return TypePhone
	case strings.HasPrefix(icon, "computer"):
		return TypeComputer
	case strings.HasPrefix(icon, "video"):
		return TypeVideo
	case strings.HasPrefix(icon, "peripheral"):
		return TypePeripheral
	default:
		return TypeUnknown
	}
}

type Controller struct {
	MAC          string
	Name         string
	Powered      bool
	Discoverable bool
	Pairable     bool
}

func RunCommand(cmd string, args ...string) (string, error) {
	fullArgs := append([]string{cmd}, args...)

	execCmd := exec.Command("bluetoothctl", fullArgs...)
	output, err := execCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := stripANSI(string(exitErr.Stderr))
			if stderr != "" {
				return "", errors.New(stderr)
			}
			return "", err
		}
		return "", err
	}
	return stripANSI(string(output)), nil
}

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func GetDevices() ([]Device, error) {
	output, err := RunCommand("devices", "Paired")
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	var macs []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "Device ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		mac := parts[1]
		if !isMAC(mac) {
			continue
		}
		macs = append(macs, mac)
	}

	if len(macs) == 0 {
		return []Device{}, nil
	}

	type result struct {
		mac    string
		name   string
		device Device
		err    error
	}

	const maxConcurrency = 5
	sem := make(chan struct{}, maxConcurrency)
	results := make(chan result, len(macs))
	var wg sync.WaitGroup

	for _, mac := range macs {
		wg.Add(1)
		go func(mac string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			name := ""
			output, err := RunCommand("info", mac)
			if err == nil {
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "Device") && strings.Contains(line, mac) {
						continue
					}
					if strings.HasPrefix(line, "Name") {
						parts := strings.SplitN(line, ":", 2)
						if len(parts) == 2 {
							name = strings.TrimSpace(parts[1])
						}
						break
					}
				}
			}

			device, err := GetDeviceInfo(mac)
			if err != nil {
				results <- result{mac: mac, name: name, err: err}
				return
			}
			if device.Name == "" {
				device.Name = name
			}
			results <- result{device: device}
		}(mac)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var devices []Device
	for r := range results {
		if r.err != nil && r.device.MAC == "" {
			devices = append(devices, Device{MAC: r.mac, Name: r.name})
		} else {
			devices = append(devices, r.device)
		}
	}

	return devices, nil
}

func isMAC(s string) bool {
	if len(s) != 17 {
		return false
	}
	for i, c := range s {
		if i == 2 || i == 5 || i == 8 || i == 11 || i == 14 {
			if c != ':' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
				return false
			}
		}
	}
	return true
}

func GetDeviceInfo(mac string) (Device, error) {
	output, err := RunCommand("info", mac)
	if err != nil {
		return Device{}, err
	}

	device := Device{MAC: mac}
	lines := strings.Split(output, "\n")

	var inDeviceSection bool
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Device") && strings.Contains(line, mac) {
			inDeviceSection = true
			continue
		}

		if inDeviceSection {
			if line == "" || strings.HasPrefix(line, "[") {
				continue
			}

			if strings.HasPrefix(line, "Name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					device.Name = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "Alias") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					device.Alias = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "Paired") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					device.Paired = strings.TrimSpace(parts[1]) == "yes"
				}
			} else if strings.HasPrefix(line, "Trusted") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					device.Trusted = strings.TrimSpace(parts[1]) == "yes"
				}
			} else if strings.HasPrefix(line, "Connected") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					device.Connected = strings.TrimSpace(parts[1]) == "yes"
				}
			} else if strings.HasPrefix(line, "Blocked") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					device.Blocked = strings.TrimSpace(parts[1]) == "yes"
				}
			} else if strings.HasPrefix(line, "Icon") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					device.Icon = strings.TrimSpace(parts[1])
					device.DeviceType = parseDeviceType(parts[1])
				}
			}
		}
	}

	return device, nil
}

func GetController() (*Controller, error) {
	output, err := RunCommand("show")
	if err != nil {
		return nil, err
	}

	ctrl := &Controller{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Controller") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ctrl.MAC = parts[1]
			}
		} else if strings.HasPrefix(line, "Name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ctrl.Name = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Powered") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ctrl.Powered = strings.TrimSpace(parts[1]) == "yes"
			}
		} else if strings.HasPrefix(line, "Discoverable") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ctrl.Discoverable = strings.TrimSpace(parts[1]) == "yes"
			}
		} else if strings.HasPrefix(line, "Pairable") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ctrl.Pairable = strings.TrimSpace(parts[1]) == "yes"
			}
		}
	}

	return ctrl, nil
}

func ConnectDevice(mac string) error {
	_, err := RunCommand("connect", mac)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	return nil
}

func DisconnectDevice(mac string) error {
	_, err := RunCommand("disconnect", mac)
	if err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}
	return nil
}

func TrustDevice(mac string) error {
	_, err := RunCommand("trust", mac)
	return err
}

func UntrustDevice(mac string) error {
	_, err := RunCommand("untrust", mac)
	return err
}

func PairDevice(mac string) error {
	_, err := RunCommand("pair", mac)
	return err
}

func RemoveDevice(mac string) error {
	_, err := RunCommand("remove", mac)
	return err
}

func SetPower(on bool) error {
	power := "off"
	if on {
		power = "on"
	}
	_, err := RunCommand("power", power)
	return err
}

func SetScan(on bool) error {
	scan := "off"
	if on {
		scan = "on"
	}
	_, err := RunCommand("scan", scan)
	return err
}
