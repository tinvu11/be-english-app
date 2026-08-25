package config

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/caarlos0/env/v11"
)

var (
	errMissingIAPConfig = errors.New("IAP is enabled but required configuration is missing")
	errRestorePolicy    = errors.New("IAP_RESTORE_POLICY must be block or transfer")
	errInvalidQuota     = errors.New("feature quota limits must not be negative")
)

type (
	// Config -.
	Config struct {
		App         app
		HTTP        http
		Log         log
		PG          pg
		Firebase    firebase
		Metrics     metrics
		Swagger     swagger
		Tracing     tracing
		YTDLP       ytdlp
		DeepSeek    deepseek
		Groq        groq
		AzureSpeech azureSpeech
		Quota       quota
		IAP         iap
	}

	// App -.
	app struct {
		Name    string `env:"APP_NAME,required"`
		Version string `env:"APP_VERSION,required"`
	}

	// HTTP -.
	http struct {
		Port           string `env:"HTTP_PORT,required"`
		UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
	}

	// Log -.
	log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	// PG -.
	pg struct {
		PoolMax int    `env:"PG_POOL_MAX,required"`
		URL     string `env:"PG_URL,required"`
	}

	// Firebase -.
	firebase struct {
		ProjectID       string `env:"FIREBASE_PROJECT_ID,required"`
		CredentialsFile string `env:"GOOGLE_APPLICATION_CREDENTIALS,required"`
	}

	// Metrics -.
	metrics struct {
		Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	}

	// Swagger -.
	swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}

	// Tracing -.
	tracing struct {
		Enabled      bool    `env:"TRACING_ENABLED" envDefault:"false"`
		OTLPEndpoint string  `env:"TRACING_OTLP_ENDPOINT" envDefault:"localhost:4317"`
		OTLPInsecure bool    `env:"TRACING_OTLP_INSECURE" envDefault:"true"`
		SampleRate   float64 `env:"TRACING_SAMPLE_RATE" envDefault:"0.1"`
	}

	ytdlp struct {
		BaseURL        string `env:"YTDLP_BASE_URL" envDefault:"http://localhost:8081"`
		TimeoutSeconds int    `env:"YTDLP_TIMEOUT_SECONDS" envDefault:"45"`
	}

	deepseek struct {
		BaseURL        string `env:"DEEPSEEK_BASE_URL" envDefault:"https://api.deepseek.com"`
		APIKey         string `env:"DEEPSEEK_API_KEY"`
		Model          string `env:"DEEPSEEK_MODEL" envDefault:"deepseek-v4-flash"`
		TimeoutSeconds int    `env:"DEEPSEEK_TIMEOUT_SECONDS" envDefault:"60"`
		MaxBatchItems  int    `env:"DEEPSEEK_MAX_BATCH_ITEMS" envDefault:"50"`
		MaxRetries     int    `env:"DEEPSEEK_MAX_RETRIES" envDefault:"2"`
	}

	groq struct {
		BaseURL        string `env:"GROQ_BASE_URL" envDefault:"https://api.groq.com"`
		APIKey         string `env:"GROQ_API_KEY"`
		Model          string `env:"GROQ_WHISPER_MODEL" envDefault:"whisper-large-v3-turbo"`
		TimeoutSeconds int    `env:"GROQ_TIMEOUT_SECONDS" envDefault:"300"`
	}

	azureSpeech struct {
		Endpoint       string `env:"AZURE_SPEECH_ENDPOINT"`
		APIKey         string `env:"AZURE_SPEECH_KEY"`
		TimeoutSeconds int    `env:"AZURE_SPEECH_TIMEOUT_SECONDS" envDefault:"30"`
	}

	quota struct {
		YouTubeImportDailyLimit       int `env:"QUOTA_YOUTUBE_IMPORT_DAILY" envDefault:"3"`
		ShadowingAssessmentDailyLimit int `env:"QUOTA_SHADOWING_ASSESSMENT_DAILY" envDefault:"5"`
	}

	iap struct {
		Enabled               bool   `env:"IAP_ENABLED" envDefault:"false"`
		RestorePolicy         string `env:"IAP_RESTORE_POLICY" envDefault:"block"`
		TimeoutSeconds        int    `env:"IAP_TIMEOUT_SECONDS" envDefault:"15"`
		AppleBaseURL          string `env:"APPLE_IAP_BASE_URL" envDefault:"https://api.storekit.apple.com"`
		AppleIssuerID         string `env:"APPLE_IAP_ISSUER_ID"`
		AppleKeyID            string `env:"APPLE_IAP_KEY_ID"`
		AppleBundleID         string `env:"APPLE_IAP_BUNDLE_ID"`
		AppleAppID            int64  `env:"APPLE_IAP_APP_ID"`
		AppleEnvironment      string `env:"APPLE_IAP_ENVIRONMENT" envDefault:"Production"`
		ApplePrivateKeyPath   string `env:"APPLE_IAP_PRIVATE_KEY_PATH"`
		AppleRootCAPath       string `env:"APPLE_IAP_ROOT_CA_PATH"`
		GooglePackageName     string `env:"GOOGLE_PLAY_PACKAGE_NAME"`
		GoogleCredentialsFile string `env:"GOOGLE_PLAY_CREDENTIALS_FILE"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	if err := validateIAP(cfg); err != nil {
		return nil, err
	}

	if cfg.Quota.YouTubeImportDailyLimit < 0 || cfg.Quota.ShadowingAssessmentDailyLimit < 0 {
		return nil, fmt.Errorf("config error: %w", errInvalidQuota)
	}

	return cfg, nil
}

//nolint:gocyclo,cyclop // Startup validation reports each credential invariant precisely.
func validateIAP(cfg *Config) error {
	if !cfg.IAP.Enabled {
		return nil
	}

	missing := make([]string, 0)

	values := map[string]string{
		"APPLE_IAP_ISSUER_ID": cfg.IAP.AppleIssuerID, "APPLE_IAP_KEY_ID": cfg.IAP.AppleKeyID,
		"APPLE_IAP_BUNDLE_ID": cfg.IAP.AppleBundleID, "APPLE_IAP_PRIVATE_KEY_PATH": cfg.IAP.ApplePrivateKeyPath,
		"APPLE_IAP_ROOT_CA_PATH":   cfg.IAP.AppleRootCAPath,
		"GOOGLE_PLAY_PACKAGE_NAME": cfg.IAP.GooglePackageName, "GOOGLE_PLAY_CREDENTIALS_FILE": cfg.IAP.GoogleCredentialsFile,
	}
	for name, value := range values {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}

	sort.Strings(missing)

	if len(missing) > 0 {
		return fmt.Errorf("config error: %w: %s", errMissingIAPConfig, strings.Join(missing, ", "))
	}

	if cfg.IAP.RestorePolicy != "block" && cfg.IAP.RestorePolicy != "transfer" {
		return fmt.Errorf("config error: %w", errRestorePolicy)
	}

	if cfg.IAP.AppleEnvironment != "Production" && cfg.IAP.AppleEnvironment != "Sandbox" {
		return fmt.Errorf("config error: %w: APPLE_IAP_ENVIRONMENT", errMissingIAPConfig)
	}

	if cfg.IAP.AppleEnvironment == "Production" && cfg.IAP.AppleAppID <= 0 {
		return fmt.Errorf("config error: %w: APPLE_IAP_APP_ID", errMissingIAPConfig)
	}

	return nil
}
