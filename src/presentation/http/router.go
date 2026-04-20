// Package http wires the Fiber application, middleware, and route groups together.
//
// Route design follows REST conventions with role-scoped path prefixes:
//   - /api/auth      — public endpoints (register, login)
//   - /api/customer  — CUSTOMER role (JWT-protected)
//   - /api/reviewer  — REVIEWER role (JWT-protected)
//   - /api/approver  — APPROVER role (JWT-protected)
package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/enterprise/trade-license/src/application/auth"
	"github.com/enterprise/trade-license/src/presentation/http/handler"
	"github.com/enterprise/trade-license/src/presentation/http/middleware"
)

// NewRouter builds and returns the fully configured Fiber application.
func NewRouter(
	authHandler *handler.AuthHandler,
	customer *handler.CustomerHandler,
	reviewer *handler.ReviewerHandler,
	approver *handler.ApproverHandler,
	authSvc *auth.Service,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")

	// ── Public auth routes ────────────────────────────────────────────────────
	a := api.Group("/auth")
	a.Post("/register", authHandler.Register)
	a.Post("/login", authHandler.Login)

	// ── Customer routes (JWT required, CUSTOMER role) ─────────────────────────
	cust := api.Group("/customer", middleware.JWTAuth(authSvc), middleware.RequireRole("CUSTOMER"))
	cust.Post("/applications", customer.SubmitApplication)
	cust.Get("/applications", customer.ListMyApplications)
	cust.Get("/applications/:id", customer.GetApplication)
	cust.Put("/applications/:id", customer.UpdateApplication)
	cust.Post("/applications/:id/resubmit", customer.ResubmitApplication)
	cust.Post("/applications/:id/cancel", customer.CancelApplication)
	cust.Delete("/applications/:id", customer.DeleteApplication)

	// ── Reviewer routes (JWT required, REVIEWER role) ─────────────────────────
	rev := api.Group("/reviewer", middleware.JWTAuth(authSvc), middleware.RequireRole("REVIEWER"))
	rev.Get("/applications", reviewer.ListPendingReview)
	rev.Get("/applications/:id", reviewer.GetApplication)
	rev.Post("/applications/:id/action", reviewer.TakeAction)

	// ── Approver routes (JWT required, APPROVER role) ─────────────────────────
	appr := api.Group("/approver", middleware.JWTAuth(authSvc), middleware.RequireRole("APPROVER"))
	appr.Get("/applications", approver.ListPendingApproval)
	appr.Get("/applications/:id", approver.GetApplication)
	appr.Post("/applications/:id/action", approver.TakeAction)

	return app
}
