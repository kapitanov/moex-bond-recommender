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

func TestProvider_ListCoupons(t *testing.T) {
	assert := assertion.New(t)

	json1 := `
[
    {
        "charsetinfo": {
            "name": "utf-8"
        }
    },
    {
        "coupons": [
            {
                "isin": null,
                "name": "Читинская область-1",
                "issuevalue": null,
                "coupondate": "1997-11-30",
                "recorddate": null,
                "startdate": "1997-05-30",
                "initialfacevalue": null,
                "facevalue": 10000,
                "faceunit": "RUB",
                "value": 500,
                "valueprc": null,
                "value_rub": 500
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
        "coupons": [
            {
                "isin": "XS2075963293",
                "name": "Eurasia Capital S.A. UNDT",
                "issuevalue": 200000000,
                "coupondate": "2111-01-01",
                "recorddate": null,
                "startdate": "2021-11-07",
                "initialfacevalue": 1000,
                "facevalue": 1000,
                "faceunit": "USD",
                "value": null,
                "valueprc": null,
                "value_rub": null
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
        "coupons": [ ]
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

	it := provider.ListCoupons(context.Background(), moex.CouponListQuery{})

	{
		coupons, err := it.Next()
		assert.Nil(err)
		assert.NotNil(coupons)
		assert.Equal(1, len(coupons))
		assert.Equal("", coupons[0].ISIN)
		assert.Equal("Читинская область-1", coupons[0].Name)
		assert.Nil(coupons[0].IssueValue)
		assert.Equal("1997-11-30", coupons[0].CouponDate.String())
		assert.Equal("", coupons[0].RecordDate.String())
		assert.Equal("1997-05-30", coupons[0].StartDate.String())
		assert.Nil(coupons[0].InitialFaceValue)
		assert.Equal(float64(10000), *coupons[0].FaceValue)
		assert.Equal(float64(500), *coupons[0].Value)
		assert.Nil(coupons[0].ValuePercent)
		assert.Equal(float64(500), *coupons[0].ValueRub)
	}

	{
		coupons, err := it.Next()
		assert.Nil(err)
		assert.NotNil(coupons)
		assert.Equal(1, len(coupons))
		assert.Equal("XS2075963293", coupons[0].ISIN)
		assert.Equal("Eurasia Capital S.A. UNDT", coupons[0].Name)
		assert.Equal(float64(200000000), *coupons[0].IssueValue)
		assert.Equal("2111-01-01", coupons[0].CouponDate.String())
		assert.Equal("", coupons[0].RecordDate.String())
		assert.Equal("2021-11-07", coupons[0].StartDate.String())
		assert.Equal(float64(1000), *coupons[0].InitialFaceValue)
		assert.Equal(float64(1000), *coupons[0].FaceValue)
		assert.Nil(coupons[0].Value)
		assert.Nil(coupons[0].ValuePercent)
		assert.Nil(coupons[0].ValueRub)
	}

	{
		_, err = it.Next()
		assert.Equal(moex.EOF, err)
	}
}
