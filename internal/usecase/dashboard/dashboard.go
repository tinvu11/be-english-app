// Package dashboard provides administrator reporting.
package dashboard

import (
	"context"
	"errors"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
)

var ErrInvalidDays = errors.New("dashboard days must be 7 or 30")

type UseCase struct{ repo repo.DashboardRepo }

func New(repository repo.DashboardRepo) usecase.Dashboard {
	return &UseCase{repo: repository}
}

func (uc *UseCase) GetDashboard(ctx context.Context, days int) (entity.Dashboard, error) {
	if days != 7 && days != 30 {
		return entity.Dashboard{}, ErrInvalidDays
	}
	return uc.repo.GetDashboard(ctx, days)
}
