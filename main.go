package main

import (
	"fmt"
	widgets "l8zykube/Widgets"
	"l8zykube/kubernetes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainModel struct {
	widgets       []widgets.Widget
	focusedWidget int
	width         int
	height        int
	kubeClient    *kubernetes.KubeClient
}

func initialModel() MainModel {
	widgets := []widgets.Widget{
		widgets.NewNameSpaceWidget(),
		widgets.NewApiResourceWidget(),
		widgets.NewMainContentWidget(),
	}

	widgets[0].SetFocused(true)

	kubeClient, err := kubernetes.NewKubeClient()
	if err != nil {
		fmt.Printf("Warning: Could not connect to Kubernetes: %v\n", err)
	}

	return MainModel{
		widgets:       widgets,
		focusedWidget: 0,
		kubeClient:    kubeClient,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "j", "k":
			// Check if we're in namespace selection mode - if so, don't switch widgets
			if mainContentWidget, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
				if mainContentWidget.SelectionNameSpace && m.focusedWidget == 2 {
					// We're in selection mode, let the MainContent widget handle j/k
					// Don't switch widgets, just pass the key to the focused widget
					var cmd tea.Cmd
					m.widgets[m.focusedWidget], cmd = m.widgets[m.focusedWidget].Update(msg)
					return m, cmd
				}
			}

			// Normal widget switching behavior
			oldIndex := m.focusedWidget

			switch msg.String() {
			case "j":
				m.focusedWidget--
			case "k":
				m.focusedWidget++

			}
			if m.focusedWidget < 0 {
				m.focusedWidget = len(m.widgets) - 1
			} else if m.focusedWidget >= len(m.widgets) {
				m.focusedWidget = 0
			}

			m.widgets[oldIndex].SetFocused(false)
			m.widgets[m.focusedWidget].SetFocused(true)
			return m, nil

		default:
			var cmd tea.Cmd

			if namespaceWidget, ok := m.widgets[0].(*widgets.NameSpaceWidget); ok {
				if mainContentWidget, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
					if m.focusedWidget == 0 && msg.String() == tea.KeyEnter.String() {
						m.widgets[0].SetFocused(false)
						m.widgets[2].SetFocused(true)
						m.focusedWidget = 2
						mainContentWidget.SetSelectionNameSpace(true)

						if m.kubeClient != nil {
							namespaces, err := m.kubeClient.GetNamespaces()
							if err != nil {
								fmt.Printf("Error fetching namespaces: %v\n", err)
							} else {
								mainContentWidget.SetNamespaceList(namespaces)
							}
						}
						return m, nil
					}

					// Handle Enter key in MainContent when in selection mode
					if m.focusedWidget == 2 && mainContentWidget.SelectionNameSpace && msg.String() == tea.KeyEnter.String() {
						// Select current item and exit selection mode
						selectedNS := mainContentWidget.GetSelectedNamespace()
						if selectedNS != "" {
							namespaceWidget.SetSelectedNameSpace(selectedNS)
							mainContentWidget.SetSelectionNameSpace(false)
							// Switch focus back to NameSpace widget
							m.widgets[2].SetFocused(false)
							m.widgets[0].SetFocused(true)
							m.focusedWidget = 0
							return m, nil
						}
					}

					// Handle Escape key in MainContent when in selection mode
					if m.focusedWidget == 2 && mainContentWidget.SelectionNameSpace && msg.String() == tea.KeyEscape.String() {
						// Exit selection mode and switch back to NameSpace widget
						mainContentWidget.SetSelectionNameSpace(false)
						m.widgets[2].SetFocused(false)
						m.widgets[0].SetFocused(true)
						m.focusedWidget = 0
						return m, nil
					}
				}
			}

			m.widgets[m.focusedWidget], cmd = m.widgets[m.focusedWidget].Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m MainModel) View() string {
	namespaceWidth := 30
	apiResourceWidth := 30
	apiResourceHeight := m.height - 9
	mainContentWidth := m.width - namespaceWidth - 4
	mainContentHeight := m.height - 4

	if namespaceWidget, ok := m.widgets[0].(*widgets.NameSpaceWidget); ok {
		namespaceWidget.SetDimensions(namespaceWidth, 0)
	}
	if apiResourceWidget, ok := m.widgets[1].(*widgets.ApiResourceWidget); ok {
		apiResourceWidget.SetDimensions(apiResourceWidth, apiResourceHeight)
	}
	if mainContentWidget, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
		mainContentWidget.SetDimensions(mainContentWidth, mainContentHeight)
	}

	vertical := lipgloss.JoinVertical(lipgloss.Top, m.widgets[0].View(), m.widgets[1].View())
	horizontal := lipgloss.JoinHorizontal(lipgloss.Top, vertical, m.widgets[2].View())

	return horizontal
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
