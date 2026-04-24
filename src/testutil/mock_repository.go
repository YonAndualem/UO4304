// Package testutil provides shared test helpers used across all layers.
// Nothing in this package is compiled into the production binary.
package testutil

import (
	"context"
	"sync"

	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/models"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

// MockRepository is an in-memory implementation of repositories.ApplicationRepository.
// It is used in unit tests to avoid a real database dependency, keeping tests fast
// and self-contained. The mutex makes it safe for concurrent test use.
type MockRepository struct {
	mu   sync.RWMutex
	data map[string]*models.TradeLicenseApplication
}

func NewMockRepository() *MockRepository {
	return &MockRepository{data: make(map[string]*models.TradeLicenseApplication)}
}

func (r *MockRepository) Save(_ context.Context, app *models.TradeLicenseApplication) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[app.ID.String()] = app
	return nil
}

func (r *MockRepository) Update(_ context.Context, app *models.TradeLicenseApplication) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[app.ID.String()] = app
	return nil
}

func (r *MockRepository) FindByID(_ context.Context, id valueobjects.ApplicationID) (*models.TradeLicenseApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	app, ok := r.data[id.String()]
	if !ok {
		return nil, domainerrors.ErrApplicationNotFound
	}
	return app, nil
}

func (r *MockRepository) FindByApplicantID(_ context.Context, applicantID string) ([]*models.TradeLicenseApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.TradeLicenseApplication
	for _, app := range r.data {
		if app.ApplicantID == applicantID {
			result = append(result, app)
		}
	}
	return result, nil
}

func (r *MockRepository) FindByStatus(_ context.Context, status valueobjects.ApplicationStatus) ([]*models.TradeLicenseApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.TradeLicenseApplication
	for _, app := range r.data {
		if app.Status == status {
			result = append(result, app)
		}
	}
	return result, nil
}

func (r *MockRepository) Delete(_ context.Context, id valueobjects.ApplicationID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id.String()]; !ok {
		return domainerrors.ErrApplicationNotFound
	}
	delete(r.data, id.String())
	return nil
}

// ─── Test fixture helpers ─────────────────────────────────────────────────────

// NewValidLicenseType returns a TRADE_LICENSE LicenseType for use in tests.
func NewValidLicenseType() valueobjects.LicenseType {
	lt, _ := valueobjects.NewLicenseType(valueobjects.TradeLicense)
	return lt
}

// NewPendingApplication creates a minimal PENDING application for use in tests.
func NewPendingApplication(applicantID string) *models.TradeLicenseApplication {
	app := models.NewTradeLicenseApplication(applicantID, NewValidLicenseType())
	return app
}

// NewReadyToSubmitApplication creates an application with all pre-conditions
// satisfied so that Submit() will succeed.
func NewReadyToSubmitApplication(applicantID string) *models.TradeLicenseApplication {
	app := NewPendingApplication(applicantID)
	app.SelectCommodity(models.NewCommodity("General Trading", "Import/Export", "Commerce"))
	app.AttachDocument(models.NewDocument("Passport", "https://storage/passport.pdf", "application/pdf"))
	app.SettlePayment(models.NewPayment(500.00, "USD", "TXN-001"))
	return app
}
