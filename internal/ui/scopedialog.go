package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"db-log-explorer/internal/events"
	"db-log-explorer/internal/explorer"
)

type scopeField int

const (
	scopeFieldFrom scopeField = iota
	scopeFieldTo
)

// ScopeDialogModel is the investigation scope selection overlay.
type ScopeDialogModel struct {
	session     *explorer.Session
	changeMode  bool
	priorScope  *events.InvestigationScope
	selected    events.ScopePreset
	customFrom  textinput.Model
	customTo    textinput.Model
	activeField scopeField
	errMsg      string
	showWarning bool
}

// NewScopeDialogModel creates a scope dialog bound to session analysis.
func NewScopeDialogModel(session *explorer.Session, changeMode bool) ScopeDialogModel {
	from := textinput.New()
	from.Width = 22
	from.CharLimit = 32
	from.Prompt = "From: "

	to := textinput.New()
	to.Width = 22
	to.CharLimit = 32
	to.Prompt = "To:   "

	m := ScopeDialogModel{
		session:    session,
		changeMode: changeMode,
		selected:   events.ScopeLastDay,
		customFrom: from,
		customTo:   to,
	}
	if changeMode && session.InvestigationScope != nil {
		cp := *session.InvestigationScope
		m.priorScope = &cp
	} else if changeMode && session.PriorScope != nil {
		cp := *session.PriorScope
		m.priorScope = &cp
	}
	m.syncCustomDefaults()
	m.customFrom.Blur()
	m.customTo.Blur()
	return m
}

func (m *ScopeDialogModel) syncCustomDefaults() {
	sum := m.session.AnalysisSummary
	layout := "2006-01-02 15:04:05"
	if !sum.MergedMin.IsZero() {
		m.customFrom.SetValue(sum.MergedMin.Format(layout))
	}
	if !sum.MergedMax.IsZero() {
		m.customTo.SetValue(sum.MergedMax.Format(layout))
	}
}

// Update handles scope dialog keys. done=true when scope confirmed or cancelled.
// cancelled=true when Esc dismisses without applying (revert in change mode).
func (m *ScopeDialogModel) Update(msg tea.Msg) (tea.Cmd, bool, bool, *events.InvestigationScope) {
	if m.showWarning {
		return m.updateWarning(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return nil, true, true, nil
		case "tab":
			if m.selected == events.ScopeCustom {
				if m.activeField == scopeFieldFrom {
					m.focusCustom(scopeFieldTo)
				} else {
					m.focusCustom(scopeFieldFrom)
				}
			}
			return nil, false, false, nil
		case "enter":
			return m.confirm()
		case "1", "2", "3", "4":
			if m.selected == events.ScopeCustom {
				return m.updateCustomField(msg)
			}
			switch msg.String() {
			case "1":
				m.selected = events.ScopeEntire
			case "2":
				m.selected = events.ScopeLastHour
			case "3":
				m.selected = events.ScopeLastDay
			case "4":
				m.selected = events.ScopeCustom
				m.focusCustom(scopeFieldFrom)
			}
			m.errMsg = ""
			return nil, false, false, nil
		default:
			if m.selected == events.ScopeCustom {
				return m.updateCustomField(msg)
			}
		}
	}
	return nil, false, false, nil
}

func (m *ScopeDialogModel) updateCustomField(msg tea.KeyMsg) (tea.Cmd, bool, bool, *events.InvestigationScope) {
	var cmd tea.Cmd
	if m.activeField == scopeFieldFrom {
		m.customFrom, cmd = m.customFrom.Update(msg)
	} else {
		m.customTo, cmd = m.customTo.Update(msg)
	}
	return cmd, false, false, nil
}

func (m *ScopeDialogModel) focusCustom(f scopeField) {
	m.activeField = f
	m.customFrom.Blur()
	m.customTo.Blur()
	if f == scopeFieldFrom {
		m.customFrom.Focus()
	} else {
		m.customTo.Focus()
	}
}

func (m *ScopeDialogModel) confirm() (tea.Cmd, bool, bool, *events.InvestigationScope) {
	if m.selected == events.ScopeEntire && m.session.NeedsLargeFileWarning() {
		m.showWarning = true
		return nil, false, false, nil
	}
	return m.applySelection()
}

