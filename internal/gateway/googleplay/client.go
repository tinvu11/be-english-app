package googleplay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

var errCredentialsRequired = errors.New("google play gateway: credentials file is required")

type Config struct {
	CredentialsFile string
}

type Client struct{ service *androidpublisher.Service }

func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.CredentialsFile == "" {
		return nil, errCredentialsRequired
	}

	service, err := androidpublisher.NewService(ctx, option.WithAuthCredentialsFile(option.ServiceAccount, cfg.CredentialsFile),
		option.WithScopes(androidpublisher.AndroidpublisherScope))
	if err != nil {
		return nil, fmt.Errorf("google play gateway - initialize: %w", err)
	}

	return &Client{service: service}, nil
}

//nolint:funlen,gocognit,gocyclo,cyclop // Normalization and token traversal are one verification operation.
func (c *Client) VerifySubscription(ctx context.Context, packageName, purchaseToken string) (entity.StoreVerificationResult, error) {
	purchase, err := c.service.Purchases.Subscriptionsv2.Get(packageName, purchaseToken).Context(ctx).Do()
	if err != nil {
		return entity.StoreVerificationResult{}, fmt.Errorf("%w: google play request: %w", entity.ErrIAPVerificationFailed, err)
	}

	if len(purchase.LineItems) == 0 {
		return entity.StoreVerificationResult{}, entity.ErrInvalidIAPPurchase
	}

	start, err := time.Parse(time.RFC3339Nano, purchase.StartTime)
	if err != nil {
		return entity.StoreVerificationResult{}, fmt.Errorf("%w: invalid google startTime", entity.ErrIAPVerificationFailed)
	}

	line := purchase.LineItems[0]

	expires, err := time.Parse(time.RFC3339Nano, line.ExpiryTime)
	if err != nil {
		return entity.StoreVerificationResult{}, fmt.Errorf("%w: invalid google expiryTime", entity.ErrIAPVerificationFailed)
	}

	for _, candidate := range purchase.LineItems[1:] {
		candidateExpiry, parseErr := time.Parse(time.RFC3339Nano, candidate.ExpiryTime)
		if parseErr == nil && candidateExpiry.After(expires) {
			line, expires = candidate, candidateExpiry
		}
	}

	transactionID := line.LatestSuccessfulOrderId
	if transactionID == "" {
		transactionID = purchase.LatestOrderId
	}

	if transactionID == "" {
		transactionID = purchaseToken
	}

	originalID := purchaseToken
	linkedToken := purchase.LinkedPurchaseToken
	// Upgrade/downgrade tokens form a chain. Walk it to the root so ownership
	// checks remain stable across more than one plan replacement.
	for depth := 0; linkedToken != "" && depth < 20; depth++ {
		originalID = linkedToken

		linked, linkedErr := c.service.Purchases.Subscriptionsv2.Get(packageName, linkedToken).Context(ctx).Do()
		if linkedErr != nil {
			return entity.StoreVerificationResult{}, fmt.Errorf("%w: resolve linked google purchase: %w", entity.ErrIAPVerificationFailed, linkedErr)
		}

		linkedToken = linked.LinkedPurchaseToken
	}

	if linkedToken != "" {
		return entity.StoreVerificationResult{}, fmt.Errorf("%w: google purchase chain is too deep", entity.ErrIAPVerificationFailed)
	}

	autoRenew := line.AutoRenewingPlan != nil && line.AutoRenewingPlan.AutoRenewEnabled
	status := googleStatus(purchase.SubscriptionState, expires, autoRenew)

	raw, err := json.Marshal(purchase)
	if err != nil {
		return entity.StoreVerificationResult{}, fmt.Errorf("google play gateway - marshal response: %w", err)
	}

	return entity.StoreVerificationResult{
		Platform: entity.PlatformAndroid, ProductID: line.ProductId,
		TransactionID: transactionID, OriginalTransactionID: originalID, PurchaseTime: start.UTC(),
		ExpiresTime: expires.UTC(), Status: status, AutoRenew: autoRenew,
		Acknowledged: purchase.AcknowledgementState == "ACKNOWLEDGEMENT_STATE_ACKNOWLEDGED",
		EventTime:    time.Now().UTC(), RawPayload: raw,
	}, nil
}

func (c *Client) Acknowledge(ctx context.Context, packageName, productID, purchaseToken string) error {
	request := &androidpublisher.SubscriptionPurchasesAcknowledgeRequest{}
	if err := c.service.Purchases.Subscriptions.Acknowledge(packageName, productID, purchaseToken, request).Context(ctx).Do(); err != nil {
		return fmt.Errorf("%w: acknowledge google purchase: %w", entity.ErrIAPVerificationFailed, err)
	}

	return nil
}

func googleStatus(state string, expires time.Time, autoRenew bool) entity.SubscriptionStatus {
	switch state {
	case "SUBSCRIPTION_STATE_ACTIVE":
		return entity.SubscriptionActive
	case "SUBSCRIPTION_STATE_IN_GRACE_PERIOD":
		return entity.SubscriptionInGracePeriod
	case "SUBSCRIPTION_STATE_CANCELED":
		if expires.After(time.Now()) {
			return entity.SubscriptionActive
		}

		return entity.SubscriptionExpired
	case "SUBSCRIPTION_STATE_EXPIRED":
		return entity.SubscriptionExpired
	default:
		_ = autoRenew

		return entity.SubscriptionCancelled
	}
}
