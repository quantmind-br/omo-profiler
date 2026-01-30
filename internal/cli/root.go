package cli

import (
	"fmt"
	"os"

	"github.com/diogenes/omo-profiler/internal/cli/cmd"
	"github.com/diogenes/omo-profiler/internal/tui"
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
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
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
	rootCmd.AddCommand(cmd.ModelsCmd)
	rootCmd.AddCommand(cmd.CreateCmd)
}
