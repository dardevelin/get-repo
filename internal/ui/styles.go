package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#5D5D5D")).
			Padding(0, 1)

	// Text styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	PendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	// Box styles
	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	// Progress styles
	ProgressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))
)
