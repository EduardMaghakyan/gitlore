package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:                "log",
	Short:              "Show git log with notes",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		gitArgs := append([]string{"log", "--notes=refs/notes/commits"}, args...)
		c := exec.Command("git", gitArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return c.Run()
	},
}
