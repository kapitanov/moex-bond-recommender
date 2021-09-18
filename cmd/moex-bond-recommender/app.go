package main

import (
	"context"
	"log"
	"time"

	"github.com/subchen/go-trylock"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/fetch"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

type App struct {
	Moex            moex.Provider
	DB              data.DB
	Fetch           fetch.Service
	fetchInProgress trylock.TryLocker
}

func NewApp() (*App, error) {
	moexLogger := log.New(log.Writer(), "moex: ", log.Flags())
	provider, err := moex.NewProvider(moex.WithURL(moexURL), moex.WithLogger(moexLogger))
	if err != nil {
		return nil, err
	}

	dataLogger := log.New(log.Writer(), "data: ", log.Flags())
	db, err := data.New(data.WithDataSource(postgresConnString), data.WithLogger(dataLogger))
	if err != nil {
		return nil, err
	}

	serviceLogger := log.New(log.Writer(), "fetch: ", log.Flags())
	service, err := fetch.New(fetch.WithDB(db), fetch.WithProvider(provider), fetch.WithLogger(serviceLogger))
	if err != nil {
		return nil, err
	}

	app := &App{
		Moex:            provider,
		DB:              db,
		Fetch:           service,
		fetchInProgress: trylock.New(),
	}
	return app, nil
}

func (app *App) FetchStaticData(ctx context.Context) error {
	app.fetchInProgress.Lock()
	defer app.fetchInProgress.Unlock()

	tx, err := app.DB.BeginTX()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Close()

	_, err = app.Fetch.FetchBonds(ctx, tx)
	if err != nil {
		return err
	}

	_, err = app.Fetch.FetchPayments(ctx, tx)
	if err != nil {
		return err
	}

	_, err = app.Fetch.FetchOffers(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (app *App) FetchMarketData(ctx context.Context) error {
	if !app.fetchInProgress.TryLock(time.Second) {
		return nil
	}

	defer app.fetchInProgress.Unlock()

	tx, err := app.DB.BeginTX()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Close()

	_, err = app.Fetch.FetchMarketData(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
