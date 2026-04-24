// Package command contains the write-side use cases of the application layer.
//
// In CQRS (Command Query Responsibility Segregation), commands mutate state
// while queries only read it. Each command handler here represents a single
// use case from the requirements document. Handlers depend only on domain
// types and the repository interface — never on HTTP, databases, or frameworks.
package command

import (
	"context"

	"github.com/enterprise/trade-license/src/domain/models"
	"github.com/enterprise/trade-license/src/domain/repositories"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

// SubmitApplicationCommand carries the data a customer provides when completing
// Steps 1–4 of the "Submit New Application for Trade License" use case.
type SubmitApplicationCommand struct {
	ApplicantID string
	LicenseType string
	Commodity   CommodityInput
	Documents   []DocumentInput
	Payment     PaymentInput
}

// CommodityInput is the data transfer object for Step 1 (Select Trade License commodity).
type CommodityInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// DocumentInput is the data transfer object for Step 2 (Attach required documents).
type DocumentInput struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
}

// PaymentInput is the data transfer object for Step 3 (Settle payment).
type PaymentInput struct {
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	TransactionID string  `json:"transaction_id"` // Reference returned by the external payment gateway
}

// SubmitApplicationHandler orchestrates the full customer submission use case.
// It builds the aggregate from raw input, delegates business rules to the domain,
// and delegates persistence to the repository — keeping orchestration logic here
// and business logic in the aggregate.
type SubmitApplicationHandler struct {
	repo repositories.ApplicationRepository
}

func NewSubmitApplicationHandler(repo repositories.ApplicationRepository) *SubmitApplicationHandler {
	return &SubmitApplicationHandler{repo: repo}
}

// Handle executes the submit use case:
//  1. Validate and construct the LicenseType value object
//  2. Create a new application aggregate (starts in PENDING)
//  3. Apply Steps 1–3 (commodity, documents, payment)
//  4. Call Submit() which enforces business rules and transitions to SUBMITTED
//  5. Persist the aggregate
//
// Returns the new application's ID on success.
func (h *SubmitApplicationHandler) Handle(ctx context.Context, cmd SubmitApplicationCommand) (string, error) {
	licenseType, err := valueobjects.NewLicenseType(cmd.LicenseType)
	if err != nil {
		return "", err
	}

	app := models.NewTradeLicenseApplication(cmd.ApplicantID, licenseType)

	commodity := models.NewCommodity(cmd.Commodity.Name, cmd.Commodity.Description, cmd.Commodity.Category)
	app.SelectCommodity(commodity)

	for _, d := range cmd.Documents {
		app.AttachDocument(models.NewDocument(d.Name, d.URL, d.ContentType))
	}

	payment := models.NewPayment(cmd.Payment.Amount, cmd.Payment.Currency, cmd.Payment.TransactionID)
	app.SettlePayment(payment)

	// Submit enforces: documents present, payment settled, status is PENDING
	if err := app.Submit(); err != nil {
		return "", err
	}

	if err := h.repo.Save(ctx, app); err != nil {
		return "", err
	}

	return app.ID.String(), nil
}
