package main

import (
	"fmt"
	"os"

	"github.com/cedev-1/jellyfin-mustui/internal/config"
	"github.com/cedev-1/jellyfin-mustui/internal/jellyfin"
	"github.com/cedev-1/jellyfin-mustui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	client := jellyfin.NewClient(cfg.ServerURL, cfg.Token, cfg.UserID)

	m := tui.NewModel(cfg, client)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
