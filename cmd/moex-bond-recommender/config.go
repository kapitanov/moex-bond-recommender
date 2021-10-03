package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
	"github.com/kapitanov/moex-bond-recommender/pkg/web"
)

// createCancellableContext создает новый контекст, чья отмена привязана к SIGINT/SIGKILL
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
	envVarName := "LISTEN_ADDR"
	defaultValue := os.Getenv(envVarName)
	if defaultValue == "" {
		defaultValue = web.DefaultAddress
	}

	usage := fmt.Sprintf("Web app listen address (defaults to $%s)", envVarName)
	cmd.Flags().StringVarP(value, "address", "a", defaultValue, usage)
}

func attachGoogleAnalyticsFlag(cmd *cobra.Command, value *string) {
	envVarName := "GOOGLE_ANALYTICS_ID"
	defaultValue := os.Getenv(envVarName)
	if defaultValue == "" {
		defaultValue = ""
	}

	usage := fmt.Sprintf("Google Analytics ID (defaults to $%s)", envVarName)
	cmd.Flags().StringVar(value, "ga-id", defaultValue, usage)
}

func parseDuration(s string) (recommender.Duration, error) {
	switch s {
	case "1y":
		return recommender.Duration1Year, nil
	case "2y":
		return recommender.Duration2Year, nil
	case "3y":
		return recommender.Duration3Year, nil
	case "4y":
		return recommender.Duration4Year, nil
	case "5y":
		return recommender.Duration5Year, nil
	default:
		return recommender.Duration1Year, fmt.Errorf("\"%s\" is not a valid duration range, valid values are: 1y, 2y, 3y, 4y, 5y", s)
	}
}
