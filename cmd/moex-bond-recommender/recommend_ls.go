package main

import (
	"fmt"
	"os"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List collections",
	}

	recommendCommand.AddCommand(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		app, err := NewApp()
		if err != nil {
			return err
		}

		collections := app.ListCollections()

		table := uitable.New()
		table.MaxColWidth = 80
		table.Wrap = true
		table.AddRow("ID", "NAME")
		for _, collection := range collections {
			table.AddRow(collection.ID(), collection.Name())
		}
		fmt.Fprintf(os.Stdout, "%s\n", table)

		return nil
	}
}
