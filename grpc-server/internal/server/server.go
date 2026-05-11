package server

import (
	"fmt"
	"net"

	"buf.build/go/protovalidate"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"grpc-server/internal/config"
	"grpc-server/internal/requestid"
)

type Server struct {
	cfg    *config.Config
	grpc   *grpc.Server
	health *health.Server
	lis    net.Listener
}

func New(cfg *config.Config) (*Server, error) {
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return nil, err
	}

	validator, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("protovalidate: %w", err)
	}

	g := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			requestid.UnaryServerInterceptor(),
			recoveryInterceptor(),
			loggingInterceptor(),
			validateInterceptor(validator),
		),
	)

	healthSrv := health.NewServer()
	healthv1.RegisterHealthServer(g, healthSrv)
	healthSrv.SetServingStatus("", healthv1.HealthCheckResponse_SERVING)

	if cfg.ReflectionEnabled {
		reflection.Register(g)
	}

	return &Server{cfg: cfg, grpc: g, health: healthSrv, lis: lis}, nil
}

// GRPC returns the underlying *grpc.Server so each service package can register itself.
func (s *Server) GRPC() *grpc.Server { return s.grpc }

// Health returns the gRPC health server so callers can flip serving status.
func (s *Server) Health() *health.Server { return s.health }

func (s *Server) Serve() error {
	log.Info().Str("addr", s.lis.Addr().String()).Msg("grpc listening")
	return s.grpc.Serve(s.lis)
}

// GracefulStop flips health to NOT_SERVING (so readiness probes drain
// connections) before draining in-flight RPCs.
func (s *Server) GracefulStop() {
	s.health.SetServingStatus("", healthv1.HealthCheckResponse_NOT_SERVING)
	s.grpc.GracefulStop()
}
