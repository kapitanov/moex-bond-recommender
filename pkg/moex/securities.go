package moex

import (
	"context"
	"fmt"
	"net/url"
)

// Security описывает ценную бумагу
type Security struct {
	ID                 int           `json:"id"`
	SecurityID         string        `json:"secid"`
	ShortName          string        `json:"shortname"`
	RegNumber          string        `json:"regnumber"`
	Name               string        `json:"name"`
	ISIN               string        `json:"isin"`
	IsTraded           TradingStatus `json:"is_traded"`
	IssuerId           int           `json:"emitent_id"`
	IssuerName         string        `json:"emitent_title"`
	IssuerINN          *string       `json:"emitent_inn"`
	IssuerOKPO         *string       `json:"emitent_okpo"`
	Type               SecurityType  `json:"type"`
	PrimaryBoardID     string        `json:"primary_boardid"`
	MarketPriceBoardID string        `json:"marketprice_boardid"`
}

// SecurityType содержит тип ценной бумаги
type SecurityType string

const (
	// SubfederalBond - субфедеральные облигации
	SubfederalBond SecurityType = "subfederal_bond"

	// OFZBond - ОФЗ
	OFZBond SecurityType = "ofz_bond"

	// ExchangeBond - биржевые облигации
	ExchangeBond SecurityType = "exchange_bond"

	// CBBond - облигации ЦБ
	CBBond SecurityType = "cb_bond"

	// MunicipalBond - мунициальные облигации
	MunicipalBond SecurityType = "municipal_bond"

	// CorporateBond - корпоративные облигации
	CorporateBond SecurityType = "corporate_bond"

	// IFIBond - облигации ИФИ
	IFIBond SecurityType = "ifi_bond"

	// EuroBond - еврооблигации
	EuroBond SecurityType = "euro_bond"
)

// Engine содержит код движка
type Engine string

const (
	// AnyEngine обозначает любую биржу
	AnyEngine Engine = ""

	// StockEngine обозначает фондовую биржу
	StockEngine Engine = "stock"
)

// Market содержит код рынка
type Market string

const (
	// AnyMarket обозначает любой рынок
	AnyMarket Market = ""

	// BondMarket обозначает рынок облигаций
	BondMarket Market = "bonds"
)

// TradingStatus содержит статус торгуемой ценной бумаги
type TradingStatus int

const (
	// IsTraded обозначает торгуемые ценные бумаги
	IsTraded TradingStatus = 1

	// IsNotTraded обозначает неторгуемые ценные бумаги
	IsNotTraded TradingStatus = 0
)

// SecurityListQuery определяет параметры запроса списка ценных бумаг
type SecurityListQuery struct {
	// Биржа
	Engine Engine

	// Рынок
	Market Market

	// Статус торгуемой ценной бумаги
	TradingStatus *TradingStatus

	// Сколько записей выводить
	Limit int

	// Сколько записей пропускать
	Start int
}

func (q SecurityListQuery) getValues(values url.Values) {
	if q.Engine != "" {
		values.Set("engine", string(q.Engine))
	}

	if q.Market != "" {
		values.Set("market", string(q.Market))
	}

	if q.Limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", q.Limit))
	}

	if q.Start > 0 {
		values.Set("start", fmt.Sprintf("%d", q.Start))
	}

	if q.TradingStatus != nil {
		switch *q.TradingStatus {
		case IsTraded:
			values.Set("is_trading", "true")
		case IsNotTraded:
			values.Set("is_trading", "false")
		}
	}
}

// SecurityListIterator определяет итератор для списка ценных бумаг
type SecurityListIterator interface {
	// Next загружает следующую страницу данных
	// Если данных больше нет, то возвращается ошибка EOF
	Next() ([]*Security, error)
}

// ListSecurities возвращает итератор на список ценных бумаг
func (p *provider) ListSecurities(ctx context.Context, query SecurityListQuery) SecurityListIterator {
	if query.Limit <= 0 {
		query.Limit = 100
	}

	return &securityListIterator{p, query, ctx}
}

type securityListIterator struct {
	provider *provider
	query    SecurityListQuery
	ctx      context.Context
}

// Next загружает следующую страницу данных
// Если данных больше нет, то возвращается ошибка EOF
func (it *securityListIterator) Next() ([]*Security, error) {
	u := it.getURL()

	resp := make([]securitiesResponse, 0)
	err := it.provider.getJSON(it.ctx, u, &resp)
	if err != nil {
		return nil, err
	}

	items := make([]*Security, 0)
	for _, respItem := range resp {
		if respItem.Securities != nil {
			items = append(items, respItem.Securities...)
		}
	}

	if len(items) == 0 {
		return nil, EOF
	}

	it.query.Start += len(items)
	return items, nil
}

func (it *securityListIterator) getURL() string {
	values := make(url.Values)

	it.query.getValues(values)
	values.Set("iss.json", "extended")
	values.Set("iss.meta", "off")

	u := fmt.Sprintf("/iss/securities.json?%s", values.Encode())
	return u
}

type securitiesResponse struct {
	Securities []*Security `json:"securities"`
}
