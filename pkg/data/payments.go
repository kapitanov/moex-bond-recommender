package data

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/kapitanov/bond-planner/pkg/moex"
)

// PaymentType содерижт тип выплаты
type PaymentType string

const (
	MaturityPayment     PaymentType = "M"
	CouponPayment       PaymentType = "C"
	OfferPayment        PaymentType = "O"
	AmortizationPayment PaymentType = "A"
)

// OfferType содержит тип оферты
type OfferType string

const (
	// GenericOffer - оферта
	GenericOffer OfferType = OfferType(moex.GenericOffer)

	// CompletedGenericOffer - состоявшаяся оферта
	CompletedGenericOffer OfferType = OfferType(moex.CompletedGenericOffer)

	// CanceledGenericOffer - отмененнная оферта
	CanceledGenericOffer OfferType = OfferType(moex.CanceledGenericOffer)

	// DefaultGenericOffer - дефолт оферты
	DefaultGenericOffer OfferType = OfferType(moex.DefaultGenericOffer)

	// TechDefaultGenericOffer - технический дефолт оферты
	TechDefaultGenericOffer OfferType = OfferType(moex.TechDefaultGenericOffer)

	// MaturityOffer - оферта-погашение
	MaturityOffer OfferType = OfferType(moex.MaturityOffer)

	// CanceledMaturityOffer - отмененнная оферта-погашение
	CanceledMaturityOffer OfferType = OfferType(moex.CanceledMaturityOffer)
)

// Payment содержит данные по выплатам (погашения, купоны, оферты, амортизации)
type Payment struct {
	/* Общие поля */
	ID               int          `gorm:"column:id; primaryKey"`
	BondID           int          `gorm:"column:bond_id"`
	Type             PaymentType  `gorm:"column:type"`
	Date             time.Time    `gorm:"column:date"`
	Value            float64      `gorm:"column:value"`
	ValuePercent     float64      `gorm:"column:value_percent"`
	ValueRub         float64      `gorm:"column:value_rub"`
	CouponRecordDate sql.NullTime `gorm:"column:coupon_record_date"`
	CouponStartDate  sql.NullTime `gorm:"column:coupon_start_date"`
	OfferPrice       *float64     `gorm:"column:offer_price"`
	OfferValue       *float64     `gorm:"column:offer_value"`
	OfferAgent       *string      `gorm:"column:offer_agent"`
	OfferType        *OfferType   `gorm:"column:offer_type"`
	CreatedAt        time.Time    `gorm:"created"`
	UpdatedAt        time.Time    `gorm:"updated"`
	Bond             Bond
}

// TableName задает название таблицы
func (Payment) TableName() string {
	return "payments"
}

// PaymentListQuery содержит параметры запроса списка выплат
type PaymentListQuery struct {
	BondID int
	Types  []PaymentType
	Since  *time.Time
}

// CreatePaymentArgs содержит параметры для создания выплаты (общие для всех типов выплат)
type CreatePaymentArgs struct {
	BondID       int
	Date         time.Time
	Value        float64
	ValuePercent float64
	ValueRub     float64
}

// CreateCouponPaymentArgs содержит параметры для создания выплаты по купону
type CreateCouponPaymentArgs struct {
	CreatePaymentArgs
	RecordDate *time.Time
	StartDate  *time.Time
}

// CreateAmortizationPaymentArgs содержит параметры для создания выплаты по амортизации
type CreateAmortizationPaymentArgs struct {
	CreatePaymentArgs
}

// CreateOfferPaymentArgs содержит параметры для создания выплаты по оферте
type CreateOfferPaymentArgs struct {
	CreatePaymentArgs
	Price *float64
	Value *float64
	Agent string
	Type  OfferType
}

// CreateMaturityPaymentArgs содержит параметры для создания выплаты по погашению
type CreateMaturityPaymentArgs struct {
	CreatePaymentArgs
}

// PaymentRepository отвечает за управление записями в таблице выплат
type PaymentRepository interface {
	// List возвращает полный список выплат, удовлетворяющих фильтру
	// Направление сортировки - по возрастанию даты
	List(query PaymentListQuery) ([]*Payment, error)

	// Get возвращает выплату по ее параметрам
	// Если указанной выплаты не существует, то возвращается ошибка ErrNotFound
	Get(bondID int, date time.Time, paymentType PaymentType) (*Payment, error)

	// CreateCoupon создает новую выплату по купону
	// Если указанная выплата уже существует, то возвращается ошибка ErrAlreadyExists
	CreateCoupon(args CreateCouponPaymentArgs) (*Payment, error)

	// CreateAmortization создает новую выплату по амортизации
	// Если указанная выплата уже существует, то возвращается ошибка ErrAlreadyExists
	CreateAmortization(args CreateAmortizationPaymentArgs) (*Payment, error)

	// CreateMaturity создает новую выплату по погашению
	// Если указанная выплата уже существует, то возвращается ошибка ErrAlreadyExists
	CreateMaturity(args CreateMaturityPaymentArgs) (*Payment, error)

	// CreateOffer создает новую выплату по оферте
	// Если указанная выплата уже существует, то возвращается ошибка ErrAlreadyExists
	CreateOffer(args CreateOfferPaymentArgs) (*Payment, error)
}

