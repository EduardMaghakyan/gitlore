package capture

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEncodePath(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"/Users/foo/projects/bar", "-Users-foo-projects-bar"},
		{"/home/user/code", "-home-user-code"},
		{"/", "-"},
	}
	for _, tt := range tests {
		got := encodePath(tt.input)
		if got != tt.want {
			t.Errorf("encodePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractUserContent(t *testing.T) {
	raw := json.RawMessage(`"hello world"`)
	got := extractUserContent(raw)
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestExtractAssistantContent(t *testing.T) {
	raw := json.RawMessage(`[
		{"type": "thinking", "thinking": "internal thought"},
		{"type": "text", "text": "visible response"},
		{"type": "tool_use", "name": "Read"}
	]`)
	got := extractAssistantContent(raw)
	if got != "visible response" {
		t.Errorf("got %q, want %q", got, "visible response")
	}
}

func TestExtractAssistantContentMultipleText(t *testing.T) {
	raw := json.RawMessage(`[
		{"type": "text", "text": "first"},
		{"type": "text", "text": "second"}
	]`)
	got := extractAssistantContent(raw)
	if got != "first\nsecond" {
		t.Errorf("got %q, want %q", got, "first\nsecond")
	}
}

func TestSkipNonMessageTypes(t *testing.T) {
	dir := t.TempDir()
	f, _ := os.Create(filepath.Join(dir, "test.jsonl"))

	entries := []map[string]interface{}{
		{"type": "file-history-snapshot", "timestamp": "2026-01-01T00:00:01Z"},
		{"type": "user", "timestamp": "2026-01-01T00:00:02Z", "message": map[string]interface{}{"role": "user", "content": "hello"}},
		{"type": "system", "timestamp": "2026-01-01T00:00:03Z"},
	}
	for _, e := range entries {
		b, _ := json.Marshal(e)
		f.Write(b)
		f.Write([]byte("\n"))
	}
	f.Close()

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := readEntriesSince(filepath.Join(dir, "test.jsonl"), since)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Errorf("got %d entries, want 1", len(result))
	}
	if result[0].Type != "user" {
		t.Errorf("got type %q, want %q", result[0].Type, "user")
	}
}

func TestTimestampFiltering(t *testing.T) {
	dir := t.TempDir()
	f, _ := os.Create(filepath.Join(dir, "test.jsonl"))

	entries := []map[string]interface{}{
		{"type": "user", "timestamp": "2026-01-01T00:00:01Z", "message": map[string]interface{}{"role": "user", "content": "old"}},
		{"type": "user", "timestamp": "2026-01-01T12:00:00Z", "message": map[string]interface{}{"role": "user", "content": "new"}},
	}
	for _, e := range entries {
		b, _ := json.Marshal(e)
		f.Write(b)
		f.Write([]byte("\n"))
	}
	f.Close()

	since := time.Date(2026, 1, 1, 6, 0, 0, 0, time.UTC)
	result, err := readEntriesSince(filepath.Join(dir, "test.jsonl"), since)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Errorf("got %d entries, want 1", len(result))
	}
}

func TestTruncation(t *testing.T) {
	s := strings.Repeat("a", 3000)
	got := truncate(s, 2000)
	if len(got) != 2003 { // 2000 + "..."
		t.Errorf("truncated length = %d, want 2003", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("should end with ...")
	}
}

func TestMalformedLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")
	content := `{"type":"user","timestamp":"2026-01-01T00:00:01Z","message":{"role":"user","content":"hello"}}
{broken json
{"type":"user","timestamp":"2026-01-01T00:00:02Z","message":{"role":"user","content":"world"}}
`
	os.WriteFile(path, []byte(content), 0644)

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := readEntriesSince(path, since)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("got %d entries, want 2 (skip malformed)", len(result))
	}
}

func TestNoConversationDir(t *testing.T) {
	result, err := Capture("/nonexistent/path", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestFormatConversation(t *testing.T) {
	entries := []entry{
		{Type: "user", Message: message{Content: json.RawMessage(`"what is 2+2?"`)}, parsed: time.Now()},
		{Type: "assistant", Message: message{Content: json.RawMessage(`[{"type":"text","text":"4"}]`)}, parsed: time.Now()},
	}
	got := formatConversation(entries)
	if !strings.Contains(got, "USER: what is 2+2?") {
		t.Errorf("missing user content in %q", got)
	}
	if !strings.Contains(got, "ASSISTANT: 4") {
		t.Errorf("missing assistant content in %q", got)
	}
}
