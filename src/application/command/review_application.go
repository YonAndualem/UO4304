package command

import (
	"context"
	"errors"

	"github.com/enterprise/trade-license/src/domain/tradelivense"
)

// ReviewAction enumerates the decisions available to a Reviewer in Step 4.
type ReviewAction string

const (
	ReviewActionAccept ReviewAction = "ACCEPT"
	ReviewActionReject ReviewAction = "REJECT"
	ReviewActionAdjust ReviewAction = "ADJUST"
)

// ReviewApplicationCommand carries the reviewer's decision for a submitted application.
// Notes are required for REJECT and ADJUST to provide the customer/approver with context.
type ReviewApplicationCommand struct {
	ApplicationID string
	ReviewerID    string
	Action        ReviewAction
	Notes         string // Required for REJECT and ADJUST actions
}

// ReviewApplicationHandler executes the "Review Submitted Application" use case.
// It routes the reviewer's chosen action to the corresponding aggregate method,
// keeping the switch-on-action logic here (orchestration) while the state
// transition rules stay in the aggregate (business logic).
type ReviewApplicationHandler struct {
	repo tradelivense.ApplicationRepository
}

func NewReviewApplicationHandler(repo tradelivense.ApplicationRepository) *ReviewApplicationHandler {
	return &ReviewApplicationHandler{repo: repo}
}

// Handle loads the application and dispatches to the appropriate aggregate method
// based on the reviewer's chosen action. Each method enforces that the application
// is in a valid state for that action (SUBMITTED or REREVIEW).
func (h *ReviewApplicationHandler) Handle(ctx context.Context, cmd ReviewApplicationCommand) error {
	appID, err := tradelivense.ApplicationIDFrom(cmd.ApplicationID)
	if err != nil {
		return err
	}

	app, err := h.repo.FindByID(ctx, appID)
	if err != nil {
		return err
	}

	switch cmd.Action {
	case ReviewActionAccept:
		err = app.Accept(cmd.ReviewerID)
	case ReviewActionReject:
		err = app.ReviewReject(cmd.ReviewerID, cmd.Notes)
	case ReviewActionAdjust:
		err = app.Adjust(cmd.ReviewerID, cmd.Notes)
	default:
		return errors.New("unknown review action: " + string(cmd.Action))
	}

	if err != nil {
		return err
	}

	return h.repo.Update(ctx, app)
}
