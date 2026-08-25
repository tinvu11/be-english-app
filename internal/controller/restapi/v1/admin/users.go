package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	adminrequest "github.com/evrone/go-clean-template/internal/controller/restapi/v1/admin/request"
	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// @Summary List application users
// @Tags admin-users
// @Produce json
// @Param search query string false "Email or username"
// @Param role query string false "Role" Enums(user,admin)
// @Param is_active query bool false "Active status"
// @Param limit query int false "Page size" default(20) maximum(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} entity.UserList
// @Failure 400,401,403,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/users [get]
func (ctrl *controller) listUsers(ctx *fiber.Ctx) error {
	filter := entity.UserFilter{Search: ctx.Query("search"), Role: ctx.Query("role"), Limit: 20}
	var err error
	if raw := ctx.Query("limit"); raw != "" {
		filter.Limit, err = strconv.Atoi(raw)
		if err != nil {
			return errorResponse(ctx, http.StatusBadRequest, "invalid pagination")
		}
	}
	if raw := ctx.Query("offset"); raw != "" {
		filter.Offset, err = strconv.Atoi(raw)
		if err != nil {
			return errorResponse(ctx, http.StatusBadRequest, "invalid pagination")
		}
	}
	if raw := ctx.Query("is_active"); raw != "" {
		value, parseErr := strconv.ParseBool(raw)
		if parseErr != nil {
			return errorResponse(ctx, http.StatusBadRequest, "invalid is_active")
		}
		filter.IsActive = &value
	}
	result, err := ctrl.adminUsers.ListUsers(ctx.UserContext(), filter)
	if err != nil {
		return ctrl.userAdminError(ctx, err)
	}
	return ctx.JSON(result)
}

// @Summary Lock or unlock user
// @Tags admin-users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param request body adminrequest.UserStatus true "Account status"
// @Success 200 {object} entity.User
// @Failure 400,401,403,404,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/users/{id}/status [patch]
func (ctrl *controller) setUserActive(ctx *fiber.Ctx) error {
	actor, ok := ctx.Locals("user").(entity.User)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var body adminrequest.UserStatus
	if err := ctx.BodyParser(&body); err != nil || ctrl.validate.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid user status")
	}
	user, err := ctrl.adminUsers.SetUserActive(ctx.UserContext(), actor.ID, ctx.Params("id"), *body.IsActive)
	if err != nil {
		return ctrl.userAdminError(ctx, err)
	}
	return ctx.JSON(user)
}

// @Summary Promote or demote user
// @Tags admin-users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param request body adminrequest.UserRole true "User role"
// @Success 200 {object} entity.User
// @Failure 400,401,403,404,500 {object} errorDoc
// @Security BearerAuth
// @Router /admin/users/{id}/role [patch]
func (ctrl *controller) setUserRole(ctx *fiber.Ctx) error {
	actor, ok := ctx.Locals("user").(entity.User)
	if !ok {
		return errorResponse(ctx, http.StatusUnauthorized, "unauthorized")
	}
	var body adminrequest.UserRole
	if err := ctx.BodyParser(&body); err != nil || ctrl.validate.Struct(body) != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid user role")
	}
	user, err := ctrl.adminUsers.SetUserRole(ctx.UserContext(), actor.ID, ctx.Params("id"), strings.ToLower(body.Role))
	if err != nil {
		return ctrl.userAdminError(ctx, err)
	}
	return ctx.JSON(user)
}

func (ctrl *controller) userAdminError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidUserFilter), errors.Is(err, entity.ErrInvalidUserRole):
		return errorResponse(ctx, http.StatusBadRequest, "invalid user request")
	case errors.Is(err, entity.ErrUserNotFound):
		return errorResponse(ctx, http.StatusNotFound, "user not found")
	case errors.Is(err, entity.ErrAdminSelfMutation):
		return errorResponse(ctx, http.StatusConflict, "admin cannot lock or demote own account")
	default:
		ctrl.log.Error(err, "restapi - v1 - admin - user")
		return errorResponse(ctx, http.StatusInternalServerError, "internal server error")
	}
}
