package notes

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Add(sha, text string) error {
	cmd := exec.Command("git", "notes", "add", "-f", "-m", text, sha)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Show(sha string) (string, error) {
	cmd := exec.Command("git", "notes", "show", sha)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no note on %s", sha)
	}
	return strings.TrimSpace(string(out)), nil
}

func Edit(sha string) error {
	text, err := Show(sha)
	if err != nil {
		text = ""
	}

	tmp, err := os.CreateTemp("", "gitlore-note-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(text); err != nil {
		return err
	}
	tmp.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	editorCmd := exec.Command(editor, tmp.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	newText, err := os.ReadFile(tmp.Name())
	if err != nil {
		return err
	}

	return Add(sha, strings.TrimSpace(string(newText)))
}

func Push(remote string) error {
	if remote == "" {
		remote = "origin"
	}
	cmd := exec.Command("git", "push", remote, "refs/notes/commits")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
