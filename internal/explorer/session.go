package explorer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"db-log-explorer/internal/events"
	"db-log-explorer/internal/filters"
	"db-log-explorer/internal/sources/mysql"
)

// SessionAnalysisSummary aggregates analysis metadata for scope dialog.
type SessionAnalysisSummary struct {
	TotalSize         int64
	TotalSizeHuman    string
	MergedMin         time.Time
	MergedMax         time.Time
	ApproxChangeCount int64
	SourceCount       int
	Complete          bool
}

// Session is the in-memory exploration state.
type Session struct {
	Sources            []*mysql.Source
	sourceByID         map[string]*mysql.Source
	Index              []events.EventSummary
	nextEventID        uint64
	Filter             filters.Criteria
	Filtered           []int
	Selected           int
	ListOffset         int
	Detail             *events.EventDetail
	DetailPreview      *events.EventSummary
	DetailLoading      bool
	detailSeq          uint64
	StatusMsg          string
	LaunchWarnings     []string
	InvestigationScope *events.InvestigationScope
	AwaitingScope      bool
	LaunchScope        *events.InvestigationScope
	AnalysisSummary    SessionAnalysisSummary
	RescopeAfterOpen   bool
	PriorScope         *events.InvestigationScope
}

// NewSession creates an empty session.
func NewSession() *Session {
	return &Session{
		sourceByID: make(map[string]*mysql.Source),
		Selected:   -1,
	}
}

// OpenSource opens a binlog file and registers it in the session.
// Returns an error if a file with the same basename is already open.
func (s *Session) OpenSource(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	name := filepath.Base(abs)

	for _, src := range s.Sources {
		if sameBinlogName(src.Path, abs) {
			return fmt.Errorf("already open: %s", name)
		}
	}

	src, err := mysql.OpenSource(abs)
	if err != nil {
		return err
	}
	s.Sources = append(s.Sources, src)
	s.sourceByID[src.ID] = src
	s.StatusMsg = fmt.Sprintf("Opened %s", filepath.Base(src.Path))
	return nil
}

