// Package repositories defines the persistence port (interface) for the Trade License bounded context.
// The domain declares what storage operations it needs; the infrastructure layer provides
// the concrete implementation. This inversion keeps the domain free of database details.
package repositories

import (
	"context"

	"github.com/enterprise/trade-license/src/domain/tradelivense/entities"
	"github.com/enterprise/trade-license/src/domain/tradelivense/valueobjects"
)

// ApplicationRepository is the domain port for persisting and querying applications.
type ApplicationRepository interface {
	// Save persists a brand-new application for the first time.
	Save(ctx context.Context, app *entities.TradeLicenseApplication) error

	// Update persists changes to an existing application (status transitions,
	// new documents, payment updates, etc.).
	Update(ctx context.Context, app *entities.TradeLicenseApplication) error

	// FindByID retrieves a single application by its unique identifier.
	// Returns ErrApplicationNotFound when no record matches.
	FindByID(ctx context.Context, id valueobjects.ApplicationID) (*entities.TradeLicenseApplication, error)

	// FindByApplicantID returns all applications belonging to a specific customer.
	FindByApplicantID(ctx context.Context, applicantID string) ([]*entities.TradeLicenseApplication, error)

	// FindByStatus returns all applications currently in the given workflow status.
	FindByStatus(ctx context.Context, status valueobjects.ApplicationStatus) ([]*entities.TradeLicenseApplication, error)

	// Delete soft-deletes an application. The aggregate's Delete() method guards
	// that only PENDING, CANCELLED, or REJECTED applications may be deleted.
	Delete(ctx context.Context, id valueobjects.ApplicationID) error
}
