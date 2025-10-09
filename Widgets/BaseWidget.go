package widgets

import tea "github.com/charmbracelet/bubbletea"

type Widget interface {
	Update(msg tea.Msg) (Widget, tea.Cmd)
	View() string
	SetFocused(bool)
	IsFocused() bool
	SetDimensions(width, height int)
}

type BaseWidget struct {
	focused bool
	width   int
	height  int
}

func (b *BaseWidget) SetFocused(focused bool) {
	b.focused = focused
}

func (b *BaseWidget) IsFocused() bool {
	return b.focused
}

func (b *BaseWidget) SetDimensions(width, height int) {
	b.width = width
	b.height = height
}
