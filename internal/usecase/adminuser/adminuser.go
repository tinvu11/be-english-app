// Package adminuser implements administrator-facing user business rules.
package adminuser

import (
	"context"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/google/uuid"
)

const maxPageSize = 100

type UseCase struct{ repo repo.AdminUserRepo }

func New(repository repo.AdminUserRepo) usecase.AdminUser {
	return newTraced(&UseCase{repo: repository})
}

func (uc *UseCase) ListUsers(ctx context.Context, filter entity.UserFilter) (entity.UserList, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Role = strings.ToLower(strings.TrimSpace(filter.Role))
	if len(filter.Search) > 255 || (filter.Role != "" && !validRole(filter.Role)) ||
		filter.Limit <= 0 || filter.Limit > maxPageSize || filter.Offset < 0 {
		return entity.UserList{}, entity.ErrInvalidUserFilter
	}
	return uc.repo.ListUsers(ctx, filter)
}

func (uc *UseCase) SetUserActive(ctx context.Context, actorID, userID string, isActive bool) (entity.User, error) {
	if !validUserIDs(actorID, userID) {
		return entity.User{}, entity.ErrInvalidUserFilter
	}
	if actorID == userID && !isActive {
		return entity.User{}, entity.ErrAdminSelfMutation
	}
	return uc.repo.SetUserActive(ctx, userID, isActive)
}

func (uc *UseCase) SetUserRole(ctx context.Context, actorID, userID, role string) (entity.User, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if !validUserIDs(actorID, userID) || !validRole(role) {
		return entity.User{}, entity.ErrInvalidUserRole
	}
	if actorID == userID && role != entity.RoleAdmin {
		return entity.User{}, entity.ErrAdminSelfMutation
	}
	return uc.repo.SetUserRole(ctx, userID, role)
}

func validRole(role string) bool { return role == entity.RoleUser || role == entity.RoleAdmin }

func validUserIDs(ids ...string) bool {
	for _, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			return false
		}
	}
	return true
}
