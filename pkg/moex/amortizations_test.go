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

func TestProvider_ListAmortizations(t *testing.T) {
	assert := assertion.New(t)

	json1 := `
[
    {
        "charsetinfo": {
            "name": "utf-8"
        }
    },
    {
        "amortizations": [
            {
                "isin": "RU0009161418",
                "name": "\"Джэй Эф Си Инт\" ОАО обл 01",
                "issuevalue": 700000000,
                "amortdate": "2004-04-08",
                "facevalue": 700,
                "initialfacevalue": 1000,
                "faceunit": "RUB",
                "valueprc": 15.00,
                "value": 150,
                "value_rub": 150,
                "data_source": "amortization"
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
        "amortizations": [
            {
                "isin": "RU0008967625",
                "name": "ОАО \"ПИК\" обл 4в.",
                "issuevalue": 750000000,
                "amortdate": "2004-10-02",
                "facevalue": 250,
                "initialfacevalue": 1000,
                "faceunit": "RUB",
                "valueprc": 25.00,
                "value": 250,
                "value_rub": 250,
                "data_source": "maturity"
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
        "amortizations": [ ]
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
			_, _ = w.Write(body)

		default:
			w.WriteHeader(404)
		}
	}))
	defer func() { testServer.Close() }()

	provider, err := moex.NewProvider(moex.WithURL(testServer.URL))
	assert.Nil(err)

	it := provider.ListAmortizations(context.Background(), moex.AmortizationListQuery{})

	{
		amortizations, err := it.Next()
		assert.Nil(err)
		assert.NotNil(amortizations)
		assert.Equal(1, len(amortizations))
		assert.Equal("RU0009161418", amortizations[0].ISIN)
		assert.Equal("\"Джэй Эф Си Инт\" ОАО обл 01", amortizations[0].Name)
		assert.Equal(float64(700000000), amortizations[0].IssueValue)
		assert.Equal("2004-04-08", amortizations[0].AmortDate.String())
		assert.Equal(float64(700), amortizations[0].FaceValue)
		assert.Equal(float64(1000), amortizations[0].InitialFaceValue)
		assert.Equal("RUB", amortizations[0].FaceUnit)
		assert.Equal(float64(15.00), amortizations[0].ValuePercent)
		assert.Equal(float64(150), amortizations[0].Value)
		assert.Equal(float64(150), amortizations[0].ValueRub)
		assert.Equal(moex.AmortizationTypeA, amortizations[0].Type)
	}

	{
		amortizations, err := it.Next()
		assert.Nil(err)
		assert.NotNil(amortizations)
		assert.Equal(1, len(amortizations))
		assert.Equal("RU0008967625", amortizations[0].ISIN)
		assert.Equal("ОАО \"ПИК\" обл 4в.", amortizations[0].Name)
		assert.Equal(float64(750000000), amortizations[0].IssueValue)
		assert.Equal("2004-10-02", amortizations[0].AmortDate.String())
		assert.Equal(float64(250), amortizations[0].FaceValue)
		assert.Equal(float64(1000), amortizations[0].InitialFaceValue)
		assert.Equal("RUB", amortizations[0].FaceUnit)
		assert.Equal(float64(25.00), amortizations[0].ValuePercent)
		assert.Equal(float64(250), amortizations[0].Value)
		assert.Equal(float64(250), amortizations[0].ValueRub)
		assert.Equal(moex.AmortizationTypeM, amortizations[0].Type)
	}

	{
		_, err = it.Next()
		assert.Equal(moex.EOF, err)
	}
}
