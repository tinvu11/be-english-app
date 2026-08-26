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
	appleGateway "github.com/evrone/go-clean-template/internal/gateway/apple"
	"github.com/evrone/go-clean-template/internal/gateway/azure"
	"github.com/evrone/go-clean-template/internal/gateway/deepseek"
	googlePlayGateway "github.com/evrone/go-clean-template/internal/gateway/googleplay"
	"github.com/evrone/go-clean-template/internal/gateway/groq"
	"github.com/evrone/go-clean-template/internal/gateway/ytdlp"
	persistAdminUserRepo "github.com/evrone/go-clean-template/internal/repo/persistent/adminuser"
	persistCaptionRepo "github.com/evrone/go-clean-template/internal/repo/persistent/caption"
	persistChannelRepo "github.com/evrone/go-clean-template/internal/repo/persistent/channel"
	persistDictionaryRepo "github.com/evrone/go-clean-template/internal/repo/persistent/dictionary"
	persistIAPRepo "github.com/evrone/go-clean-template/internal/repo/persistent/iap"
	persistLanguageRepo "github.com/evrone/go-clean-template/internal/repo/persistent/language"
	persistLearningContentRepo "github.com/evrone/go-clean-template/internal/repo/persistent/learningcontent"
	persistLevelRepo "github.com/evrone/go-clean-template/internal/repo/persistent/level"
	persistQuotaRepo "github.com/evrone/go-clean-template/internal/repo/persistent/quota"
	persistShadowingRepo "github.com/evrone/go-clean-template/internal/repo/persistent/shadowing"
	persistTopicRepo "github.com/evrone/go-clean-template/internal/repo/persistent/topic"
	persistUserRepo "github.com/evrone/go-clean-template/internal/repo/persistent/user"
	persistVideoRepo "github.com/evrone/go-clean-template/internal/repo/persistent/video"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/internal/usecase/adminuser"
	"github.com/evrone/go-clean-template/internal/usecase/caption"
	"github.com/evrone/go-clean-template/internal/usecase/channel"
	iapGateway "github.com/evrone/go-clean-template/internal/usecase/gateway"
	iapUseCase "github.com/evrone/go-clean-template/internal/usecase/iap"
	"github.com/evrone/go-clean-template/internal/usecase/language"
	"github.com/evrone/go-clean-template/internal/usecase/learningcontent"
	"github.com/evrone/go-clean-template/internal/usecase/level"
	quotaUseCase "github.com/evrone/go-clean-template/internal/usecase/quota"
	"github.com/evrone/go-clean-template/internal/usecase/shadowing"
	"github.com/evrone/go-clean-template/internal/usecase/topic"
	"github.com/evrone/go-clean-template/internal/usecase/user"
	"github.com/evrone/go-clean-template/internal/usecase/video"
	"github.com/evrone/go-clean-template/internal/usecase/vocabulary"
	"github.com/evrone/go-clean-template/pkg/firebaseauth"
	"github.com/evrone/go-clean-template/pkg/httpserver"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/evrone/go-clean-template/pkg/tracing"
)

type useCases struct {
	user       usecase.User
	language   usecase.Language
	level      usecase.Level
	topic      usecase.Topic
	channel    usecase.Channel
	adminUser  usecase.AdminUser
	video      usecase.Video
	caption    usecase.Caption
	vocabulary usecase.Vocabulary
	learning   usecase.LearningContent
	shadowing  usecase.Shadowing
	iap        usecase.IAP
}

type servers struct {
	http *httpserver.Server
}

