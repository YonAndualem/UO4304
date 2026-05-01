package models

import (
	"time"

	"github.com/enterprise/trade-license/src/domain/common"
	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/events"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

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
