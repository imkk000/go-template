package greeter

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	greeterv1 "grpc-server/gen/go/greeter/v1"
)

// Handler implements the generated greeterv1.GreeterServiceServer interface.
// It holds the Service interface — never a concrete type.
type Handler struct {
	greeterv1.UnimplementedGreeterServiceServer
	svc Service
}

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) SayHello(ctx context.Context, req *greeterv1.SayHelloRequest) (*greeterv1.SayHelloResponse, error) {
	msg, count, err := h.svc.Greet(ctx, req.GetName(), req.GetLang())
	if err != nil {
		return nil, toStatus(err)
	}
	return &greeterv1.SayHelloResponse{
		Message: msg,
		Count:   int32(count),
	}, nil
}

// Register wires repository + client + service + handler in one place,
// then registers the gRPC service. Call this once from main.go.
func Register(s *grpc.Server) {
	repo := NewRepository()
	client := NewClient()
	svc := NewService(repo, client)
	greeterv1.RegisterGreeterServiceServer(s, NewHandler(svc))
}

func toStatus(err error) error {
	switch {
	case errors.Is(err, ErrInvalidName):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
