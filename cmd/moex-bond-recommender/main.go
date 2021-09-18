package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

var rootCommand = &cobra.Command{
	Use:              "moex-bond-recommender",
	TraverseChildren: true,
}

var (
	postgresConnString string
	moexURL            string
)

var appLogger *log.Logger

func init() {
	rootCommand.PersistentFlags().StringVarP(&postgresConnString, "psql", "p", data.DefaultDataSource, "Postgres connnection string")
	rootCommand.PersistentFlags().StringVar(&moexURL, "moex", moex.DefaultURL, "ISS root URL")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	appLogger = log.New(log.Writer(), "app:  ", log.Flags())

	err := rootCommand.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
}

func createCancellableContext() context.Context {
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
