// Package middleware contains Fiber middleware used across all route groups.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/enterprise/trade-license/src/application/auth"
)

// JWTAuth validates the Bearer token from the Authorization header and stores
// the parsed user ID and role in fiber locals for downstream handlers to read.
func JWTAuth(svc *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := ""
		if header := c.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
			token = strings.TrimPrefix(header, "Bearer ")
		} else {
			// Fall back to ?token= query param so browsers can load documents
			// directly in <embed>/<iframe> without custom headers.
			token = c.Query("token")
		}
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing or invalid Authorization header",
			})
		}
		claims, err := svc.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}
		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

// RequireRole enforces that the JWT-authenticated user has the expected role.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Locals("role") != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden: role '" + role + "' required",
			})
		}
		return c.Next()
	}
}

// UserID returns the authenticated user's ID from fiber locals.
func UserID(c *fiber.Ctx) string {
	if uid, ok := c.Locals("userID").(string); ok {
		return uid
	}
	return ""
}
