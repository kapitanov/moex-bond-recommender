package main

import (
	"github.com/spf13/cobra"
)

var fetchCommand = &cobra.Command{
	Use:              "fetch",
	Short:            "fetchService commands",
	TraverseChildren: true,
}

func init() {
	rootCommand.AddCommand(fetchCommand)
}
