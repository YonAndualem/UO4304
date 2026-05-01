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
