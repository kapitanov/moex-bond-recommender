package main

import (
	"log"

	"github.com/madflojo/tasks"
	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/web"
)

func init() {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run web app",
	}

	rootCommand.AddCommand(cmd)

	var (
		postgresConnString, moexURL, address, googleAnalyticsID string
	)
	attachPostgresUrlFlag(cmd, &postgresConnString)
	attachMoexUrlFlag(cmd, &moexURL)
	attachListenAddressFlag(cmd, &address)
	attachGoogleAnalyticsFlag(cmd, &googleAnalyticsID)
	debugMode := cmd.Flags().Bool("debug", false, "enable debug mode")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := createCancellableContext()

		app, err := app.New(app.WithMoexURL(moexURL), app.WithDataSource(postgresConnString))
		if err != nil {
			return err
		}
		defer app.Close()

		err = app.StartBackgroundTasks()
		if err != nil {
			return err
		}

		webappLogger := log.New(log.Writer(), "web:  ", log.Flags())
		webapp, err := web.New(
			web.WithListenAddress(address),
			web.WithLogger(webappLogger), web.WithApp(app),
			web.WithGoogleAnalyticsID(googleAnalyticsID),
			web.WithDebugMode(*debugMode))
		if err != nil {
			return err
		}

		scheduler := tasks.New()
		defer scheduler.Stop()

		err = webapp.Start()
		if err != nil {
			return err
		}
		defer webapp.Close()

		// Ожидание SIGINT
		appLogger.Printf("web app is running, press <Ctrl+C> to exit")
		<-ctx.Done()

		return nil
	}
}
