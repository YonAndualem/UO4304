package command

import (
	"context"

	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/repositories"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
)

// DeleteApplicationCommand removes a PENDING, CANCELLED, or REJECTED application
// from the customer's list via a soft delete.
type DeleteApplicationCommand struct {
	ApplicationID string
	ApplicantID   string
}

type DeleteApplicationHandler struct {
	repo repositories.ApplicationRepository
}

func NewDeleteApplicationHandler(repo repositories.ApplicationRepository) *DeleteApplicationHandler {
	return &DeleteApplicationHandler{repo: repo}
}

func (h *DeleteApplicationHandler) Handle(ctx context.Context, cmd DeleteApplicationCommand) error {
	id, err := valueobjects.ApplicationIDFrom(cmd.ApplicationID)
	if err != nil {
		return domainerrors.ErrApplicationNotFound
	}

	app, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if app.ApplicantID != cmd.ApplicantID {
		return domainerrors.ErrForbidden
	}

	if err := app.Delete(); err != nil {
		return err
	}

	return h.repo.Delete(ctx, id)
}
