package main

import (
	"github.com/spf13/cobra"
)

var fetchCommand = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch commands",
}

func init() {
	rootCommand.AddCommand(fetchCommand)
}
