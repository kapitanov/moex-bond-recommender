package main

import (
	"github.com/spf13/cobra"
)

var runCommand = &cobra.Command{
	Use:              "run",
	Short:            "Run commands",
	TraverseChildren: true,
}

func init() {
	rootCommand.AddCommand(runCommand)
}
