package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// NewMux builds a grpc-gateway ServeMux. Add one RegisterXxxServiceHandler
// call per proto service — routes come straight from the `google.api.http`
// annotations on each rpc.
//
//	import greeterv1 "<your-proto-module>/gen/go/greeter/v1"
//	import userv1    "<your-proto-module>/gen/go/user/v1"
//
//	if err := greeterv1.RegisterGreeterServiceHandler(ctx, mux, conn); err != nil {
//	    return nil, err
//	}
//	if err := userv1.RegisterUserServiceHandler(ctx, mux, conn); err != nil {
//	    return nil, err
//	}
//
// Your proto module must be generated with `protoc-gen-grpc-gateway`
// (it emits the `*.pb.gw.go` file containing those Register functions).
func NewMux(ctx context.Context, conn *grpc.ClientConn) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(incomingHeaderMatcher),
	)
	_ = ctx
	_ = conn
	return mux, nil
}

// incomingHeaderMatcher forwards a small allowlist of HTTP headers into the
// gRPC metadata so backends can read Authorization, request id, etc.
func incomingHeaderMatcher(key string) (string, bool) {
	switch http.CanonicalHeaderKey(key) {
	case "Authorization", "X-Request-Id":
		return key, true
	}
	return runtime.DefaultHeaderMatcher(key)
}
