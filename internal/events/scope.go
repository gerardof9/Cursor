package events

import "time"

// ScopePreset identifies how investigation scope was chosen.
type ScopePreset string

const (
	ScopeEntire   ScopePreset = "Entire"
	ScopeLastHour ScopePreset = "LastHour"
	ScopeLastDay  ScopePreset = "LastDay"
	ScopeCustom   ScopePreset = "Custom"
)

// InvestigationScope is the inclusive time window for scoped indexing.
type InvestigationScope struct {
	From   time.Time
	To     time.Time
	Preset ScopePreset
}

// Contains reports whether ts falls within [From, To] inclusive.
func (s InvestigationScope) Contains(ts time.Time) bool {
	if ts.Before(s.From) {
		return false
	}
	return !ts.After(s.To)
}

// FileAnalysisResult holds lightweight analysis metadata for one source.
type FileAnalysisResult struct {
	SourceID          string
	FileSize          int64
	FileSizeHuman     string
	MinTimestamp      time.Time
	MaxTimestamp      time.Time
	ApproxChangeCount int64
	Complete          bool
	Error             string
}
