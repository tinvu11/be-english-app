package config

import "testing"

func TestValidateIAP(t *testing.T) {
	t.Parallel()

	cfg := &Config{IAP: iap{
		Enabled: true, RestorePolicy: "block",
		GooglePackageName: "com.example.app", GoogleCredentialsFile: "google-service-account.json",
	}}

	if err := validateIAP(cfg); err != nil {
		t.Fatalf("validate Google IAP: %v", err)
	}
}
