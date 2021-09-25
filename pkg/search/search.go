package search

import (
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
)

// Request содержит параметры запроса поикса облигаций
type Request struct {
	Text  string
	Skip  int
	Limit int
}

// DefaultLimit содержит значение по умолчанию для Request.Limit
const DefaultLimit = 10

// Result содержит результаты поиска облигаций
type Result struct {
	Bonds      []*data.Bond
	TotalCount int
}

// Service предоставляет доступ к сервису поиска облигаций по тексту
type Service interface {
	// Do выполняет поиск по тексту
	Do(tx *data.TX, req Request) (*Result, error)

	// Rebuild выполняет перестроение поискового индекса
	Rebuild(ctx context.Context, tx *data.TX) error
}

// Option настраивает сервис поиска
type Option func(s *service) error

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

	return service, nil
}

type service struct {
	log *log.Logger
}

// Do выполняет поиск по тексту
func (s *service) Do(tx *data.TX, req Request) (*Result, error) {
	text := req.Text
	text = strings.TrimSpace(text)
	text = regexp.MustCompile("[^a-zA-Z0-9а-яА-Я]+").ReplaceAllString(text, "")
	if text == "" {
		res := Result{
			Bonds:      make([]*data.Bond, 0),
			TotalCount: 0,
		}
		return &res, nil
	}

	text = fmt.Sprintf("%s:*", text)

	skip := req.Skip
	if skip < 0 {
		skip = 0
	}

	limit := req.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}

	bonds, totalCount, err := tx.Search.Exec(text, skip, limit)
	if err != nil {
		return nil, err
	}

	res := Result{
		Bonds:      bonds,
		TotalCount: totalCount,
	}
	return &res, nil
}

// Rebuild выполняет перестроение поискового индекса
func (s *service) Rebuild(ctx context.Context, tx *data.TX) error {
	s.log.Printf("rebuilding search index")

	err := tx.Search.Rebuild()
	if err != nil {
		return err
	}

	s.log.Printf("search index in now up to date")
	return nil
}
