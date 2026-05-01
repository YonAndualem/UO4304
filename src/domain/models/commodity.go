package models

import "github.com/google/uuid"

// Commodity represents the specific trade activity or product the applicant
// intends to operate under the license. Owned by TradeLicenseApplication.
type Commodity struct {
	ID          string
	Name        string
	Description string
	Category    string
}

func NewCommodity(name, description, category string) Commodity {
	return Commodity{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Category:    category,
	}
}
