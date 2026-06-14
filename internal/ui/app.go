package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"db-log-explorer/internal/events"
	"db-log-explorer/internal/explorer"
	"db-log-explorer/internal/filters"
	"db-log-explorer/internal/sources/mysql"
)

const minWidth = 80
const minHeight = 24

type modal int

const (
	modalNone modal = iota
	modalOpen
	modalFilter
	modalHelp
	modalScope
)

// AnalysisProgressMsg delivers analysis progress from background goroutine.
type AnalysisProgressMsg struct {
	SourceID string
}

// AnalysisDoneMsg delivers analysis completion from background goroutine.
type AnalysisDoneMsg struct {
	SourceID string
	Err      error
}

// IndexBatchMsg delivers parsed summaries from background indexer.
type IndexBatchMsg struct {
	SourceID  string
	Summaries []events.EventSummary
	Done      bool
	Err       error
}

// DetailLoadedMsg delivers on-demand event detail.
type DetailLoadedMsg struct {
	Seq    uint64
	Detail events.EventDetail
	Err    error
}

// Model is the root Bubble Tea model.
type Model struct {
	session      *explorer.Session
	width        int
	height       int
	activeModal  modal
	openFile     OpenFileModel
	filterEditor FilterModel
	scopeDialog  ScopeDialogModel
	showHelp     bool
	quitting     bool
	analysisCh        map[string]chan AnalysisDoneMsg
	indexCh           map[string]chan IndexBatchMsg
	detailDebounceGen uint64
	exitCode          int
}

// NewModel creates the application model.
func NewModel(session *explorer.Session, initialExitCode int) Model {
	return Model{
		session:      session,
		openFile:     NewOpenFileModel(),
		filterEditor: NewFilterModel(),
		analysisCh:   make(map[string]chan AnalysisDoneMsg),
		indexCh:      make(map[string]chan IndexBatchMsg),
		exitCode:     initialExitCode,
	}
}

// ExitCode returns process exit code after quit.
func (m Model) ExitCode() int {
	return m.exitCode
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, src := range m.session.Sources {
		cmds = append(cmds, m.startAnalyzer(src))
	}
	return tea.Batch(cmds...)
}

func (m *Model) startAnalyzer(src *mysql.Source) tea.Cmd {
	ch := make(chan AnalysisDoneMsg, 4)
	m.analysisCh[src.ID] = ch
	go m.runAnalyzer(src, ch)
	return waitAnalysisDone(ch)
}

func (m *Model) runAnalyzer(src *mysql.Source, ch chan AnalysisDoneMsg) {
	_, err := mysql.AnalyzeStream(src, func(p mysql.AnalysisProgress) {
		// progress polled via BytesRead on source during tick-less updates
	})
	ch <- AnalysisDoneMsg{SourceID: src.ID, Err: err}
}

