package explorer

import (
	"fmt"
	"strings"
	"time"

	"db-log-explorer/internal/events"
)

const (
	timeLayoutFull = "2006-01-02 15:04:05"
	timeLayoutDate = "2006-01-02"
)

// ParseCLITimestamp parses CLI --from/--to values.
// Date-only: from = start of day; to = end of day 23:59:59.
func ParseCLITimestamp(value string, isEnd bool) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	if t, err := time.ParseInLocation(timeLayoutFull, value, time.Local); err == nil {
		return t, nil
	}
	t, err := time.ParseInLocation(timeLayoutDate, value, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp %q (use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)", value)
	}
	if isEnd {
		return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location()), nil
	}
	return t, nil
}

// ParseLaunchScope parses paired --from/--to into InvestigationScope.
func ParseLaunchScope(fromStr, toStr string) (*events.InvestigationScope, error) {
	from, err := ParseCLITimestamp(fromStr, false)
	if err != nil {
		return nil, fmt.Errorf("--from: %w", err)
	}
	to, err := ParseCLITimestamp(toStr, true)
	if err != nil {
		return nil, fmt.Errorf("--to: %w", err)
	}
	if from.After(to) {
		return nil, fmt.Errorf("--from must be before or equal to --to")
	}
	return &events.InvestigationScope{
		From:   from,
		To:     to,
		Preset: events.ScopeCustom,
	}, nil
}

// SetLaunchScope stores CLI-provided scope on the session.
func (s *Session) SetLaunchScope(scope *events.InvestigationScope) {
	s.LaunchScope = scope
}
