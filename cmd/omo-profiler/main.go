package main

import (
	"os"

	"github.com/diogenes/omo-profiler/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
