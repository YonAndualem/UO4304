// Package http wires the Fiber application, middleware, and route groups together.
//
// This is the outermost layer of Clean Architecture — it knows about HTTP concerns
// (verbs, paths, status codes) but delegates all business logic to the handlers,
// which in turn delegate to the application layer.
//
// Route design follows REST conventions with role-scoped path prefixes:
//   - /api/customer  — endpoints for the CUSTOMER role
//   - /api/reviewer  — endpoints for the REVIEWER role
//   - /api/approver  — endpoints for the APPROVER role
//
// Each group is protected by a RequireRole middleware that checks the X-Role header.
// In production this would verify a signed JWT; here it trusts the header directly.
package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/enterprise/trade-license/src/presentation/http/handler"
	"github.com/enterprise/trade-license/src/presentation/http/middleware"
)

// NewRouter builds and returns the fully configured Fiber application.
// All route groups, middleware chains, and error handlers are registered here.
func NewRouter(
	customer *handler.CustomerHandler,
	reviewer *handler.ReviewerHandler,
	approver *handler.ApproverHandler,
) *fiber.App {
	app := fiber.New(fiber.Config{
		// Global error handler: ensures unhandled panics or errors always return
		// a consistent JSON body rather than an empty response or HTML stack trace.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New()) // Recover from panics and return 500 rather than crashing.
	app.Use(logger.New())  // Structured request/response logging for observability.

	// Health check — used by Docker/Kubernetes probes to confirm the process is up.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")

	// ── Customer routes ───────────────────────────────────────────────────────
	// Use Case: Applicant requests new application for Trade License
	// Role: CUSTOMER
	cust := api.Group("/customer", middleware.RequireRole("CUSTOMER"))
	cust.Post("/applications", customer.SubmitApplication)                 // Steps 1–4: submit new application
	cust.Get("/applications", customer.ListMyApplications)                 // View all my applications
	cust.Get("/applications/:id", customer.GetApplication)                 // View a single application
	cust.Put("/applications/:id", customer.UpdateApplication)              // Steps 1–3: edit PENDING/ADJUSTED
	cust.Post("/applications/:id/resubmit", customer.ResubmitApplication)  // Step 4: resubmit ADJUSTED → SUBMITTED
	cust.Post("/applications/:id/cancel", customer.CancelApplication)      // Step 4: cancel PENDING/ADJUSTED
	cust.Delete("/applications/:id", customer.DeleteApplication)           // Soft-delete PENDING/CANCELLED/REJECTED

	// ── Reviewer routes ───────────────────────────────────────────────────────
	// Use Case: Review Submitted New Application for Trade License
	// Role: REVIEWER
	rev := api.Group("/reviewer", middleware.RequireRole("REVIEWER"))
	rev.Get("/applications", reviewer.ListPendingReview)           // Steps 1–3: reviewer work queue
	rev.Get("/applications/:id", reviewer.GetApplication)          // Steps 1–3: view application details
	rev.Post("/applications/:id/action", reviewer.TakeAction)      // Step 4: Accept / Reject / Adjust

	// ── Approver routes ───────────────────────────────────────────────────────
	// Use Case: Approve Reviewed New Application for Trade License
	// Role: APPROVER
	appr := api.Group("/approver", middleware.RequireRole("APPROVER"))
	appr.Get("/applications", approver.ListPendingApproval)         // Steps 1–3: approver work queue
	appr.Get("/applications/:id", approver.GetApplication)          // Steps 1–3: view application details
	appr.Post("/applications/:id/action", approver.TakeAction)      // Step 4: Approve / Reject / Rereview

	return app
}
