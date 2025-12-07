# jellyfin-mustui

A terminal-based user interface (TUI) for Jellyfin, built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss). Browse your music library, play tracks, and enjoy a sleek terminal experience.

<p align="center">
  <img width="926" height="489" alt="image" src="https://github.com/user-attachments/assets/65c3efde-57b1-434c-85b0-2babce60468b" />
</p>

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/cedev-1/jellyfin-mustui.git
   cd jellyfin-mustui
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   go build -o jellyfin-mustui ./cmd/jellyfin-mustui
   ```

   Or run directly:
   ```bash
   go run ./cmd/jellyfin-mustui/main.go
   ```

## Usage

1. Run the application:
   ```bash
   ./jellyfin-mustui
   ```

2. On first launch, enter your Jellyfin server details (URL, username, password).

3. Browse your music:
   - Use arrow keys or Vim keys (`h`, `j`, `k`, `l`) to navigate.
   - Press `Tab` to switch between Artists and Tracks panels.
   - Press `Enter` to select an artist/album or play a track.
   - Press `/` to filter/search in lists.
   - Press `Space` to play/pause, `n`/`p` for next/previous track.

4. Press `?` for help, `q` to quit.

## Keybindings

- **Navigation**: `↑/↓` or `k/j` (up/down), `Tab` (switch panels)
- **Selection**: `Enter` (select/play)
- **Albums**: `h/l` or `←/→` (previous/next album in Tracks panel)
- **Playback**: `Space` (play/pause), `n` (next), `p` (previous)
- **Search**: `/` (filter in lists)
- **Help**: `?` (toggle help), `Esc` (close help or cancel)
- **Quit**: `q` (or `Ctrl+C`)

## Configuration

The app saves your login details in a config file (e.g., `~/.config/jellyfin-mustui/config.json` on Linux).

## License

MIT [LICENSE](LICENSE).
