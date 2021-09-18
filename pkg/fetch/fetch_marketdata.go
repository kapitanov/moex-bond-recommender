package fetch

import (
	"context"
	"log"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

type marketDataFetchWorker struct {
	provider moex.Provider
	tx       *data.TX
	log      *log.Logger
	stats    *MarketDataFetchStats
	bondIDs  map[string]int
}

// FetchMarketData выполняет выгрузку рыночных данных из биржи в БД
func (w *marketDataFetchWorker) FetchMarketData(ctx context.Context) error {
	items, err := w.provider.GetMarketData(ctx)
	if err != nil {
		return err
	}

	for _, item := range items {
		err = ctx.Err()
		if err != nil {
			return err
		}

		bondID, err := w.GetBondID(item.SecurityID)
		if err != nil {
			if err == data.ErrNotFound {
				continue
			}

			return err
		}

		if item.Time == nil {
			continue
		}

		var currency *string = nil
		if item.Currency != nil {
			c := normalizeCurrency(*item.Currency)
			currency = &c
		}

		args := data.PutMarketDataArgs{
			Time:            item.Time.Time(),
			FaceValue:       item.FaceValue,
			Currency:        currency,
			Last:            item.Last,
			LastChange:      item.LastChange,
			ClosePrice:      item.ClosePrice,
			LegalClosePrice: item.LegalClosePrice,
			AccruedInterest: item.AccruedInterest,
		}

		_, err = w.tx.MarketData.Put(bondID, args)
		if err != nil {
			return err
		}

		w.stats.NewMarketData++
	}
	return nil
}

// GetBondID возвращает ID облигации по ее SecurityID.
// Если облигации не найдено, то возвращается data.ErrNotFound
func (w *marketDataFetchWorker) GetBondID(securityID string) (int, error) {
	id, exists := w.bondIDs[securityID]
	if exists {
		if id == 0 {
			return 0, data.ErrNotFound
		}
		return id, nil
	}

	bond, err := w.tx.Bonds.GetBySecurityID(securityID)
	if err != nil {
		if err == data.ErrNotFound {
			w.bondIDs[securityID] = 0
		}
		return 0, err
	}

	w.bondIDs[securityID] = bond.ID
	return bond.ID, nil
}
