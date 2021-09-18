package data

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/kapitanov/moex-bond-recommender/pkg/data/migrations"
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
	Issuers    IssuerRepository
	Bonds      BondRepository
	Payments   PaymentRepository
	Offers     OfferRepository
	MarketData MarketDataRepository
	db         *gorm.DB
	committed  bool
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
		Log: log.New(io.Discard, "", 0),
	}

	for _, f := range options {
		err := f(conf)
		if err != nil {
			return nil, err
		}
	}

	db, err := conf.Connect()
	if err != nil {
		return nil, err
	}

	conf.Log.Printf("migrating db schema")
	err = migrations.Up(db)
	if err != nil {
		return nil, err
	}

	return &dbContext{db, conf.Log}, nil
}

const (
	// DefaultDataSource содержит строку соединения с БД по умолчанию
	DefaultDataSource = "postgres://postgres:postgres@localhost:5432/bond_planner"
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

// WithLogger задает логгер
func WithLogger(log *log.Logger) Option {
	return func(s *dbContextConfig) error {
		if log == nil {
			return fmt.Errorf("logger option is nil")
		}

		s.Log = log
		return nil
	}
}

type dbContextConfig struct {
	DSN string
	Log *log.Logger
}

func (c *dbContextConfig) Connect() (*gorm.DB, error) {
	c.Log.Printf("connecting to postgres at \"%s\"", c.DSN)
	cfg, err := pgx.ParseConfig(c.DSN)
	if err != nil {
		return nil, err
	}

	conn := stdlib.OpenDB(*cfg)
	r, err := conn.Query("SELECT 1")
	if err != nil {
		err = errors.Unwrap(err)
		switch err.(type) {
		case *pgconn.PgError:
			pgError := err.(*pgconn.PgError)
			if pgError.Code == "3D000" { // database <name> doesn't exists
				err = c.CreateDB(*cfg)
				if err != nil {
					return nil, err
				}

				r, err = conn.Query("SELECT 1")
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
			break

		default:
			return nil, err
		}
	}

	err = r.Close()
	if err != nil {
		return nil, err
	}

	dialector := postgres.New(postgres.Config{
		Conn: conn,
	})
	gormConf := &gorm.Config{
		Logger: logger.Discard,
	}
	db, err := gorm.Open(dialector, gormConf)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// CreateDB создает БД, считая, что она не существует
func (c *dbContextConfig) CreateDB(cfg pgx.ConnConfig) error {
	dbName := cfg.Database
	c.Log.Printf("database %s doesn't exist and will be created", dbName)

	cfg.Database = "postgres" // Default database

	conn := stdlib.OpenDB(cfg)
	defer conn.Close()

	sqlCommand := fmt.Sprintf("CREATE DATABASE \"%s\"", dbName)
	_, err := conn.Exec(sqlCommand)
	if err != nil {
		return err
	}

	return nil
}

type dbContext struct {
	db  *gorm.DB
	log *log.Logger
}

// BeginTX начинает новую транзакцию
func (ctx *dbContext) BeginTX() (*TX, error) {
	db := ctx.db.Begin(&sql.TxOptions{Isolation: sql.LevelSnapshot})
	if db.Error != nil {
		return nil, db.Error
	}

	tx := &TX{
		Issuers:    &issuerRepository{db},
		Bonds:      &bondRepository{db},
		Payments:   &paymentRepository{db},
		Offers:     &offerRepository{db},
		MarketData: &marketDataRepository{db},
		db:         db,
		committed:  false,
	}

	return tx, nil
}
