package config

import "testing"

func TestValidateIAPAllowsGoogleOnly(t *testing.T) {
	t.Parallel()

	cfg := &Config{IAP: iap{
		Enabled: true, AppleEnabled: false, RestorePolicy: "block",
		GooglePackageName: "com.example.app", GoogleCredentialsFile: "google-service-account.json",
	}}

	if err := validateIAP(cfg); err != nil {
		t.Fatalf("validate Google-only IAP: %v", err)
	}
}
