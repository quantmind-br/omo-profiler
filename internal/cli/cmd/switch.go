package cmd

import (
	"fmt"
	"os"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/spf13/cobra"
)

var SwitchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Apply a profile to ~/.omo/omo.json",
	Long: `Applies a profile by writing its keys into ~/.omo/omo.json.

The profile block replaces the matching keys at the document root, so the
profile is live as soon as the command returns. If the current root
configuration matches no profile it is saved as a new profile first.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		applied, err := profile.Apply(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to apply profile: %v\n", err)
			os.Exit(1)
		}
		if applied.Snapshot != "" {
			fmt.Printf("Saved the previous configuration as profile %q\n", applied.Snapshot)
		}
		fmt.Printf("Applied profile %q to %s\n", applied.Name, config.OmoFile())
		os.Exit(0)
	},
}
