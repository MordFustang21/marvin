# Marvin - A Spotlight Alternative

Marvin is a lightweight Spotlight alternative built with Go and Fyne.io. It provides quick search functionality for files, folders, and applications on your system using macOS Spotlight index.

<img src="docs/screenshots/marvin-screenshot.png" alt="Marvin Screenshot" width="600"/>

## Features

- Fast searching using macOS Spotlight index
- Global hotkey activation (Cmd+Space)
- File/application launching
- Lightweight and responsive

## Requirements

- Go 1.18 or higher
- macOS (due to dependency on macOS Spotlight)
- Fyne.io dependencies (automatically installed via Go modules)

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/marvin-go.git

# Build and install
cd marvin-go
go build -o marvin ./cmd/marvin
```

## Usage

Run the application:

```bash
./marvin
```

Press `Cmd+Space` (or `Alt+Space`) to activate the search interface. Type your search query to find files, folders, and applications.

## Configuration

Currently, Marvin has minimal configuration options. Future versions will include:

- Customizable themes
- Configurable keyboard shortcuts
- Plugin system for extended functionality

## Architecture

Marvin is built using the following components:

- **UI Layer**: Uses Fyne.io for cross-platform GUI
- **Search Layer**: Interfaces with macOS Spotlight index
- **Hotkey System**: Global keyboard shortcut handling

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
