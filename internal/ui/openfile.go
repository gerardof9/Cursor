package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// OpenFileModel is the in-session file path modal.
type OpenFileModel struct {
	input   textinput.Model
	cancel  bool
}

// NewOpenFileModel creates the open-file overlay.
func NewOpenFileModel() OpenFileModel {
	ti := textinput.New()
	ti.Placeholder = "/path/to/mysql-bin.000001"
	ti.CharLimit = 512
	ti.Width = 60
	ti.Prompt = "Path: "
	ti.Focus()
	return OpenFileModel{input: ti}
}

// Update handles modal input. Returns done=true when Enter/Esc pressed.
func (m *OpenFileModel) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.cancel = false
			return nil, true
		case "esc":
			m.cancel = true
			m.input.SetValue("")
			return nil, true
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd, false
}

// Cancelled reports esc dismiss without apply.
func (m OpenFileModel) Cancelled() bool {
	return m.cancel
}

// Value returns entered path.
func (m OpenFileModel) Value() string {
	return m.input.Value()
}

// Render draws the modal.
func (m OpenFileModel) Render() string {
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	return box.Render("Open binlog file\n" + m.input.View())
}

// Reset clears input for next open.
func (m *OpenFileModel) Reset() {
	m.input.SetValue("")
	m.cancel = false
	m.input.Focus()
}
