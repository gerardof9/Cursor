package ui

// HelpText returns keybinding help for the overlay.
func HelpText() string {
	return `Navigation: ↑/k ↓/j  PgUp/PgDn  Home/g End/G
Sources:    o open file
Scope:      s change investigation scope
Filter:     f secondary filter (schema/table/op)   c clear filters
Analysis:   Esc cancel during analysis/open
View:       ? help   q quit`
}
