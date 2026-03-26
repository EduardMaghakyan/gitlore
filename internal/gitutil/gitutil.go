package gitutil

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", args[0], strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return strings.TrimSpace(string(out)), nil
}

func RepoRoot() (string, error) {
	return run("rev-parse", "--show-toplevel")
}

func HeadSHA() (string, error) {
	return run("rev-parse", "HEAD")
}

func CommitMessage(sha string) (string, error) {
	return run("log", "-1", "--format=%B", sha)
}

func CommitTime(sha string) (time.Time, error) {
	out, err := run("log", "-1", "--format=%aI", sha)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, out)
}

func PreviousCommitTime() (time.Time, error) {
	t, err := CommitTime("HEAD~1")
	if err != nil {
		return time.Unix(0, 0), nil
	}
	return t, nil
}

func Diff(from, to string) (string, error) {
	return run("diff", from+".."+to)
}

func DiffHead() (string, error) {
	out, err := run("diff", "HEAD~1..HEAD")
	if err != nil {
		// First commit — diff against empty tree
		return run("diff", "--cached", "4b825dc642cb6eb9a060e54bf899d69f82cf7108", "HEAD")
	}
	return out, nil
}

func IsInsideWorkTree() bool {
	out, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}
