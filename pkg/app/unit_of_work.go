package app

import (
	"context"
	"strconv"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
	"github.com/kapitanov/moex-bond-recommender/pkg/search"
)

// UnitOfWork представляет собой реализацию паттерна "unit of work"
type UnitOfWork interface {
	// Search выполняет поиск облигации по тексту
	Search(req search.Request) (*search.Result, error)

	// ListCollections возвращает список коллекций рекомендаций
	ListCollections() []recommender.Collection

	// GetCollection возвращает коллекцию рекомендаций по ее ID
	GetCollection(id string) (recommender.Collection, error)

	// ListCollectionBonds возвращает облигации из коллекции рекомендаций по ее ID
	ListCollectionBonds(id string, duration recommender.Duration) ([]*recommender.Report, error)

	// GetReport возвращает отчет по отдельной облигации
	GetReport(idOrISIN string) (*recommender.Report, error)

	// Close закрывает unit of work
	Close()
}

type unitOfWork struct {
	tx                 *data.TX
	ctx                context.Context
	searchService      search.Service
	recommenderService recommender.Service
}

// Search выполняет поиск облигации по тексту
func (u *unitOfWork) Search(req search.Request) (*search.Result, error) {
	return u.searchService.Do(u.tx, req)
}

// ListCollections возвращает список коллекций рекомендаций
func (u *unitOfWork) ListCollections() []recommender.Collection {
	return u.recommenderService.ListCollections()
}

// GetCollection возвращает коллекцию рекомендаций по ее ID
func (u *unitOfWork) GetCollection(id string) (recommender.Collection, error) {
	return u.recommenderService.GetCollection(id)
}

// ListCollectionBonds возвращает облигации из коллекции рекомендаций по ее ID
func (u *unitOfWork) ListCollectionBonds(id string, duration recommender.Duration) ([]*recommender.Report, error) {
	collection, err := u.recommenderService.GetCollection(id)
	if err != nil {
		return nil, err
	}

	limit := 25
	items, err := collection.ListBonds(u.ctx, u.tx, limit, duration)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetReport возвращает отчет по отдельной облигации
func (u *unitOfWork) GetReport(idOrISIN string) (*recommender.Report, error) {
	id, err := strconv.Atoi(idOrISIN)
	if err != nil {
		bond, err := u.tx.Bonds.GetByISIN(idOrISIN)
		if err != nil {
			if err == data.ErrNotFound {
				bond, err = u.tx.Bonds.GetBySecurityID(idOrISIN)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		id = bond.ID
	}

	report, err := u.recommenderService.GetReport(u.ctx, u.tx, id)
	if err != nil {
		return nil, err
	}

	return report, nil
}

// Close закрывает unit of work
func (u *unitOfWork) Close() {
	u.tx.Close()
}
