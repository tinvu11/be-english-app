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

		ctx.Locals("userID", localUser.ID)
		ctx.Locals("firebaseUID", identity.UID)

		return ctx.Next()
	}
}
