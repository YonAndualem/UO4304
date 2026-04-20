package command

import (
	"context"

	"github.com/enterprise/trade-license/src/domain/tradelivense"
)

// DeleteApplicationCommand removes a PENDING, CANCELLED, or REJECTED application
// from the customer's list via a soft delete.
type DeleteApplicationCommand struct {
	ApplicationID string
	ApplicantID   string
}

type DeleteApplicationHandler struct {
	repo tradelivense.ApplicationRepository
}

func NewDeleteApplicationHandler(repo tradelivense.ApplicationRepository) *DeleteApplicationHandler {
	return &DeleteApplicationHandler{repo: repo}
}

func (h *DeleteApplicationHandler) Handle(ctx context.Context, cmd DeleteApplicationCommand) error {
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

	if err := app.Delete(); err != nil {
		return err
	}

	return h.repo.Delete(ctx, id)
}
