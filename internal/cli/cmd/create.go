package cmd

import (
	"fmt"
	"os"

	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/spf13/cobra"
)

var fromTemplate string

var CreateCmd = &cobra.Command{
	Use:   "create [new-profile-name]",
	Short: "Create a new profile",
	Long:  `Create a new profile block in ~/.omo/omo.json. Use --from to create from an existing template.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if fromTemplate == "" {
			fmt.Fprintln(os.Stderr, "Error: TUI mode not implemented. Use --from to create from template.")
			os.Exit(1)
		}

		if !profile.Exists(fromTemplate) {
			fmt.Fprintf(os.Stderr, "Error: template '%s' not found\n", fromTemplate)
			os.Exit(1)
		}

		var newProfileName string
		if len(args) > 0 {
			newProfileName = profile.SanitizeName(args[0])
		} else {
			fmt.Fprintln(os.Stderr, "Error: new profile name required")
			os.Exit(1)
		}

		if newProfileName == "" {
			fmt.Fprintln(os.Stderr, "Error: invalid profile name")
			os.Exit(1)
		}

		if profile.Exists(newProfileName) {
			fmt.Fprintf(os.Stderr, "Error: profile '%s' already exists\n", newProfileName)
			os.Exit(1)
		}

		// Clones the whole profile block ([opencode] plus its unknown keys and
		// sibling blocks) in one transaction.
		if err := profile.CreateFrom(newProfileName, fromTemplate); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create profile: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created profile '%s' from template '%s'\n", newProfileName, fromTemplate)
		os.Exit(0)
	},
}

func init() {
	CreateCmd.Flags().StringVarP(&fromTemplate, "from", "f", "", "Create from existing template profile")
}
