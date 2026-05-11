package greeter

import (
	"context"
	"errors"
)

// Hand-rolled mocks — no extra dependencies needed.

type mockRepository struct {
	recordErr     error
	countResult   int
	countErr      error
	recordedNames []string
}

func (m *mockRepository) RecordGreeting(_ context.Context, name string) error {
	m.recordedNames = append(m.recordedNames, name)
	return m.recordErr
}

func (m *mockRepository) CountGreetings(_ context.Context, _ string) (int, error) {
	return m.countResult, m.countErr
}

type mockClient struct {
	out    string
	err    error
	inputs []translateCall
}

type translateCall struct{ text, lang string }

func (m *mockClient) Translate(_ context.Context, text, lang string) (string, error) {
	m.inputs = append(m.inputs, translateCall{text: text, lang: lang})
	if m.err != nil {
		return "", m.err
	}
	if m.out != "" {
		return m.out, nil
	}
	return text, nil
}

type mockService struct {
	msg   string
	count int
	err   error
	calls []greetCall
}

type greetCall struct{ name, lang string }

func (m *mockService) Greet(_ context.Context, name, lang string) (string, int, error) {
	m.calls = append(m.calls, greetCall{name: name, lang: lang})
	return m.msg, m.count, m.err
}

var (
	_ Repository = (*mockRepository)(nil)
	_ Client     = (*mockClient)(nil)
	_ Service    = (*mockService)(nil)

	errBoom = errors.New("boom")
)
