package tui

import "github.com/charmbracelet/lipgloss"

// Color Palette
var (
	colorPrimary    = lipgloss.Color("#B7BDF8")
	colorSecondary  = lipgloss.Color("#6E738D")
	colorText       = lipgloss.Color("#CAD3F5")
	colorSubtext    = lipgloss.Color("#A5ADCB")
	colorError      = lipgloss.Color("#ED8796")
	colorBackground = lipgloss.Color("#24273A")
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary)

	loginBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSecondary).
			Padding(1, 4).
			Width(60).
			Align(lipgloss.Left)

	inputFocusedStyle = lipgloss.NewStyle().Foreground(colorPrimary)
	inputBlurredStyle = lipgloss.NewStyle().Foreground(colorSubtext)

	buttonStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorSecondary).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.
				Foreground(colorBackground).
				Background(colorPrimary).
				Bold(true)

	listTitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			MarginLeft(2).
			UnsetBackground()

	listItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedListItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(colorPrimary).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorPrimary)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorSubtext)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSubtext).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1)

	nowPlayingStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSecondary).
			Padding(0, 2).
			MarginTop(1)

	albumHeaderStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Bold(true).
				MarginTop(1)
)
