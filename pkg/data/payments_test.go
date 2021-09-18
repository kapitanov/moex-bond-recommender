package data_test

import (
	"testing"

	assertion "github.com/stretchr/testify/assert"
	"gopkg.in/data-dog/go-sqlmock.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
)

func TestPayment_Scan(t *testing.T) {
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

	mock.ExpectQuery("SELECT \\* FROM \"payments\"").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "bond_id", "type", "value"}).
				AddRow(123, 456, string(data.MaturityPayment), 123.45))

	var payment data.Payment
	err = db.First(&payment).Error
	assert.Nil(err)
	assert.Equal(123, payment.ID)
	assert.Equal(456, payment.BondID)
	assert.Equal(data.MaturityPayment, payment.Type)
	assert.Equal(123.45, payment.Value)
}
