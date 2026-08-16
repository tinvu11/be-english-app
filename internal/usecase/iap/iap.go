package iap

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	iapgateway "github.com/evrone/go-clean-template/internal/usecase/gateway"
	iaprepo "github.com/evrone/go-clean-template/internal/usecase/repo"
)

type Config struct {
	GooglePackageName string
	RestorePolicy     string // "block" (default) or "transfer"
}

type UseCase struct {
	apple  iapgateway.AppleGateway
	google iapgateway.GoogleGateway
	uow    iaprepo.UnitOfWork
	repo   iaprepo.SubscriptionRepo
	cfg    Config
	now    func() time.Time
}

func New(apple iapgateway.AppleGateway, google iapgateway.GoogleGateway, uow iaprepo.UnitOfWork,
	repo iaprepo.SubscriptionRepo, cfg Config,
) *UseCase {
	return &UseCase{apple: apple, google: google, uow: uow, repo: repo, cfg: cfg, now: time.Now}
}

//nolint:gocyclo,cyclop // Store-specific verification and acknowledgement branches are intentionally explicit.
func (uc *UseCase) VerifyPurchase(ctx context.Context, userID string, input entity.VerifyPurchaseInput) (entity.IAPStatus, error) {
	if userID == "" || !input.Platform.Valid() || strings.TrimSpace(input.ProductID) == "" || strings.TrimSpace(input.PurchaseToken) == "" {
		return entity.IAPStatus{}, entity.ErrInvalidIAPPurchase
	}

	var (
		result entity.StoreVerificationResult
		err    error
	)

	switch input.Platform {
	case entity.PlatformIOS:
		result, err = uc.apple.VerifyTransaction(ctx, input.PurchaseToken)
	case entity.PlatformAndroid:
		if input.PackageName == "" || input.PackageName != uc.cfg.GooglePackageName {
			return entity.IAPStatus{}, entity.ErrInvalidIAPPurchase
		}

		result, err = uc.google.VerifySubscription(ctx, input.PackageName, input.PurchaseToken)
	}

	if err != nil {
		return entity.IAPStatus{}, err
	}

	if result.Platform != input.Platform || result.ProductID != input.ProductID {
		return entity.IAPStatus{}, entity.ErrIAPProductMismatch
	}

	if persistErr := uc.persistResult(ctx, userID, &result, "", ""); persistErr != nil {
		return entity.IAPStatus{}, persistErr
	}

	if input.Platform == entity.PlatformAndroid && !result.Acknowledged {
		if acknowledgeErr := uc.google.Acknowledge(ctx, input.PackageName, result.ProductID, input.PurchaseToken); acknowledgeErr != nil {
			return entity.IAPStatus{}, acknowledgeErr
		}
	}

	return uc.GetStatus(ctx, userID)
}

func (uc *UseCase) GetStatus(ctx context.Context, userID string) (entity.IAPStatus, error) {
	if userID == "" {
		return entity.IAPStatus{}, entity.ErrInvalidIAPPurchase
	}

	sub, err := uc.repo.GetActiveByUserID(ctx, userID)
	if errors.Is(err, entity.ErrSubscriptionNotFound) {
		return entity.IAPStatus{IsPremium: false}, nil
	}

	if err != nil {
		return entity.IAPStatus{}, err
	}

	isPremium := sub.Status.Entitled() && sub.ExpiresAt.After(uc.now())

	status := entity.IAPStatus{IsPremium: isPremium, Subscription: sub}
	if isPremium {
		until := sub.ExpiresAt
		status.PremiumUntil = &until
	}

	return status, nil
}

func (uc *UseCase) HandleAppleWebhook(ctx context.Context, signedPayload string) error {
	if signedPayload == "" {
		return entity.ErrInvalidIAPPurchase
	}

	event, err := uc.apple.ParseNotification(ctx, signedPayload)
	if err != nil {
		return err
	}

	return uc.persistWebhook(ctx, &event)
}

type pubSubEnvelope struct {
	Message struct {
		Data      string `json:"data"`
		MessageID string `json:"messageId"`
	} `json:"message"`
}

type developerNotification struct {
	PackageName              string          `json:"packageName"`
	EventTimeMillis          string          `json:"eventTimeMillis"`
	TestNotification         json.RawMessage `json:"testNotification"`
	SubscriptionNotification *struct {
		PurchaseToken  string `json:"purchaseToken"`
		SubscriptionID string `json:"subscriptionId"`
	} `json:"subscriptionNotification"`
}

