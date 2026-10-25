package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/request"
	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary Verify an in-app purchase
// @Description Verify an App Store or Google Play purchase and update the current user's subscription
// @ID iap-verify-purchase
// @Tags iap
// @Accept json
// @Produce json
// @Param request body entity.VerifyPurchaseInput true "Purchase details"
// @Success 200 {object} entity.IAPStatus
// @Failure 400,401,409,500,502,503 {object} response.Error
// @Security BearerAuth
// @Router /iap/verify [post]
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

// @Summary Get in-app purchase status
// @Description Get the current user's latest subscription and premium entitlement
// @ID iap-get-status
// @Tags iap
// @Produce json
// @Success 200 {object} entity.IAPStatus
// @Failure 401,500 {object} response.Error
// @Security BearerAuth
// @Router /iap/status [get]
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

// @Summary Receive an Apple IAP notification
// @Description Process an App Store Server Notifications V2 signed payload. This public endpoint authenticates the notification by verifying its JWS signature.
// @ID iap-apple-webhook
// @Tags iap-webhooks
// @Accept json
// @Produce json
// @Param request body request.AppleIAPWebhook true "App Store notification"
// @Success 200 {string} string "OK"
// @Failure 400,500 {object} response.Error
// @Router /iap/webhooks/apple [post]
func (r *V1) appleIAPWebhook(ctx *fiber.Ctx) error {
	var body request.AppleIAPWebhook
	if err := ctx.BodyParser(&body); err != nil || body.SignedPayload == "" {
		return errorResponse(ctx, http.StatusBadRequest, "invalid Apple notification")
	}
	if err := r.iap.HandleAppleWebhook(ctx.UserContext(), body.SignedPayload); err != nil {
		return r.webhookError(ctx, err)
	}
	return ctx.SendStatus(http.StatusOK)
}

// @Summary Receive a Google Play RTDN notification
// @Description Process a Google Cloud Pub/Sub push envelope containing a Base64-encoded Real-time Developer Notification. This is a public store callback.
// @ID iap-google-webhook
// @Tags iap-webhooks
// @Accept json
// @Produce json
// @Param request body request.GoogleIAPWebhook true "Google Pub/Sub push envelope"
// @Success 200 {string} string "OK"
// @Failure 400,500 {object} response.Error
// @Router /iap/webhooks/google [post]
func (r *V1) googleIAPWebhook(ctx *fiber.Ctx) error {
	if err := r.iap.HandleGoogleWebhook(ctx.UserContext(), ctx.Body()); err != nil {
		return r.webhookError(ctx, err)
	}
	return ctx.SendStatus(http.StatusOK)
}

// Keep the Swagger response package discoverable when annotations are parsed.
var _ response.Error

func (r *V1) iapError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidIAPPurchase), errors.Is(err, entity.ErrIAPProductMismatch):
		return errorResponse(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrIAPAlreadyOwned):
		return errorResponse(ctx, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrIAPVerificationFailed):
		return errorResponse(ctx, http.StatusBadGateway, "store verification failed")
	case errors.Is(err, entity.ErrIAPPlatformDisabled):
		return errorResponse(ctx, http.StatusServiceUnavailable, err.Error())
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
