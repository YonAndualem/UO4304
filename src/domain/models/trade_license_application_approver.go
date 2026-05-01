package models

import (
	"time"

	"github.com/enterprise/trade-license/src/domain/common"
	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/events"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

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
