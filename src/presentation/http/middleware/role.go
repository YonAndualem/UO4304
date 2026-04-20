// Package middleware contains Fiber middleware used across all route groups.
package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// RequireRole is a Fiber middleware that enforces role-based access control
// by inspecting the X-Role request header.
//
// In a production system this header would be derived from a validated JWT token
// rather than trusted from the client directly. The middleware is intentionally
// kept thin here so it can be swapped for a JWT-based implementation without
// changing any handler code.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("X-Role") != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden: role '" + role + "' required",
			})
		}
		return c.Next()
	}
}

// UserID extracts the caller's identity from the X-User-ID header.
// This is a placeholder for real authentication — in production the user ID
// would be extracted from a verified JWT claim.
func UserID(c *fiber.Ctx) string {
	return c.Get("X-User-ID")
}
