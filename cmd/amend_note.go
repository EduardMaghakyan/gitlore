package cmd

import (
	"github.com/eduardmaghakyan/gitlore/internal/gitutil"
	"github.com/eduardmaghakyan/gitlore/internal/notes"
	"github.com/spf13/cobra"
)

var amendNoteCmd = &cobra.Command{
	Use:   "amend-note [SHA]",
	Short: "Edit the note on a commit",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sha := "HEAD"
		if len(args) > 0 {
			sha = args[0]
		}

		if sha == "HEAD" {
			resolved, err := gitutil.HeadSHA()
			if err != nil {
				return err
			}
			sha = resolved
		}

		return notes.Edit(sha)
	},
}
