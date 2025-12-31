package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a new state graph with map state
	g := graph.NewStateGraph[map[string]any]()

	g.AddNode("process", "process", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		// Access configuration from context
		config := graph.GetConfig(ctx)
		limit := 5 // Default
		if config != nil && config.Configurable != nil {
			if val, ok := config.Configurable["limit"].(int); ok {
				limit = val
			}
		}

		fmt.Printf("Processing with limit: %d\n", limit)
		state["processed"] = true
		state["limit_used"] = limit
		return state, nil
	})

	g.SetEntryPoint("process")
	g.AddEdge("process", graph.END)

	runnable, err := g.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// Run with limit 3 (via config)
	config := &graph.Config{
		Configurable: map[string]any{"limit": 3},
	}
	res, _ := runnable.InvokeWithConfig(context.Background(), map[string]any{"input": "start"}, config)
	fmt.Printf("Result with limit 3: %v\n", res)

	// Run with limit 10 (via config)
	config2 := &graph.Config{
		Configurable: map[string]any{"limit": 10},
	}
	res2, _ := runnable.InvokeWithConfig(context.Background(), map[string]any{"input": "start"}, config2)
	fmt.Printf("Result with limit 10: %v\n", res2)
}
