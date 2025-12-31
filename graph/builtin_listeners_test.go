package graph_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

const (
	step2Result = "step2_result"
)

func TestProgressListener_OnNodeEvent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewProgressListenerWithWriter(&buf).
		WithTiming(false). // Disable timing for predictable output
		WithPrefix("🔄")

	ctx := context.Background()

	// Test start event
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	output := buf.String()

	if !strings.Contains(output, "🔄 Starting test_node") {
		t.Errorf("Expected start message, got: %s", output)
	}

	// Test complete event
	buf.Reset()
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", nil, nil)
	output = buf.String()

	if !strings.Contains(output, "✅ test_node completed") {
		t.Errorf("Expected complete message, got: %s", output)
	}

	// Test error event
	buf.Reset()
	listener.OnNodeEvent(ctx, graph.NodeEventError, "test_node", nil, fmt.Errorf("test error"))
	output = buf.String()

	if !strings.Contains(output, "❌ test_node failed: test error") {
		t.Errorf("Expected error message, got: %s", output)
	}
}

func TestProgressListener_CustomSteps(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewProgressListenerWithWriter(&buf).
		WithTiming(false)

	// Set custom step message
	listener.SetNodeStep("process", "Analyzing data")

	ctx := context.Background()
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "process", nil, nil)

	output := buf.String()
	if !strings.Contains(output, "🔄 Analyzing data") {
		t.Errorf("Expected custom message, got: %s", output)
	}
}

func TestProgressListener_WithDetails(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewProgressListenerWithWriter(&buf).
		WithTiming(false).
		WithDetails(true)

	ctx := context.Background()
	state := map[string]any{"key": "value"}

	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", state, nil)

	output := buf.String()
	if !strings.Contains(output, "State: map[key:value]") {
		t.Errorf("Expected state details, got: %s", output)
	}
}

func TestLoggingListener_OnNodeEvent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "[TEST] ", 0) // No timestamp for predictable output

	listener := graph.NewLoggingListenerWithLogger(logger).
		WithLogLevel(graph.LogLevelDebug)

	ctx := context.Background()

	// Test different event types
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventError, "test_node", nil, fmt.Errorf("test error"))

	output := buf.String()

	if !strings.Contains(output, "[TEST] START test_node") {
		t.Errorf("Expected start log, got: %s", output)
	}

	if !strings.Contains(output, "[TEST] COMPLETE test_node") {
		t.Errorf("Expected complete log, got: %s", output)
	}

	if !strings.Contains(output, "[TEST] ERROR test_node: test error") {
		t.Errorf("Expected error log, got: %s", output)
	}
}

func TestLoggingListener_LogLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "[TEST] ", 0)

	listener := graph.NewLoggingListenerWithLogger(logger).
		WithLogLevel(graph.LogLevelError) // Only error level and above

	ctx := context.Background()

	// These should be filtered out
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventProgress, "test_node", nil, nil)

	// This should be logged
	listener.OnNodeEvent(ctx, graph.NodeEventError, "test_node", nil, fmt.Errorf("test error"))

	output := buf.String()

	if strings.Contains(output, "START") || strings.Contains(output, "PROGRESS") {
		t.Errorf("Expected debug/info messages to be filtered, got: %s", output)
	}

	if !strings.Contains(output, "ERROR test_node") {
		t.Errorf("Expected error message, got: %s", output)
	}
}

func TestLoggingListener_WithState(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "[TEST] ", 0)

	listener := graph.NewLoggingListenerWithLogger(logger).
		WithState(true)

	ctx := context.Background()
	state := map[string]any{"state": testState}

	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", state, nil)

	output := buf.String()
	// State is now map[string]any{"state": "test_state"}
	if !strings.Contains(output, "State: map[state:test_state]") {
		t.Errorf("Expected state in log, got: %s", output)
	}
}

func TestMetricsListener_OnNodeEvent(t *testing.T) {
	t.Parallel()

	listener := graph.NewMetricsListener()
	ctx := context.Background()

	// Simulate node execution
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	time.Sleep(1 * time.Millisecond) // Small delay to measure
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", nil, nil)

	// Check metrics
	executions := listener.GetNodeExecutions()
	if executions["test_node"] != 1 {
		t.Errorf("Expected 1 execution, got %d", executions["test_node"])
	}

	avgDurations := listener.GetNodeAverageDuration()
	if _, exists := avgDurations["test_node"]; !exists {
		t.Error("Expected duration to be recorded")
	}

	if listener.GetTotalExecutions() != 1 {
		t.Errorf("Expected 1 total execution, got %d", listener.GetTotalExecutions())
	}
}

