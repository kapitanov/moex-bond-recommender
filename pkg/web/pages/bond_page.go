package pages

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

// BondPage обрабатывает запросы "GET /bonds/:id"
func (ctrl *Controller) BondPage(c *gin.Context) {
	id := c.Param("id")

	model, err := NewBondPageModel(ctrl.app, c, id)
	if err != nil {
		if err == recommender.ErrNotFound || err == data.ErrNotFound {
			ctrl.renderHTML(c, http.StatusNotFound, "pages/bond_not_found", id)
			return
		}

		panic(err)
	}

	ctrl.renderHTML(c, http.StatusOK, "pages/bond", model)
}

// BondPageModel - модель для страницы "pages/bond.html"
type BondPageModel struct {
	Bond   *data.Bond
	Issuer *data.Issuer
	Report *recommender.Report
}

// NewBondPageModel создает новые объекты типа BondPageModel
func NewBondPageModel(app app.App, context context.Context, id string) (*BondPageModel, error) {
	u, err := app.NewUnitOfWork(context)
	if err != nil {
		return nil, err
	}
	defer u.Close()

	report, err := u.GetReport(id)
	if err != nil {
		return nil, err
	}

	model := BondPageModel{
		Bond:   report.Bond,
		Issuer: report.Issuer,
		Report: report,
	}
	return &model, nil
}
