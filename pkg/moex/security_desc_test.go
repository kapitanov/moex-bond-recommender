package moex_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	assertion "github.com/stretchr/testify/assert"

	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

func TestProperty_AsString(t *testing.T) {
	assert := assertion.New(t)

	str := `{"name": "FACEUNIT", "title": "Валюта номинала", "value": "SUR", "type": "string", "sort_order": 11, "is_hidden": 0, "precision": null}`
	var prop moex.Property
	err := json.Unmarshal([]byte(str), &prop)
	assert.Nil(err)
	assert.Equal(moex.FaceUnitProperty, prop.Name)

	value, err := prop.AsString()
	assert.Nil(err)
	assert.Equal("SUR", value)

	_, err = prop.AsDate()
	assert.Equal(moex.ErrWrongPropertyType, err)

	_, err = prop.AsFloat64()
	assert.Equal(moex.ErrWrongPropertyType, err)

	_, err = prop.AsBool()
	assert.Equal(moex.ErrWrongPropertyType, err)
}

func TestProperty_AsDate(t *testing.T) {
	assert := assertion.New(t)

	str := `{"name": "ISSUEDATE", "title": "Дата начала торгов", "value": "2019-05-22", "type": "date", "sort_order": 7, "is_hidden": 0, "precision": null}`
	var prop moex.Property
	err := json.Unmarshal([]byte(str), &prop)
	assert.Nil(err)
	assert.Equal(moex.IssueDateProperty, prop.Name)

	value, err := prop.AsDate()
	assert.Nil(err)
	assert.Equal("2019-05-22", value.String())

	_, err = prop.AsString()
	assert.Equal(moex.ErrWrongPropertyType, err)

	_, err = prop.AsFloat64()
	assert.Equal(moex.ErrWrongPropertyType, err)

	_, err = prop.AsBool()
	assert.Equal(moex.ErrWrongPropertyType, err)
}

func TestProperty_AsFloat64(t *testing.T) {
	assert := assertion.New(t)

	str := `{"name": "INITIALFACEVALUE", "title": "Первоначальная номинальная стоимость", "value": "1000", "type": "number", "sort_order": 10, "is_hidden": 0, "precision": null}`
	var prop moex.Property
	err := json.Unmarshal([]byte(str), &prop)
	assert.Nil(err)
	assert.Equal(moex.InitialFaceValueProperty, prop.Name)

	value, err := prop.AsFloat64()
	assert.Nil(err)
	assert.Equal(float64(1000), value)

	_, err = prop.AsString()
	assert.Equal(moex.ErrWrongPropertyType, err)

	_, err = prop.AsDate()
	assert.Equal(moex.ErrWrongPropertyType, err)

	_, err = prop.AsBool()
	assert.Equal(moex.ErrWrongPropertyType, err)
}

func TestProperty_AsBool(t *testing.T) {
	assert := assertion.New(t)

	str := `{"name": "ISQUALIFIEDINVESTORS", "title": "Бумаги для квалифицированных инвесторов", "value": "1", "type": "boolean", "sort_order": 46, "is_hidden": 0, "precision": 0}`
	var prop moex.Property
	err := json.Unmarshal([]byte(str), &prop)
	assert.Nil(err)
	assert.Equal(moex.IsForQualifiedInvestorsOnlyProperty, prop.Name)

	value, err := prop.AsBool()
	assert.Nil(err)
	assert.Equal(true, value)

	_, err = prop.AsString()
	assert.Equal(moex.ErrWrongPropertyType, err)

	_, err = prop.AsDate()
	assert.Equal(moex.ErrWrongPropertyType, err)

	_, err = prop.AsFloat64()
	assert.Equal(moex.ErrWrongPropertyType, err)
}

func TestProvider_GetSecurityDescription(t *testing.T) {
	assert := assertion.New(t)

	json := `
[
    {
        "charsetinfo": {
            "name": "utf-8"
        }
    },
    {
        "description": [
            {
                "name": "ISSUEDATE",
                "title": "Дата начала торгов",
                "value": "2019-05-22",
                "type": "date",
                "sort_order": 7,
                "is_hidden": 0,
                "precision": null
            },
            {
                "name": "INITIALFACEVALUE",
                "title": "Первоначальная номинальная стоимость",
                "value": "1000",
                "type": "number",
                "sort_order": 10,
                "is_hidden": 0,
                "precision": null
            },
            {
                "name": "FACEUNIT",
                "title": "Валюта номинала",
                "value": "SUR",
                "type": "string",
                "sort_order": 11,
                "is_hidden": 0,
                "precision": null
            },
            {
                "name": "ISQUALIFIEDINVESTORS",
                "title": "Бумаги для квалифицированных инвесторов",
                "value": "0",
                "type": "boolean",
                "sort_order": 46,
                "is_hidden": 0,
                "precision": 0
            }
        ]
    }
]`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(json))
	}))
	defer func() { testServer.Close() }()

	provider, err := moex.NewProvider(moex.WithURL(testServer.URL))
	if !assert.Nil(err) {
		return
	}

	desc, err := provider.GetSecurityDescription(context.Background(), "TESTISIN")
	assert.Nil(err)
	assert.NotNil(desc)
	assert.NotNil(desc.Properties)

	assert.Equal(4, len(desc.Properties))

	_, exists := desc.Properties[moex.IssueDateProperty]
	assert.True(exists)

	_, exists = desc.Properties[moex.InitialFaceValueProperty]
	assert.True(exists)

	_, exists = desc.Properties[moex.FaceUnitProperty]
	assert.True(exists)

	_, exists = desc.Properties[moex.IsForQualifiedInvestorsOnlyProperty]
	assert.True(exists)
}
