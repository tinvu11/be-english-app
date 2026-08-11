// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/controller/restapi"
	"github.com/evrone/go-clean-template/internal/gateway/deepseek"
	"github.com/evrone/go-clean-template/internal/gateway/ytdlp"
	persistAdminUserRepo "github.com/evrone/go-clean-template/internal/repo/persistent/adminuser"
	persistCaptionRepo "github.com/evrone/go-clean-template/internal/repo/persistent/caption"
	persistChannelRepo "github.com/evrone/go-clean-template/internal/repo/persistent/channel"
	persistLanguageRepo "github.com/evrone/go-clean-template/internal/repo/persistent/language"
	persistLevelRepo "github.com/evrone/go-clean-template/internal/repo/persistent/level"
	persistTopicRepo "github.com/evrone/go-clean-template/internal/repo/persistent/topic"
	persistUserRepo "github.com/evrone/go-clean-template/internal/repo/persistent/user"
	persistVideoRepo "github.com/evrone/go-clean-template/internal/repo/persistent/video"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/internal/usecase/adminuser"
	"github.com/evrone/go-clean-template/internal/usecase/caption"
	"github.com/evrone/go-clean-template/internal/usecase/channel"
	"github.com/evrone/go-clean-template/internal/usecase/language"
	"github.com/evrone/go-clean-template/internal/usecase/level"
	"github.com/evrone/go-clean-template/internal/usecase/topic"
	"github.com/evrone/go-clean-template/internal/usecase/user"
	"github.com/evrone/go-clean-template/internal/usecase/video"
	"github.com/evrone/go-clean-template/pkg/firebaseauth"
	"github.com/evrone/go-clean-template/pkg/httpserver"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/evrone/go-clean-template/pkg/tracing"
)

type useCases struct {
	user      usecase.User
	language  usecase.Language
	level     usecase.Level
	topic     usecase.Topic
	channel   usecase.Channel
	adminUser usecase.AdminUser
	video     usecase.Video
	caption   usecase.Caption
}

type servers struct {
	http *httpserver.Server
}

func initUseCases(cfg *config.Config, pg *postgres.Postgres) useCases {
	userRepo := persistUserRepo.New(pg)
	languageRepo := persistLanguageRepo.New(pg)
	levelRepo := persistLevelRepo.New(pg)
	topicRepo := persistTopicRepo.New(pg)
	channelRepo := persistChannelRepo.New(pg)
	adminUserRepo := persistAdminUserRepo.New(pg)
	videoRepo := persistVideoRepo.New(pg)
	captionRepo := persistCaptionRepo.New(pg)
	subtitleProvider := ytdlp.New(cfg.YTDLP.BaseURL, &http.Client{Timeout: time.Duration(cfg.YTDLP.TimeoutSeconds) * time.Second})
	translator := deepseek.New(deepseek.Config{BaseURL: cfg.DeepSeek.BaseURL, APIKey: cfg.DeepSeek.APIKey,
		Model: cfg.DeepSeek.Model, MaxRetries: cfg.DeepSeek.MaxRetries},
		&http.Client{Timeout: time.Duration(cfg.DeepSeek.TimeoutSeconds) * time.Second})

	return useCases{
		user:      user.New(userRepo, languageRepo),
		language:  language.New(languageRepo),
		level:     level.New(levelRepo),
		topic:     topic.New(topicRepo),
		channel:   channel.New(channelRepo),
		adminUser: adminuser.New(adminUserRepo),
		video:     video.New(videoRepo, userRepo, subtitleProvider),
		caption:   caption.New(captionRepo, videoRepo, languageRepo, subtitleProvider, translator, cfg.DeepSeek.MaxBatchItems),
	}
}

func initServers(cfg *config.Config, uc useCases, verifier *firebaseauth.Verifier, l logger.Interface) servers {
	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	restapi.NewRouter(httpServer.App, cfg, uc.user, uc.language, uc.level, uc.topic, uc.channel, uc.adminUser, uc.video, uc.caption, verifier, l)

	return servers{
		http: httpServer,
	}
}

func (s *servers) startServers() {
	s.http.Start()
}

func (s *servers) waitForShutdown(l logger.Interface) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error

	select {
	case sig := <-interrupt:
		l.Info("app - Run - signal: %s", sig.String())
	case err = <-s.http.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	s.shutdownServers(l)
}

func (s *servers) shutdownServers(l logger.Interface) {
	if err := s.http.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	ctx := context.Background()

	// Tracing
	shutdownTracing, err := tracing.New(ctx, tracing.Config{
		Enabled:     cfg.Tracing.Enabled,
		ServiceName: cfg.App.Name,
		Version:     cfg.App.Version,
		Endpoint:    cfg.Tracing.OTLPEndpoint,
		Insecure:    cfg.Tracing.OTLPInsecure,
		SampleRate:  cfg.Tracing.SampleRate,
	})
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - tracing.New: %w", err))
	}
	defer func() {
		if err := shutdownTracing(ctx); err != nil {
			l.Error(fmt.Errorf("app - Run - shutdownTracing: %w", err))
		}
	}()

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Firebase Authentication
	firebaseVerifier, err := firebaseauth.New(ctx, cfg.Firebase.ProjectID, cfg.Firebase.CredentialsFile)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - firebaseauth.New: %w", err))
	}

	uc := initUseCases(cfg, pg)
	s := initServers(cfg, uc, firebaseVerifier, l)
	s.startServers()
	s.waitForShutdown(l)
}
