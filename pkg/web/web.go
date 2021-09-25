package web

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
)

// DefaultAddress - адрес для прослушивания по умолчанию
const DefaultAddress = "0.0.0.0:5000"

// Service - сервис веб приложения
type Service interface {
	// Start запускает веб приложение
	Start() error

	// Close завершает работу веб приложения
	Close()
}

// New создает новые объекты типа Service
func New(options ...Option) (Service, error) {
	gin.SetMode(gin.ReleaseMode)

	s := &service{
		router:  gin.New(),
		logger:  log.New(io.Discard, "", 0),
		address: DefaultAddress,
		done:    &sync.WaitGroup{},
	}

	for _, fn := range options {
		err := fn(s)
		if err != nil {
			return nil, err
		}
	}

	s.router.HTMLRender = ginview.New(goview.Config{
		Root:         "templates",
		Extension:    ".html",
		Master:       "layout",
		Partials:     []string{},
		Funcs:        defineFunctions(s.googleAnalyticsID),
		DisableCache: false,
		Delims:       goview.Delims{Left: "{{", Right: "}}"},
	})
	s.pagesController = &pagesController{app: s.app}
	s.ConfigureEndpoints()
	return s, nil
}

// Option настраивает веб приложение
type Option func(s *service) error

// WithLogger задает логгер
func WithLogger(logger *log.Logger) Option {
	return func(s *service) error {
		s.logger = logger
		return nil
	}
}

// WithListenAddress задает адрес для прослушивания
func WithListenAddress(address string) Option {
	return func(s *service) error {
		s.address = address
		return nil
	}
}

// WithApp задает экземпляр приложения app.App
func WithApp(app app.App) Option {
	return func(s *service) error {
		s.app = app
		return nil
	}
}

// WithGoogleAnalyticsID задает ID для Google Analytics
func WithGoogleAnalyticsID(value string) Option {
	return func(s *service) error {
		s.googleAnalyticsID = value
		return nil
	}
}

type service struct {
	router            *gin.Engine
	logger            *log.Logger
	address           string
	done              *sync.WaitGroup
	app               app.App
	pagesController   *pagesController
	server            *http.Server
	googleAnalyticsID string
}

// Start запускает веб приложение
func (s *service) Start() error {
	s.server = &http.Server{Addr: s.address}
	s.server.Handler = s.router

	go func() {
		s.logger.Printf("listening on \"%s\"\n", s.server.Addr)

		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("could not listen on \"%s\": %v\n", s.server.Addr, err)
		}

		s.done.Done()
	}()

	return nil
}

// Close завершает работу веб приложения
func (s *service) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.server.SetKeepAlivesEnabled(false)
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.logger.Fatalf("could not gracefully shutdown the server: %v\n", err)
	}
}

type pagesController struct {
	app app.App
}
