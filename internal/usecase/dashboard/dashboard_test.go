package dashboard

import (
	"context"
	"errors"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	days int
	err  error
}

func (r *repositoryStub) GetDashboard(_ context.Context, days int) (entity.Dashboard, error) {
	r.days = days
	return entity.Dashboard{Days: days}, r.err
}

func TestGetDashboardValidatesPeriod(t *testing.T) {
	for _, days := range []int{-1, 0, 1, 8, 31, 999} {
		repository := &repositoryStub{}
		_, err := New(repository).GetDashboard(context.Background(), days)
		require.ErrorIs(t, err, ErrInvalidDays)
		require.Zero(t, repository.days)
	}
	for _, days := range []int{7, 30} {
		repository := &repositoryStub{}
		result, err := New(repository).GetDashboard(context.Background(), days)
		require.NoError(t, err)
		require.Equal(t, days, result.Days)
		require.Equal(t, days, repository.days)
	}
}

func TestGetDashboardPropagatesRepositoryFailure(t *testing.T) {
	expected := errors.New("database unavailable")
	_, err := New(&repositoryStub{err: expected}).GetDashboard(context.Background(), 7)
	require.ErrorIs(t, err, expected)
}
