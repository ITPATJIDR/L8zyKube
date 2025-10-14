package widgets

import (
	"l8zykube/components"
	kubetypes "l8zykube/kubernetes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainContentWidget struct {
	BaseWidget
	SelectionNameSpace bool
	namespaceSelector  *components.NamespaceSelector
	resourceTable      *components.ResourceTable
	welcomeScreen      *components.WelcomeScreen
}

func NewMainContentWidget() *MainContentWidget {
	return &MainContentWidget{
		BaseWidget: BaseWidget{
			focused: false,
		},
		namespaceSelector: components.NewNamespaceSelector(),
		resourceTable:     components.NewResourceTable(),
		welcomeScreen:     components.NewWelcomeScreen(),
	}
}

func (m *MainContentWidget) Update(msg tea.Msg) (Widget, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	var cmd tea.Cmd

	// If we're in namespace selection mode, handle list interactions
	if m.SelectionNameSpace {
		cmd = m.namespaceSelector.Update(msg)
		return m, cmd
	}

	// Default behavior when not in selection mode
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case tea.KeyEnter.String():
			if len(m.resourceTable.Resources) > 0 {
				m.resourceTable.SetActive(true)
			}
		case tea.KeyEscape.String():
			m.resourceTable.SetActive(false)
		}

		if m.resourceTable.Active && len(m.resourceTable.Resources) > 0 {
			switch key {
			case "down", "j":
				m.resourceTable.ScrollDown()
			case "up", "k":
				m.resourceTable.ScrollUp()
			case "pgdown":
				m.resourceTable.PageDown()
			case "pgup":
				m.resourceTable.PageUp()
			case "home", "g":
				m.resourceTable.ScrollToTop()
			case "end", "G":
				m.resourceTable.ScrollToBottom()
			}
		}
	}
	return m, nil
}

func (m *MainContentWidget) SetSelectionNameSpace(isSelection bool) {
	m.SelectionNameSpace = isSelection
}

func (m *MainContentWidget) SetNamespaceList(namespaces []string) {
	m.namespaceSelector.SetNamespaceList(namespaces)
}

func (m *MainContentWidget) GetSelectedNamespace() string {
	return m.namespaceSelector.GetSelectedNamespace()
}

func (m *MainContentWidget) SetDimensions(width, height int) {
	m.BaseWidget.SetDimensions(width, height)
	m.namespaceSelector.SetDimensions(width, height)
	m.resourceTable.SetDimensions(width, height)
	m.welcomeScreen.SetDimensions(width, height)
}

func (m *MainContentWidget) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Height(m.height) // Force fixed height to prevent overflow

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
		content = m.namespaceSelector.Render()
	} else if len(m.resourceTable.Resources) > 0 {
		content = m.resourceTable.Render()
	} else {
		content = m.welcomeScreen.Render()
	}

	return style.Render(content)
}

// Public API to set resources for rendering
func (m *MainContentWidget) SetResourcesDetailed(title string, resources []kubetypes.ResourceInfo) {
	m.resourceTable.SetResources(title, resources)
}

func (m *MainContentWidget) IsResourcesActive() bool {
	return m.resourceTable.Active
}

func (m *MainContentWidget) GetSelectedResource() *kubetypes.ResourceInfo {
	return m.resourceTable.GetSelectedResource()
}
