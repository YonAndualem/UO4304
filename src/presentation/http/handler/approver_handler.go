package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/enterprise/trade-license/src/application/command"
	"github.com/enterprise/trade-license/src/application/query"
	"github.com/enterprise/trade-license/src/domain/tradelivense"
	"github.com/enterprise/trade-license/src/presentation/http/middleware"
)

// ApproverHandler groups all HTTP endpoints available to the APPROVER role.
//
// The approver use case is the final gate in the workflow:
//   - Steps 1–3: view the reviewed application, documents, and payment
//   - Step 4: grant final approval, reject, or send back for re-review
type ApproverHandler struct {
	approveHandler *command.ApproveApplicationHandler
	getHandler     *query.GetApplicationHandler
	listHandler    *query.ListByStatusHandler
}

// NewApproverHandler constructs the handler with all dependencies injected.
func NewApproverHandler(
	approve *command.ApproveApplicationHandler,
	get *query.GetApplicationHandler,
	list *query.ListByStatusHandler,
) *ApproverHandler {
	return &ApproverHandler{approveHandler: approve, getHandler: get, listHandler: list}
}

// ListPendingApproval handles GET /api/approver/applications[?status=ACCEPTED].
//
// Returns all applications that have been accepted by a reviewer and are
// awaiting the approver's final decision. Defaults to status=ACCEPTED.
// The oldest applications are listed first to prevent starvation.
func (h *ApproverHandler) ListPendingApproval(c *fiber.Ctx) error {
	status := c.Query("status", string(tradelivense.StatusAccepted))

	dtos, err := h.listHandler.Handle(c.Context(), status)
	if err != nil {
		return domainError(c, err)
	}
	return c.JSON(dtos)
}

// GetApplication handles GET /api/approver/applications/:id.
//
// Returns the full ApplicationDTO so the approver can inspect the reviewed
// application (Step 1), documents (Step 2), and payment (Step 3) before
// deciding on an action.
func (h *ApproverHandler) GetApplication(c *fiber.Ctx) error {
	dto, err := h.getHandler.Handle(c.Context(), c.Params("id"))
	if err != nil {
		return domainError(c, err)
	}
	return c.JSON(dto)
}

// TakeAction handles POST /api/approver/applications/:id/action.
//
// Implements Step 4 of the Approver use case: Action => {Approve, Reject, Rereview}.
//
//   - APPROVE  → ACCEPTED → APPROVED (workflow complete, license granted)
//   - REJECT   → ACCEPTED → REJECTED (terminal state, notes required)
//   - REREVIEW → ACCEPTED → REREVIEW (returns to reviewer queue for deeper scrutiny, notes required)
//
// Notes are mandatory for REJECT and REREVIEW to provide the reviewer or auditor
// with context about why the approval was withheld. Returns 204 No Content on success.
func (h *ApproverHandler) TakeAction(c *fiber.Ctx) error {
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
	if (req.Action == string(command.ApproveActionReject) || req.Action == string(command.ApproveActionRereview)) && req.Notes == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "notes are required for REJECT and REREVIEW actions"})
	}

	approverID := middleware.UserID(c)
	if approverID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-User-ID header is required"})
	}

	if err := h.approveHandler.Handle(c.Context(), command.ApproveApplicationCommand{
		ApplicationID: c.Params("id"),
		ApproverID:    approverID,
		Action:        command.ApproveAction(req.Action),
		Notes:         req.Notes,
	}); err != nil {
		return domainError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
