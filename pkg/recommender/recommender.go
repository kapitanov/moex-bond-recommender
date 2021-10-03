package recommender

import (
	"context"
	"errors"
	"time"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
)

// ErrNotFound возвращается, если запрошенный объект не найден
var ErrNotFound = errors.New("not found")

// Service предоставляет доступ к модулю рекомендаций
type Service interface {
	// ListCollections возвращает список коллекций рекомендаций
	ListCollections() []Collection

	// GetCollection возвращает коллекцию рекомендаций по ее ID
	// Если коллекция не найдена, то возвращается ошибка ErrNotFound
	GetCollection(id string) (Collection, error)

	// GetReport возвращает отчет по отдельной облигации
	// Если отчет не найден, то возвращается ошибка ErrNotFound
	GetReport(ctx context.Context, tx *data.TX, id int) (*Report, error)

	// Suggest выполняет расчет предложений по инвестированию
	Suggest(ctx context.Context, tx *data.TX, request *SuggestRequest) (*SuggestResult, error)

	// Rebuild выполняет обновление данных рекомендаций
	Rebuild(ctx context.Context, tx *data.TX) error
}

// Duration представляет срок до погашения облигации
type Duration string

const (
	// Duration1Year обозначает облигации, до погашения которых осталось не более 1 года
	Duration1Year Duration = "1y"

	// Duration2Year обозначает облигации, до погашения которых осталось не более 2 лет
	Duration2Year Duration = "2y"

	// Duration3Year обозначает облигации, до погашения которых осталось не более 3 лет
	Duration3Year Duration = "3y"

	// Duration4Year обозначает облигации, до погашения которых осталось не более 4 лет
	Duration4Year Duration = "4y"

	// Duration5Year обозначает облигации, до погашения которых осталось не более 5 лет
	Duration5Year Duration = "5y"
)

// Durations содержит список всех возможных значений Duration
var Durations = []Duration{Duration1Year, Duration2Year, Duration3Year, Duration4Year, Duration5Year}

// Collection содержит данные коллекции рекомендаций
type Collection interface {
	// ID возвращает ID коллекции
	ID() string

	// Name возвращает название коллекции
	Name() string

	// ListBonds возвращает список облигаций из коллекции
	ListBonds(ctx context.Context, tx *data.TX, limit int, duration Duration) ([]*Report, error)
}

// Report содержит данные отчета по облигации
type Report struct {
	// Облигация
	Bond *data.Bond

	// Эмитент
	Issuer *data.Issuer

	// Текущие рыночные данные
	MarketData *data.MarketData

	// Дней до погашения
	DaysTillMaturity int

	// Валюта торгов
	Currency string

	// Чистая цена открытия, в %
	OpenPrice float64

	// НКД на момент открытия, в валюте
	OpenAccruedInterest float64

	// Номинал на момент открытия, в валюте
	OpenFaceValue float64

	// Комиссия за сделку покупки, в валюте
	OpenFee float64

	// Сумма затрат на покупку 1 облигации, в валюте
	OpenValue float64

	// Сумма выплат по купонам от момента покупки до момента погашения, в валюте
	CouponPayments float64

	// Сумма выплат по амортизациям от момента покупки до момента погашения, в валюте
	AmortizationPayments float64

	// Сумма выплат по погашению, в валюте
	MaturityPayment float64

	// Сумма налогов, в валюте
	Taxes float64

	// Суммарный доход, в валюте
	Revenue float64

	// Прибыль, в валюте
	ProfitLoss float64

	// Прибыль, в % по отношению к сумме вложений
	RelativeProfitLoss float64

	// Приведенная доходность, % годовых
	InterestRate float64

	// Таблица выплат
	CashFlow []*CashFlowItem
}

// CashFlowItemType кодирует тип выплаты
type CashFlowItemType string

const (
	// Coupon обозначает тип выплаты "выплата по купону"
	Coupon CashFlowItemType = "C"

	// Amortization обозначает тип выплаты "выплата по амортизации"
	Amortization CashFlowItemType = "A"

	// Maturity обозначает тип выплаты "выплата по погашению"
	Maturity CashFlowItemType = "M"
)

// CashFlowItem содержит данные выплат по облигации
type CashFlowItem struct {
	// Тип выплаты
	Type CashFlowItemType

	// Дата выплаты
	Date time.Time

	// Сумма выплаты, в рублях
	ValueRub float64
}

// SuggestRequest - запрос на формирование предложений по инвестированию
type SuggestRequest struct {
	// Сумма для инвестирования
	Amount float64

	// Максимальный срок инвестирования
	MaxDuration Duration

	// Ограничения по составу портфеля
	Parts []*SuggestRequestPart
}

// SuggestRequestPart - ограничения по составу портфеля для запроса SuggestRequest
type SuggestRequestPart struct {
	// Тип коллекции
	Collection Collection

	// Вес в портфеле
	Weight float64
}

// SuggestResult - результат расчета предложений по инвестированию
type SuggestResult struct {
	// Список позиций
	Positions []*SuggestedPortfolioPosition

	// Общая сумма инвестирования
	Amount float64

	// Дней до погашения всего портфеля
	DurationDays int

	// Прибыль, в валюте
	ProfitLoss float64

	// Прибыль, в % по отношению к сумме вложений
	RelativeProfitLoss float64

	// Приведенная доходность, % годовых
	InterestRate float64
}

// SuggestedPortfolioPosition - позиция в предложенном портфеле
type SuggestedPortfolioPosition struct {
	Report

	// Размер позиции
	Quantity int

	// Доля в составе портфеля (0..1)
	Weight float64
}

// New создает новый объект Service
func New() (Service, error) {
	return &service{}, nil
}
