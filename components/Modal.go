package components

import (
	"github.com/charmbracelet/lipgloss"
)

type ModalType int

const (
	ModalError ModalType = iota
	ModalWarning
	ModalInfo
	ModalSuccess
)

type Modal struct {
	Width       int
	Height      int
	Title       string
	Message     string
	Visible     bool
	Type        ModalType
	Buttons     []string
	SelectedBtn int
	OnConfirm   func()
	OnCancel    func()
}

func NewModal() *Modal {
	return &Modal{
		Visible:     false,
		Type:        ModalInfo,
		Buttons:     []string{"OK"},
		SelectedBtn: 0,
	}
}

func (m *Modal) SetDimensions(width, height int) {
	m.Width = width
	m.Height = height
}

func (m *Modal) Show(title, message string) {
	m.Title = title
	m.Message = message
	m.Visible = true
	m.SelectedBtn = 0
}

func (m *Modal) ShowWithType(title, message string, modalType ModalType) {
	m.Title = title
	m.Message = message
	m.Type = modalType
	m.Visible = true
	m.SelectedBtn = 0
}

func (m *Modal) ShowWithButtons(title, message string, modalType ModalType, buttons []string) {
	m.Title = title
	m.Message = message
	m.Type = modalType
	m.Buttons = buttons
	m.Visible = true
	m.SelectedBtn = 0
}

func (m *Modal) Hide() {
	m.Visible = false
}

func (m *Modal) NextButton() {
	if len(m.Buttons) > 0 {
		m.SelectedBtn = (m.SelectedBtn + 1) % len(m.Buttons)
	}
}

func (m *Modal) PrevButton() {
	if len(m.Buttons) > 0 {
		m.SelectedBtn = (m.SelectedBtn - 1 + len(m.Buttons)) % len(m.Buttons)
	}
}

func (m *Modal) SelectButton() {
	if m.SelectedBtn < len(m.Buttons) {
		button := m.Buttons[m.SelectedBtn]
		if button == "OK" || button == "Yes" || button == "Confirm" {
			if m.OnConfirm != nil {
				m.OnConfirm()
			}
		} else if button == "Cancel" || button == "No" {
			if m.OnCancel != nil {
				m.OnCancel()
			}
		}
		m.Hide()
	}
}

func (m *Modal) getModalColors() (borderColor, titleColor string) {
	switch m.Type {
	case ModalError:
		return "196", "196" // Red
	case ModalWarning:
		return "214", "214" // Orange
	case ModalSuccess:
		return "46", "46" // Green
	case ModalInfo:
		return "33", "33" // Blue
	default:
		return "240", "240" // Gray
	}
}

func (m *Modal) Render() string {
	if !m.Visible {
		return ""
	}

	// Calculate modal dimensions (centered)
	modalWidth := 60
	modalHeight := 10

	borderColor, titleColor := m.getModalColors()

	// Create the modal border
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(titleColor)).
		Bold(true).
		Align(lipgloss.Center).
		Width(modalWidth-4).
		Margin(0, 0, 1, 0)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(modalWidth-4).
		Margin(1, 0, 0, 0)

	// Button style
	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(modalWidth-4).
		Margin(1, 0, 0, 0)

	// Selected button style
	selectedButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center).
		Width(modalWidth-4).
		Margin(1, 0, 0, 0)

	title := titleStyle.Render(m.Title)
	message := messageStyle.Render(m.Message)

	// Render buttons
	var buttonText string
	if len(m.Buttons) > 0 {
		buttons := make([]string, len(m.Buttons))
		for i, btn := range m.Buttons {
			if i == m.SelectedBtn {
				buttons[i] = selectedButtonStyle.Render("[" + btn + "]")
			} else {
				buttons[i] = buttonStyle.Render(" " + btn + " ")
			}
		}
		buttonText = lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
	} else {
		buttonText = buttonStyle.Render("Press q to close")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		message,
		buttonText,
	)

	return modalStyle.Render(content)
}

// Helper functions for common modal types
func (m *Modal) ShowError(title, message string, buttons string) {
	m.ShowWithType(title, message, ModalError)
	if buttons != "" {
		m.Buttons = []string{buttons}
	}
}

func (m *Modal) ShowWarning(title, message string) {
	m.ShowWithType(title, message, ModalWarning)
}

func (m *Modal) ShowInfo(title, message string) {
	m.ShowWithType(title, message, ModalInfo)
}

func (m *Modal) ShowSuccess(title, message string) {
	m.ShowWithType(title, message, ModalSuccess)
}

func (m *Modal) ShowConfirm(title, message string, onConfirm, onCancel func()) {
	m.ShowWithButtons(title, message, ModalInfo, []string{"Yes", "No"})
	m.OnConfirm = onConfirm
	m.OnCancel = onCancel
}
