package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/internal/usecase/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var errUserRepo = errors.New("repository error")

func newUserUseCase(t *testing.T) (usecase.User, *MockUserRepo, *MockLanguageRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := NewMockUserRepo(ctrl)
	languages := NewMockLanguageRepo(ctrl)

	return user.New(repo, languages), repo, languages
}

func TestAuthenticateExistingFirebaseUser(t *testing.T) {
	t.Parallel()

	uc, repo, _ := newUserUseCase(t)
	identity := entity.AuthIdentity{UID: "firebase-123", Email: "test@example.com", Name: "Test User", Picture: "https://example.com/avatar.jpg"}
	expected := entity.User{ID: "local-123", FirebaseUID: identity.UID, Email: identity.Email, Username: identity.Name, AvatarURL: identity.Picture}

	repo.EXPECT().GetByFirebaseUID(gomock.Any(), identity.UID).Return(expected, nil)
	repo.EXPECT().UpdateFirebaseProfile(gomock.Any(), expected.ID, identity.Name, identity.Picture).Return(nil)

	result, err := uc.Authenticate(context.Background(), identity)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestAuthenticateProvisionsFirebaseUser(t *testing.T) {
	t.Parallel()

	uc, repo, _ := newUserUseCase(t)
	identity := entity.AuthIdentity{UID: "firebase-123456789", Email: "john@example.com", Name: "John Doe", Picture: "https://example.com/john.jpg"}

	repo.EXPECT().GetByFirebaseUID(gomock.Any(), identity.UID).Return(entity.User{}, entity.ErrUserNotFound)
	repo.EXPECT().Store(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, stored *entity.User) error {
		assert.NotEmpty(t, stored.ID)
		assert.Equal(t, identity.UID, stored.FirebaseUID)
		assert.Equal(t, identity.Email, stored.Email)
		assert.Equal(t, "John Doe", stored.Username)
		assert.Equal(t, identity.Picture, stored.AvatarURL)
		assert.Equal(t, entity.RoleUser, stored.Role)
		return nil
	})

	result, err := uc.Authenticate(context.Background(), identity)

	require.NoError(t, err)
	assert.Equal(t, identity.UID, result.FirebaseUID)
	assert.Equal(t, entity.RoleUser, result.Role)
}

func TestAuthenticateRejectsEmptyUID(t *testing.T) {
	t.Parallel()

	uc, _, _ := newUserUseCase(t)

	_, err := uc.Authenticate(context.Background(), entity.AuthIdentity{})

	require.Error(t, err)
}

func TestAuthenticateRepositoryError(t *testing.T) {
	t.Parallel()

	uc, repo, _ := newUserUseCase(t)
	repo.EXPECT().GetByFirebaseUID(gomock.Any(), "firebase-123").Return(entity.User{}, errUserRepo)

	_, err := uc.Authenticate(context.Background(), entity.AuthIdentity{UID: "firebase-123"})

	require.ErrorIs(t, err, errUserRepo)
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	uc, repo, _ := newUserUseCase(t)
	expected := entity.User{ID: "local-123", FirebaseUID: "firebase-123"}
	repo.EXPECT().GetByID(gomock.Any(), expected.ID).Return(expected, nil)

	result, err := uc.GetUser(context.Background(), expected.ID)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestLocalAuthenticationDisabled(t *testing.T) {
	t.Parallel()

	uc, _, _ := newUserUseCase(t)

	_, registerErr := uc.Register(context.Background(), "name", "email@example.com", "password")
	_, loginErr := uc.Login(context.Background(), "email@example.com", "password")

	require.ErrorIs(t, registerErr, entity.ErrLocalAuthDisabled)
	require.ErrorIs(t, loginErr, entity.ErrLocalAuthDisabled)
}

func TestUpdateLanguages(t *testing.T) {
	t.Parallel()

	uc, users, languages := newUserUseCase(t)
	nativeID, targetID := 2, 1
	expected := entity.User{ID: "local-123", NativeLanguageID: &nativeID, TargetLanguageID: &targetID}
	languages.EXPECT().GetLanguage(gomock.Any(), nativeID).Return(entity.Language{ID: nativeID, IsActive: true, IsLearnable: false}, nil)
	languages.EXPECT().GetLanguage(gomock.Any(), targetID).Return(entity.Language{ID: targetID, IsActive: true, IsLearnable: true}, nil)
	users.EXPECT().UpdateLanguages(gomock.Any(), expected.ID, nativeID, targetID).Return(nil)
	users.EXPECT().GetByID(gomock.Any(), expected.ID).Return(expected, nil)

	got, err := uc.UpdateLanguages(context.Background(), expected.ID, nativeID, targetID)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestUpdateLanguagesRejectsWrongKinds(t *testing.T) {
	t.Parallel()

	uc, _, languages := newUserUseCase(t)
	languages.EXPECT().GetLanguage(gomock.Any(), 1).Return(entity.Language{ID: 1, IsActive: true, IsLearnable: true}, nil)
	languages.EXPECT().GetLanguage(gomock.Any(), 2).Return(entity.Language{ID: 2, IsActive: true, IsLearnable: true}, nil)

	_, err := uc.UpdateLanguages(context.Background(), "local-123", 1, 2)

	require.ErrorIs(t, err, entity.ErrInvalidLanguage)
}
