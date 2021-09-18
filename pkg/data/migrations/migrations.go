package migrations

import (
	"sort"
	"strings"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var list = make([]*gormigrate.Migration, 0)

// Up применяет миграции к БД
func Up(db *gorm.DB) error {
	sort.Slice(list, func(i, j int) bool {
		return strings.Compare(list[i].ID, list[j].ID) < 0
	})

	options := &gormigrate.Options{
		TableName:                 "migrations",
		IDColumnName:              "id",
		IDColumnSize:              255,
		UseTransaction:            true,
		ValidateUnknownMigrations: true,
	}
	migrator := gormigrate.New(db, options, list)
	err := migrator.Migrate()
	if err != nil {
		return err
	}

	return nil
}

func register(id string, migrate gormigrate.MigrateFunc, rollback gormigrate.RollbackFunc) {
	m := &gormigrate.Migration{
		ID:       id,
		Migrate:  migrate,
		Rollback: rollback,
	}

	list = append(list, m)
}

func registerSQL(id, migrateSQL, rollbackSQL string) {
	migrate := func(db *gorm.DB) error {
		err := db.Exec(migrateSQL).Error
		return err
	}

	rollback := func(db *gorm.DB) error {
		err := db.Exec(rollbackSQL).Error
		return err
	}

	register(id, migrate, rollback)
}
