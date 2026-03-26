package cmd

import (
	"fmt"
	"os"

	"github.com/eduardmaghakyan/gitlore/internal/gitutil"
	"github.com/eduardmaghakyan/gitlore/internal/hooks"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install gitlore hooks into the current repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitutil.RepoRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, "not inside a git repository")
			os.Exit(1)
		}
		return hooks.Install(root)
	},
}
