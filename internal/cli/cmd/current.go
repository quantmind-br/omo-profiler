package cmd

import (
	"fmt"
	"os"

	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/spf13/cobra"
)

var CurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the name of the active profile",
	Long:  `Displays the name of the currently active oh-my-opencode profile.`,
	Run: func(cmd *cobra.Command, args []string) {
		active, err := profile.GetActive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if !active.Exists {
			fmt.Println("(none)")
			os.Exit(1)
		}

		if active.IsOrphan {
			fmt.Println("(custom - unsaved)")
			os.Exit(0)
		}

		fmt.Println(active.ProfileName)
		os.Exit(0)
	},
}
