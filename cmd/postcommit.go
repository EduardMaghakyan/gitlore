package cmd

import (
	"fmt"
	"os"

	"github.com/eduardmaghakyan/gitlore/internal/hooks"
	"github.com/spf13/cobra"
)

var postCommitCmd = &cobra.Command{
	Use:    "_post-commit",
	Short:  "Post-commit hook handler (internal)",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := hooks.RunPostCommit(); err != nil {
			fmt.Fprintf(os.Stderr, "gitlore: %v\n", err)
		}
	},
}
