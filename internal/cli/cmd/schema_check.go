package cmd

import (
	"fmt"
	"os"

	"github.com/diogenes/omo-profiler/internal/schema"
	"github.com/spf13/cobra"
)

var schemaCheckOutput string

var SchemaCheckCmd = &cobra.Command{
	Use:   "schema-check",
	Short: "Check if embedded schema differs from upstream",
	Long:  `Compares the embedded schema with the upstream version and generates a diff file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if schemaCheckOutput == "" {
			fmt.Fprintln(os.Stderr, "Error: --output flag is required")
			os.Exit(1)
		}

		result, err := schema.CompareSchemas()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if result.Identical {
			fmt.Println("Schemas are identical")
			os.Exit(0)
		}

		path, err := schema.SaveDiff(schemaCheckOutput, result.Diff)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Diff saved to: %s\n", path)
		os.Exit(0)
	},
}

func init() {
	SchemaCheckCmd.Flags().StringVarP(&schemaCheckOutput, "output", "o", "", "Directory to save diff file (required)")
	SchemaCheckCmd.MarkFlagRequired("output")
}
