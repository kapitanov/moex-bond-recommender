package moex

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Coupon описывает купон облигации
type Coupon struct {
	ISIN             string       `json:"isin"`
	Name             string       `json:"name"`
	IssueValue       *float64     `json:"issuevalue"`
	CouponDate       NullableDate `json:"coupondate"`
	RecordDate       NullableDate `json:"recorddate"`
	StartDate        NullableDate `json:"startdate"`
	InitialFaceValue *float64     `json:"initialfacevalue"`
	FaceValue        *float64     `json:"facevalue"`
	FaceUnit         string       `json:"faceunit"`
	Value            *float64     `json:"value"`
	ValuePercent     *float64     `json:"valueprc"`
	ValueRub         *float64     `json:"value_rub"`
}

// CouponListQuery определяет параметры запроса списка купонов
type CouponListQuery struct {
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

func (q CouponListQuery) getValues(values url.Values) {
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
	case IsNotTraded:
		values.Set("is_traded", "false")
	}
}

// CouponListIterator определяет итератор для списка купонов
type CouponListIterator interface {
	// Next загружает следующую страницу данных
	// Если данных больше нет, то возвращается ошибка EOF
	Next() ([]*Coupon, error)
}

// ListCoupons возвращает итератор на список купонов
func (p *provider) ListCoupons(ctx context.Context, query CouponListQuery) CouponListIterator {
	if query.Limit <= 0 {
		query.Limit = 100
	}

	return &couponListIterator{p, query, ctx}
}

type couponListIterator struct {
	provider *provider
	query    CouponListQuery
	ctx      context.Context
}

// Next загружает следующую страницу данных
// Если данных больше нет, то возвращается ошибка EOF
func (it *couponListIterator) Next() ([]*Coupon, error) {
	u := it.getURL()

	resp := make([]couponsResponse, 0)
	err := it.provider.getJSON(it.ctx, u, &resp)
	if err != nil {
		return nil, err
	}

	items := make([]*Coupon, 0)
	for _, respItem := range resp {
		if respItem.Coupons != nil {
			items = append(items, respItem.Coupons...)
		}
	}

	if len(items) == 0 {
		return nil, EOF
	}

	it.query.Start += len(items)
	return items, nil
}

func (it *couponListIterator) getURL() string {
	values := make(url.Values)

	it.query.getValues(values)

	values.Set("iss.only", "coupons")
	values.Set("sort_order", "asc")
	values.Set("iss.json", "extended")
	values.Set("iss.meta", "off")

	u := fmt.Sprintf("/iss/statistics/engines/stock/markets/bonds/bondization.json?%s", values.Encode())
	return u
}

type couponsResponse struct {
	Coupons []*Coupon `json:"coupons"`
}
