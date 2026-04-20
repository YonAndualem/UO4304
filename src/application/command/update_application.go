package command

import (
	"context"

	"github.com/enterprise/trade-license/src/domain/tradelivense"
)

// UpdateApplicationCommand carries the fields a customer may change on a PENDING
// or ADJUSTED application. Payment is optional: a nil pointer means "keep the
// existing payment unchanged"; a non-nil pointer triggers ReplacePayment so the
// customer can correct amount or transaction ID before (re)submitting.
//
// This command satisfies use-case Steps 1–3 for the editing workflow:
//   - Step 1: updated commodity (SelectCommodity)
//   - Step 2: updated documents (AttachDocument)
//   - Step 3: optional payment re-settlement (ReplacePayment)
type UpdateApplicationCommand struct {
	ApplicationID string
	ApplicantID   string
	Commodity     CommodityInput
	Documents     []DocumentInput
	Payment       *PaymentInput // Optional — nil preserves the current payment.
}

// UpdateApplicationHandler executes the "Edit Application" use case.
// It is the application-layer orchestrator: load → authorise → mutate → persist.
// All business invariants (valid status, doc count, payment presence) are
// enforced by the domain aggregate, not here.
type UpdateApplicationHandler struct {
	repo tradelivense.ApplicationRepository
}

// NewUpdateApplicationHandler constructs the handler with its repository dependency.
func NewUpdateApplicationHandler(repo tradelivense.ApplicationRepository) *UpdateApplicationHandler {
	return &UpdateApplicationHandler{repo: repo}
}

// Handle runs the update use case:
//  1. Resolve the ApplicationID value object (invalid UUID → ErrApplicationNotFound).
//  2. Load the aggregate — returns ErrApplicationNotFound if it does not exist.
//  3. Ownership check — returns ErrForbidden if the caller is not the applicant.
//  4. Replace commodity and documents via UpdateDetails.
//  5. Optionally re-settle payment via ReplacePayment.
//  6. Persist the updated aggregate.
//
// Returns the updated ApplicationDTO on success (populated by the HTTP handler
// from a subsequent GetApplication query — this handler is write-only).
func (h *UpdateApplicationHandler) Handle(ctx context.Context, cmd UpdateApplicationCommand) error {
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

	// Only re-settle payment if the caller explicitly provided new payment details.
	if cmd.Payment != nil {
		if err := app.ReplacePayment(cmd.Payment.Amount, cmd.Payment.Currency, cmd.Payment.TransactionID); err != nil {
			return err
		}
	}

	return h.repo.Update(ctx, app)
}
