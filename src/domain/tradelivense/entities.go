package tradelivense

import (
	"time"

	"github.com/google/uuid"
)

// Commodity is a domain entity representing the specific trade activity or
// product category the applicant intends to operate under the license.
// It is owned by TradeLicenseApplication and has no independent lifecycle.
type Commodity struct {
	ID          string
	Name        string
	Description string
	Category    string
}

// NewCommodity constructs a Commodity with a generated ID.
// The ID is assigned here so the domain controls identity, not the database.
func NewCommodity(name, description, category string) Commodity {
	return Commodity{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Category:    category,
	}
}

// ─── Document ────────────────────────────────────────────────────────────────

// Document is a domain entity representing a single supporting file attached
// to an application (e.g. passport copy, business registration certificate).
// The URL is a reference to where the file is stored (object storage, etc.);
// the domain does not manage file bytes directly.
type Document struct {
	ID          string
	Name        string
	URL         string
	ContentType string
	UploadedAt  time.Time
}

// NewDocument constructs a Document with a generated ID and the current timestamp.
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

// HistoryEntry records one status transition so the full audit trail of an
// application is preserved across cycles (submit → adjust → resubmit → …).
type HistoryEntry struct {
	ID         string
	ActorID    string            // Who triggered the transition
	Action     string            // SUBMIT, CANCEL, ACCEPT, REJECT, ADJUST, APPROVE, REREVIEW, RESUBMIT, UPDATE
	FromStatus ApplicationStatus // Status before the transition
	ToStatus   ApplicationStatus // Status after the transition
	Notes      string            // Reason / reviewer comment captured at this step
	OccurredAt time.Time
}

// NewHistoryEntry creates a history entry with a generated ID.
func NewHistoryEntry(actorID, action string, from, to ApplicationStatus, notes string) HistoryEntry {
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

// PaymentStatus tracks whether the payment associated with an application has
// been settled by the external payment gateway.
type PaymentStatus string

const (
	PaymentSettled PaymentStatus = "SETTLED"
	PaymentPending PaymentStatus = "PENDING"
	PaymentFailed  PaymentStatus = "FAILED"
)

// Payment is a domain entity recording the fee settlement for an application.
// The application may only be submitted once payment status is SETTLED, enforcing
// that no application proceeds without proof of payment.
type Payment struct {
	ID            string
	Amount        float64
	Currency      string
	TransactionID string // External reference from the payment gateway
	PaidAt        time.Time
	Status        PaymentStatus
}

// NewPayment constructs a Payment in the SETTLED state.
// Only call this after the payment gateway has confirmed the transaction.
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
