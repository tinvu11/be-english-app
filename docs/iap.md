# In-App Purchase configuration

The IAP module is disabled by default. Apply migration `20260816000001_create_iap`, provision both store integrations, then set `IAP_ENABLED=true`.

## Environment variables

| Variable | Required when enabled | Description |
|---|---:|---|
| `IAP_ENABLED` | yes | Registers the IAP gateways and HTTP routes. |
| `IAP_RESTORE_POLICY` | yes | `block` rejects an active purchase owned by another app user; `transfer` closes the old entitlement and transfers it atomically. Default: `block`. |
| `IAP_TIMEOUT_SECONDS` | no | HTTP timeout for App Store calls. Default: `15`. |
| `APPLE_IAP_BASE_URL` | yes | Production: `https://api.storekit.apple.com`; sandbox: `https://api.storekit-sandbox.apple.com`. |
| `APPLE_IAP_ISSUER_ID` | yes | App Store Connect In-App Purchase issuer ID. |
| `APPLE_IAP_KEY_ID` | yes | ID of the In-App Purchase API key. |
| `APPLE_IAP_BUNDLE_ID` | yes | Exact iOS bundle ID. It is checked against every signed transaction and notification. |
| `APPLE_IAP_APP_ID` | production | Numeric App Store app ID, checked on production notifications. |
| `APPLE_IAP_ENVIRONMENT` | yes | Signed-data environment: `Production` or `Sandbox`; keep it consistent with `APPLE_IAP_BASE_URL`. |
| `APPLE_IAP_PRIVATE_KEY_PATH` | yes | Path to the downloaded `.p8` EC private key. Mount it as a secret; never commit it. |
| `APPLE_IAP_ROOT_CA_PATH` | yes | PEM bundle containing the trusted Apple Root CA certificate(s) used to pin validation of the `x5c` JWS chain. The generic OS trust store is intentionally not accepted. |
| `GOOGLE_PLAY_PACKAGE_NAME` | yes | Exact Android application ID. Client input and RTDN payloads must match it. |
| `GOOGLE_PLAY_CREDENTIALS_FILE` | yes | Path to a dedicated Google service-account JSON key. Do not reuse the Firebase credential implicitly. |

Apple request JWTs use ES256 and are generated per request with a five-minute lifetime. Transaction responses, notification `signedPayload`, `signedTransactionInfo`, and `signedRenewalInfo` are verified against a pinned Apple certificate chain before parsing. Verification also enforces the App Store signing/WWDR certificate OIDs, bundle ID, app ID in production, and environment.

For Google Play, enable the Google Play Android Developer API, invite the service account in Play Console, and grant only the permissions needed to view orders/subscriptions and manage subscription acknowledgements. The server calls `purchases.subscriptionsv2.get`; an unacknowledged purchase is acknowledged only after the entitlement transaction commits.

## Store callbacks

- Apple App Store Server Notifications V2: `POST /v1/iap/webhooks/apple`
- Google Cloud Pub/Sub push subscription for RTDN: `POST /v1/iap/webhooks/google`

Both callbacks are public HTTP routes because the stores call them. Apple authenticity comes from JWS certificate/signature validation. Google messages never directly grant access: the purchase token is always re-fetched through the authenticated Play Developer API, and the configured package name is enforced. At the edge, additionally restrict request size, enable rate limiting, and configure an authenticated Pub/Sub push identity/API gateway when available.

Pub/Sub must forward the standard envelope containing `message.data` and `message.messageId`. `message.data` is the Base64-encoded DeveloperNotification JSON. Apple `notificationUUID` and Google `messageId` are stored in `iap_webhook_events` inside the same transaction as entitlement changes.

## Client API

- `POST /v1/iap/verify` requires the existing Firebase bearer token. Body:

  ```json
  {
    "platform": "android",
    "product_id": "premium_monthly",
    "purchase_token": "store-token-or-apple-transaction-id",
    "package_name": "com.example.app"
  }
  ```

  For iOS, `purchase_token` is the StoreKit 2 transaction ID and `package_name` may be omitted.

- `GET /v1/iap/status` requires the same bearer token and returns the latest subscription plus `isPremium` and `premiumUntil`.

## Operational notes

- Use HTTPS only. Keep `.p8` and service-account files in a secret manager/volume with read-only permissions.
- Configure separate deployments or environment-specific `APPLE_IAP_BASE_URL` values for sandbox and production.
- Alert on store API failures and webhook 5xx responses; stores/PubSub retry non-2xx deliveries.
- Periodically reconcile active rows with both stores. Webhooks are an update signal, not the sole source of truth.
- `premium_until` is a cache. The transactional subscription row remains the source of truth.
