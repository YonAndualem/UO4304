package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/enterprise/trade-license/src/application/command"
	"github.com/enterprise/trade-license/src/application/query"
	"github.com/enterprise/trade-license/src/domain/tradelivense"
	"github.com/enterprise/trade-license/src/presentation/http/middleware"
)

// ReviewerHandler groups all HTTP endpoints available to the REVIEWER role.
//
// The reviewer use case maps directly to the three-step inspection + one-step
// action pattern defined in the requirements:
//   - Steps 1–3: view commodity, documents, payment (GET endpoints)
//   - Step 4: take action — Accept, Reject, or request Adjustment (POST action)
type ReviewerHandler struct {
	reviewHandler *command.ReviewApplicationHandler
	getHandler    *query.GetApplicationHandler
	listHandler   *query.ListByStatusHandler
}

// NewReviewerHandler constructs the handler with all dependencies injected.
func NewReviewerHandler(
	review *command.ReviewApplicationHandler,
	get *query.GetApplicationHandler,
	list *query.ListByStatusHandler,
) *ReviewerHandler {
	return &ReviewerHandler{reviewHandler: review, getHandler: get, listHandler: list}
}

// ListPendingReview handles GET /api/reviewer/applications[?status=...].
//
// Returns the reviewer's actionable work queue. Without a ?status filter it
// combines SUBMITTED and REREVIEW results so the reviewer sees their full queue
// in a single call. Pass ?status=SUBMITTED or ?status=REREVIEW to filter.
//
// REREVIEW applications are ones that an approver has sent back for additional
// examination; they appear alongside new SUBMITTED applications because the
// reviewer's actions are identical for both.
func (h *ReviewerHandler) ListPendingReview(c *fiber.Ctx) error {
	status := c.Query("status")

	// When no filter is given, merge both queues so the reviewer sees everything
	// that needs their attention without needing two separate API calls.
	if status == "" {
		submitted, err := h.listHandler.Handle(c.Context(), string(tradelivense.StatusSubmitted))
		if err != nil {
			return domainError(c, err)
		}
		rereview, err := h.listHandler.Handle(c.Context(), string(tradelivense.StatusRereview))
		if err != nil {
			return domainError(c, err)
		}
		return c.JSON(append(submitted, rereview...))
	}

	dtos, err := h.listHandler.Handle(c.Context(), status)
	if err != nil {
		return domainError(c, err)
	}
	return c.JSON(dtos)
}

// GetApplication handles GET /api/reviewer/applications/:id.
//
// Returns the full ApplicationDTO so the reviewer can inspect commodity (Step 1),
// documents (Step 2), and payment (Step 3) before deciding on an action.
func (h *ReviewerHandler) GetApplication(c *fiber.Ctx) error {
	dto, err := h.getHandler.Handle(c.Context(), c.Params("id"))
	if err != nil {
		return domainError(c, err)
	}
	return c.JSON(dto)
}

// TakeAction handles POST /api/reviewer/applications/:id/action.
//
// Implements Step 4 of the Reviewer use case: Action => {Accept, Reject, Adjust}.
//
//   - ACCEPT  → SUBMITTED|REREVIEW → ACCEPTED (forwards to approver queue)
//   - REJECT  → SUBMITTED|REREVIEW → REJECTED (terminal state, notes required)
//   - ADJUST  → SUBMITTED|REREVIEW → ADJUSTED (returns to customer for correction, notes required)
//
// Notes are mandatory for REJECT and ADJUST so the customer or auditor always
// knows why the application was not accepted. Returns 204 No Content on success.
func (h *ReviewerHandler) TakeAction(c *fiber.Ctx) error {
	type request struct {
		Action string `json:"action"`
		Notes  string `json:"notes"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "action is required"})
	}
	// Notes are required when the reviewer is explaining a negative or corrective decision.
	if (req.Action == string(command.ReviewActionReject) || req.Action == string(command.ReviewActionAdjust)) && req.Notes == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "notes are required for REJECT and ADJUST actions"})
	}

	reviewerID := middleware.UserID(c)
	if reviewerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-User-ID header is required"})
	}

	if err := h.reviewHandler.Handle(c.Context(), command.ReviewApplicationCommand{
		ApplicationID: c.Params("id"),
		ReviewerID:    reviewerID,
		Action:        command.ReviewAction(req.Action),
		Notes:         req.Notes,
	}); err != nil {
		return domainError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