func (m *ScopeDialogModel) updateWarning(msg tea.Msg) (tea.Cmd, bool, bool, *events.InvestigationScope) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil, false, false, nil
	}
	switch strings.ToLower(key.String()) {
	case "y":
		m.showWarning = false
		return m.applySelection()
	case "f":
		m.showWarning = false
		m.selected = events.ScopeCustom
		m.focusCustom(scopeFieldFrom)
		return nil, false, false, nil
	case "esc":
		m.showWarning = false
		return nil, false, false, nil
	}
	return nil, false, false, nil
}

func (m *ScopeDialogModel) applySelection() (tea.Cmd, bool, bool, *events.InvestigationScope) {
	var from, to time.Time
	var err error
	if m.selected == events.ScopeCustom {
		from, err = parseScopeTime(m.customFrom.Value())
		if err != nil {
			m.errMsg = "Invalid From: " + err.Error()
			return nil, false, false, nil
		}
		to, err = parseScopeTime(m.customTo.Value())
		if err != nil {
			m.errMsg = "Invalid To: " + err.Error()
			return nil, false, false, nil
		}
	}
	scope, err := m.session.ScopeFromPreset(m.selected, from, to)
	if err != nil {
		m.errMsg = err.Error()
		return nil, false, false, nil
	}
	return nil, true, false, &scope
}

func parseScopeTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS")
}

// Render draws the scope dialog or warning modal.
func (m ScopeDialogModel) Render() string {
	if m.showWarning {
		return m.renderWarning()
	}
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	sum := m.session.AnalysisSummary

	title := "Investigation Scope"
	if m.changeMode {
		title = "Change Investigation Scope"
	}

	names := m.session.SourceNames()
	selectedName := "none"
	if len(names) > 0 {
		selectedName = names[0]
		if len(names) > 1 {
			selectedName += fmt.Sprintf(" (+ %d others)", len(names)-1)
		}
	}

	lines := []string{
		title,
		fmt.Sprintf("Selected: %s", selectedName),
		fmt.Sprintf("Size: %s", sum.TotalSizeHuman),
	}
	if !sum.MergedMin.IsZero() && !sum.MergedMax.IsZero() {
		layout := "2006-01-02 15:04:05"
		lines = append(lines,
			fmt.Sprintf("Time span: %s .. %s", sum.MergedMin.Format(layout), sum.MergedMax.Format(layout)),
		)
	}
	lines = append(lines, fmt.Sprintf("Change events: ~%s", formatCount(sum.ApproxChangeCount)), "")

	if m.session.InvestigationScope == nil {
		lines = append(lines, "No investigation scope defined.", "")
	}

	lines = append(lines,
		optionLine("1", events.ScopeEntire, m.selected),
		optionLine("2", events.ScopeLastHour, m.selected),
		optionLine("3", events.ScopeLastDay, m.selected),
		optionLine("4", events.ScopeCustom, m.selected),
		"",
	)

	if m.selected == events.ScopeCustom {
		lines = append(lines, "Custom "+m.customFrom.View(), "Custom "+m.customTo.View(), "")
	}

	lines = append(lines, "1-4 select preset   Enter confirm   Esc cancel")
	if m.errMsg != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.errMsg))
	}

	return box.Render(strings.Join(lines, "\n"))
}

func (m ScopeDialogModel) renderWarning() string {
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	sum := m.session.AnalysisSummary
	layout := "2006-01-02"
	lines := []string{
		"Warning",
		fmt.Sprintf("Large file detected (%s).", sum.TotalSizeHuman),
		fmt.Sprintf("~%s change events", formatCount(sum.ApproxChangeCount)),
	}
	if !sum.MergedMin.IsZero() && !sum.MergedMax.IsZero() {
		lines = append(lines, fmt.Sprintf("Time span: %s .. %s", sum.MergedMin.Format(layout), sum.MergedMax.Format(layout)))
	}
	lines = append(lines,
		"",
		"Full indexing may take several minutes and use significant memory.",
		"",
		"[Y] Continue with entire file   [F] Define date range",
	)
	return box.Render(strings.Join(lines, "\n"))
}

func optionLine(key string, preset events.ScopePreset, selected events.ScopePreset) string {
	label := presetLabel(preset)
	marker := " "
	if preset == selected {
		marker = "*"
	}
	return fmt.Sprintf("[%s]%s %s", key, marker, label)
}

func presetLabel(p events.ScopePreset) string {
	switch p {
	case events.ScopeEntire:
		return "Entire file"
	case events.ScopeLastHour:
		return "Last hour (of file)"
	case events.ScopeLastDay:
		return "Last day (of file)"
	case events.ScopeCustom:
		return "Custom range"
	default:
		return string(p)
	}
}

func formatCount(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var out []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}
