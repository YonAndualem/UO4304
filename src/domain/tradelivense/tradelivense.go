// Package tradelivense is the Trade License bounded context.
//
// This file re-exports all types from the domain subpackages so that callers
// (application, infrastructure, testutil) continue to import a single path:
//
//	"github.com/enterprise/trade-license/src/domain/tradelivense"
//
// Internal organisation:
//
//	entities/       — TradeLicenseApplication aggregate root + domain entities
//	errors/         — domain error sentinels
//	events/         — domain events raised by the aggregate
//	repositories/   — ApplicationRepository port (interface)
//	valueobjects/   — ApplicationID, LicenseType, ApplicationStatus
package tradelivense

import (
	"github.com/enterprise/trade-license/src/domain/tradelivense/entities"
	domainerrors "github.com/enterprise/trade-license/src/domain/tradelivense/errors"
	"github.com/enterprise/trade-license/src/domain/tradelivense/repositories"
	"github.com/enterprise/trade-license/src/domain/tradelivense/valueobjects"
)

// ─── Entity type aliases ─────────────────────────────────────────────────────

type TradeLicenseApplication = entities.TradeLicenseApplication
type Commodity = entities.Commodity
type Document = entities.Document
type Payment = entities.Payment
type PaymentStatus = entities.PaymentStatus
type HistoryEntry = entities.HistoryEntry

// ─── Value object type aliases ───────────────────────────────────────────────

type ApplicationID = valueobjects.ApplicationID
type LicenseType = valueobjects.LicenseType
type ApplicationStatus = valueobjects.ApplicationStatus

// ─── Repository type alias ───────────────────────────────────────────────────

type ApplicationRepository = repositories.ApplicationRepository

// ─── Error re-exports ────────────────────────────────────────────────────────
// These point to the same sentinel values so errors.Is() keeps working correctly.

var (
	ErrInvalidStatusTransition = domainerrors.ErrInvalidStatusTransition
	ErrDocumentRequired        = domainerrors.ErrDocumentRequired
	ErrPaymentRequired         = domainerrors.ErrPaymentRequired
	ErrApplicationNotFound     = domainerrors.ErrApplicationNotFound
	ErrForbidden               = domainerrors.ErrForbidden
)

// ─── Status constant re-exports ──────────────────────────────────────────────

const (
	StatusPending   = valueobjects.StatusPending
	StatusSubmitted = valueobjects.StatusSubmitted
	StatusCancelled = valueobjects.StatusCancelled
	StatusAccepted  = valueobjects.StatusAccepted
	StatusRejected  = valueobjects.StatusRejected
	StatusAdjusted  = valueobjects.StatusAdjusted
	StatusApproved  = valueobjects.StatusApproved
	StatusRereview  = valueobjects.StatusRereview

	TradeLicense = valueobjects.TradeLicense

	PaymentSettled = entities.PaymentSettled
	PaymentPending = entities.PaymentPending
	PaymentFailed  = entities.PaymentFailed
)

// ─── Constructor re-exports ──────────────────────────────────────────────────

func NewTradeLicenseApplication(applicantID string, licenseType LicenseType) *TradeLicenseApplication {
	return entities.NewTradeLicenseApplication(applicantID, licenseType)
}

func NewCommodity(name, description, category string) Commodity {
	return entities.NewCommodity(name, description, category)
}

func NewDocument(name, url, contentType string) Document {
	return entities.NewDocument(name, url, contentType)
}

func NewPayment(amount float64, currency, transactionID string) Payment {
	return entities.NewPayment(amount, currency, transactionID)
}

func NewHistoryEntry(actorID, action string, from, to ApplicationStatus, notes string) HistoryEntry {
	return entities.NewHistoryEntry(actorID, action, from, to, notes)
}

func NewApplicationID() ApplicationID {
	return valueobjects.NewApplicationID()
}

func ApplicationIDFrom(value string) (ApplicationID, error) {
	return valueobjects.ApplicationIDFrom(value)
}

func NewLicenseType(value string) (LicenseType, error) {
	return valueobjects.NewLicenseType(value)
}
