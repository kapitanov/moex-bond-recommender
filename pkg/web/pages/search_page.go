package pages

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
func (ctrl *Controller) SearchPage(c *gin.Context) {
	var query SearchQueryModel
	err := c.BindQuery(&query)
	if err != nil {
		panic(NewError(400, "malformed query"))
	}

	text := query.Text
	text = strings.TrimSpace(text)
	if text == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	model, err := NewSearchPageModel(ctrl.app, c, text, query.Skip)
	if err != nil {
		panic(err)
	}

	if len(model.Bonds) == 1 {
		c.Redirect(http.StatusFound, fmt.Sprintf("/bonds/%s", url.PathEscape(model.Bonds[0].ISIN)))
		return
	}

	if query.Partial {
		ctrl.renderHTML(c, http.StatusOK, "pages/search_partial.html", model)
	} else {
		ctrl.renderHTML(c, http.StatusOK, "pages/search", model)
	}
}

// SearchQueryModel - модель для запроса "GET /search"
type SearchQueryModel struct {
	Text    string `form:"q"`
	Skip    int    `form:"skip"`
	Partial bool   `form:"partial"`
}

// SearchPageModel - модель для страницы "pages/search.html"
type SearchPageModel struct {
	Query          string
	Bonds          []*data.Bond
	Skip           int
	TotalCount     int
	DisplayedCount int
}

// NewSearchPageModel создает объекты типа SearchPageModel
func NewSearchPageModel(app app.App, context context.Context, query string, skip int) (*SearchPageModel, error) {
	u, err := app.NewUnitOfWork(context)
	if err != nil {
		return nil, err
	}
	defer u.Close()

	response, err := u.Search(search.Request{Text: query, Skip: skip, Limit: search.DefaultLimit})
	if err != nil {
		return nil, err
	}

	model := SearchPageModel{
		Query:          query,
		Bonds:          response.Bonds,
		Skip:           skip,
		TotalCount:     response.TotalCount,
		DisplayedCount: skip + len(response.Bonds),
	}
	return &model, nil
}
