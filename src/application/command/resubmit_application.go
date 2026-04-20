package command

import (
	"context"

	"github.com/enterprise/trade-license/src/domain/tradelivense"
)

// ResubmitApplicationCommand carries the updated commodity, documents, and
// optional payment that a customer provides when resubmitting an ADJUSTED application.
//
// This command covers the full customer workflow for an adjustment cycle:
//   - Step 1: updated commodity
//   - Step 2: updated documents
//   - Step 3: optional payment correction (nil = keep existing payment)
//   - Step 4: Resubmit action → ADJUSTED → SUBMITTED
type ResubmitApplicationCommand struct {
	ApplicationID string
	ApplicantID   string
	Commodity     CommodityInput
	Documents     []DocumentInput
	Payment       *PaymentInput // Optional — nil preserves the current payment.
}

// ResubmitApplicationHandler executes the "Resubmit Application" use case.
// This is the closing step of the ADJUSTED cycle: the customer has addressed
// the reviewer's notes and wants to put the application back in the review queue.
type ResubmitApplicationHandler struct {
	repo tradelivense.ApplicationRepository
}

// NewResubmitApplicationHandler constructs the handler with its repository dependency.
func NewResubmitApplicationHandler(repo tradelivense.ApplicationRepository) *ResubmitApplicationHandler {
	return &ResubmitApplicationHandler{repo: repo}
}

// Handle runs the resubmit use case:
//  1. Resolve the ApplicationID value object.
//  2. Load the aggregate.
//  3. Ownership check.
//  4. Replace commodity and documents.
//  5. Optionally re-settle payment.
//  6. Call Resubmit() — transitions ADJUSTED → SUBMITTED and archives reviewer notes.
//  7. Persist the aggregate.
func (h *ResubmitApplicationHandler) Handle(ctx context.Context, cmd ResubmitApplicationCommand) error {
	id, err := tradelivense.ApplicationIDFrom(cmd.ApplicationID)
	if err != nil {
		return tradelivense.ErrApplicationNotFound
	}

	app, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if app.ApplicantID != cmd.ApplicantID {
		return tradelivense.ErrForbidden
	}

	commodity := tradelivense.NewCommodity(cmd.Commodity.Name, cmd.Commodity.Description, cmd.Commodity.Category)
	docs := make([]tradelivense.Document, 0, len(cmd.Documents))
	for _, d := range cmd.Documents {
		docs = append(docs, tradelivense.NewDocument(d.Name, d.URL, d.ContentType))
	}

	if err := app.UpdateDetails(commodity, docs); err != nil {
		return err
	}

	if cmd.Payment != nil {
		if err := app.ReplacePayment(cmd.Payment.Amount, cmd.Payment.Currency, cmd.Payment.TransactionID); err != nil {
			return err
		}
	}

	// Resubmit archives the reviewer's adjustment notes into history, clears the
	// Notes field, and transitions ADJUSTED → SUBMITTED.
	if err := app.Resubmit(); err != nil {
		return err
	}

	return h.repo.Update(ctx, app)
}
