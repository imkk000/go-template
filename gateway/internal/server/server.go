package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"

	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/middleware"
)

type Server struct {
	engine *gin.Engine
}

func New(cfg *config.Config, mux *runtime.ServeMux) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	r.GET("/healthz", handler.Health)

	if cfg.SwaggerEnabled {
		if err := mountSwagger(r, cfg.SwaggerSpec); err != nil {
			log.Warn().Err(err).Str("spec", cfg.SwaggerSpec).Msg("swagger disabled: cannot read spec")
		} else {
			log.Info().Str("path", "/docs/").Msg("swagger UI mounted")
		}
	}

	// Anything that isn't an explicit Gin route falls through to the
	// grpc-gateway mux. Auth (when enabled) gates the mux but not /healthz or /docs.
	var fallback []gin.HandlerFunc
	if cfg.AuthEnabled {
		fallback = append(fallback, middleware.Auth(middleware.AuthOptions{
			JWTKey:     cfg.AuthJWTKey,
			APIURL:     cfg.AuthAPIURL,
			APITimeout: cfg.AuthAPITimeout,
		}))
	}
	fallback = append(fallback, gin.WrapH(mux))
	r.NoRoute(fallback...)

	return &Server{engine: r}
}

func (s *Server) Handler() http.Handler { return s.engine }