func TestMetricsListener_ErrorTracking(t *testing.T) {
	t.Parallel()

	listener := graph.NewMetricsListener()
	ctx := context.Background()

	// Simulate node execution with error
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "error_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventError, "error_node", nil, fmt.Errorf("test error"))

	// Check error metrics
	errors := listener.GetNodeErrors()
	if errors["error_node"] != 1 {
		t.Errorf("Expected 1 error, got %d", errors["error_node"])
	}
}

func TestMetricsListener_PrintSummary(t *testing.T) {
	t.Parallel()

	listener := graph.NewMetricsListener()
	ctx := context.Background()

	// Generate some metrics
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "node1", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "node1", nil, nil)

	listener.OnNodeEvent(ctx, graph.NodeEventStart, "node2", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventError, "node2", nil, fmt.Errorf("error"))

	var buf bytes.Buffer
	listener.PrintSummary(&buf)

	output := buf.String()

	if !strings.Contains(output, "Node Execution Metrics") {
		t.Error("Expected metrics header")
	}

	if !strings.Contains(output, "Total Executions: 2") {
		t.Error("Expected total executions count")
	}

	if !strings.Contains(output, "node1: 1") {
		t.Error("Expected node1 execution count")
	}

	if !strings.Contains(output, "node2: 1 errors") {
		t.Error("Expected node2 error count")
	}
}

func TestMetricsListener_Reset(t *testing.T) {
	t.Parallel()

	listener := graph.NewMetricsListener()
	ctx := context.Background()

	// Generate some metrics
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test_node", nil, nil)
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "test_node", nil, nil)

	// Verify metrics exist
	if listener.GetTotalExecutions() != 1 {
		t.Error("Expected metrics to be recorded")
	}

	// Reset and verify metrics are cleared
	listener.Reset()

	if listener.GetTotalExecutions() != 0 {
		t.Error("Expected metrics to be reset")
	}

	executions := listener.GetNodeExecutions()
	if len(executions) != 0 {
		t.Error("Expected executions to be cleared")
	}
}

func TestChatListener_OnNodeEvent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewChatListenerWithWriter(&buf).
		WithTime(false) // Disable time for predictable output

	ctx := context.Background()

	// Test start event
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "process", nil, nil)
	output := buf.String()

	if !strings.Contains(output, "🤖 Starting process...") {
		t.Errorf("Expected start message, got: %s", output)
	}

	// Test complete event
	buf.Reset()
	listener.OnNodeEvent(ctx, graph.NodeEventComplete, "process", nil, nil)
	output = buf.String()

	if !strings.Contains(output, "✅ process finished") {
		t.Errorf("Expected complete message, got: %s", output)
	}
}

func TestChatListener_CustomMessages(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewChatListenerWithWriter(&buf).
		WithTime(false)

	// Set custom message
	listener.SetNodeMessage("analyze", "Analyzing your document")

	ctx := context.Background()
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "analyze", nil, nil)

	output := buf.String()
	if !strings.Contains(output, "Analyzing your document") {
		t.Errorf("Expected custom message, got: %s", output)
	}
}

func TestChatListener_WithTime(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	listener := graph.NewChatListenerWithWriter(&buf).
		WithTime(true)

	ctx := context.Background()
	listener.OnNodeEvent(ctx, graph.NodeEventStart, "test", nil, nil)

	output := buf.String()
	// Should contain timestamp in format [HH:MM:SS]
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Errorf("Expected timestamp in output, got: %s", output)
	}
}

