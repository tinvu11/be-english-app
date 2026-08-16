package iap

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	iaprepo "github.com/evrone/go-clean-template/internal/usecase/repo"
	"github.com/stretchr/testify/require"
)

type appleStub struct {
	result entity.StoreVerificationResult
}

var errNotImplemented = errors.New("not implemented")

func (s *appleStub) VerifyTransaction(context.Context, string) (entity.StoreVerificationResult, error) {
	return s.result, nil
}

func (s *appleStub) ParseNotification(context.Context, string) (entity.WebhookEvent, error) {
	return entity.WebhookEvent{}, errNotImplemented
}

type googleStub struct {
	result       entity.StoreVerificationResult
	acknowledged int
	verifyCalls  int
}

func (s *googleStub) VerifySubscription(context.Context, string, string) (entity.StoreVerificationResult, error) {
	s.verifyCalls++

	return s.result, nil
}

func (s *googleStub) Acknowledge(context.Context, string, string, string) error {
	s.acknowledged++

	return nil
}

type memoryRepo struct {
	subs     map[string]*entity.UserSubscription
	receipts map[string]*entity.IAPReceipt
	events   map[string]struct{}
	premium  map[string]*time.Time
	upserts  int
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		subs: map[string]*entity.UserSubscription{}, receipts: map[string]*entity.IAPReceipt{},
		events: map[string]struct{}{}, premium: map[string]*time.Time{},
	}
}

func (m *memoryRepo) WithinTransaction(_ context.Context, fn func(iaprepo.Repositories) error) error {
	return fn(iaprepo.Repositories{Subscriptions: m, Receipts: m, Users: m, WebhookEvents: m})
}

func (m *memoryRepo) Create(_ context.Context, sub *entity.UserSubscription) error {
	cloned := *sub
	m.subs[sub.OriginalTransactionID] = &cloned

	return nil
}

func (m *memoryRepo) Update(_ context.Context, sub *entity.UserSubscription) error {
	cloned := *sub
	m.subs[sub.OriginalTransactionID] = &cloned

	return nil
}

func (m *memoryRepo) GetActiveByUserID(_ context.Context, userID string) (*entity.UserSubscription, error) {
	for _, sub := range m.subs {
		if sub.UserID == userID && sub.Status.Entitled() {
			cloned := *sub

			return &cloned, nil
		}
	}

	return nil, entity.ErrSubscriptionNotFound
}

func (m *memoryRepo) GetByOriginalTransactionID(_ context.Context, id string) (*entity.UserSubscription, error) {
	sub, ok := m.subs[id]
	if !ok {
		return nil, entity.ErrSubscriptionNotFound
	}

	cloned := *sub

	return &cloned, nil
}

func (m *memoryRepo) UpsertSubscriptionTx(_ context.Context, sub *entity.UserSubscription) (bool, error) {
	existing, ok := m.subs[sub.OriginalTransactionID]
	if ok && !sub.LastEventAt.After(existing.LastEventAt) {
		return false, nil
	}

	if ok && existing.UserID != sub.UserID && existing.Status.Entitled() && !sub.AllowTransfer {
		return false, entity.ErrIAPAlreadyOwned
	}

	m.upserts++
	cloned := *sub
	m.subs[sub.OriginalTransactionID] = &cloned

	return true, nil
}

func (m *memoryRepo) InsertReceipt(_ context.Context, receipt *entity.IAPReceipt) error {
	cloned := *receipt
	m.receipts[receipt.TransactionID] = &cloned

	return nil
}

func (m *memoryRepo) CheckTransactionExists(_ context.Context, id string) (bool, error) {
	_, ok := m.receipts[id]

	return ok, nil
}

func (m *memoryRepo) UpdatePremiumUntil(_ context.Context, userID string, until *time.Time) error {
	m.premium[userID] = until

	return nil
}

func (m *memoryRepo) InsertWebhookEvent(_ context.Context, provider, eventID string, _ time.Time) (bool, error) {
	key := provider + ":" + eventID
	if _, ok := m.events[key]; ok {
		return false, nil
	}

	m.events[key] = struct{}{}

	return true, nil
}