func initUseCases(ctx context.Context, cfg *config.Config, pg *postgres.Postgres) (useCases, error) {
	userRepo := persistUserRepo.New(pg)
	languageRepo := persistLanguageRepo.New(pg)
	levelRepo := persistLevelRepo.New(pg)
	topicRepo := persistTopicRepo.New(pg)
	channelRepo := persistChannelRepo.New(pg)
	adminUserRepo := persistAdminUserRepo.New(pg)
	videoRepo := persistVideoRepo.New(pg)
	captionRepo := persistCaptionRepo.New(pg)
	dictionaryRepo := persistDictionaryRepo.New(pg)
	learningContentRepo := persistLearningContentRepo.New(pg)
	shadowingRepo := persistShadowingRepo.New(pg)
	quotaRepo := persistQuotaRepo.New(pg)
	quota := quotaUseCase.New(quotaRepo, userRepo, quotaUseCase.Config{
		YouTubeImportDailyLimit:       cfg.Quota.YouTubeImportDailyLimit,
		ShadowingAssessmentDailyLimit: cfg.Quota.ShadowingAssessmentDailyLimit,
	})
	youtubeProvider := ytdlp.New(cfg.YTDLP.BaseURL, &http.Client{Timeout: time.Duration(cfg.YTDLP.TimeoutSeconds) * time.Second})
	translator := deepseek.New(deepseek.Config{
		BaseURL: cfg.DeepSeek.BaseURL, APIKey: cfg.DeepSeek.APIKey,
		Model: cfg.DeepSeek.Model, MaxRetries: cfg.DeepSeek.MaxRetries,
	},
		&http.Client{Timeout: time.Duration(cfg.DeepSeek.TimeoutSeconds) * time.Second})
	transcriber := groq.New(groq.Config{BaseURL: cfg.Groq.BaseURL, APIKey: cfg.Groq.APIKey, Model: cfg.Groq.Model},
		&http.Client{Timeout: time.Duration(cfg.Groq.TimeoutSeconds) * time.Second})
	pronunciationAssessor := azure.New(azure.Config{Endpoint: cfg.AzureSpeech.Endpoint, APIKey: cfg.AzureSpeech.APIKey},
		&http.Client{Timeout: time.Duration(cfg.AzureSpeech.TimeoutSeconds) * time.Second})

	result := useCases{
		user:       user.New(userRepo, languageRepo),
		language:   language.New(languageRepo),
		level:      level.New(levelRepo),
		topic:      topic.New(topicRepo),
		channel:    channel.New(channelRepo),
		adminUser:  adminuser.New(adminUserRepo),
		video:      video.New(videoRepo, userRepo, youtubeProvider, quota),
		caption:    caption.New(captionRepo, videoRepo, languageRepo, userRepo, youtubeProvider, transcriber, translator, cfg.DeepSeek.MaxBatchItems),
		vocabulary: vocabulary.New(dictionaryRepo, userRepo, languageRepo, translator),
		learning:   learningcontent.New(learningContentRepo, dictionaryRepo, captionRepo, videoRepo, userRepo, languageRepo, translator),
		shadowing:  shadowing.New(shadowingRepo, pronunciationAssessor, quota),
	}
	if cfg.IAP.Enabled {
		httpClient := &http.Client{Timeout: time.Duration(cfg.IAP.TimeoutSeconds) * time.Second}
		var appleClient iapGateway.AppleGateway
		if cfg.IAP.AppleEnabled {
			client, err := appleGateway.New(&appleGateway.Config{
				BaseURL:  cfg.IAP.AppleBaseURL,
				IssuerID: cfg.IAP.AppleIssuerID, KeyID: cfg.IAP.AppleKeyID, BundleID: cfg.IAP.AppleBundleID,
				AppAppleID: cfg.IAP.AppleAppID, Environment: cfg.IAP.AppleEnvironment,
				PrivateKeyPath: cfg.IAP.ApplePrivateKeyPath, RootCAPath: cfg.IAP.AppleRootCAPath,
			}, httpClient)
			if err != nil {
				return useCases{}, fmt.Errorf("initialize Apple IAP: %w", err)
			}
			appleClient = client
		}
		googleClient, err := googlePlayGateway.New(ctx, googlePlayGateway.Config{CredentialsFile: cfg.IAP.GoogleCredentialsFile})
		if err != nil {
			return useCases{}, fmt.Errorf("initialize Google Play IAP: %w", err)
		}
		iapRepo := persistIAPRepo.New(pg)
		result.iap = iapUseCase.New(appleClient, googleClient, iapRepo, iapRepo,
			iapUseCase.Config{GooglePackageName: cfg.IAP.GooglePackageName, RestorePolicy: cfg.IAP.RestorePolicy})
	}
	return result, nil
}

func initServers(cfg *config.Config, uc useCases, verifier *firebaseauth.Verifier, l logger.Interface) servers {
	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	restapi.NewRouter(httpServer.App, cfg, uc.user, uc.language, uc.level, uc.topic, uc.channel, uc.adminUser, uc.video, uc.caption, uc.vocabulary, uc.learning, uc.shadowing, uc.iap, verifier, l)

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

	uc, err := initUseCases(ctx, cfg, pg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - initUseCases: %w", err))
	}
	s := initServers(cfg, uc, firebaseVerifier, l)
	s.startServers()
	s.waitForShutdown(l)
}
