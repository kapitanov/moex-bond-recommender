package data

import (
	"time"

	"gorm.io/gorm"
)

// CashFlowItem представляет запись в таблице текущих выплат
type CashFlowItem struct {
	BondID   int         `gorm:"column:bond_id"`
	Type     PaymentType `gorm:"column:type"`
	Date     time.Time   `gorm:"column:date"`
	ValueRub float64     `gorm:"column:value_rub"`
}

// TableName задает название таблицы
func (CashFlowItem) TableName() string {
	return "cashflows"
}

// CashFlowRepository отвечает за управление записями в таблице текущих выплат
type CashFlowRepository interface {
	// List возвращает полный список текущих выплат для облигации
	// Направление сортировки - по возрастанию даты
	List(id int) ([]*CashFlowItem, error)

	// Rebuild выполняет перерасчет текущих выплат для всех облигаций
	Rebuild() error
}

type cashFlowRepository struct {
	db *gorm.DB
}

// List возвращает полный список текущих выплат для облигации
// Направление сортировки - по возрастанию даты
func (repo *cashFlowRepository) List(id int) ([]*CashFlowItem, error) {
	sqlQuery := `
SELECT *
FROM cashflows
WHERE bond_id = ?
ORDER BY date ASC, type ASC
`
	var items []*CashFlowItem
	err := repo.db.Raw(sqlQuery, id).Scan(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

// Rebuild выполняет перерасчет текущих выплат для всех облигаций
func (repo *cashFlowRepository) Rebuild() error {
	return repo.db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY cashflows").Error
}
