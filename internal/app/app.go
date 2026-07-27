// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/internal/controller/restapi"
	persistContentRepo "github.com/evrone/go-clean-template/internal/repo/persistent/content"
	persistTranslationRepo "github.com/evrone/go-clean-template/internal/repo/persistent/translation"
	persistUserRepo "github.com/evrone/go-clean-template/internal/repo/persistent/user"
	"github.com/evrone/go-clean-template/internal/repo/webapi"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/internal/usecase/content"
	"github.com/evrone/go-clean-template/internal/usecase/translation"
	"github.com/evrone/go-clean-template/internal/usecase/user"
	"github.com/evrone/go-clean-template/pkg/firebaseauth"
	"github.com/evrone/go-clean-template/pkg/httpserver"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/evrone/go-clean-template/pkg/tracing"
)

type useCases struct {
	translation usecase.Translation
	user        usecase.User
	content     usecase.Content
}

type servers struct {
	http *httpserver.Server
}

func initUseCases(pg *postgres.Postgres) useCases {
	translationRepo := persistTranslationRepo.New(pg)
	userRepo := persistUserRepo.New(pg)
	contentRepo := persistContentRepo.New(pg)

	return useCases{
		translation: translation.New(translationRepo, webapi.New()),
		user:        user.New(userRepo),
		content:     content.New(contentRepo),
	}
}

func initServers(cfg *config.Config, uc useCases, verifier *firebaseauth.Verifier, l logger.Interface) servers {
	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	restapi.NewRouter(httpServer.App, cfg, uc.translation, uc.user, uc.content, verifier, l)

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

	uc := initUseCases(pg)
	s := initServers(cfg, uc, firebaseVerifier, l)
	s.startServers()
	s.waitForShutdown(l)
}
