package cmd

import (
	"fmt"

	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Long:  `Lists all available profiles. The active profile is marked with an asterisk (*).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := profile.List()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		if len(profiles) == 0 {
			fmt.Println("(No profiles found)")
			return nil
		}

		active, err := profile.GetActive()
		if err != nil {
			return fmt.Errorf("failed to get active profile: %w", err)
		}

		activeProfileName := ""
		if active.Exists && !active.IsOrphan {
			activeProfileName = active.ProfileName
		}

		for _, name := range profiles {
			if name == activeProfileName {
				fmt.Printf("* %s (active)\n", name)
			} else {
				fmt.Printf("  %s\n", name)
			}
		}

		return nil
	},
}
