package tui

import "github.com/charmbracelet/lipgloss"

// Color Palette
var (
	colorPrimary    = lipgloss.Color("#AA5CDB")
	colorSecondary  = lipgloss.Color("#00A4DC")
	colorText       = lipgloss.Color("#FAFAFA")
	colorSubtext    = lipgloss.Color("#A1A1A1")
	colorError      = lipgloss.Color("#FF5555")
	colorBackground = lipgloss.Color("#101010")
)

var (
	baseStyle = lipgloss.NewStyle().Foreground(colorText)
	appStyle  = lipgloss.NewStyle().Padding(1, 2)

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
			Align(lipgloss.Center)

	inputFocusedStyle = lipgloss.NewStyle().Foreground(colorPrimary)
	inputBlurredStyle = lipgloss.NewStyle().Foreground(colorSubtext)

	buttonStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorSecondary).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.Copy().
				Foreground(colorBackground).
				Background(colorPrimary).
				Bold(true)

	listTitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			MarginLeft(2)

	listItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedListItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(colorPrimary).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorPrimary)

	detailTitleStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true).
				MarginBottom(1).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(colorSubtext)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Width(12)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorText)

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

	progressBarStyle = lipgloss.NewStyle().
				Foreground(colorPrimary)

	albumHeaderStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Bold(true).
				MarginTop(1)
)
