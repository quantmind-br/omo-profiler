package cmd

import (
	"io"
	"testing"
)

func TestSchemaCheckCmd_RequiresOutputFlag(t *testing.T) {
	SchemaCheckCmd.SetArgs([]string{})

	SchemaCheckCmd.SetOut(io.Discard)
	SchemaCheckCmd.SetErr(io.Discard)

	err := SchemaCheckCmd.Execute()
	if err == nil {
		t.Error("Expected error when --output flag is missing, but got nil")
	}
}

func TestSchemaCheckCmd_Registration(t *testing.T) {
	if SchemaCheckCmd.Use != "schema-check" {
		t.Errorf("Expected command Use to be 'schema-check', got %q", SchemaCheckCmd.Use)
	}

	flag := SchemaCheckCmd.Flags().Lookup("output")
	if flag == nil {
		t.Error("Expected 'output' flag to be defined")
	}

	if !SchemaCheckCmd.Flags().Changed("output") && flag.DefValue != "" {
		t.Errorf("Expected default value of 'output' to be empty, got %q", flag.DefValue)
	}
}
