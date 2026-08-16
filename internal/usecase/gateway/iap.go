package gateway

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

type AppleGateway interface {
	VerifyTransaction(ctx context.Context, transactionID string) (entity.StoreVerificationResult, error)
	ParseNotification(ctx context.Context, signedPayload string) (entity.WebhookEvent, error)
}

type GoogleGateway interface {
	VerifySubscription(ctx context.Context, packageName, purchaseToken string) (entity.StoreVerificationResult, error)
	Acknowledge(ctx context.Context, packageName, productID, purchaseToken string) error
}
