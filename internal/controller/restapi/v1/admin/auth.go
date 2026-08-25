package admin

import (
	"errors"
	"net/http"

	adminrequest "github.com/evrone/go-clean-template/internal/controller/restapi/v1/admin/request"
	adminresponse "github.com/evrone/go-clean-template/internal/controller/restapi/v1/admin/response"
	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary     Admin login
// @Description Verify a Firebase ID token and require the local user to have the admin role
// @ID          admin-login
// @Tags        admin
// @Accept      json
// @Produce     json
// @Param       request body adminrequest.Login true "Firebase credentials"
// @Success     200 {object} adminresponse.Login
// @Failure     400 {object} errorDoc
// @Failure     401 {object} errorDoc
// @Failure     403 {object} errorDoc
// @Failure     500 {object} errorDoc
// @Router      /admin/auth/login [post]
func (ctrl *controller) login(ctx *fiber.Ctx) error {
	var body adminrequest.Login
	if err := ctx.BodyParser(&body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	if err := ctrl.validate.Struct(body); err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "idToken is required")
	}

	identity, err := ctrl.verifier.Verify(ctx.UserContext(), body.IDToken)
	if err != nil {
		return errorResponse(ctx, http.StatusUnauthorized, "invalid or expired Firebase ID token")
	}

	user, err := ctrl.users.Authenticate(ctx.UserContext(), identity)
	if err != nil {
		ctrl.log.Error(err, "restapi - v1 - admin - login")

		return errorResponse(ctx, http.StatusInternalServerError, "failed to authenticate admin")
	}
	if user.Role != entity.RoleAdmin {
		return errorResponse(ctx, http.StatusForbidden, "forbidden: admin access required")
	}

	return ctx.Status(http.StatusOK).JSON(adminresponse.Login{Authenticated: true, User: user})
}

// @Summary     Get current admin
// @Description Return the authenticated administrator profile
// @ID          admin-me
// @Tags        admin
// @Produce     json
// @Success     200 {object} entity.User
// @Failure     401 {object} errorDoc
// @Failure     403 {object} errorDoc
// @Failure     404 {object} errorDoc
// @Failure     500 {object} errorDoc
// @Security    BearerAuth
// @Router      /admin/me [get]
func (ctrl *controller) me(ctx *fiber.Ctx) error {
	authenticated, ok := ctx.Locals("user").(entity.User)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}

	user, err := ctrl.users.GetUser(ctx.UserContext(), authenticated.ID)
	if err != nil {
		ctrl.log.Error(err, "restapi - v1 - admin - me")
		if errors.Is(err, entity.ErrUserNotFound) {
			return errorResponse(ctx, http.StatusNotFound, "admin not found")
		}

		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
	if user.Role != entity.RoleAdmin {
		return errorResponse(ctx, http.StatusForbidden, "forbidden: admin access required")
	}

	return ctx.Status(http.StatusOK).JSON(user)
}
