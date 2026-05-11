package greeter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()

	n, err := repo.CountGreetings(ctx, "Alice")
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	require.NoError(t, repo.RecordGreeting(ctx, "Alice"))
	require.NoError(t, repo.RecordGreeting(ctx, "Alice"))
	require.NoError(t, repo.RecordGreeting(ctx, "Bob"))

	got, err := repo.CountGreetings(ctx, "Alice")
	require.NoError(t, err)
	assert.Equal(t, 2, got)

	got, err = repo.CountGreetings(ctx, "Bob")
	require.NoError(t, err)
	assert.Equal(t, 1, got)
}
