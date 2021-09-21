package main

import (
	"fmt"
	"os"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
)

func init() {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List bond collections",
	}
	recommendCommand.AddCommand(cmd)

	var (
		postgresConnString, moexURL string
	)
	attachPostgresUrlFlag(cmd, &postgresConnString)
	attachMoexUrlFlag(cmd, &moexURL)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := CreateCancellableContext()

		app, err := app.New(app.WithMoexURL(moexURL), app.WithDataSource(postgresConnString))
		if err != nil {
			return err
		}
		defer app.Close()

		u, err := app.NewUnitOfWork(ctx)
		if err != nil {
			return err
		}
		defer u.Close()

		collections := u.ListCollections()

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
