package main

import (
	"fmt"
	"l8zykube/kubernetes"
	widgets "l8zykube/widgets"

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

	// Load API resources immediately using default namespace (list is cluster-wide types)
	if kubeClient != nil {
		if apiResources, err := kubeClient.GetAPIResources(); err != nil {
			fmt.Printf("Error fetching API resources: %v\n", err)
		} else {
			// Avoid package alias shadowing by asserting against a local interface
			if arw, ok := widgets[1].(interface{ SetApiResourceList([]string) }); ok {
				arw.SetApiResourceList(apiResources)
			}
		}
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
			// If MainContent is in selection mode or resources-active, route j/k to it
			if mainContentWidget, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
				if m.focusedWidget == 2 && (mainContentWidget.SelectionNameSpace || mainContentWidget.IsResourcesActive()) {
					var cmd tea.Cmd
					m.widgets[m.focusedWidget], cmd = m.widgets[m.focusedWidget].Update(msg)
					return m, cmd
				}
			}

			// If ApiResourceWidget is focused and activated, route j/k to it
			if apiResourceWidget, ok := m.widgets[1].(*widgets.ApiResourceWidget); ok {
				if m.focusedWidget == 1 && apiResourceWidget.IsListActive() {
					var cmd tea.Cmd
					m.widgets[m.focusedWidget], cmd = m.widgets[m.focusedWidget].Update(msg)
					return m, cmd
				}
			}

			// Otherwise, switch widgets
			oldIndex := m.focusedWidget
			if msg.String() == "j" {
				m.focusedWidget--
			} else {
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

			if apiResourceWidget, ok := m.widgets[1].(*widgets.ApiResourceWidget); ok {
				if m.focusedWidget == 1 && msg.String() == tea.KeyEnter.String() {
					var cmd tea.Cmd
					m.widgets[1], cmd = m.widgets[1].Update(msg)
					// If already active, Enter selects and we fetch details
					if apiResourceWidget.IsListActive() {
						selectedResource := apiResourceWidget.GetSelectedApiResource()
						if selectedResource != "" && m.kubeClient != nil {
							// Determine current namespace
							currentNS := "default"
							if namespaceWidget, ok := m.widgets[0].(*widgets.NameSpaceWidget); ok {
								currentNS = namespaceWidget.GetSelectedNameSpace()
							}
							// Fetch resources and render in main content
							resources, err := m.kubeClient.GetResourceListDetailed(selectedResource, currentNS)
							if err != nil {
								fmt.Printf("Error fetching %s in %s: %v\n", selectedResource, currentNS, err)
							} else if mainContent, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
								mainContent.SetResourcesDetailed(fmt.Sprintf("%s in %s", selectedResource, currentNS), resources)
							}
						}
					}
					return m, cmd
				}
			}

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

							// Load API resources for the selected namespace
							if m.kubeClient != nil {
								if apiResourceWidget, ok := m.widgets[1].(*widgets.ApiResourceWidget); ok {
									apiResources, err := m.kubeClient.GetAPIResources()
									if err != nil {
										fmt.Printf("Error fetching API resources: %v\n", err)
									} else {
										apiResourceWidget.SetApiResourceList(apiResources)
									}
								}
							}

							m.widgets[2].SetFocused(false)
							m.widgets[0].SetFocused(true)
							m.focusedWidget = 0
							return m, nil
						}
					}

					if m.focusedWidget == 2 && mainContentWidget.SelectionNameSpace && msg.String() == tea.KeyEscape.String() {
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
	apiResourceHeight := m.height - 8
	mainContentWidth := m.width - namespaceWidth - 4
	mainContentHeight := m.height - 3

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
