package recommender

import (
	"context"
	"sort"
	"strings"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
)

type service struct {
}

// ListCollections возвращает список коллекций рекомендаций
func (s *service) ListCollections() []Collection {
	array := make([]Collection, len(collections))
	i := 0
	for _, coll := range collections {
		array[i] = coll
		i++
	}

	sort.Slice(array, func(i, j int) bool {
		return strings.Compare(array[i].ID(), array[j].ID()) < 0
	})

	return array
}

// GetCollection возвращает коллекцию рекомендаций по ее ID
// Если коллекция не найдена, то возвращается ошибка ErrNotFound
func (s *service) GetCollection(id string) (Collection, error) {
	coll, exists := collections[id]
	if !exists {
		return nil, ErrNotFound
	}

	return coll, nil
}

// GetReport возвращает отчет по отдельной облигации
// Если отчет не найден, то возвращается ошибка ErrNotFound
func (s *service) GetReport(ctx context.Context, tx *data.TX, id int) (*Report, error) {
	entity, err := tx.Reports.Get(id)
	if err != nil {
		if err == data.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	payments, err := tx.CashFlow.List(id)
	if err != nil {
		return nil, err
	}

	report := mapReport(entity)

	report.CashFlow = make([]CashFlowItem, len(payments))
	for i, payment := range payments {
		report.CashFlow[i] = CashFlowItem{
			Type:     CashFlowItemType(payment.Type),
			Date:     payment.Date,
			ValueRub: payment.ValueRub,
		}
	}

	return report, nil
}

// Rebuild выполняет обновление данных рекомендаций
func (s *service) Rebuild(ctx context.Context, tx *data.TX) error {
	// Обновляем данные текущих выплат по облигациям
	err := tx.CashFlow.Rebuild()
	if err != nil {
		return err
	}

	// Обновляем данные отчетов по облигациям
	err = tx.Reports.Rebuild()
	if err != nil {
		return err
	}

	// Обновляем данные коллекций
	for _, coll := range collections {
		err := coll.Rebuild(ctx, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

var emptyCashFlowArray = make([]CashFlowItem, 0)

func mapReport(entity *data.Report) *Report {
	report := Report{
		Bond:                 &entity.Bond,
		Issuer:               &entity.Issuer,
		MarketData:           &entity.MarketData,
		DaysTillMaturity:     entity.DaysTillMaturity,
		Currency:             entity.Currency,
		OpenPrice:            entity.OpenPrice,
		OpenAccruedInterest:  entity.OpenAccruedInterest,
		OpenFaceValue:        entity.OpenFaceValue,
		OpenFee:              entity.OpenFee,
		OpenValue:            entity.OpenValue,
		CouponPayments:       entity.CouponPayments,
		AmortizationPayments: entity.AmortizationPayments,
		MaturityPayment:      entity.MaturityPayment,
		Taxes:                entity.Taxes,
		Revenue:              entity.Revenue,
		ProfitLoss:           entity.ProfitLoss,
		RelativeProfitLoss:   entity.RelativeProfitLoss,
		InterestRate:         entity.InterestRate,
		CashFlow:             emptyCashFlowArray,
	}
	return &report
}
