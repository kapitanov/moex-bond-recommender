package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
	"github.com/kapitanov/moex-bond-recommender/pkg/web"
)

var rootCommand = &cobra.Command{
	Use:              "moex-bond-recommender",
	TraverseChildren: true,
	SilenceUsage:     true,
}

var (
	quietMode bool
)

var appLogger *log.Logger

func init() {
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

func attachPostgresUrlFlag(cmd *cobra.Command, value *string) {
	envVarName := "POSTGRES_URL"
	defaultValue := os.Getenv(envVarName)
	if defaultValue == "" {
		defaultValue = data.DefaultDataSource
	}

	usage := fmt.Sprintf("Postgres connection string (defaults to $%s)", envVarName)
	cmd.Flags().StringVarP(value, "psql", "p", defaultValue, usage)
}

func attachMoexUrlFlag(cmd *cobra.Command, value *string) {
	envVarName := "ISS_URL"
	defaultValue := os.Getenv(envVarName)
	if defaultValue == "" {
		defaultValue = moex.DefaultURL
	}

	usage := fmt.Sprintf("ISS root URL (defaults to $%s)", envVarName)
	cmd.Flags().StringVar(value, "moex", defaultValue, usage)
}

func attachListenAddressFlag(cmd *cobra.Command, value *string) {
	envVarName := "LISTEN_ADDRR"
	defaultValue := os.Getenv(envVarName)
	if defaultValue == "" {
		defaultValue = web.DefaultAddress
	}

	usage := fmt.Sprintf("Web app listen address (defaults to $%s)", envVarName)
	cmd.Flags().StringVarP(value, "address", "a", defaultValue, usage)
}
