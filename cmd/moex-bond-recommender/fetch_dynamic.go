package main

import (
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "dynamic",
		Short: "fetchService market data only",
	}

	fetchCommand.AddCommand(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		app, err := NewApp()
		if err != nil {
			return err
		}

		ctx := CreateCancellableContext()

		err = app.FetchMarketData(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}
