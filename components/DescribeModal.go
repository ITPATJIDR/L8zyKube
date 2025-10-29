package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type DescribeMode int

const (
	DescribeModeRead DescribeMode = iota
	DescribeModeWrite
)

type DescribeModal struct {
	Width     int
	Height    int
	Title     string
	Content   string
	Visible   bool
	scrollPos int
	lines     []string
	mode      DescribeMode

	resourceType string
	resourceName string
	namespace    string
}

func NewDescribeModal() *DescribeModal {
	return &DescribeModal{
		Visible: false,
		mode:    DescribeModeRead,
	}
}

func (dm *DescribeModal) SetDimensions(width, height int) {
	dm.Width = width
	dm.Height = height
}

func (dm *DescribeModal) Show(title, content, resourceType, namespace, name string) {
	dm.Title = title
	dm.Content = content
	dm.lines = strings.Split(content, "\n")
	dm.Visible = true
	dm.scrollPos = 0
	dm.mode = DescribeModeRead
	dm.resourceType = resourceType
	dm.resourceName = name
	dm.namespace = namespace
}

func (dm *DescribeModal) Hide() {
	dm.Visible = false
}

func (dm *DescribeModal) Mode() DescribeMode {
	return dm.mode
}

func (dm *DescribeModal) SetMode(mode DescribeMode) {
	dm.mode = mode
}

func (dm *DescribeModal) UpdateContent(content string) {
	dm.Content = content
	dm.lines = strings.Split(content, "\n")
	dm.scrollPos = 0
	dm.mode = DescribeModeRead
}

func (dm *DescribeModal) TargetInfo() (string, string, string) {
	return dm.resourceType, dm.namespace, dm.resourceName
}

func (dm *DescribeModal) CanEdit() bool {
	return strings.TrimSpace(dm.resourceType) != "" && strings.TrimSpace(dm.resourceName) != ""
}

func (dm *DescribeModal) EditCommandArgs() []string {
	if !dm.CanEdit() {
		return nil
	}
	args := []string{"kubectl", "edit", dm.resourceType, dm.resourceName}
	if ns := strings.TrimSpace(dm.namespace); ns != "" {
		args = append(args, "-n", ns)
	}
	return args
}

func (dm *DescribeModal) ScrollUp() {
	if dm.scrollPos > 0 {
		dm.scrollPos--
	}
}

func (dm *DescribeModal) ScrollDown() {
	maxScroll := len(dm.lines) - dm.visibleLineCount()
	if maxScroll < 0 {
		maxScroll = 0
	}
	if dm.scrollPos < maxScroll {
		dm.scrollPos++
	}
}

func (dm *DescribeModal) PageUp() {
	lines := dm.visibleLineCount()
	dm.scrollPos -= lines
	if dm.scrollPos < 0 {
		dm.scrollPos = 0
	}
}

func (dm *DescribeModal) PageDown() {
	lines := dm.visibleLineCount()
	maxScroll := len(dm.lines) - lines
	if maxScroll < 0 {
		maxScroll = 0
	}
	dm.scrollPos += lines
	if dm.scrollPos > maxScroll {
		dm.scrollPos = maxScroll
	}
}

func (dm *DescribeModal) ScrollToTop() {
	dm.scrollPos = 0
}

func (dm *DescribeModal) ScrollToBottom() {
	maxScroll := len(dm.lines) - dm.visibleLineCount()
	if maxScroll < 0 {
		maxScroll = 0
	}
	dm.scrollPos = maxScroll
}

func (dm *DescribeModal) visibleLineCount() int {
	height := dm.Height - 10
	if height < 1 {
		height = 1
	}
	return height
}

func (dm *DescribeModal) Render() string {
	if !dm.Visible {
		return ""
	}

	modalWidth := dm.Width - 10
	if modalWidth < 80 {
		modalWidth = 80
	}
	modalHeight := dm.Height - 4
	if modalHeight < 20 {
		modalHeight = 20
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("34")).
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight - 3)

	titleLabel := "Read"
	if dm.mode == DescribeModeWrite {
		titleLabel = "Write"
	}
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("34")).
		Bold(true).
		Align(lipgloss.Left).
		Width(modalWidth-4).
		Margin(0, 0, 1, 0)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(modalWidth - 4).
		Height(modalHeight - 8).
		MaxHeight(modalHeight - 4)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(modalWidth-4).
		Margin(1, 0, 0, 0).
		Italic(true)

	modeTitle := fmt.Sprintf("%s [%s mode]", dm.Title, titleLabel)
	title := titleStyle.Render(modeTitle)

	var body string
	var scrollInfo string
	var instruction string

	if dm.mode == DescribeModeWrite {
		editCommand := strings.Join(dm.EditCommandArgs(), " ")
		if editCommand == "" {
			editCommand = "kubectl edit <resource> <name>"
		}
		message := []string{
			"Launching editor...",
			fmt.Sprintf("Command: %s", editCommand),
			"After closing the editor, the describe view will refresh automatically.",
		}
		body = contentStyle.Render(strings.Join(message, "\n"))
		instruction = instructionStyle.Render("esc/q: close")
	} else {
		visibleLines := dm.visibleLineCount()
		if visibleLines < 1 {
			visibleLines = 1
		}

		start := dm.scrollPos
		if start < 0 {
			start = 0
		}
		if start >= len(dm.lines) {
			start = len(dm.lines) - 1
			if start < 0 {
				start = 0
			}
		}
		end := start + visibleLines
		if end > len(dm.lines) {
			end = len(dm.lines)
		}
		if start > end {
			start = end
		}

		visible := "No description available"
		if len(dm.lines) > 0 {
			visible = strings.Join(dm.lines[start:end], "\n")
		}

		body = contentStyle.Render(visible)
		if len(dm.lines) > visibleLines {
			scrollInfo = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Align(lipgloss.Right).
				Width(modalWidth - 4).
				Render(fmt.Sprintf("Lines %d-%d of %d", start+1, end, len(dm.lines)))
		}
		instruction = instructionStyle.Render("↑/↓: scroll | PgUp/PgDown: page | g/G: top/bottom | ctrl+e: edit | esc/q: close")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		body,
		scrollInfo,
		instruction,
	)

	return modalStyle.Render(content)
}
