package widgets

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ApiResourceItem struct {
	title string
}

type ApiResourceWidget struct {
	BaseWidget
	SelectedNameSpace   string
	SelectionNameSpace  bool
	ApiResourceList     []string
	filteredList        []string
	selectedApiResource string
	selectedIndex       int
	scrollOffset        int
	listActive          bool
	searchActive        bool
	searchQuery         string
}

func NewApiResourceWidget() *ApiResourceWidget {
	return &ApiResourceWidget{
		BaseWidget: BaseWidget{
			focused: false,
		},
		selectedIndex:   0,
		scrollOffset:    0,
		ApiResourceList: []string{},
		listActive:      false,
		searchActive:    false,
	}
}

func (a *ApiResourceWidget) Update(msg tea.Msg) (Widget, tea.Cmd) {
	if !a.focused {
		return a, nil
	}

	switch m := msg.(type) {
	case tea.KeyMsg:
		key := m.String()
		switch key {
		case "enter":
			if !a.listActive {
				a.listActive = true
			} else if len(a.filteredList) > 0 && a.selectedIndex >= 0 && a.selectedIndex < len(a.filteredList) {
				a.selectedApiResource = a.filteredList[a.selectedIndex]
			}
		case tea.KeyEscape.String():
			a.listActive = false
		}

		if a.listActive {
			switch key {
			case "up", "k":
				if a.selectedIndex > 0 {
					a.selectedIndex--
				}
			case "down", "j":
				if a.selectedIndex < len(a.filteredList)-1 {
					a.selectedIndex++
				}
			case "home", "g":
				a.selectedIndex = 0
			case "end", "G":
				if len(a.filteredList) > 0 {
					a.selectedIndex = len(a.filteredList) - 1
				}
			case "/":
				a.searchActive = true
				a.selectedIndex = 0
			}
		}

		if a.searchActive {
			switch key {
			case "esc":
				a.searchActive = false
				a.searchQuery = ""
				a.updateFilteredList()
			case "backspace", tea.KeyBackspace.String():
				if len(a.searchQuery) > 0 {
					a.searchQuery = a.searchQuery[:len(a.searchQuery)-1]
					a.updateFilteredList()
				}
			case "up", "k":
				if a.selectedIndex > 0 {
					// Purpose: for increase selected index by 1 cause duplicate key
					a.selectedIndex += 1
					a.selectedIndex--
				}
			case "down", "j":
				if a.selectedIndex < len(a.filteredList)-1 {
					// Purpose: for decrease selected index by 1 cause duplicate key
					a.selectedIndex -= 1
					a.selectedIndex++
				}
			default:
				if m.Type == tea.KeyRunes {
					a.searchQuery += m.String()
					a.updateFilteredList()
				}
			}
		}

		a.ensureSelectionVisible()
	}

	return a, nil
}

func (a *ApiResourceWidget) SetApiResourceList(resources []string) {
	a.ApiResourceList = resources
	a.updateFilteredList()
	if a.selectedIndex >= len(a.filteredList) {
		a.selectedIndex = maxInt(len(a.filteredList)-1, 0)
	}
	a.ensureSelectionVisible()
}

func (a *ApiResourceWidget) updateFilteredList() {
	if a.searchQuery == "" {
		a.filteredList = a.ApiResourceList
	} else {
		query := strings.ToLower(a.searchQuery)
		a.filteredList = []string{}
		for _, resource := range a.ApiResourceList {
			if strings.Contains(strings.ToLower(resource), query) {
				a.filteredList = append(a.filteredList, resource)
			}
		}
	}

	// Reset selected index if it's out of bounds
	if a.selectedIndex >= len(a.filteredList) {
		a.selectedIndex = maxInt(len(a.filteredList)-1, 0)
	}
	a.ensureSelectionVisible()
}

func (a *ApiResourceWidget) GetSelectedApiResource() string {
	if len(a.filteredList) == 0 {
		return a.selectedApiResource
	}
	if a.selectedIndex >= 0 && a.selectedIndex < len(a.filteredList) {
		return a.filteredList[a.selectedIndex]
	}
	return a.selectedApiResource
}

