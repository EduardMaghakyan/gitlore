package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/eduardmaghakyan/gitlore/internal/notes"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:                "push",
	Short:              "Push commits and notes",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Push commits
		gitArgs := append([]string{"push"}, args...)
		c := exec.Command("git", gitArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}

		// Push notes
		if err := notes.Push("origin"); err != nil {
			fmt.Fprintln(os.Stderr, "gitlore: warning: could not push notes ref")
		}
		return nil
	},
}
