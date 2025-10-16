package main

import (
	"fmt"
	"l8zykube/components"
	"l8zykube/kubernetes"
	widgets "l8zykube/widgets"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainModel struct {
	widgets       []widgets.Widget
	focusedWidget int
	width         int
	height        int
	kubeClient    *kubernetes.KubeClient
	modal         *components.Modal
	logsModal     *components.LogsModal
	showModal     bool
	showLogsModal bool
}

func initialModel() MainModel {
	widgets := []widgets.Widget{
		widgets.NewNameSpaceWidget(),
		widgets.NewApiResourceWidget(),
		widgets.NewMainContentWidget(),
	}

	widgets[0].SetFocused(true)

	kubeClient, err := kubernetes.NewKubeClient()
	showModal := false
	if err != nil {
		fmt.Printf("Warning: Could not connect to Kubernetes: %v\n", err)
		showModal = true
	}

	if kubeClient != nil {
		if apiResources, err := kubeClient.GetAPIResources(); err != nil {
			fmt.Printf("Error fetching API resources: %v\n", err)
			showModal = true
		} else {
			if arw, ok := widgets[1].(interface{ SetApiResourceList([]string) }); ok {
				arw.SetApiResourceList(apiResources)
			}
		}
	}

	modal := components.NewModal()
	logsModal := components.NewLogsModal()
	if showModal {
		modal.ShowError("Kubernetes Connection Failed", "Could not connect to Kubernetes cluster.\nPlease check your kubeconfig and cluster status.\nMake sure minikube is running: minikube start", "Q")
	}

	return MainModel{
		widgets:       widgets,
		focusedWidget: 0,
		kubeClient:    kubeClient,
		modal:         modal,
		logsModal:     logsModal,
		showModal:     showModal,
		showLogsModal: false,
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
		case "ctrl+c", "q", "Q":
			if m.showModal {
				m.modal.Hide()
				m.showModal = false
				return m, tea.Quit
			}
			if m.showLogsModal {
				m.logsModal.Hide()
				m.showLogsModal = false
				return m, tea.Quit
			}
			return m, tea.Quit

		case tea.KeyEscape.String():
			if m.showModal {
				m.modal.Hide()
				m.showModal = false
				return m, nil
			}
			if m.showLogsModal {
				m.logsModal.Hide()
				m.showLogsModal = false
				return m, nil
			}
			// If ApiResourceWidget is active, deactivate it
			if apiResourceWidget, ok := m.widgets[1].(*widgets.ApiResourceWidget); ok {
				if apiResourceWidget.IsListActive() {
					var cmd tea.Cmd
					m.widgets[1], cmd = m.widgets[1].Update(msg)
					return m, cmd
				}
			}
			// If MainContent is in selection mode, exit it
			if mainContentWidget, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
				if mainContentWidget.SelectionNameSpace {
					mainContentWidget.SetSelectionNameSpace(false)
					m.widgets[2].SetFocused(false)
					m.widgets[0].SetFocused(true)
					m.focusedWidget = 0
					return m, nil
				}
			}
			// Route Escape key to the focused widget for other cases
			if m.focusedWidget < len(m.widgets) {
				var cmd tea.Cmd
				m.widgets[m.focusedWidget], cmd = m.widgets[m.focusedWidget].Update(msg)
				return m, cmd
			}

		case "ctrl+l":
			if m.kubeClient != nil {
				if mainContentWidget, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
					selected := mainContentWidget.GetSelectedResource()
					if selected != nil && selected.Type == "Pod" {
						logs, err := m.kubeClient.GetPodLogs(selected.Namespace, selected.Name, 1000)
						if err != nil {
							m.modal.ShowError("Logs Error", fmt.Sprintf("Failed to get logs:\n%v", err), "Close")
							m.showModal = true
						} else {
							logLineCount := len(strings.Split(logs, "\n"))
							m.logsModal.Show(fmt.Sprintf("Pod Logs: %s (namespace: %s) - %d lines", selected.Name, selected.Namespace, logLineCount), logs)
							m.logsModal.SetDimensions(m.width, m.height)
							m.showLogsModal = true
						}
						return m, nil
					} else {
						m.modal.ShowError("No Pod Selected", "Please select a pod to view logs", "Close")
						m.showModal = true
						return m, nil
					}
				}
			} else {
				m.modal.ShowError("No Connection", "Not connected to Kubernetes cluster", "Close")
				m.showModal = true
				return m, nil
			}
			return m, nil

		case "up":
			if m.showLogsModal {
				m.logsModal.ScrollUp()
				return m, nil
			}

		case "down":
			if m.showLogsModal {
				m.logsModal.ScrollDown()
				return m, nil
			}

		case "pgup":
			if m.showLogsModal {
				m.logsModal.PageUp()
				return m, nil
			}

		case "pgdown":
			if m.showLogsModal {
				m.logsModal.PageDown()
				return m, nil
			}

		case "home", "g":
			if m.showLogsModal {
				m.logsModal.ScrollToTop()
				return m, nil
			}

		case "end", "G":
			if m.showLogsModal {
				m.logsModal.ScrollToBottom()
				return m, nil
			}

		case "left", "h":
			if m.showModal {
				m.modal.PrevButton()
				return m, nil
			}

		case "right", "l":
			if m.showModal {
				m.modal.NextButton()
				return m, nil
			}

		case "j", "k":
			// If logs modal is showing, handle scrolling
			if m.showLogsModal {
				if msg.String() == "j" {
					m.logsModal.ScrollDown()
				} else {
					m.logsModal.ScrollUp()
				}
				return m, nil
			}

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

	if m.showLogsModal {
		logsContent := m.logsModal.Render()
		logsStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center)

		overlay := logsStyle.Render(logsContent)
		return overlay
	}

	if m.showModal {
		modalContent := m.modal.Render()
		modalStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center)

		overlay := modalStyle.Render(modalContent)
		return overlay
	}

	// Append dynamic footer with keybindings
	footer := m.renderFooter()
	bodyWithFooter := lipgloss.JoinVertical(lipgloss.Left, horizontal, footer)
	return bodyWithFooter
}

