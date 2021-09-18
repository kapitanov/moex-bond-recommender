package data_test

import (
	"testing"

	assertion "github.com/stretchr/testify/assert"
	"gopkg.in/data-dog/go-sqlmock.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
)

func TestIssuer_Scan(t *testing.T) {
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

	mock.ExpectQuery("SELECT \\* FROM \"issuers\"").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "moex_id", "name", "inn", "okpo"}).
				AddRow("123", "456", "FooBar", nil, "OKPO"))

	var issuer data.Issuer
	err = db.First(&issuer).Error
	assert.Nil(err)
	assert.Equal(123, issuer.ID)
	assert.Equal(456, issuer.MoexID)
	assert.Equal("FooBar", issuer.Name)
	assert.Nil(issuer.INN)
	assert.NotNil(issuer.OKPO)
	assert.Equal("OKPO", *issuer.OKPO)
}
