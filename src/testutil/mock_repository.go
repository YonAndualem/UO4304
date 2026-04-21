// Package testutil provides shared test helpers used across all layers.
// Nothing in this package is compiled into the production binary.
package testutil

import (
	"context"
	"sync"

	"github.com/enterprise/trade-license/src/domain/tradelivense"
)

// MockRepository is an in-memory implementation of tradelivense.ApplicationRepository.
// It is used in unit tests to avoid a real database dependency, keeping tests fast
// and self-contained. The mutex makes it safe for concurrent test use.
type MockRepository struct {
	mu   sync.RWMutex
	data map[string]*tradelivense.TradeLicenseApplication
}

func NewMockRepository() *MockRepository {
	return &MockRepository{data: make(map[string]*tradelivense.TradeLicenseApplication)}
}

func (r *MockRepository) Save(_ context.Context, app *tradelivense.TradeLicenseApplication) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[app.ID.String()] = app
	return nil
}

func (r *MockRepository) Update(_ context.Context, app *tradelivense.TradeLicenseApplication) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[app.ID.String()] = app
	return nil
}

func (r *MockRepository) FindByID(_ context.Context, id tradelivense.ApplicationID) (*tradelivense.TradeLicenseApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	app, ok := r.data[id.String()]
	if !ok {
		return nil, tradelivense.ErrApplicationNotFound
	}
	return app, nil
}

func (r *MockRepository) FindByApplicantID(_ context.Context, applicantID string) ([]*tradelivense.TradeLicenseApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*tradelivense.TradeLicenseApplication
	for _, app := range r.data {
		if app.ApplicantID == applicantID {
			result = append(result, app)
		}
	}
	return result, nil
}

func (r *MockRepository) FindByStatus(_ context.Context, status tradelivense.ApplicationStatus) ([]*tradelivense.TradeLicenseApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*tradelivense.TradeLicenseApplication
	for _, app := range r.data {
		if app.Status == status {
			result = append(result, app)
		}
	}
	return result, nil
}

func (r *MockRepository) Delete(_ context.Context, id tradelivense.ApplicationID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id.String()]; !ok {
		return tradelivense.ErrApplicationNotFound
	}
	delete(r.data, id.String())
	return nil
}

// ─── Test fixture helpers ─────────────────────────────────────────────────────

// NewValidLicenseType returns a TRADE_LICENSE LicenseType for use in tests.
func NewValidLicenseType() tradelivense.LicenseType {
	lt, _ := tradelivense.NewLicenseType(tradelivense.TradeLicense)
	return lt
}

// NewPendingApplication creates a minimal PENDING application for use in tests.
func NewPendingApplication(applicantID string) *tradelivense.TradeLicenseApplication {
	app := tradelivense.NewTradeLicenseApplication(applicantID, NewValidLicenseType())
	return app
}

// NewReadyToSubmitApplication creates an application with all pre-conditions
// satisfied so that Submit() will succeed.
func NewReadyToSubmitApplication(applicantID string) *tradelivense.TradeLicenseApplication {
	app := NewPendingApplication(applicantID)
	app.SelectCommodity(tradelivense.NewCommodity("General Trading", "Import/Export", "Commerce"))
	app.AttachDocument(tradelivense.NewDocument("Passport", "https://storage/passport.pdf", "application/pdf"))
	app.SettlePayment(tradelivense.NewPayment(500.00, "USD", "TXN-001"))
	return app
}
