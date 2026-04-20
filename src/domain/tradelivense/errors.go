package tradelivense

import "errors"

// Domain error sentinels allow callers to distinguish error types with errors.Is
// without importing infrastructure-level error packages into the domain.
var (
	// ErrInvalidStatusTransition is returned when a business action is attempted
	// on an application that is not in the correct state for that action.
	// For example, trying to Accept an application that is already Approved.
	ErrInvalidStatusTransition = errors.New("invalid status transition for current state")

	// ErrDocumentRequired is returned when a customer attempts to submit an
	// application without having attached at least one supporting document.
	ErrDocumentRequired = errors.New("at least one document must be attached before submitting")

	// ErrPaymentRequired is returned when a customer attempts to submit an
	// application before settling the required fee.
	ErrPaymentRequired = errors.New("payment must be settled before submitting")

	// ErrApplicationNotFound is returned by the repository when no application
	// matches the requested ID. The application layer maps this to an HTTP 404.
	ErrApplicationNotFound = errors.New("application not found")

	// ErrForbidden is returned when the caller does not own the application
	// they are trying to modify.
	ErrForbidden = errors.New("you do not own this application")
)
