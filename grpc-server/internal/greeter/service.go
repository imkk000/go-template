package greeter

import (
	"context"
	"errors"
	"fmt"
)

// ErrInvalidName is returned when the caller omits a required name.
// Handlers map it to codes.InvalidArgument.
var ErrInvalidName = errors.New("name is required")

// Service is the business-logic layer. Handlers call this through the interface.
type Service interface {
	Greet(ctx context.Context, name, lang string) (message string, count int, err error)
}

type service struct {
	repo   Repository
	client Client
}

func NewService(repo Repository, client Client) Service {
	return &service{repo: repo, client: client}
}

func (s *service) Greet(ctx context.Context, name, lang string) (string, int, error) {
	if name == "" {
		return "", 0, ErrInvalidName
	}
	base := fmt.Sprintf("Hello, %s!", name)
	msg, err := s.client.Translate(ctx, base, lang)
	if err != nil {
		return "", 0, fmt.Errorf("translate: %w", err)
	}
	if err := s.repo.RecordGreeting(ctx, name); err != nil {
		return "", 0, fmt.Errorf("record greeting: %w", err)
	}
	count, err := s.repo.CountGreetings(ctx, name)
	if err != nil {
		return "", 0, fmt.Errorf("count greetings: %w", err)
	}
	return msg, count, nil
}
