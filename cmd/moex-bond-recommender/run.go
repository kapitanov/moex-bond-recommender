package main

import (
	"github.com/spf13/cobra"
)

var runCommand = &cobra.Command{
	Use:   "run",
	Short: "Run commands",
}

func init() {
	rootCommand.AddCommand(runCommand)
}
