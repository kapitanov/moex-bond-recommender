package main

import (
	"context"
	"log"
	"time"

	"github.com/subchen/go-trylock"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/fetch"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
	"github.com/kapitanov/moex-bond-recommender/pkg/search"
)

type App struct {
	Moex            moex.Provider
	DB              data.DB
	Fetch           fetch.Service
	Search          search.Service
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

	fetchLogger := log.New(log.Writer(), "fetch: ", log.Flags())
	fetchService, err := fetch.New(fetch.WithProvider(provider), fetch.WithLogger(fetchLogger))
	if err != nil {
		return nil, err
	}

	searchLogger := log.New(log.Writer(), "search: ", log.Flags())
	searchService, err := search.New(search.WithLogger(searchLogger))
	if err != nil {
		return nil, err
	}

	app := &App{
		Moex:            provider,
		DB:              db,
		Fetch:           fetchService,
		Search:          searchService,
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

func (app *App) ExecSearch(req search.Request) (*search.Result, error) {
	tx, err := app.DB.BeginTX()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Close()

	return app.Search.Do(tx, req)
}
