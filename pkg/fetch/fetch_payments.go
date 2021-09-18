package fetch

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

type paymentFetchWorker struct {
	provider moex.Provider
	tx       *data.TX
	log      *log.Logger
	stats    *PaymentFetchStats
	bondIDs  map[string]int
}

// FetchCoupons выполняет выгрузку купонов
func (w *paymentFetchWorker) FetchCoupons(ctx context.Context) error {
	since, err := w.GetStartDate(data.CouponPayment)
	if err != nil {
		return err
	}

	it := w.provider.ListCoupons(ctx, moex.CouponListQuery{From: since})
	count := 0
	for {
		w.log.Printf("fetch coupons: %d item(s) processed", count)

		coupons, err := it.Next()
		if err != nil {
			if err == moex.EOF {
				break
			}
			return err
		}

		for _, coupon := range coupons {
			err = ctx.Err()
			if err != nil {
				return err
			}

			count++

			bondID, err := w.GetBondID(coupon.ISIN)
			if err != nil {
				if err == data.ErrNotFound {
					continue
				}

				return err
			}

			var date time.Time
			if coupon.CouponDate.HasValue() {
				date = *coupon.CouponDate.Time()
			} else if coupon.StartDate.HasValue() {
				date = *coupon.StartDate.Time()
			} else if coupon.RecordDate.HasValue() {
				date = *coupon.RecordDate.Time()
			} else {
				return fmt.Errorf("a coupon on %s has no date", coupon.ISIN)
			}

			value := float64(0)
			if coupon.Value != nil {
				value = *coupon.Value
			}

			valuePercent := float64(0)
			if coupon.ValuePercent != nil {
				valuePercent = *coupon.ValuePercent
			}

			valueRub := float64(0)
			if coupon.ValueRub != nil {
				valueRub = *coupon.ValueRub
			}

			args := data.CreateCouponPaymentArgs{
				CreatePaymentArgs: data.CreatePaymentArgs{
					BondID:       bondID,
					Date:         date,
					Value:        value,
					ValuePercent: valuePercent,
					ValueRub:     valueRub,
				},
				RecordDate: coupon.RecordDate.Time(),
				StartDate:  coupon.StartDate.Time(),
			}

			_, err = w.tx.Payments.CreateCoupon(args)
			if err != nil {
				if err == data.ErrAlreadyExists {
					continue
				}
				return err
			}

			w.stats.NewCoupons++
		}
	}

	return nil
}

// FetchAmortizations выполняет выгрузку амортизаций и погашений
func (w *paymentFetchWorker) FetchAmortizations(ctx context.Context) error {
	sinceAmort, err := w.GetStartDate(data.AmortizationPayment)
	if err != nil {
		return err
	}

	sinceMat, err := w.GetStartDate(data.MaturityPayment)
	if err != nil {
		return err
	}

	since := sinceAmort
	if sinceMat != nil && sinceAmort != nil && sinceMat.After(*sinceAmort) {
		since = sinceMat
	}

	it := w.provider.ListAmortizations(ctx, moex.AmortizationListQuery{From: since})
	count := 0
	for {
		w.log.Printf("fetch amortizations: %d item(s) processed", count)

		amortizations, err := it.Next()
		if err != nil {
			if err == moex.EOF {
				break
			}
			return err
		}

		for _, amortization := range amortizations {
			err = ctx.Err()
			if err != nil {
				return err
			}

			count++

			bondID, err := w.GetBondID(amortization.ISIN)
			if err != nil {
				if err == data.ErrNotFound {
					continue
				}

				return err
			}

			var date time.Time
			if amortization.AmortDate.HasValue() {
				date = *amortization.AmortDate.Time()
			} else {
				return fmt.Errorf("an amortization on %s has no date", amortization.ISIN)
			}

			if amortization.Type == moex.AmortizationTypeM {
				args := data.CreateMaturityPaymentArgs{
					CreatePaymentArgs: data.CreatePaymentArgs{
						BondID:       bondID,
						Date:         date,
						Value:        amortization.Value,
						ValuePercent: amortization.ValuePercent,
						ValueRub:     amortization.ValueRub,
					},
				}

				_, err = w.tx.Payments.CreateMaturity(args)
				if err != nil {
					if err == data.ErrAlreadyExists {
						continue
					}
					return err
				}

				w.stats.NewMaturities++
			} else {
				args := data.CreateAmortizationPaymentArgs{
					CreatePaymentArgs: data.CreatePaymentArgs{
						BondID:       bondID,
						Date:         date,
						Value:        amortization.Value,
						ValuePercent: amortization.ValuePercent,
						ValueRub:     amortization.ValueRub,
					},
				}

				_, err = w.tx.Payments.CreateAmortization(args)
				if err != nil {
					if err == data.ErrAlreadyExists {
						continue
					}
					return err
				}

				w.stats.NewAmortizations++
			}
		}
	}

	return nil
}

// GetStartDate возвращает дату начала синхронизации для выплат указанного типа
func (w *paymentFetchWorker) GetStartDate(t data.PaymentType) (*time.Time, error) {
	lastPayment, err := w.tx.Payments.Last(t)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	ny, nm, nd := time.Now().Date()
	ly, lm, ld := lastPayment.Date.Date()

	if ny == ly && nm == lm && nd == ld {
		return nil, nil
	}

	return &lastPayment.Date, nil
}

// GetBondID возвращает ID облигации по ее ISIN.
// Если облигации не найдено, то возвращается data.ErrNotFound
func (w *paymentFetchWorker) GetBondID(isin string) (int, error) {
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
