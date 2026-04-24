// Package models contains the domain models and the TradeLicenseApplication
// aggregate root for the Trade License bounded context.
package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/enterprise/trade-license/src/domain/common"
	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/events"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

// ─── Commodity ───────────────────────────────────────────────────────────────

// Commodity represents the specific trade activity or product the applicant
// intends to operate under the license. Owned by TradeLicenseApplication.
type Commodity struct {
	ID          string
	Name        string
	Description string
	Category    string
}

func NewCommodity(name, description, category string) Commodity {
	return Commodity{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Category:    category,
	}
}

// ─── Document ────────────────────────────────────────────────────────────────

// Document represents a single supporting file attached to an application.
// The URL is a reference to object storage — the domain does not manage bytes.
type Document struct {
	ID          string
	Name        string
	URL         string
	ContentType string
	UploadedAt  time.Time
}

func NewDocument(name, url, contentType string) Document {
	return Document{
		ID:          uuid.New().String(),
		Name:        name,
		URL:         url,
		ContentType: contentType,
		UploadedAt:  time.Now(),
	}
}

// ─── HistoryEntry ─────────────────────────────────────────────────────────────

// HistoryEntry records one status transition in the append-only audit trail.
type HistoryEntry struct {
	ID         string
	ActorID    string
	Action     string
	FromStatus valueobjects.ApplicationStatus
	ToStatus   valueobjects.ApplicationStatus
	Notes      string
	OccurredAt time.Time
}

func NewHistoryEntry(actorID, action string, from, to valueobjects.ApplicationStatus, notes string) HistoryEntry {
	return HistoryEntry{
		ID:         uuid.New().String(),
		ActorID:    actorID,
		Action:     action,
		FromStatus: from,
		ToStatus:   to,
		Notes:      notes,
		OccurredAt: time.Now(),
	}
}

// ─── Payment ─────────────────────────────────────────────────────────────────

// PaymentStatus tracks whether the fee associated with an application has been settled.
type PaymentStatus string

const (
	PaymentSettled PaymentStatus = "SETTLED"
	PaymentPending PaymentStatus = "PENDING"
	PaymentFailed  PaymentStatus = "FAILED"
)

// Payment records the fee settlement for an application.
type Payment struct {
	ID            string
	Amount        float64
	Currency      string
	TransactionID string
	PaidAt        time.Time
	Status        PaymentStatus
}

func NewPayment(amount float64, currency, transactionID string) Payment {
	return Payment{
		ID:            uuid.New().String(),
		Amount:        amount,
		Currency:      currency,
		TransactionID: transactionID,
		PaidAt:        time.Now(),
		Status:        PaymentSettled,
	}
}

// ─── TradeLicenseApplication (Aggregate Root) ────────────────────────────────

