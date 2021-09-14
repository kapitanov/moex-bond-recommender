package moex

import (
	"fmt"
	"net/url"
)

// MarketData описывает итоги торгов по облигации
type MarketData struct {
	SecurityID      string   `json:"SECID"`
	Last            *float64 `json:"LAST"`
	LastChange      *float64 `json:"LASTCHANGE"`
	ClosePrice      *float64 `json:"CLOSEPRICE"`
	LegalClosePrice *float64 `json:"LCLOSEPRICE"`
	BoardID         string   `json:"BOARDID"`
	SeqNum          int64    `json:"SEQNUM"`
	SysTime         DateTime `json:"SYSTIME"`
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

	items := make([]*MarketData, 0)
	for _, respItem := range resp {
		if respItem.MarketData != nil {
			for _, item := range respItem.MarketData {
				items = append(items, item)
			}
		}
	}

	return items, nil
}

type marketDataResponse struct {
	MarketData []*MarketData `json:"marketdata"`
}
