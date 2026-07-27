package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
)

const _bearerParts = 2

type errorResponse struct {
	Error string `json:"error"`
}

// TokenVerifier verifies a Firebase ID token.
type TokenVerifier interface {
	Verify(ctx context.Context, idToken string) (entity.AuthIdentity, error)
}

// UserAuthenticator maps a verified Firebase identity to a local user.
type UserAuthenticator interface {
	Authenticate(ctx context.Context, identity entity.AuthIdentity) (entity.User, error)
}

type contextKey string

const UserKey contextKey = "user"

// GetUser extracts the user entity from context.
func GetUser(ctx context.Context) (entity.User, bool) {
	val := ctx.Value(UserKey)
	user, ok := val.(entity.User)

	return user, ok
}

// Auth verifies Firebase ID tokens and places the local user ID in Fiber locals.
func Auth(verifier TokenVerifier, users UserAuthenticator) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		header := ctx.Get("Authorization")
		if header == "" {
			return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "missing authorization header"})
		}

		parts := strings.SplitN(header, " ", _bearerParts)
		if len(parts) != _bearerParts || parts[0] != "Bearer" {
			return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "invalid authorization header format"})
		}

		identity, err := verifier.Verify(ctx.UserContext(), parts[1])
		if err != nil {
			return ctx.Status(http.StatusUnauthorized).JSON(errorResponse{Error: "invalid or expired Firebase ID token"})
		}

		localUser, err := users.Authenticate(ctx.UserContext(), identity)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).JSON(errorResponse{Error: "failed to provision user"})
		}

		ctx.Locals("user", localUser)
		ctx.Locals("userID", localUser.ID)
		ctx.Locals("firebaseUID", identity.UID)

		userCtx := context.WithValue(ctx.UserContext(), UserKey, localUser)
		ctx.SetUserContext(userCtx)

		return ctx.Next()
	}
}

// RequireRole checks if the authenticated user has the specified role.
func RequireRole(role string) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		userVal := ctx.Locals("user")

		var (
			localUser entity.User
			ok        bool
		)

		if userVal != nil {
			localUser, ok = userVal.(entity.User)
		} else {
			localUser, ok = GetUser(ctx.UserContext())
		}

		if !ok || localUser.Role != role {
			errResp := errorResponse{Error: "forbidden: access denied"}

			return ctx.Status(http.StatusForbidden).JSON(errResp)
		}

		return ctx.Next()
	}
}

// AdminOnly checks if the authenticated user has the 'admin' role.
func AdminOnly() func(*fiber.Ctx) error {
	return RequireRole("admin")
}
