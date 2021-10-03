package recommender

import (
	"context"
	"math"
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

	report := mapReport(entity)

	err = s.enrichWithCashFlow(tx, report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

// enrichWithCashFlow дозагружает в отчет данные по выплатам
func (s *service) enrichWithCashFlow(tx *data.TX, report *Report) error {
	payments, err := tx.CashFlow.List(report.Bond.ID)
	if err != nil {
		return err
	}

	report.CashFlow = make([]*CashFlowItem, len(payments))
	for i, payment := range payments {
		report.CashFlow[i] = &CashFlowItem{
			Type:     CashFlowItemType(payment.Type),
			Date:     payment.Date,
			ValueRub: payment.ValueRub,
		}
	}

	return nil
}

// Suggest выполняет расчет предложений по инвестированию
func (s *service) Suggest(ctx context.Context, tx *data.TX, request *SuggestRequest) (*SuggestResult, error) {
	// Формирование позиций
	positions, err := s.generatePositionsForSuggestion(tx, request)
	if err != nil {
		return nil, err
	}

	// Формирование портфеля
	result := &SuggestResult{
		Positions:          positions,
		Amount:             0, // Рассчитывается отдельно
		DurationDays:       0, // Рассчитывается отдельно
		ProfitLoss:         0, // Рассчитывается отдельно
		RelativeProfitLoss: 0, // Рассчитывается отдельно
		InterestRate:       0, // Рассчитывается отдельно
	}

	for _, p := range positions {
		result.Amount += p.OpenValue
		if result.DurationDays < p.DaysTillMaturity {
			result.DurationDays = p.DaysTillMaturity
		}
		result.ProfitLoss += p.ProfitLoss
	}

	result.RelativeProfitLoss = 100.0 * result.ProfitLoss / result.Amount
	result.InterestRate = result.RelativeProfitLoss / (float64(result.DurationDays) / 356.25)

	return result, nil
}

// generatePositionsForSuggestion выполняет генерацию позиций по запросу
func (s *service) generatePositionsForSuggestion(tx *data.TX, request *SuggestRequest) ([]*SuggestedPortfolioPosition, error) {
	var positions []*SuggestedPortfolioPosition

	// Расчет позиций согласно ограничениям
	if request.Parts != nil && len(request.Parts) > 0 {
		// Нормализация весов и сортировка групп по убыванию веса
		sumOfWeights := 0.0
		for _, part := range request.Parts {
			sumOfWeights += part.Weight
		}
		for _, part := range request.Parts {
			part.Weight = part.Weight / sumOfWeights
		}

		sort.Slice(request.Parts, func(i, j int) bool {
			return request.Parts[i].Weight < request.Parts[j].Weight
		})

		// Расчет позиций по каждой из групп
		positions = make([]*SuggestedPortfolioPosition, 0)
		unusedAmount := float64(0)
		for _, part := range request.Parts {
			maxAmount := math.Floor(request.Amount*part.Weight) + unusedAmount
			ps, remainingAmount, err := s.generatePositionsForSuggestionPart(tx, part.Collection, request.MaxDuration, maxAmount)
			if err != nil {
				return nil, err
			}

			for _, p := range ps {
				positions = append(positions, p)
			}
			unusedAmount = remainingAmount
		}

	} else {
		var err error
		positions, _, err = s.generatePositionsForSuggestionPart(tx, nil, request.MaxDuration, request.Amount)
		if err != nil {
			return nil, err
		}
	}

	// Расчет весов позиций
	totalAmount := float64(0)
	for _, p := range positions {
		totalAmount += p.OpenValue
	}
	for _, p := range positions {
		p.Weight = p.OpenValue / totalAmount
	}

	return positions, nil

}

// generatePositionsForSuggestionPart выполняет генерацию позиций по запросу для отдельно взятой коллекции
func (s *service) generatePositionsForSuggestionPart(
	tx *data.TX,
	collection Collection,
	duration Duration,
	maxAmount float64) ([]*SuggestedPortfolioPosition, float64, error) {

	// Выбираем подходящие облигации
	reports, err := s.getBondForSuggestion(tx, collection, duration)
	if err != nil {
		return nil, 0, err
	}

	// До тех пор, пока у нас имеется неиспользованный объем инвестиций,
	// перебираем все подходящие облигации, рассчитывая объем позиции
	positions := make([]*SuggestedPortfolioPosition, 0)
	for _, report := range reports {
		// Рассчитываем объем позиции
		quantity := int(math.Floor(maxAmount / report.OpenValue))
		if quantity <= 0 {
			continue
		}
		quantityF := float64(quantity)

		// Корректируем объем инвестиций
		maxAmount -= report.OpenValue * float64(quantity)

		// Формируем позицию
		r := mapReport(report)
		err = s.enrichWithCashFlow(tx, r)
		if err != nil {
			return nil, 0, err
		}

		r.OpenFee *= quantityF
		r.OpenValue *= quantityF
		r.CouponPayments *= quantityF
		r.AmortizationPayments *= quantityF
		r.MaturityPayment *= quantityF
		r.Taxes *= quantityF
		r.Revenue *= quantityF
		r.ProfitLoss *= quantityF
		for _, c := range r.CashFlow {
			c.ValueRub *= quantityF
		}

		position := &SuggestedPortfolioPosition{
			Report:   *r,
			Quantity: quantity,
			Weight:   0, // Веса рассчитываются позже, т.к. они считаются по всему портфелю
		}
		positions = append(positions, position)

		if maxAmount <= 0 {
			break
		}
	}

	return positions, maxAmount, nil
}

// getBondForSuggestion выполняет выборку облигаций по запросу
func (s *service) getBondForSuggestion(tx *data.TX, collection Collection, duration Duration) ([]*data.Report, error) {
	if collection == nil {
		// Выборка облигаций по критериям:
		// - погашение в пределах срока инвестирования
		// - доходность в рамках трех сигм
		// - не более 10 облигаций
		sql := `
WITH cte AS (
    SELECT b.id,
           r.interest_rate,
           AVG(r.interest_rate) OVER ()    AS mean,
           STDDEV(r.interest_rate) OVER () AS stddev
    FROM reports r
    INNER JOIN bonds b ON b.id = r.bond_id
    WHERE b.high_risk = FALSE
      AND b.maturity_date <= NOW() + ? * '1y'::interval
      AND b.maturity_date >= NOW() + (0.5 * ? * '1y'::interval)
      AND r.interest_rate > 0
    ORDER BY r.interest_rate DESC
)
SELECT id AS bond_id, row_number() OVER () AS index
FROM cte
WHERE (interest_rate <= mean + 3 * stddev)
`
		d := getAge(duration)
		return tx.Reports.List(10, sql, d, d)
	} else {
		// Выборка облигаций по критериям:
		// - погашение в пределах срока инвестирования
		// - погашение в пределах срока инвестирования
		// - доходность в рамках: [max - 1, max]
		// - не более 10 облигаций

		sql := `
WITH cte_bonds AS (
    SELECT DISTINCT b.id
    FROM bonds b
    INNER JOIN collection_bonds cb ON b.id = cb.bond_id
    WHERE cb.collection_id = ?
),
     cte AS (
         SELECT b.id,
                r.interest_rate,
                MAX(r.interest_rate) OVER ()                     AS max_interest_rate,
                (MAX(r.interest_rate) OVER () - r.interest_rate) AS delta_interest_rate
         FROM reports r
         INNER JOIN bonds b ON b.id = r.bond_id
         INNER JOIN cte_bonds ON cte_bonds.id = r.bond_id
         WHERE b.maturity_date <= NOW() + ? * '1y'::interval
           AND b.maturity_date >= NOW() + (0.5 * ? * '1y'::interval)
           AND r.interest_rate > 0
         ORDER BY r.interest_rate DESC
     )
SELECT id AS bond_id, row_number() OVER () AS index
FROM cte
WHERE (max_interest_rate - interest_rate) <= 1
`
		d := getAge(duration)
		return tx.Reports.List(10, sql, collection.ID(), d, d)
	}
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

var emptyCashFlowArray = make([]*CashFlowItem, 0)

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
