package requestid

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestFromContext_AbsentReturnsEmpty(t *testing.T) {
	assert.Empty(t, FromContext(context.Background()))
}

func TestServerInterceptor_PropagatesIncomingHeader(t *testing.T) {
	in := metadata.NewIncomingContext(context.Background(), metadata.Pairs(Header, "abc123"))

	var seen string
	handler := func(ctx context.Context, _ any) (any, error) {
		seen = FromContext(ctx)
		return nil, nil
	}

	_, err := UnaryServerInterceptor()(in, nil, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, "abc123", seen)
}

func TestServerInterceptor_GeneratesWhenMissing(t *testing.T) {
	var seen string
	handler := func(ctx context.Context, _ any) (any, error) {
		seen = FromContext(ctx)
		return nil, nil
	}

	_, err := UnaryServerInterceptor()(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Len(t, seen, 32, "generated id should be 16 hex bytes")
}

func TestClientInterceptor_AppendsToOutgoingMetadata(t *testing.T) {
	ctx := WithValue(context.Background(), "abc123")

	var got metadata.MD
	invoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		got, _ = metadata.FromOutgoingContext(ctx)
		return nil
	}

	err := UnaryClientInterceptor()(ctx, "/svc/Method", nil, nil, nil, invoker)
	require.NoError(t, err)
	assert.Equal(t, []string{"abc123"}, got.Get(Header))
}

func TestClientInterceptor_NoIdNoMetadata(t *testing.T) {
	var got metadata.MD
	invoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		got, _ = metadata.FromOutgoingContext(ctx)
		return nil
	}

	err := UnaryClientInterceptor()(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	require.NoError(t, err)
	assert.Empty(t, got.Get(Header))
}
