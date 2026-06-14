package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"db-log-explorer/internal/events"
	"db-log-explorer/internal/filters"
)

// FilterField identifies focused filter editor field.
type FilterField int

const (
	FieldOps FilterField = iota
	FieldSchema
	FieldTable
)

// FilterModel is the filter editor modal.
type FilterModel struct {
	opsInput    textinput.Model
	schemaInput textinput.Model
	tableInput  textinput.Model
	field       FilterField
}

// NewFilterModel creates filter inputs.
func NewFilterModel() FilterModel {
	mk := func(prompt, ph string) textinput.Model {
		ti := textinput.New()
		ti.Prompt = prompt
		ti.Placeholder = ph
		ti.Width = 40
		return ti
	}
	m := FilterModel{
		opsInput:    mk("Ops: ", "INSERT,UPDATE,DELETE,DDL"),
		schemaInput: mk("Schema: ", "mydb"),
		tableInput:  mk("Table: ", "orders or mydb.orders"),
		field:       FieldOps,
	}
	m.opsInput.Focus()
	return m
}

func (m *FilterModel) activeInput() *textinput.Model {
	switch m.field {
	case FieldSchema:
		return &m.schemaInput
	case FieldTable:
		return &m.tableInput
	default:
		return &m.opsInput
	}
}

// CycleField advances focus (Tab).
func (m *FilterModel) CycleField() {
	m.blurAll()
	m.field = (m.field + 1) % 3
	m.activeInput().Focus()
}

func (m *FilterModel) blurAll() {
	m.opsInput.Blur()
	m.schemaInput.Blur()
	m.tableInput.Blur()
}

// ToCriteria builds filter criteria from inputs.
func (m FilterModel) ToCriteria() filters.Criteria {
	c := filters.Criteria{
		Schema: strings.TrimSpace(m.schemaInput.Value()),
		Table:  strings.TrimSpace(m.tableInput.Value()),
	}
	if ops := strings.TrimSpace(m.opsInput.Value()); ops != "" {
		for _, part := range strings.Split(ops, ",") {
			part = strings.TrimSpace(strings.ToUpper(part))
			switch events.Operation(part) {
			case events.OpInsert, events.OpUpdate, events.OpDelete, events.OpDDL:
				c.Operations = append(c.Operations, events.Operation(part))
			}
		}
	}
	return c
}

// LoadFromCriteria populates inputs from active filter.
func (m *FilterModel) LoadFromCriteria(c filters.Criteria) {
	m.opsInput.SetValue("")
	if len(c.Operations) > 0 {
		ops := make([]string, len(c.Operations))
		for i, o := range c.Operations {
			ops[i] = o.String()
		}
		m.opsInput.SetValue(strings.Join(ops, ","))
	}
	m.schemaInput.SetValue(c.Schema)
	m.tableInput.SetValue(c.Table)
}

// Render draws the filter modal.
func (m FilterModel) Render() string {
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	body := strings.Join([]string{
		"Secondary filter (Tab next field, Enter apply, Esc cancel)",
		m.opsInput.View(),
		m.schemaInput.View(),
		m.tableInput.View(),
	}, "\n")
	return box.Render(body)
}
