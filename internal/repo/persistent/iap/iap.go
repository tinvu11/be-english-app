package iap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	iaprepo "github.com/evrone/go-clean-template/internal/usecase/repo"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type db interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Repo struct {
	pool *postgres.Postgres
	db   db
}

func New(pg *postgres.Postgres) *Repo { return &Repo{pool: pg, db: pg.Pool} }

func (r *Repo) WithinTransaction(ctx context.Context, fn func(iaprepo.Repositories) error) (returnErr error) {
	tx, err := r.pool.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("IAPRepo - begin: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			returnErr = errors.Join(returnErr, fmt.Errorf("IAPRepo - rollback: %w", rollbackErr))
		}
	}()

	txRepo := &Repo{pool: r.pool, db: tx}

	ports := iaprepo.Repositories{Subscriptions: txRepo, Receipts: txRepo, Users: txRepo, WebhookEvents: txRepo}
	if transactionErr := fn(ports); transactionErr != nil {
		return transactionErr
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("IAPRepo - commit: %w", err)
	}

	return nil
}

func (r *Repo) Create(ctx context.Context, sub *entity.UserSubscription) error {
	err := r.db.QueryRow(ctx, `INSERT INTO user_subscriptions
		(user_id,platform,product_id,original_transaction_id,status,starts_at,expires_at,auto_renew,last_event_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id,created_at,updated_at`, sub.UserID, sub.Platform, sub.ProductID,
		sub.OriginalTransactionID, sub.Status, sub.StartsAt, sub.ExpiresAt, sub.AutoRenew, sub.LastEventAt).
		Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("IAPRepo - Create: %w", err)
	}

	return nil
}

func (r *Repo) Update(ctx context.Context, sub *entity.UserSubscription) error {
	tag, err := r.db.Exec(ctx, `UPDATE user_subscriptions SET product_id=$2,status=$3,starts_at=$4,
		expires_at=$5,auto_renew=$6,last_event_at=$7,updated_at=CURRENT_TIMESTAMP WHERE id=$1`,
		sub.ID, sub.ProductID, sub.Status, sub.StartsAt, sub.ExpiresAt, sub.AutoRenew, sub.LastEventAt)
	if err != nil {
		return fmt.Errorf("IAPRepo - Update: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entity.ErrSubscriptionNotFound
	}

	return nil
}

func (r *Repo) GetActiveByUserID(ctx context.Context, userID string) (*entity.UserSubscription, error) {
	return r.getSubscription(ctx, `WHERE user_id=$1 AND status IN ('active','in_grace_period')
		ORDER BY expires_at DESC,id DESC LIMIT 1`, userID)
}

func (r *Repo) GetByOriginalTransactionID(ctx context.Context, originalID string) (*entity.UserSubscription, error) {
	return r.getSubscription(ctx, `WHERE original_transaction_id=$1 ORDER BY last_event_at DESC,id DESC LIMIT 1`, originalID)
}

func (r *Repo) getSubscription(ctx context.Context, where string, arg any) (*entity.UserSubscription, error) {
	var sub entity.UserSubscription

	err := r.db.QueryRow(ctx, `SELECT id,user_id,platform,product_id,original_transaction_id,status,
		starts_at,expires_at,auto_renew,last_event_at,created_at,updated_at FROM user_subscriptions `+where, arg).
		Scan(&sub.ID, &sub.UserID, &sub.Platform, &sub.ProductID, &sub.OriginalTransactionID,
			&sub.Status, &sub.StartsAt, &sub.ExpiresAt, &sub.AutoRenew, &sub.LastEventAt, &sub.CreatedAt, &sub.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrSubscriptionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("IAPRepo - getSubscription: %w", err)
	}

	return &sub, nil
}

// UpsertSubscriptionTx serializes updates for one store purchase and ignores
// stale events based on the store timestamp, not their arrival order.
//
//nolint:gocognit,gocyclo,cyclop,nestif // Ownership transfer and the unique invariant share one serialized path.
func (r *Repo) UpsertSubscriptionTx(ctx context.Context, sub *entity.UserSubscription) (bool, error) {
	if _, err := r.db.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, sub.OriginalTransactionID); err != nil {
		return false, fmt.Errorf("IAPRepo - advisory lock: %w", err)
	}

	existing, err := r.GetByOriginalTransactionID(ctx, sub.OriginalTransactionID)
	if err == nil {
		if !sub.LastEventAt.After(existing.LastEventAt) {
			return false, nil
		}
		// A restore transfer creates a new ownership history row and closes the old one.
		if existing.UserID != sub.UserID {
			if existing.Status.Entitled() && existing.ExpiresAt.After(time.Now()) && !sub.AllowTransfer {
				return false, entity.ErrIAPAlreadyOwned
			}

			if _, err = r.db.Exec(ctx, `UPDATE user_subscriptions SET status='cancelled',auto_renew=FALSE,
				last_event_at=$2,updated_at=CURRENT_TIMESTAMP WHERE id=$1`, existing.ID, sub.LastEventAt); err != nil {
				return false, fmt.Errorf("IAPRepo - close previous owner: %w", err)
			}

			if sub.Status.Entitled() {
				if _, err = r.db.Exec(ctx, `UPDATE user_subscriptions SET status='cancelled',auto_renew=FALSE,
					updated_at=CURRENT_TIMESTAMP WHERE user_id=$1 AND status IN ('active','in_grace_period')`, sub.UserID); err != nil {
					return false, fmt.Errorf("IAPRepo - close transfer target subscription: %w", err)
				}
			}

			return true, r.Create(ctx, sub)
		}

		if sub.Status.Entitled() {
			if _, err = r.db.Exec(ctx, `UPDATE user_subscriptions SET status='cancelled',auto_renew=FALSE,
				updated_at=CURRENT_TIMESTAMP WHERE user_id=$1 AND id<>$2 AND status IN ('active','in_grace_period')`, sub.UserID, existing.ID); err != nil {
				return false, fmt.Errorf("IAPRepo - close superseded subscription: %w", err)
			}
		}

		sub.ID = existing.ID

		return true, r.Update(ctx, sub)
	}

	if !errors.Is(err, entity.ErrSubscriptionNotFound) {
		return false, err
	}

	// Product changes or re-signups with a new store identifier replace the
	// user's previous entitlement while retaining its history.
	if sub.Status.Entitled() {
		if _, err = r.db.Exec(ctx, `UPDATE user_subscriptions SET status='cancelled',auto_renew=FALSE,
			updated_at=CURRENT_TIMESTAMP WHERE user_id=$1 AND status IN ('active','in_grace_period')`, sub.UserID); err != nil {
			return false, fmt.Errorf("IAPRepo - close replaced subscription: %w", err)
		}
	}

	return true, r.Create(ctx, sub)
}

