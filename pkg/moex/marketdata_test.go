package moex_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	assertion "github.com/stretchr/testify/assert"

	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
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
        "securities": [
            {
                "SECID": "RU000A103D60",
                "BOARDID": "AUCT",
                "SHORTNAME": "КОБР-47",
                "PREVWAPRICE": null,
                "YIELDATPREVWAPRICE": 0,
                "COUPONVALUE": 16.27,
                "NEXTCOUPON": "2021-10-13",
                "ACCRUEDINT": 11.09,
                "PREVPRICE": null,
                "LOTSIZE": 1,
                "FACEVALUE": 1000,
                "BOARDNAME": "Размещение: Аукцион - безадрес.",
                "STATUS": "A",
                "MATDATE": "2021-10-13",
                "DECIMALS": 4,
                "COUPONPERIOD": 92,
                "ISSUESIZE": 300000000,
                "PREVLEGALCLOSEPRICE": null,
                "PREVADMITTEDQUOTE": null,
                "PREVDATE": "2021-09-13",
                "SECNAME": "КОБР-47",
                "REMARKS": null,
                "MARKETCODE": "FOND",
                "INSTRID": "BOBR",
                "SECTORID": null,
                "MINSTEP": 0.0001,
                "FACEUNIT": "SUR",
                "BUYBACKPRICE": 100,
                "BUYBACKDATE": "2021-10-13",
                "ISIN": "RU000A103D60",
                "LATNAME": "KOBR-47",
                "REGNUMBER": "4-47-22BR2-1",
                "CURRENCYID": "SUR",
                "ISSUESIZEPLACED": 189102266,
                "LISTLEVEL": 3,
                "SECTYPE": "5",
                "COUPONPERCENT": 6.750,
                "OFFERDATE": null,
                "SETTLEDATE": "2021-09-15",
                "LOTVALUE": 1000
            }
        ],
        "marketdata": [
            {
                "SECID": "RU000A103D60",
                "BID": null,
                "BIDDEPTH": null,
                "OFFER": null,
                "OFFERDEPTH": null,
                "SPREAD": 0,
                "BIDDEPTHT": null,
                "OFFERDEPTHT": null,
                "OPEN": null,
                "LOW": null,
                "HIGH": null,
                "LAST": 99.34,
                "LASTCHANGE": -0.01,
                "LASTCHANGEPRCNT": 0,
                "QTY": 0,
                "VALUE": 0.00,
                "YIELD": 0,
                "VALUE_USD": 0,
                "WAPRICE": null,
                "LASTCNGTOLASTWAPRICE": 0,
                "WAPTOPREVWAPRICEPRCNT": 0,
                "WAPTOPREVWAPRICE": 0,
                "YIELDATWAPRICE": 0,
                "YIELDTOPREVYIELD": 0,
                "CLOSEYIELD": 0,
                "CLOSEPRICE": null,
                "MARKETPRICETODAY": null,
                "MARKETPRICE": null,
                "LASTTOPREVPRICE": 0,
                "NUMTRADES": 0,
                "VOLTODAY": 0,
                "VALTODAY": 0,
                "VALTODAY_USD": 0,
                "BOARDID": "AUCT",
                "TRADINGSTATUS": "N",
                "UPDATETIME": "19:00:13",
                "DURATION": 29,
                "NUMBIDS": null,
                "NUMOFFERS": null,
                "CHANGE": null,
                "TIME": "19:00:13",
                "HIGHBID": null,
                "LOWOFFER": null,
                "PRICEMINUSPREVWAPRICE": null,
                "LASTBID": null,
                "LASTOFFER": null,
                "LCURRENTPRICE": null,
                "LCLOSEPRICE": null,
                "MARKETPRICE2": null,
                "ADMITTEDQUOTE": null,
                "OPENPERIODPRICE": null,
                "SEQNUM": 1985569,
                "SYSTIME": "2021-09-14 19:15:51",
                "VALTODAY_RUR": 0,
                "IRICPICLOSE": null,
                "BEICLOSE": null,
                "CBRCLOSE": null,
                "YIELDTOOFFER": null,
                "YIELDLASTCOUPON": null,
                "TRADINGSESSION": null
            }
        ]
    }
]
`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		u, err := url.Parse(req.RequestURI)
		if err != nil {
			panic(err)
		}

		switch u.Path {
		case "/iss/engines/stock/markets/bonds/securities.json":
			w.WriteHeader(200)
			w.Header().Set("content-type", "application/json")
			_, _ = w.Write([]byte(json))

		default:
			w.WriteHeader(404)
		}
	}))
	defer func() { testServer.Close() }()

	provider, err := moex.NewProvider(moex.WithURL(testServer.URL))
	if !assert.Nil(err) {
		return
	}

	list, err := provider.GetMarketData(context.Background())
	assert.Nil(err)
	assert.NotNil(list)
	assert.Equal(1, len(list))
	assert.Equal("RU000A103D60", list[0].SecurityID)
	assert.Equal("AUCT", list[0].BoardID)

	// "securities"
	assert.NotNil(list[0].AccruedInterest)
	assert.Equal(float64(11.09), *list[0].AccruedInterest)
	assert.NotNil(list[0].FaceValue)
	assert.Equal(float64(1000), *list[0].FaceValue)
	assert.NotNil(list[0].Currency)
	assert.Equal("SUR", *list[0].Currency)

	// "marketdata"
	assert.NotNil(list[0].Last)
	assert.Equal(float64(99.34), *list[0].Last)
	assert.NotNil(list[0].LastChange)
	assert.Equal(float64(-0.01), *list[0].LastChange)
	assert.Nil(list[0].ClosePrice)
	assert.Nil(list[0].LegalClosePrice)
	assert.Equal("2021-09-14 19:15:51", list[0].Time.String())
}
