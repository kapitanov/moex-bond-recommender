package data_test

import (
	"testing"

	assertion "github.com/stretchr/testify/assert"
	"gopkg.in/data-dog/go-sqlmock.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
)

func TestCashFlowItem_Scan(t *testing.T) {
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

	mock.ExpectQuery("SELECT \\* FROM \"cashflow\"").
		WillReturnRows(
			sqlmock.NewRows([]string{"bond_id", "type", "value_rub"}).
				AddRow(123, "C", 45.67))

	var item data.CashFlowItem
	err = db.First(&item).Error
	assert.Nil(err)
	assert.Equal(123, item.BondID)
	assert.Equal(data.CouponPayment, item.Type)
	assert.Equal(float64(45.67), item.ValueRub)
}
