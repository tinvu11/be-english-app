package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/google/uuid"
)

// UseCase -.
type UseCase struct {
	repo repo.UserRepo
}

// New returns a User usecase instrumented with OpenTelemetry tracing spans.
func New(r repo.UserRepo) usecase.User {
	return newTraced(&UseCase{repo: r})
}

// Authenticate finds or provisions the local user for a verified Firebase identity.
func (uc *UseCase) Authenticate(ctx context.Context, identity entity.AuthIdentity) (entity.User, error) {
	if identity.UID == "" {
		return entity.User{}, fmt.Errorf("UserUseCase - Authenticate: empty Firebase UID")
	}

	existing, err := uc.repo.GetByFirebaseUID(ctx, identity.UID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, entity.ErrUserNotFound) {
		return entity.User{}, fmt.Errorf("UserUseCase - Authenticate - uc.repo.GetByFirebaseUID: %w", err)
	}

	now := time.Now().UTC()
	email := identity.Email
	if email == "" {
		email = identity.UID + "@firebase.local"
	}

	user := entity.User{
		ID:          uuid.New().String(),
		FirebaseUID: identity.UID,
		Username:    firebaseUsername(identity),
		Email:       email,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err = uc.repo.Store(ctx, &user); err != nil {
		if errors.Is(err, entity.ErrUserAlreadyExists) {
			concurrentUser, lookupErr := uc.repo.GetByFirebaseUID(ctx, identity.UID)
			if lookupErr == nil {
				return concurrentUser, nil
			}
		}

		return entity.User{}, fmt.Errorf("UserUseCase - Authenticate - uc.repo.Store: %w", err)
	}

	return user, nil
}

// Register is retained only for unused legacy transports.
func (uc *UseCase) Register(context.Context, string, string, string) (entity.User, error) {
	return entity.User{}, entity.ErrLocalAuthDisabled
}

// Login is retained only for unused legacy transports.
func (uc *UseCase) Login(context.Context, string, string) (string, error) {
	return "", entity.ErrLocalAuthDisabled
}

// GetUser -.
func (uc *UseCase) GetUser(ctx context.Context, userID string) (entity.User, error) {
	user, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserUseCase - GetUser - uc.repo.GetByID: %w", err)
	}

	return user, nil
}

func firebaseUsername(identity entity.AuthIdentity) string {
	base := identity.Name
	if base == "" {
		base, _, _ = strings.Cut(identity.Email, "@")
	}

	base = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return '-'
	}, base)
	base = strings.Trim(base, "-")
	if base == "" {
		base = "user"
	}
	if len(base) > 32 {
		base = base[:32]
	}

	suffix := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, identity.UID)
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}

	return base + "-" + suffix
}
