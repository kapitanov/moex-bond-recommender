package main

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

func init() {
	cmd := &cobra.Command{
		Use:   "suggest",
		Short: "Suggest a portfolio",
		Args:  cobra.ExactArgs(0),
	}

	rootCommand.AddCommand(cmd)

	var (
		postgresConnString, moexURL string
	)
	attachPostgresUrlFlag(cmd, &postgresConnString)
	attachMoexUrlFlag(cmd, &moexURL)
	amount := cmd.Flags().Float64("amount", 1000.0, "amount to invest (RUB)")
	durationStr := cmd.Flags().StringP(
		"duration",
		"d",
		string(recommender.Duration1Year),
		"Bond duration range (1y/2y/3y/4y/5y)")
	partsRaw := cmd.Flags().StringArray("part", []string{}, "define portfolio part (format: COLLECTION_NAME=WEIGHT)")

	parsePart := func(u app.UnitOfWork, partRaw string) (recommender.SuggestRequestPart, error) {
		parts := strings.SplitN(partRaw, "=", 2)

		weight := 1.0
		if len(parts) == 2 {
			w, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return recommender.SuggestRequestPart{}, err
			}
			weight = w
		}

		collection, err := u.GetCollection(parts[0])
		if err != nil {
			return recommender.SuggestRequestPart{}, err
		}

		return recommender.SuggestRequestPart{
			Collection: collection,
			Weight:     weight,
		}, nil
	}

	formatDate := func(v sql.NullTime) string {
		if !v.Valid {
			return ""
		}

		return v.Time.Format("2006-01-02")
	}

	printPortfolio := func(result *recommender.SuggestResult) {
		indent := ""

		// Overview
		table := uitable.New()
		table.RightAlign(2)
		table.AddRow(indent, "Invested", fmt.Sprintf("%0.2f %s", result.Amount, "RUB"))
		table.AddRow(indent, "Duration", fmt.Sprintf("%0.1f years", float64(result.DurationDays)/365.25))
		table.AddRow(indent, "Duration", fmt.Sprintf("%d days", result.DurationDays))
		table.AddRow(indent, "Profit", fmt.Sprintf("%0.2f %s", result.ProfitLoss, "RUB"))
		table.AddRow(indent, "", fmt.Sprintf("%0.2f%%", result.RelativeProfitLoss))
		table.AddRow(indent, "Interest rate", fmt.Sprintf("%0.2f%%", result.InterestRate))
		fmt.Fprintf(os.Stdout, "OVERVIEW\n\n%s\n\n", table)

		// Positions
		table = uitable.New()
		table.RightAlign(2)
		table.RightAlign(3)
		table.RightAlign(4)
		table.RightAlign(5)
		table.RightAlign(6)
		table.RightAlign(7)
		table.AddRow("", "ISIN", "NAME", "MATURITY DATE", "Q", "INVESTED", "PROFIT/LOSS", "INTEREST RATE", "PART IN PORTFOLIO")
		for _, position := range result.Positions {
			table.AddRow(
				indent,
				position.Bond.ISIN,
				position.Bond.ShortName,
				formatDate(position.Bond.MaturityDate),
				fmt.Sprintf("%d", position.Quantity),
				fmt.Sprintf("%0.2f %s", position.OpenValue, position.Currency),
				fmt.Sprintf("%0.2f %s", position.ProfitLoss, position.Currency),
				fmt.Sprintf("%0.2f%%", position.InterestRate),
				fmt.Sprintf("%0.2f%%", position.Weight*100.0))
		}
		fmt.Fprintf(os.Stdout, "POSITIONS\n\n%s\n\n", table)

		// Cash flow
		type CashFlowRow struct {
			Date                                    time.Time
			Amount                                  float64
			HasCoupon, HasAmortization, HasMaturity bool
		}
		cashflows := make(map[time.Time]*CashFlowRow)
		for _, p := range result.Positions {
			for _, c := range p.CashFlow {
				row, exists := cashflows[c.Date]
				if !exists {
					row = &CashFlowRow{Date: c.Date, Amount: 0}
					cashflows[c.Date] = row
				}

				row.Amount += c.ValueRub
				switch c.Type {
				case recommender.Coupon:
					row.HasCoupon = true
				case recommender.Amortization:
					row.HasAmortization = true
				case recommender.Maturity:
					row.HasMaturity = true
				}
			}
		}

		if len(cashflows) > 0 {
			array := make([]*CashFlowRow, len(cashflows))
			i := 0
			for _, r := range cashflows {
				array[i] = r
				i++
			}
			sort.Slice(array, func(i, j int) bool {
				return array[i].Date.Before(array[j].Date)
			})

			table = uitable.New()
			table.RightAlign(6)
			table.AddRow("", "DATE", "COUPON", "AMORTIZATION", "MATURITY", "AMOUNT")

			boolToStr := func(b bool) string {
				if !b {
					return ""
				}

				return "+"
			}

			for _, r := range array {
				types := ""
				if r.HasCoupon {
					types += "C "
				}
				if r.HasAmortization {
					types += "A "
				}
				if r.HasMaturity {
					types += "M "
				}
				table.AddRow(
					indent,
					r.Date.Format("2006-01-02"),
					boolToStr(r.HasCoupon),
					boolToStr(r.HasAmortization),
					boolToStr(r.HasMaturity),
					fmt.Sprintf("%0.2f %s", r.Amount, "RUB"))
			}
			fmt.Fprintf(os.Stdout, "CASH FLOW\n\n%s\n", table)
		}
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		duration, err := parseDuration(*durationStr)
		if err != nil {
			return err
		}

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

		request := &recommender.SuggestRequest{
			Amount:      *amount,
			MaxDuration: duration,
		}

		if partsRaw != nil && len(*partsRaw) > 0 {
			request.Parts = make([]*recommender.SuggestRequestPart, len(*partsRaw))
			for i, partRaw := range *partsRaw {
				part, err := parsePart(u, partRaw)
				if err != nil {
					return err
				}

				request.Parts[i] = &part
			}
		}

		result, err := u.Suggest(request)
		if err != nil {
			return err
		}

		printPortfolio(result)

		return nil
	}
}
