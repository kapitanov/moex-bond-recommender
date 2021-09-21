package main

import (
	"log"
	"time"

	"github.com/madflojo/tasks"
	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/web"
)

func init() {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Run web app",
	}

	runCommand.AddCommand(cmd)

	var address string
	cmd.Flags().StringVarP(&address, "address", "a", string(web.DefaultAddress), "Web app listen address")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		app, err := NewApp()
		if err != nil {
			return err
		}

		ctx := CreateCancellableContext()

		webappLogger := log.New(log.Writer(), "web:  ", log.Flags())
		webapp, err := web.New(web.WithListenAddress(address), web.WithLogger(webappLogger), web.WithApp(app))
		if err != nil {
			return err
		}
/*
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
 */

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

		err = webapp.Start(ctx)
		if err != nil {
			return err
		}

		// Ожидание SIGINT
		appLogger.Printf("web app is running, press <Ctrl+C> to exit")
		<-ctx.Done()

		return nil
	}
}
