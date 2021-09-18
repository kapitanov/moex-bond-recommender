package moex

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Amortization описывает амортизацию/погашение облигации
type Amortization struct {
	ISIN             string           `json:"isin"`
	Name             string           `json:"name"`
	IssueValue       float64          `json:"issuevalue"`
	AmortDate        NullableDate     `json:"amortdate"`
	InitialFaceValue float64          `json:"initialfacevalue"`
	FaceValue        float64          `json:"facevalue"`
	FaceUnit         string           `json:"faceunit"`
	Value            float64          `json:"value"`
	ValuePercent     float64          `json:"valueprc"`
	ValueRub         float64          `json:"value_rub"`
	Type             AmortizationType `json:"data_source"`
}

// AmortizationType описывает тип амортизации
type AmortizationType string

const (
	// AmortizationTypeA обозначает тип амортизации - собственно амортизацию
	AmortizationTypeA AmortizationType = "amortization"

	// AmortizationTypeM обозначает тип амортизации - погашение
	AmortizationTypeM AmortizationType = "maturity"
)

// AmortizationListQuery определяет параметры запроса списка амортизаций
type AmortizationListQuery struct {
	// Дата, больше либо равно
	From *time.Time

	// Дата, меньше либо равно
	Till *time.Time

	// Статус торгуемой ценной бумаги
	TradingStatus TradingStatus

	// Сколько записей выводить
	Limit int

	// Сколько записей пропускать
	Start int
}

func (q AmortizationListQuery) getValues(values url.Values) {
	if q.From != nil {
		values.Set("from", q.From.Format("2006-01-02"))
	}

	if q.Till != nil {
		values.Set("till", q.Till.Format("2006-01-02"))
	}

	if q.Limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", q.Limit))
	}

	if q.Start > 0 {
		values.Set("start", fmt.Sprintf("%d", q.Start))
	}

	switch q.TradingStatus {
	case IsTraded:
		values.Set("is_traded", "true")
		break
	case IsNotTraded:
		values.Set("is_traded", "false")
		break
	}
}

// AmortizationListIterator определяет итератор для списка амортизаций
type AmortizationListIterator interface {
	// Next загружает следующую страницу данных
	// Если данных больше нет, то возвращается ошибка EOF
	Next() ([]*Amortization, error)
}

// ListAmortizations возвращает итератор на список амортизаций
func (p *provider) ListAmortizations(ctx context.Context, query AmortizationListQuery) AmortizationListIterator {
	if query.Limit <= 0 {
		query.Limit = 100
	}

	return &amortizationListIterator{p, query, ctx}
}

type amortizationListIterator struct {
	provider *provider
	query    AmortizationListQuery
	ctx      context.Context
}

// Next загружает следующую страницу данных
// Если данных больше нет, то возвращается ошибка EOF
func (it *amortizationListIterator) Next() ([]*Amortization, error) {
	u := it.getURL()

	resp := make([]amortizationsResponse, 0)
	err := it.provider.getJSON(it.ctx, u, &resp)
	if err != nil {
		return nil, err
	}

	items := make([]*Amortization, 0)
	for _, respItem := range resp {
		if respItem.Amortizations != nil {
			for _, item := range respItem.Amortizations {
				items = append(items, item)
			}
		}
	}

	if len(items) == 0 {
		return nil, EOF
	}

	it.query.Start += len(items)
	return items, nil
}

func (it *amortizationListIterator) getURL() string {
	values := make(url.Values)

	it.query.getValues(values)

	values.Set("iss.only", "amortizations")
	values.Set("sort_order", "asc")
	values.Set("iss.json", "extended")
	values.Set("iss.meta", "off")

	u := fmt.Sprintf("/iss/statistics/engines/stock/markets/bonds/bondization.json?%s", values.Encode())
	return u
}

type amortizationsResponse struct {
	Amortizations []*Amortization `json:"amortizations"`
}
