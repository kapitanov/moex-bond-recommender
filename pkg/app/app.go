package app

import (
	"context"
	"log"

	"github.com/subchen/go-trylock"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/fetch"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
	"github.com/kapitanov/moex-bond-recommender/pkg/search"
)

// App является корневым контейнером для сервисов приложения
type App interface {
	// FetchStaticData выполняет выгрузку статических данных
	FetchStaticData(ctx context.Context) error

	// FetchMarketData выполняет выгрузку рыночных данных
	FetchMarketData(ctx context.Context) error

	// Search выполняет поиск облигации по тексту
	Search(req search.Request) (*search.Result, error)

	// ListCollections возвращает список коллекций рекомендаций
	ListCollections() []recommender.Collection

	// GetCollection возвращает коллекцию рекомендаций по ее ID
	GetCollection(ctx context.Context, id string, duration recommender.Duration) (recommender.Collection, []*recommender.Report, error)

	// GetReport возвращает отчет по отдельной облигации
	GetReport(ctx context.Context, idOrISIN string) (*recommender.Report, error)
}

type config struct {
	MoexURL     string
	PostgresURL string
}

// Option конфигурирует объект App
type Option func(c *config) error

// WithMoexURL задает корневой URL сервиса ISS
// По умолчанию используется moex.DefaultURL
func WithMoexURL(value string) Option {
	return func(c *config) error {
		c.MoexURL = value
		return nil
	}
}

// WithDataSource задает строку соединения с БД
func WithDataSource(value string) Option {
	return func(c *config) error {
		c.PostgresURL = value
		return nil
	}
}

// New создает новый объект App
func New(options ...Option) (App, error) {
	c := &config{
		MoexURL:     moex.DefaultURL,
		PostgresURL: data.DefaultDataSource,
	}

	for _, fn := range options {
		err := fn(c)
		if err != nil {
			return nil, err
		}
	}

	moexLogger := log.New(log.Writer(), "moex: ", log.Flags())
	provider, err := moex.NewProvider(moex.WithURL(c.MoexURL), moex.WithLogger(moexLogger))
	if err != nil {
		return nil, err
	}

	dataLogger := log.New(log.Writer(), "data: ", log.Flags())
	db, err := data.New(data.WithDataSource(c.PostgresURL), data.WithLogger(dataLogger))
	if err != nil {
		return nil, err
	}

	fetchLogger := log.New(log.Writer(), "fetch: ", log.Flags())
	fetchService, err := fetch.New(fetch.WithProvider(provider), fetch.WithLogger(fetchLogger))
	if err != nil {
		return nil, err
	}

	searchLogger := log.New(log.Writer(), "search: ", log.Flags())
	searchService, err := search.New(search.WithLogger(searchLogger))
	if err != nil {
		return nil, err
	}

	recommenderService, err := recommender.New()
	if err != nil {
		return nil, err
	}

	app := &appImpl{
		moexProvider:       provider,
		db:                 db,
		fetchService:       fetchService,
		searchService:      searchService,
		recommenderService: recommenderService,
		fetchInProgress:    trylock.New(),
	}
	return app, nil
}
