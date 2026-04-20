// Package postgres is the infrastructure implementation of the domain repository interface.
//
// This package is the only place in the codebase that knows about PostgreSQL and GORM.
// The domain and application layers reference only the tradelivense.ApplicationRepository
// interface — they have no import path into this package. This is Clean Architecture's
// Dependency Rule: source code dependencies point inward (toward domain), never outward.
package postgres

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/enterprise/trade-license/src/infrastructure/persistence/postgres/models"
)

// NewDB opens a PostgreSQL connection and runs GORM's AutoMigrate to ensure
// the schema is up to date. In production you would replace AutoMigrate with
// versioned SQL migrations (see migrations/001_initial.sql), but AutoMigrate
// is convenient for development and testing.
func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// AutoMigrate creates or alters tables to match the model structs.
	// It never drops columns, so it is safe to run on an existing database.
	if err := db.AutoMigrate(
		&models.Application{},
		&models.Commodity{},
		&models.Document{},
		&models.Payment{},
		&models.ApplicationHistory{},
		&models.User{},
	); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}