func (a *ApiResourceWidget) SetDimensions(width, height int) {
	a.BaseWidget.SetDimensions(width, height)
	a.ensureSelectionVisible()
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
	if len(a.ApiResourceList) == 0 {
		placeholderText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Center).
			Width(a.innerContentWidth()).
			Render("No API resources loaded\nSelect a namespace first")

		content = placeholderText
	} else {
		title := lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			MarginLeft(2).
			Render("API Resources")

		// Add search bar if search is active
		searchBar := ""
		if a.searchActive {
			searchBar = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true).
				MarginLeft(2).
				Render(fmt.Sprintf("Search: %s_", a.searchQuery))
		}

		listHeight := a.innerHeight() - 1
		if a.searchActive {
			listHeight -= 1 // Reserve space for search bar
		}
		if listHeight < 1 {
			listHeight = 1
		}

		visibleItems := a.visibleItems(listHeight)

		contentWidth := a.innerContentWidth()
		textWidth := contentWidth - 2
		if textWidth < 1 {
			textWidth = 1
		}

		normalStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("87"))
		selectedStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Background(lipgloss.Color("236"))
		descStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("240"))

		lines := make([]string, 0, len(visibleItems)*2+1)
		lines = append(lines, title)
		if a.searchActive && searchBar != "" {
			lines = append(lines, searchBar)
		}
		for _, idx := range visibleItems {
			name := a.filteredList[idx]
			nameLine := truncateWithEllipsis(name, textWidth)
			descText := fmt.Sprintf("Resource: %s", name)
			descLine := truncateWithEllipsis(descText, textWidth)
			if idx == a.selectedIndex && a.listActive {
				lines = append(lines, selectedStyle.Render(nameLine))
				lines = append(lines, descStyle.Render(descLine))
			} else {
				lines = append(lines, normalStyle.Render(nameLine))
				lines = append(lines, descStyle.Render(descLine))
			}
		}

		// If there's extra vertical space, pad it so the border stays consistent
		currentHeight := len(lines)
		for currentHeight < a.innerHeight() {
			lines = append(lines, "")
			currentHeight++
		}

		content = lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	return style.Render(content)
}

func (a *ApiResourceWidget) IsListActive() bool {
	return a.listActive
}

func (a *ApiResourceWidget) innerHeight() int {
	h := a.height
	if h <= 0 {
		h = 32
	}
	return maxInt(h-2, 1)
}

func (a *ApiResourceWidget) innerContentWidth() int {
	w := a.width
	if w <= 0 {
		w = 30
	}
	return maxInt(w-4, 1)
}

func (a *ApiResourceWidget) pageSize() int {
	h := a.innerHeight() - 1
	if h < 2 {
		return 1
	}
	itemsFit := h / 2
	if itemsFit < 1 {
		itemsFit = 1
	}
	return itemsFit
}

func (a *ApiResourceWidget) ensureSelectionVisible() {
	if len(a.filteredList) == 0 {
		a.scrollOffset = 0
		a.selectedIndex = 0
		return
	}

	h := a.innerHeight() - 1
	if a.searchActive {
		h -= 1 // Reserve space for search bar
	}
	if h < 2 {
		h = 2
	}

	visibleItemCount := maxInt(h/2, 1)

	if a.selectedIndex < 0 {
		a.selectedIndex = 0
	}
	if a.selectedIndex > len(a.filteredList)-1 {
		a.selectedIndex = len(a.filteredList) - 1
	}

	if a.selectedIndex < a.scrollOffset {
		a.scrollOffset = a.selectedIndex
	}

	bottomIndex := a.scrollOffset + visibleItemCount - 1
	if a.selectedIndex > bottomIndex {
		a.scrollOffset = a.selectedIndex - visibleItemCount + 1
	}

	maxOffset := maxInt(len(a.filteredList)-visibleItemCount, 0)
	if a.scrollOffset > maxOffset {
		a.scrollOffset = maxOffset
	}
	if a.scrollOffset < 0 {
		a.scrollOffset = 0
	}
}

func (a *ApiResourceWidget) visibleItems(visibleRows int) []int {
	rowsForItems := maxInt(visibleRows-1, 1)
	itemSlots := maxInt(rowsForItems/2, 1)
	start := a.scrollOffset
	end := start + itemSlots
	if end > len(a.filteredList) {
		end = len(a.filteredList)
	}
	idxs := make([]int, 0, maxInt(end-start, 0))
	for i := start; i < end; i++ {
		idxs = append(idxs, i)
	}
	return idxs
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncateWithEllipsis(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxWidth {
		return s
	}
	if maxWidth == 1 {
		return "…"
	}
	return string(r[:maxWidth-1]) + "…"
}
