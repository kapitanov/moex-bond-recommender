package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

func (ctrl *pagesController) IndexPage(c *gin.Context) {
	model, err := NewIndexPageModel(ctrl.app)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "pages/index", model)
}

type IndexPageModel struct {
	Collections []CollectionModel
}

func NewIndexPageModel(app app.App) (*IndexPageModel, error) {
	collections := app.ListCollections()

	model := IndexPageModel{
		Collections: make([]CollectionModel, len(collections)),
	}

	for i, collection := range collections {
		c, err := NewCollectionModel(app, collection)
		if err != nil {
			return nil, err
		}

		model.Collections[i] = *c
	}

	return &model, nil
}

type CollectionModel struct {
	ID       string
	Name     string
	Duration recommender.Duration
	Bonds    []CollectionItemModel
}

func NewCollectionModel(app app.App, collection recommender.Collection) (*CollectionModel, error) {
	// TODO нужен рефакторинг - не хватает unit of work
	duration := recommender.Duration5Year
	_, reports, err := app.GetCollection(context.Background(), collection.ID(), duration)
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

type CollectionItemModel struct {
	Bond   *data.Bond
	Issuer *data.Issuer
	Report *recommender.Report
}
