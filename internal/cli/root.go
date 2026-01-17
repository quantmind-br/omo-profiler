package cli

import (
	"github.com/diogenes/omo-profiler/internal/cli/cmd"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:     "omo-profiler",
	Short:   "TUI profile manager for oh-my-opencode",
	Long:    `omo-profiler is a TUI application for managing oh-my-opencode configuration profiles.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Launch TUI when no subcommand is provided
		cmd.Help()
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(cmd.ListCmd)
	rootCmd.AddCommand(cmd.CurrentCmd)
	rootCmd.AddCommand(cmd.ExportCmd)
	rootCmd.AddCommand(cmd.SwitchCmd)
	rootCmd.AddCommand(cmd.ImportCmd)
}
