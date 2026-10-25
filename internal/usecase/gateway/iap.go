package gateway

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

type GoogleGateway interface {
	VerifySubscription(ctx context.Context, packageName, purchaseToken string) (entity.StoreVerificationResult, error)
	Acknowledge(ctx context.Context, packageName, productID, purchaseToken string) error
}
