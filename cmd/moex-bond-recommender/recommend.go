package main

import (
	"github.com/spf13/cobra"
)

var recommendCommand = &cobra.Command{
	Use:              "recommend",
	Short:            "Recommend commands",
	TraverseChildren: true,
}

func init() {
	rootCommand.AddCommand(recommendCommand)
}
