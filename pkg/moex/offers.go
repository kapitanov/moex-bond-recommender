package moex

import (
	"fmt"
	"net/url"
	"time"
)

// Offer описывает оферту по облигации
type Offer struct {
	ISIN           string       `json:"isin"`
	Name           string       `json:"name"`
	IssueValue     *float64     `json:"issuevalue"`
	OfferDate      NullableDate `json:"offerdate"`
	StartOfferDate NullableDate `json:"offerdatestart"`
	EndOfferDate   NullableDate `json:"offerdateend"`
	FaceValue      *float64     `json:"facevalue"`
	FaceUnit       string       `json:"faceunit"`
	Price          *float64     `json:"price"`
	Value          *float64     `json:"value"`
	Agent          string       `json:"agent"`
	Type           OfferType    `json:"offertype"`
}

// OfferType содержит тип оферты
type OfferType string

const (
	// GenericOffer - оферта
	GenericOffer OfferType = "Оферта"

	// CompletedGenericOffer - состоявшаяся оферта
	CompletedGenericOffer OfferType = "Оферта (состоялось)"

	// CanceledGenericOffer - отмененнная оферта
	CanceledGenericOffer OfferType = "Оферта (отменено)"

	// DefaultGenericOffer - дефолт оферты
	DefaultGenericOffer OfferType = "Оферта (дефолт)"

	// TechDefaultGenericOffer - технический дефолт оферты
	TechDefaultGenericOffer OfferType = "Оферта (технический дефолт)"

	// MaturityOffer - оферта-погашение
	MaturityOffer OfferType = "Оферта/Погашение"

	// CanceledMaturityOffer - отмененнная оферта-погашение
	CanceledMaturityOffer OfferType = "Оферта/Погашение(отменено)"
)

// OfferListQuery определяет параметры запроса списка оферт
type OfferListQuery struct {
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

func (q OfferListQuery) getValues(values url.Values) {
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

// OfferListIterator определяет итератор для списка оферт
type OfferListIterator interface {
	// Next загружает следующую страницу данных
	// Если данных больше нет, то возвращается ошибка EOF
	Next() ([]*Offer, error)
}

// ListOffers возвращает итератор на список оферт
func (p *provider) ListOffers(query OfferListQuery) OfferListIterator {
	if query.Limit <= 0 {
		query.Limit = 100
	}

	return &offersListIterator{p, query}
}

type offersListIterator struct {
	provider *provider
	query    OfferListQuery
}

// Next загружает следующую страницу данных
// Если данных больше нет, то возвращается ошибка EOF
func (it *offersListIterator) Next() ([]*Offer, error) {
	u := it.getURL()

	resp := make([]offersResponse, 0)
	err := it.provider.getJSON(u, &resp)
	if err != nil {
		return nil, err
	}

	items := make([]*Offer, 0)
	for _, respItem := range resp {
		if respItem.Offers != nil {
			for _, item := range respItem.Offers {
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

func (it *offersListIterator) getURL() string {
	values := make(url.Values)

	it.query.getValues(values)

	values.Set("iss.only", "offers")
	values.Set("sort_order", "asc")
	values.Set("iss.json", "extended")
	values.Set("iss.meta", "off")

	u := fmt.Sprintf("/iss/statistics/engines/stock/markets/bonds/bondization.json?%s", values.Encode())
	return u
}

type offersResponse struct {
	Offers []*Offer `json:"offers"`
}
