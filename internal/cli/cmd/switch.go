package cmd

import (
	"fmt"
	"os"

	"github.com/diogenes/omo-profiler/internal/backup"
	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/spf13/cobra"
)

var SwitchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Activate a profile",
	Long:  `Switches to the specified profile by copying it to oh-my-opencode.json. Creates a timestamped backup of the current config before switching.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if !profile.Exists(name) {
			fmt.Fprintf(os.Stderr, "Error: profile not found: %s\n", name)
			os.Exit(1)
		}

		configPath := config.ConfigFile()
		if _, err := os.Stat(configPath); err == nil {
			_, err := backup.Create(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to create backup: %v\n", err)
				os.Exit(1)
			}
		}

		if err := profile.SetActive(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to switch profile: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Switched to profile: %s\n", name)
		os.Exit(0)
	},
}
