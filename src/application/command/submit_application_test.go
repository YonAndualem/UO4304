package command_test

// Tests for the SubmitApplicationHandler use case.
// The mock repository is injected so no real database is needed.

import (
	"context"
	"testing"

	"github.com/enterprise/trade-license/src/application/command"
	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
	"github.com/enterprise/trade-license/src/testutil"
)

func TestSubmitApplicationHandler_HappyPath(t *testing.T) {
	repo := testutil.NewMockRepository()
	handler := command.NewSubmitApplicationHandler(repo)

	cmd := command.SubmitApplicationCommand{
		ApplicantID: "customer-1",
		LicenseType: valueobjects.TradeLicense,
		Commodity: command.CommodityInput{
			Name:        "General Trading",
			Description: "Import and export of consumer goods",
			Category:    "Commerce",
		},
		Documents: []command.DocumentInput{
			{Name: "Passport", URL: "https://storage/passport.pdf", ContentType: "application/pdf"},
		},
		Payment: command.PaymentInput{
			Amount:        500.00,
			Currency:      "USD",
			TransactionID: "TXN-SUBMIT-001",
		},
	}

	id, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id == "" {
		t.Error("expected a non-empty application ID")
	}

	// Verify the application was persisted with SUBMITTED status
	appID, _ := valueobjects.ApplicationIDFrom(id)
	app, err := repo.FindByID(context.Background(), appID)
	if err != nil {
		t.Fatalf("expected application to be persisted: %v", err)
	}
	if app.Status != valueobjects.StatusSubmitted {
		t.Errorf("expected SUBMITTED, got %s", app.Status)
	}
}

func TestSubmitApplicationHandler_InvalidLicenseType(t *testing.T) {
	repo := testutil.NewMockRepository()
	handler := command.NewSubmitApplicationHandler(repo)

	cmd := command.SubmitApplicationCommand{
		ApplicantID: "customer-1",
		LicenseType: "INVALID_TYPE",
		Documents:   []command.DocumentInput{{Name: "doc", URL: "url", ContentType: "pdf"}},
		Payment:     command.PaymentInput{Amount: 100, Currency: "USD", TransactionID: "TXN-002"},
	}

	_, err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Error("expected error for invalid license type")
	}
}

func TestSubmitApplicationHandler_MissingDocuments(t *testing.T) {
	repo := testutil.NewMockRepository()
	handler := command.NewSubmitApplicationHandler(repo)

	cmd := command.SubmitApplicationCommand{
		ApplicantID: "customer-1",
		LicenseType: valueobjects.TradeLicense,
		Documents:   []command.DocumentInput{}, // Empty — should fail
		Payment:     command.PaymentInput{Amount: 500, Currency: "USD", TransactionID: "TXN-003"},
	}

	_, err := handler.Handle(context.Background(), cmd)
	if err != domainerrors.ErrDocumentRequired {
		t.Errorf("expected ErrDocumentRequired, got %v", err)
	}
}
