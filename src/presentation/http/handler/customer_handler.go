// Package handler contains the HTTP handler structs that translate between the
// HTTP world (Fiber context, JSON bodies, status codes) and the application layer
// (command and query handlers). Handlers are thin adapters: they parse and validate
// input, delegate to the appropriate use-case handler, and translate errors into
// HTTP responses. No business logic lives here.
package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/enterprise/trade-license/src/application/command"
	"github.com/enterprise/trade-license/src/application/query"
	"github.com/enterprise/trade-license/src/presentation/http/middleware"
)

// CustomerHandler groups all HTTP endpoints available to the CUSTOMER role.
// Each method corresponds to one customer-facing use case from the requirements.
type CustomerHandler struct {
	submitHandler   *command.SubmitApplicationHandler
	cancelHandler   *command.CancelApplicationHandler
	updateHandler   *command.UpdateApplicationHandler
	resubmitHandler *command.ResubmitApplicationHandler
	deleteHandler   *command.DeleteApplicationHandler
	getHandler      *query.GetApplicationHandler
	listHandler     *query.ListByApplicantHandler
}

// NewCustomerHandler constructs a CustomerHandler with all its dependencies injected.
// Called once at startup from main.go (the composition root).
func NewCustomerHandler(
	submit *command.SubmitApplicationHandler,
	cancel *command.CancelApplicationHandler,
	update *command.UpdateApplicationHandler,
	resubmit *command.ResubmitApplicationHandler,
	delete *command.DeleteApplicationHandler,
	get *query.GetApplicationHandler,
	list *query.ListByApplicantHandler,
) *CustomerHandler {
	return &CustomerHandler{
		submitHandler:   submit,
		cancelHandler:   cancel,
		updateHandler:   update,
		resubmitHandler: resubmit,
		deleteHandler:   delete,
		getHandler:      get,
		listHandler:     list,
	}
}

