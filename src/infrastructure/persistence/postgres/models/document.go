package models

import "time"

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
