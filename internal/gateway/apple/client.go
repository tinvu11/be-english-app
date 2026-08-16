package apple

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/golang-jwt/jwt/v5"
)

const (
	maxResponseBytes = 1 << 20
	apiTokenTTL      = 5 * time.Minute
	compactJWSParts  = 3
)

type Config struct {
	BaseURL        string
	IssuerID       string
	KeyID          string
	BundleID       string
	AppAppleID     int64
	Environment    string
	PrivateKeyPath string
	RootCAPath     string
}

type Client struct {
	cfg        Config
	httpClient *http.Client
	privateKey *ecdsa.PrivateKey
	roots      *x509.CertPool
}

var (
	errInvalidPrivateKey = errors.New("apple gateway: invalid private key")
	errInvalidRootCA     = errors.New("apple gateway: invalid root CA")
	errInvalidX5C        = errors.New("apple gateway: invalid x5c certificate chain")
	errInvalidCompactJWS = errors.New("apple gateway: invalid compact JWS")
)

func New(cfg *Config, httpClient *http.Client) (*Client, error) {
	keyPEM, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("apple gateway - read private key: %w", err)
	}

	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errInvalidPrivateKey
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("apple gateway - parse private key: %w", err)
	}

	privateKey, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errInvalidPrivateKey
	}

	if cfg.RootCAPath == "" {
		return nil, errInvalidRootCA
	}

	roots := x509.NewCertPool()

	rootPEM, readErr := os.ReadFile(cfg.RootCAPath)
	if readErr != nil {
		return nil, fmt.Errorf("apple gateway - read root CA: %w", readErr)
	}

	if !roots.AppendCertsFromPEM(rootPEM) {
		return nil, errInvalidRootCA
	}

	return &Client{cfg: *cfg, httpClient: httpClient, privateKey: privateKey, roots: roots}, nil
}

func (c *Client) VerifyTransaction(ctx context.Context, transactionID string) (entity.StoreVerificationResult, error) {
	token, err := c.apiToken()
	if err != nil {
		return entity.StoreVerificationResult{}, err
	}

	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/inApps/v1/transactions/" + url.PathEscape(transactionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return entity.StoreVerificationResult{}, fmt.Errorf("apple gateway - request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return entity.StoreVerificationResult{}, fmt.Errorf("%w: apple request: %w", entity.ErrIAPVerificationFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return entity.StoreVerificationResult{}, fmt.Errorf("apple gateway - read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return entity.StoreVerificationResult{}, fmt.Errorf("%w: apple returned HTTP %d", entity.ErrIAPVerificationFailed, resp.StatusCode)
	}

	var response struct {
		SignedTransactionInfo string `json:"signedTransactionInfo"`
	}
	if err = json.Unmarshal(body, &response); err != nil || response.SignedTransactionInfo == "" {
		return entity.StoreVerificationResult{}, fmt.Errorf("%w: malformed apple response", entity.ErrIAPVerificationFailed)
	}

	return c.parseTransaction(response.SignedTransactionInfo)
}

//nolint:funlen,gocyclo,cyclop // Mapping signed Apple event variants is one audited protocol switch.
func (c *Client) ParseNotification(_ context.Context, signedPayload string) (entity.WebhookEvent, error) {
	var payload struct {
		NotificationType string `json:"notificationType"`
		Subtype          string `json:"subtype"`
		NotificationUUID string `json:"notificationUUID"`
		SignedDate       int64  `json:"signedDate"`
		Data             struct {
			BundleID              string `json:"bundleId"`
			AppAppleID            int64  `json:"appAppleId"`
			Environment           string `json:"environment"`
			SignedTransactionInfo string `json:"signedTransactionInfo"`
			SignedRenewalInfo     string `json:"signedRenewalInfo"`
		} `json:"data"`
	}
	if err := c.verifyJWS(signedPayload, &payload); err != nil {
		return entity.WebhookEvent{}, fmt.Errorf("%w: apple notification: %w", entity.ErrInvalidIAPPurchase, err)
	}

	if payload.NotificationUUID == "" || payload.Data.BundleID != c.cfg.BundleID ||
		payload.Data.Environment != c.cfg.Environment ||
		(c.cfg.Environment == "Production" && payload.Data.AppAppleID != c.cfg.AppAppleID) {
		return entity.WebhookEvent{}, entity.ErrInvalidIAPPurchase
	}

	eventTime := time.UnixMilli(payload.SignedDate).UTC()
	switch payload.NotificationType {
	case "DID_RENEW", "DID_FAIL_TO_RENEW", "EXPIRED", "REVOKE":
	default:
		return entity.WebhookEvent{
			Provider: "apple", EventID: payload.NotificationUUID,
			EventTime: eventTime, Ignored: true,
		}, nil
	}

	result, err := c.parseTransaction(payload.Data.SignedTransactionInfo)
	if err != nil {
		return entity.WebhookEvent{}, err
	}

	if payload.Data.SignedRenewalInfo != "" {
		var renewal struct {
			AutoRenewStatus int `json:"autoRenewStatus"`
		}
		if err = c.verifyJWS(payload.Data.SignedRenewalInfo, &renewal); err != nil {
			return entity.WebhookEvent{}, fmt.Errorf("%w: apple renewal: %w", entity.ErrInvalidIAPPurchase, err)
		}

		result.AutoRenew = renewal.AutoRenewStatus == 1
	}

	result.EventTime = eventTime

	switch payload.NotificationType {
	case "DID_RENEW":
		result.Status = entity.SubscriptionActive
	case "DID_FAIL_TO_RENEW":
		if payload.Subtype == "GRACE_PERIOD" {
			result.Status = entity.SubscriptionInGracePeriod
		} else {
			result.Status = entity.SubscriptionExpired
		}
	case "EXPIRED":
		result.Status = entity.SubscriptionExpired
		result.AutoRenew = false
	case "REVOKE":
		result.Status = entity.SubscriptionCancelled
		result.AutoRenew = false
	}

	return entity.WebhookEvent{Provider: "apple", EventID: payload.NotificationUUID, EventTime: eventTime, Result: result}, nil
}

func (c *Client) apiToken() (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss": c.cfg.IssuerID, "iat": now.Unix(), "exp": now.Add(apiTokenTTL).Unix(),
		"aud": "appstoreconnect-v1", "bid": c.cfg.BundleID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.cfg.KeyID
	token.Header["typ"] = "JWT"

	signed, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("apple gateway - sign API token: %w", err)
	}

	return signed, nil
}