// SubmitApplication handles POST /api/customer/applications.
//
// This is the primary customer use case — it combines all four steps into one
// atomic HTTP call:
//   - Step 1: license_type selects the Trade License type
//   - Step 2: documents[] attaches supporting files
//   - Step 3: payment settles the required fee
//   - Step 4: the aggregate is submitted (PENDING → SUBMITTED)
//
// Returns 201 Created with {"application_id": "<uuid>"}.
func (h *CustomerHandler) SubmitApplication(c *fiber.Ctx) error {
	type request struct {
		LicenseType string                 `json:"license_type"`
		Commodity   command.CommodityInput `json:"commodity"`
		Documents   []command.DocumentInput `json:"documents"`
		Payment     command.PaymentInput   `json:"payment"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Presentation-layer validation: catch structurally invalid input before
	// it reaches the domain. Domain rules (e.g. status guards) are still enforced
	// by the aggregate — this layer only prevents obviously malformed requests.
	if req.LicenseType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "license_type is required"})
	}
	if len(req.Documents) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "at least one document is required"})
	}
	for _, d := range req.Documents {
		if d.Name == "" || d.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "each document must have a name and storage key"})
		}
	}
	if req.Payment.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payment.amount must be greater than zero"})
	}
	if req.Payment.TransactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payment.transaction_id is required"})
	}

	applicantID := middleware.UserID(c)
	if applicantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-User-ID header is required"})
	}

	id, err := h.submitHandler.Handle(c.Context(), command.SubmitApplicationCommand{
		ApplicantID: applicantID,
		LicenseType: req.LicenseType,
		Commodity:   req.Commodity,
		Documents:   req.Documents,
		Payment:     req.Payment,
	})
	if err != nil {
		return domainError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"application_id": id})
}

// UpdateApplication handles PUT /api/customer/applications/:id.
//
// Allows a customer to correct the commodity, documents, and optionally the
// payment on a PENDING or ADJUSTED application. Payment is optional: omitting
// the "payment" key preserves the existing settlement.
//
// Satisfies use-case Steps 1–3 for the editing workflow.
// Returns 200 OK with the updated ApplicationDTO.
func (h *CustomerHandler) UpdateApplication(c *fiber.Ctx) error {
	type request struct {
		Commodity command.CommodityInput  `json:"commodity"`
		Documents []command.DocumentInput `json:"documents"`
		Payment   *command.PaymentInput   `json:"payment"` // Pointer — nil means keep current payment.
	}

	applicantID := middleware.UserID(c)
	if applicantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-User-ID header is required"})
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate payment if the caller chose to update it.
	if req.Payment != nil {
		if req.Payment.Amount <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payment.amount must be greater than zero"})
		}
		if req.Payment.TransactionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payment.transaction_id is required"})
		}
	}

	if err := h.updateHandler.Handle(c.Context(), command.UpdateApplicationCommand{
		ApplicationID: c.Params("id"),
		ApplicantID:   applicantID,
		Commodity:     req.Commodity,
		Documents:     req.Documents,
		Payment:       req.Payment,
	}); err != nil {
		return domainError(c, err)
	}

	dto, err := h.getHandler.Handle(c.Context(), c.Params("id"))
	if err != nil {
		return domainError(c, err)
	}
	return c.JSON(dto)
}

// ResubmitApplication handles POST /api/customer/applications/:id/resubmit.
//
// Transitions an ADJUSTED application back to SUBMITTED after the customer has
// addressed the reviewer's notes. Accepts updated commodity, documents, and an
// optional payment correction (Steps 1–3), then triggers the status transition
// (Step 4 — Resubmit action).
//
// The reviewer's adjustment notes are archived into the History audit trail and
// cleared from the Notes field so the next reviewer sees a clean slate.
// Returns 200 OK with the updated ApplicationDTO.
func (h *CustomerHandler) ResubmitApplication(c *fiber.Ctx) error {
	type request struct {
		Commodity command.CommodityInput  `json:"commodity"`
		Documents []command.DocumentInput `json:"documents"`
		Payment   *command.PaymentInput   `json:"payment"` // Pointer — nil means keep current payment.
	}

	applicantID := middleware.UserID(c)
	if applicantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-User-ID header is required"})
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if req.Payment != nil {
		if req.Payment.Amount <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payment.amount must be greater than zero"})
		}
		if req.Payment.TransactionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payment.transaction_id is required"})
		}
	}

	if err := h.resubmitHandler.Handle(c.Context(), command.ResubmitApplicationCommand{
		ApplicationID: c.Params("id"),
		ApplicantID:   applicantID,
		Commodity:     req.Commodity,
		Documents:     req.Documents,
		Payment:       req.Payment,
	}); err != nil {
		return domainError(c, err)
	}

	dto, err := h.getHandler.Handle(c.Context(), c.Params("id"))
	if err != nil {
		return domainError(c, err)
	}
	return c.JSON(dto)
}

// CancelApplication handles POST /api/customer/applications/:id/cancel.
//
// Allows a customer to abandon a PENDING or ADJUSTED application.
// ADJUSTED is included so a customer who disagrees with the reviewer's adjustment
// request can withdraw rather than being forced to resubmit.
// Returns 204 No Content on success.
func (h *CustomerHandler) CancelApplication(c *fiber.Ctx) error {
	applicantID := middleware.UserID(c)
	if applicantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-User-ID header is required"})
	}

	if err := h.cancelHandler.Handle(c.Context(), command.CancelApplicationCommand{
		ApplicationID: c.Params("id"),
		ApplicantID:   applicantID,
	}); err != nil {
		return domainError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteApplication handles DELETE /api/customer/applications/:id.
//
// Permanently removes (soft-deletes) an application that is in a terminal or
// pre-submission state: PENDING, CANCELLED, or REJECTED. Applications under
// active review (SUBMITTED, ACCEPTED, REREVIEW) or approved cannot be deleted —
// the aggregate's Delete() method enforces this guard.
// Returns 204 No Content on success.
func (h *CustomerHandler) DeleteApplication(c *fiber.Ctx) error {
	applicantID := middleware.UserID(c)
	if applicantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-User-ID header is required"})
	}

	if err := h.deleteHandler.Handle(c.Context(), command.DeleteApplicationCommand{
		ApplicationID: c.Params("id"),
		ApplicantID:   applicantID,
	}); err != nil {
		return domainError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetApplication handles GET /api/customer/applications/:id.
// Returns the full ApplicationDTO including commodity, documents, payment, and
// the complete audit trail history for the specified application.
func (h *CustomerHandler) GetApplication(c *fiber.Ctx) error {
	dto, err := h.getHandler.Handle(c.Context(), c.Params("id"))
	if err != nil {
		return domainError(c, err)
	}
	return c.JSON(dto)
}

// ListMyApplications handles GET /api/customer/applications.
// Returns all non-deleted applications belonging to the authenticated customer,
// ordered newest-first. The full history and associations are included so the
// list page can display status badges and attention indicators without additional
// API calls.
func (h *CustomerHandler) ListMyApplications(c *fiber.Ctx) error {
	applicantID := middleware.UserID(c)
	if applicantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-User-ID header is required"})
	}

	dtos, err := h.listHandler.Handle(c.Context(), applicantID)
	if err != nil {
		return domainError(c, err)
	}
	return c.JSON(dtos)
}
