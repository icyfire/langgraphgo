package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a listenable graph
	g := graph.NewListenableStateGraph[map[string]any]()

	// Define nodes
	processNode := g.AddNode("process", "process", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		time.Sleep(100 * time.Millisecond)
		return map[string]any{"processed": true}, nil
	})

	analyzeNode := g.AddNode("analyze", "analyze", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		time.Sleep(100 * time.Millisecond)
		return map[string]any{"analyzed": true}, nil
	})

	reportNode := g.AddNode("report", "report", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		time.Sleep(100 * time.Millisecond)
		return map[string]any{"reported": true}, nil
	})

	// Add global listener (logs everything)
	g.AddGlobalListener(graph.NewLoggingListener().WithLogLevel(graph.LogLevelInfo))

	// Add specific listener to process node (metrics)
	processNode.AddListener(graph.NewMetricsListener())

	// Add listener to analyze node
	analyzeNode.AddListener(graph.NewLoggingListener().WithLogLevel(graph.LogLevelDebug))

	// Add progress listener to report node
	reportNode.AddListener(graph.NewProgressListener().WithPrefix("📊"))

	// Define flow
	g.SetEntryPoint("process")
	g.AddEdge("process", "analyze")
	g.AddEdge("analyze", "report")
	g.AddEdge("report", graph.END)

	// Compile
	runnable, err := g.CompileListenable()
	if err != nil {
		panic(err)
	}

	// Run
	fmt.Println("Running graph with listeners...")
	_, err = runnable.Invoke(context.Background(), map[string]any{})
	if err != nil {
		panic(err)
	}
}
