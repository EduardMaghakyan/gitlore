package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type pane int

const (
	paneCommits pane = iota
	paneDetail
)

type Model struct {
	commits      []Commit
	cursor       int
	fileCursor   int
	focus        pane
	scrollOffset int
	diffLines    int
	width        int
	height       int
	ready        bool
}

func NewModel(commits []Commit) Model {
	m := Model{
		commits: commits,
		focus:   paneCommits,
	}
	// Pre-load detail for first commit
	if len(commits) > 0 {
		LoadCommitDetail(&m.commits[0])
		m.updateDiffLines()
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.focus == paneCommits {
				m.focus = paneDetail
			} else {
				m.focus = paneCommits
			}
		case "up", "k":
			if m.focus == paneCommits {
				if m.cursor > 0 {
					m.cursor--
					m.loadCurrentCommit()
				}
			} else {
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			}
		case "down", "j":
			if m.focus == paneCommits {
				if m.cursor < len(m.commits)-1 {
					m.cursor++
					m.loadCurrentCommit()
				}
			} else {
				if m.diffLines == 0 || m.scrollOffset < m.diffLines-1 {
					m.scrollOffset++
				}
			}
		case "left", "h":
			if m.focus == paneDetail {
				m.switchFile(-1)
			}
		case "right", "l":
			if m.focus == paneDetail {
				m.switchFile(1)
			}
		}
	}
	return m, nil
}

func (m *Model) loadCurrentCommit() {
	c := &m.commits[m.cursor]
	if c.Files == nil {
		LoadCommitDetail(c)
	}
	m.fileCursor = 0
	m.scrollOffset = 0
	m.updateDiffLines()
}

func (m *Model) switchFile(delta int) {
	c := &m.commits[m.cursor]
	next := m.fileCursor + delta
	if next >= 0 && next < len(c.Files) {
		m.fileCursor = next
		m.scrollOffset = 0
		LoadFileDiff(c.SHA, &c.Files[m.fileCursor])
		m.updateDiffLines()
	}
}

func (m *Model) updateDiffLines() {
	c := m.commits[m.cursor]
	if m.fileCursor < len(c.Files) && c.Files[m.fileCursor].Diff != "" {
		m.diffLines = strings.Count(c.Files[m.fileCursor].Diff, "\n") + 1
	} else {
		m.diffLines = 0
	}
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	return m.renderView()
}

// Run starts the TUI.
func Run(commits []Commit) error {
	if len(commits) == 0 {
		fmt.Println("No commits with gitlore notes found.")
		return nil
	}

	p := tea.NewProgram(NewModel(commits), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}
