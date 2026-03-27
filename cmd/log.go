package cmd

import (
	"github.com/eduardmaghakyan/gitlore/internal/tui"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Browse commits with gitlore notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		commits, err := tui.LoadCommits(100)
		if err != nil {
			return err
		}
		return tui.Run(commits)
	},
}
