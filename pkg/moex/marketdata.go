package moex

import (
	"fmt"
	"net/url"
)

// MarketData описывает итоги торгов по облигации
type MarketData struct {
	SecurityID      string
	BoardID         string
	AccruedInterest *float64
	FaceValue       *float64
	Currency        *string
	Last            *float64
	LastChange      *float64
	ClosePrice      *float64
	LegalClosePrice *float64
	Time            *DateTime
}

// rawSecurityData описывает параметры облигации, зависящие от даты
type rawSecurityData struct {
	SecurityID      string   `json:"SECID"`
	BoardID         string   `json:"BOARDID"`
	AccruedInterest *float64 `json:"ACCRUEDINT"`
	FaceValue       float64  `json:"FACEVALUE"`
	Currency        string   `json:"CURRENCYID"`
}

// rawMarketData описывает итоги торгов по облигации (сырые данные)
type rawMarketData struct {
	SecurityID      string   `json:"SECID"`
	BoardID         string   `json:"BOARDID"`
	Last            *float64 `json:"LAST"`
	LastChange      *float64 `json:"LASTCHANGE"`
	ClosePrice      *float64 `json:"CLOSEPRICE"`
	LegalClosePrice *float64 `json:"LCLOSEPRICE"`
	Time            DateTime `json:"SYSTIME"`
}

// GetMarketData возвращает текущие рыночные данные
func (p *provider) GetMarketData() ([]*MarketData, error) {
	values := make(url.Values)

	values.Set("iss.only", "marketdata")
	values.Set("iss.json", "extended")
	values.Set("iss.meta", "off")

	u := fmt.Sprintf("/iss/engines/stock/markets/bonds/securities.json?%s", values.Encode())

	resp := make([]marketDataResponse, 0)
	err := p.getJSON(u, &resp)
	if err != nil {
		return nil, err
	}

	collection := newMarketDataCollection()
	for _, respItem := range resp {
		if respItem.MarketData != nil {
			for _, s := range respItem.Securities {
				item := collection.GetOrAdd(s.SecurityID, s.BoardID)
				item.AccruedInterest = s.AccruedInterest
				item.FaceValue = &s.FaceValue
				item.Currency = &s.Currency
			}

			for _, m := range respItem.MarketData {
				item := collection.GetOrAdd(m.SecurityID, m.BoardID)

				item.Last = m.Last
				item.LastChange = m.LastChange
				item.ClosePrice = m.ClosePrice
				item.LegalClosePrice = m.LegalClosePrice
				item.Time = &m.Time
			}
		}
	}

	array := collection.GetItems()
	return array, nil
}

type marketDataResponse struct {
	Securities []*rawSecurityData `json:"securities"`
	MarketData []*rawMarketData   `json:"marketdata"`
}

type marketDataCollection struct {
	data map[string]map[string]*MarketData
}

func newMarketDataCollection() *marketDataCollection {
	return &marketDataCollection{
		data: make(map[string]map[string]*MarketData),
	}
}

func (collection *marketDataCollection) GetOrAdd(securityID, boardID string) *MarketData {
	inner, exists := collection.data[securityID]
	if !exists {
		inner = make(map[string]*MarketData)
		collection.data[securityID] = inner
	}

	item, exists := inner[boardID]
	if !exists {
		item = &MarketData{
			SecurityID: securityID,
			BoardID:    boardID,
		}

		inner[boardID] = item
	}

	return item
}

func (collection *marketDataCollection) GetItems() []*MarketData {
	array := make([]*MarketData, len(collection.data))
	array = array[0:0]

	for _, inner := range collection.data {
		for _, item := range inner {
			array = append(array, item)
		}
	}

	return array
}
