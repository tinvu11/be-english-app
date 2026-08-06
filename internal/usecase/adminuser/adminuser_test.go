package adminuser

import (
	"context"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const actorID = "6fc48ba8-b5ca-46d9-87f1-d5e51f063763"
const targetID = "5f5cb6d2-0043-4b85-a085-8b94efc14ac3"

type repoStub struct {
	filter entity.UserFilter
	active bool
	role   string
}

func (r *repoStub) ListUsers(_ context.Context, f entity.UserFilter) (entity.UserList, error) {
	r.filter = f
	return entity.UserList{}, nil
}
func (r *repoStub) SetUserActive(_ context.Context, _ string, a bool) (entity.User, error) {
	r.active = a
	return entity.User{ID: targetID, IsActive: a}, nil
}
func (r *repoStub) SetUserRole(_ context.Context, _ string, role string) (entity.User, error) {
	r.role = role
	return entity.User{ID: targetID, Role: role}, nil
}

func TestListUsersNormalizesFilter(t *testing.T) {
	t.Parallel()
	repository := &repoStub{}
	uc := &UseCase{repo: repository}
	_, err := uc.ListUsers(t.Context(), entity.UserFilter{Search: " test ", Role: " ADMIN ", Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, "test", repository.filter.Search)
	assert.Equal(t, "admin", repository.filter.Role)
}

func TestAdminCannotLockSelf(t *testing.T) {
	t.Parallel()
	uc := &UseCase{repo: &repoStub{}}
	_, err := uc.SetUserActive(t.Context(), actorID, actorID, false)
	require.ErrorIs(t, err, entity.ErrAdminSelfMutation)
}

func TestSetUserRole(t *testing.T) {
	t.Parallel()
	repository := &repoStub{}
	uc := &UseCase{repo: repository}
	_, err := uc.SetUserRole(t.Context(), actorID, targetID, "ADMIN")
	require.NoError(t, err)
	assert.Equal(t, entity.RoleAdmin, repository.role)
}
