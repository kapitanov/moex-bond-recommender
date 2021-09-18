package main

import (
	"fmt"
	"os"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/search"
)

func init() {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search bonds",
		Args:  cobra.ExactArgs(1),
	}

	var (
		skip, limit int
	)

	cmd.Flags().IntVar(&skip, "skip", 0, "How many items to skip")
	cmd.Flags().IntVar(&limit, "limit", 10, "How many items to show")

	rootCommand.AddCommand(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		app, err := NewApp()
		if err != nil {
			return err
		}

		req := search.Request{
			Text:  args[0],
			Skip:  skip,
			Limit: limit,
		}
		result, err := app.ExecSearch(req)
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
