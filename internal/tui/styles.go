package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorCyan   = lipgloss.Color("6")
	colorYellow = lipgloss.Color("3")
	colorGreen  = lipgloss.Color("2")
	colorRed    = lipgloss.Color("1")
	colorDim    = lipgloss.Color("8")

	styleSHA = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)

	styleDate = lipgloss.NewStyle().Foreground(colorDim)

	styleTag = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)

	styleNote = lipgloss.NewStyle()

	styleSubject = lipgloss.NewStyle().Foreground(colorDim)

	styleHeader = lipgloss.NewStyle().Bold(true)

	styleSelectedBorder = lipgloss.NewStyle().
				BorderLeft(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(colorCyan).
				PaddingLeft(1)

	styleUnselectedPadding = lipgloss.NewStyle().PaddingLeft(3)

	styleHelp = lipgloss.NewStyle().Foreground(colorDim)

	styleDiffAdd = lipgloss.NewStyle().Foreground(colorGreen)
	styleDiffDel = lipgloss.NewStyle().Foreground(colorRed)
	styleDiffHdr = lipgloss.NewStyle().Foreground(colorCyan)

	styleFileSelected   = lipgloss.NewStyle().Bold(true)
	styleFileUnselected = lipgloss.NewStyle().Foreground(colorDim)

	styleFileAdded = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	styleFileMod   = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	styleFileDel   = lipgloss.NewStyle().Foreground(colorRed).Bold(true)

	styleFocusActive = lipgloss.NewStyle().
				BorderLeft(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(colorCyan)

	styleSectionLine = lipgloss.NewStyle().Foreground(colorDim)
)