type paymentRepository struct {
	db *gorm.DB
}

// List возвращает полный список выплат, удовлетворяющих фильтру
// Направление сортировки - по возрастанию даты
func (repo *paymentRepository) List(query PaymentListQuery) ([]*Payment, error) {
	q := repo.db

	q = q.Where("bond_id = ?", query.BondID)
	if query.Types != nil && len(query.Types) > 0 {
		q = q.Where("type IN ?", query.Types)
	}
	if query.Since != nil {
		q = q.Where("date >= ?", query.Since)
	}

	q = q.Order("date ASC").Order("type ASC")

	var payments []*Payment
	err := q.Find(&payments).Error
	if err != nil {
		return nil, err
	}

	return payments, nil
}

// Get возвращает выплату по ее параметрам
// Если указанной выплаты не существует, то возвращается ошибка ErrNotFound
func (repo *paymentRepository) Get(bondID int, date time.Time, paymentType PaymentType) (*Payment, error) {
	var payment Payment
	err := repo.db.First(&payment, "bond_id = ? AND date::date = (?)::date AND type = ?", bondID, date, paymentType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &payment, nil
}

// CreateCoupon создает новую выплату по купону
// Если указанная выплата уже существует, то возвращается ошибка ErrAlreadyExists
func (repo *paymentRepository) CreateCoupon(args CreateCouponPaymentArgs) (*Payment, error) {
	var fn = func(payment *Payment) error {
		if args.RecordDate != nil {
			payment.CouponRecordDate = sql.NullTime{Time: *args.RecordDate, Valid: true}
		}

		if args.StartDate != nil {
			payment.CouponStartDate = sql.NullTime{Time: *args.StartDate, Valid: true}
		}

		return nil
	}

	return repo.Create(args.CreatePaymentArgs, CouponPayment, fn)
}

// CreateAmortization создает новую выплату по амортизации
// Если указанная выплата уже существует, то возвращается ошибка ErrAlreadyExists
func (repo *paymentRepository) CreateAmortization(args CreateAmortizationPaymentArgs) (*Payment, error) {
	var fn = func(payment *Payment) error {
		return nil
	}

	return repo.Create(args.CreatePaymentArgs, AmortizationPayment, fn)
}

// CreateMaturity создает новую выплату по погашению
// Если указанная выплата уже существует, то возвращается ошибка ErrAlreadyExists
func (repo *paymentRepository) CreateMaturity(args CreateMaturityPaymentArgs) (*Payment, error) {
	var fn = func(payment *Payment) error {
		return nil
	}

	return repo.Create(args.CreatePaymentArgs, MaturityPayment, fn)
}

// CreateOffer создает новую выплату по оферте
// Если указанная выплата уже существует, то возвращается ошибка ErrAlreadyExists
func (repo *paymentRepository) CreateOffer(args CreateOfferPaymentArgs) (*Payment, error) {
	var fn = func(payment *Payment) error {
		payment.OfferPrice = args.Price
		payment.OfferAgent = &args.Agent
		payment.OfferType = &args.Type
		return nil
	}

	return repo.Create(args.CreatePaymentArgs, OfferPayment, fn)
}

// Create создает новую выплату
// Если указанная выплата уже существует, то возвращается ошибка ErrAlreadyExists
func (repo *paymentRepository) Create(args CreatePaymentArgs, t PaymentType, fn func(*Payment) error) (*Payment, error) {
	_, err := repo.Get(args.BondID, args.Date, t)
	if err != nil {
		if err != ErrNotFound {
			return nil, err
		}
	} else {
		return nil, ErrAlreadyExists
	}

	now := time.Now().UTC()
	payment := &Payment{
		BondID:       args.BondID,
		Type:         t,
		Date:         args.Date,
		Value:        args.Value,
		ValuePercent: args.ValuePercent,
		ValueRub:     args.ValueRub,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = fn(payment)
	if err != nil {
		return nil, err
	}

	err = repo.db.Create(payment).Error
	if err != nil {
		return nil, err
	}

	return payment, nil
}
