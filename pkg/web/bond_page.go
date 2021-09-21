package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

func (ctrl *pagesController) BondPage(c *gin.Context) {
	id := c.Param("id")

	model, err := NewBondPageModel(ctrl.app, id)
	if err != nil {
		if err == recommender.ErrNotFound || err == data.ErrNotFound {
			c.HTML(http.StatusNotFound, "pages/bond_not_found", id)
			return
		}

		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "pages/bond", model)
}

type BondPageModel struct {
	Bond   *data.Bond
	Issuer *data.Issuer
	Report *recommender.Report
}

func NewBondPageModel(app app.App, id string) (*BondPageModel, error) {
	report, err := app.GetReport(context.Background(), id)
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
