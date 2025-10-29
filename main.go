package main

import (
	"fmt"
	"l8zykube/components"
	"l8zykube/kubernetes"
	widgets "l8zykube/widgets"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainModel struct {
	widgets        []widgets.Widget
	focusedWidget  int
	width          int
	height         int
	kubeClient     *kubernetes.KubeClient
	modal          *components.Modal
	logsModal      *components.LogsModal
	showModal      bool
	showLogsModal  bool
	watching       bool
	watchResource  string
	watchNamespace string
}

type WatchTick struct{}

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
		modal.ShowError("Kubernetes Connection Failed", "Could not connect to Kubernetes cluster.\nPlease check your kubeconfig and cluster status.\nMake sure minikube is running: minikube start", "Ctrl+Q")
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

func watchTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg { return WatchTick{} })
}

func normalizeResourceTypeForFetch(rt string) string {
	r := strings.ToLower(strings.TrimSpace(rt))
	switch r {
	case "pod":
		return "pods"
	case "service":
		return "services"
	case "deployment":
		return "deployments"
	}
	return r
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q":
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
			if apiResourceWidget, ok := m.widgets[1].(*widgets.ApiResourceWidget); ok {
				if apiResourceWidget.IsListActive() {
					var cmd tea.Cmd
					m.widgets[1], cmd = m.widgets[1].Update(msg)
					return m, cmd
				}
			}
			if mainContentWidget, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
				if mainContentWidget.SelectionNameSpace {
					mainContentWidget.SetSelectionNameSpace(false)
					m.widgets[2].SetFocused(false)
					m.widgets[0].SetFocused(true)
					m.focusedWidget = 0
					return m, nil
				}
			}
			if m.focusedWidget < len(m.widgets) {
				var cmd tea.Cmd
				m.widgets[m.focusedWidget], cmd = m.widgets[m.focusedWidget].Update(msg)
				return m, cmd
			}

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
			if m.showLogsModal {
				if msg.String() == "j" {
					m.logsModal.ScrollDown()
				} else {
					m.logsModal.ScrollUp()
				}
				return m, nil
			}

			if mainContentWidget, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
				if m.focusedWidget == 2 && (mainContentWidget.SelectionNameSpace || mainContentWidget.IsResourcesActive()) {
					var cmd tea.Cmd
					m.widgets[m.focusedWidget], cmd = m.widgets[m.focusedWidget].Update(msg)
					return m, cmd
				}
			}

			if apiResourceWidget, ok := m.widgets[1].(*widgets.ApiResourceWidget); ok {
				if m.focusedWidget == 1 && apiResourceWidget.IsListActive() {
					var cmd tea.Cmd
					m.widgets[m.focusedWidget], cmd = m.widgets[m.focusedWidget].Update(msg)
					return m, cmd
				}
			}

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
					if apiResourceWidget.IsListActive() {
						selectedResource := apiResourceWidget.GetSelectedApiResource()
						if selectedResource != "" && m.kubeClient != nil {
							currentNS := "default"
							if namespaceWidget, ok := m.widgets[0].(*widgets.NameSpaceWidget); ok {
								currentNS = namespaceWidget.GetSelectedNameSpace()
							}
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

					if m.focusedWidget == 2 && mainContentWidget.SelectionNameSpace && msg.String() == tea.KeyEnter.String() {
						selectedNS := mainContentWidget.GetSelectedNamespace()
						if selectedNS != "" {
							namespaceWidget.SetSelectedNameSpace(selectedNS)
							mainContentWidget.SetSelectionNameSpace(false)

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

	switch msg := msg.(type) {
	case widgets.ShowLogsRequest:
		if m.kubeClient == nil {
			m.modal.ShowError("No Connection", "Not connected to Kubernetes cluster", "Close")
			m.showModal = true
			return m, nil
		}
		if strings.EqualFold(msg.Resource.Type, "Pod") || strings.EqualFold(msg.Resource.Type, "Pods") {
			logs, err := m.kubeClient.GetPodLogs(msg.Resource.Namespace, msg.Resource.Name, 1000)
			if err != nil {
				m.modal.ShowError("Logs Error", fmt.Sprintf("Failed to get logs:\n%v", err), "Close")
				m.showModal = true
				return m, nil
			}
			logLineCount := len(strings.Split(logs, "\n"))
			m.logsModal.Show(fmt.Sprintf("Pod Logs: %s (namespace: %s) - %d lines", msg.Resource.Name, msg.Resource.Namespace, logLineCount), logs)
			m.logsModal.SetDimensions(m.width, m.height)
			m.showLogsModal = true
			return m, nil
		}
		m.modal.ShowError("No Pod Selected", "Please select a pod to view logs", "Close")
		m.showModal = true
		return m, nil

	case widgets.ToggleWatchRequest:
		if m.kubeClient == nil {
			m.modal.ShowError("No Connection", "Not connected to Kubernetes cluster", "Close")
			m.showModal = true
			return m, nil
		}
		rt := normalizeResourceTypeForFetch(msg.ResourceType)
		ns := msg.Namespace
		if ns == "" {
			if namespaceWidget, ok := m.widgets[0].(*widgets.NameSpaceWidget); ok {
				ns = namespaceWidget.GetSelectedNameSpace()
			}
			if ns == "" {
				ns = "default"
			}
		}

		if m.watching && m.watchResource == rt && m.watchNamespace == ns {
			m.watching = false
			if mainContent, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
				mainContent.SetWatching(false)
			}
			return m, nil
		}

		m.watching = true
		m.watchResource = rt
		m.watchNamespace = ns

		if resources, err := m.kubeClient.GetResourceListDetailed(rt, ns); err == nil {
			if mainContent, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
				mainContent.UpdateResourcesOnly(fmt.Sprintf("%s in %s", rt, ns), resources)
				mainContent.SetWatching(true)
			}
		}
		return m, watchTickCmd()

	case WatchTick:
		if !m.watching || m.kubeClient == nil {
			return m, nil
		}
		if resources, err := m.kubeClient.GetResourceListDetailed(m.watchResource, m.watchNamespace); err == nil {
			if mainContent, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
				mainContent.UpdateResourcesOnly(fmt.Sprintf("%s in %s", m.watchResource, m.watchNamespace), resources)
			}
		}
		return m, watchTickCmd()

	case widgets.ShowDescribeRequest:
		if m.kubeClient == nil {
			m.modal.ShowError("No Connection", "Not connected to Kubernetes cluster", "Close")
			m.showModal = true
			return m, nil
		}
		rt := strings.ToLower(msg.Resource.Type)
		switch rt {
		case "pod":
			rt = "pods"
		case "service":
			rt = "services"
		case "deployment":
			rt = "deployments"
		}
		desc, err := m.kubeClient.DescribeResource(rt, msg.Resource.Namespace, msg.Resource.Name)
		if err != nil {
			m.modal.ShowError("Describe Error", fmt.Sprintf("Failed to describe resource:\n%v", err), "Close")
			m.showModal = true
			return m, nil
		}
		title := fmt.Sprintf("Describe: %s/%s (namespace: %s)", rt, msg.Resource.Name, msg.Resource.Namespace)
		m.logsModal.Show(title, desc)
		m.logsModal.SetDimensions(m.width, m.height)
		m.showLogsModal = true
		return m, nil
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

	footer := m.renderFooter()
	bodyWithFooter := lipgloss.JoinVertical(lipgloss.Left, horizontal, footer)
	return bodyWithFooter
}

func (m MainModel) renderFooter() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(m.width)

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

	switch m.focusedWidget {
	case 0:
		hints = append(hints, "j/k: move focus", "enter: choose namespace", "q: quit")
	case 1:
		if arw, ok := m.widgets[1].(*widgets.ApiResourceWidget); ok && arw.IsListActive() {
			hints = append(hints, "j/k: move", "enter: select", "esc: back", "/: search", "q: quit")
		} else {
			hints = append(hints, "j/k: move focus", "enter: open resources", "q: quit")
		}
	case 2:
		if mcw, ok := m.widgets[2].(*widgets.MainContentWidget); ok {
			if mcw.SelectionNameSpace {
				hints = append(hints, "j/k: move", "enter: select namespace", "esc: cancel", "q: quit")
			} else if mcw.IsResourcesActive() {
				hints = append(hints, "j/k, up/down: scroll", "esc: exit")
				hints = append(hints, "ctrl+w: toggle watch")
				if sel := mcw.GetSelectedResource(); sel != nil && sel.Type == "Pod" {
					hints = append(hints, "ctrl+l: view logs")
					hints = append(hints, "ctrl+d: describe resource")
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
