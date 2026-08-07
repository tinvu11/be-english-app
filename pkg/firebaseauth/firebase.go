// Package firebaseauth verifies Firebase Authentication ID tokens.
package firebaseauth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/evrone/go-clean-template/internal/entity"
	"google.golang.org/api/option"
)

// Verifier verifies Firebase ID tokens.
type Verifier struct {
	client *auth.Client
}

// New creates a Firebase ID-token verifier using a service-account file.
func New(ctx context.Context, projectID, credentialsFile string) (*Verifier, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("firebaseauth - New - firebase.NewApp: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebaseauth - New - app.Auth: %w", err)
	}

	return &Verifier{client: client}, nil
}

// Verify validates an ID token and returns the trusted Firebase identity.
func (v *Verifier) Verify(ctx context.Context, idToken string) (entity.AuthIdentity, error) {
	token, err := v.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return entity.AuthIdentity{}, fmt.Errorf("firebaseauth - Verify - client.VerifyIDToken: %w", err)
	}

	identity := entity.AuthIdentity{
		UID: token.UID,
	}

	if email, ok := token.Claims["email"].(string); ok {
		identity.Email = email
	}

	if name, ok := token.Claims["name"].(string); ok {
		identity.Name = name
	}

	if picture, ok := token.Claims["picture"].(string); ok {
		identity.Picture = picture
	}

	return identity, nil
}
