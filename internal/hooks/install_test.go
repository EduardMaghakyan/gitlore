package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallHookNewFile(t *testing.T) {
	dir := t.TempDir()
	err := installHook(dir, "post-commit", postCommitSnippet)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "post-commit"))
	content := string(data)

	if !strings.HasPrefix(content, "#!/bin/sh\n") {
		t.Error("missing shebang")
	}
	if !strings.Contains(content, "gitlore _post-commit >/dev/null 2>&1 &") {
		t.Error("missing gitlore command")
	}
	if !strings.Contains(content, markerBegin) || !strings.Contains(content, markerEnd) {
		t.Error("missing markers")
	}
}

func TestInstallHookIdempotent(t *testing.T) {
	dir := t.TempDir()

	// Install twice
	installHook(dir, "post-commit", postCommitSnippet)
	installHook(dir, "post-commit", postCommitSnippet)

	data, _ := os.ReadFile(filepath.Join(dir, "post-commit"))
	content := string(data)

	count := strings.Count(content, "gitlore _post-commit >/dev/null 2>&1 &")
	if count != 1 {
		t.Errorf("found %d instances of gitlore command, want 1", count)
	}
}

func TestInstallHookPreservesExisting(t *testing.T) {
	dir := t.TempDir()
	existing := "#!/bin/sh\necho 'existing hook'\n"
	os.WriteFile(filepath.Join(dir, "post-commit"), []byte(existing), 0755)

	installHook(dir, "post-commit", postCommitSnippet)

	data, _ := os.ReadFile(filepath.Join(dir, "post-commit"))
	content := string(data)

	if !strings.Contains(content, "echo 'existing hook'") {
		t.Error("existing hook content was lost")
	}
	if !strings.Contains(content, "gitlore _post-commit >/dev/null 2>&1 &") {
		t.Error("gitlore command not added")
	}
}

func TestRemoveSection(t *testing.T) {
	input := "before\n# gitlore:begin\nstuff\n# gitlore:end\nafter"
	got := removeSection(input, markerBegin, markerEnd)
	if strings.Contains(got, "stuff") {
		t.Error("section not removed")
	}
	if !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Error("surrounding content was lost")
	}
}
