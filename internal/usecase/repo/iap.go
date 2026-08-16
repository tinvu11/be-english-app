package repo

import (
	"context"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
)

type SubscriptionRepo interface {
	Create(ctx context.Context, subscription *entity.UserSubscription) error
	Update(ctx context.Context, subscription *entity.UserSubscription) error
	GetActiveByUserID(ctx context.Context, userID string) (*entity.UserSubscription, error)
	GetByOriginalTransactionID(ctx context.Context, originalTransactionID string) (*entity.UserSubscription, error)
	UpsertSubscriptionTx(ctx context.Context, subscription *entity.UserSubscription) (bool, error)
}

type ReceiptRepo interface {
	InsertReceipt(ctx context.Context, receipt *entity.IAPReceipt) error
	CheckTransactionExists(ctx context.Context, transactionID string) (bool, error)
}

type UserRepo interface {
	UpdatePremiumUntil(ctx context.Context, userID string, premiumUntil *time.Time) error
}

type WebhookEventRepo interface {
	InsertWebhookEvent(ctx context.Context, provider, eventID string, eventTime time.Time) (bool, error)
}

type Repositories struct {
	Subscriptions SubscriptionRepo
	Receipts      ReceiptRepo
	Users         UserRepo
	WebhookEvents WebhookEventRepo
}

type UnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(Repositories) error) error
}
