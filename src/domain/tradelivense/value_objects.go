// Package tradelivense contains the Trade License bounded context.
//
// This is the heart of the system — the domain layer. It has zero dependencies
// on any framework, database, or HTTP library. All business rules and invariants
// live here and are enforced by the types and methods in this package.
package tradelivense

import (
	"errors"

	"github.com/google/uuid"
)

// ApplicationID is a strongly-typed value object wrapping a UUID string.
//
// Using a dedicated type instead of a plain string prevents callers from
// accidentally passing any random string as an ID — the compiler enforces
// the distinction between an ApplicationID and, say, an ApplicantID.
type ApplicationID struct {
	value string
}

// NewApplicationID generates a new random UUID-based application identifier.
func NewApplicationID() ApplicationID {
	return ApplicationID{value: uuid.New().String()}
}

// ApplicationIDFrom reconstructs an ApplicationID from a string (e.g. from a URL param).
// Returns an error if the string is not a valid UUID, preventing corrupt identifiers
// from reaching the domain.
func ApplicationIDFrom(value string) (ApplicationID, error) {
	if _, err := uuid.Parse(value); err != nil {
		return ApplicationID{}, errors.New("invalid application ID format")
	}
	return ApplicationID{value: value}, nil
}

// String returns the raw UUID string, used when persisting or serialising the ID.
func (id ApplicationID) String() string { return id.value }

// ─── LicenseType ─────────────────────────────────────────────────────────────

// LicenseType is a constrained value object representing the category of license
// being applied for. Only values registered in validLicenseTypes are accepted,
// making it impossible to create an application with an unknown license type.
type LicenseType struct {
	value string
}

// TradeLicense is the only license type supported in this bounded context.
const TradeLicense = "TRADE_LICENSE"

// validLicenseTypes acts as an allow-list. Add new types here as the system grows.
var validLicenseTypes = map[string]bool{
	TradeLicense: true,
}

// NewLicenseType validates and constructs a LicenseType.
// Returns an error for any value not in the allow-list.
func NewLicenseType(value string) (LicenseType, error) {
	if !validLicenseTypes[value] {
		return LicenseType{}, errors.New("invalid license type: " + value)
	}
	return LicenseType{value: value}, nil
}

// String returns the raw string value for persistence or serialisation.
func (lt LicenseType) String() string { return lt.value }

// ─── ApplicationStatus ───────────────────────────────────────────────────────

// ApplicationStatus is the current position of an application in the workflow.
//
// Valid transitions are enforced exclusively by the TradeLicenseApplication
// aggregate — nowhere else. This prevents rogue code from jumping the workflow
// by writing directly to the status field.
//
// State machine:
//
//	PENDING ──Submit──► SUBMITTED ──Accept──► ACCEPTED ──Approve──► APPROVED
//	PENDING ──Cancel──► CANCELLED
//	SUBMITTED ──Reject──► REJECTED
//	SUBMITTED ──Adjust──► ADJUSTED
//	ACCEPTED  ──Reject──► REJECTED
//	ACCEPTED  ──Rereview──► REREVIEW ──(back to Reviewer)──► ACCEPTED / REJECTED / ADJUSTED
type ApplicationStatus string

const (
	StatusPending   ApplicationStatus = "PENDING"
	StatusSubmitted ApplicationStatus = "SUBMITTED"
	StatusCancelled ApplicationStatus = "CANCELLED"
	StatusAccepted  ApplicationStatus = "ACCEPTED"
	StatusRejected  ApplicationStatus = "REJECTED"
	StatusAdjusted  ApplicationStatus = "ADJUSTED"
	StatusApproved  ApplicationStatus = "APPROVED"
	StatusRereview  ApplicationStatus = "REREVIEW"
)