type transactionPayload struct {
	TransactionID         string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	ProductID             string `json:"productId"`
	BundleID              string `json:"bundleId"`
	Environment           string `json:"environment"`
	PurchaseDate          int64  `json:"purchaseDate"`
	ExpiresDate           int64  `json:"expiresDate"`
	RevocationDate        int64  `json:"revocationDate"`
}

func (c *Client) parseTransaction(signed string) (entity.StoreVerificationResult, error) {
	var payload transactionPayload
	if err := c.verifyJWS(signed, &payload); err != nil {
		return entity.StoreVerificationResult{}, fmt.Errorf("%w: apple transaction JWS: %w", entity.ErrIAPVerificationFailed, err)
	}

	if payload.BundleID != c.cfg.BundleID || payload.Environment != c.cfg.Environment ||
		payload.TransactionID == "" || payload.OriginalTransactionID == "" {
		return entity.StoreVerificationResult{}, entity.ErrInvalidIAPPurchase
	}

	purchaseAt := time.UnixMilli(payload.PurchaseDate).UTC()
	expiresAt := time.UnixMilli(payload.ExpiresDate).UTC()

	status := entity.SubscriptionActive
	if payload.RevocationDate > 0 {
		status = entity.SubscriptionCancelled
	} else if !expiresAt.After(time.Now()) {
		status = entity.SubscriptionExpired
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return entity.StoreVerificationResult{}, fmt.Errorf("apple gateway - marshal transaction: %w", err)
	}

	return entity.StoreVerificationResult{
		Platform: entity.PlatformIOS, ProductID: payload.ProductID,
		TransactionID: payload.TransactionID, OriginalTransactionID: payload.OriginalTransactionID,
		PurchaseTime: purchaseAt, ExpiresTime: expiresAt, Status: status, AutoRenew: status == entity.SubscriptionActive,
		EventTime: purchaseAt, RawPayload: raw,
	}, nil
}

//nolint:funlen,gocognit,gocyclo,cyclop // Chain parsing and signature validation remain inseparable.
func (c *Client) verifyJWS(signed string, destination any) error {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}))

	_, err := parser.Parse(signed, func(token *jwt.Token) (any, error) {
		chainValue, ok := token.Header["x5c"].([]any)
		if !ok || len(chainValue) != compactJWSParts {
			return nil, errInvalidX5C
		}

		certs := make([]*x509.Certificate, 0, len(chainValue))
		for _, value := range chainValue {
			encoded, ok := value.(string)
			if !ok {
				return nil, errInvalidX5C
			}

			der, decodeErr := base64.StdEncoding.DecodeString(encoded)
			if decodeErr != nil {
				return nil, decodeErr
			}

			cert, parseErr := x509.ParseCertificate(der)
			if parseErr != nil {
				return nil, parseErr
			}

			certs = append(certs, cert)
		}

		intermediates := x509.NewCertPool()
		for _, cert := range certs[1:] {
			intermediates.AddCert(cert)
		}

		if !hasExtension(certs[0], asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 11, 1}) ||
			!hasExtension(certs[1], asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 2, 1}) {
			return nil, errInvalidX5C
		}

		if _, verifyErr := certs[0].Verify(x509.VerifyOptions{
			Roots: c.roots, Intermediates: intermediates,
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		}); verifyErr != nil {
			return nil, verifyErr
		}

		publicKey, ok := certs[0].PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, errInvalidX5C
		}

		return publicKey, nil
	})
	if err != nil {
		return err
	}

	parts := strings.Split(signed, ".")
	if len(parts) != compactJWSParts {
		return errInvalidCompactJWS
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	return json.Unmarshal(payload, destination)
}

func hasExtension(cert *x509.Certificate, oid asn1.ObjectIdentifier) bool {
	for _, extension := range cert.Extensions {
		if extension.Id.Equal(oid) {
			return true
		}
	}

	return false
}
