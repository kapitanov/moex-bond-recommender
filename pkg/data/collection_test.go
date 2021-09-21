package data_test

import (
	"testing"

	assertion "github.com/stretchr/testify/assert"
	"gopkg.in/data-dog/go-sqlmock.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
)

func TestCollectionBondRef_Scan(t *testing.T) {
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

	mock.ExpectQuery("SELECT \\* FROM").
		WillReturnRows(
			sqlmock.NewRows([]string{"bond_id"}).
				AddRow(123))

	var item data.CollectionBondRef
	err = db.First(&item).Error
	assert.Nil(err)
	assert.Equal(123, item.BondID)
}
