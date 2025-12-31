package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	fmt.Println("=== Subgraph Example ===")

	// 1. Define Main Graph
	main := graph.NewStateGraph[map[string]any]()

	// 2. Define Subgraph 1 (Validation)
	validationSubgraph := graph.NewStateGraph[map[string]any]()
	validationSubgraph.AddNode("check_format", "check_format", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("[Validation] Checking format...")
		return map[string]any{"format_ok": true}, nil
	})
	validationSubgraph.AddNode("sanitize", "sanitize", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("[Validation] Sanitizing input...")
		return map[string]any{"sanitized": true}, nil
	})
	validationSubgraph.SetEntryPoint("check_format")
	validationSubgraph.AddEdge("check_format", "sanitize")
	validationSubgraph.AddEdge("sanitize", graph.END)

	// 3. Define Subgraph 2 (Processing)
	processingSubgraph := graph.NewStateGraph[map[string]any]()
	processingSubgraph.AddNode("transform", "transform", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("[Processing] Transforming data...")
		return map[string]any{"transformed": true}, nil
	})
	processingSubgraph.AddNode("enrich", "enrich", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("[Processing] Enriching data...")
		return map[string]any{"enriched": true}, nil
	})
	processingSubgraph.SetEntryPoint("transform")
	processingSubgraph.AddEdge("transform", "enrich")
	processingSubgraph.AddEdge("enrich", graph.END)

	// 4. Add Nodes to Main Graph
	main.AddNode("receive", "receive", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("[Main] Received request")
		return state, nil
	})

	// Add subgraphs
	// Use AddSubgraph with generic types
	graph.AddSubgraph(main, "validation", validationSubgraph,
		func(s map[string]any) map[string]any { return s },
		func(s map[string]any) map[string]any { return s })

	graph.AddSubgraph(main, "processing", processingSubgraph,
		func(s map[string]any) map[string]any { return s },
		func(s map[string]any) map[string]any { return s })

	main.AddNode("finalize", "finalize", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("[Main] Finalizing response")
		return map[string]any{"status": "completed"}, nil
	})

	// 5. Connect Main Graph
	main.SetEntryPoint("receive")
	main.AddEdge("receive", "validation")
	main.AddEdge("validation", "processing")
	main.AddEdge("processing", "finalize")
	main.AddEdge("finalize", graph.END)

	// 6. Compile and Run
	runnable, err := main.Compile()
	if err != nil {
		panic(err)
	}

	fmt.Println("Running combined workflow...")
	res, err := runnable.Invoke(context.Background(), map[string]any{"input": "data"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Final State: %v\n", res)

	// Example of using CreateSubgraph builder
	fmt.Println("\n=== CreateSubgraph Builder Example ===")
	g2 := graph.NewStateGraph[map[string]any]()
	graph.CreateSubgraph(g2, "dynamic_sub", func(sg *graph.StateGraph[map[string]any]) error {
		sg.AddNode("step1", "step1", func(ctx context.Context, state map[string]any) (map[string]any, error) {
			return map[string]any{"dynamic_step1": true}, nil
		})
		sg.AddNode("step2", "step2", func(ctx context.Context, state map[string]any) (map[string]any, error) {
			return map[string]any{"dynamic_step2": true}, nil
		})
		sg.SetEntryPoint("step1")
		sg.AddEdge("step1", "step2")
		sg.AddEdge("step2", graph.END)
		return nil
	},
		func(s map[string]any) map[string]any { return s },
		func(s map[string]any) map[string]any { return s })

	g2.SetEntryPoint("dynamic_sub")
	g2.AddEdge("dynamic_sub", graph.END)

	r2, _ := g2.Compile()
	res2, _ := r2.Invoke(context.Background(), map[string]any{})
	fmt.Printf("Dynamic Subgraph Result: %v\n", res2)
}
