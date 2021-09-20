package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

var rootCommand = &cobra.Command{
	Use:              "moex-bond-recommender",
	TraverseChildren: true,
	SilenceUsage:     true,
}

var (
	postgresConnString string
	moexURL            string
	quietMode          bool
)

var appLogger *log.Logger

func init() {
	rootCommand.PersistentFlags().StringVarP(&postgresConnString, "psql", "p", data.DefaultDataSource, "Postgres connnection string")
	rootCommand.PersistentFlags().StringVar(&moexURL, "moex", moex.DefaultURL, "ISS root URL")
	rootCommand.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "Suppress logging")

	rootCommand.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		log.SetFlags(log.LstdFlags | log.Lmsgprefix)
		if quietMode {
			log.SetOutput(io.Discard)
		} else {
			log.SetOutput(os.Stderr)
		}

		appLogger = log.New(log.Writer(), "app:  ", log.Flags())
	}
}

func main() {
	err := rootCommand.Execute()
	if err != nil {
		os.Exit(-1)
	}
}

// CreateCancellableContext создает новый контекст, чья отмена привязана к SIGINT/SIGKILL
func CreateCancellableContext() context.Context {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	go func() {
		_ = <-signals
		cancel()
	}()

	return ctx
}

// NewApp создает новый объект приложения
func NewApp() (app.App, error) {
	return app.New(app.WithMoexURL(moexURL), app.WithDataSource(postgresConnString))
}
