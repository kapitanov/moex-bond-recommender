package recommender

import (
	"context"
	"fmt"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
)

type internalCollection struct {
	id        string
	name      string
	filterSQL func(duration Duration) string
}

var collections = make(map[string]*internalCollection)

func register(id, name string, filterSQL func(duration Duration) string) {
	if _, exists := collections[id]; exists {
		panic(fmt.Sprintf("collection \"%s\" already exists", id))
	}

	coll := &internalCollection{
		id:        id,
		name:      name,
		filterSQL: filterSQL,
	}
	collections[id] = coll
}

// ID возвращает ID коллекции
func (c *internalCollection) ID() string {
	return c.id
}

// Name возвращает название коллекции
func (c *internalCollection) Name() string {
	return c.name
}

// ListBonds возвращает список облигаций из коллекции
func (c *internalCollection) ListBonds(ctx context.Context, tx *data.TX, duration Duration) ([]*Report, error) {
	filterSQL := `
SELECT bond_id, index
FROM collection_bonds
WHERE collection_id = ? and duration = ?
ORDER BY index ASC
`

	entities, err := tx.Reports.List(filterSQL, c.id, getAge(duration))
	if err != nil {
		return nil, err
	}

	reports := make([]*Report, len(entities))
	for i, entity := range entities {
		reports[i] = mapReport(entity)
	}
	return reports, err
}

// Rebuild выполняет обновление данных коллекции
func (c *internalCollection) Rebuild(ctx context.Context, tx *data.TX) error {
	for _, duration := range Durations {
		err := tx.CollectionBondReferences.Rebuild(c.id, getAge(duration), c.filterSQL(duration))
		if err != nil {
			return err
		}
	}

	return nil
}

func getAge(duration Duration) int {
	switch duration {
	case Duration1Year:
		return 1
	case Duration2Year:
		return 2
	case Duration3Year:
		return 3
	case Duration4Year:
		return 4
	case Duration5Year:
		return 5
	}

	panic(fmt.Errorf("\"%s\" is out of valid range for Duration", duration))
}
