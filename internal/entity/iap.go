package entity

import (
	"encoding/json"
	"time"
)

type IAPPlatform string

const (
	PlatformIOS     IAPPlatform = "ios"
	PlatformAndroid IAPPlatform = "android"
)

func (p IAPPlatform) Valid() bool { return p == PlatformIOS || p == PlatformAndroid }

type SubscriptionStatus string

const (
	SubscriptionActive        SubscriptionStatus = "active"
	SubscriptionExpired       SubscriptionStatus = "expired"
	SubscriptionInGracePeriod SubscriptionStatus = "in_grace_period"
	SubscriptionCancelled     SubscriptionStatus = "cancelled"
)

func (s SubscriptionStatus) Entitled() bool {
	return s == SubscriptionActive || s == SubscriptionInGracePeriod
}

type UserSubscription struct {
	ID                    int64              `json:"id"`
	UserID                string             `json:"userId"`
	Platform              IAPPlatform        `json:"platform"`
	ProductID             string             `json:"productId"`
	OriginalTransactionID string             `json:"originalTransactionId"`
	Status                SubscriptionStatus `json:"status"`
	StartsAt              time.Time          `json:"startsAt"`
	ExpiresAt             time.Time          `json:"expiresAt"`
	AutoRenew             bool               `json:"autoRenew"`
	LastEventAt           time.Time          `json:"lastEventAt"`
	CreatedAt             time.Time          `json:"createdAt"`
	UpdatedAt             time.Time          `json:"updatedAt"`
	AllowTransfer         bool               `json:"-"`
}

type IAPReceipt struct {
	ID                    int64           `json:"id"`
	UserID                string          `json:"userId"`
	Platform              IAPPlatform     `json:"platform"`
	ProductID             string          `json:"productId"`
	TransactionID         string          `json:"transactionId"`
	OriginalTransactionID string          `json:"originalTransactionId"`
	PurchaseTime          time.Time       `json:"purchaseTime"`
	ExpiresTime           time.Time       `json:"expiresTime"`
	RawPayload            json.RawMessage `json:"-"`
	CreatedAt             time.Time       `json:"createdAt"`
}

type StoreVerificationResult struct {
	Platform              IAPPlatform
	ProductID             string
	TransactionID         string
	OriginalTransactionID string
	PurchaseTime          time.Time
	ExpiresTime           time.Time
	Status                SubscriptionStatus
	AutoRenew             bool
	Acknowledged          bool
	EventTime             time.Time
	RawPayload            json.RawMessage
}

type VerifyPurchaseInput struct {
	Platform      IAPPlatform `json:"platform" validate:"required"`
	ProductID     string      `json:"product_id" validate:"required,max=100"`
	PurchaseToken string      `json:"purchase_token" validate:"required"`
	PackageName   string      `json:"package_name"`
}

type IAPStatus struct {
	IsPremium    bool              `json:"isPremium"`
	PremiumUntil *time.Time        `json:"premiumUntil,omitempty"`
	Subscription *UserSubscription `json:"subscription,omitempty"`
}

type WebhookEvent struct {
	Provider  string
	EventID   string
	EventTime time.Time
	Result    StoreVerificationResult
	Ignored   bool
}
