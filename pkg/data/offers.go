package data

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

// OfferType содержит тип оферты
type OfferType string

const (
	// GenericOffer - оферта
	GenericOffer OfferType = "offer"

	// CompletedGenericOffer - состоявшаяся оферта
	CompletedGenericOffer OfferType = "completed_offer"

	// CanceledGenericOffer - отмененнная оферта
	CanceledGenericOffer OfferType = "canceled_offer"

	// DefaultGenericOffer - дефолт оферты
	DefaultGenericOffer OfferType = "default_offer"

	// TechDefaultGenericOffer - технический дефолт оферты
	TechDefaultGenericOffer OfferType = "tech_default_offer"

	// MaturityOffer - оферта-погашение
	MaturityOffer OfferType = "maturity"

	// CanceledMaturityOffer - отмененнная оферта-погашение
	CanceledMaturityOffer OfferType = "canceled_maturity"
)

// Offer содержит данные по офертам
type Offer struct {
	ID         int          `gorm:"column:id; primaryKey"`
	BondID     int          `gorm:"column:bond_id"`
	IssueValue *float64     `gorm:"issue_value"`
	Date       sql.NullTime `gorm:"date"`
	StartDate  sql.NullTime `gorm:"start_date"`
	EndDate    sql.NullTime `gorm:"end_date"`
	FaceValue  *float64     `gorm:"face_value"`
	FaceUnit   string       `gorm:"face_unit"`
	Price      *float64     `gorm:"column:price"`
	Value      *float64     `gorm:"column:value"`
	Agent      *string      `gorm:"column:agent"`
	Type       *OfferType   `gorm:"column:type"`
	CreatedAt  time.Time    `gorm:"column:created"`
	UpdatedAt  time.Time    `gorm:"column:updated"`
	Bond       Bond
}

// TableName задает название таблицы
func (Offer) TableName() string {
	return "offers"
}

// CreateOfferArgs содержит параметры для создания оферты
type CreateOfferArgs struct {
	BondID     int
	IssueValue *float64
	Date       sql.NullTime
	StartDate  sql.NullTime
	EndDate    sql.NullTime
	FaceValue  *float64
	FaceUnit   string
	Price      *float64
	Value      *float64
	Agent      *string
	Type       *OfferType
}

// OfferRepository отвечает за управление записями в таблице оферт
type OfferRepository interface {
	// Create создает новую выплату по оферте
	// Если указанная оферта уже существует, то возвращается ошибка ErrAlreadyExists
	Create(args CreateOfferArgs) (*Offer, error)
}

type offerRepository struct {
	db *gorm.DB
}

// Create создает новую выплату по оферте
// Если указанная оферта уже существует, то возвращается ошибка ErrAlreadyExists
func (repo *offerRepository) Create(args CreateOfferArgs) (*Offer, error) {
	var offer Offer
	err := repo.db.First(&offer, "bond_id = ? AND date::date = (?)::date AND start_date::date = (?)::date AND end_date::date = (?)::date",
		args.BondID, args.Date, args.StartDate, args.EndDate).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		return nil, ErrAlreadyExists
	}

	now := time.Now().UTC()
	offer = Offer{
		BondID:     args.BondID,
		IssueValue: args.IssueValue,
		Date:       args.Date,
		StartDate:  args.StartDate,
		EndDate:    args.EndDate,
		FaceValue:  args.FaceValue,
		FaceUnit:   args.FaceUnit,
		Price:      args.Price,
		Value:      args.Value,
		Agent:      args.Agent,
		Type:       args.Type,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	err = repo.db.Create(&offer).Error
	if err != nil {
		return nil, err
	}

	return &offer, nil
}
