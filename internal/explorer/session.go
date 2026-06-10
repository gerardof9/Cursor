package explorer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"db-log-explorer/internal/events"
	"db-log-explorer/internal/filters"
	"db-log-explorer/internal/sources/mysql"
)

// Session is the in-memory exploration state.
type Session struct {
	Sources        []*mysql.Source
	sourceByID     map[string]*mysql.Source
	Index          []events.EventSummary
	nextEventID    uint64
	Filter         filters.Criteria
	Filtered       []int
	Selected       int
	Detail         *events.EventDetail
	DetailLoading  bool
	detailSeq      uint64
	StatusMsg      string
	LaunchWarnings []string
}

// NewSession creates an empty session.
func NewSession() *Session {
	return &Session{
		sourceByID: make(map[string]*mysql.Source),
		Selected:   -1,
	}
}

// OpenSource opens a binlog file and registers it in the session.
func (s *Session) OpenSource(path string) error {
	src, err := mysql.OpenSource(path)
	if err != nil {
		return err
	}
	s.Sources = append(s.Sources, src)
	s.sourceByID[src.ID] = src
	s.StatusMsg = fmt.Sprintf("Opened %s", filepath.Base(src.Path))
	return nil
}

// SourceByID returns a registered source.
func (s *Session) SourceByID(id string) *mysql.Source {
	return s.sourceByID[id]
}

// SourcesPaths returns basenames for status display.
func (s *Session) SourceNames() []string {
	out := make([]string, len(s.Sources))
	for i, src := range s.Sources {
		out[i] = filepath.Base(src.Path)
	}
	return out
}

// AppendSummaries merges new summaries and resorts the index.
func (s *Session) AppendSummaries(batch []events.EventSummary) {
	for _, sum := range batch {
		sum.ID = s.nextEventID
		s.nextEventID++
		s.Index = append(s.Index, sum)
	}
	s.sortIndex()
	s.recomputeFiltered(false)
}

// FinishSourceIndexing marks a source ready after indexing completes.
func (s *Session) FinishSourceIndexing(sourceID string, err error) {
	src := s.sourceByID[sourceID]
	if src == nil {
		return
	}
	if err != nil {
		src.MarkError(err)
		s.StatusMsg = fmt.Sprintf("Index error %s: %v", filepath.Base(src.Path), err)
		return
	}
	src.MarkReady()
	s.recomputeFiltered(false)
}

// sortIndex sorts by timestamp, source, offset.
func (s *Session) sortIndex() {
	sort.SliceStable(s.Index, func(i, j int) bool {
		a, b := s.Index[i], s.Index[j]
		if !a.Timestamp.Equal(b.Timestamp) {
			return a.Timestamp.Before(b.Timestamp)
		}
		if a.SourceID != b.SourceID {
			return a.SourceID < b.SourceID
		}
		return a.Offset < b.Offset
	})
}

// ApplyFilter sets criteria and recomputes filtered indices.
func (s *Session) ApplyFilter(c filters.Criteria) {
	s.Filter = c
	s.recomputeFiltered(true)
}

// ClearFilter removes all filters.
func (s *Session) ClearFilter() {
	s.Filter = filters.Criteria{}
	s.recomputeFiltered(true)
	s.StatusMsg = "Filters cleared"
}

func (s *Session) recomputeFiltered(adjustSelection bool) {
	prevEventID := uint64(0)
	if s.Selected >= 0 && s.Selected < len(s.Filtered) {
		prevEventID = s.Index[s.Filtered[s.Selected]].ID
	}

	s.Filtered = filters.Apply(s.Index, s.Filter)

	if !adjustSelection {
		return
	}

	s.Selected = -1
	s.Detail = nil
	s.DetailLoading = false

	if prevEventID == 0 {
		if len(s.Filtered) > 0 {
			s.Selected = 0
		}
		return
	}

	for i, idx := range s.Filtered {
		if s.Index[idx].ID == prevEventID {
			s.Selected = i
			return
		}
	}

	// Nearest following event by index order.
	for i, idx := range s.Filtered {
		if s.Index[idx].ID > prevEventID {
			s.Selected = i
			s.StatusMsg = "Selection moved to nearest match"
			return
		}
	}

	if len(s.Filtered) > 0 {
		s.Selected = len(s.Filtered) - 1
		s.StatusMsg = "Selection moved to nearest match"
	} else if s.Filter.Active() {
		s.StatusMsg = "No events match filter"
	}
}

