package models

import "time"

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
