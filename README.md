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

### Layers

**1. `main.go`** — Trivial entry point. Calls `ui.Run()`.

**2. `bluetooth/bluetooth.go`** — Low-level wrapper around the `bluetoothctl` CLI tool.
- Defines `Device` and `Controller` structs
- Functions like `GetDevices()`, `ConnectDevice()`, `SetPower()` all exec `bluetoothctl` and parse the text output
- All I/O is blocking and synchronous (executes a subprocess)

**3. `ui/app.go`** — The TUI. Contains three components:

| Component | Role |
|---|---|
| `model` struct | Holds all app state (devices, controller, selection, loading flags, etc.) |
| `Update()` | Event handler — receives user input or async results, mutates state, returns commands |
| `View()` | Renders the terminal UI by returning a string (called after every `Update`) |

**4. `ui/styles.go`** — Defines reusable Lipgloss style objects (colors, borders, bold text) used by `View()`.

### The MVU Flow

1. **Init** — Fetches devices and controller info
2. **Loop** — For each user keypress or async result:
   - `Update()` processes the message, mutates `model`, returns a `tea.Cmd` (background task)
   - `View()` renders the current model to a string
3. **Commands** — Background goroutines (e.g., `connectDevice()`) run bluetoothctl, then send the result back as a new message to `Update()`

### Key Design Choices

- **No separate "View" struct** — The `View()` method is on the model itself (Bubbletea convention)
- **Messages as state change protocol** — All state mutations happen in `Update()` based on incoming messages
- **Synchronous bluetoothctl calls** — Each operation (connect, disconnect, scan) blocks until the subprocess completes

