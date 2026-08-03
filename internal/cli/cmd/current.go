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
	Long:  `Displays the profile whose configuration is currently written at the root of ~/.omo/omo.json.`,
	Run: func(cmd *cobra.Command, args []string) {
		active, err := profile.GetActive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if active.Modified {
			fmt.Fprintln(os.Stderr, "Warning: the root configuration matches no profile")
			fmt.Println("(none)")
			os.Exit(1)
		}
		if active.ProfileName == "" {
			fmt.Println("(none)")
			os.Exit(1)
		}
		fmt.Println(active.ProfileName)
		os.Exit(0)
	},
}
