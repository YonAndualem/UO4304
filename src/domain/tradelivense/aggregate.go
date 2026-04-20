package tradelivense

import (
	"time"

	"github.com/enterprise/trade-license/src/domain/common"
)

// TradeLicenseApplication is the aggregate root for the Trade License bounded context.
//
// All state mutations go through this struct's methods so that business invariants
// are enforced in one place. No external layer may modify Status, Notes, or child
// entities directly.
//
// Workflow roles:
//
//	Customer  → Submit / Cancel / Resubmit / UpdateDetails / Delete
//	Reviewer  → Accept / ReviewReject / Adjust
//	Approver  → Approve / ApproveReject / Rereview
type TradeLicenseApplication struct {
	common.AggregateRoot

	ID          ApplicationID
	LicenseType LicenseType
	ApplicantID string
	Commodity   *Commodity
	Documents   []Document
	Payment     *Payment
	Status      ApplicationStatus
	Notes       string        // Latest reviewer/approver note (cleared on Resubmit)
	History     []HistoryEntry // Append-only audit trail of every status transition
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewTradeLicenseApplication creates a new application in the PENDING state.
func NewTradeLicenseApplication(applicantID string, licenseType LicenseType) *TradeLicenseApplication {
	return &TradeLicenseApplication{
		ID:          NewApplicationID(),
		LicenseType: licenseType,
		ApplicantID: applicantID,
		Status:      StatusPending,
		Documents:   []Document{},
		History:     []HistoryEntry{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ─── Customer setup mutations (Steps 1–3) ────────────────────────────────────

func (a *TradeLicenseApplication) SelectCommodity(commodity Commodity) {
	a.Commodity = &commodity
	a.UpdatedAt = time.Now()
}

func (a *TradeLicenseApplication) AttachDocument(doc Document) {
	a.Documents = append(a.Documents, doc)
	a.UpdatedAt = time.Now()
}

func (a *TradeLicenseApplication) SettlePayment(payment Payment) {
	a.Payment = &payment
	a.UpdatedAt = time.Now()
}

// ─── Customer actions ────────────────────────────────────────────────────────

// Submit transitions PENDING → SUBMITTED.
func (a *TradeLicenseApplication) Submit() error {
	if a.Status != StatusPending {
		return ErrInvalidStatusTransition
	}
	if len(a.Documents) == 0 {
		return ErrDocumentRequired
	}
	if a.Payment == nil {
		return ErrPaymentRequired
	}
	a.addHistory(a.ApplicantID, "SUBMIT", StatusPending, StatusSubmitted, "")
	a.Status = StatusSubmitted
	a.UpdatedAt = time.Now()
	a.AddEvent(ApplicationSubmittedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApplicantID:     a.ApplicantID,
	})
	return nil
}

// Cancel transitions PENDING or ADJUSTED → CANCELLED.
// A customer may abandon an application that has not yet been accepted into review.
func (a *TradeLicenseApplication) Cancel() error {
	if a.Status != StatusPending && a.Status != StatusAdjusted {
		return ErrInvalidStatusTransition
	}
	prev := a.Status
	a.addHistory(a.ApplicantID, "CANCEL", prev, StatusCancelled, "")
	a.Status = StatusCancelled
	a.UpdatedAt = time.Now()
	a.AddEvent(ApplicationCancelledEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApplicantID:     a.ApplicantID,
	})
	return nil
}

// UpdateDetails replaces commodity and documents on a PENDING or ADJUSTED application.
// Payment is handled separately via ReplacePayment; this method intentionally
// decouples content edits from payment re-settlement.
func (a *TradeLicenseApplication) UpdateDetails(commodity Commodity, documents []Document) error {
	if a.Status != StatusPending && a.Status != StatusAdjusted {
		return ErrInvalidStatusTransition
	}
	if len(documents) == 0 {
		return ErrDocumentRequired
	}
	a.Commodity = &commodity
	a.Documents = documents
	a.addHistory(a.ApplicantID, "UPDATE", a.Status, a.Status, "")
	a.UpdatedAt = time.Now()
	return nil
}

// ReplacePayment re-settles the payment on a PENDING or ADJUSTED application.
// This satisfies Step 3 ("Settle Payment") when a customer needs to correct the
// payment details before submitting or resubmitting.
//
// When an existing payment record is present, its fields are mutated in-place so
// that the repository's Save will issue an UPDATE rather than an INSERT — this
// preserves the primary key and avoids violating the unique constraint on
// TransactionID across applications.
func (a *TradeLicenseApplication) ReplacePayment(amount float64, currency, transactionID string) error {
	if a.Status != StatusPending && a.Status != StatusAdjusted {
		return ErrInvalidStatusTransition
	}
	if a.Payment != nil {
		// Mutate in-place: keep the same PK so the DB row is updated, not replaced.
		a.Payment.Amount = amount
		a.Payment.Currency = currency
		a.Payment.TransactionID = transactionID
		a.Payment.PaidAt = time.Now()
	} else {
		p := NewPayment(amount, currency, transactionID)
		a.Payment = &p
	}
	a.UpdatedAt = time.Now()
	return nil
}

// Resubmit transitions ADJUSTED → SUBMITTED after the customer has addressed
// the reviewer's correction notes. Notes are archived into History then cleared
// so the reviewer sees a clean slate on the next review cycle.
func (a *TradeLicenseApplication) Resubmit() error {
	if a.Status != StatusAdjusted {
		return ErrInvalidStatusTransition
	}
	if len(a.Documents) == 0 {
		return ErrDocumentRequired
	}
	if a.Payment == nil {
		return ErrPaymentRequired
	}
	// Archive the reviewer's adjustment reason before clearing it
	a.addHistory(a.ApplicantID, "RESUBMIT", StatusAdjusted, StatusSubmitted, a.Notes)
	a.Notes = ""
	a.Status = StatusSubmitted
	a.UpdatedAt = time.Now()
	a.AddEvent(ApplicationResubmittedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApplicantID:     a.ApplicantID,
	})
	return nil
}

// Delete validates that an application is in a terminal or pre-submission state
// before the repository soft-deletes it. Active applications (under review or
// approved) cannot be deleted.
func (a *TradeLicenseApplication) Delete() error {
	switch a.Status {
	case StatusPending, StatusCancelled, StatusRejected:
		return nil
	default:
		return ErrInvalidStatusTransition
	}
}

// ─── Reviewer actions ────────────────────────────────────────────────────────

// Accept transitions SUBMITTED|REREVIEW → ACCEPTED.
func (a *TradeLicenseApplication) Accept(reviewerID string) error {
	if a.Status != StatusSubmitted && a.Status != StatusRereview {
		return ErrInvalidStatusTransition
	}
	prev := a.Status
	a.addHistory(reviewerID, "ACCEPT", prev, StatusAccepted, "")
	a.Status = StatusAccepted
	a.UpdatedAt = time.Now()
	a.AddEvent(ApplicationAcceptedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ReviewerID:      reviewerID,
	})
	return nil
}