// renderFooter builds a dynamic footer string showing current keybindings
// based on the active UI state and focused widget.
func (m MainModel) renderFooter() string {
	// Base style for footer
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(m.width)

	// Determine current context
	var hints []string

	if m.showModal {
		hints = append(hints, "esc: close modal", "q: quit")
		return style.Render(strings.Join(hints, "  |  "))
	}

	if m.showLogsModal {
		hints = append(hints,
			"up/down, j/k: scroll",
			"pgup/pgdown: page",
			"g/G, home/end: jump",
			"esc: close",
			"q: quit",
		)
		return style.Render(strings.Join(hints, "  |  "))
	}

	// Focus-specific hints
	switch m.focusedWidget {
	case 0: // Namespace widget
		hints = append(hints, "j/k: move focus", "enter: choose namespace", "q: quit")
	case 1: // API resource widget
		if arw, ok := m.widgets[1].(*widgets.ApiResourceWidget); ok && arw.IsListActive() {
			hints = append(hints, "j/k: move", "enter: select", "esc: back", "q: quit")
		} else {
			hints = append(hints, "j/k: move focus", "enter: open resources", "q: quit")
		}
	case 2: // Main content widget
		if mcw, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
			if mcw.SelectionNameSpace {
				hints = append(hints, "j/k: move", "enter: select namespace", "esc: cancel", "q: quit")
			} else if mcw.IsResourcesActive() {
				hints = append(hints, "j/k, up/down: scroll", "pgup/pgdown: page", "g/G: jump", "esc: exit", "q: quit")
				if sel := mcw.GetSelectedResource(); sel != nil && sel.Type == "Pod" {
					hints = append(hints, "ctrl+l: view logs")
				}
			} else {
				hints = append(hints, "enter: activate list", "j/k: move focus", "q: quit")
			}
		}
	default:
		hints = append(hints, "j/k: move focus", "q: quit")
	}

	return style.Render(strings.Join(hints, "  |  "))
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
