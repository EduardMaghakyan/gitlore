package hooks

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduardmaghakyan/gitlore/internal/capture"
	"github.com/eduardmaghakyan/gitlore/internal/config"
	"github.com/eduardmaghakyan/gitlore/internal/distill"
	"github.com/eduardmaghakyan/gitlore/internal/gitutil"
	"github.com/eduardmaghakyan/gitlore/internal/notes"
)

// RunPostCommit is the main orchestration for the post-commit hook.
// It captures conversation, distills it, and attaches a note.
func RunPostCommit() error {
	sha, err := gitutil.HeadSHA()
	if err != nil {
		return err
	}

	repoRoot, err := gitutil.RepoRoot()
	if err != nil {
		return err
	}

	since, err := gitutil.PreviousCommitTime()
	if err != nil {
		return err
	}

	conversation, err := capture.Capture(repoRoot, since)
	if err != nil {
		return err
	}

	if conversation == "" {
		return nil // no agent conversation — nothing to attach
	}

	cfg, err := config.Load(repoRoot)
	if err != nil {
		return err
	}

	diff, err := gitutil.DiffHead()
	if err != nil {
		diff = "" // non-fatal — distill without diff
	}

	summary := distill.Distill(conversation, diff, cfg)

	tag := detectTag(sha)
	note := fmt.Sprintf("%s %s", tag, summary)

	if err := notes.Add(sha, note); err != nil {
		return err
	}

	// Print to terminal so the user sees it
	fmt.Fprintf(os.Stderr, "\n\033[36mgitlore\033[0m %s\n%s\n", tag, summary)
	return nil
}

// detectTag checks if the commit was agent-assisted.
func detectTag(sha string) string {
	msg, err := gitutil.CommitMessage(sha)
	if err != nil {
		return "[assisted]"
	}
	if strings.Contains(msg, "Co-Authored-By: Claude") {
		return "[agent-assisted]"
	}
	return "[assisted]"
}
