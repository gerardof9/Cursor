package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"db-log-explorer/internal/events"
	"db-log-explorer/internal/explorer"
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
)

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
	showHelp     bool
	quitting     bool
	indexCh      map[string]chan IndexBatchMsg
	exitCode     int
}

// NewModel creates the application model.
func NewModel(session *explorer.Session, initialExitCode int) Model {
	return Model{
		session:      session,
		openFile:     NewOpenFileModel(),
		filterEditor: NewFilterModel(),
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
		cmds = append(cmds, m.startIndexer(src))
	}
	return tea.Batch(cmds...)
}

func (m *Model) startIndexer(src *mysql.Source) tea.Cmd {
	ch := make(chan IndexBatchMsg, 8)
	m.indexCh[src.ID] = ch
	go m.runIndexer(src, ch)
	return waitIndexBatch(ch)
}

func (m *Model) runIndexer(src *mysql.Source, ch chan IndexBatchMsg) {
	var batch []events.EventSummary
	emit := func(s events.EventSummary) {
		batch = append(batch, s)
		if len(batch) >= 64 {
			ch <- IndexBatchMsg{SourceID: src.ID, Summaries: batch}
			batch = nil
		}
	}
	err := mysql.IndexStream(src, emit)
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

func loadDetailCmd(session *explorer.Session, seq uint64) tea.Cmd {
	return func() tea.Msg {
		sum, ok := session.SelectedSummary()
		if !ok {
			return DetailLoadedMsg{Seq: seq, Err: fmt.Errorf("no selection")}
		}
		src := session.SourceByID(sum.SourceID)
		if src == nil {
			return DetailLoadedMsg{Seq: seq, Err: fmt.Errorf("source not found")}
		}
		detail, err := mysql.LoadDetail(src, sum)
		return DetailLoadedMsg{Seq: seq, Detail: detail, Err: err}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
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
		if m.activeModal == modalFilter {
			switch msg.String() {
			case "tab":
				m.filterEditor.CycleField()
				return m, nil
			case "enter":
				m.session.ApplyFilter(m.filterEditor.ToCriteria())
				m.activeModal = modalNone
				seq := m.session.BeginDetailLoad()
				return m, loadDetailCmd(m.session, seq)
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

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.session.Close()
			return m, tea.Quit
		case "o":
			m.activeModal = modalOpen
			m.openFile.Reset()
			return m, nil
		case "f":
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
			m.session.MoveSelection(-1)
			seq := m.session.BeginDetailLoad()
			return m, loadDetailCmd(m.session, seq)
		case "down", "j":
			m.session.MoveSelection(1)
			seq := m.session.BeginDetailLoad()
			return m, loadDetailCmd(m.session, seq)
		case "pgup":
			m.session.MoveSelection(-10)
			seq := m.session.BeginDetailLoad()
			return m, loadDetailCmd(m.session, seq)
		case "pgdown":
			m.session.MoveSelection(10)
			seq := m.session.BeginDetailLoad()
			return m, loadDetailCmd(m.session, seq)
		case "home", "g":
			m.session.SetSelection(0)
			seq := m.session.BeginDetailLoad()
			return m, loadDetailCmd(m.session, seq)
		case "end", "G":
			m.session.SetSelection(len(m.session.Filtered) - 1)
			seq := m.session.BeginDetailLoad()
			return m, loadDetailCmd(m.session, seq)
		}

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
			m.session.SetSelection(0)
			seq := m.session.BeginDetailLoad()
			cmds = append(cmds, loadDetailCmd(m.session, seq))
		}
		return m, tea.Batch(cmds...)

	case DetailLoadedMsg:
		if msg.Err != nil {
			m.session.StatusMsg = msg.Err.Error()
			m.session.DetailLoading = false
			return m, nil
		}
		m.session.SetDetail(msg.Seq, msg.Detail)
		return m, nil
	}

	return m, nil
}

func (m *Model) openPath(path string) tea.Cmd {
	if err := m.session.OpenSource(path); err != nil {
		m.session.StatusMsg = err.Error()
		return nil
	}
	src := m.session.SourceByID(m.session.Sources[len(m.session.Sources)-1].ID)
	return m.startIndexer(src)
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
	listPane := RenderList(listW, bodyH, listItems, selected, emptyMsg)

	filteredEmpty := m.session.HasActiveFilter() && len(m.session.Filtered) == 0
	detailPane := RenderDetail(detailW, bodyH, m.session.DetailLoading, m.session.Detail, filteredEmpty)

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(listW).Height(bodyH).Render(listPane),
		lipgloss.NewStyle().Width(detailW).Height(bodyH).Render(detailPane),
	)

	view := lipgloss.JoinVertical(lipgloss.Left, top, body, status)

	if m.activeModal == modalOpen {
		view += "\n" + m.openFile.Render()
	}
	if m.activeModal == modalFilter {
		view += "\n" + m.filterEditor.Render()
	}
	if m.activeModal == modalHelp {
		view += "\n" + lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Render(HelpText())
	}

	return view
}

func (m Model) visibleEvents() []events.EventSummary {
	out := make([]events.EventSummary, 0, len(m.session.Filtered))
	for _, idx := range m.session.Filtered {
		out = append(out, m.session.Index[idx])
	}
	return out
}

func (m Model) renderTopBar() string {
	filter := m.session.Filter.Summary()
	if filter == "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("binlog-explorer  |  o open  f filter  ? help  q quit")
	}
	return lipgloss.NewStyle().Render("Filters: " + filter)
}

func (m Model) renderStatusBar() string {
	pct, indexing := m.session.IndexingProgress()
	status := m.session.FilteredLabel()
	if indexing {
		status = fmt.Sprintf("Indexing: %d%% | %s", pct, status)
	}
	if msg := m.session.StatusMsg; msg != "" {
		status += " | " + msg
	}
	if warn := m.session.WarningsText(); warn != "" {
		status += " | " + warn
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(status)
}
