package models

import (
	"time"

	"github.com/google/uuid"
)

// Document represents a single supporting file attached to an application.
// The URL is a reference to object storage — the domain does not manage bytes.
type Document struct {
	ID          string
	Name        string
	URL         string
	ContentType string
	UploadedAt  time.Time
}

func NewDocument(name, url, contentType string) Document {
	return Document{
		ID:          uuid.New().String(),
		Name:        name,
		URL:         url,
		ContentType: contentType,
		UploadedAt:  time.Now(),
	}
}
