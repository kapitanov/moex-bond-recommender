package main

import (
	"time"

	"github.com/madflojo/tasks"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Run fetch scheduler",
	}

	runCommand.AddCommand(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		app, err := NewApp()
		if err != nil {
			return err
		}

		ctx := CreateCancellableContext()

		appLogger.Printf("fetching static data (initial fetch)")
		err = app.FetchStaticData(ctx)
		if err != nil {
			return err
		}

		appLogger.Printf("fetching market data (initial fetch)")
		err = app.FetchMarketData(ctx)
		if err != nil {
			return err
		}

		scheduler := tasks.New()
		defer scheduler.Stop()

		// Каждые 15 минут выполняется выгрузка маркетдаты
		fetchMarketDataInterval := 15 * time.Minute
		_, err = scheduler.Add(&tasks.Task{
			Interval: fetchMarketDataInterval,
			TaskFunc: func() error {
				return app.FetchMarketData(ctx)
			},
		})
		if err != nil {
			return err
		}
		appLogger.Printf("will fetch market data every %s", fetchMarketDataInterval)

		// Каждый день в 9:00 (MSK) выполняется выгрузка данных по облигациям
		y, m, d := time.Now().UTC().Date()
		tz, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			return err
		}
		startTime := time.Date(y, m, d, 9, 00, 00, 000, tz)
		startTime = startTime.Add(-24 * time.Hour)
		fetchStaticDataInterval := 24 * time.Hour
		_, err = scheduler.Add(&tasks.Task{
			Interval:   fetchStaticDataInterval,
			StartAfter: startTime,
			TaskFunc: func() error {
				return app.FetchStaticData(ctx)
			},
		})
		if err != nil {
			return err
		}
		appLogger.Printf("will fetch static data every %s starting %s", fetchStaticDataInterval, startTime)

		// Ожидание SIGINT
		appLogger.Printf("foreground fetch is running, press <Ctrl+C> to exit")
		<-ctx.Done()

		return nil
	}
}
