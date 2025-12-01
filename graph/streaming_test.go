package graph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnhancedStreaming(t *testing.T) {
	g := NewStreamingMessageGraph()

	g.AddNode("A", func(ctx context.Context, state interface{}) (interface{}, error) {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		// Simulate LLM call via callback
		config := GetConfig(ctx)
		if config != nil {
			for _, cb := range config.Callbacks {
				cb.OnLLMStart(ctx, nil, []string{"prompt"}, "run-1", nil, nil, nil)
				time.Sleep(5 * time.Millisecond)
				cb.OnLLMEnd(ctx, "response", "run-1")
			}
		}

		return "A done", nil
	})

	g.SetEntryPoint("A")
	g.AddEdge("A", END)

	runnable, err := g.CompileStreaming()
	assert.NoError(t, err)

	streamResult := runnable.Stream(context.Background(), nil)
	defer streamResult.Cancel()

	var events []StreamEvent
	for event := range streamResult.Events {
		events = append(events, event)
	}

	// Verify events
	// Expected: ChainStart, NodeStart(A), LLMStart, LLMEnd, NodeEnd(A), ChainEnd
	// Note: NodeStart/End are emitted by ListenableNode logic.
	// ChainStart/End are emitted by StreamingListener via callbacks?
	// Wait, StreamingListener is added as a NodeListener, so it gets Node events.
	// It is also added as a CallbackHandler in config.
	// Who calls OnChainStart?
	// In graph.go InvokeWithConfig calls OnChainStart.
	// ListenableRunnable.InvokeWithConfig does NOT call OnChainStart yet.
	// I need to update ListenableRunnable.InvokeWithConfig to call callbacks if I want Chain events.

	// Let's check what we have.
	// NodeStart/End come from ListenableNode.Execute calling NotifyListeners.
	// LLMStart/End come from manual calls in Node A.

	hasNodeStart := false
	hasNodeEnd := false
	hasLLMStart := false
	hasLLMEnd := false

	for _, e := range events {
		switch e.Event {
		case NodeEventStart:
			if e.NodeName == "A" {
				hasNodeStart = true
			}
		case NodeEventComplete:
			if e.NodeName == "A" {
				hasNodeEnd = true
			}
		case EventLLMStart:
			hasLLMStart = true
		case EventLLMEnd:
			hasLLMEnd = true
		}
	}

	assert.True(t, hasNodeStart, "Should have NodeStart event")
	assert.True(t, hasNodeEnd, "Should have NodeEnd event")
	assert.True(t, hasLLMStart, "Should have LLMStart event")
	assert.True(t, hasLLMEnd, "Should have LLMEnd event")
}
