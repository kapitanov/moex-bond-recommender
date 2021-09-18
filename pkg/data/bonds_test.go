package data_test

import (
	"testing"

	assertion "github.com/stretchr/testify/assert"
	"gopkg.in/data-dog/go-sqlmock.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
)

func TestBond_Scan(t *testing.T) {
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

	mock.ExpectQuery("SELECT \\* FROM \"bonds\"").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "moex_id", "isin"}).
				AddRow("123", "456", "FooBar"))

	var bond data.Bond
	err = db.First(&bond).Error
	assert.Nil(err)
	assert.Equal(123, bond.ID)
	assert.Equal(456, bond.MoexID)
	assert.Equal("FooBar", bond.ISIN)
}
