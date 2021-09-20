package data

import (
	"fmt"

	"gorm.io/gorm"
)

// CollectionBondRef представляет запись в таблице связей "коллекция-облигация"
type CollectionBondRef struct {
	CollectionID string `gorm:"column:collection_id"`
	Duration     int    `gorm:"column:duration"`
	BondID       int    `gorm:"column:bond_id"`
	Order        int    `gorm:"column:order"`
}

// TableName задает название таблицы
func (CollectionBondRef) TableName() string {
	return "collection_bonds"
}

// CollectionBondRefRepository отвечает за управление записями в таблице связей "коллекция-облигация"
type CollectionBondRefRepository interface {
	// Rebuild выполняет перерасчет списка облигаций для отдельной коллекции
	Rebuild(collectionID string, duration int, filter string) error
}

type collectionBondRefRepository struct {
	db *gorm.DB
}

// Rebuild выполняет перерасчет списка облигаций для отдельной коллекции
func (repo *collectionBondRefRepository) Rebuild(collectionID string, duration int, filter string) error {
	err := repo.db.
		Exec("DELETE FROM collection_bonds WHERE collection_id = ? AND duration = ?", collectionID, duration).
		Error
	if err != nil {
		return err
	}

	sqlQuery := `
INSERT INTO collection_bonds(collection_id, duration, bond_id, index)
SELECT ?,
       ?,
       bond_id,
       ROW_NUMBER() OVER (ORDER BY interest_rate DESC) AS index
FROM reports
INNER JOIN bonds on reports.bond_id = bonds.id
WHERE bond_id IN (
%s
)
AND interest_rate > 0
AND (AGE(bonds.maturity_date::date, NOW()::date) >= '3 day'::interval)
AND (AGE(bonds.maturity_date::date, NOW()::date) <= '%d year'::interval)
ORDER BY interest_rate DESC;
`
	sqlQuery = fmt.Sprintf(sqlQuery, filter, duration)

	err = repo.db.Exec(sqlQuery, collectionID, duration).Error
	if err != nil {
		return err
	}

	return nil
}
