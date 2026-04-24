package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
)

// domainError maps a domain or application error to the correct HTTP status code
// and returns a consistent JSON error body.
//
// Mapping rationale:
//   - 404: the requested resource does not exist
//   - 422: the request is well-formed but violates a business rule (invalid transition, missing docs, etc.)
//   - 400: the caller provided structurally invalid input
//   - 500: unexpected infrastructure failure (DB, network, etc.)
func domainError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domainerrors.ErrApplicationNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})

	case errors.Is(err, domainerrors.ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})

	case errors.Is(err, domainerrors.ErrInvalidStatusTransition),
		errors.Is(err, domainerrors.ErrDocumentRequired),
		errors.Is(err, domainerrors.ErrPaymentRequired):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})

	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
