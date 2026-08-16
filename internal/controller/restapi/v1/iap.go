package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

func (r *V1) verifyIAPPurchase(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var input entity.VerifyPurchaseInput
	if err := ctx.BodyParser(&input); err != nil || r.v.Struct(input) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid purchase payload")
	}
	status, err := r.iap.VerifyPurchase(ctx.UserContext(), userID, input)
	if err != nil {
		return r.iapError(ctx, err)
	}
	return ctx.Status(http.StatusOK).JSON(status)
}

func (r *V1) getIAPStatus(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(string)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	status, err := r.iap.GetStatus(ctx.UserContext(), userID)
	if err != nil {
		return r.iapError(ctx, err)
	}
	return ctx.Status(http.StatusOK).JSON(status)
}

func (r *V1) appleIAPWebhook(ctx *fiber.Ctx) error {
	var body struct {
		SignedPayload string `json:"signedPayload"`
	}
	if err := ctx.BodyParser(&body); err != nil || body.SignedPayload == "" {
		return errorResponse(ctx, http.StatusBadRequest, "invalid Apple notification")
	}
	if err := r.iap.HandleAppleWebhook(ctx.UserContext(), body.SignedPayload); err != nil {
		return r.webhookError(ctx, err)
	}
	return ctx.SendStatus(http.StatusOK)
}

func (r *V1) googleIAPWebhook(ctx *fiber.Ctx) error {
	if err := r.iap.HandleGoogleWebhook(ctx.UserContext(), ctx.Body()); err != nil {
		return r.webhookError(ctx, err)
	}
	return ctx.SendStatus(http.StatusOK)
}

func (r *V1) iapError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidIAPPurchase), errors.Is(err, entity.ErrIAPProductMismatch):
		return errorResponse(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrIAPAlreadyOwned):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrIAPVerificationFailed):
		return errorResponse(ctx, http.StatusBadGateway, "store verification failed")
	default:
		r.l.Error(err, "restapi - v1 - iap")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}

func (r *V1) webhookError(ctx *fiber.Ctx, err error) error {
	if errors.Is(err, entity.ErrInvalidIAPPurchase) {
		return errorResponse(ctx, http.StatusBadRequest, "invalid store notification")
	}
	r.l.Error(err, "restapi - v1 - iap webhook")
	// Non-2xx makes Apple/PubSub retry transient failures.
	return errorResponse(ctx, http.StatusInternalServerError, "notification processing failed")
}
