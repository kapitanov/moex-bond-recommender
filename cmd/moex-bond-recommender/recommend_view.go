package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

func init() {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View bond collection",
		Args:  cobra.ExactArgs(1),
	}

	recommendCommand.AddCommand(cmd)

	var (
		postgresConnString, moexURL string
	)
	attachPostgresUrlFlag(cmd, &postgresConnString)
	attachMoexUrlFlag(cmd, &moexURL)

	var durationStr string
	cmd.Flags().StringVarP(
		&durationStr,
		"duration",
		"d",
		string(recommender.Duration1Year),
		"Bond duration range (1y/2y/3y/4y/5y)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var duration recommender.Duration
		switch durationStr {
		case "1y":
			duration = recommender.Duration1Year
			break
		case "2y":
			duration = recommender.Duration2Year
			break
		case "3y":
			duration = recommender.Duration3Year
			break
		case "4y":
			duration = recommender.Duration4Year
			break
		case "5y":
			duration = recommender.Duration5Year
			break
		default:
			return fmt.Errorf("\"%s\" is not a valid duration range, valid values are: 1y, 2y, 3y, 4y, 5y", durationStr)
		}

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

		collection, err := u.GetCollection(args[0])
		if err != nil {
			return err
		}

		reports, err := u.ListCollectionBonds(collection.ID(), duration)
		if err != nil {
			return err
		}

		var formatDate = func(v sql.NullTime) string {
			if !v.Valid {
				return ""
			}

			return v.Time.Format("2006-01-02")
		}

		table := uitable.New()
		table.RightAlign(2)
		table.RightAlign(3)
		table.RightAlign(4)
		table.RightAlign(5)
		table.RightAlign(6)
		table.AddRow("ISIN", "NAME", "MATURITY DATE", "PRICE", "OPEN VALUE", "PROFIT/LOSS", "INTEREST RATE")
		for _, report := range reports {
			table.AddRow(
				report.Bond.ISIN,
				report.Bond.ShortName,
				formatDate(report.Bond.MaturityDate),
				fmt.Sprintf("%0.2f%%", report.OpenPrice),
				fmt.Sprintf("%0.2f %s", report.OpenValue, report.Currency),
				fmt.Sprintf("%0.2f %s", report.ProfitLoss, report.Currency),
				fmt.Sprintf("%0.2f%%", report.InterestRate))
		}
		fmt.Fprintf(os.Stdout, "%s (%s)\n\n%s\n", collection.Name(), duration, table)

		return nil
	}
}
