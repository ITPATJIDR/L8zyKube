package components

import (
	"fmt"
	kubetypes "l8zykube/kubernetes"
	"strings"
	"unicode/utf8"

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

const (
	resourceTableVerticalChrome = 4
	defaultResourceTableHeight  = 33
	columnSeparator             = "  "
)

type tableColumn struct {
	title     string
	minWidth  int
	maxWidth  int
	extractor func(kubetypes.ResourceInfo) string
	width     int
}

var placeholderValues = map[string]struct{}{
	"":        {},
	"<none>":  {},
	"unknown": {},
	"n/a":     {},
	"-":       {},
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

	_, rowsForItems := rt.layoutMetrics()
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

	_, rowsForItems := rt.layoutMetrics()
	maxOff := maxInt(len(rt.Resources)-rowsForItems, 0)
	if rt.ScrollOffset > maxOff {
		rt.ScrollOffset = maxOff
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

	_, rowsForItems := rt.layoutMetrics()

	// If selected item is below visible area, scroll down
	if rt.SelectedIndex >= rt.ScrollOffset+rowsForItems {
		rt.ScrollOffset = rt.SelectedIndex - rowsForItems + 1
	}
	maxOff := maxInt(len(rt.Resources)-rowsForItems, 0)
	if rt.ScrollOffset > maxOff {
		rt.ScrollOffset = maxOff
	}
	if rt.ScrollOffset < 0 {
		rt.ScrollOffset = 0
	}
}

func (rt *ResourceTable) ScrollToTop() {
	rt.ScrollOffset = 0
	rt.SelectedIndex = 0
}

func (rt *ResourceTable) ScrollToBottom() {
	_, rowsForItems := rt.layoutMetrics()

	rt.SelectedIndex = len(rt.Resources) - 1
	rt.ScrollOffset = maxInt(len(rt.Resources)-rowsForItems, 0)
}

func (rt *ResourceTable) PageUp() {
	_, rowsForItems := rt.layoutMetrics()

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
	_, rowsForItems := rt.layoutMetrics()

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

	columns := rt.determineColumns(showNamespace)
	if len(columns) == 0 {
		columns = []tableColumn{
			rt.newColumn("NAME", 12, 48, func(r kubetypes.ResourceInfo) string { return r.Name }),
		}
	}
	columns = rt.layoutColumns(columns, contentWidth)

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

	headerCells := make([]string, len(columns))
	for i, col := range columns {
		headerCells[i] = pad(trunc(col.title, col.width), col.width)
	}
	header := lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("240")).
		Bold(true).
		Render(strings.Join(headerCells, columnSeparator))

	// Build rows within available height and apply scrolling
	innerHeight, rowsForItems := rt.layoutMetrics()

	rows := make([]string, 0, innerHeight)
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
		cells := make([]string, len(columns))
		for ci, col := range columns {
			value := strings.TrimSpace(col.extractor(r))
			cells[ci] = pad(trunc(value, col.width), col.width)
		}
		line := strings.Join(cells, columnSeparator)

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

func (rt *ResourceTable) contentHeight() int {
	if rt.Height <= 0 {
		return defaultResourceTableHeight
	}

	h := rt.Height - resourceTableVerticalChrome
	if h < 1 {
		h = 1
	}
	return h
}

func (rt *ResourceTable) layoutMetrics() (int, int) {
	innerHeight := rt.contentHeight()
	availableRows := innerHeight - 2
	if availableRows < 2 {
		availableRows = 2
	}
	rowsForItems := availableRows - 1
	if rowsForItems < 1 {
		rowsForItems = 1
	}
	return innerHeight, rowsForItems
}

func (rt *ResourceTable) newColumn(title string, minWidth, maxWidth int, extractor func(kubetypes.ResourceInfo) string) tableColumn {
	minTitleWidth := utf8.RuneCountInString(title)
	if minWidth < minTitleWidth {
		minWidth = minTitleWidth
	}
	if minWidth < 1 {
		minWidth = 1
	}
	if maxWidth > 0 && maxWidth < minWidth {
		maxWidth = minWidth
	}
	return tableColumn{
		title:     title,
		minWidth:  minWidth,
		maxWidth:  maxWidth,
		extractor: extractor,
	}
}

func (rt *ResourceTable) determineColumns(showNamespace bool) []tableColumn {
	columns := []tableColumn{
		rt.newColumn("NAME", 12, 48, func(r kubetypes.ResourceInfo) string { return r.Name }),
	}

	if showNamespace {
		columns = append(columns, rt.newColumn("NAMESPACE", 12, 24, func(r kubetypes.ResourceInfo) string { return r.Namespace }))
	}

	if rt.hasMeaningfulValue(func(r kubetypes.ResourceInfo) string { return r.Ready }, "0/0") {
		columns = append(columns, rt.newColumn("READY", 5, 12, func(r kubetypes.ResourceInfo) string { return r.Ready }))
	}
	if rt.hasMeaningfulValue(func(r kubetypes.ResourceInfo) string { return r.Status }) {
		columns = append(columns, rt.newColumn("STATUS", 8, 24, func(r kubetypes.ResourceInfo) string { return r.Status }))
	}
	if rt.hasMeaningfulValue(func(r kubetypes.ResourceInfo) string { return r.Restarts }, "0") {
		columns = append(columns, rt.newColumn("RESTARTS", 7, 16, func(r kubetypes.ResourceInfo) string { return r.Restarts }))
	}
	if rt.hasMeaningfulValue(func(r kubetypes.ResourceInfo) string { return r.Age }) {
		columns = append(columns, rt.newColumn("AGE", 6, 16, func(r kubetypes.ResourceInfo) string { return r.Age }))
	}
	if rt.hasMeaningfulValue(func(r kubetypes.ResourceInfo) string { return r.IP }) {
		columns = append(columns, rt.newColumn("IP", 8, 24, func(r kubetypes.ResourceInfo) string { return r.IP }))
	}
	if rt.hasMeaningfulValue(func(r kubetypes.ResourceInfo) string { return r.Node }) {
		columns = append(columns, rt.newColumn("NODE", 8, 24, func(r kubetypes.ResourceInfo) string { return r.Node }))
	}

	return columns
}

func (rt *ResourceTable) hasMeaningfulValue(extractor func(kubetypes.ResourceInfo) string, extraIgnores ...string) bool {
	ignore := make(map[string]struct{}, len(placeholderValues)+len(extraIgnores))
	for k := range placeholderValues {
		ignore[k] = struct{}{}
	}
	for _, v := range extraIgnores {
		key := strings.ToLower(strings.TrimSpace(v))
		if key == "" {
			continue
		}
		ignore[key] = struct{}{}
	}

	for _, res := range rt.Resources {
		val := strings.ToLower(strings.TrimSpace(extractor(res)))
		if val == "" {
			continue
		}
		if _, found := ignore[val]; !found {
			return true
		}
	}
	return false
}

func (rt *ResourceTable) layoutColumns(columns []tableColumn, contentWidth int) []tableColumn {
	if len(columns) == 0 {
		return columns
	}

	for i := range columns {
		maxLen := utf8.RuneCountInString(columns[i].title)
		for _, res := range rt.Resources {
			value := strings.TrimSpace(columns[i].extractor(res))
			length := utf8.RuneCountInString(value)
			if length > maxLen {
				maxLen = length
			}
		}
		width := columns[i].minWidth
		if maxLen > width {
			width = maxLen
		}
		if columns[i].maxWidth > 0 && width > columns[i].maxWidth {
			width = columns[i].maxWidth
		}
		columns[i].width = width
	}

	separatorWidth := len(columnSeparator) * (len(columns) - 1)
	totalOther := separatorWidth
	for i := 1; i < len(columns); i++ {
		totalOther += columns[i].width
	}

	nameCol := &columns[0]
	available := contentWidth - totalOther
	if available < nameCol.minWidth {
		deficit := nameCol.minWidth - available
		columns = shrinkColumns(columns, deficit)
		totalOther = separatorWidth
		for i := 1; i < len(columns); i++ {
			totalOther += columns[i].width
		}
		available = contentWidth - totalOther
	}
	if available < 1 {
		available = 1
	}
	if nameCol.maxWidth > 0 && available > nameCol.maxWidth {
		available = nameCol.maxWidth
	}
	if available < nameCol.minWidth {
		available = nameCol.minWidth
	}
	nameCol.width = available
	return columns
}

func shrinkColumns(columns []tableColumn, deficit int) []tableColumn {
	if deficit <= 0 {
		return columns
	}
	for deficit > 0 {
		madeProgress := false
		for i := len(columns) - 1; i >= 1 && deficit > 0; i-- {
			if columns[i].width > columns[i].minWidth {
				columns[i].width--
				deficit--
				madeProgress = true
			}
		}
		if !madeProgress {
			break
		}
	}
	return columns
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
