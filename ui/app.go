package ui

import (
	"fmt"
	"os"

	"bt/bluetooth"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	devices    []bluetooth.Device
	controller *bluetooth.Controller
	selected   int
	scanning   bool
	loading    bool
	statusMsg  string
	err        error
}

func InitialModel() model {
	return model{
		selected: 0,
		loading:  true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		refreshDevices,
		getController,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.devices)-1 {
				m.selected++
			}
		case "enter":
			if len(m.devices) > 0 && m.selected < len(m.devices) {
				return m, toggleDevice(m.devices[m.selected])
			}
		case "c":
			if len(m.devices) > 0 && m.selected < len(m.devices) {
				return m, connectDevice(m.devices[m.selected].MAC)
			}
		case "d":
			if len(m.devices) > 0 && m.selected < len(m.devices) {
				return m, disconnectDevice(m.devices[m.selected].MAC)
			}
		case "t":
			if len(m.devices) > 0 && m.selected < len(m.devices) {
				return m, toggleTrust(m.devices[m.selected])
			}
		case "r":
			m.loading = true
			return m, refreshDevices
		case "s":
			m.scanning = !m.scanning
			return m, toggleScan(m.scanning)
		case "p":
			if len(m.devices) > 0 && m.selected < len(m.devices) {
				return m, pairDevice(m.devices[m.selected].MAC)
			}
		}
	case []bluetooth.Device:
		m.devices = msg
		m.loading = false
	case *bluetooth.Controller:
		m.controller = msg
	case string:
		m.statusMsg = msg
	case error:
		m.err = msg
	}
	return m, nil
}

func (m model) View() string {
	s := BorderStyle.Render

	header := TitleStyle.Render(" Bluetooth ")
	if m.controller != nil {
		power := "●"
		powerColor := lipgloss.Color("46")
		if !m.controller.Powered {
			power = "○"
			powerColor = lipgloss.Color("241")
		}
		header = fmt.Sprintf("%s %s %s", header, lipgloss.Style{}.
			Foreground(powerColor).Render(power), lipgloss.Style{}.
			Foreground(lipgloss.Color("250")).Render(m.controller.Name))
	}
	if m.scanning {
		header = fmt.Sprintf("%s %s", header, lipgloss.Style{}.
			Foreground(lipgloss.Color("46")).Render("◐ Scanning"))
	}

	output := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
	)

	if m.loading {
		output = lipgloss.JoinVertical(lipgloss.Left, output, "  Loading...")
	} else if len(m.devices) == 0 {
		output = lipgloss.JoinVertical(lipgloss.Left, output, "  No paired devices")
		output = lipgloss.JoinVertical(lipgloss.Left, output, "  Press 's' to scan for nearby devices")
	} else {
		deviceHeader := HeaderStyle.Render(" Paired Devices ")
		output = lipgloss.JoinVertical(lipgloss.Left, output, deviceHeader)

		for i, device := range m.devices {
			var status string
			if device.Connected {
				status = "Connected"
			} else {
				status = "Disconnected"
			}

			trustMark := " "
			if device.Trusted {
				trustMark = "★"
			}

			name := bluetooth.PadRight(device.DisplayName(), 25)
			deviceLine := fmt.Sprintf("  %s %s  %s %s",
				device.TypeIcon(),
				name,
				status,
				trustMark,
			)

			if i == m.selected {
				deviceLine = SelectedStyle.Render(deviceLine)
			} else {
				deviceLine = NormalStyle.Render(deviceLine)
			}

			output = lipgloss.JoinVertical(lipgloss.Left, output, deviceLine)
		}
	}

	output = lipgloss.JoinVertical(lipgloss.Left, output, "")
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Render("  ↑↓ Navigate  |  c: Connect  |  d: Disconnect  |  t: Trust  |  s: Scan  |  r: Refresh  |  q: Quit"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, lipgloss.Style{}.Foreground(lipgloss.Color("245")).Render("  ★ = Trusted (auto-reconnect)"))

	if m.statusMsg != "" {
		if m.err != nil {
			output = lipgloss.JoinVertical(lipgloss.Left, output, ErrorStyle.Render("  ✗ "+m.statusMsg))
		} else {
			output = lipgloss.JoinVertical(lipgloss.Left, output, SuccessStyle.Render("  ✓ "+m.statusMsg))
		}
	}

	return s(output)
}

func toggleDevice(device bluetooth.Device) tea.Cmd {
	return func() tea.Msg {
		var err error
		if device.Connected {
			err = bluetooth.DisconnectDevice(device.MAC)
		} else {
			err = bluetooth.ConnectDevice(device.MAC)
		}
		if err != nil {
			return err
		}
		devices, err := bluetooth.GetDevices()
		if err != nil {
			return err
		}
		return devices
	}
}

func connectDevice(mac string) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.ConnectDevice(mac)
		if err != nil {
			return err
		}
		devices, err := bluetooth.GetDevices()
		if err != nil {
			return err
		}
		return devices
	}
}

func disconnectDevice(mac string) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.DisconnectDevice(mac)
		if err != nil {
			return err
		}
		devices, err := bluetooth.GetDevices()
		if err != nil {
			return err
		}
		return devices
	}
}

func toggleTrust(device bluetooth.Device) tea.Cmd {
	return func() tea.Msg {
		var err error
		if device.Trusted {
			err = bluetooth.UntrustDevice(device.MAC)
		} else {
			err = bluetooth.TrustDevice(device.MAC)
		}
		if err != nil {
			return err
		}
		return "Trust updated"
	}
}

func toggleScan(on bool) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.SetScan(on)
		if err != nil {
			return err
		}
		return "Scan updated"
	}
}

func refreshDevices() tea.Msg {
	devices, err := bluetooth.GetDevices()
	if err != nil {
		return err
	}
	return devices
}

func getController() tea.Msg {
	ctrl, err := bluetooth.GetController()
	if err != nil {
		return err
	}
	return ctrl
}

func pairDevice(mac string) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.PairDevice(mac)
		if err != nil {
			return err
		}
		return "Pairing initiated"
	}
}

func Run() {
	p := tea.NewProgram(InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
