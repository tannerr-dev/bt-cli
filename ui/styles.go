package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(true).
			Width(50)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Width(50)

	StatusConnected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	StatusDisconnected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))
)
