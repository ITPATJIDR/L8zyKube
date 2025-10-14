package components

import (
	"github.com/charmbracelet/lipgloss"
)

type WelcomeScreen struct {
	Width  int
	Height int
}

func NewWelcomeScreen() *WelcomeScreen {
	return &WelcomeScreen{}
}

func (ws *WelcomeScreen) SetDimensions(width, height int) {
	ws.Width = width
	ws.Height = height
}

func (ws *WelcomeScreen) Render() string {
	// ASCII art text
	asciiArt := `
 __    ___ _____ __ __ _____ _____ _____ _____ 
|  |  | . |   __|  |  |  |  |  |  | __  |   __|
|  |__| . |__   |_   _|    -|  |  | __ -|   __|
|_____|___|_____| |_| |__|__|_____|_____|_____|
`

	styledAscii := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center).
		Width(ws.Width - 4).
		Render(asciiArt)

	welcomeText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("87")).
		Bold(true).
		Align(lipgloss.Center).
		Width(ws.Width-4).
		Margin(0, 0, 1, 0).
		Render("Welcome to L8zyKube!")

	instructionText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(ws.Width-4).
		Margin(1, 0, 0, 0).
		Render("Press Enter in NameSpace widget to select a namespace")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		welcomeText,
		styledAscii,
		instructionText,
	)
}
