package greeter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	greeterv1 "grpc-server/gen/go/greeter/v1"
)

func TestHandler_SayHello_OK(t *testing.T) {
	svc := &mockService{msg: "Hello, Alice!", count: 5}
	h := NewHandler(svc)

	resp, err := h.SayHello(context.Background(), &greeterv1.SayHelloRequest{
		Name: "Alice",
		Lang: "en",
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello, Alice!", resp.GetMessage())
	assert.Equal(t, int32(5), resp.GetCount())
	require.Len(t, svc.calls, 1)
	assert.Equal(t, "Alice", svc.calls[0].name)
	assert.Equal(t, "en", svc.calls[0].lang)
}

func TestHandler_SayHello_InvalidNameMapsToInvalidArgument(t *testing.T) {
	svc := &mockService{err: ErrInvalidName}
	h := NewHandler(svc)

	_, err := h.SayHello(context.Background(), &greeterv1.SayHelloRequest{})

	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestHandler_SayHello_UnknownErrorMapsToInternal(t *testing.T) {
	svc := &mockService{err: errBoom}
	h := NewHandler(svc)

	_, err := h.SayHello(context.Background(), &greeterv1.SayHelloRequest{Name: "Alice"})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}
