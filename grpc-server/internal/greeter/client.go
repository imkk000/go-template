package greeter

import "context"

// Client is an outbound dependency the service calls (e.g. a translation API
// or another microservice). Replace NoopClient with a real implementation.
type Client interface {
	Translate(ctx context.Context, text, lang string) (string, error)
}

type NoopClient struct{}

func NewClient() *NoopClient { return &NoopClient{} }

func (NoopClient) Translate(_ context.Context, text, _ string) (string, error) {
	return text, nil
}
