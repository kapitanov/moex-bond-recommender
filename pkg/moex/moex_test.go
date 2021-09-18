package moex_test

import (
	"encoding/json"
	"testing"

	assertion "github.com/stretchr/testify/assert"

	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

type testDateStruct struct {
	Date moex.Date `json:"date"`
}

func TestDate_UnmarshalJSON(t *testing.T) {
	assert := assertion.New(t)

	{
		var s testDateStruct
		err := json.Unmarshal([]byte("{ \"date\": \"2021-01-15\" }"), &s)

		assert.Nil(err)
		assert.NotNil(s.Date)
		assert.Equal("2021-01-15", s.Date.Format("2006-01-02"))
	}
}

type testNullableDateStruct struct {
	Date moex.NullableDate `json:"date"`
}

func TestNullableDate_UnmarshalJSON(t *testing.T) {
	assert := assertion.New(t)

	{
		var s testNullableDateStruct
		err := json.Unmarshal([]byte("{}"), &s)

		assert.Nil(err)
		assert.Nil(s.Date.Time())
	}

	{
		var s testNullableDateStruct
		err := json.Unmarshal([]byte("{ \"date\": null }"), &s)

		assert.Nil(err)
		assert.Nil(s.Date.Time())
	}

	{
		var s testNullableDateStruct
		err := json.Unmarshal([]byte("{ \"date\": \"0000-00-00\" }"), &s)

		assert.Nil(err)
		assert.Nil(s.Date.Time())
	}

	{
		var s testNullableDateStruct
		err := json.Unmarshal([]byte("{ \"date\": \"2021-01-15\" }"), &s)

		assert.Nil(err)
		assert.NotNil(s.Date.Time())
		assert.Equal("2021-01-15", s.Date.Format("2006-01-02"))
	}
}

type testDateTimeStruct struct {
	Date moex.DateTime `json:"date"`
}

func TestDateTime_UnmarshalJSON(t *testing.T) {
	assert := assertion.New(t)

	{
		var s testDateTimeStruct
		err := json.Unmarshal([]byte("{ \"date\": \"2021-09-13 09:41:52\" }"), &s)

		assert.Nil(err)
		assert.NotNil(s.Date)
		assert.Equal("2021-09-13 09:41:52", s.Date.Format("2006-01-02 15:04:05"))
	}
}
