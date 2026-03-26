package capture

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxChars = 8000

// entry represents a single JSONL line from Claude Code conversation logs.
type entry struct {
	Type      string    `json:"type"`
	Timestamp string    `json:"timestamp"`
	Message   message   `json:"message"`
	parsed    time.Time // set after parsing
}

type message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Capture reads Claude Code conversation logs for the given repo
// and returns formatted conversation text since the given time.
func Capture(repoRoot string, since time.Time) (string, error) {
	claudeHome, err := claudeDir()
	if err != nil {
		return "", nil
	}

	encoded := encodePath(repoRoot)
	projectDir := filepath.Join(claudeHome, "projects", encoded)

	files, err := findConversationFiles(projectDir)
	if err != nil || len(files) == 0 {
		return "", nil
	}

	var entries []entry
	// Check the 2 most recent files
	limit := 2
	if len(files) < limit {
		limit = len(files)
	}
	for _, f := range files[:limit] {
		e, err := readEntriesSince(f, since)
		if err != nil {
			continue
		}
		entries = append(entries, e...)
	}

	if len(entries) == 0 {
		return "", nil
	}

	// Sort by timestamp ascending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].parsed.Before(entries[j].parsed)
	})

	return formatConversation(entries), nil
}

func claudeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".claude")
	if _, err := os.Stat(dir); err != nil {
		return "", err
	}
	return dir, nil
}

// encodePath converts a filesystem path to Claude Code's directory encoding.
// /Users/foo/projects/bar -> -Users-foo-projects-bar
func encodePath(path string) string {
	return strings.ReplaceAll(path, "/", "-")
}

// findConversationFiles returns JSONL files sorted by modification time (newest first).
func findConversationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	type fileWithTime struct {
		path    string
		modTime time.Time
	}

	var files []fileWithTime
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileWithTime{
			path:    filepath.Join(dir, e.Name()),
			modTime: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})

	result := make([]string, len(files))
	for i, f := range files {
		result[i] = f.path
	}
	return result, nil
}

// readEntriesSince reads a JSONL file and returns entries after the given time.
// Reads from the end of the file for efficiency.
func readEntriesSince(path string, since time.Time) ([]entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read from end: seek to last 512KB (sufficient for recent conversation)
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	readSize := int64(512 * 1024)
	offset := info.Size() - readSize
	if offset < 0 {
		offset = 0
	}
	if offset > 0 {
		f.Seek(offset, 0)
	}

	var entries []entry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 256*1024)

	first := offset > 0 // skip potentially partial first line
	for scanner.Scan() {
		if first {
			first = false
			continue
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var e entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue // skip malformed lines
		}

		if e.Type != "user" && e.Type != "assistant" {
			continue
		}

		t, err := time.Parse(time.RFC3339Nano, e.Timestamp)
		if err != nil {
			continue
		}
		e.parsed = t

		if t.Before(since) {
			continue
		}

		entries = append(entries, e)
	}

	return entries, nil
}

// formatConversation turns entries into a readable conversation string.
func formatConversation(entries []entry) string {
	var b strings.Builder
	totalChars := 0

	// Build from the end (most recent is most relevant), then reverse
	var lines []string
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		var line string

		switch e.Type {
		case "user":
			text := extractUserContent(e.Message.Content)
			if text == "" {
				continue
			}
			line = fmt.Sprintf("USER: %s", text)
		case "assistant":
			text := extractAssistantContent(e.Message.Content)
			if text == "" {
				continue
			}
			line = fmt.Sprintf("ASSISTANT: %s", text)
		}

		if totalChars+len(line) > maxChars {
			break
		}
		totalChars += len(line) + 1
		lines = append(lines, line)
	}

	// Reverse to chronological order
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	for _, l := range lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}

	return b.String()
}

func extractUserContent(raw json.RawMessage) string {
	// User content is a plain string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return truncate(s, 2000)
	}
	return ""
}

func extractAssistantContent(raw json.RawMessage) string {
	// Assistant content is an array of content blocks
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}

	var texts []string
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			texts = append(texts, b.Text)
		}
	}
	return truncate(strings.Join(texts, "\n"), 2000)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
