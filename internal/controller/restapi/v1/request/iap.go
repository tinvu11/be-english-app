package request

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
