package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"grpc-server/internal/config"
	"grpc-server/internal/greeter"
	applog "grpc-server/internal/logger"
	"grpc-server/internal/server"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	applog.Setup(cfg.LogLevel, cfg.LogPretty)

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("create server")
	}

	// Register services here — one line each. Each package wires its own
	// repository + client + service + handler behind interfaces.
	greeter.Register(srv.GRPC())
	// user.Register(srv.GRPC())
	// order.Register(srv.GRPC())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Serve(); err != nil {
			log.Fatal().Err(err).Msg("serve")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutdown signal received")

	done := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(cfg.ShutdownTimeout):
		log.Warn().Msg("graceful shutdown timed out")
	}
	log.Info().Msg("grpc server stopped")
}
