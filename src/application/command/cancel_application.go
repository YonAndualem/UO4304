package command

import (
	"context"

	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/repositories"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

// CancelApplicationCommand carries the data needed to cancel a pending application.
// The ApplicantID is included so that higher-level authorization logic can
// verify that the requester owns the application (enforced at the handler layer).
type CancelApplicationCommand struct {
	ApplicationID string
	ApplicantID   string
}

// CancelApplicationHandler executes the "Cancel Application" use case (Step 4 — Cancel action).
// It loads the aggregate, delegates the cancellation rule to the domain, and persists
// the updated state. All status-transition logic remains in the aggregate.
type CancelApplicationHandler struct {
	repo repositories.ApplicationRepository
}

func NewCancelApplicationHandler(repo repositories.ApplicationRepository) *CancelApplicationHandler {
	return &CancelApplicationHandler{repo: repo}
}

func (h *CancelApplicationHandler) Handle(ctx context.Context, cmd CancelApplicationCommand) error {
	appID, err := valueobjects.ApplicationIDFrom(cmd.ApplicationID)
	if err != nil {
		return domainerrors.ErrApplicationNotFound
	}

	app, err := h.repo.FindByID(ctx, appID)
	if err != nil {
		return err
	}

	if app.ApplicantID != cmd.ApplicantID {
		return domainerrors.ErrForbidden
	}

	if err := app.Cancel(); err != nil {
		return err
	}

	return h.repo.Update(ctx, app)
}
