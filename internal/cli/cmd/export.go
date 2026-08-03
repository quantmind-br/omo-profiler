package cmd

import (
	"fmt"
	"os"

	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/spf13/cobra"
)

var exportForce bool

var ExportCmd = &cobra.Command{
	Use:   "export <name> <path>",
	Short: "Export a profile to a file",
	Long:  `Exports the specified profile block from ~/.omo/omo.json to a flat JSON file at the given path.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		path := args[1]

		if !profile.Exists(name) {
			fmt.Fprintf(os.Stderr, "Error: profile not found: %s\n", name)
			os.Exit(1)
		}

		if !exportForce {
			if _, err := os.Stat(path); err == nil {
				fmt.Fprintf(os.Stderr, "Error: destination file already exists: %s. Use --force to overwrite\n", path)
				os.Exit(1)
			}
		}

		// Export the stored block verbatim so an export/import round-trip is
		// lossless; marshalling the typed Config drops explicit zero values.
		data, err := profile.ExportOpenCode(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load profile: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to write file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Exported profile \"%s\" to %s\n", name, path)
		os.Exit(0)
	},
}

func init() {
	ExportCmd.Flags().BoolVarP(&exportForce, "force", "f", false, "Overwrite destination if it exists")
}
