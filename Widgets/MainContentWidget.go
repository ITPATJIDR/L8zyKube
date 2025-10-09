package widgets

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NamespaceItem represents an item in the namespace list
type NamespaceItem struct {
	title, desc string
}

func (i NamespaceItem) Title() string       { return i.title }
func (i NamespaceItem) Description() string { return i.desc }
func (i NamespaceItem) FilterValue() string { return i.title }

type MainContentWidget struct {
	BaseWidget
	SelectionNameSpace bool
	NamespaceList      []string
	list               list.Model
	selectedNamespace  string
}

func NewMainContentWidget() *MainContentWidget {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Namespace"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.MarginLeft(2)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.MarginLeft(2)

	return &MainContentWidget{
		BaseWidget: BaseWidget{
			focused: false,
		},
		list: l,
	}
}

func (m *MainContentWidget) Update(msg tea.Msg) (Widget, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	var cmd tea.Cmd

	// If we're in namespace selection mode, handle list interactions
	if m.SelectionNameSpace {
		// Update the list with the message
		m.list, cmd = m.list.Update(msg)

		// Update our selected namespace from the list
		if selectedItem := m.list.SelectedItem(); selectedItem != nil {
			if item, ok := selectedItem.(NamespaceItem); ok {
				m.selectedNamespace = item.title
			}
		}

		return m, cmd
	}

	// Default behavior when not in selection mode
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyEnter.String():

		case tea.KeyEscape.String():

		}
	}
	return m, nil
}

func (m *MainContentWidget) SetSelectionNameSpace(isSelection bool) {
	m.SelectionNameSpace = isSelection
}

func (m *MainContentWidget) SetNamespaceList(namespaces []string) {
	m.NamespaceList = namespaces

	items := make([]list.Item, len(namespaces))
	for i, ns := range namespaces {
		items[i] = NamespaceItem{
			title: ns,
			desc:  fmt.Sprintf("Namespace: %s", ns),
		}
	}

	m.list.SetItems(items)
}

func (m *MainContentWidget) GetSelectedNamespace() string {
	// Get the currently selected item from the list
	if selectedItem := m.list.SelectedItem(); selectedItem != nil {
		if item, ok := selectedItem.(NamespaceItem); ok {
			return item.title
		}
	}
	return m.selectedNamespace
}

func (m *MainContentWidget) SetDimensions(width, height int) {
	m.BaseWidget.SetDimensions(width, height)
	m.list.SetSize(width-4, height-4)
}

func (m *MainContentWidget) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)

	if m.width > 0 {
		style = style.Width(m.width)
	} else {
		style = style.Width(90)
	}

	if m.height > 0 {
		style = style.Height(m.height)
	} else {
		style = style.Height(37)
	}

	if m.focused {
		style = style.BorderForeground(lipgloss.Color("205"))
	} else {
		style = style.BorderForeground(lipgloss.Color("240"))
	}

	var content string
	if m.SelectionNameSpace {
		content = m.list.View()
	} else {
		// ASCII art text
		asciiArt := `
 __    ___ _____ __ __ _____ _____ _____ _____ 
|  |  | . |   __|  |  |  |  |  |  | __  |   __|
|  |__| . |__   |_   _|    -|  |  | __ -|   __|
|_____|___|_____| |_| |__|__|_____|_____|_____|
`

		// Create styled ASCII art with proper width centering
		styledAscii := lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")). // Pink/magenta color
			Bold(true).
			Align(lipgloss.Center).
			Width(m.width - 4). // Account for widget padding
			Render(asciiArt)

		// Add welcome text with proper width centering
		welcomeText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("87")). // Light blue
			Bold(true).
			Align(lipgloss.Center).
			Width(m.width-4). // Account for widget padding
			Margin(0, 0, 1, 0).
			Render("Welcome to L8zyKube!")

		// Add instruction text with proper width centering
		instructionText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Gray
			Align(lipgloss.Center).
			Width(m.width-4). // Account for widget padding
			Margin(1, 0, 0, 0).
			Render("Press Enter in NameSpace widget to select a namespace")

		// Combine all elements and center them
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			welcomeText,
			styledAscii,
			instructionText,
		)
	}

	return style.Render(content)
}
