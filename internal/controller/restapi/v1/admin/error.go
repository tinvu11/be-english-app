package admin

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/gofiber/fiber/v2"
)

// errorDoc keeps the shared API error type discoverable by Swagger.
type errorDoc = response.Error

func errorResponse(ctx *fiber.Ctx, status int, message string) error {
	return ctx.Status(status).JSON(response.Error{Error: message})
}
