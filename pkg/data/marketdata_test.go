package data_test

import (
	"testing"

	assertion "github.com/stretchr/testify/assert"
	"gopkg.in/data-dog/go-sqlmock.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kapitanov/bond-planner/pkg/data"
)

func TestMarketData_Scan(t *testing.T) {
	assert := assertion.New(t)

	conn, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		panic(err)
	}

	mock.ExpectQuery("SELECT \\* FROM \"marketdata\"").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "bond_id", "last"}).
				AddRow(123, 456, 123.45))

	var marketData data.MarketData
	err = db.First(&marketData).Error
	assert.Nil(err)
	assert.Equal(123, marketData.ID)
	assert.Equal(456, marketData.BondID)
	assert.NotNil(marketData.Last)
	assert.Equal(123.45, *marketData.Last)
}