func verifiedPurchase(now time.Time) entity.StoreVerificationResult {
	return entity.StoreVerificationResult{
		Platform: entity.PlatformAndroid, ProductID: "premium_monthly",
		TransactionID: "GPA.1", OriginalTransactionID: "token-root", PurchaseTime: now.Add(-time.Hour),
		ExpiresTime: now.Add(30 * 24 * time.Hour), Status: entity.SubscriptionActive, AutoRenew: true,
		Acknowledged: false, EventTime: now,
	}
}

func TestVerifyGooglePersistsBeforeAcknowledging(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepo()
	google := &googleStub{result: verifiedPurchase(now)}
	uc := New(&appleStub{}, google, repo, repo, Config{GooglePackageName: "com.example.app", RestorePolicy: "block"})
	uc.now = func() time.Time { return now }

	status, err := uc.VerifyPurchase(context.Background(), "user-1", entity.VerifyPurchaseInput{
		Platform: entity.PlatformAndroid, ProductID: "premium_monthly", PurchaseToken: "token", PackageName: "com.example.app",
	})
	require.NoError(t, err)
	require.True(t, status.IsPremium)
	require.Equal(t, 1, google.acknowledged)
	require.Contains(t, repo.receipts, "GPA.1")
	require.NotNil(t, repo.premium["user-1"])
}

func TestVerifyBlocksActivePurchaseOwnedByAnotherUser(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	repo := newMemoryRepo()
	repo.subs["token-root"] = &entity.UserSubscription{
		UserID: "first-user", OriginalTransactionID: "token-root",
		Status: entity.SubscriptionActive, ExpiresAt: now.Add(time.Hour), LastEventAt: now.Add(-time.Minute),
	}
	google := &googleStub{result: verifiedPurchase(now)}
	uc := New(&appleStub{}, google, repo, repo, Config{GooglePackageName: "com.example.app", RestorePolicy: "block"})
	uc.now = func() time.Time { return now }

	_, err := uc.VerifyPurchase(context.Background(), "second-user", entity.VerifyPurchaseInput{
		Platform: entity.PlatformAndroid, ProductID: "premium_monthly", PurchaseToken: "token", PackageName: "com.example.app",
	})
	require.ErrorIs(t, err, entity.ErrIAPAlreadyOwned)
	require.Zero(t, google.acknowledged)
}

func TestGoogleWebhookIsIdempotentAndRejectsOlderState(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepo()
	repo.subs["token-root"] = &entity.UserSubscription{
		UserID: "user-1", OriginalTransactionID: "token-root",
		Status: entity.SubscriptionActive, ExpiresAt: now.Add(time.Hour), LastEventAt: now.Add(-time.Hour),
	}
	google := &googleStub{result: verifiedPurchase(now)}
	uc := New(&appleStub{}, google, repo, repo, Config{GooglePackageName: "com.example.app"})
	uc.now = func() time.Time { return now }

	body := googleEnvelope(t, "message-1", now, "token")
	require.NoError(t, uc.HandleGoogleWebhook(context.Background(), body))
	require.NoError(t, uc.HandleGoogleWebhook(context.Background(), body))
	require.Equal(t, 1, repo.upserts)

	older := googleEnvelope(t, "message-2", now.Add(-2*time.Hour), "token")
	require.NoError(t, uc.HandleGoogleWebhook(context.Background(), older))
	require.Equal(t, 1, repo.upserts)
}

func googleEnvelope(t *testing.T, messageID string, eventTime time.Time, token string) []byte {
	t.Helper()

	data, err := json.Marshal(map[string]any{
		"packageName":     "com.example.app",
		"eventTimeMillis": strconv.FormatInt(eventTime.UnixMilli(), 10), "subscriptionNotification": map[string]any{"purchaseToken": token, "subscriptionId": "premium_monthly"},
	})
	require.NoError(t, err)
	body, err := json.Marshal(map[string]any{"message": map[string]any{
		"messageId": messageID,
		"data":      base64.StdEncoding.EncodeToString(data),
	}})
	require.NoError(t, err)

	return body
}
