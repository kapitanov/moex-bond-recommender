package fetch

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

// BondFetchStats содержит статистику выгрузки облигаций
type BondFetchStats struct {
	NewIssuers int
	NewBonds   int
}

// PaymentFetchStats содержит статистику выгрузки выплат
type PaymentFetchStats struct {
	NewCoupons       int
	NewAmortizations int
	NewMaturities    int
}

// OfferFetchStats содержит статистику выгрузки оферт
type OfferFetchStats struct {
	NewOffers int
}

// MarketDataFetchStats содержит статистику выгрузки рыночных данных
type MarketDataFetchStats struct {
	NewMarketData int
}

// Service содержит функции для выгрузки данных из биржи в БД
type Service interface {
	// FetchBonds выполняет выгрузку облигаций из биржи в БД
	FetchBonds(ctx context.Context, tx *data.TX) (*BondFetchStats, error)

	// FetchPayments выполняет выгрузку выплат из биржи в БД
	FetchPayments(ctx context.Context, tx *data.TX) (*PaymentFetchStats, error)

	// FetchOffers выполняет выгрузку оферт из биржи в БД
	FetchOffers(ctx context.Context, tx *data.TX) (*OfferFetchStats, error)

	// FetchMarketData выполняет выгрузку рыночных данных из биржи в БД
	FetchMarketData(ctx context.Context, tx *data.TX) (*MarketDataFetchStats, error)
}

// Option настраивает сервис
type Option func(s *service) error

// WithProvider задает провайдера данных с биржи
func WithProvider(provider moex.Provider) Option {
	return func(s *service) error {
		s.provider = provider
		return nil
	}
}

// WithDB задает контекст БД
func WithDB(db data.DB) Option {
	return func(s *service) error {
		s.db = db
		return nil
	}
}

// WithLogger задает логгер
func WithLogger(log *log.Logger) Option {
	return func(s *service) error {
		if log == nil {
			return fmt.Errorf("logger option is nil")
		}

		s.log = log
		return nil
	}
}

// New создает новый объект типа Service
func New(options ...Option) (Service, error) {
	service := &service{
		log: log.New(io.Discard, "", 0),
	}
	for _, fn := range options {
		err := fn(service)
		if err != nil {
			return nil, err
		}
	}

	if service.db == nil {
		return nil, fmt.Errorf("missing required option WithDB")
	}
	if service.provider == nil {
		return nil, fmt.Errorf("missing required option WithProvider")
	}

	return service, nil
}

type service struct {
	provider moex.Provider
	db       data.DB
	log      *log.Logger
}

// FetchBonds выполняет выгрузку облигаций из биржи в БД
func (s *service) FetchBonds(ctx context.Context, tx *data.TX) (*BondFetchStats, error) {
	start := time.Now()

	w := &bondFetchWorker{
		provider: s.provider,
		tx:       tx,
		log:      s.log,
		stats:    &BondFetchStats{},
	}

	err := w.FetchBonds(ctx)
	if err != nil {
		return nil, err
	}

	end := time.Now()

	duration := end.Sub(start)
	s.log.Printf("fetch completed, %d new issuer(s) and %d new bond(s) were fetched in %s",
		w.stats.NewIssuers, w.stats.NewBonds, duration.Round(time.Second))

	return w.stats, nil
}

// FetchPayments выполняет выгрузку выплат из биржи в БД
func (s *service) FetchPayments(ctx context.Context, tx *data.TX) (*PaymentFetchStats, error) {
	start := time.Now()

	w := &paymentFetchWorker{
		provider: s.provider,
		tx:       tx,
		log:      s.log,
		stats:    &PaymentFetchStats{},
		bondIDs:  make(map[string]int),
	}

	err := w.FetchCoupons(ctx)
	if err != nil {
		return nil, err
	}

	err = w.FetchAmortizations(ctx)
	if err != nil {
		return nil, err
	}

	end := time.Now()

	duration := end.Sub(start)
	s.log.Printf("fetch completed, %d new coupon(s), %d new amortization(s) and %d maturity(es) were fetched in %s",
		w.stats.NewCoupons, w.stats.NewAmortizations, w.stats.NewMaturities, duration.Round(time.Second))

	return w.stats, nil
}

// FetchOffers выполняет выгрузку оферт из биржи в БД
func (s *service) FetchOffers(ctx context.Context, tx *data.TX) (*OfferFetchStats, error) {
	start := time.Now()

	w := &offerFetchWorker{
		provider: s.provider,
		tx:       tx,
		log:      s.log,
		stats:    &OfferFetchStats{},
		bondIDs:  make(map[string]int),
	}

	err := w.FetchOffers(ctx)
	if err != nil {
		return nil, err
	}

	end := time.Now()

	duration := end.Sub(start)
	s.log.Printf("fetch completed, %d new offer(s) were fetched in %s", w.stats.NewOffers, duration.Round(time.Second))

	return w.stats, nil
}

// FetchMarketData выполняет выгрузку рыночных данных из биржи в БД
func (s *service) FetchMarketData(ctx context.Context, tx *data.TX) (*MarketDataFetchStats, error) {
	start := time.Now()

	w := &marketDataFetchWorker{
		provider: s.provider,
		tx:       tx,
		log:      s.log,
		stats:    &MarketDataFetchStats{},
		bondIDs:  make(map[string]int),
	}

	err := w.FetchMarketData(ctx)
	if err != nil {
		return nil, err
	}

	end := time.Now()

	duration := end.Sub(start)
	s.log.Printf("fetch completed, %d new market data record(s) were fetched in %s", w.stats.NewMarketData, duration.Round(time.Second))

	return w.stats, nil
}
