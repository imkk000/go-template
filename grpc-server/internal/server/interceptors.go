package server

import (
	"context"
	"runtime/debug"
	"time"

	"buf.build/go/protovalidate"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"grpc-server/internal/requestid"
)

func loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		evt := log.Info()
		if err != nil {
			evt = log.Error().Err(err).Str("code", status.Code(err).String())
		}
		if rid := requestid.FromContext(ctx); rid != "" {
			evt = evt.Str("request_id", rid)
		}
		evt.
			Str("method", info.FullMethod).
			Dur("latency", time.Since(start)).
			Msg("grpc request")
		return resp, err
	}
}

// recoveryInterceptor catches panics from any downstream interceptor or
// handler, logs the stack with the request id, and returns codes.Internal.
func recoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Interface("panic", r).
					Bytes("stack", debug.Stack()).
					Str("method", info.FullMethod).
					Str("request_id", requestid.FromContext(ctx)).
					Msg("panic recovered")
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// validateInterceptor runs protovalidate rules declared in the .proto file
// against every incoming request message. Rule violations become
// codes.InvalidArgument.
func validateInterceptor(v protovalidate.Validator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if msg, ok := req.(proto.Message); ok {
			if err := v.Validate(msg); err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
		return handler(ctx, req)
	}
}
