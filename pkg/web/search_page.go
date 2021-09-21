package web

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/search"
)

func (ctrl *pagesController) SearchPage(c *gin.Context) {
	query, exists := c.GetQuery("q")
	if !exists {
		c.Redirect(http.StatusFound, "/")
		return
	}

	model, err := NewSearchPageModel(ctrl.app, query)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(model.Bonds) == 1 {
		c.Redirect(http.StatusFound, fmt.Sprintf("/bonds/%s", url.PathEscape(model.Bonds[0].ISIN)))
		return
	}

	c.HTML(http.StatusOK, "pages/search", model)
}

type SearchPageModel struct {
	Query      string
	Bonds      []*data.Bond
	TotalCount int
}

func NewSearchPageModel(app app.App, query string) (*SearchPageModel, error) {
	response, err := app.Search(search.Request{Text: query})
	if err != nil {
		return nil, err
	}

	model := SearchPageModel{
		Query:      query,
		Bonds:      response.Bonds,
		TotalCount: response.TotalCount,
	}
	return &model, nil
}
