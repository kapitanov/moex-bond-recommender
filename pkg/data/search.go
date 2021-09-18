package data

import (
	"gorm.io/gorm"
)

// SearchIndex содержит данные для поиска облигаций
type SearchIndex struct {
	ID     int    `gorm:"column:id; primaryKey"`
	BondID int    `gorm:"column:bond_id"`
	Vector string `gorm:"column:vector; type:tsvector"`
}

// TableName задает название таблицы
func (SearchIndex) TableName() string {
	return "search"
}

// SearchRepository предоставляет доступ к данным поискового индекса
type SearchRepository interface {
	// Exec выполняет поиск по поисковому индексу
	Exec(filter string, skip, limit int) (bonds []*Bond, totalCount int, err error)

	// Rebuild выполняет пересборку поискового индекса
	Rebuild() error
}

type searchRepository struct {
	db *gorm.DB
}

// Exec выполняет поиск по поисковому индексу
func (repo *searchRepository) Exec(filter string, skip, limit int) ([]*Bond, int, error) {
	listSQL := `
SELECT bonds.*
FROM (
         SELECT bond_id,
                ROW_NUMBER() OVER () AS row
         FROM (
                  SELECT bond_id,
                         MAX(rank)
                  FROM (
                           SELECT bond_id, TS_RANK(vector, TO_TSQUERY(?)) AS rank
                           FROM search
                           WHERE vector @@ TO_TSQUERY(?)
                       ) xs
                  GROUP BY bond_id
              ) xs
     ) xs
         INNER JOIN bonds ON bonds.id = xs.bond_id
WHERE xs.row > ?
LIMIT ?
`
	var items []*Bond
	err := repo.db.Raw(listSQL, filter, filter, skip, limit).Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}

	countSQL := `
SELECT COUNT(*)
FROM (
         SELECT bond_id,
                MAX(rank)
         FROM (
                  SELECT bond_id, TS_RANK(vector, TO_TSQUERY(?)) AS rank
                  FROM search
                  WHERE vector @@ TO_TSQUERY(?)
              ) xs
         GROUP BY bond_id
     ) xs
`
	var totalCount int
	err = repo.db.Raw(countSQL, filter, filter).Scan(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	return items, totalCount, nil
}

// Rebuild выполняет пересборку поискового индекса
func (repo *searchRepository) Rebuild() error {
	return repo.db.Exec("CALL proc_rebuild_search_index()").Error
}
