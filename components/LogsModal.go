package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type LogsModal struct {
	Width     int
	Height    int
	Title     string
	Logs      string
	Visible   bool
	scrollPos int
	logLines  []string
}

func NewLogsModal() *LogsModal {
	return &LogsModal{
		Visible:   false,
		scrollPos: 0,
	}
}

func (lm *LogsModal) SetDimensions(width, height int) {
	lm.Width = width
	lm.Height = height
}

func (lm *LogsModal) Show(title, logs string) {
	lm.Title = title
	lm.Logs = logs
	lm.logLines = strings.Split(logs, "\n")
	lm.Visible = true
	lm.scrollPos = 0

	lm.ScrollToBottom()
}

func (lm *LogsModal) Hide() {
	lm.Visible = false
}

func (lm *LogsModal) ScrollUp() {
	if lm.scrollPos > 0 {
		lm.scrollPos--
	}
}

func (lm *LogsModal) ScrollDown() {
	maxScroll := len(lm.logLines) - lm.getVisibleLines()
	if maxScroll < 0 {
		maxScroll = 0
	}
	if lm.scrollPos < maxScroll {
		lm.scrollPos++
	}
}

func (lm *LogsModal) PageUp() {
	visibleLines := lm.getVisibleLines()
	lm.scrollPos -= visibleLines
	if lm.scrollPos < 0 {
		lm.scrollPos = 0
	}
}

func (lm *LogsModal) PageDown() {
	visibleLines := lm.getVisibleLines()
	maxScroll := len(lm.logLines) - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	lm.scrollPos += visibleLines
	if lm.scrollPos > maxScroll {
		lm.scrollPos = maxScroll
	}
}

func (lm *LogsModal) ScrollToTop() {
	lm.scrollPos = 0
}

func (lm *LogsModal) ScrollToBottom() {
	maxScroll := len(lm.logLines) - lm.getVisibleLines()
	if maxScroll < 0 {
		maxScroll = 0
	}
	lm.scrollPos = maxScroll
}

func (lm *LogsModal) getVisibleLines() int {
	visibleLines := lm.Height - 10
	if visibleLines < 1 {
		visibleLines = 1
	}
	return visibleLines
}

func (lm *LogsModal) Render() string {
	if !lm.Visible {
		return ""
	}

	// Calculate modal dimensions
	modalWidth := lm.Width - 10
	if modalWidth < 80 {
		modalWidth = 80
	}
	modalHeight := lm.Height - 4
	if modalHeight < 20 {
		modalHeight = 20
	}

	// Create the modal border
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")). // Blue border
		Padding(1, 2).
		Width(modalWidth).
		Height(modalHeight - 3)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")). // Blue text
		Bold(true).
		Align(lipgloss.Left).
		Width(modalWidth-4).
		Margin(0, 0, 1, 0)

	// Logs content style
	logsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(modalWidth - 4).
		Height(modalHeight - 8).
		MaxHeight(modalHeight - 4)

	// Instruction style
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(modalWidth-4).
		Margin(1, 0, 0, 0).
		Italic(true)

	title := titleStyle.Render(lm.Title)

	// Get visible portion of logs
	visibleLines := lm.getVisibleLines()
	if visibleLines < 1 {
		visibleLines = 1
	}

	startLine := lm.scrollPos
	if startLine < 0 {
		startLine = 0
	}
	if startLine >= len(lm.logLines) {
		startLine = len(lm.logLines) - 1
		if startLine < 0 {
			startLine = 0
		}
	}

	endLine := startLine + visibleLines
	if endLine > len(lm.logLines) {
		endLine = len(lm.logLines)
	}

	if startLine > endLine {
		startLine = endLine
	}

	var visibleLogs string
	if len(lm.logLines) > 0 && startLine < len(lm.logLines) && endLine <= len(lm.logLines) && startLine <= endLine {
		// Show all log lines in the visible range
		visibleLogs = strings.Join(lm.logLines[startLine:endLine], "\n")
	} else {
		visibleLogs = "No logs available"
	}

	logsContent := logsStyle.Render(visibleLogs)

	// Show scroll position
	scrollInfo := ""
	if len(lm.logLines) > visibleLines {
		scrollInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Right).
			Width(modalWidth - 4).
			Render(fmt.Sprintf("Lines %d-%d of %d", startLine+1, endLine, len(lm.logLines)))
	}

	instruction := instructionStyle.Render("↑/↓: Scroll | PgUp/PgDown: Page | Home/End: Top/Bottom | q/ESC: Close")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		logsContent,
		scrollInfo,
		instruction,
	)

	return modalStyle.Render(content)
}
