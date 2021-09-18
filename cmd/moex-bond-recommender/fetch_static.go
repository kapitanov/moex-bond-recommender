package main

import (
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "static",
		Short: "Fetch static data only",
	}

	fetchCommand.AddCommand(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		app, err := NewApp()
		if err != nil {
			return err
		}

		ctx := createCancellableContext()

		err = app.FetchStaticData(ctx)
		if err != nil {
			return err
		}

		return nil
	}
}