func waitAnalysisDone(ch chan AnalysisDoneMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (m *Model) startIndexer(src *mysql.Source) tea.Cmd {
	ch := make(chan IndexBatchMsg, 8)
	m.indexCh[src.ID] = ch
	go m.runIndexer(src, ch)
	return waitIndexBatch(ch)
}

func (m *Model) runIndexer(src *mysql.Source, ch chan IndexBatchMsg) {
	var batch []events.EventSummary
	scope := m.session.InvestigationScope
	emit := func(s events.EventSummary) {
		batch = append(batch, s)
		if len(batch) >= 64 {
			ch <- IndexBatchMsg{SourceID: src.ID, Summaries: batch}
			batch = nil
		}
	}
	err := mysql.IndexStream(src, scope, emit)
	if len(batch) > 0 {
		ch <- IndexBatchMsg{SourceID: src.ID, Summaries: batch}
	}
	ch <- IndexBatchMsg{SourceID: src.ID, Done: true, Err: err}
}

func waitIndexBatch(ch chan IndexBatchMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.session.ClampListOffset(m.listViewportRows())
		return m, nil

	case tea.KeyMsg:
		if m.handleBlockingEsc(msg) {
			return m, nil
		}
		if m.activeModal == modalHelp {
			m.activeModal = modalNone
			m.showHelp = false
			return m, nil
		}
		if m.activeModal == modalOpen {
			cmd, done := m.openFile.Update(msg)
			if done {
				m.activeModal = modalNone
				if !m.openFile.Cancelled() {
					path := strings.TrimSpace(m.openFile.Value())
					if path != "" {
						return m, m.openPath(path)
					}
				}
				m.openFile.Reset()
			}
			return m, cmd
		}
		if m.activeModal == modalScope {
			cmd, done, cancelled, scope := m.scopeDialog.Update(msg)
			if done {
				m.activeModal = modalNone
				if cancelled {
					m.session.AwaitingScope = false
					if m.scopeDialog.priorScope != nil {
						cp := *m.scopeDialog.priorScope
						m.session.InvestigationScope = &cp
						m.session.PriorScope = nil
						if len(m.session.Index) == 0 {
							return m, m.startAllIndexers()
						}
					}
					return m, cmd
				}
				if scope != nil {
					if m.scopeDialog.changeMode || m.session.RescopeAfterOpen {
						return m, m.applyScopeChange(*scope)
					}
					return m, m.applyScope(*scope)
				}
			}
			return m, cmd
		}
		if m.activeModal == modalFilter {
			switch msg.String() {
			case "tab":
				m.filterEditor.CycleField()
				return m, nil
			case "enter":
				m.session.ApplyFilter(m.filterEditor.ToCriteria())
				m.activeModal = modalNone
				return m, m.scheduleDetailLoad()
			case "esc":
				m.activeModal = modalNone
				return m, nil
			default:
				if ti := m.filterEditor.activeInput(); ti != nil {
					var cmd tea.Cmd
					*ti, cmd = ti.Update(msg)
					return m, cmd
				}
			}
			return m, nil
		}

		if m.session.PendingAnalysisCount() > 0 {
			return m, nil
		}
		if m.session.AwaitingScope || m.activeModal == modalScope {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.session.Close()
			return m, tea.Quit
		case "o":
			m.activeModal = modalOpen
			m.openFile.Reset()
			return m, nil
		case "s":
			if m.session.PendingAnalysisCount() > 0 || !m.session.AnalysisSummary.Complete {
				return m, nil
			}
			m.scopeDialog = NewScopeDialogModel(m.session, m.session.InvestigationScope != nil)
			m.activeModal = modalScope
			return m, nil
		case "f":
			if m.session.InvestigationScope == nil {
				return m, nil
			}
			m.filterEditor.LoadFromCriteria(m.session.Filter)
			m.activeModal = modalFilter
			return m, nil
		case "c":
			m.session.ClearFilter()
			return m, nil
		case "?":
			m.activeModal = modalHelp
			m.showHelp = true
			return m, nil
		case "up", "k":
			m.session.MoveSelection(-1, m.listViewportRows())
			return m, m.scheduleDetailLoad()
		case "down", "j":
			m.session.MoveSelection(1, m.listViewportRows())
			return m, m.scheduleDetailLoad()
		case "pgup":
			m.session.MoveSelection(-10, m.listViewportRows())
			return m, m.scheduleDetailLoad()
		case "pgdown":
			m.session.MoveSelection(10, m.listViewportRows())
			return m, m.scheduleDetailLoad()
		case "home", "g":
			m.session.SetSelection(0, m.listViewportRows())
			return m, m.scheduleDetailLoad()
		case "end", "G":
			m.session.SetSelection(len(m.session.Filtered)-1, m.listViewportRows())
			return m, m.scheduleDetailLoad()
		}

	case detailDebounceMsg:
		if msg.gen != m.detailDebounceGen {
			return m, nil
		}
		return m, m.startDetailLoad()

	case AnalysisDoneMsg:
		m.session.FinishSourceAnalysis(msg.SourceID, msg.Err)
		delete(m.analysisCh, msg.SourceID)
		if m.session.PendingAnalysisCount() == 0 {
			return m, m.onAnalysisComplete()
		}
		return m, nil

	case IndexBatchMsg:
		var cmds []tea.Cmd
		if len(msg.Summaries) > 0 {
			m.session.AppendSummaries(msg.Summaries)
		}
		if msg.Done {
			m.session.FinishSourceIndexing(msg.SourceID, msg.Err)
			delete(m.indexCh, msg.SourceID)
		} else if ch, ok := m.indexCh[msg.SourceID]; ok {
			cmds = append(cmds, waitIndexBatch(ch))
		}
		if m.session.Selected < 0 && len(m.session.Filtered) > 0 {
			m.session.SetSelection(0, m.listViewportRows())
		}
		if _, indexing := m.session.IndexingProgress(); !indexing && m.session.Selected >= 0 && len(m.session.Filtered) > 0 {
			cmds = append(cmds, m.scheduleDetailLoad())
		}
		return m, tea.Batch(cmds...)

	case DetailLoadedMsg:
		return m.handleDetailLoaded(msg)
	}

	return m, nil
}

func (m *Model) handleBlockingEsc(msg tea.KeyMsg) bool {
	if msg.String() != "esc" {
		return false
	}
	if m.session.PendingAnalysisCount() > 0 {
		m.cancelPendingOpen()
		return true
	}
	return false
}

func (m *Model) cancelPendingOpen() {
	m.session.RemoveIncompleteSources()
	m.session.StatusMsg = "Open cancelled"
}

func (m *Model) onAnalysisComplete() tea.Cmd {
	if m.session.AnalysisSummary.SourceCount == 0 {
		m.session.StatusMsg = "Analysis failed for all sources"
		return nil
	}
	if m.session.RescopeAfterOpen {
		m.session.RescopeAfterOpen = false
		m.session.AwaitingScope = true
		m.scopeDialog = NewScopeDialogModel(m.session, true)
		m.activeModal = modalScope
		return nil
	}
	if m.session.LaunchScope != nil {
		if err := m.session.ValidateScopeBounds(m.session.LaunchScope.From, m.session.LaunchScope.To); err != nil {
			m.session.StatusMsg = err.Error()
			m.session.AwaitingScope = false
			return nil
		}
		scope := *m.session.LaunchScope
		return m.applyScope(scope)
	}
	m.session.AwaitingScope = true
	m.scopeDialog = NewScopeDialogModel(m.session, false)
	m.activeModal = modalScope
	return nil
}

