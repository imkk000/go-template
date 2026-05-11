// Package requestid carries an x-request-id value through gRPC metadata and
// the request context, so logs and downstream gRPC calls share a correlation
// id with the upstream HTTP gateway.
package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Header is the metadata / HTTP header key. It must be lowercase for gRPC.
const Header = "x-request-id"

type ctxKey struct{}

// FromContext returns the request id stored on ctx, or "" if none.
func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}

// WithValue attaches a request id to ctx.
func WithValue(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// UnaryServerInterceptor extracts x-request-id from incoming metadata (or
// generates one), stores it on the context for handlers/services to read,
// and echoes it back to the caller in the response header.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		id := fromIncoming(ctx)
		if id == "" {
			id = generate()
		}
		ctx = WithValue(ctx, id)
		_ = grpc.SetHeader(ctx, metadata.Pairs(Header, id))
		return handler(ctx, req)
	}
}

// UnaryClientInterceptor copies the request id from the calling context into
// the outgoing gRPC metadata, so an upstream correlation id flows through
// nested gRPC calls.
//
// Use it when constructing an outbound gRPC client:
//
//	conn, _ := grpc.NewClient(addr,
//	    grpc.WithTransportCredentials(insecure.NewCredentials()),
//	    grpc.WithChainUnaryInterceptor(requestid.UnaryClientInterceptor()),
//	)
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if id := FromContext(ctx); id != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, Header, id)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func fromIncoming(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get(Header)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func generate() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(b[:])
}
