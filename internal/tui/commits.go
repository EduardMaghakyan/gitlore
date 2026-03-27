package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Commit struct {
	SHA          string
	ShortSHA     string
	Author       string
	Date         time.Time
	DateRelative string
	Subject      string
	Note         string
	Body         string       // loaded lazily
	Files        []FileChange // loaded lazily
}

type FileChange struct {
	Status string // A, M, D, R, etc.
	Path   string
	Diff   string // loaded when selected
}

const recordMarker = "---GITLORE---"

// LoadCommits loads recent commits that have gitlore notes.
func LoadCommits(limit int) ([]Commit, error) {
	// Use per-line fields with a record separator marker.
	// %N (notes) can be multiline, so we place it last and delimit records.
	format := fmt.Sprintf("%%H%%n%%h%%n%%an%%n%%aI%%n%%cr%%n%%s%%n%%N%s", recordMarker)
	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", limit),
		"--notes=refs/notes/commits", fmt.Sprintf("--format=%s", format))
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	records := strings.Split(string(out), recordMarker)
	var commits []Commit
	for _, rec := range records {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}
		// Split into lines: first 6 are fixed fields, rest is the note
		lines := strings.SplitN(rec, "\n", 7)
		if len(lines) < 6 {
			continue
		}

		note := ""
		if len(lines) >= 7 {
			note = strings.TrimSpace(lines[6])
		}
		if note == "" {
			continue // only show commits with notes
		}

		date, _ := time.Parse(time.RFC3339, lines[3])
		commits = append(commits, Commit{
			SHA:          lines[0],
			ShortSHA:     lines[1],
			Author:       lines[2],
			Date:         date,
			DateRelative: lines[4],
			Subject:      lines[5],
			Note:         note,
		})
	}
	return commits, nil
}

// LoadCommitDetail loads the full body and changed files for a commit.
func LoadCommitDetail(c *Commit) error {
	// Load full body
	cmd := exec.Command("git", "log", "-1", "--format=%B", c.SHA)
	out, err := cmd.Output()
	if err == nil {
		c.Body = strings.TrimSpace(string(out))
	}

	// Load changed files
	cmd = exec.Command("git", "diff-tree", "--no-commit-id", "--name-status", "-r", c.SHA)
	out, err = cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	c.Files = nil
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		c.Files = append(c.Files, FileChange{
			Status: parts[0],
			Path:   parts[1],
		})
	}

	// Auto-load first file's diff
	if len(c.Files) > 0 {
		LoadFileDiff(c.SHA, &c.Files[0])
	}

	return nil
}

// LoadFileDiff loads the diff for a specific file in a commit.
func LoadFileDiff(sha string, f *FileChange) {
	if f.Diff != "" {
		return // already loaded
	}
	cmd := exec.Command("git", "diff", sha+"~1.."+sha, "--", f.Path)
	out, err := cmd.Output()
	if err != nil {
		// First commit or other issue — try show
		cmd = exec.Command("git", "show", sha, "--", f.Path)
		out, _ = cmd.Output()
	}
	f.Diff = string(out)
}