func sameBinlogName(existingPath, newPath string) bool {
	return strings.EqualFold(filepath.Base(existingPath), filepath.Base(newPath))
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

// FinishSourceAnalysis records analysis completion for a source.
func (s *Session) FinishSourceAnalysis(sourceID string, err error) {
	src := s.sourceByID[sourceID]
	if src == nil {
		return
	}
	if err != nil {
		src.MarkError(err)
		if src.Analysis == nil {
			src.Analysis = &events.FileAnalysisResult{
				SourceID: sourceID,
				Error:    err.Error(),
			}
		}
		s.StatusMsg = fmt.Sprintf("Analysis error %s: %v", filepath.Base(src.Path), err)
		return
	}
	s.RecomputeAnalysisSummary()
}

// RecomputeAnalysisSummary merges per-source analysis into session summary.
func (s *Session) RecomputeAnalysisSummary() {
	var sum SessionAnalysisSummary
	for _, src := range s.Sources {
		if src.Analysis == nil || !src.Analysis.Complete {
			continue
		}
		sum.SourceCount++
		sum.TotalSize += src.Analysis.FileSize
		sum.ApproxChangeCount += src.Analysis.ApproxChangeCount
		if !src.Analysis.MinTimestamp.IsZero() {
			if sum.MergedMin.IsZero() || src.Analysis.MinTimestamp.Before(sum.MergedMin) {
				sum.MergedMin = src.Analysis.MinTimestamp
			}
		}
		if !src.Analysis.MaxTimestamp.IsZero() {
			if sum.MergedMax.IsZero() || src.Analysis.MaxTimestamp.After(sum.MergedMax) {
				sum.MergedMax = src.Analysis.MaxTimestamp
			}
		}
	}
	sum.TotalSizeHuman = mysql.HumanSize(sum.TotalSize)
	sum.Complete = s.allSourcesAnalyzed()
	s.AnalysisSummary = sum
}

func (s *Session) allSourcesAnalyzed() bool {
	if len(s.Sources) == 0 {
		return false
	}
	for _, src := range s.Sources {
		if src.State == mysql.StateError {
			continue
		}
		if src.Analysis == nil || !src.Analysis.Complete {
			return false
		}
	}
	return true
}

// AnalysisProgress returns aggregate analysis percent across sources.
func (s *Session) AnalysisProgress() (pct int, active bool) {
	if len(s.Sources) == 0 {
		return 0, false
	}
	total := 0
	activeCount := 0
	for _, src := range s.Sources {
		if src.State == mysql.StateError {
			continue
		}
		if src.Analysis == nil || !src.Analysis.Complete {
			if src.State == mysql.StateAnalyzing {
				total += src.AnalysisProgress()
			}
			activeCount++
		} else {
			total += 100
		}
	}
	if activeCount == 0 {
		return total / len(s.Sources), false
	}
	return total / len(s.Sources), true
}

// PendingAnalysisCount returns sources still analyzing.
func (s *Session) PendingAnalysisCount() int {
	n := 0
	for _, src := range s.Sources {
		if src.State == mysql.StateError {
			continue
		}
		if src.Analysis == nil || !src.Analysis.Complete {
			n++
		}
	}
	return n
}

// ScopeFromPreset builds investigation scope from preset and merged analysis.
func (s *Session) ScopeFromPreset(preset events.ScopePreset, customFrom, customTo time.Time) (events.InvestigationScope, error) {
	sum := s.AnalysisSummary
	switch preset {
	case events.ScopeEntire:
		if sum.MergedMin.IsZero() || sum.MergedMax.IsZero() {
			return events.InvestigationScope{}, fmt.Errorf("no timestamps in analyzed file(s)")
		}
		return events.InvestigationScope{
			From:   sum.MergedMin,
			To:     sum.MergedMax,
			Preset: events.ScopeEntire,
		}, nil
	case events.ScopeLastHour:
		if sum.MergedMax.IsZero() {
			return events.InvestigationScope{}, fmt.Errorf("no timestamps in analyzed file(s)")
		}
		from := sum.MergedMax.Add(-time.Hour)
		if from.Before(sum.MergedMin) {
			from = sum.MergedMin
		}
		return events.InvestigationScope{
			From:   from,
			To:     sum.MergedMax,
			Preset: events.ScopeLastHour,
		}, nil
	case events.ScopeLastDay:
		if sum.MergedMax.IsZero() {
			return events.InvestigationScope{}, fmt.Errorf("no timestamps in analyzed file(s)")
		}
		from := sum.MergedMax.Add(-24 * time.Hour)
		if from.Before(sum.MergedMin) {
			from = sum.MergedMin
		}
		return events.InvestigationScope{
			From:   from,
			To:     sum.MergedMax,
			Preset: events.ScopeLastDay,
		}, nil
	case events.ScopeCustom:
		if customFrom.IsZero() || customTo.IsZero() {
			return events.InvestigationScope{}, fmt.Errorf("custom range requires From and To")
		}
		if customFrom.After(customTo) {
			return events.InvestigationScope{}, fmt.Errorf("From must be before or equal to To")
		}
		if err := s.ValidateScopeBounds(customFrom, customTo); err != nil {
			return events.InvestigationScope{}, err
		}
		return events.InvestigationScope{
			From:   customFrom,
			To:     customTo,
			Preset: events.ScopeCustom,
		}, nil
	default:
		return events.InvestigationScope{}, fmt.Errorf("unknown scope preset")
	}
}

// ValidateScopeBounds checks custom range against merged analysis min/max.
func (s *Session) ValidateScopeBounds(from, to time.Time) error {
	sum := s.AnalysisSummary
	if sum.MergedMin.IsZero() || sum.MergedMax.IsZero() {
		return nil
	}
	if from.Before(sum.MergedMin) || to.After(sum.MergedMax) {
		return fmt.Errorf("range must be within analyzed span %s .. %s",
			sum.MergedMin.Format("2006-01-02 15:04:05"),
			sum.MergedMax.Format("2006-01-02 15:04:05"))
	}
	return nil
}

// ApplyScopeChange updates scope, clears secondary filters, and narrows in memory or marks for re-index.
func (s *Session) ApplyScopeChange(scope events.InvestigationScope) bool {
	s.Filter = filters.Criteria{}
	s.Selected = -1
	s.ListOffset = 0
	s.Detail = nil
	s.DetailPreview = nil
	s.DetailLoading = false
	s.AwaitingScope = false

	old := s.InvestigationScope
	cp := scope
	s.InvestigationScope = &cp
	s.StatusMsg = fmt.Sprintf("Scope: %s .. %s",
		scope.From.Format("2006-01-02 15:04:05"),
		scope.To.Format("2006-01-02 15:04:05"))

	if old != nil && scopeSubset(&scope, old) && len(s.Index) > 0 {
		s.filterIndexByScope(scope)
		s.recomputeFiltered(true)
		return false
	}

	s.ClearIndex()
	return true
}

func scopeSubset(inner, outer *events.InvestigationScope) bool {
	if inner == nil || outer == nil {
		return false
	}
	return !inner.From.Before(outer.From) && !inner.To.After(outer.To)
}

func (s *Session) filterIndexByScope(scope events.InvestigationScope) {
	kept := s.Index[:0]
	for _, ev := range s.Index {
		if scope.Contains(ev.Timestamp) {
			kept = append(kept, ev)
		}
	}
	s.Index = kept
	s.nextEventID = 0
	for i := range s.Index {
		s.Index[i].ID = s.nextEventID
		s.nextEventID++
	}
	s.sortIndex()
}

// ConfirmScope sets active scope and clears awaiting flag.
func (s *Session) ConfirmScope(scope events.InvestigationScope) {
	cp := scope
	s.InvestigationScope = &cp
	s.AwaitingScope = false
	s.ClearIndex()
	s.StatusMsg = fmt.Sprintf("Scope: %s .. %s",
		scope.From.Format("2006-01-02 15:04:05"),
		scope.To.Format("2006-01-02 15:04:05"))
}

// ClearIndex removes indexed events and resets selection.
func (s *Session) ClearIndex() {
	s.Index = nil
	s.Filtered = nil
	s.Selected = -1
	s.ListOffset = 0
	s.Detail = nil
	s.DetailPreview = nil
	s.DetailLoading = false
	s.nextEventID = 0
	for _, src := range s.Sources {
		src.IndexedCount = 0
	}
}

// NeedsLargeFileWarning reports whether Entire-file choice should warn.
func (s *Session) NeedsLargeFileWarning() bool {
	sum := s.AnalysisSummary
	if sum.TotalSize >= mysql.LargeFileSizeBytes {
		return true
	}
	return sum.ApproxChangeCount >= mysql.LargeFileEventCount
}

// ScopeLabel returns a short display label for active scope.
func (s *Session) ScopeLabel() string {
	if s.InvestigationScope == nil {
		return ""
	}
	sc := s.InvestigationScope
	return fmt.Sprintf("%s .. %s",
		sc.From.Format("2006-01-02 15:04:05"),
		sc.To.Format("2006-01-02 15:04:05"))
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
	s.ListOffset = 0
	s.Detail = nil
	s.DetailPreview = nil
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
	s.DetailPreview = nil
	s.DetailLoading = false
}

// CancelDetailLoad invalidates an in-flight detail parse (e.g. user moved selection).
func (s *Session) CancelDetailLoad() {
	s.detailSeq++
	s.DetailLoading = false
}

// SetDetailPreview shows index metadata while full detail is debounced or loading.
func (s *Session) SetDetailPreview(sum events.EventSummary) {
	cp := sum
	s.DetailPreview = &cp
	s.Detail = nil
}

// ClearDetailPreview removes the interim summary pane.
func (s *Session) ClearDetailPreview() {
	s.DetailPreview = nil
}

// DetailGeneration returns the current detail load generation.
func (s *Session) DetailGeneration() uint64 {
	return s.detailSeq
}

// BeginDetailLoad increments detail generation for an in-flight full parse.
func (s *Session) BeginDetailLoad() uint64 {
	s.detailSeq++
	s.DetailLoading = true
	s.Detail = nil
	return s.detailSeq
}

// MoveSelection adjusts selection within filtered list and keeps it visible.
func (s *Session) MoveSelection(delta, viewportRows int) {
	if len(s.Filtered) == 0 {
		s.Selected = -1
		s.ListOffset = 0
		return
	}
	if s.Selected < 0 {
		s.Selected = 0
		s.clampListOffset(viewportRows)
		return
	}
	s.Selected += delta
	if s.Selected < 0 {
		s.Selected = 0
	}
	if s.Selected >= len(s.Filtered) {
		s.Selected = len(s.Filtered) - 1
	}
	s.clampListOffset(viewportRows)
}

// SetSelection sets absolute selection in filtered list and keeps it visible.
func (s *Session) SetSelection(pos, viewportRows int) {
	if len(s.Filtered) == 0 {
		s.Selected = -1
		s.ListOffset = 0
		return
	}
	if pos < 0 {
		pos = 0
	}
	if pos >= len(s.Filtered) {
		pos = len(s.Filtered) - 1
	}
	s.Selected = pos
	s.clampListOffset(viewportRows)
}

// ClampListOffset adjusts the list scroll position for the current selection.
func (s *Session) ClampListOffset(viewportRows int) {
	s.clampListOffset(viewportRows)
}

func (s *Session) clampListOffset(viewportRows int) {
	if s.Selected < 0 || len(s.Filtered) == 0 {
		s.ListOffset = 0
		return
	}
	maxRows := viewportRows
	if maxRows < 1 {
		maxRows = 1
	}
	if s.Selected < s.ListOffset {
		s.ListOffset = s.Selected
	}
	if s.Selected >= s.ListOffset+maxRows {
		s.ListOffset = s.Selected - maxRows + 1
	}
	maxOffset := len(s.Filtered) - maxRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.ListOffset > maxOffset {
		s.ListOffset = maxOffset
	}
	if s.ListOffset < 0 {
		s.ListOffset = 0
	}
}

// IndexingProgress returns aggregate indexing percent across sources.
func (s *Session) IndexingProgress() (pct int, active bool) {
	if len(s.Sources) == 0 {
		return 0, false
	}
	if s.InvestigationScope == nil {
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
		return fmt.Sprintf("%d / %d events", shown, total)
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
	if s.AwaitingScope {
		return "Select investigation scope to begin indexing"
	}
	if len(s.Index) == 0 {
		for _, src := range s.Sources {
			if src.State == mysql.StateError {
				return fmt.Sprintf("Source error: %s", src.Error)
			}
			if src.State == mysql.StateAnalyzing && (src.Analysis == nil || !src.Analysis.Complete) {
				return "Analyzing binlog..."
			}
			if src.State == mysql.StateIndexing {
				return "Indexing scoped events..."
			}
		}
		if s.InvestigationScope == nil && s.allSourcesAnalyzed() {
			return "Select investigation scope to begin indexing"
		}
		return "No user-data change events in scope"
	}
	return ""
}

// RemoveIncompleteSources drops sources without completed analysis and rebuilds sourceByID.
func (s *Session) RemoveIncompleteSources() {
	var kept []*mysql.Source
	for _, src := range s.Sources {
		if src.Analysis != nil && src.Analysis.Complete {
			kept = append(kept, src)
			continue
		}
		_ = src.Close()
		delete(s.sourceByID, src.ID)
	}
	s.Sources = kept
	s.RecomputeAnalysisSummary()
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
