package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a streaming graph for a text processing pipeline
	g := graph.NewStreamingStateGraph[map[string]any]()

	// Define nodes
	analyze := g.AddNode("analyze", "analyze", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		input := state["input"].(string)
		time.Sleep(200 * time.Millisecond)
		return map[string]any{"analysis": fmt.Sprintf("Length: %d", len(input))}, nil
	})

	enhance := g.AddNode("enhance", "enhance", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		input := state["input"].(string)
		time.Sleep(300 * time.Millisecond)
		return map[string]any{"enhanced": strings.ToUpper(input)}, nil
	})

	summarize := g.AddNode("summarize", "summarize", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		analysis := state["analysis"].(string)
		enhanced := state["enhanced"].(string)
		time.Sleep(200 * time.Millisecond)
		return map[string]any{"summary": fmt.Sprintf("%s -> %s", analysis, enhanced)}, nil
	})

	// Add progress listeners to nodes
	progressListener := graph.NewProgressListener().WithTiming(true)
	analyze.AddListener(progressListener)
	enhance.AddListener(progressListener)
	summarize.AddListener(progressListener)

	// Define flow
	g.SetEntryPoint("analyze")
	g.AddEdge("analyze", "enhance")
	g.AddEdge("enhance", "summarize")
	g.AddEdge("summarize", graph.END)

	// Compile
	runnable, err := g.CompileStreaming()
	if err != nil {
		panic(err)
	}

	// Execute with streaming
	fmt.Println("Starting pipeline...")
	input := map[string]any{"input": "hello world"}

	// Stream returns a channel wrapper
	streamResult := runnable.Stream(context.Background(), input)

	// Process events
	for event := range streamResult.Events {
		if event.Error != nil {
			fmt.Printf("Error: %v\n", event.Error)
			return
		}

		// We can react to specific events if needed,
		// but the ProgressListener attached to nodes handles printing.
		// Here we just consume the channel to let it run.

		if event.Event == graph.NodeEventComplete {
			// Maybe show partial results
			if event.NodeName == "enhance" {
				fmt.Printf(">> Intermediate result: %v\n", event.State)
			}
		}
	}

	fmt.Println("Pipeline finished.")
}
