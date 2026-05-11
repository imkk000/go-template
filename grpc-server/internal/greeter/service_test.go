package greeter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Greet_HappyPath(t *testing.T) {
	repo := &mockRepository{countResult: 3}
	client := &mockClient{out: "Hola, Alice!"}
	svc := NewService(repo, client)

	msg, count, err := svc.Greet(context.Background(), "Alice", "es")

	require.NoError(t, err)
	assert.Equal(t, "Hola, Alice!", msg)
	assert.Equal(t, 3, count)
	assert.Equal(t, []string{"Alice"}, repo.recordedNames)
	require.Len(t, client.inputs, 1)
	assert.Equal(t, "Hello, Alice!", client.inputs[0].text)
	assert.Equal(t, "es", client.inputs[0].lang)
}

func TestService_Greet_EmptyName(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockClient{})

	_, _, err := svc.Greet(context.Background(), "", "en")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidName)
}

func TestService_Greet_ClientFailureWrapsError(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockClient{err: errBoom})

	_, _, err := svc.Greet(context.Background(), "Bob", "en")

	require.Error(t, err)
	assert.ErrorIs(t, err, errBoom)
}

func TestService_Greet_RepositoryRecordFails(t *testing.T) {
	repo := &mockRepository{recordErr: errBoom}
	svc := NewService(repo, &mockClient{})

	_, _, err := svc.Greet(context.Background(), "Bob", "en")

	require.Error(t, err)
	assert.ErrorIs(t, err, errBoom)
}
