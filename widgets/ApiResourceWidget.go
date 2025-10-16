package widgets

import (
	"fmt"

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
	selectedApiResource string
	selectedIndex       int
	scrollOffset        int
	listActive          bool
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
			} else if len(a.ApiResourceList) > 0 && a.selectedIndex >= 0 && a.selectedIndex < len(a.ApiResourceList) {
				a.selectedApiResource = a.ApiResourceList[a.selectedIndex]
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
				if a.selectedIndex < len(a.ApiResourceList)-1 {
					a.selectedIndex++
				}
			case "home", "g":
				a.selectedIndex = 0
			case "end", "G":
				if len(a.ApiResourceList) > 0 {
					a.selectedIndex = len(a.ApiResourceList) - 1
				}
			}
		}

		a.ensureSelectionVisible()
	}

	return a, nil
}

func (a *ApiResourceWidget) SetApiResourceList(resources []string) {
	a.ApiResourceList = resources
	if a.selectedIndex >= len(a.ApiResourceList) {
		a.selectedIndex = maxInt(len(a.ApiResourceList)-1, 0)
	}
	a.ensureSelectionVisible()
}

func (a *ApiResourceWidget) GetSelectedApiResource() string {
	if len(a.ApiResourceList) == 0 {
		return a.selectedApiResource
	}
	if a.selectedIndex >= 0 && a.selectedIndex < len(a.ApiResourceList) {
		return a.ApiResourceList[a.selectedIndex]
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

		listHeight := a.innerHeight() - 1
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
		for _, idx := range visibleItems {
			name := a.ApiResourceList[idx]
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

// Helpers
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
	if len(a.ApiResourceList) == 0 {
		a.scrollOffset = 0
		a.selectedIndex = 0
		return
	}

	h := a.innerHeight() - 1
	if h < 2 {
		h = 2
	}

	visibleItemCount := maxInt(h/2, 1)

	if a.selectedIndex < 0 {
		a.selectedIndex = 0
	}
	if a.selectedIndex > len(a.ApiResourceList)-1 {
		a.selectedIndex = len(a.ApiResourceList) - 1
	}

	if a.selectedIndex < a.scrollOffset {
		a.scrollOffset = a.selectedIndex
	}

	bottomIndex := a.scrollOffset + visibleItemCount - 1
	if a.selectedIndex > bottomIndex {
		a.scrollOffset = a.selectedIndex - visibleItemCount + 1
	}

	maxOffset := maxInt(len(a.ApiResourceList)-visibleItemCount, 0)
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
	if end > len(a.ApiResourceList) {
		end = len(a.ApiResourceList)
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