//nolint:gocyclo,cyclop // Pub/Sub envelope validation is kept linear to fail closed at every boundary.
func (uc *UseCase) HandleGoogleWebhook(ctx context.Context, body []byte) error {
	var envelope pubSubEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil || envelope.Message.MessageID == "" || envelope.Message.Data == "" {
		return entity.ErrInvalidIAPPurchase
	}

	decoded, err := base64.StdEncoding.DecodeString(envelope.Message.Data)
	if err != nil {
		return entity.ErrInvalidIAPPurchase
	}

	var notification developerNotification
	if err = json.Unmarshal(decoded, &notification); err != nil {
		return entity.ErrInvalidIAPPurchase
	}

	if notification.PackageName != uc.cfg.GooglePackageName {
		return entity.ErrInvalidIAPPurchase
	}

	eventMillis, err := strconv.ParseInt(notification.EventTimeMillis, 10, 64)
	if err != nil {
		return entity.ErrInvalidIAPPurchase
	}

	eventTime := time.UnixMilli(eventMillis).UTC()
	if notification.SubscriptionNotification == nil {
		event := entity.WebhookEvent{
			Provider: "google", EventID: envelope.Message.MessageID,
			EventTime: eventTime, Ignored: true,
		}

		return uc.persistWebhook(ctx, &event)
	}

	if notification.SubscriptionNotification.PurchaseToken == "" {
		return entity.ErrInvalidIAPPurchase
	}

	result, err := uc.google.VerifySubscription(ctx, notification.PackageName, notification.SubscriptionNotification.PurchaseToken)
	if err != nil {
		return err
	}

	result.EventTime = eventTime
	event := entity.WebhookEvent{Provider: "google", EventID: envelope.Message.MessageID, EventTime: eventTime, Result: result}

	return uc.persistWebhook(ctx, &event)
}

func (uc *UseCase) persistWebhook(ctx context.Context, event *entity.WebhookEvent) error {
	if event.Ignored {
		return uc.uow.WithinTransaction(ctx, func(repos iaprepo.Repositories) error {
			_, err := repos.WebhookEvents.InsertWebhookEvent(ctx, event.Provider, event.EventID, event.EventTime)

			return err
		})
	}

	existing, err := uc.repo.GetByOriginalTransactionID(ctx, event.Result.OriginalTransactionID)
	if errors.Is(err, entity.ErrSubscriptionNotFound) {
		// A notification can precede the authenticated client verification. There
		// is no trustworthy app user to attach it to yet.
		return nil
	}

	if err != nil {
		return err
	}

	return uc.persistResult(ctx, existing.UserID, &event.Result, event.Provider, event.EventID)
}

//nolint:funlen,gocognit,gocyclo,cyclop // This atomic entitlement boundary keeps rollback semantics visible.
func (uc *UseCase) persistResult(ctx context.Context, userID string, result *entity.StoreVerificationResult,
	provider, eventID string,
) error {
	if result.TransactionID == "" || result.OriginalTransactionID == "" || result.ProductID == "" || result.EventTime.IsZero() {
		return entity.ErrInvalidIAPPurchase
	}

	return uc.uow.WithinTransaction(ctx, func(repos iaprepo.Repositories) error {
		if eventID != "" {
			inserted, err := repos.WebhookEvents.InsertWebhookEvent(ctx, provider, eventID, result.EventTime)
			if err != nil || !inserted {
				return err
			}
		}

		previous, err := repos.Subscriptions.GetByOriginalTransactionID(ctx, result.OriginalTransactionID)
		if err != nil && !errors.Is(err, entity.ErrSubscriptionNotFound) {
			return err
		}

		transfer := uc.cfg.RestorePolicy == "transfer"
		if err == nil && previous.UserID != userID && previous.Status.Entitled() && previous.ExpiresAt.After(uc.now()) && !transfer {
			return entity.ErrIAPAlreadyOwned
		}

		sub := &entity.UserSubscription{
			UserID: userID, Platform: result.Platform, ProductID: result.ProductID,
			OriginalTransactionID: result.OriginalTransactionID, Status: result.Status, StartsAt: result.PurchaseTime,
			ExpiresAt: result.ExpiresTime, AutoRenew: result.AutoRenew, LastEventAt: result.EventTime, AllowTransfer: transfer,
		}

		applied, err := repos.Subscriptions.UpsertSubscriptionTx(ctx, sub)
		if err != nil {
			return err
		}

		if !applied {
			return nil
		}

		receipt := &entity.IAPReceipt{
			UserID: userID, Platform: result.Platform, ProductID: result.ProductID,
			TransactionID: result.TransactionID, OriginalTransactionID: result.OriginalTransactionID,
			PurchaseTime: result.PurchaseTime, ExpiresTime: result.ExpiresTime, RawPayload: result.RawPayload,
		}
		if receiptErr := repos.Receipts.InsertReceipt(ctx, receipt); receiptErr != nil {
			return receiptErr
		}

		var premiumUntil *time.Time

		if result.Status.Entitled() && result.ExpiresTime.After(uc.now()) {
			until := result.ExpiresTime
			premiumUntil = &until
		}

		if premiumErr := repos.Users.UpdatePremiumUntil(ctx, userID, premiumUntil); premiumErr != nil {
			return premiumErr
		}

		if previous != nil && previous.UserID != userID && transfer {
			if err = repos.Users.UpdatePremiumUntil(ctx, previous.UserID, nil); err != nil {
				return fmt.Errorf("clear previous IAP owner: %w", err)
			}
		}

		return nil
	})
}
