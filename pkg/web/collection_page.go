package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

func (ctrl *pagesController) CollectionPage(c *gin.Context) {
	id := c.Param("id")
	model, err := NewCollectionPageModel(ctrl.app, id)
	if err != nil {
		if err == recommender.ErrNotFound {
			c.HTML(http.StatusNotFound, "pages/collection_not_found", id)
			return
		}

		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "pages/collection", model)
}

type CollectionPageModel struct {
	ID               string
	Name             string
	ItemsPerDuration map[recommender.Duration][]CollectionPageItemModel
}

func NewCollectionPageModel(app app.App, id string) (*CollectionPageModel, error) {
	// TODO нужен рефакторинг - не хватает unit of work

	model := CollectionPageModel{}
	model.ItemsPerDuration = make(map[recommender.Duration][]CollectionPageItemModel)
	for _, duration := range recommender.Durations {
		collection, reports, err := app.GetCollection(context.Background(), id, duration)
		if err != nil {
			return nil, err
		}

		array := make([]CollectionPageItemModel, len(reports))
		for i, report := range reports {
			array[i] = CollectionPageItemModel{
				Bond:   report.Bond,
				Issuer: report.Issuer,
				Report: report,
			}
		}

		model.ID = collection.ID()
		model.Name = collection.Name()
		model.ItemsPerDuration[duration] = array
	}

	return &model, nil
}

type CollectionPageItemModel struct {
	Bond   *data.Bond
	Issuer *data.Issuer
	Report *recommender.Report
}
