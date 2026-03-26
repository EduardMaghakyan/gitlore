package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	markerBegin = "# gitlore:begin"
	markerEnd   = "# gitlore:end"
)

var postCommitSnippet = strings.Join([]string{
	markerBegin,
	"gitlore _post-commit >/dev/null 2>&1 &",
	markerEnd,
}, "\n")

var prePushSnippet = strings.Join([]string{
	markerBegin,
	`[ -z "$GITLORE_PUSHING_NOTES" ] && GITLORE_PUSHING_NOTES=1 git push origin refs/notes/commits 2>/dev/null || true`,
	markerEnd,
}, "\n")

// Install sets up gitlore hooks and config in the current repo.
func Install(repoRoot string) error {
	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	if _, err := os.Stat(hooksDir); err != nil {
		return fmt.Errorf("not a git repository (no .git/hooks): %s", repoRoot)
	}

	if err := installHook(hooksDir, "post-commit", postCommitSnippet); err != nil {
		return fmt.Errorf("post-commit hook: %w", err)
	}

	if err := installHook(hooksDir, "pre-push", prePushSnippet); err != nil {
		return fmt.Errorf("pre-push hook: %w", err)
	}

	// Configure git to display notes in log/show
	cmd := exec.Command("git", "config", "--local", "notes.displayRef", "refs/notes/commits")
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git config: %w", err)
	}

	fmt.Println("gitlore installed:")
	fmt.Println("  - post-commit hook (distills conversation -> note)")
	fmt.Println("  - pre-push hook (pushes notes ref)")
	fmt.Println("  - notes.displayRef configured (git log shows notes)")
	return nil
}

func installHook(hooksDir, name, snippet string) error {
	path := filepath.Join(hooksDir, name)

	existing := ""
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	// Remove old gitlore section if present
	cleaned := removeSection(existing, markerBegin, markerEnd)

	var content string
	if cleaned == "" {
		content = "#!/bin/sh\n" + snippet + "\n"
	} else {
		// Ensure shebang exists
		if !strings.HasPrefix(cleaned, "#!") {
			cleaned = "#!/bin/sh\n" + cleaned
		}
		content = strings.TrimRight(cleaned, "\n") + "\n" + snippet + "\n"
	}

	return os.WriteFile(path, []byte(content), 0755)
}

func removeSection(s, begin, end string) string {
	startIdx := strings.Index(s, begin)
	if startIdx == -1 {
		return s
	}
	endIdx := strings.Index(s[startIdx:], end)
	if endIdx == -1 {
		return s
	}
	endIdx += startIdx + len(end)
	// Remove trailing newline if present
	if endIdx < len(s) && s[endIdx] == '\n' {
		endIdx++
	}
	return s[:startIdx] + s[endIdx:]
}
