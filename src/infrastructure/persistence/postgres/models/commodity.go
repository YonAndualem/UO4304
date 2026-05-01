package models

// Commodity is the GORM model for the commodities table.
type Commodity struct {
	ID            string `gorm:"primaryKey;type:uuid"`
	ApplicationID string `gorm:"not null;type:uuid;index"`
	Name          string `gorm:"not null"`
	Description   string
	Category      string
}
