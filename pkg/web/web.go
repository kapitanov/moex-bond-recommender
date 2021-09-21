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

const DefaultAddress = "0.0.0.0:5000"

type Service interface {
	Start(ctx context.Context) error
	Close()
}

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
		Funcs:        defineFunctions(),
		DisableCache: false,
		Delims:       goview.Delims{Left: "{{", Right: "}}"},
	})
	s.pagesController = &pagesController{s.app}
	s.ConfigureEndpoints()
	return s, nil
}

type Option func(s *service) error

func WithLogger(logger *log.Logger) Option {
	return func(s *service) error {
		s.logger = logger
		return nil
	}
}

func WithListenAddress(address string) Option {
	return func(s *service) error {
		s.address = address
		return nil
	}
}

func WithApp(app app.App) Option {
	return func(s *service) error {
		s.app = app
		return nil
	}
}

type service struct {
	router          *gin.Engine
	logger          *log.Logger
	address         string
	done            *sync.WaitGroup
	app             app.App
	pagesController *pagesController
}

func (s *service) Start(ctx context.Context) error {
	http.Handle("/", s.router)

	server := &http.Server{Addr: s.address}

	go func() {
		s.logger.Printf("listening on \"%s\"\n", server.Addr)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("could not listen on \"%s\": %v\n", server.Addr, err)
		}

		s.done.Done()
	}()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		err := server.Shutdown(ctx)
		if err != nil {
			s.logger.Fatalf("could not gracefully shutdown the server: %v\n", err)
		}
	}()

	return nil
}

func (s *service) Close() {

}

type pagesController struct {
	app app.App
}
