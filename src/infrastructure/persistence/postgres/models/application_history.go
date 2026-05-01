package models

import "time"

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
