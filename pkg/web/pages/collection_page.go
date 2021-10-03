package pages

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

// CollectionPage обрабатывает запросы "GET /collections/:id"
func (ctrl *Controller) CollectionPage(c *gin.Context) {
	id := c.Param("id")
	model, err := NewCollectionPageModel(ctrl.app, c, id)
	if err != nil {
		if err == recommender.ErrNotFound {
			ctrl.renderHTML(c, http.StatusNotFound, "pages/collection_not_found", id)
			return
		}

		panic(err)
	}

	ctrl.renderHTML(c, http.StatusOK, "pages/collection", model)
}

// CollectionPageModel - модель для страницы "pages/collection.html"
type CollectionPageModel struct {
	ID               string
	Name             string
	ItemsPerDuration map[recommender.Duration][]CollectionPageItemModel
}

// NewCollectionPageModel создает новые объекты типа CollectionPageModel
func NewCollectionPageModel(app app.App, context context.Context, id string) (*CollectionPageModel, error) {
	u, err := app.NewUnitOfWork(context)
	if err != nil {
		return nil, err
	}
	defer u.Close()

	model := CollectionPageModel{}
	model.ItemsPerDuration = make(map[recommender.Duration][]CollectionPageItemModel)
	for _, duration := range recommender.Durations {
		collection, err := u.GetCollection(id)
		if err != nil {
			return nil, err
		}

		reports, err := u.ListCollectionBonds(collection.ID(), duration)
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

// CollectionPageItemModel - модель отдельной записи в коллекции (для страницы "pages/collection.html")
type CollectionPageItemModel struct {
	Bond   *data.Bond
	Issuer *data.Issuer
	Report *recommender.Report
}
