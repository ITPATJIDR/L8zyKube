package widgets

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ApiResourceWidget struct {
	BaseWidget
	SelectedNameSpace  string
	SelectionNameSpace bool
}

func NewApiResourceWidget() *ApiResourceWidget {
	return &ApiResourceWidget{
		BaseWidget: BaseWidget{
			focused: false,
		},
	}
}

func (a *ApiResourceWidget) Update(msg tea.Msg) (Widget, tea.Cmd) {
	if !a.focused {
		return a, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyEnter.String():
			a.SelectionNameSpace = true
		case tea.KeyEscape.String():
			a.SelectionNameSpace = false
		}
	}
	return a, nil
}

func (a *ApiResourceWidget) SetDimensions(width, height int) {
	a.BaseWidget.SetDimensions(30, height)
}

func (a *ApiResourceWidget) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(30)

	if a.height > 0 {
		style = style.Height(a.height)
	} else {
		style = style.Height(32)
	}

	if a.focused {
		style = style.BorderForeground(lipgloss.Color("205"))
	} else {
		style = style.BorderForeground(lipgloss.Color("240"))
	}

	content := fmt.Sprintf("ApiResource: %s", a.SelectedNameSpace)
	return style.Render(content)
}
