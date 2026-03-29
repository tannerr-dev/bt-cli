# bt-cli

*CLI/TUI wrapper around bluetoothctl commands*

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
- `github.com/mattn/go-runewidth` — Unicode width for padding


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

