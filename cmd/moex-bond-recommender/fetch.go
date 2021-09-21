package main

import (
	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
)

func init() {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch data and exit",
	}
	rootCommand.AddCommand(cmd)

	var (
		postgresConnString, moexURL      string
		fetchStaticData, fetchMarketData bool
	)
	attachPostgresUrlFlag(cmd, &postgresConnString)
	attachMoexUrlFlag(cmd, &moexURL)
	cmd.Flags().BoolVarP(&fetchStaticData, "static", "s", false, "Fetch static data")
	cmd.Flags().BoolVarP(&fetchMarketData, "market", "m", false, "Fetch market data")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := CreateCancellableContext()

		app, err := app.New(app.WithMoexURL(moexURL), app.WithDataSource(postgresConnString))
		if err != nil {
			return err
		}
		defer app.Close()

		if fetchStaticData || !fetchStaticData && !fetchMarketData {
			err = app.FetchStaticData(ctx)
			if err != nil {
				return err
			}
		}

		if fetchMarketData || !fetchStaticData && !fetchMarketData {
			err = app.FetchMarketData(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
