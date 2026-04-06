# bt-cli

*CLI/TUI wrapper around bluetoothctl commands*
* this is vibe coded *

Control your bluetooth devices via the command line. `bt`
![screenshot](https://github.com/user-attachments/assets/3c6414d2-0f6a-442c-89f1-b62ae98ca3e4)

---

## Building

### Prerequisites
- Go 1.21+
- Linux with `bluetoothctl` available (part of the `bluez` package)

### Build

```bash
go build -o bt .
```

This produces an executable named `bt` in the current directory.

### Run

```bash
./bt
```

### Install globally

```bash
go install
# or
go build -o /usr/local/bin/bt .
```

### Dependencies

Dependencies are managed via Go modules and installed automatically with `go build` or `go mod download`.

Required packages:
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — Terminal styling


## Code Structure

The app follows a **Model-View-Update (MVU)** architecture using the [Charmbracelet Bubbletea](https://github.com/charmbracelet/bubbletea) framework.

```
bt-cli/
├── main.go              # Entry point
├── bluetooth/            # Bluetooth abstraction layer
│   └── bluetooth.go     # Wraps bluetoothctl commands
└── ui/                  # Terminal UI layer
    ├── app.go           # TUI logic (Model + Update + View)
    └── styles.go        # Lipgloss styling definitions
```

## TROUBLESHOOTING bluetooth controls

```

#### bluetooth controls
```
bluetoothctl power on
bluetoothctl power off

bluetoothctl scan on
bluetoothctl scan off

bluetoothctl pair <MAC_ADDRESS>

bluetoothctl connect <MAC_ADDRESS>
bluetoothctl disconnect <MAC_ADDRESS>

bluetoothctl devices
```
To create a connection with the built-in utils, you can follow this slightly more manual process using bluetoothctl.

hcitool scan  # to get the MAC address of your device
bluetoothctl
power on  # in case the bluez controller power is off 
agent on
scan on  # wait for your device's address to show up here
scan off
trust MAC_ADDRESS
pair MAC_ADDRRESS
connect MAC_ADDRESS

sudo hcitool cc 94:23:6E:6F:23:9D

then quickly pair and connect with bluetoothctl