// Integration test with actual graph execution
func TestBuiltinListeners_Integration(t *testing.T) {
	t.Parallel()

	// Create graph
	g := graph.NewListenableStateGraph[map[string]any]()

	node1 := g.AddNode("step1", "step1", func(_ context.Context, state map[string]any) (map[string]any, error) {
		// Return updated state map
		return map[string]any{"state": "step1_result"}, nil
	})

	node2 := g.AddNode("step2", "step2", func(_ context.Context, state map[string]any) (map[string]any, error) {
		// Return updated state map
		return map[string]any{"state": step2Result}, nil
	})

	g.AddEdge("step1", "step2")
	g.AddEdge("step2", graph.END)
	g.SetEntryPoint("step1")

	// Add listeners
	var progressBuf, logBuf, chatBuf bytes.Buffer

	progressListener := graph.NewProgressListenerWithWriter(&progressBuf).WithTiming(false)
	logListener := graph.NewLoggingListenerWithLogger(log.New(&logBuf, "[GRAPH] ", 0))
	chatListener := graph.NewChatListenerWithWriter(&chatBuf).WithTime(false)
	metricsListener := graph.NewMetricsListener()

	// Builtin listeners might be implementing NodeListener[any] (untyped)?
	// If Builtin listeners are not generic, we might need adapters.
	// But let's check builtin_listeners.go.
	// They implement NodeListener (which was untyped interface).
	// Now NodeListener is NodeListener[S].
	// If builtin listeners implement OnNodeEvent(..., state any, ...), they only satisfy NodeListener[any].
	// But our graph is NodeListener[map[string]any].
	// So we can't directly add them if they don't implement the generic interface with map[string]any.
	// Unless I make builtin listeners generic or use an adapter.
	//
	// However, NodeListener[S] interface has OnNodeEvent(..., state S, ...).
	// If builtin listener has OnNodeEvent(..., state any, ...), it DOES NOT match NodeListener[S] in Go because methods must match exactly.
	//
	// So I probably broke builtin listeners.
	// I should check builtin_listeners.go.
	// If they use `any`, they are compatible with `NodeListener[any]`.
	// But I am using `ListenableStateGraph[map[string]any]`.
	// Its `AddListener` expects `NodeListener[map[string]any]`.
	//
	// I need to genericize builtin listeners OR create an adapter.
	// Given the scope, creating an adapter is easier.
	// But "unify to generics" suggests making them generic.
	//
	// Let's assume for now I will fix builtin_listeners.go later or they are already fixed (I didn't touch them).
	// I didn't touch `builtin_listeners.go`.
	// So they are broken.
	//
	// I will update the test to assume I will fix them.
	// OR I can use an adapter in the test.
	//
	// Actually, `NewProgressListenerWithWriter` returns `*ProgressListener`.
	// `ProgressListener` has `OnNodeEvent(..., state any, ...)`.
	// I need `OnNodeEvent(..., state map[string]any, ...)`.
	//
	// I will make `BuiltinListeners` generic-friendly by adding a method or type alias?
	// Or just make them implement `OnNodeEvent` with `any`.
	// But `ListenableNode[S]` calls `OnNodeEvent(..., state S, ...)`.
	// If `S` is `map[string]any`, it passes a map.
	// `ProgressListener` expects `any`.
	// `map` satisfies `any`.
	// But the INTERFACE `NodeListener[map[string]any]` requires method taking `map[string]any`.
	// `ProgressListener` method takes `any`.
	// Does `func(any)` satisfy `interface { func(map[string]any) }`? NO.
	//
	// So I MUST genericize builtin listeners or use adapters.
	// `ProgressListener` should probably be `ProgressListener[S any]`.
	//
	// I will check `builtin_listeners.go` in next step.
	// For now, I will comment out the listener addition in test or wrap them.
	//
	// Better: I will create a simple adapter in the test for now.
	// type anyAdapter struct { L graph.NodeListener[any] }
	// func (a anyAdapter) OnNodeEvent(..., state map[string]any, ...) { a.L.OnNodeEvent(..., state, ...) }
	//
	// Wait, `NodeListener` is generic now. `graph.NodeListener` refers to `graph.NodeListener[S]`.
	// `builtin_listeners.go` defines `ProgressListener`. It implements `OnNodeEvent`.
	// But `NodeListener` definition in `graph` package CHANGED.
	// So `builtin_listeners.go` might fail to compile if it refers to `NodeListener`.
	// `builtin_listeners.go` imports `graph`.
	// It likely uses `NodeListener` interface?
	//
	// Let's check `builtin_listeners.go`.
	// `type ProgressListener struct ...`
	// `func (l *ProgressListener) OnNodeEvent(...)`
	// It probably doesn't explicitly say "implements NodeListener".
	// But to be used as one, it must satisfy the interface.
	//
	// I will verify this in next steps.
	// For the test, I will update it to use the new graph API.

	// Adapter for builtin listeners.
	// Builtin listeners seems to have OnNodeEvent(..., map[string]any, ...).
	mapAdapter := func(l interface {
		OnNodeEvent(context.Context, graph.NodeEvent, string, map[string]any, error)
	}) graph.NodeListener[map[string]any] {
		return graph.NodeListenerFunc[map[string]any](func(ctx context.Context, e graph.NodeEvent, n string, s map[string]any, err error) {
			l.OnNodeEvent(ctx, e, n, s, err)
		})
	}

	// BUT, if ProgressListener takes map[string]any, it ALREADY implements NodeListener[map[string]any]!
	// So I shouldn't need an adapter?
	//
	// Why did I think I needed an adapter?
	// Because originally it was untyped (any).
	// If it was untyped, it implemented NodeListener (untyped).
	// Now NodeListener is generic.
	// NodeListener[map[string]any] requires OnNodeEvent(..., map[string]any, ...).
	//
	// If ProgressListener has OnNodeEvent(..., any, ...), it does NOT implement NodeListener[map[string]any].
	// The error message:
	// cannot use progressListener ... as interface{... any ...} value in argument to adapter:
	// *graph.ProgressListener does not implement interface{... any ...} (wrong type for method OnNodeEvent)
	// have OnNodeEvent(..., map[string]any, ...)
	// want OnNodeEvent(..., any, ...)
	//
	// This CONFIRMS ProgressListener has `map[string]any` in its signature.
	// How? I must have changed builtin_listeners.go?
	// Or maybe it was always map[string]any?
	// If so, I can just use it directly!
	//
	// Let's try adding directly first. If that fails, I'll know why.
	// But wait, if it has `map[string]any`, why did it work with `StateGraphUntyped` (which dealt with `any`)?
	// `StateGraphUntyped` listeners took `any`.
	// If ProgressListener took `map[string]any`, it wouldn't satisfy `NodeListener` (untyped).
	//
	// Unless... I updated `builtin_listeners.go` in a previous step without realizing?
	// I don't recall editing it.
	//
	// Maybe `NodeEvent` definition change affected it? No.
	//
	// Maybe I should just check `builtin_listeners.go` content.
	// But I am in the middle of `replace`.
	//
	// I will update the test to use an adapter that matches what the compiler says ProgressListener has.
	// If compiler says it has `map[string]any`, then my adapter should accept `map[string]any`.
	// And if it has `map[string]any`, it implements `NodeListener[map[string]any]`.
	// So I can cast/use directly?
	//
	// Let's try using `mapAdapter` defined above.

	node1.AddListener(mapAdapter(progressListener))
	node1.AddListener(mapAdapter(logListener))
	node1.AddListener(mapAdapter(chatListener))
	node1.AddListener(mapAdapter(metricsListener))

	node2.AddListener(mapAdapter(progressListener))
	node2.AddListener(mapAdapter(logListener))
	node2.AddListener(mapAdapter(chatListener))
	node2.AddListener(mapAdapter(metricsListener))

	// Execute graph - pass input as map to avoid wrapping
	runnable, err := g.CompileListenable()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	ctx := context.Background()
	result, err := runnable.Invoke(ctx, map[string]any{"state": "input"})
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// result is already map[string]any
	actualResult := result["state"]
	if actualResult != step2Result {
		t.Errorf("Expected 'step2_result', got %v", actualResult)
	}

	// Wait for async listeners
	time.Sleep(50 * time.Millisecond)

	// Check outputs
	progressOutput := progressBuf.String()
	if !strings.Contains(progressOutput, "Starting step1") {
		t.Errorf("Progress listener should show step1, got: %s", progressOutput)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "START step1") {
		t.Errorf("Log listener should show START step1, got: %s", logOutput)
	}

	chatOutput := chatBuf.String()
	if !strings.Contains(chatOutput, "🤖 Starting step1") {
		t.Errorf("Chat listener should show start message, got: %s", chatOutput)
	}

	// Check metrics
	executions := metricsListener.GetNodeExecutions()
	if executions["step1"] != 1 || executions["step2"] != 1 {
		t.Errorf("Expected 1 execution each, got: %v", executions)
	}

	if metricsListener.GetTotalExecutions() != 2 {
		t.Errorf("Expected 2 total executions, got %d", metricsListener.GetTotalExecutions())
	}
}
