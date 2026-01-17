package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a new state graph
	g := graph.NewStateGraph[any]()

	// Set up state merger to properly merge map updates
	// This ensures that Command.Update merges fields instead of replacing the entire state
	g.SetStateMerger(func(ctx context.Context, current any, results []any) (any, error) {
		state := current.(map[string]any)
		if state == nil {
			state = make(map[string]any)
		}
		for _, res := range results {
			resMap := res.(map[string]any)
			for k, v := range resMap {
				state[k] = v
			}
		}
		return state, nil
	})

	g.AddNode("router", "router", func(ctx context.Context, state any) (any, error) {
		m := state.(map[string]any)
		val := m["value"].(int)
		if val > 10 {
			return &graph.Command{
				Goto:   "end_high",
				Update: map[string]any{"path": "high"},
			}, nil
		}
		return &graph.Command{
			Goto:   "process",
			Update: map[string]any{"path": "normal"},
		}, nil
	})

	g.AddNode("process", "process", func(ctx context.Context, state any) (any, error) {
		m := state.(map[string]any)
		val := m["value"].(int)
		return map[string]any{"value": val * 2}, nil
	})

	g.AddNode("end_high", "end_high", func(ctx context.Context, state any) (any, error) {
		m := state.(map[string]any)
		val := m["value"].(int)
		return map[string]any{"value": val + 100}, nil
	})

	g.SetEntryPoint("router")
	g.AddEdge("process", graph.END)
	g.AddEdge("end_high", graph.END)

	runnable, err := g.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// Test 1: Normal path
	fmt.Println("--- Test 1: Normal Path ---")
	res1, _ := runnable.Invoke(context.Background(), map[string]any{"value": 5})
	fmt.Printf("Result (value=5): %v\n", res1)

	// Test 2: High path
	fmt.Println("\n--- Test 2: High Path ---")
	res2, _ := runnable.Invoke(context.Background(), map[string]any{"value": 15})
	fmt.Printf("Result (value=15): %v\n", res2)
}
