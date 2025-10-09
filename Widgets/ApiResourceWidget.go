package widgets

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ApiResourceItem represents an item in the API resource list
type ApiResourceItem struct {
	title, desc string
}

func (i ApiResourceItem) Title() string       { return i.title }
func (i ApiResourceItem) Description() string { return i.desc }
func (i ApiResourceItem) FilterValue() string { return i.title }

type ApiResourceWidget struct {
	BaseWidget
	SelectedNameSpace   string
	SelectionNameSpace  bool
	ApiResourceList     []string
	list                list.Model
	selectedApiResource string
}

func NewApiResourceWidget() *ApiResourceWidget {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "API Resources"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	// Enhanced styling
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginLeft(2)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.
		MarginLeft(2).
		Foreground(lipgloss.Color("240"))
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.
		MarginLeft(2).
		Foreground(lipgloss.Color("240"))

	// Style the list items using the delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Background(lipgloss.Color("236"))
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236"))
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("87"))
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("240"))

	l.SetDelegate(delegate)

	return &ApiResourceWidget{
		BaseWidget: BaseWidget{
			focused: false,
		},
		list: l,
	}
}

func (a *ApiResourceWidget) Update(msg tea.Msg) (Widget, tea.Cmd) {
	if !a.focused {
		return a, nil
	}

	var cmd tea.Cmd

	// Handle list interactions when focused
	a.list, cmd = a.list.Update(msg)

	// Update our selected API resource from the list
	if selectedItem := a.list.SelectedItem(); selectedItem != nil {
		if item, ok := selectedItem.(ApiResourceItem); ok {
			a.selectedApiResource = item.title
		}
	}

	return a, cmd
}

func (a *ApiResourceWidget) SetApiResourceList(resources []string) {
	a.ApiResourceList = resources

	items := make([]list.Item, len(resources))
	for i, resource := range resources {
		items[i] = ApiResourceItem{
			title: resource,
			desc:  fmt.Sprintf("Resource: %s", resource),
		}
	}

	a.list.SetItems(items)
}

func (a *ApiResourceWidget) GetSelectedApiResource() string {
	// Get the currently selected item from the list
	if selectedItem := a.list.SelectedItem(); selectedItem != nil {
		if item, ok := selectedItem.(ApiResourceItem); ok {
			return item.title
		}
	}
	return a.selectedApiResource
}

func (a *ApiResourceWidget) SetDimensions(width, height int) {
	a.BaseWidget.SetDimensions(width, height)
	a.list.SetSize(width-4, height-4)
}

func (a *ApiResourceWidget) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)

	if a.width > 0 {
		style = style.Width(a.width)
	} else {
		style = style.Width(30)
	}

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

	var content string
	if len(a.ApiResourceList) > 0 {
		// Show the API resource list
		content = a.list.View()
	} else {
		// Show placeholder when no resources are loaded
		placeholderText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Center).
			Width(a.width - 4).
			Render("No API resources loaded\nSelect a namespace first")

		content = placeholderText
	}

	return style.Render(content)
}
