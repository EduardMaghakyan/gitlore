package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderView() string {
	if len(m.commits) == 0 {
		return styleHelp.Render("  No commits with gitlore notes found.")
	}

	var b strings.Builder

	// Calculate pane heights
	topHeight := len(m.commits) + 1 // commits + divider
	maxTop := m.height * 40 / 100
	if topHeight > maxTop {
		topHeight = maxTop
	}
	if topHeight < 3 {
		topHeight = 3
	}
	bottomHeight := m.height - topHeight - 2 // 2 for dividers

	// Top divider with help
	var topHelp string
	if m.focus == paneCommits {
		topHelp = styleHelp.Render("↑↓ select  tab detail  q quit")
	} else {
		topHelp = styleHelp.Render("tab commits  q quit")
	}
	topTitle := fmt.Sprintf("COMMITS (%d of %d)", m.cursor+1, len(m.commits))
	b.WriteString(m.renderDivider(topTitle, topHelp))

	// Top pane: compact commit list
	commitLines := topHeight - 1
	start, end := m.commitRange(commitLines)
	for i := start; i < end; i++ {
		c := m.commits[i]
		selected := i == m.cursor

		sha := styleSHA.Render(c.ShortSHA)
		date := styleDate.Render(c.DateRelative)

		// Truncate subject to fit
		subjectWidth := m.width - 14 - lipgloss.Width(c.DateRelative) // sha(7) + spaces + date
		subject := c.Subject
		if len(subject) > subjectWidth && subjectWidth > 3 {
			subject = subject[:subjectWidth-3] + "..."
		}

		line := fmt.Sprintf(" %s  %s", sha, subject)
		// Right-align date
		pad := m.width - lipgloss.Width(line) - lipgloss.Width(date) - 1
		if pad < 1 {
			pad = 1
		}
		line += strings.Repeat(" ", pad) + date

		if selected {
			prefix := styleSHA.Render("▸")
			b.WriteString(prefix + line + "\n")
		} else {
			b.WriteString(" " + line + "\n")
		}
	}

	// Middle divider
	var detailHelp string
	if m.focus == paneDetail {
		detailHelp = styleHelp.Render("↑↓ scroll  ←→ files  tab commits")
	} else {
		detailHelp = styleHelp.Render("tab detail")
	}
	b.WriteString(m.renderDivider("WHY", detailHelp))

	// Bottom pane: fixed note + file tabs, then scrollable diff
	c := m.commits[m.cursor]

	// Fixed: note
	note := c.Note
	for _, tag := range []string{"[agent-assisted]", "[assisted]"} {
		if strings.HasPrefix(note, tag) {
			note = styleTag.Render(tag) + note[len(tag):]
			break
		}
	}
	noteRendered := " " + wrapText(note, m.width-2) + "\n"
	b.WriteString(noteRendered)
	usedLines := strings.Count(noteRendered, "\n")

	// Fixed: file tabs
	if len(c.Files) > 0 {
		b.WriteString("\n")
		usedLines++
		var fileParts []string
		for i, f := range c.Files {
			status := colorFileStatus(f.Status)
			name := f.Path
			if parts := strings.Split(name, "/"); len(parts) > 1 {
				name = parts[len(parts)-1]
			}
			if i == m.fileCursor {
				fileParts = append(fileParts, styleSHA.Render("▸")+" "+status+" "+styleFileSelected.Render(name))
			} else {
				fileParts = append(fileParts, "  "+status+" "+styleFileUnselected.Render(name))
			}
		}
		b.WriteString(" " + strings.Join(fileParts, "   ") + "\n")
		usedLines++
	}

	// Scrollable: diff only
	diffHeight := bottomHeight - usedLines - 1
	if diffHeight < 1 {
		diffHeight = 1
	}

	if len(c.Files) > 0 && m.fileCursor < len(c.Files) {
		f := c.Files[m.fileCursor]
		if f.Diff != "" {
			b.WriteString("\n")
			diffLines := strings.Split(colorizeDiff(f.Diff, m.width-1), "\n")

			// Apply scroll offset to diff only
			if m.scrollOffset > 0 && m.scrollOffset < len(diffLines) {
				diffLines = diffLines[m.scrollOffset:]
			}

			for i := 0; i < diffHeight && i < len(diffLines); i++ {
				b.WriteString(diffLines[i] + "\n")
			}
		}
	}

	return b.String()
}

func (m Model) renderDivider(title, help string) string {
	styledTitle := styleHeader.Render(" " + title + " ")
	lineWidth := m.width - lipgloss.Width(styledTitle) - lipgloss.Width(help) - 1
	if lineWidth < 2 {
		lineWidth = 2
	}
	return styleSectionLine.Render("─") + styledTitle +
		styleSectionLine.Render(strings.Repeat("─", lineWidth)) +
		help + "\n"
}

func (m Model) commitRange(visible int) (int, int) {
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	end := start + visible
	if end > len(m.commits) {
		end = len(m.commits)
	}
	return start, end
}

func colorizeDiff(diff string, width int) string {
	lines := strings.Split(diff, "\n")
	var result []string
	for _, line := range lines {
		if width > 0 && len(line) > width {
			line = line[:width]
		}
		switch {
		case strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "--- "):
			result = append(result, " "+styleDiffHdr.Render(line))
		case strings.HasPrefix(line, "@@"):
			result = append(result, " "+styleDiffHdr.Render(line))
		case strings.HasPrefix(line, "+"):
			result = append(result, " "+styleDiffAdd.Render(line))
		case strings.HasPrefix(line, "-"):
			result = append(result, " "+styleDiffDel.Render(line))
		case strings.HasPrefix(line, "diff "):
			result = append(result, " "+styleDiffHdr.Render(line))
		default:
			result = append(result, " "+line)
		}
	}
	return strings.Join(result, "\n")
}

func colorFileStatus(status string) string {
	switch status {
	case "A":
		return styleFileAdded.Render("A")
	case "D":
		return styleFileDel.Render("D")
	default:
		return styleFileMod.Render(status)
	}
}

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if lipgloss.Width(line+" "+w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}