func (r *Repo) InsertReceipt(ctx context.Context, receipt *entity.IAPReceipt) error {
	tag, err := r.db.Exec(ctx, `INSERT INTO iap_receipts
		(user_id,platform,product_id,transaction_id,original_transaction_id,purchase_time,expires_time,raw_payload)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(transaction_id) DO NOTHING`,
		receipt.UserID, receipt.Platform, receipt.ProductID, receipt.TransactionID,
		receipt.OriginalTransactionID, receipt.PurchaseTime, receipt.ExpiresTime, receipt.RawPayload)
	if err != nil {
		return fmt.Errorf("IAPRepo - InsertReceipt: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return nil
	}

	return nil
}

func (r *Repo) CheckTransactionExists(ctx context.Context, transactionID string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM iap_receipts WHERE transaction_id=$1)`, transactionID).Scan(&exists); err != nil {
		return false, fmt.Errorf("IAPRepo - CheckTransactionExists: %w", err)
	}

	return exists, nil
}

func (r *Repo) UpdatePremiumUntil(ctx context.Context, userID string, premiumUntil *time.Time) error {
	tag, err := r.db.Exec(ctx, `UPDATE users SET premium_until=$2,updated_at=CURRENT_TIMESTAMP WHERE id=$1`, userID, premiumUntil)
	if err != nil {
		return fmt.Errorf("IAPRepo - UpdatePremiumUntil: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

func (r *Repo) InsertWebhookEvent(ctx context.Context, provider, eventID string, eventTime time.Time) (bool, error) {
	tag, err := r.db.Exec(ctx, `INSERT INTO iap_webhook_events(provider,event_id,event_time)
		VALUES($1,$2,$3) ON CONFLICT(provider,event_id) DO NOTHING`, provider, eventID, eventTime)
	if err != nil {
		return false, fmt.Errorf("IAPRepo - InsertWebhookEvent: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}
