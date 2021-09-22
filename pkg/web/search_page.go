package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/search"
)

// SearchPage обрабатывает запросы "GET /search"
func (ctrl *pagesController) SearchPage(c *gin.Context) {
	query, exists := c.GetQuery("q")
	if !exists {
		c.Redirect(http.StatusFound, "/")
		return
	}

	query = strings.TrimSpace(query)
	if query == "" {
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

// SearchPageModel - модель для страницы "pages/search.html"
type SearchPageModel struct {
	Query      string
	Bonds      []*data.Bond
	TotalCount int
}

// NewSearchPageModel создает объекты типа SearchPageModel
func NewSearchPageModel(app app.App, query string) (*SearchPageModel, error) {
	u, err := app.NewUnitOfWork(context.Background())
	if err != nil {
		return nil, err
	}
	defer u.Close()

	response, err := u.Search(search.Request{Text: query})
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
