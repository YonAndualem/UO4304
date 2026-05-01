package models

import (
	"time"

	"github.com/enterprise/trade-license/src/domain/common"
	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/events"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

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
