package widgets

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NameSpaceWidget struct {
	BaseWidget
	SelectedNameSpace  string
	SelectionNameSpace bool
}

func NewNameSpaceWidget() *NameSpaceWidget {
	return &NameSpaceWidget{
		BaseWidget: BaseWidget{
			focused: false,
		},
		SelectedNameSpace: "default",
	}
}

func (n *NameSpaceWidget) SetSelectedNameSpace(nameSpace string) {
	n.SelectedNameSpace = nameSpace
}

func (n *NameSpaceWidget) GetSelectedNameSpace() string {
	return n.SelectedNameSpace
}

func (n *NameSpaceWidget) Update(msg tea.Msg) (Widget, tea.Cmd) {
	if !n.focused {
		return n, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyEnter.String():
			n.SelectionNameSpace = true
		case tea.KeyEscape.String():
			n.SelectionNameSpace = false
		}
	}
	return n, nil
}

func (n *NameSpaceWidget) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(30)

	if n.focused {
		style = style.BorderForeground(lipgloss.Color("205"))
	} else {
		style = style.BorderForeground(lipgloss.Color("240"))
	}

	content := fmt.Sprintf("NameSpace: %s", n.SelectedNameSpace)
	return style.Render(content)
}
