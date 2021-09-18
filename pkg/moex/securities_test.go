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

func TestProvider_ListSecurities(t *testing.T) {
	assert := assertion.New(t)

	json1 := `
[
    {
        "charsetinfo": {
            "name": "utf-8"
        }
    },
    {
        "securities": [
            {
                "id": 66310675,
                "secid": "RU000A100CN3",
                "shortname": "Якут-12 об",
                "regnumber": "RU35012RSY0",
                "name": "Республика Саха (Якутия) об.12",
                "isin": "RU000A100CN3",
                "is_traded": 1,
                "emitent_id": 1372,
                "emitent_title": "Министерство финансов Республики Саха (Якутия)",
                "emitent_inn": "1435027673",
                "emitent_okpo": "00063006",
                "gosreg": "RU35012RSY0",
                "type": "subfederal_bond",
                "group": "stock_bonds",
                "primary_boardid": "TQCB",
                "marketprice_boardid": "TQCB"
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
        "securities": [
            {
                "id": 93129041,
                "secid": "RU000A100JC1",
                "shortname": "ЕАБР 1Р-04",
                "regnumber": null,
                "name": "ЕАБР БО 001Р-04",
                "isin": "RU000A100JC1",
                "is_traded": 1,
                "emitent_id": 2258,
                "emitent_title": "Евразийский банк развития",
                "emitent_inn": "9909220306",
                "emitent_okpo": null,
                "gosreg": null,
                "type": "exchange_bond",
                "group": "stock_bonds",
                "primary_boardid": "TQCB",
                "marketprice_boardid": "TQCB"
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
        "securities": [ ]
    }
]`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		u, err := url.Parse(req.RequestURI)
		if err != nil {
			panic(err)
		}

		switch u.Path {
		case "/iss/securities.json":
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

	it := provider.ListSecurities(context.Background(), moex.SecurityListQuery{})

	{
		securities, err := it.Next()
		assert.Nil(err)
		assert.NotNil(securities)
		assert.Equal(1, len(securities))
		assert.Equal(int64(66310675), securities[0].ID)
		assert.Equal("RU000A100CN3", securities[0].SecurityID)
		assert.Equal("Якут-12 об", securities[0].ShortName)
		assert.Equal("RU35012RSY0", securities[0].RegNumber)
		assert.Equal("Республика Саха (Якутия) об.12", securities[0].Name)
		assert.Equal("RU000A100CN3", securities[0].ISIN)
		assert.Equal(moex.IsTraded, securities[0].IsTraded)
		assert.Equal(int64(1372), securities[0].IssuerId)
		assert.Equal("Министерство финансов Республики Саха (Якутия)", securities[0].IssuerName)
		assert.Equal("1435027673", securities[0].IssuerINN)
		assert.Equal("00063006", securities[0].IssuerOKPO)
		assert.Equal(moex.SubfederalBond, securities[0].Type)
		assert.Equal("TQCB", securities[0].PrimaryBoardID)
		assert.Equal("TQCB", securities[0].MarketPriceBoardID)
	}

	{
		securities, err := it.Next()
		assert.Nil(err)
		assert.NotNil(securities)
		assert.Equal(1, len(securities))
		assert.Equal(int64(93129041), securities[0].ID)
		assert.Equal("RU000A100JC1", securities[0].SecurityID)
		assert.Equal("ЕАБР 1Р-04", securities[0].ShortName)
		assert.Equal("", securities[0].RegNumber)
		assert.Equal("ЕАБР БО 001Р-04", securities[0].Name)
		assert.Equal("RU000A100JC1", securities[0].ISIN)
		assert.Equal(moex.IsTraded, securities[0].IsTraded)
		assert.Equal(int64(2258), securities[0].IssuerId)
		assert.Equal("Евразийский банк развития", securities[0].IssuerName)
		assert.Equal("9909220306", securities[0].IssuerINN)
		assert.Equal("", securities[0].IssuerOKPO)
		assert.Equal(moex.ExchangeBond, securities[0].Type)
		assert.Equal("TQCB", securities[0].PrimaryBoardID)
		assert.Equal("TQCB", securities[0].MarketPriceBoardID)
	}

	{
		_, err = it.Next()
		assert.Equal(moex.EOF, err)
	}
}
