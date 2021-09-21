package data

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

// BondType содержит тип облигации
type BondType string

const (
	// SubfederalBond - субфедеральные облигации
	SubfederalBond BondType = BondType(moex.SubfederalBond)

	// OFZBond - ОФЗ
	OFZBond BondType = BondType(moex.OFZBond)

	// ExchangeBond - биржевые облигации
	ExchangeBond BondType = BondType(moex.ExchangeBond)

	// CBBond - облигации ЦБ
	CBBond BondType = BondType(moex.CBBond)

	// MunicipalBond - мунициальные облигации
	MunicipalBond BondType = BondType(moex.MunicipalBond)

	// CorporateBond - корпоративные облигации
	CorporateBond BondType = BondType(moex.CorporateBond)

	// IFIBond - облигации МФО
	IFIBond BondType = BondType(moex.IFIBond)

	// EuroBond - еврооблигации
	EuroBond BondType = BondType(moex.EuroBond)
)

// Bond содержит данные облигаций
type Bond struct {
	ID                 int          `gorm:"column:id; primaryKey"`
	IssuerID           int          `gorm:"column:issuer_id"`
	MoexID             int          `gorm:"column:moex_id; unique"`
	SecurityID         string       `gorm:"column:security_id; unique"`
	ShortName          string       `gorm:"column:short_name"`
	FullName           string       `gorm:"column:full_name"`
	ISIN               string       `gorm:"column:isin"`
	IsTraded           bool         `gorm:"column:is_traded"`
	QualifiedOnly      bool         `gorm:"column:qualified_only"`
	IsHighRisk         bool         `gorm:"column:high_risk"`
	Type               BondType     `gorm:"column:type"`
	PrimaryBoardID     string       `gorm:"column:primary_board_id"`
	MarketPriceBoardID string       `gorm:"column:market_price_board_id"`
	InitialFaceValue   float64      `gorm:"column:initial_face_value"`
	FaceUnit           string       `gorm:"column:face_unit"`
	IssueDate          sql.NullTime `gorm:"column:issue_date"`
	MaturityDate       sql.NullTime `gorm:"column:maturity_date"`
	ListingLevel       int          `gorm:"column:listing_level"`
	CouponFrequency    int          `gorm:"column:coupon_freq"`
	CreatedAt          time.Time    `gorm:"column:created"`
	UpdatedAt          time.Time    `gorm:"column:updated"`
	Issuer             Issuer
	Payments           []Payment `gorm:"foreignKey:BondID"`
}

// TableName задает название таблицы
func (Bond) TableName() string {
	return "bonds"
}

// CreateBondArgs содержит данные для создания облигации
type CreateBondArgs struct {
	IssuerID           int
	MoexID             int
	SecurityID         string
	ShortName          string
	FullName           string
	ISIN               string
	IsTraded           bool
	QualifiedOnly      bool
	IsHighRisk         bool
	Type               BondType
	PrimaryBoardID     string
	MarketPriceBoardID string
	InitialFaceValue   float64
	FaceUnit           string
	IssueDate          sql.NullTime
	MaturityDate       sql.NullTime
	ListingLevel       int
	CouponFrequency    int
}

// UpdateBondArgs содержит данные для создания облигации
type UpdateBondArgs struct {
	IsTraded           bool
	QualifiedOnly      bool
	PrimaryBoardID     string
	MarketPriceBoardID string
	IssueDate          sql.NullTime
	MaturityDate       sql.NullTime
}

// BondRepository отвечает за управление записями в таблице облигаций
type BondRepository interface {
	// GetByID выполняет поиски облигации по полю Bond.ID
	// Если облигация не найдена, возвращается ошибка ErrNotFound
	GetByID(id int) (*Bond, error)

	// GetByMoexID выполняет поиски облигации по полю Bond.MoexID
	// Если облигация не найдена, возвращается ошибка ErrNotFound
	GetByMoexID(moexID int) (*Bond, error)

	// GetByISIN выполняет поиски облигации по полю Bond.ISIN
	// Если облигация не найдена, возвращается ошибка ErrNotFound
	GetByISIN(isin string) (*Bond, error)

	// GetBySecurityID выполняет поиски облигации по полю Bond.SecurityID
	// Если облигация не найдена, возвращается ошибка ErrNotFound
	GetBySecurityID(securityID string) (*Bond, error)

	// Create создает облигацию
	// Если облигация уже существует, то возвращается ErrAlreadyExists
	Create(args CreateBondArgs) (*Bond, error)

	// Update обновляет данные эмитента
	// Если эмитент не найден, возвращается ошибка ErrNotFound
	Update(id int, args UpdateBondArgs) (*Bond, error)
}

