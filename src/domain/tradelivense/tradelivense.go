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
	"github.com/enterprise/trade-license/src/domain/tradelivense/models"
	domainerrors "github.com/enterprise/trade-license/src/domain/tradelivense/errors"
	"github.com/enterprise/trade-license/src/domain/tradelivense/repositories"
	"github.com/enterprise/trade-license/src/domain/tradelivense/valueobjects"
)

// ─── Entity type aliases ─────────────────────────────────────────────────────

type TradeLicenseApplication = models.TradeLicenseApplication
type Commodity = models.Commodity
type Document = models.Document
type Payment = models.Payment
type PaymentStatus = models.PaymentStatus
type HistoryEntry = models.HistoryEntry

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

	PaymentSettled = models.PaymentSettled
	PaymentPending = models.PaymentPending
	PaymentFailed  = models.PaymentFailed
)

// ─── Constructor re-exports ──────────────────────────────────────────────────

func NewTradeLicenseApplication(applicantID string, licenseType LicenseType) *TradeLicenseApplication {
	return models.NewTradeLicenseApplication(applicantID, licenseType)
}

func NewCommodity(name, description, category string) Commodity {
	return models.NewCommodity(name, description, category)
}

func NewDocument(name, url, contentType string) Document {
	return models.NewDocument(name, url, contentType)
}

func NewPayment(amount float64, currency, transactionID string) Payment {
	return models.NewPayment(amount, currency, transactionID)
}

func NewHistoryEntry(actorID, action string, from, to ApplicationStatus, notes string) HistoryEntry {
	return models.NewHistoryEntry(actorID, action, from, to, notes)
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
