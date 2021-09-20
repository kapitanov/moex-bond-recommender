package data

import (
	"fmt"

	"gorm.io/gorm"
)

// Report содержит данные отчета по облигации
type Report struct {
	Bond                 Bond       `gorm:"embedded;embeddedPrefix:bond_"`
	Issuer               Issuer     `gorm:"embedded;embeddedPrefix:issuer_"`
	MarketData           MarketData `gorm:"embedded;embeddedPrefix:marketdata_"`
	DaysTillMaturity     int        `gorm:"column:days_till_maturity"`
	Currency             string     `gorm:"column:currency"`
	OpenPrice            float64    `gorm:"column:open_price"`
	OpenAccruedInterest  float64    `gorm:"column:open_accrued_interest"`
	OpenFaceValue        float64    `gorm:"column:open_face_value"`
	OpenFee              float64    `gorm:"column:open_fee"`
	OpenValue            float64    `gorm:"column:open_value"`
	CouponPayments       float64    `gorm:"column:coupon_payments"`
	AmortizationPayments float64    `gorm:"column:amortization_payments"`
	MaturityPayment      float64    `gorm:"column:maturity_payments"`
	Taxes                float64    `gorm:"column:taxes"`
	Revenue              float64    `gorm:"column:revenue"`
	ProfitLoss           float64    `gorm:"column:profit_loss"`
	RelativeProfitLoss   float64    `gorm:"column:relative_profit_loss"`
	InterestRate         float64    `gorm:"column:interest_rate"`
}

// TableName задает название таблицы
func (Report) TableName() string {
	return "reports"
}

// ReportRepository отвечает за управление записями в таблице отчетов по облигациям
type ReportRepository interface {
	// Get возвращает отчет по облигации
	// Если данные по указанной облигации не найдены, то возвращается ErrNotFound
	Get(id int) (*Report, error)

	// List возвращает отчеты по облигациям, которые удовлетворяют указанному подзапросу
	List(filter string, values ...interface{}) ([]*Report, error)

	// Rebuild выполняет перерасчет отчетов по облигациям
	Rebuild() error
}

type reportRepository struct {
	db *gorm.DB
}

// Get возвращает отчет по облигации
// Если данные по указанной облигации не найдены, то возвращается ErrNotFound
func (repo *reportRepository) Get(id int) (*Report, error) {
	sqlQuery := `
SELECT *
FROM reports
WHERE bond_id = ?
LIMIT 1
`
	var report Report
	err := repo.db.Raw(sqlQuery, id).First(&report).Error
	if err != nil {
		return nil, err
	}

	return &report, nil
}

// List возвращает отчеты по облигациям, которые удовлетворяют указанному подзапросу
func (repo *reportRepository) List(filter string, values ...interface{}) ([]*Report, error) {
	sqlQuery := `
WITH cte AS (
%s
)
SELECT reports.*
FROM reports
INNER JOIN cte ON cte.bond_id = reports.bond_id
ORDER BY cte.index ASC
LIMIT 10;
`

	sqlQuery = fmt.Sprintf(sqlQuery, filter)

	var reports []*Report
	err := repo.db.Raw(sqlQuery, values...).Scan(&reports).Error
	if err != nil {
		return nil, err
	}

	return reports, nil
}

// Rebuild выполняет перерасчет текущих выплат для всех облигаций
func (repo *reportRepository) Rebuild() error {
	return repo.db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY reports").Error
}
