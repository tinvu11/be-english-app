package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type (
	// Config -.
	Config struct {
		App      app
		HTTP     http
		Log      log
		PG       pg
		Firebase firebase
		Metrics  metrics
		Swagger  swagger
		Tracing  tracing
		YTDLP    ytdlp
		DeepSeek deepseek
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
		BinaryPath     string `env:"YTDLP_BINARY_PATH" envDefault:"yt-dlp"`
		TimeoutSeconds int    `env:"YTDLP_TIMEOUT_SECONDS" envDefault:"45"`
		MaxFileMB      int64  `env:"YTDLP_MAX_FILE_MB" envDefault:"10"`
	}

	deepseek struct {
		BaseURL        string `env:"DEEPSEEK_BASE_URL" envDefault:"https://api.deepseek.com"`
		APIKey         string `env:"DEEPSEEK_API_KEY"`
		Model          string `env:"DEEPSEEK_MODEL" envDefault:"deepseek-v4-flash"`
		TimeoutSeconds int    `env:"DEEPSEEK_TIMEOUT_SECONDS" envDefault:"60"`
		MaxBatchItems  int    `env:"DEEPSEEK_MAX_BATCH_ITEMS" envDefault:"50"`
		MaxRetries     int    `env:"DEEPSEEK_MAX_RETRIES" envDefault:"2"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
