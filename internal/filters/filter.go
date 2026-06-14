package filters

import (
	"strings"

	"db-log-explorer/internal/events"
)

// Criteria holds active filter constraints (AND combination).
type Criteria struct {
	Operations []events.Operation
	Schema     string
	Table      string
}

func (c Criteria) Active() bool {
	return len(c.Operations) > 0 || c.Schema != "" || c.Table != ""
}

func (c Criteria) Summary() string {
	var parts []string
	if len(c.Operations) > 0 {
		ops := make([]string, len(c.Operations))
		for i, o := range c.Operations {
			ops[i] = o.String()
		}
		parts = append(parts, "op="+strings.Join(ops, "|"))
	}
	if c.Schema != "" {
		parts = append(parts, "schema="+c.Schema)
	}
	if c.Table != "" {
		parts = append(parts, "table="+c.Table)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

// Apply returns indices into index that pass all active criteria.
func Apply(index []events.EventSummary, c Criteria) []int {
	if !c.Active() {
		out := make([]int, len(index))
		for i := range index {
			out[i] = i
		}
		return out
	}

	var out []int
	for i, ev := range index {
		if matches(ev, c) {
			out = append(out, i)
		}
	}
	return out
}

func matches(ev events.EventSummary, c Criteria) bool {
	if len(c.Operations) > 0 && !containsOp(c.Operations, ev.Operation) {
		return false
	}
	if c.Schema != "" && !strings.EqualFold(ev.Schema, c.Schema) {
		return false
	}
	if c.Table != "" && !tableMatches(ev, c.Table) {
		return false
	}
	return true
}

func containsOp(ops []events.Operation, op events.Operation) bool {
	for _, o := range ops {
		if o == op {
			return true
		}
	}
	return false
}

func tableMatches(ev events.EventSummary, table string) bool {
	table = strings.TrimSpace(table)
	if table == "" {
		return true
	}
	if strings.Contains(table, ".") {
		parts := strings.SplitN(table, ".", 2)
		return strings.EqualFold(ev.Schema, parts[0]) && strings.EqualFold(ev.Table, parts[1])
	}
	return strings.EqualFold(ev.Table, table)
}
