package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"gateway/internal/config"
	"gateway/internal/grpcclient"
	"gateway/internal/handler"
	applog "gateway/internal/logger"
	"gateway/internal/server"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	applog.Setup(cfg.LogLevel, cfg.LogPretty)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn, err := grpcclient.Dial(ctx, cfg.GRPCBackendAddr, cfg.GRPCDialTimeout)
	if err != nil {
		log.Fatal().Err(err).Str("addr", cfg.GRPCBackendAddr).Msg("dial grpc backend")
	}
	defer conn.Close()

	mux, err := handler.NewMux(ctx, conn)
	if err != nil {
		log.Fatal().Err(err).Msg("build grpc-gateway mux")
	}

	srv := server.New(cfg, mux)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("gateway listening")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown")
	}
	log.Info().Msg("gateway stopped")
}
