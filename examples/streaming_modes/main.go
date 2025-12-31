package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a streaming graph
	g := graph.NewStreamingStateGraph[map[string]any]()

	// Define nodes
	g.AddNode("step_1", "step_1", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		time.Sleep(500 * time.Millisecond)
		return map[string]any{"step_1": "completed"}, nil
	})

	g.AddNode("step_2", "step_2", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		time.Sleep(500 * time.Millisecond)
		return map[string]any{"step_2": "completed"}, nil
	})

	g.SetEntryPoint("step_1")
	g.AddEdge("step_1", "step_2")
	g.AddEdge("step_2", graph.END)

	// 1. Stream Mode: Updates (Default)
	fmt.Println("=== Streaming Updates ===")
	g.SetStreamConfig(graph.StreamConfig{Mode: graph.StreamModeUpdates, BufferSize: 10})
	runnable, _ := g.CompileStreaming()

	updates := runnable.Stream(context.Background(), map[string]any{})
	for event := range updates.Events {
		fmt.Printf("Event: %s, Node: %s, State: %v\n", event.Event, event.NodeName, event.State)
	}

	// 2. Stream Mode: Values (Full State)
	fmt.Println("\n=== Streaming Values ===")
	g.SetStreamConfig(graph.StreamConfig{Mode: graph.StreamModeValues, BufferSize: 10})
	runnable, _ = g.CompileStreaming()

	values := runnable.Stream(context.Background(), map[string]any{})
	for event := range values.Events {
		fmt.Printf("Event: %s, State: %v\n", event.Event, event.State)
	}

	// 3. Stream Mode: Debug (All Events)
	fmt.Println("\n=== Streaming Debug ===")
	g.SetStreamConfig(graph.StreamConfig{Mode: graph.StreamModeDebug, BufferSize: 10})
	runnable, _ = g.CompileStreaming()

	debug := runnable.Stream(context.Background(), map[string]any{})
	for event := range debug.Events {
		fmt.Printf("[%s] %s: %v\n", event.Timestamp.Format(time.StampMilli), event.Event, event.NodeName)
	}
}
