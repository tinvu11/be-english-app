package admin

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// @Summary     Administrator dashboard statistics
// @Description Live catalogue totals, daily registrations/uploads (Vietnam calendar days), and the five newest system videos. Totals are all-time; period metrics include today.
// @ID          admin-dashboard
// @Tags        admin
// @Produce     json
// @Param       days query int false "Calendar days including today" Enums(7,30) default(7)
// @Success     200 {object} entity.Dashboard
// @Failure     400,401,403,500,503 {object} errorDoc
// @Security    BearerAuth
// @Router      /admin/dashboard [get]
func (ctrl *controller) dashboard(ctx *fiber.Ctx) error {
	days, err := strconv.Atoi(ctx.Query("days", "7"))
	if err != nil || (days != 7 && days != 30) {
		return errorResponse(ctx, http.StatusBadRequest, "days must be 7 or 30")
	}
	if ctrl.dashboardUseCase == nil {
		return errorResponse(ctx, http.StatusServiceUnavailable, "dashboard unavailable")
	}
	queryCtx, cancel := context.WithTimeout(ctx.UserContext(), 10*time.Second)
	defer cancel()
	result, err := ctrl.dashboardUseCase.GetDashboard(queryCtx, days)
	if err != nil {
		ctrl.log.Error(err)
		return errorResponse(ctx, http.StatusInternalServerError, "could not load dashboard")
	}
	ctx.Set(fiber.HeaderCacheControl, "no-store")
	return ctx.JSON(result)
}
