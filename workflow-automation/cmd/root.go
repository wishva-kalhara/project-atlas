package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	headedFlag bool
	debugFlag  bool
	indexFlag  int
)

var rootCmd = &cobra.Command{
	Use:   "workflow-exec",
	Short: "Marine Masterpieces Agent Orchestration & UI Automation CLI",
	Long:  "A unified Go CLI for orchestrating AI video creation, metadata enrichment, and multi-platform reel scheduling.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&headedFlag, "headed", true, "Run browser in headed mode (visible UI)")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable verbose debug logs")
	rootCmd.PersistentFlags().IntVar(&indexFlag, "index", 0, "Index of the metadata entry to process (default 0)")
	// Register sub-commands
	rootCmd.AddCommand(scheduleFbCmd)
	rootCmd.AddCommand(videoGenCmd)
	rootCmd.AddCommand(youtubeCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(setupCmd)
}