// ReviewReject transitions SUBMITTED|REREVIEW → REJECTED.
func (a *TradeLicenseApplication) ReviewReject(reviewerID, reason string) error {
	if a.Status != StatusSubmitted && a.Status != StatusRereview {
		return ErrInvalidStatusTransition
	}
	prev := a.Status
	a.addHistory(reviewerID, "REJECT", prev, StatusRejected, reason)
	a.Status = StatusRejected
	a.Notes = reason
	a.UpdatedAt = time.Now()
	a.AddEvent(ApplicationRejectedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ActorID:         reviewerID,
		Reason:          reason,
	})
	return nil
}

// Adjust transitions SUBMITTED|REREVIEW → ADJUSTED.
func (a *TradeLicenseApplication) Adjust(reviewerID, notes string) error {
	if a.Status != StatusSubmitted && a.Status != StatusRereview {
		return ErrInvalidStatusTransition
	}
	prev := a.Status
	a.addHistory(reviewerID, "ADJUST", prev, StatusAdjusted, notes)
	a.Status = StatusAdjusted
	a.Notes = notes
	a.UpdatedAt = time.Now()
	a.AddEvent(ApplicationAdjustedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ReviewerID:      reviewerID,
		Notes:           notes,
	})
	return nil
}

// ─── Approver actions ────────────────────────────────────────────────────────

// Approve transitions ACCEPTED → APPROVED.
func (a *TradeLicenseApplication) Approve(approverID string) error {
	if a.Status != StatusAccepted {
		return ErrInvalidStatusTransition
	}
	a.addHistory(approverID, "APPROVE", StatusAccepted, StatusApproved, "")
	a.Status = StatusApproved
	a.UpdatedAt = time.Now()
	a.AddEvent(ApplicationApprovedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApproverID:      approverID,
	})
	return nil
}

// ApproveReject transitions ACCEPTED → REJECTED.
func (a *TradeLicenseApplication) ApproveReject(approverID, reason string) error {
	if a.Status != StatusAccepted {
		return ErrInvalidStatusTransition
	}
	a.addHistory(approverID, "REJECT", StatusAccepted, StatusRejected, reason)
	a.Status = StatusRejected
	a.Notes = reason
	a.UpdatedAt = time.Now()
	a.AddEvent(ApplicationRejectedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ActorID:         approverID,
		Reason:          reason,
	})
	return nil
}

// Rereview transitions ACCEPTED → REREVIEW.
func (a *TradeLicenseApplication) Rereview(approverID, notes string) error {
	if a.Status != StatusAccepted {
		return ErrInvalidStatusTransition
	}
	a.addHistory(approverID, "REREVIEW", StatusAccepted, StatusRereview, notes)
	a.Status = StatusRereview
	a.Notes = notes
	a.UpdatedAt = time.Now()
	a.AddEvent(ApplicationRereviewEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApproverID:      approverID,
		Notes:           notes,
	})
	return nil
}

// ─── Private helpers ─────────────────────────────────────────────────────────

func (a *TradeLicenseApplication) addHistory(actorID, action string, from, to ApplicationStatus, notes string) {
	a.History = append(a.History, NewHistoryEntry(actorID, action, from, to, notes))
}
