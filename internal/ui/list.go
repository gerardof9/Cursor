package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"db-log-explorer/internal/events"
)

var (
	listHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	listRowStyle    = lipgloss.NewStyle()
	listSelStyle    = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("236")).Foreground(lipgloss.Color("230"))
	listDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// RenderList draws the event list pane.
func RenderList(width, height int, items []events.EventSummary, selected, listOffset int, emptyMsg string) string {
	if emptyMsg != "" {
		return listDimStyle.Width(width).Render(emptyMsg)
	}
	if height < 2 {
		return ""
	}

	header := fmt.Sprintf("%-19s %-7s %-10s %-12s %-8s", "TIME", "OP", "SCHEMA", "TABLE", "SRC")
	lines := []string{listHeaderStyle.Render(truncate(header, width))}

	maxRows := height - 1
	if maxRows < 1 {
		maxRows = 1
	}
	start := listOffset
	if start < 0 {
		start = 0
	}
	if start > len(items) {
		start = len(items)
	}
	end := start + maxRows
	if end > len(items) {
		end = len(items)
	}

	for i := start; i < end; i++ {
		line := formatEventRow(items[i])
		if i == selected {
			lines = append(lines, listSelStyle.Render(truncate(line, width)))
		} else {
			lines = append(lines, listRowStyle.Render(truncate(line, width)))
		}
	}

	return strings.Join(lines, "\n")
}

func formatEventRow(ev events.EventSummary) string {
	ts := ev.Timestamp.Format("2006-01-02 15:04:05")
	src := filepath.Base(ev.SourcePath)
	return fmt.Sprintf("%-19s %-7s %-10s %-12s %-8s", ts, ev.Operation, ev.Schema, ev.Table, src)
}

func truncate(s string, width int) string {
	if width <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return string(r[:width])
	}
	return string(r[:width-1]) + "…"
}
