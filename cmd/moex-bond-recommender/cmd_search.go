package main

import (
	"fmt"
	"os"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/search"
)

func init() {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for bonds",
		Args:  cobra.ExactArgs(1),
	}

	var (
		skip, limit                 int
		postgresConnString, moexURL string
	)
	cmd.Flags().IntVar(&skip, "skip", 0, "How many items to skip")
	cmd.Flags().IntVar(&limit, "limit", 10, "How many items to show")
	attachPostgresUrlFlag(cmd, &postgresConnString)
	attachMoexUrlFlag(cmd, &moexURL)

	rootCommand.AddCommand(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := createCancellableContext()

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

		req := search.Request{
			Text:  args[0],
			Skip:  skip,
			Limit: limit,
		}
		result, err := u.Search(req)
		if err != nil {
			return err
		}

		table := uitable.New()
		table.MaxColWidth = 80
		table.Wrap = true
		table.AddRow("ISIN", "SHORT NAME", "FULL NAME")
		for _, bond := range result.Bonds {
			table.AddRow(bond.ISIN, bond.ShortName, bond.FullName)
		}
		fmt.Fprintf(os.Stdout, "%s\n\nShown %d items out of %d\n", table, len(result.Bonds), result.TotalCount)

		return nil
	}
}
