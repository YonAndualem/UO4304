package models

import (
	"time"

	"github.com/enterprise/trade-license/src/domain/common"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

// TradeLicenseApplication is the aggregate root for the Trade License bounded context.
// All state mutations go through this struct's methods so that business invariants
// are enforced in one place.
type TradeLicenseApplication struct {
	common.AggregateRoot

	ID          valueobjects.ApplicationID
	LicenseType valueobjects.LicenseType
	ApplicantID string
	Commodity   *Commodity
	Documents   []Document
	Payment     *Payment
	Status      valueobjects.ApplicationStatus
	Notes       string
	History     []HistoryEntry
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewTradeLicenseApplication creates a new application in the PENDING state.
func NewTradeLicenseApplication(applicantID string, licenseType valueobjects.LicenseType) *TradeLicenseApplication {
	return &TradeLicenseApplication{
		ID:          valueobjects.NewApplicationID(),
		LicenseType: licenseType,
		ApplicantID: applicantID,
		Status:      valueobjects.StatusPending,
		Documents:   []Document{},
		History:     []HistoryEntry{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (a *TradeLicenseApplication) addHistory(actorID, action string, from, to valueobjects.ApplicationStatus, notes string) {
	a.History = append(a.History, NewHistoryEntry(actorID, action, from, to, notes))
}
