package data

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// MarketData содержит последние рыночные данные по облигации
type MarketData struct {
	ID              int       `gorm:"column:id; primaryKey"`
	BondID          int       `gorm:"column:bond_id"`
	Time            time.Time `gorm:"column:time"`
	FaceValue       *float64  `gorm:"column:face_value"`
	Currency        *string   `gorm:"column:currency"`
	Last            *float64  `gorm:"column:last"`
	LastChange      *float64  `gorm:"column:last_change"`
	ClosePrice      *float64  `gorm:"column:close_price"`
	LegalClosePrice *float64  `gorm:"column:legal_close_price"`
	AccruedInterest *float64  `gorm:"column:accrued_interest"`
	Bond            Bond
}

// TableName задает название таблицы
func (MarketData) TableName() string {
	return "marketdata"
}

// PutMarketDataArgs содержит параметры для записи рыночных данных
type PutMarketDataArgs struct {
	Time            time.Time
	FaceValue       *float64
	Currency        *string
	Last            *float64
	LastChange      *float64
	ClosePrice      *float64
	LegalClosePrice *float64
	AccruedInterest *float64
}

// MarketDataRepository отвечает за управление записями в таблице рыночных данных
type MarketDataRepository interface {
	// Get возвращает запись для указанной облигации
	// Если указанной записи не существует, то возвращается ошибка ErrNotFound
	Get(bondID int) (*MarketData, error)

	// Put записывает рыночные данные для указанной облигации
	// Если рыночные данные уже существуют, они обновляются
	Put(bondID int, args PutMarketDataArgs) (*MarketData, error)
}

type marketDataRepository struct {
	db *gorm.DB
}

// Get возвращает запись для указанной облигации
// Если указанной записи не существует, то возвращается ошибка ErrNotFound
func (repo *marketDataRepository) Get(bondID int) (*MarketData, error) {
	var marketData MarketData
	err := repo.db.First(&marketData, "bond_id = ?", bondID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &marketData, nil
}

// Put записывает рыночные данные для указанной облигации
// Если рыночные данные уже существуют, они обновляются
func (repo *marketDataRepository) Put(bondID int, args PutMarketDataArgs) (*MarketData, error) {
	marketData, err := repo.Get(bondID)
	if err != nil {
		if err != ErrNotFound {
			return nil, err
		}

		marketData = &MarketData{
			BondID: bondID,
		}
		repo.PutValues(marketData, &args)
		err = repo.db.Create(marketData).Error
		if err != nil {
			return nil, err
		}

		return marketData, nil
	}

	repo.PutValues(marketData, &args)
	err = repo.db.Updates(marketData).Error
	if err != nil {
		return nil, err
	}

	return marketData, nil
}

// PutValues записывает значения из PutMarketDataArgs в MarketData
func (repo *marketDataRepository) PutValues(marketData *MarketData, args *PutMarketDataArgs) {
	marketData.Time = args.Time
	marketData.FaceValue = args.FaceValue
	marketData.Currency = args.Currency
	marketData.Last = args.Last
	marketData.LastChange = args.LastChange
	marketData.ClosePrice = args.ClosePrice
	marketData.LegalClosePrice = args.LegalClosePrice
	marketData.AccruedInterest = args.AccruedInterest
}
