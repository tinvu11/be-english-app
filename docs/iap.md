# In-App Purchase configuration (Android / Google Play)

The IAP module is disabled by default. Apply migration `20260816000001_create_iap`, provision Google Play integration, then set `IAP_ENABLED=true`.

## Environment variables

| Variable | Required when enabled | Description |
|---|---:|---|
| `IAP_ENABLED` | yes | Registers the IAP gateways and HTTP routes. |
| `IAP_RESTORE_POLICY` | yes | `block` rejects an active purchase owned by another app user; `transfer` closes the old entitlement and transfers it atomically. Default: `block`. |
| `IAP_TIMEOUT_SECONDS` | no | HTTP timeout for store calls. Default: `15`. |
| `GOOGLE_PLAY_PACKAGE_NAME` | yes | Exact Android application ID. Client input and RTDN payloads must match it. |
| `GOOGLE_PLAY_CREDENTIALS_FILE` | yes | Path to a dedicated Google service-account JSON key. Do not reuse the Firebase credential implicitly. |

For Google Play, enable the Google Play Android Developer API, invite the service account in Play Console, and grant only the permissions needed to view orders/subscriptions and manage subscription acknowledgements. The server calls `purchases.subscriptionsv2.get`; an unacknowledged purchase is acknowledged only after the entitlement transaction commits.

## Store callbacks

- Google Cloud Pub/Sub push subscription for RTDN: `POST /v1/iap/webhooks/google`

This callback is a public HTTP route because Google Cloud Pub/Sub calls it. Google messages never directly grant access: the purchase token is always re-fetched through the authenticated Play Developer API, and the configured package name is enforced. At the edge, additionally restrict request size, enable rate limiting, and configure an authenticated Pub/Sub push identity/API gateway when available.

Pub/Sub must forward the standard envelope containing `message.data` and `message.messageId`. `message.data` is the Base64-encoded DeveloperNotification JSON. Google `messageId` is stored in `iap_webhook_events` inside the same transaction as entitlement changes.

## Client API

- `POST /v1/iap/verify` requires the existing Firebase bearer token. Body:

  ```json
  {
    "platform": "android",
    "product_id": "premium_monthly",
    "purchase_token": "google-play-purchase-token",
    "package_name": "com.example.app"
  }
  ```

- `GET /v1/iap/status` requires the same bearer token and returns the latest subscription plus `isPremium` and `premiumUntil`.

## Operational notes

- Use HTTPS only. Keep service-account files in a secret manager/volume with read-only permissions.
- Alert on store API failures and webhook 5xx responses; PubSub retries non-2xx deliveries.
- Periodically reconcile active rows with Google Play. Webhooks are an update signal, not the sole source of truth.
- `premium_until` is a cache. The transactional subscription row remains the source of truth.
