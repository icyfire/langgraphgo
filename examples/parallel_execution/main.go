package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a new state graph with typed state
	g := graph.NewStateGraph[map[string]any]()

	// Define Schema
	// Using map schema where "results" accumulates values
	schema := graph.NewMapSchema()
	schema.RegisterReducer("results", graph.AppendReducer)
	g.SetSchema(schema)

	// Define Nodes
	g.AddNode("start", "start", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("Starting execution...")
		return map[string]any{}, nil
	})

	g.AddNode("branch_a", "branch_a", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Branch A executed")
		return map[string]any{"results": "A"}, nil
	})

	g.AddNode("branch_b", "branch_b", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("Branch B executed")
		return map[string]any{"results": "B"}, nil
	})

	g.AddNode("branch_c", "branch_c", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		time.Sleep(150 * time.Millisecond)
		fmt.Println("Branch C executed")
		return map[string]any{"results": "C"}, nil
	})

	g.AddNode("aggregator", "aggregator", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		results := state["results"]
		fmt.Printf("Aggregated results: %v\n", results)
		return map[string]any{"final": "done"}, nil
	})

	// Define Graph Structure
	g.SetEntryPoint("start")

	// Fan-out from start to branches
	g.AddEdge("start", "branch_a")
	g.AddEdge("start", "branch_b")
	g.AddEdge("start", "branch_c")

	// Fan-in from branches to aggregator
	g.AddEdge("branch_a", "aggregator")
	g.AddEdge("branch_b", "aggregator")
	g.AddEdge("branch_c", "aggregator")

	g.AddEdge("aggregator", graph.END)

	// Compile
	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	// Execute
	initialState := map[string]any{
		"results": []string{},
	}

	res, err := runnable.Invoke(context.Background(), initialState)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Final state: %v\n", res)
}
