package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

// IndexPage обрабатывает запросы "GET /"
func (ctrl *pagesController) IndexPage(c *gin.Context) {
	model, err := NewIndexPageModel(ctrl.app)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "pages/index", model)
}

// IndexPageModel - модель для страницы "pages/index.html"
type IndexPageModel struct {
	Collections []CollectionModel
}

// NewIndexPageModel создает объекты типа IndexPageModel
func NewIndexPageModel(app app.App) (*IndexPageModel, error) {
	u, err := app.NewUnitOfWork(context.Background())
	if err != nil {
		return nil, err
	}
	defer u.Close()

	collections := u.ListCollections()

	model := IndexPageModel{
		Collections: make([]CollectionModel, len(collections)),
	}

	for i, collection := range collections {
		c, err := NewCollectionModel(u, collection)
		if err != nil {
			return nil, err
		}

		model.Collections[i] = *c
	}

	return &model, nil
}

// CollectionModel - модель коллекции для страницы "pages/index.html"
type CollectionModel struct {
	ID       string
	Name     string
	Duration recommender.Duration
	Bonds    []CollectionItemModel
}

// NewCollectionModel создает объекты типа CollectionModel
func NewCollectionModel(u app.UnitOfWork, collection recommender.Collection) (*CollectionModel, error) {
	duration := recommender.Duration5Year
	reports, err := u.ListCollectionBonds(collection.ID(), duration)
	if err != nil {
		return nil, err
	}

	array := make([]CollectionItemModel, len(reports))
	for i, report := range reports {
		array[i] = CollectionItemModel{
			Bond:   report.Bond,
			Issuer: report.Issuer,
			Report: report,
		}
	}

	model := CollectionModel{
		ID:       collection.ID(),
		Name:     collection.Name(),
		Duration: duration,
		Bonds:    array,
	}
	return &model, nil
}

// CollectionItemModel - модель элемента коллекции для страницы "pages/index.html"
type CollectionItemModel struct {
	Bond   *data.Bond
	Issuer *data.Issuer
	Report *recommender.Report
}
