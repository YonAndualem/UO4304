// Package models contains GORM-annotated persistence structs.
//
// These structs exist solely to satisfy GORM's ORM conventions (table names,
// foreign keys, soft deletes). They are kept strictly inside the infrastructure
// layer — the domain layer never sees them. The mapper in the parent package
// translates between these structs and domain objects in both directions.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Application is the GORM model for the applications table.
// It mirrors the TradeLicenseApplication aggregate but is shaped for relational storage.
type Application struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	LicenseType string `gorm:"not null"`
	ApplicantID string `gorm:"not null;index"`
	Status      string `gorm:"not null;index"`
	Notes       string

	// HasOne / HasMany associations — GORM will preload these on FindByID queries.
	Commodity *Commodity           `gorm:"foreignKey:ApplicationID;constraint:OnDelete:CASCADE"`
	Documents []Document           `gorm:"foreignKey:ApplicationID;constraint:OnDelete:CASCADE"`
	Payment   *Payment             `gorm:"foreignKey:ApplicationID;constraint:OnDelete:CASCADE"`
	History   []ApplicationHistory `gorm:"foreignKey:ApplicationID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft-delete support
}

// ApplicationHistory records every status transition for an application.
// It is append-only — rows are never updated or deleted.
type ApplicationHistory struct {
	ID            string    `gorm:"primaryKey;type:uuid"`
	ApplicationID string    `gorm:"not null;type:uuid;index"`
	ActorID       string    `gorm:"not null"`
	Action        string    `gorm:"not null"`
	FromStatus    string    `gorm:"not null"`
	ToStatus      string    `gorm:"not null"`
	Notes         string
	OccurredAt    time.Time `gorm:"not null;index"`
}

// Commodity is the GORM model for the commodities table.
type Commodity struct {
	ID            string `gorm:"primaryKey;type:uuid"`
	ApplicationID string `gorm:"not null;type:uuid;index"`
	Name          string `gorm:"not null"`
	Description   string
	Category      string
}

// Document is the GORM model for the documents table.
// An application can have many documents (one-to-many relationship).
type Document struct {
	ID            string `gorm:"primaryKey;type:uuid"`
	ApplicationID string `gorm:"not null;type:uuid;index"`
	Name          string `gorm:"not null"`
	URL           string `gorm:"not null"`
	ContentType   string
	UploadedAt    time.Time
}

// Payment is the GORM model for the payments table.
// An application has at most one payment (one-to-one, enforced by uniqueIndex).
type Payment struct {
	ID            string  `gorm:"primaryKey;type:uuid"`
	ApplicationID string  `gorm:"not null;type:uuid;uniqueIndex"`
	Amount        float64
	Currency      string
	TransactionID string `gorm:"uniqueIndex"` // Prevents duplicate payment records
	PaidAt        time.Time
	Status        string
}
