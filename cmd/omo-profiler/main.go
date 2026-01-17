package main

import (
	"os"

	"github.com/diogenes/omo-profiler/internal/cli"

	// Dependencies for future implementation
	_ "github.com/charmbracelet/bubbles/list"
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/lipgloss"
	_ "github.com/sergi/go-diff/diffmatchpatch"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/xeipuuv/gojsonschema"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
