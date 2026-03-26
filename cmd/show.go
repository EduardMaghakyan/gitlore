package cmd

import (
	"fmt"
	"os"

	"github.com/eduardmaghakyan/gitlore/internal/gitutil"
	"github.com/eduardmaghakyan/gitlore/internal/notes"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show [SHA]",
	Short: "Show the note on a commit",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sha := "HEAD"
		if len(args) > 0 {
			sha = args[0]
		}

		// Resolve SHA
		resolved, err := gitutil.HeadSHA()
		if sha != "HEAD" {
			resolved = sha
		}
		if err != nil {
			return err
		}

		text, err := notes.Show(resolved)
		if err != nil {
			fmt.Fprintf(os.Stderr, "no gitlore note on %s\n", sha)
			return nil
		}

		fmt.Println(text)
		return nil
	},
}
