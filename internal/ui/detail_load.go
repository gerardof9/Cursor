package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"db-log-explorer/internal/events"
	"db-log-explorer/internal/explorer"
	"db-log-explorer/internal/sources/mysql"
)

const detailDebounce = 250 * time.Millisecond

// detailDebounceMsg fires after navigation settles to load one detail payload.
type detailDebounceMsg struct {
	gen uint64
}

func loadDetailCmd(session *explorer.Session, seq uint64, sum events.EventSummary) tea.Cmd {
	srcID := sum.SourceID
	return func() tea.Msg {
		src := session.SourceByID(srcID)
		if src == nil {
			return DetailLoadedMsg{Seq: seq, Err: fmt.Errorf("source not found")}
		}
		detail, err := mysql.LoadDetail(src, sum)
		return DetailLoadedMsg{Seq: seq, Detail: detail, Err: err}
	}
}

func (m *Model) scheduleDetailLoad() tea.Cmd {
	m.detailDebounceGen++
	gen := m.detailDebounceGen
	m.session.CancelDetailLoad()
	if sum, ok := m.session.SelectedSummary(); ok {
		m.session.SetDetailPreview(sum)
	} else {
		m.session.ClearDetailPreview()
	}
	return tea.Tick(detailDebounce, func(time.Time) tea.Msg {
		return detailDebounceMsg{gen: gen}
	})
}

func (m *Model) startDetailLoad() tea.Cmd {
	sum, ok := m.session.SelectedSummary()
	if !ok {
		m.session.ClearDetailPreview()
		return nil
	}
	seq := m.session.BeginDetailLoad()
	return loadDetailCmd(m.session, seq, sum)
}

func (m *Model) handleDetailLoaded(msg DetailLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Seq != m.session.DetailGeneration() {
		return m, nil
	}
	if msg.Err != nil {
		m.session.StatusMsg = msg.Err.Error()
		m.session.DetailLoading = false
		return m, nil
	}
	m.session.SetDetail(msg.Seq, msg.Detail)
	return m, nil
}
