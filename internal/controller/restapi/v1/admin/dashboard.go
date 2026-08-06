package admin

import (
	"net/http"

	_ "github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/gofiber/fiber/v2"
)

type dashboardResponse struct {
	Message string `json:"message" example:"Welcome to the admin dashboard!"`
}

// @Summary     Admin dashboard
// @Description Get the administrator dashboard message
// @ID          admin-dashboard
// @Tags        admin
// @Produce     json
// @Success     200 {object} dashboardResponse
// @Failure     401 {object} response.Error
// @Failure     403 {object} response.Error
// @Security    BearerAuth
// @Router      /admin/dashboard [get]
func (ctrl *controller) dashboard(ctx *fiber.Ctx) error {
	return ctx.Status(http.StatusOK).JSON(dashboardResponse{Message: "Welcome to the admin dashboard!"})
}
