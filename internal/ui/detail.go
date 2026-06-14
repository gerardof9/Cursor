package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"db-log-explorer/internal/events"
)

var detailTitleStyle = lipgloss.NewStyle().Bold(true)
var detailNoteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

// RenderDetail draws the detail pane.
func RenderDetail(width, height int, loading bool, preview *events.EventSummary, detail *events.EventDetail, filteredEmpty bool) string {
	if filteredEmpty {
		return lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("240")).Render("(detail cleared — no filter matches)")
	}
	if detail != nil {
		return lipgloss.NewStyle().Width(width).MaxHeight(height).Render(renderFullDetail(detail))
	}
	if preview != nil {
		var b strings.Builder
		writeSummaryHeader(&b, *preview)
		if loading {
			b.WriteString("\n")
			b.WriteString(detailNoteStyle.Render("Loading row data..."))
		}
		return lipgloss.NewStyle().Width(width).MaxHeight(height).Render(b.String())
	}
	if loading {
		return lipgloss.NewStyle().Width(width).Render("Loading...")
	}
	return lipgloss.NewStyle().Width(width).Foreground(lipgloss.Color("240")).Render("Select an event")
}

func renderFullDetail(detail *events.EventDetail) string {
	var b strings.Builder
	s := detail.Summary
	writeSummaryHeader(&b, s)
	if detail.SQL != "" {
		b.WriteString("\nSQL:\n")
		b.WriteString(detail.SQL)
		b.WriteString("\n")
	}
	for i, row := range detail.RowValues {
		b.WriteString(fmt.Sprintf("\nRow %d:\n", i+1))
		if len(row.Before) > 0 {
			b.WriteString("  Before: " + strings.Join(row.Before, ", ") + "\n")
		}
		if len(row.After) > 0 {
			b.WriteString("  After:  " + strings.Join(row.After, ", ") + "\n")
		}
	}
	for _, note := range detail.Notes {
		b.WriteString(detailNoteStyle.Render(note) + "\n")
	}
	if !detail.Complete {
		b.WriteString(detailNoteStyle.Render("Partial metadata only") + "\n")
	}
	return b.String()
}

func writeSummaryHeader(b *strings.Builder, s events.EventSummary) {
	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("%s %s.%s", s.Operation, s.Schema, s.Table)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Time: %s\n", s.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("Format: %s\n", s.Format))
	if s.TxHint != "" {
		b.WriteString(fmt.Sprintf("Transaction: %s\n", s.TxHint))
	}
}
