package app

import (
	"context"
	"strconv"
	"time"

	"github.com/subchen/go-trylock"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/fetch"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
	"github.com/kapitanov/moex-bond-recommender/pkg/search"
)

type appImpl struct {
	moexProvider       moex.Provider
	db                 data.DB
	fetchService       fetch.Service
	searchService      search.Service
	recommenderService recommender.Service
	fetchInProgress    trylock.TryLocker
}

// FetchStaticData выполняет выгрузку статических данных
func (app *appImpl) FetchStaticData(ctx context.Context) error {
	app.fetchInProgress.Lock()
	defer app.fetchInProgress.Unlock()

	tx, err := app.db.BeginTX()
	if err != nil {
		return err
	}
	defer tx.Close()

	_, err = app.fetchService.FetchBonds(ctx, tx)
	if err != nil {
		return err
	}

	_, err = app.fetchService.FetchPayments(ctx, tx)
	if err != nil {
		return err
	}

	_, err = app.fetchService.FetchOffers(ctx, tx)
	if err != nil {
		return err
	}

	err = app.searchService.Rebuild(ctx, tx)
	if err != nil {
		return err
	}

	err = app.recommenderService.Rebuild(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// FetchMarketData выполняет выгрузку рыночных данных
func (app *appImpl) FetchMarketData(ctx context.Context) error {
	if !app.fetchInProgress.TryLock(time.Second) {
		return nil
	}

	defer app.fetchInProgress.Unlock()

	tx, err := app.db.BeginTX()
	if err != nil {
		return err
	}
	defer tx.Close()

	_, err = app.fetchService.FetchMarketData(ctx, tx)
	if err != nil {
		return err
	}

	err = app.recommenderService.Rebuild(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Search выполняет поиск облигации по тексту
func (app *appImpl) Search(req search.Request) (*search.Result, error) {
	tx, err := app.db.BeginTX()
	if err != nil {
		return nil, err
	}
	defer tx.Close()

	return app.searchService.Do(tx, req)
}

// ListCollections возвращает список коллекций рекомендаций
func (app *appImpl) ListCollections() []recommender.Collection {
	return app.recommenderService.ListCollections()
}

// GetCollection возвращает коллекцию рекомендаций по ее ID
func (app *appImpl) GetCollection(ctx context.Context, id string, duration recommender.Duration) (recommender.Collection, []*recommender.Report, error) {
	tx, err := app.db.BeginTX()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Close()

	collection, err := app.recommenderService.GetCollection(id)
	if err != nil {
		return nil, nil, err
	}

	items, err := collection.ListBonds(ctx, tx, duration)
	if err != nil {
		return nil, nil, err
	}

	return collection, items, nil
}

// GetReport возвращает отчет по отдельной облигации
func (app *appImpl) GetReport(ctx context.Context, idOrISIN string) (*recommender.Report, error) {
	tx, err := app.db.BeginTX()
	if err != nil {
		return nil, err
	}
	defer tx.Close()

	id, err := strconv.Atoi(idOrISIN)
	if err != nil {
		bond, err := tx.Bonds.GetByISIN(idOrISIN)
		if err != nil {
			if err == data.ErrNotFound {
				bond, err = tx.Bonds.GetBySecurityID(idOrISIN)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		id = bond.ID
	}

	report, err := app.recommenderService.GetReport(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return report, nil
}
