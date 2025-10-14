package components

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

type NamespaceSelector struct {
	NamespaceList     []string
	list              list.Model
	selectedNamespace string
	Width             int
	Height            int
}

func NewNamespaceSelector() *NamespaceSelector {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Namespace"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.MarginLeft(2)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.MarginLeft(2)

	return &NamespaceSelector{
		list: l,
	}
}

func (ns *NamespaceSelector) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	ns.list, cmd = ns.list.Update(msg)

	// Update our selected namespace from the list
	if selectedItem := ns.list.SelectedItem(); selectedItem != nil {
		if item, ok := selectedItem.(NamespaceItem); ok {
			ns.selectedNamespace = item.title
		}
	}

	return cmd
}

func (ns *NamespaceSelector) SetNamespaceList(namespaces []string) {
	ns.NamespaceList = namespaces

	items := make([]list.Item, len(namespaces))
	for i, ns := range namespaces {
		items[i] = NamespaceItem{
			title: ns,
			desc:  fmt.Sprintf("Namespace: %s", ns),
		}
	}

	ns.list.SetItems(items)
}

func (ns *NamespaceSelector) GetSelectedNamespace() string {
	// Get the currently selected item from the list
	if selectedItem := ns.list.SelectedItem(); selectedItem != nil {
		if item, ok := selectedItem.(NamespaceItem); ok {
			return item.title
		}
	}
	return ns.selectedNamespace
}

func (ns *NamespaceSelector) SetDimensions(width, height int) {
	ns.Width = width
	ns.Height = height
	ns.list.SetSize(width-4, height-4)
}

func (ns *NamespaceSelector) Render() string {
	return ns.list.View()
}
