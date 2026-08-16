package request

// AppleIAPWebhook is the App Store Server Notifications V2 request body.
type AppleIAPWebhook struct {
	SignedPayload string `json:"signedPayload" validate:"required" example:"eyJhbGciOiJFUzI1NiIsIng1YyI6Wy4uLl19..."`
}

// GoogleIAPWebhook is the Google Cloud Pub/Sub push envelope used for RTDN.
type GoogleIAPWebhook struct {
	Message      GooglePubSubMessage `json:"message" validate:"required"`
	Subscription string              `json:"subscription,omitempty" example:"projects/example/subscriptions/iap-rtdn"`
}

// GooglePubSubMessage contains the Base64-encoded DeveloperNotification payload.
type GooglePubSubMessage struct {
	Data      string `json:"data" validate:"required" example:"eyJ2ZXJzaW9uIjoiMS4wIiwicGFja2FnZU5hbWUiOiJjb20uZXhhbXBsZS5hcHAifQ=="`
	MessageID string `json:"messageId" validate:"required" example:"1234567890"`
}