// TradeLicenseApplication is the aggregate root for the Trade License bounded context.
// All state mutations go through this struct's methods so that business invariants
// are enforced in one place.
type TradeLicenseApplication struct {
	common.AggregateRoot

	ID          valueobjects.ApplicationID
	LicenseType valueobjects.LicenseType
	ApplicantID string
	Commodity   *Commodity
	Documents   []Document
	Payment     *Payment
	Status      valueobjects.ApplicationStatus
	Notes       string
	History     []HistoryEntry
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewTradeLicenseApplication creates a new application in the PENDING state.
func NewTradeLicenseApplication(applicantID string, licenseType valueobjects.LicenseType) *TradeLicenseApplication {
	return &TradeLicenseApplication{
		ID:          valueobjects.NewApplicationID(),
		LicenseType: licenseType,
		ApplicantID: applicantID,
		Status:      valueobjects.StatusPending,
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

func (a *TradeLicenseApplication) Submit() error {
	if a.Status != valueobjects.StatusPending {
		return domainerrors.ErrInvalidStatusTransition
	}
	if len(a.Documents) == 0 {
		return domainerrors.ErrDocumentRequired
	}
	if a.Payment == nil {
		return domainerrors.ErrPaymentRequired
	}
	a.addHistory(a.ApplicantID, "SUBMIT", valueobjects.StatusPending, valueobjects.StatusSubmitted, "")
	a.Status = valueobjects.StatusSubmitted
	a.UpdatedAt = time.Now()
	a.AddEvent(events.ApplicationSubmittedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApplicantID:     a.ApplicantID,
	})
	return nil
}

func (a *TradeLicenseApplication) Cancel() error {
	if a.Status != valueobjects.StatusPending && a.Status != valueobjects.StatusAdjusted {
		return domainerrors.ErrInvalidStatusTransition
	}
	prev := a.Status
	a.addHistory(a.ApplicantID, "CANCEL", prev, valueobjects.StatusCancelled, "")
	a.Status = valueobjects.StatusCancelled
	a.UpdatedAt = time.Now()
	a.AddEvent(events.ApplicationCancelledEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApplicantID:     a.ApplicantID,
	})
	return nil
}

func (a *TradeLicenseApplication) UpdateDetails(commodity Commodity, documents []Document) error {
	if a.Status != valueobjects.StatusPending && a.Status != valueobjects.StatusAdjusted {
		return domainerrors.ErrInvalidStatusTransition
	}
	if len(documents) == 0 {
		return domainerrors.ErrDocumentRequired
	}
	a.Commodity = &commodity
	a.Documents = documents
	a.addHistory(a.ApplicantID, "UPDATE", a.Status, a.Status, "")
	a.UpdatedAt = time.Now()
	return nil
}

func (a *TradeLicenseApplication) ReplacePayment(amount float64, currency, transactionID string) error {
	if a.Status != valueobjects.StatusPending && a.Status != valueobjects.StatusAdjusted {
		return domainerrors.ErrInvalidStatusTransition
	}
	if a.Payment != nil {
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

func (a *TradeLicenseApplication) Resubmit() error {
	if a.Status != valueobjects.StatusAdjusted {
		return domainerrors.ErrInvalidStatusTransition
	}
	if len(a.Documents) == 0 {
		return domainerrors.ErrDocumentRequired
	}
	if a.Payment == nil {
		return domainerrors.ErrPaymentRequired
	}
	a.addHistory(a.ApplicantID, "RESUBMIT", valueobjects.StatusAdjusted, valueobjects.StatusSubmitted, a.Notes)
	a.Notes = ""
	a.Status = valueobjects.StatusSubmitted
	a.UpdatedAt = time.Now()
	a.AddEvent(events.ApplicationResubmittedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApplicantID:     a.ApplicantID,
	})
	return nil
}

func (a *TradeLicenseApplication) Delete() error {
	switch a.Status {
	case valueobjects.StatusPending, valueobjects.StatusCancelled, valueobjects.StatusRejected:
		return nil
	default:
		return domainerrors.ErrInvalidStatusTransition
	}
}

// ─── Reviewer actions ────────────────────────────────────────────────────────

func (a *TradeLicenseApplication) Accept(reviewerID string) error {
	if a.Status != valueobjects.StatusSubmitted && a.Status != valueobjects.StatusRereview {
		return domainerrors.ErrInvalidStatusTransition
	}
	prev := a.Status
	a.addHistory(reviewerID, "ACCEPT", prev, valueobjects.StatusAccepted, "")
	a.Status = valueobjects.StatusAccepted
	a.UpdatedAt = time.Now()
	a.AddEvent(events.ApplicationAcceptedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ReviewerID:      reviewerID,
	})
	return nil
}

func (a *TradeLicenseApplication) ReviewReject(reviewerID, reason string) error {
	if a.Status != valueobjects.StatusSubmitted && a.Status != valueobjects.StatusRereview {
		return domainerrors.ErrInvalidStatusTransition
	}
	prev := a.Status
	a.addHistory(reviewerID, "REJECT", prev, valueobjects.StatusRejected, reason)
	a.Status = valueobjects.StatusRejected
	a.Notes = reason
	a.UpdatedAt = time.Now()
	a.AddEvent(events.ApplicationRejectedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ActorID:         reviewerID,
		Reason:          reason,
	})
	return nil
}

func (a *TradeLicenseApplication) Adjust(reviewerID, notes string) error {
	if a.Status != valueobjects.StatusSubmitted && a.Status != valueobjects.StatusRereview {
		return domainerrors.ErrInvalidStatusTransition
	}
	prev := a.Status
	a.addHistory(reviewerID, "ADJUST", prev, valueobjects.StatusAdjusted, notes)
	a.Status = valueobjects.StatusAdjusted
	a.Notes = notes
	a.UpdatedAt = time.Now()
	a.AddEvent(events.ApplicationAdjustedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ReviewerID:      reviewerID,
		Notes:           notes,
	})
	return nil
}

// ─── Approver actions ────────────────────────────────────────────────────────

func (a *TradeLicenseApplication) Approve(approverID string) error {
	if a.Status != valueobjects.StatusAccepted {
		return domainerrors.ErrInvalidStatusTransition
	}
	a.addHistory(approverID, "APPROVE", valueobjects.StatusAccepted, valueobjects.StatusApproved, "")
	a.Status = valueobjects.StatusApproved
	a.UpdatedAt = time.Now()
	a.AddEvent(events.ApplicationApprovedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApproverID:      approverID,
	})
	return nil
}

func (a *TradeLicenseApplication) ApproveReject(approverID, reason string) error {
	if a.Status != valueobjects.StatusAccepted {
		return domainerrors.ErrInvalidStatusTransition
	}
	a.addHistory(approverID, "REJECT", valueobjects.StatusAccepted, valueobjects.StatusRejected, reason)
	a.Status = valueobjects.StatusRejected
	a.Notes = reason
	a.UpdatedAt = time.Now()
	a.AddEvent(events.ApplicationRejectedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ActorID:         approverID,
		Reason:          reason,
	})
	return nil
}

func (a *TradeLicenseApplication) Rereview(approverID, notes string) error {
	if a.Status != valueobjects.StatusAccepted {
		return domainerrors.ErrInvalidStatusTransition
	}
	a.addHistory(approverID, "REREVIEW", valueobjects.StatusAccepted, valueobjects.StatusRereview, notes)
	a.Status = valueobjects.StatusRereview
	a.Notes = notes
	a.UpdatedAt = time.Now()
	a.AddEvent(events.ApplicationRereviewEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(),
		ApplicationID:   a.ID.String(),
		ApproverID:      approverID,
		Notes:           notes,
	})
	return nil
}

// ─── Private helpers ─────────────────────────────────────────────────────────

func (a *TradeLicenseApplication) addHistory(actorID, action string, from, to valueobjects.ApplicationStatus, notes string) {
	a.History = append(a.History, NewHistoryEntry(actorID, action, from, to, notes))
}
