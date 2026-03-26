package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "gitlore",
	Short: "Accumulated knowledge for AI-assisted codebases",
	Long:  "gitlore distills agent conversations into semantic commit notes.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(amendNoteCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(postCommitCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gitlore", version)
	},
}
