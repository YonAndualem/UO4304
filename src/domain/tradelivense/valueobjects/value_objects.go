// Package valueobjects defines the immutable value types for the Trade License bounded context.
// Value objects are defined entirely by their value — two instances with the same
// data are considered equal. They validate themselves at construction time.
package valueobjects

import (
	"errors"

	"github.com/google/uuid"
)

// ApplicationID is a strongly-typed value object wrapping a UUID string.
// Using a dedicated type prevents callers from accidentally passing any string
// where an application identifier is expected.
type ApplicationID struct {
	value string
}

// NewApplicationID generates a new random UUID-based application identifier.
func NewApplicationID() ApplicationID {
	return ApplicationID{value: uuid.New().String()}
}

// ApplicationIDFrom reconstructs an ApplicationID from a string (e.g. from a URL param).
// Returns an error if the string is not a valid UUID.
func ApplicationIDFrom(value string) (ApplicationID, error) {
	if _, err := uuid.Parse(value); err != nil {
		return ApplicationID{}, errors.New("invalid application ID format")
	}
	return ApplicationID{value: value}, nil
}

// String returns the raw UUID string for persistence or serialisation.
func (id ApplicationID) String() string { return id.value }

// ─── LicenseType ─────────────────────────────────────────────────────────────

// LicenseType is a constrained value object representing the category of license
// being applied for. Only values in the allow-list are accepted.
type LicenseType struct {
	value string
}

// TradeLicense is the only license type supported in this bounded context.
const TradeLicense = "TRADE_LICENSE"

var validLicenseTypes = map[string]bool{
	TradeLicense: true,
}

// NewLicenseType validates and constructs a LicenseType.
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
// Valid transitions are enforced exclusively by the TradeLicenseApplication
// aggregate — nowhere else.
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
