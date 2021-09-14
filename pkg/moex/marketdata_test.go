package moex_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	assertion "github.com/stretchr/testify/assert"

	"github.com/kapitanov/bond-planner/pkg/moex"
)

func TestProvider_GetMarketData(t *testing.T) {
	assert := assertion.New(t)

	json := `
[
    {
        "charsetinfo": {
            "name": "utf-8"
        }
    },
    {
        "marketdata": [
            {
                "SECID": "RU000A0JNYN1",
                "BID": 99.34,
                "BIDDEPTH": null,
                "OFFER": 99.35,
                "OFFERDEPTH": null,
                "SPREAD": 0.01,
                "BIDDEPTHT": 11572,
                "OFFERDEPTHT": 40084,
                "OPEN": 99.9,
                "LOW": 99.22,
                "HIGH": 99.9,
                "LAST": 99.34,
                "LASTCHANGE": -0.01,
                "LASTCHANGEPRCNT": -0.01,
                "QTY": 1,
                "VALUE": 993.40,
                "YIELD": 7.03,
                "VALUE_USD": 13.65,
                "WAPRICE": 99.43,
                "LASTCNGTOLASTWAPRICE": -0.05,
                "WAPTOPREVWAPRICEPRCNT": 0.04,
                "WAPTOPREVWAPRICE": 0.04,
                "YIELDATWAPRICE": 6.9,
                "YIELDTOPREVYIELD": -0.06,
                "CLOSEYIELD": 0,
                "CLOSEPRICE": null,
                "MARKETPRICETODAY": null,
                "MARKETPRICE": 99.39,
                "LASTTOPREVPRICE": -0.06,
                "NUMTRADES": 117,
                "VOLTODAY": 1099,
                "VALTODAY": 1092739,
                "VALTODAY_USD": 15018,
                "BOARDID": "TQCB",
                "TRADINGSTATUS": "T",
                "UPDATETIME": "17:02:00",
                "DURATION": 266,
                "NUMBIDS": null,
                "NUMOFFERS": null,
                "CHANGE": -0.06,
                "TIME": "16:57:05",
                "HIGHBID": null,
                "LOWOFFER": null,
                "PRICEMINUSPREVWAPRICE": -0.05,
                "LASTBID": null,
                "LASTOFFER": null,
                "LCURRENTPRICE": 99.34,
                "LCLOSEPRICE": null,
                "MARKETPRICE2": null,
                "ADMITTEDQUOTE": null,
                "OPENPERIODPRICE": null,
                "SEQNUM": 863392,
                "SYSTIME": "2021-09-13 17:17:05",
                "VALTODAY_RUR": 1092739,
                "IRICPICLOSE": null,
                "BEICLOSE": null,
                "CBRCLOSE": null,
                "YIELDTOOFFER": null,
                "YIELDLASTCOUPON": null,
                "TRADINGSESSION": "1"
            }
        ]
    }
]`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		u, err := url.Parse(req.RequestURI)
		if err != nil {
			panic(err)
		}

		switch u.Path {
		case "/iss/engines/stock/markets/bonds/securities.json":
			w.WriteHeader(200)
			w.Header().Set("content-type", "application/json")
			w.Write([]byte(json))
			break

		default:
			w.WriteHeader(404)
			break
		}
	}))
	defer func() { testServer.Close() }()

	provider, err := moex.NewProvider(moex.WithURL(testServer.URL))
	if !assert.Nil(err) {
		return
	}

	list, err := provider.GetMarketData()
	assert.Nil(err)
	assert.NotNil(list)
	assert.Equal(1, len(list))
	assert.Equal("RU000A0JNYN1", list[0].SecurityID)
	assert.NotNil(list[0].Last)
	assert.Equal(float64(99.34), *list[0].Last)
	assert.NotNil(list[0].LastChange)
	assert.Equal(float64(-0.01), *list[0].LastChange)
	assert.Nil(list[0].ClosePrice)
	assert.Nil(list[0].LegalClosePrice)
	assert.Equal("TQCB", list[0].BoardID)
	assert.Equal(int64(863392), list[0].SeqNum)
	assert.Equal("2021-09-13 17:17:05", list[0].SysTime.String())
}
