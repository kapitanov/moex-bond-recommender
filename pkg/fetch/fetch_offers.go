package fetch

import (
	"context"
	"log"
	"time"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

type offerFetchWorker struct {
	provider moex.Provider
	tx       *data.TX
	log      *log.Logger
	stats    *OfferFetchStats
	bondIDs  map[string]int
}

// FetchOffers выполняет выгрузку оферт
func (w *offerFetchWorker) FetchOffers(ctx context.Context) error {
	it := w.provider.ListOffers(ctx, moex.OfferListQuery{})
	count := 0
	for {
		w.log.Printf("fetch offers: %d item(s) processed", count)

		offers, err := it.Next()
		if err != nil {
			if err == moex.EOF {
				break
			}
			return err
		}

		for _, offer := range offers {
			err = ctx.Err()
			if err != nil {
				return err
			}

			count++

			bondID, err := w.GetBondID(offer.ISIN)
			if err != nil {
				if err == data.ErrNotFound {
					continue
				}

				return err
			}

			var date *time.Time
			if offer.OfferDate.HasValue() {
				date = offer.OfferDate.Time()
			} else if offer.StartOfferDate.HasValue() {
				date = offer.StartOfferDate.Time()
			} else if offer.EndOfferDate.HasValue() {
				date = offer.EndOfferDate.Time()
			}

			args := data.CreateOfferArgs{
				BondID:     bondID,
				IssueValue: offer.Value,
				Date:       timeToNullTime(date),
				StartDate:  nullableDateToNullTime(offer.StartOfferDate),
				EndDate:    nullableDateToNullTime(offer.EndOfferDate),
				FaceValue:  offer.FaceValue,
				FaceUnit:   normalizeCurrency(offer.FaceUnit),
				Price:      offer.Price,
				Value:      offer.Value,
				Agent:      offer.Agent,
				Type:       w.MapOfferType(offer.Type),
			}

			_, err = w.tx.Offers.Create(args)
			if err != nil {
				if err == data.ErrAlreadyExists {
					continue
				}
				return err
			}

			w.stats.NewOffers++
		}
	}

	return nil
}

// GetBondID возвращает ID облигации по ее ISIN.
// Если облигации не найдено, то возвращается data.ErrNotFound
func (w *offerFetchWorker) GetBondID(isin string) (int, error) {
	id, exists := w.bondIDs[isin]
	if exists {
		if id == 0 {
			return 0, data.ErrNotFound
		}
		return id, nil
	}

	bond, err := w.tx.Bonds.GetByISIN(isin)
	if err != nil {
		if err == data.ErrNotFound {
			w.bondIDs[isin] = 0
		}
		return 0, err
	}

	w.bondIDs[isin] = bond.ID
	return bond.ID, nil
}

// MapOfferType выполняет преобразование из moex.OfferType в data.OfferType
func (w *offerFetchWorker) MapOfferType(t *moex.OfferType) *data.OfferType {
	if t == nil {
		return nil
	}

	var result = data.OfferType(*t)
	switch *t {
	case moex.GenericOffer:
		result = data.GenericOffer
	case moex.CompletedGenericOffer:
		result = data.CompletedGenericOffer
	case moex.CanceledGenericOffer:
		result = data.CanceledGenericOffer
	case moex.DefaultGenericOffer:
		result = data.DefaultGenericOffer
	case moex.TechDefaultGenericOffer:
		result = data.TechDefaultGenericOffer
	case moex.MaturityOffer:
		result = data.MaturityOffer
	case moex.CanceledMaturityOffer:
		result = data.CanceledMaturityOffer
	}

	return &result
}
