package tradelivense

import "context"

// ApplicationRepository is the domain's port (interface) for persistence.
//
// In Clean Architecture this interface lives in the domain layer while its
// implementation lives in the infrastructure layer. The domain depends on the
// abstraction, not on PostgreSQL, GORM, or any other storage technology.
// This is the Dependency Inversion Principle in practice — high-level policy
// (domain rules) does not depend on low-level detail (database driver).
//
// The infrastructure layer satisfies this contract, and the application layer
// wires them together at startup.
type ApplicationRepository interface {
	// Save persists a brand-new application for the first time.
	Save(ctx context.Context, app *TradeLicenseApplication) error

	// Update persists changes to an existing application (status transitions,
	// new documents, payment updates, etc.).
	Update(ctx context.Context, app *TradeLicenseApplication) error

	// FindByID retrieves a single application by its unique identifier.
	// Returns ErrApplicationNotFound when no record matches.
	FindByID(ctx context.Context, id ApplicationID) (*TradeLicenseApplication, error)

	// FindByApplicantID returns all applications belonging to a specific customer.
	FindByApplicantID(ctx context.Context, applicantID string) ([]*TradeLicenseApplication, error)

	// FindByStatus returns all applications currently in the given workflow status.
	// Used by reviewers and approvers to list their work queues.
	FindByStatus(ctx context.Context, status ApplicationStatus) ([]*TradeLicenseApplication, error)

	// Delete soft-deletes an application. Only valid for PENDING, CANCELLED, or
	// REJECTED applications — the aggregate's Delete() method enforces this guard.
	Delete(ctx context.Context, id ApplicationID) error
}
