package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/diogenes/omo-profiler/internal/schema"
	"github.com/spf13/cobra"
)

var importName string

var ImportCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "Import a profile from a JSON file",
	Long:  `Imports a flat JSON config as a profile block in ~/.omo/omo.json. The file must conform to the omo config schema ([opencode] / flat-config shape).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sourcePath := args[0]

		data, err := os.ReadFile(sourcePath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", sourcePath)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}

		// Type-check the payload before storing it verbatim; the file itself
		// is what gets written, so explicit empty values survive.
		if err := json.Unmarshal(data, &config.Config{}); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid JSON: %v\n", err)
			os.Exit(1)
		}

		validator, err := schema.GetValidator()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create validator: %v\n", err)
			os.Exit(1)
		}

		validationErrors, err := validator.ValidateJSONForSave(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: validation failed: %v\n", err)
			os.Exit(2)
		}
		if len(validationErrors) > 0 {
			fmt.Fprintln(os.Stderr, "Error: validation failed:")
			for _, ve := range validationErrors {
				fmt.Fprintf(os.Stderr, "  - %s: %s\n", ve.Path, ve.Message)
			}
			os.Exit(2)
		}

		var originalName string
		var profileName string

		if importName != "" {
			originalName = importName
			profileName = profile.SanitizeName(importName)
		} else {
			filename := filepath.Base(sourcePath)
			originalName = strings.TrimSuffix(filename, ".json")
			profileName = profile.SanitizeName(originalName)
		}

		if profileName == "" {
			fmt.Fprintf(os.Stderr, "Error: cannot derive valid profile name from filename %q. Use --name <name> to specify.\n", originalName)
			os.Exit(1)
		}

		// Name selection and the write share one transaction.
		baseName := profileName
		profileName, hadCollision, err := profile.CreateAvailable(baseName, data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to save profile: %v\n", err)
			os.Exit(1)
		}

		if hadCollision {
			fmt.Printf("Profile %q exists, imported as %q\n", baseName, profileName)
		} else {
			fmt.Printf("Imported profile: %s\n", profileName)
		}
		os.Exit(0)
	},
}

func init() {
	ImportCmd.Flags().StringVarP(&importName, "name", "n", "", "Name for the imported profile")
}
