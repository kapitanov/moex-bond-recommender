package data

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kapitanov/bond-planner/pkg/data/migrations"
)

var (
	// ErrNotFound возвращается, если объект не был найден в БД
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists возвращается, если объект уже существует в БД
	ErrAlreadyExists = errors.New("already exists")
)

// DB предоставляет доступ к репозиториям БД
type DB interface {
	// BeginTX начинает новую транзакцию
	BeginTX() (*TX, error)
}

// TX представляет транзакцию БД
type TX struct {
	Issuers   IssuerRepository
	Bonds     BondRepository
	Payments  PaymentRepository
	db        *gorm.DB
	committed bool
}

// Commit фиксирует транзакцию
func (tx *TX) Commit() error {
	err := tx.db.Commit().Error
	if err != nil {
		return err
	}

	tx.committed = true
	return nil
}

// Close завершает транзакцию
func (tx *TX) Close() {
	if !tx.committed {
		tx.db.Rollback()
	}
}

// New создает новый объект DB
func New(options ...Option) (DB, error) {
	conf := &dbContextConfig{
		DSN: DefaultDataSource,
	}

	for _, f := range options {
		err := f(conf)
		if err != nil {
			return nil, err
		}
	}

	dialector := postgres.New(postgres.Config{
		DSN: conf.DSN,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = migrations.Up(db)
	if err != nil {
		return nil, err
	}

	return &dbContext{db}, nil
}

const (
	// DefaultDataSource содержит строку соединения с БД по умолчанию
	DefaultDataSource = "postgres://localhost:5432/bond_planner"
)

// Option конфигурирует контекст БД
type Option func(c *dbContextConfig) error

// WithDataSource задает строку соединения с БД
func WithDataSource(dsn string) Option {
	return func(c *dbContextConfig) error {
		c.DSN = dsn
		return nil
	}
}

type dbContextConfig struct {
	DSN string
}

type dbContext struct {
	db *gorm.DB
}

// BeginTX начинает новую транзакцию
func (ctx *dbContext) BeginTX() (*TX, error) {
	db := ctx.db.Begin()
	if db.Error != nil {
		return nil, db.Error
	}

	tx := &TX{
		Issuers:   &issuerRepository{db},
		Bonds:     &bondRepository{db},
		Payments:  &paymentRepository{db},
		db:        db,
		committed: false,
	}

	return tx, nil
}
