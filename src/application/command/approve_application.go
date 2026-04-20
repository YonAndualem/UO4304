package command

import (
	"context"
	"errors"

	"github.com/enterprise/trade-license/src/domain/tradelivense"
)

// ApproveAction enumerates the decisions available to an Approver in Step 4.
type ApproveAction string

const (
	ApproveActionApprove  ApproveAction = "APPROVE"
	ApproveActionReject   ApproveAction = "REJECT"
	ApproveActionRereview ApproveAction = "REREVIEW"
)

// ApproveApplicationCommand carries the approver's decision for a reviewed application.
// Notes are required for REJECT and REREVIEW to explain the decision.
type ApproveApplicationCommand struct {
	ApplicationID string
	ApproverID    string
	Action        ApproveAction
	Notes         string // Required for REJECT and REREVIEW actions
}

// ApproveApplicationHandler executes the "Approve Reviewed Application" use case.
// Mirrors ReviewApplicationHandler in structure — the action routing lives here
// while the transition guards and domain events live in the aggregate.
type ApproveApplicationHandler struct {
	repo tradelivense.ApplicationRepository
}

func NewApproveApplicationHandler(repo tradelivense.ApplicationRepository) *ApproveApplicationHandler {
	return &ApproveApplicationHandler{repo: repo}
}

// Handle loads the application and dispatches to the appropriate aggregate method
// based on the approver's action. Each method requires the application to be
// in ACCEPTED status; REREVIEW sends it back to the reviewer queue.
func (h *ApproveApplicationHandler) Handle(ctx context.Context, cmd ApproveApplicationCommand) error {
	appID, err := tradelivense.ApplicationIDFrom(cmd.ApplicationID)
	if err != nil {
		return err
	}

	app, err := h.repo.FindByID(ctx, appID)
	if err != nil {
		return err
	}

	switch cmd.Action {
	case ApproveActionApprove:
		err = app.Approve(cmd.ApproverID)
	case ApproveActionReject:
		err = app.ApproveReject(cmd.ApproverID, cmd.Notes)
	case ApproveActionRereview:
		err = app.Rereview(cmd.ApproverID, cmd.Notes)
	default:
		return errors.New("unknown approve action: " + string(cmd.Action))
	}

	if err != nil {
		return err
	}

	return h.repo.Update(ctx, app)
}