type bondRepository struct {
	db *gorm.DB
}

// GetByID выполняет поиски облигации по полю Bond.ID
// Если облигация не найдена, возвращается ошибка ErrNotFound
func (repo *bondRepository) GetByID(id int) (*Bond, error) {
	var bond Bond
	err := repo.db.First(&bond, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &bond, nil
}

// GetByMoexID выполняет поиски облигации по полю Bond.MoexID
// Если облигация не найдена, возвращается ошибка ErrNotFound
func (repo *bondRepository) GetByMoexID(moexID int) (*Bond, error) {
	var bond Bond
	err := repo.db.First(&bond, "moex_id = ?", moexID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &bond, nil
}

// GetByISIN выполняет поиски облигации по полю Bond.ISIN
// Если облигация не найдена, возвращается ошибка ErrNotFound
func (repo *bondRepository) GetByISIN(isin string) (*Bond, error) {
	var bond Bond
	err := repo.db.First(&bond, "isin = ?", isin).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &bond, nil
}

// GetBySecurityID выполняет поиски облигации по полю Bond.SecurityID
// Если облигация не найдена, возвращается ошибка ErrNotFound
func (repo *bondRepository) GetBySecurityID(securityID string) (*Bond, error) {
	var bond Bond
	err := repo.db.First(&bond, "security_id = ?", securityID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &bond, nil
}

// Create создает облигацию
// Если облигация уже существует, то возвращается ErrAlreadyExists
func (repo *bondRepository) Create(args CreateBondArgs) (*Bond, error) {
	_, err := repo.GetByMoexID(args.MoexID)
	if err != nil {
		if err != ErrNotFound {
			return nil, err
		}
	} else {
		return nil, ErrAlreadyExists
	}

	_, err = repo.GetByISIN(args.ISIN)
	if err != nil {
		if err != ErrNotFound {
			return nil, err
		}
	} else {
		return nil, ErrAlreadyExists
	}

	_, err = repo.GetBySecurityID(args.SecurityID)
	if err != nil {
		if err != ErrNotFound {
			return nil, err
		}
	} else {
		return nil, ErrAlreadyExists
	}

	now := time.Now().UTC()
	bond := Bond{
		ID:                 0,
		IssuerID:           args.IssuerID,
		MoexID:             args.MoexID,
		SecurityID:         args.SecurityID,
		ShortName:          args.ShortName,
		FullName:           args.FullName,
		ISIN:               args.ISIN,
		IsTraded:           args.IsTraded,
		QualifiedOnly:      args.QualifiedOnly,
		IsHighRisk:         args.IsHighRisk,
		Type:               args.Type,
		PrimaryBoardID:     args.PrimaryBoardID,
		MarketPriceBoardID: args.MarketPriceBoardID,
		InitialFaceValue:   args.InitialFaceValue,
		FaceUnit:           args.FaceUnit,
		IssueDate:          args.IssueDate,
		MaturityDate:       args.MaturityDate,
		ListingLevel:       args.ListingLevel,
		CouponFrequency:    args.CouponFrequency,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	err = repo.db.Create(&bond).Error
	if err != nil {
		return nil, err
	}

	return &bond, nil
}

// Update обновляет данные эмитента
// Если эмитент не найден, возвращается ошибка ErrNotFound
func (repo *bondRepository) Update(id int, args UpdateBondArgs) (*Bond, error) {
	bond, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	hasChanges := false
	if bond.IsTraded != args.IsTraded {
		bond.IsTraded = args.IsTraded
		hasChanges = true
	}
	if bond.QualifiedOnly != args.QualifiedOnly {
		bond.QualifiedOnly = args.QualifiedOnly
		hasChanges = true
	}
	if bond.PrimaryBoardID != args.PrimaryBoardID {
		bond.PrimaryBoardID = args.PrimaryBoardID
		hasChanges = true
	}
	if bond.MarketPriceBoardID != args.MarketPriceBoardID {
		bond.MarketPriceBoardID = args.MarketPriceBoardID
		hasChanges = true
	}
	if bond.IssueDate != args.IssueDate {
		bond.IssueDate = args.IssueDate
		hasChanges = true
	}
	if bond.MaturityDate != args.MaturityDate {
		bond.MaturityDate = args.MaturityDate
		hasChanges = true
	}

	if !hasChanges {
		return bond, nil
	}

	bond.UpdatedAt = time.Now().UTC()
	err = repo.db.Updates(bond).Error
	if err != nil {
		return nil, err
	}

	return bond, nil
}
