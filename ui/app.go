package ui

import (
	"fmt"
	"os"
	"time"

	"bt/bluetooth"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	width        int
	height       int
	devices      []bluetooth.Device
	controller   *bluetooth.Controller
	selected     int
	scanning     bool
	loading      bool
	scanInFlight bool
	statusMsg    string
	err          error
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
		tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return refreshControllerMsg{}
		}),
	)
}

type refreshControllerMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
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
			return m, tea.Batch(refreshDevices, getController)
		case "s":
			if m.scanInFlight {
				return m, nil
			}
			m.scanInFlight = true
			m.scanning = !m.scanning
			return m, toggleScan(m.scanning)
		case "p":
			if len(m.devices) > 0 && m.selected < len(m.devices) {
				return m, pairDevice(m.devices[m.selected].MAC)
			}
		case "P":
			return m, togglePower()
		case "x":
			if len(m.devices) > 0 && m.selected < len(m.devices) {
				return m, removeDevice(m.devices[m.selected].MAC)
			}
		}
	case []bluetooth.Device:
		m.devices = msg
		m.loading = false
		m.err = nil
	case *bluetooth.Controller:
		m.controller = msg
	case devicesAndControllerMsg:
		devices, ctrl, err := getDevicesAndController()
		if err != nil {
			m.err = err
			m.loading = false
		} else {
			m.devices = devices
			m.controller = ctrl
			m.loading = false
			m.err = nil
		}
	case refreshControllerMsg:
		return m, tea.Batch(
			getController,
			tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
				return refreshControllerMsg{}
			}),
		)
	case scanDoneMsg:
		m.scanInFlight = false
	case string:
		m.statusMsg = msg
	case error:
		m.err = msg
		m.loading = false
	}
	return m, nil
}

func (m model) View() string {
	s := BorderStyle.Render

	minWidth := 50
	if m.width > 0 {
		minWidth = m.width
	}

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

	statusMode := 2
	trustMode := 2

	if m.loading {
		output = lipgloss.JoinVertical(lipgloss.Left, output, "  Loading...")
	} else if len(m.devices) == 0 {
		output = lipgloss.JoinVertical(lipgloss.Left, output, "  No paired devices")
		output = lipgloss.JoinVertical(lipgloss.Left, output, "  Press 's' to scan")
	} else {
		deviceHeader := HeaderStyle.Render(" Paired Devices ")
		output = lipgloss.JoinVertical(lipgloss.Left, output, deviceHeader)

		nameWidth := 25

		if minWidth < 60 {
			nameWidth = 15
			statusMode = 1
			trustMode = 1
		}
		if minWidth < 45 {
			nameWidth = 10
			statusMode = 1
			trustMode = 1
		}
		if minWidth < 35 {
			nameWidth = 8
			statusMode = 0
			trustMode = 0
		}

		for i, device := range m.devices {
			trustMark := " "
			if trustMode > 0 && device.Trusted {
				trustMark = "★"
			}

			displayName := device.DisplayName()
			if lipgloss.Width(displayName) > nameWidth {
				displayName = truncate(displayName, nameWidth)
			}

			var statusStr string
			var statusStyle lipgloss.Style
			connected := device.Connected

			switch statusMode {
			case 0:
				statusStyle = StatusDisconnected
				if connected {
					statusStyle = StatusConnected
					statusStr = "●"
				} else {
					statusStr = "○"
				}
			case 1:
				statusStyle = StatusDisconnected
				if connected {
					statusStyle = StatusConnected
					statusStr = "Conn"
				} else {
					statusStr = "Disc"
				}
			default:
				statusStyle = StatusDisconnected
				if connected {
					statusStyle = StatusConnected
					statusStr = "Connected   "
				} else {
					statusStr = "Disconnected"
				}
			}

			deviceLine := fmt.Sprintf("  %s %-*s %s %s",
				device.TypeIcon(),
				nameWidth,
				displayName,
				statusStyle.Render(statusStr),
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
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("  Navigation"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    ↑↓  Navigate"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render(""))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("  Device Actions"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    Enter  Connect/Disconnect"))
	if minWidth >= 45 {
		output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    c      Connect"))
		output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    d      Disconnect"))
	}
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    t      Toggle Trust"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    x      Remove Device"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render(""))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("  System"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    s      Toggle Scan"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    P      Toggle Power"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    r      Refresh"))
	output = lipgloss.JoinVertical(lipgloss.Left, output, HelpStyle.Width(minWidth-4).Render("    q      Quit"))

	if trustMode > 0 {
		output = lipgloss.JoinVertical(lipgloss.Left, output, "")
		switch trustMode {
		case 1:
			output = lipgloss.JoinVertical(lipgloss.Left, output, lipgloss.Style{}.Foreground(lipgloss.Color("245")).Render("  ★ = Trusted"))
		default:
			output = lipgloss.JoinVertical(lipgloss.Left, output, lipgloss.Style{}.Foreground(lipgloss.Color("245")).Render("  ★ = Trusted (auto-reconnect)"))
		}
	}

	if m.statusMsg != "" {
		if m.err != nil {
			output = lipgloss.JoinVertical(lipgloss.Left, output, ErrorStyle.Width(minWidth-4).Render("  ✗ "+m.statusMsg))
		} else {
			output = lipgloss.JoinVertical(lipgloss.Left, output, SuccessStyle.Width(minWidth-4).Render("  ✓ "+m.statusMsg))
		}
	}

	return s(output)
}

func truncate(s string, maxWidth int) string {
	width := 0
	for i, r := range s {
		w := 2
		if r < 128 {
			w = 1
		}
		if width+w > maxWidth {
			return s[:i] + "…"
		}
		width += w
	}
	return s
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
		return devicesAndControllerMsg{}
	}
}

func connectDevice(mac string) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.ConnectDevice(mac)
		if err != nil {
			return err
		}
		return devicesAndControllerMsg{}
	}
}

func disconnectDevice(mac string) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.DisconnectDevice(mac)
		if err != nil {
			return err
		}
		return devicesAndControllerMsg{}
	}
}

type devicesAndControllerMsg struct{}

func getDevicesAndController() ([]bluetooth.Device, *bluetooth.Controller, error) {
	devices, err := bluetooth.GetDevices()
	if err != nil {
		return nil, nil, err
	}
	ctrl, err := bluetooth.GetController()
	if err != nil {
		return devices, nil, err
	}
	return devices, ctrl, nil
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
		return scanDoneMsg{on: on}
	}
}

type scanDoneMsg struct {
	on bool
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
		devices, err := bluetooth.GetDevices()
		if err != nil {
			return err
		}
		return devices
	}
}

func togglePower() tea.Cmd {
	return func() tea.Msg {
		ctrl, err := bluetooth.GetController()
		if err != nil {
			return err
		}
		err = bluetooth.SetPower(!ctrl.Powered)
		if err != nil {
			return err
		}
		newCtrl, err := bluetooth.GetController()
		if err != nil {
			return err
		}
		return newCtrl
	}
}

func removeDevice(mac string) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.RemoveDevice(mac)
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

func Run() {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
