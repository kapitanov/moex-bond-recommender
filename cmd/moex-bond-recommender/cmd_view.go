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
		Short: "View bond report",
		Args:  cobra.ExactArgs(1),
	}

	rootCommand.AddCommand(cmd)

	var (
		postgresConnString, moexURL string
	)
	attachPostgresUrlFlag(cmd, &postgresConnString)
	attachMoexUrlFlag(cmd, &moexURL)

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

		report, err := u.GetReport(args[0])
		if err != nil {
			return err
		}

		var formatDate = func(v sql.NullTime) string {
			if !v.Valid {
				return ""
			}

			return v.Time.Format("2006-01-02")
		}

		var formatCashFlowType = func(t recommender.CashFlowItemType) string {
			switch t {
			case recommender.Coupon:
				return "CPN"
			case recommender.Amortization:
				return "AMR"
			case recommender.Maturity:
				return "MAT"
			}
			return string(t)
		}

		overview := uitable.New()
		overview.RightAlign(1)
		overview.RightAlign(3)

		overview.AddRow("OPEN:", "", "CLOSE:", formatDate(report.Bond.MaturityDate))
		overview.AddRow("  Price", fmt.Sprintf("%0.2f%%", report.OpenPrice), "  Coupons", fmt.Sprintf("%0.2f %s", report.CouponPayments, report.Currency))
		overview.AddRow("  Face value", fmt.Sprintf("%0.2f %s", report.OpenFaceValue, report.Currency), "  Amortizations", fmt.Sprintf("%0.2f %s", report.AmortizationPayments, report.Currency))
		overview.AddRow("  Accrued interest", fmt.Sprintf("%0.2f %s", report.OpenAccruedInterest, report.Currency), "  Maturity", fmt.Sprintf("%0.2f %s", report.MaturityPayment, report.Currency))
		overview.AddRow("  Fee", fmt.Sprintf("%0.2f %s", report.OpenFee, report.Currency), "  Revenue", fmt.Sprintf("%0.2f %s", report.Revenue, report.Currency))
		overview.AddRow("  Expenses", fmt.Sprintf("%0.2f %s", report.OpenValue, report.Currency), "  Taxes", fmt.Sprintf("%0.2f %s", report.Taxes, report.Currency))
		overview.AddRow("Days till maturity", "", "", fmt.Sprintf("%d", report.DaysTillMaturity))
		overview.AddRow("Profit/loss", "", "", fmt.Sprintf("%0.2f %s", report.ProfitLoss, report.Currency))
		overview.AddRow("Interest rate", "", "", fmt.Sprintf("%0.2f%%", report.InterestRate))

		table := uitable.New()
		table.AddRow("DATE", "TYPE", "VALUE")
		table.RightAlign(2)
		for _, item := range report.CashFlow {
			table.AddRow(
				item.Date.Format("2006-01-05"),
				formatCashFlowType(item.Type),
				fmt.Sprintf("%0.2f %s", item.ValueRub, "RUB"))
		}
		fmt.Fprintf(
			os.Stdout,
			"%s \"%s\"\n%s\n\n%s\n\n%s\n",
			report.Bond.ISIN,
			report.Bond.FullName,
			report.Issuer.Name,
			overview,
			table)

		return nil
	}
}