// SelectedSummary returns the currently selected event summary.
func (s *Session) SelectedSummary() (events.EventSummary, bool) {
	if s.Selected < 0 || s.Selected >= len(s.Filtered) {
		return events.EventSummary{}, false
	}
	return s.Index[s.Filtered[s.Selected]], true
}

// SetDetail stores loaded detail and clears loading flag.
func (s *Session) SetDetail(seq uint64, detail events.EventDetail) {
	if seq != s.detailSeq {
		return
	}
	s.Detail = &detail
	s.DetailLoading = false
}

// BeginDetailLoad increments detail generation.
func (s *Session) BeginDetailLoad() uint64 {
	s.detailSeq++
	s.DetailLoading = true
	s.Detail = nil
	return s.detailSeq
}

// MoveSelection adjusts selection within filtered list.
func (s *Session) MoveSelection(delta int) {
	if len(s.Filtered) == 0 {
		s.Selected = -1
		return
	}
	if s.Selected < 0 {
		s.Selected = 0
		return
	}
	s.Selected += delta
	if s.Selected < 0 {
		s.Selected = 0
	}
	if s.Selected >= len(s.Filtered) {
		s.Selected = len(s.Filtered) - 1
	}
}

// SetSelection sets absolute selection in filtered list.
func (s *Session) SetSelection(pos int) {
	if len(s.Filtered) == 0 {
		s.Selected = -1
		return
	}
	if pos < 0 {
		pos = 0
	}
	if pos >= len(s.Filtered) {
		pos = len(s.Filtered) - 1
	}
	s.Selected = pos
}

// IndexingProgress returns aggregate indexing percent across sources.
func (s *Session) IndexingProgress() (pct int, active bool) {
	if len(s.Sources) == 0 {
		return 0, false
	}
	total := 0
	activeCount := 0
	for _, src := range s.Sources {
		total += src.IndexProgress()
		if src.State == mysql.StateIndexing {
			activeCount++
		}
	}
	return total / len(s.Sources), activeCount > 0
}

// FilteredLabel returns status text for event counts.
func (s *Session) FilteredLabel() string {
	total := len(s.Index)
	shown := len(s.Filtered)
	if s.Filter.Active() {
		return fmt.Sprintf("%d / %d events (filtered)", shown, total)
	}
	return fmt.Sprintf("%d events", total)
}

// HasActiveFilter returns whether filters are applied.
func (s *Session) HasActiveFilter() bool {
	return s.Filter.Active()
}

// EmptyIndexMessage returns a user-facing list placeholder.
func (s *Session) EmptyIndexMessage() string {
	if len(s.Sources) == 0 {
		return "Open a binlog: press o"
	}
	if s.HasActiveFilter() && len(s.Filtered) == 0 {
		return "No events match filter"
	}
	if len(s.Index) == 0 {
		for _, src := range s.Sources {
			if src.State == mysql.StateError {
				return fmt.Sprintf("Source error: %s", src.Error)
			}
			if src.State == mysql.StateIndexing {
				return "Indexing events..."
			}
		}
		return "No user-data change events in open source(s)"
	}
	return ""
}

// Close releases open sources.
func (s *Session) Close() {
	for _, src := range s.Sources {
		_ = src.Close()
	}
}

// AddLaunchWarning records a non-fatal startup issue.
func (s *Session) AddLaunchWarning(msg string) {
	s.LaunchWarnings = append(s.LaunchWarnings, msg)
}

// WarningsText joins launch warnings for status display.
func (s *Session) WarningsText() string {
	return strings.Join(s.LaunchWarnings, "; ")
}