func (m *Model) applyScope(scope events.InvestigationScope) tea.Cmd {
	m.session.ConfirmScope(scope)
	return m.startAllIndexers()
}

func (m *Model) applyScopeChange(scope events.InvestigationScope) tea.Cmd {
	m.session.PriorScope = nil
	needReindex := m.session.ApplyScopeChange(scope)
	if !needReindex {
		return m.scheduleDetailLoad()
	}
	return m.startAllIndexers()
}

func (m *Model) startAllIndexers() tea.Cmd {
	var cmds []tea.Cmd
	for _, src := range m.session.Sources {
		if src.State == mysql.StateError {
			continue
		}
		cmds = append(cmds, m.startIndexer(src))
	}
	return tea.Batch(cmds...)
}

func (m *Model) openPath(path string) tea.Cmd {
	hadScope := m.session.InvestigationScope != nil
	if err := m.session.OpenSource(path); err != nil {
		m.session.StatusMsg = err.Error()
		return nil
	}
	if hadScope {
		cp := *m.session.InvestigationScope
		m.session.PriorScope = &cp
		m.session.RescopeAfterOpen = true
		m.session.ClearIndex()
		m.session.InvestigationScope = nil
		m.session.AwaitingScope = false
		m.session.Filter = filters.Criteria{}
	}
	src := m.session.Sources[len(m.session.Sources)-1]
	return m.startAnalyzer(src)
}

func (m Model) View() string {
	if m.width < minWidth || m.height < minHeight {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Terminal too small; resize to 80x24")
	}

	topH := 1
	statusH := 1
	bodyH := m.height - topH - statusH
	if bodyH < 3 {
		bodyH = 3
	}
	listW := (m.width * 60) / 100
	if listW < 30 {
		listW = m.width / 2
	}
	detailW := m.width - listW - 1
	if detailW < 20 {
		detailW = 20
	}

	top := m.renderTopBar()
	status := m.renderStatusBar()

	listItems := m.visibleEvents()
	selected := m.session.Selected
	emptyMsg := m.session.EmptyIndexMessage()
	listPane := RenderList(listW, bodyH, listItems, selected, m.session.ListOffset, emptyMsg)

	filteredEmpty := m.session.HasActiveFilter() && len(m.session.Filtered) == 0
	detailPane := RenderDetail(detailW, bodyH, m.session.DetailLoading, m.session.DetailPreview, m.session.Detail, filteredEmpty)

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(listW).Height(bodyH).Render(listPane),
		lipgloss.NewStyle().Width(detailW).Height(bodyH).Render(detailPane),
	)

	view := lipgloss.JoinVertical(lipgloss.Left, top, body, status)

	if m.activeModal == modalOpen {
		view += "\n" + m.openFile.Render()
	}
	if m.activeModal == modalScope {
		view += "\n" + m.scopeDialog.Render()
	}
	if m.activeModal == modalFilter {
		view += "\n" + m.filterEditor.Render()
	}
	if m.activeModal == modalHelp {
		view += "\n" + lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Render(HelpText())
	}

	return view
}

func (m Model) listViewportRows() int {
	bodyH := m.height - 2 // top bar + status bar
	if bodyH < 3 {
		bodyH = 3
	}
	rows := bodyH - 1 // list header row
	if rows < 1 {
		return 1
	}
	return rows
}

func (m Model) visibleEvents() []events.EventSummary {
	if m.session.InvestigationScope == nil {
		return nil
	}
	out := make([]events.EventSummary, 0, len(m.session.Filtered))
	for _, idx := range m.session.Filtered {
		out = append(out, m.session.Index[idx])
	}
	return out
}

func (m Model) renderTopBar() string {
	parts := []string{"binlog-explorer"}
	if label := m.session.ScopeLabel(); label != "" {
		parts = append(parts, "scope: "+label)
	}
	filter := m.session.Filter.Summary()
	if filter != "" {
		parts = append(parts, "Filters: "+filter)
	} else {
		parts = append(parts, "o open  s scope  f filter  ? help  q quit")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Join(parts, "  |  "))
}

func (m Model) renderStatusBar() string {
	status := m.session.FilteredLabel()

	if pct, active := m.session.AnalysisProgress(); active {
		names := strings.Join(m.session.SourceNames(), ", ")
		status = fmt.Sprintf("Analyzing: %d%% | %s | %s", pct, names, status)
	} else if m.session.AwaitingScope {
		status = "Analysis complete — select investigation scope | " + status
	} else if pct, active := m.session.IndexingProgress(); active {
		scope := m.session.ScopeLabel()
		status = fmt.Sprintf("Indexing scope: %s | %d%% | %s", scope, pct, status)
	}

	if msg := m.session.StatusMsg; msg != "" {
		status += " | " + msg
	}
	if warn := m.session.WarningsText(); warn != "" {
		status += " | " + warn
	}
	if m.session.PendingAnalysisCount() > 0 {
		status += " | Esc cancel"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(status)
}
