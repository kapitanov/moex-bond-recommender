package app

import (
	"context"
	"time"

	"github.com/reugn/go-quartz/quartz"
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
	scheduler          quartz.Scheduler
	isSchedulerRunning bool
}

// IsStaticDataUpToDate возвращает false, если статические данные нуждаются в обновлении
func (app *appImpl) IsStaticDataUpToDate(ctx context.Context) (bool, error) {
	tx, err := app.db.BeginTX()
	if err != nil {
		return false, err
	}
	defer tx.Close()

	lastTime, err := tx.Bonds.GetLastUpdateTime()
	if err != nil {
		return false, err
	}

	if lastTime == nil {
		return false, nil
	}

	isUpToDate := time.Since(*lastTime).Hours() < 24.0
	return isUpToDate, nil
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

// NewUnitOfWork создает новый unit of work
func (app *appImpl) NewUnitOfWork(ctx context.Context) (UnitOfWork, error) {
	tx, err := app.db.BeginTX()
	if err != nil {
		return nil, err
	}

	u := &unitOfWork{
		tx:                 tx,
		ctx:                ctx,
		searchService:      app.searchService,
		recommenderService: app.recommenderService,
	}
	return u, nil
}
