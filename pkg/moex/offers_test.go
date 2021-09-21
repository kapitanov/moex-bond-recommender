package moex_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	assertion "github.com/stretchr/testify/assert"

	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

func TestProvider_ListOffers(t *testing.T) {
	assert := assertion.New(t)

	json1 := `
[
    {
        "charsetinfo": {
            "name": "utf-8"
        }
    },
    {
        "offers": [
            {
                "isin": "RU000A0JRDY3",
                "name": "ДОМ.РФ (АО) обл. сер.А18",
                "issuevalue": 7000000000,
                "offerdate": "0000-00-00",
                "offerdatestart": "2020-01-09",
                "offerdateend": "2020-01-15",
                "facevalue": 500,
                "faceunit": "RUB",
                "price": 100,
                "value": null,
                "agent": null,
                "offertype": "Оферта"
            }
        ]
    }
]`
	json2 := `
[
    {
        "charsetinfo": {
            "name": "utf-8"
        }
    },
    {
        "offers": [
            {
                "isin": "RU000A0JRJC6",
                "name": "Волга-Спорт АО обл. 01",
                "issuevalue": 1400000000,
                "offerdate": "0000-00-00",
                "offerdatestart": "2020-07-30",
                "offerdateend": "2020-08-06",
                "facevalue": 1000,
                "faceunit": "RUB",
                "price": 100,
                "value": null,
                "agent": null,
                "offertype": "Оферта"
            }
        ]
    }
]`
	json3 := `
[
    {
        "charsetinfo": {
            "name": "utf-8"
        }
    },
    {
        "offers": [ ]
    }
]`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		u, err := url.Parse(req.RequestURI)
		if err != nil {
			panic(err)
		}

		switch u.Path {
		case "/iss/statistics/engines/stock/markets/bonds/bondization.json":
			str := u.Query().Get("start")

			var body []byte
			body = []byte(json1)
			if str != "" {
				start, err := strconv.Atoi(str)
				if err != nil {
					panic(err)
				}

				if start > 0 {
					body = []byte(json2)
				}

				if start > 1 {
					body = []byte(json3)
				}
			}

			w.WriteHeader(200)
			w.Header().Set("content-type", "application/json")
			w.Write(body)
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

	it := provider.ListOffers(context.Background(), moex.OfferListQuery{})

	{
		offers, err := it.Next()
		assert.Nil(err)
		assert.NotNil(offers)
		assert.Equal(1, len(offers))
		assert.Equal("RU000A0JRDY3", offers[0].ISIN)
		assert.Equal("ДОМ.РФ (АО) обл. сер.А18", offers[0].Name)
		assert.Equal(float64(7000000000), *offers[0].IssueValue)
		assert.Equal("", offers[0].OfferDate.String())
		assert.Equal("2020-01-09", offers[0].StartOfferDate.String())
		assert.Equal("2020-01-15", offers[0].EndOfferDate.String())
		assert.Equal(float64(500), *offers[0].FaceValue)
		assert.Equal("RUB", offers[0].FaceUnit)
		assert.Equal(float64(100), *offers[0].Price)
		assert.Nil(offers[0].Value)
		assert.Nil(offers[0].Agent)
		assert.NotNil(offers[0].Type)
		assert.Equal(moex.GenericOffer, *offers[0].Type)
	}

	{
		offers, err := it.Next()
		assert.Nil(err)
		assert.NotNil(offers)
		assert.Equal(1, len(offers))
		assert.Equal("RU000A0JRJC6", offers[0].ISIN)
		assert.Equal("Волга-Спорт АО обл. 01", offers[0].Name)
		assert.Equal(float64(1400000000), *offers[0].IssueValue)
		assert.Equal("", offers[0].OfferDate.String())
		assert.Equal("2020-07-30", offers[0].StartOfferDate.String())
		assert.Equal("2020-08-06", offers[0].EndOfferDate.String())
		assert.Equal(float64(1000), *offers[0].FaceValue)
		assert.Equal("RUB", offers[0].FaceUnit)
		assert.Equal(float64(100), *offers[0].Price)
		assert.Nil(offers[0].Value)
		assert.Nil(offers[0].Agent)
		assert.NotNil(offers[0].Type)
		assert.Equal(moex.GenericOffer, *offers[0].Type)
	}

	{
		_, err = it.Next()
		assert.Equal(moex.EOF, err)
	}
}
