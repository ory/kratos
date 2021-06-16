package corp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type contextMock struct {
	ContextNoOp
}

func (*contextMock) ContextualizeTableName(_ context.Context, name string) string {
	return "foo"
}

func TestSetContextualizer(t *testing.T) {
	ctx := context.Background()
	require.Panics(t, func() {
		ContextualizeTableName(ctx, "")
	})

	SetContextualizer(new(ContextNoOp))
	assert.Equal(t, "bar", ContextualizeTableName(ctx, "bar"))

	SetContextualizer(new(contextMock))
	assert.Equal(t, "foo", ContextualizeTableName(ctx, "bar"))

	SetContextualizer(new(ContextNoOp))
	assert.Equal(t, "foo", ContextualizeTableName(ctx, "bar"))
}
