package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// quotaErrorDoc keeps the Swagger response type discoverable across v1 handlers.
type quotaErrorDoc = response.QuotaError

func errorResponse(ctx *fiber.Ctx, code int, msg string) error {
	return ctx.Status(code).JSON(response.Error{Error: msg})
}

func quotaErrorResponse(ctx *fiber.Ctx, err error) error {
	var quotaErr *entity.QuotaExceededError
	if !errors.As(err, &quotaErr) {
		return errorResponse(ctx, http.StatusTooManyRequests, entity.ErrQuotaExceeded.Error())
	}

	status := quotaErr.Status

	return ctx.Status(http.StatusTooManyRequests).JSON(response.QuotaError{
		Error: entity.ErrQuotaExceeded.Error(), Code: "FEATURE_QUOTA_EXCEEDED", Feature: status.Feature,
		Limit: status.Limit, Used: status.Used, Remaining: status.Remaining, ResetAt: status.ResetAt, UpgradeRequired: true,
	})
}
