package v1

import (
	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

type adminDashboardResponse struct {
	Message string `json:"message" example:"Welcome to the admin dashboard!"`
}

// @Summary     Admin Dashboard
// @Description Get admin dashboard message
// @ID          admin-dashboard
// @Tags        admin
// @Produce     json
// @Success     200 {object} adminDashboardResponse
// @Failure     401 {object} response.Error
// @Failure     403 {object} response.Error
// @Failure     500 {object} response.Error
// @Security    BearerAuth
// @Router      /admin/dashboard [get]
func (r *V1) adminDashboard(ctx *fiber.Ctx) error {
	resp := adminDashboardResponse{
		Message: "Welcome to the admin dashboard!",
	}

	return ctx.Status(http.StatusOK).JSON(resp)
}
