package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/enterprise/trade-license/src/application/auth"
)

// AuthHandler exposes the register and login endpoints.
type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		UserID   string `json:"user_id"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.UserID == "" || req.Password == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id, password, and role are required"})
	}
	if req.Role != "CUSTOMER" && req.Role != "REVIEWER" && req.Role != "APPROVER" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role must be CUSTOMER, REVIEWER, or APPROVER"})
	}
	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password must be at least 6 characters"})
	}

	if err := h.svc.Register(req.UserID, req.Password, req.Role); err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "user ID already taken"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "registration failed"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user_id": req.UserID, "role": req.Role})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		UserID   string `json:"user_id"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	token, role, err := h.svc.Login(req.UserID, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidPassword) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user ID or password"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "login failed"})
	}

	return c.JSON(fiber.Map{
		"token":   token,
		"user_id": req.UserID,
		"role":    role,
	})
}
