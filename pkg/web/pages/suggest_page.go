package pages

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kapitanov/moex-bond-recommender/pkg/app"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

// SuggestPage обрабатывает запросы "GET /suggest"
func (ctrl *Controller) SuggestPage(c *gin.Context) {
	u, err := ctrl.app.NewUnitOfWork(c)
	if err != nil {
		panic(err)
	}
	defer u.Close()

	req, err := NewSuggestPortfolioRequest(c, u)
	if err != nil {
		panic(err)
	}

	collections := u.ListCollections()
	model := SuggestPageModel{
		Collections: make([]SuggestPageCollectionModel, len(collections)),
	}
	for i, c := range collections {
		model.Collections[i] = SuggestPageCollectionModel{
			ID:   c.ID(),
			Name: c.Name(),
		}
	}
	if req == nil {
		ctrl.renderHTML(c, http.StatusOK, "pages/suggest", &model)
		return
	}

	portfolio, err := u.Suggest(req.toSuggestRequest())
	if err != nil {
		panic(err)
	}

	extModel := &SuggestViewPageModel{
		SuggestPageModel: model,
		Request:          req,
		Portfolio:        portfolio,
		ShareUrl:         fmt.Sprintf("/suggests?json=%s", req),
	}
	cashFlowDict := make(map[time.Time]*SuggestViewCashFlowPageModel)
	for _, p := range portfolio.Positions {
		for _, c := range p.CashFlow {
			item, exists := cashFlowDict[c.Date]
			if !exists {
				item = &SuggestViewCashFlowPageModel{Date: c.Date}
				cashFlowDict[c.Date] = item
			}
			item.Amount += c.ValueRub
			switch c.Type {
			case recommender.Coupon:
				item.HasCoupon = true
				break
			case recommender.Amortization:
				item.HasAmortization = true
				break
			case recommender.Maturity:
				item.HasMaturity = true
				break
			}
		}
	}
	extModel.CashFlow = make([]*SuggestViewCashFlowPageModel, len(cashFlowDict))
	i := 0
	for _, item := range cashFlowDict {
		extModel.CashFlow[i] = item
		i++
	}
	sort.Slice(extModel.CashFlow, func(i, j int) bool {
		return extModel.CashFlow[i].Date.Before(extModel.CashFlow[j].Date)
	})

	ctrl.renderHTML(c, http.StatusOK, "pages/suggest_view", extModel)
}

// SuggestPortfolioRequest - параметры для запроса GET /api/suggest-portfolio
type SuggestPortfolioRequest struct {
	Amount         float64                        `json:"amount"`
	MaxDuration    recommender.Duration           `json:"-"`
	MaxDurationRaw int                            `json:"max_duration"`
	Parts          []*SuggestPortfolioRequestPart `json:"parts"`
}

// SuggestPortfolioRequestPart - элемент параметра запроса GET /api/suggest-portfolio
type SuggestPortfolioRequestPart struct {
	Collection     recommender.Collection `json:"-"`
	CollectionName string                 `json:"-"`
	CollectionID   string                 `json:"collection"`
	Weight         float64                `json:"weight"`
}

// NewSuggestPortfolioRequest создает объект SuggestPortfolioRequest из строки
func NewSuggestPortfolioRequest(c *gin.Context, u app.UnitOfWork) (*SuggestPortfolioRequest, error) {
	raw, exists := c.GetQuery("json")
	if !exists || raw == "" {
		return nil, nil
	}

	var request SuggestPortfolioRequest
	err := json.Unmarshal([]byte(raw), &request)
	if err != nil {
		return nil, NewError(400, "malformed \"json\" parameter")
	}

	if request.Amount <= 0 {
		return nil, NewError(400, "invalid value for \"amount\" parameter")
	}

	switch request.MaxDurationRaw {
	case 1:
		request.MaxDuration = recommender.Duration1Year
		break
	case 2:
		request.MaxDuration = recommender.Duration2Year
		break
	case 3:
		request.MaxDuration = recommender.Duration3Year
		break
	case 4:
		request.MaxDuration = recommender.Duration4Year
		break
	case 5:
		request.MaxDuration = recommender.Duration5Year
		break
	default:
		return nil, NewError(400, "invalid value for \"max_duration\" parameter")
	}

	if request.Parts != nil && len(request.Parts) > 0 {
		sumOfWeights := 0.0
		for _, part := range request.Parts {
			collection, err := u.GetCollection(part.CollectionID)
			if err != nil {
				if err == recommender.ErrNotFound {
					return nil, NewError(400, "collection \"%s\" doesn't exist", part.Collection)
				}
				return nil, err
			}

			part.Collection = collection
			part.CollectionName = collection.Name()

			if part.Weight < 0 {
				return nil, NewError(400, "invalid value for \"weight\" parameter")
			}

			sumOfWeights += part.Weight
		}

		for _, part := range request.Parts {
			part.Weight = 100.0 * part.Weight / sumOfWeights
		}
	} else {
		request.Parts = nil
	}

	return &request, nil
}

// String преобразует значение в строку
func (r *SuggestPortfolioRequest) String() string {
	bytes, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	str := string(bytes)
	str = url.QueryEscape(str)
	return str
}

// toSuggestRequest создает recommender.SuggestRequest из SuggestPortfolioRequest
func (r *SuggestPortfolioRequest) toSuggestRequest() *recommender.SuggestRequest {
	req := recommender.SuggestRequest{
		Amount:      r.Amount,
		MaxDuration: r.MaxDuration,
		Parts:       nil,
	}

	if r.Parts != nil && len(r.Parts) > 0 {
		req.Parts = make([]*recommender.SuggestRequestPart, len(r.Parts))
		for i, part := range r.Parts {
			req.Parts[i] = &recommender.SuggestRequestPart{
				Collection: part.Collection,
				Weight:     part.Weight,
			}
		}
	}

	return &req
}

// SuggestPageModel - модель для страницы "pages/suggest.html"
type SuggestPageModel struct {
	Collections []SuggestPageCollectionModel
}

// SuggestPageCollectionModel описывает коллекцию на странице "pages/suggest.html"
type SuggestPageCollectionModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SuggestViewPageModel - модель для страницы "pages/suggest_view.html"
type SuggestViewPageModel struct {
	SuggestPageModel
	Request   *SuggestPortfolioRequest
	Portfolio *recommender.SuggestResult
	ShareUrl  string
	CashFlow  []*SuggestViewCashFlowPageModel
}

// SuggestViewCashFlowPageModel - модель выплаты для страницы "pages/suggest_view.html"
type SuggestViewCashFlowPageModel struct {
	Date            time.Time
	Amount          float64
	HasCoupon       bool
	HasAmortization bool
	HasMaturity     bool
}
