package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/diogenes/omo-profiler/internal/models"
	"github.com/spf13/cobra"
)

var ModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Manage registered models",
	Long:  `Manage the registry of AI models that can be used in profiles.`,
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered models",
	RunE: func(cmd *cobra.Command, args []string) error {
		registry, err := models.Load()
		if err != nil {
			return fmt.Errorf("failed to load models: %w", err)
		}

		groups := registry.ListByProvider()
		if len(groups) == 0 {
			fmt.Println("(No models registered)")
			return nil
		}

		totalCount := 0
		for _, group := range groups {
			providerName := group.Provider
			if providerName == "" {
				providerName = "Other"
			}
			fmt.Println(providerName)
			for _, m := range group.Models {
				fmt.Printf("  %s (%s)\n", m.DisplayName, m.ModelID)
				totalCount++
			}
			fmt.Println()
		}
		fmt.Printf("(%d models total)\n", totalCount)
		return nil
	},
}

var modelsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new model",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Display name: ")
		displayName, _ := reader.ReadString('\n')
		displayName = strings.TrimSpace(displayName)
		if displayName == "" {
			return fmt.Errorf("display name is required")
		}

		fmt.Print("Model ID: ")
		modelId, _ := reader.ReadString('\n')
		modelId = strings.TrimSpace(modelId)
		if modelId == "" {
			return fmt.Errorf("model ID is required")
		}

		fmt.Print("Provider: ")
		provider, _ := reader.ReadString('\n')
		provider = strings.TrimSpace(provider)

		registry, err := models.Load()
		if err != nil {
			return fmt.Errorf("failed to load models: %w", err)
		}

		newModel := models.RegisteredModel{
			DisplayName: displayName,
			ModelID:     modelId,
			Provider:    provider,
		}

		if err := registry.Add(newModel); err != nil {
			return err
		}

		fmt.Println("✓ Model added")
		return nil
	},
}

var modelsEditCmd = &cobra.Command{
	Use:   "edit <modelId>",
	Short: "Edit an existing model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelId := args[0]

		registry, err := models.Load()
		if err != nil {
			return fmt.Errorf("failed to load models: %w", err)
		}

		existing := registry.Get(modelId)
		if existing == nil {
			return fmt.Errorf("model '%s' not found", modelId)
		}

		fmt.Printf("Editing model: %s (%s)\n\n", existing.DisplayName, existing.ModelID)

		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Display name [%s]: ", existing.DisplayName)
		displayName, _ := reader.ReadString('\n')
		displayName = strings.TrimSpace(displayName)
		if displayName == "" {
			displayName = existing.DisplayName
		}

		fmt.Printf("Model ID [%s]: ", existing.ModelID)
		newModelId, _ := reader.ReadString('\n')
		newModelId = strings.TrimSpace(newModelId)
		if newModelId == "" {
			newModelId = existing.ModelID
		}

		fmt.Printf("Provider [%s]: ", existing.Provider)
		provider, _ := reader.ReadString('\n')
		provider = strings.TrimSpace(provider)
		if provider == "" {
			provider = existing.Provider
		}

		updatedModel := models.RegisteredModel{
			DisplayName: displayName,
			ModelID:     newModelId,
			Provider:    provider,
		}

		if err := registry.Update(modelId, updatedModel); err != nil {
			return err
		}

		fmt.Println("✓ Model updated")
		return nil
	},
}

var modelsDeleteCmd = &cobra.Command{
	Use:   "delete <modelId>",
	Short: "Delete a model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelId := args[0]

		registry, err := models.Load()
		if err != nil {
			return fmt.Errorf("failed to load models: %w", err)
		}

		existing := registry.Get(modelId)
		if existing == nil {
			return fmt.Errorf("model '%s' not found", modelId)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Delete '%s'? (y/n): ", existing.DisplayName)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled")
			return nil
		}

		if err := registry.Delete(modelId); err != nil {
			return err
		}

		fmt.Println("✓ Model deleted")
		return nil
	},
}

func init() {
	ModelsCmd.AddCommand(modelsListCmd)
	ModelsCmd.AddCommand(modelsAddCmd)
	ModelsCmd.AddCommand(modelsEditCmd)
	ModelsCmd.AddCommand(modelsDeleteCmd)
}
