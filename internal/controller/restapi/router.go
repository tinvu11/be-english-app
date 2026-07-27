package restapi

import (
	"net/http"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/evrone/go-clean-template/config"
	_ "github.com/evrone/go-clean-template/docs" // Swagger docs.
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	v1 "github.com/evrone/go-clean-template/internal/controller/restapi/v1"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
)

// NewRouter -.
// Swagger spec:
//
//	@title       Go Clean Template API
//	@description Content catalog and translation API with user authentication
//	@version     1.0
//	@host        localhost:8080
//	@BasePath    /v1
//	@securityDefinitions.apikey BearerAuth
//	@in header
//	@name Authorization
func NewRouter(app *fiber.App, cfg *config.Config, t usecase.Translation, u usecase.User, c usecase.Content, verifier middleware.TokenVerifier, l logger.Interface) {
	// Options
	app.Use(middleware.Logger(l))
	app.Use(middleware.Recovery(l))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// Prometheus metrics
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New("my-service-name")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// Swagger
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// K8s probe
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })

	// Routers
	apiV1Group := app.Group("/v1")
	{
		if cfg.Tracing.Enabled {
			apiV1Group.Use(otelfiber.Middleware())
		}

		v1.NewRoutes(apiV1Group, t, u, c, verifier, l)
	}
}
