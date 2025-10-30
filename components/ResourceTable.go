package components

import (
	"fmt"
	kubetypes "l8zykube/kubernetes"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ResourceTable struct {
	Resources     []kubetypes.ResourceInfo
	Title         string
	ScrollOffset  int
	SelectedIndex int
	Active        bool
	Watching      bool
	Width         int
	Height        int
}

func NewResourceTable() *ResourceTable {
	return &ResourceTable{
		Resources:    []kubetypes.ResourceInfo{},
		ScrollOffset: 0,
		Active:       false,
	}
}

func (rt *ResourceTable) SetResources(title string, resources []kubetypes.ResourceInfo) {
	rt.Title = title
	rt.Resources = resources
	rt.ScrollOffset = 0
	rt.SelectedIndex = 0
	rt.Active = false
}

func (rt *ResourceTable) UpdateResourcesOnly(title string, resources []kubetypes.ResourceInfo) {
	rt.Title = title
	rt.Resources = resources

	if rt.SelectedIndex >= len(rt.Resources) {
		rt.SelectedIndex = maxInt(len(rt.Resources)-1, 0)
	}

	innerHeight := rt.Height
	if innerHeight <= 0 {
		innerHeight = 37
	}
	availableRows := innerHeight - 2
	if availableRows < 2 {
		availableRows = 2
	}
	rowsForItems := availableRows - 1
	if rowsForItems < 1 {
		rowsForItems = 1
	}

	maxOff := maxInt(len(rt.Resources)-rowsForItems, 0)
	if rt.ScrollOffset > maxOff {
		rt.ScrollOffset = maxOff
	}
	if rt.ScrollOffset < 0 {
		rt.ScrollOffset = 0
	}
}

func (rt *ResourceTable) SetActive(active bool) {
	rt.Active = active
}

func (rt *ResourceTable) SetWatching(watching bool) {
	rt.Watching = watching
}

func (rt *ResourceTable) GetSelectedResource() *kubetypes.ResourceInfo {
	if rt.SelectedIndex >= 0 && rt.SelectedIndex < len(rt.Resources) {
		return &rt.Resources[rt.SelectedIndex]
	}
	return nil
}

func (rt *ResourceTable) SetDimensions(width, height int) {
	rt.Width = width
	rt.Height = height
}

func (rt *ResourceTable) ScrollUp() {
	if rt.SelectedIndex > 0 {
		rt.SelectedIndex--
	}

	// Adjust scroll offset to keep selected item visible
	innerHeight := rt.Height
	if innerHeight <= 0 {
		innerHeight = 37
	}
	availableRows := innerHeight - 2
	if availableRows < 2 {
		availableRows = 2
	}
	rowsForItems := availableRows - 1
	if rowsForItems < 1 {
		rowsForItems = 1
	}

	// If selected item is above visible area, scroll up
	if rt.SelectedIndex < rt.ScrollOffset {
		rt.ScrollOffset = rt.SelectedIndex
	}
}

func (rt *ResourceTable) ScrollDown() {
	if rt.SelectedIndex < len(rt.Resources)-1 {
		rt.SelectedIndex++
	}

	// Adjust scroll offset to keep selected item visible
	innerHeight := rt.Height
	if innerHeight <= 0 {
		innerHeight = 37
	}
	availableRows := innerHeight - 2
	if availableRows < 2 {
		availableRows = 2
	}
	rowsForItems := availableRows - 1
	if rowsForItems < 1 {
		rowsForItems = 1
	}

	// If selected item is below visible area, scroll down
	if rt.SelectedIndex >= rt.ScrollOffset+rowsForItems {
		rt.ScrollOffset = rt.SelectedIndex - rowsForItems + 1
	}
}

func (rt *ResourceTable) ScrollToTop() {
	rt.ScrollOffset = 0
	rt.SelectedIndex = 0
}

func (rt *ResourceTable) ScrollToBottom() {
	innerHeight := rt.Height
	if innerHeight <= 0 {
		innerHeight = 37
	}
	availableRows := innerHeight - 2
	if availableRows < 2 {
		availableRows = 2
	}
	rowsForItems := availableRows - 1
	if rowsForItems < 1 {
		rowsForItems = 1
	}

	rt.SelectedIndex = len(rt.Resources) - 1
	rt.ScrollOffset = maxInt(len(rt.Resources)-rowsForItems, 0)
}

func (rt *ResourceTable) PageUp() {
	innerHeight := rt.Height
	if innerHeight <= 0 {
		innerHeight = 37
	}
	availableRows := innerHeight - 2
	if availableRows < 2 {
		availableRows = 2
	}
	rowsForItems := availableRows - 1
	if rowsForItems < 1 {
		rowsForItems = 1
	}

	rt.SelectedIndex -= rowsForItems
	if rt.SelectedIndex < 0 {
		rt.SelectedIndex = 0
	}

	rt.ScrollOffset -= rowsForItems
	if rt.ScrollOffset < 0 {
		rt.ScrollOffset = 0
	}
}

func (rt *ResourceTable) PageDown() {
	innerHeight := rt.Height
	if innerHeight <= 0 {
		innerHeight = 37
	}
	availableRows := innerHeight - 2
	if availableRows < 2 {
		availableRows = 2
	}
	rowsForItems := availableRows - 1
	if rowsForItems < 1 {
		rowsForItems = 1
	}

	rt.SelectedIndex += rowsForItems
	if rt.SelectedIndex >= len(rt.Resources) {
		rt.SelectedIndex = len(rt.Resources) - 1
	}

	rt.ScrollOffset += rowsForItems
	maxOff := maxInt(len(rt.Resources)-rowsForItems, 0)
	if rt.ScrollOffset > maxOff {
		rt.ScrollOffset = maxOff
	}
}

func (rt *ResourceTable) Render() string {
	if len(rt.Resources) == 0 {
		return ""
	}

	// Title
	titlePrefix := "Resources"
	if rt.Watching {
		titlePrefix = "Watch Resources"
	}
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginLeft(2).
		Render(fmt.Sprintf("%s: %s (%d)", titlePrefix, rt.Title, len(rt.Resources)))

	contentWidth := rt.Width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	// Determine if namespace column should be rendered
	showNamespace := false
	if len(rt.Resources) > 0 {
		unique := make(map[string]struct{})
		for _, res := range rt.Resources {
			ns := strings.TrimSpace(res.Namespace)
			if ns != "" {
				unique[strings.ToLower(ns)] = struct{}{}
			}
			if len(unique) > 1 {
				showNamespace = true
				break
			}
		}
		if !showNamespace {
			lowerTitle := strings.ToLower(rt.Title)
			if strings.Contains(lowerTitle, "all namespace") {
				showNamespace = true
			}
		}
	}

	// Column widths
	readyW, statusW, ageW, ipW := 7, 12, 12, 15
	nsW := 0
	spaces := 4
	if showNamespace {
		nsW = 15
		spaces++
	}
	nameW := contentWidth - (readyW + statusW + ageW + ipW + nsW + spaces)
	if nameW < 10 {
		delta := 10 - nameW
		if showNamespace {
			reduceNS := minInt(delta, nsW-8)
			nsW -= reduceNS
			delta -= reduceNS
		}
		reduceIP := minInt(delta, ipW-8)
		ipW -= reduceIP
		delta -= reduceIP
		if delta > 0 {
			reduceStatus := minInt(delta, statusW-8)
			statusW -= reduceStatus
			delta -= reduceStatus
		}
		if delta > 0 {
			reduceAge := minInt(delta, ageW-6)
			ageW -= reduceAge
			delta -= reduceAge
		}
		nameW = 10
	}

	pad := func(s string, w int) string { return fmt.Sprintf("%-*s", w, s) }
	trunc := func(s string, w int) string {
		if w <= 0 {
			return ""
		}
		r := []rune(s)
		if len(r) <= w {
			return s
		}
		if w == 1 {
			return "…"
		}
		return string(r[:w-1]) + "…"
	}

	headerColumns := []string{pad(trunc("NAME", nameW), nameW)}
	if showNamespace {
		headerColumns = append(headerColumns, pad(trunc("NAMESPACE", nsW), nsW))
	}
	headerColumns = append(headerColumns,
		pad(trunc("READY", readyW), readyW),
		pad(trunc("STATUS", statusW), statusW),
		pad(trunc("AGE", ageW), ageW),
		pad(trunc("IP", ipW), ipW),
	)
	header := lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("240")).
		Bold(true).
		Render(strings.Join(headerColumns, " "))

	// Build rows within available height and apply scrolling
	innerHeight := rt.Height
	if innerHeight <= 0 {
		innerHeight = 37
	}
	availableRows := innerHeight - 2 // content rows inside padding
	if availableRows < 2 {
		availableRows = 2
	}
	rowsForItems := availableRows - 1 // leave one row for footer
	if rowsForItems < 1 {
		rowsForItems = 1
	}

	rows := make([]string, 0, availableRows)
	rows = append(rows, title)
	rows = append(rows, header)
	rowStyle := lipgloss.NewStyle().PaddingLeft(2)

	// Clamp scroll offset
	maxOff := maxInt(len(rt.Resources)-rowsForItems, 0)
	if rt.ScrollOffset > maxOff {
		rt.ScrollOffset = maxOff
	}
	if rt.ScrollOffset < 0 {
		rt.ScrollOffset = 0
	}

	start := rt.ScrollOffset
	end := start + rowsForItems
	if end > len(rt.Resources) {
		end = len(rt.Resources)
	}
	for i := start; i < end; i++ {
		r := rt.Resources[i]
		lineColumns := []string{pad(trunc(r.Name, nameW), nameW)}
		if showNamespace {
			ns := strings.TrimSpace(r.Namespace)
			if ns == "" {
				ns = "<cluster>"
			}
			lineColumns = append(lineColumns, pad(trunc(ns, nsW), nsW))
		}
		lineColumns = append(lineColumns,
			pad(trunc(r.Ready, readyW), readyW),
			pad(trunc(r.Status, statusW), statusW),
			pad(trunc(r.Age, ageW), ageW),
			pad(trunc(r.IP, ipW), ipW),
		)
		line := strings.Join(lineColumns, " ")

		// Highlight selected row
		if i == rt.SelectedIndex && rt.Active {
			selectedStyle := lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("205")).
				Bold(true).
				Background(lipgloss.Color("236"))
			rows = append(rows, selectedStyle.Render(line))
		} else {
			rows = append(rows, rowStyle.Render(line))
		}
	}

	// Footer indicator
	footerHint := ""
	if rt.Active {
		footerHint = "  (Esc exit, j/k scroll)"
	}
	footer := lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("%d-%d of %d%s", start+1, end, len(rt.Resources), footerHint))

	rows = append(rows, footer)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
