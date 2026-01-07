package graph

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateState(t *testing.T) {
	g := NewCheckpointableStateGraph[map[string]any]()

	// Setup schema with reducer
	schema := NewMapSchema()
	schema.RegisterReducer("count", func(curr, new any) (any, error) {
		if curr == nil {
			return new, nil
		}
		return curr.(int) + new.(int), nil
	})
	g.SetSchema(schema)

	g.AddNode("A", "A", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"count": 1}, nil
	})
	g.SetEntryPoint("A")
	g.AddEdge("A", END)

	runnable, err := g.CompileCheckpointable()
	assert.NoError(t, err)

	// 1. Run initial graph with thread_id in config
	ctx := context.Background()
	threadID := runnable.GetExecutionID()
	config := &Config{
		Configurable: map[string]any{
			"thread_id": threadID,
		},
	}
	res, err := runnable.InvokeWithConfig(ctx, map[string]any{"count": 10}, config)
	assert.NoError(t, err)

	mRes := res
	assert.Equal(t, 11, mRes["count"]) // 10 + 1 = 11

	// 2. Update state manually (Human-in-the-loop)
	// We want to add 5 to the count
	// config already has thread_id from previous Invoke
	updateConfig := &Config{
		Configurable: map[string]any{
			"thread_id": threadID,
		},
	}

	newConfig, err := runnable.UpdateState(ctx, updateConfig, "human", map[string]any{"count": 5})
	assert.NoError(t, err)
	assert.NotEmpty(t, newConfig.Configurable["checkpoint_id"])

	// 3. Verify state is updated
	snapshot, err := runnable.GetState(ctx, newConfig)
	assert.NoError(t, err)

	mSnap := snapshot.Values.(map[string]any)
	// Should be 11 (previous) + 5 (update) = 16
	assert.Equal(t, 16, mSnap["count"])
}
