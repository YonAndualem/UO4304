package models

import "time"

// User is the GORM model for the users table.
// Stores registered accounts with bcrypt-hashed passwords.
type User struct {
	ID           string    `gorm:"primaryKey;type:uuid"`
	UserID       string    `gorm:"not null;uniqueIndex"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"not null"`
	CreatedAt    time.Time
}
