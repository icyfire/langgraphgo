package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a new state graph with typed state
	g := graph.NewStateGraph[map[string]any]()

	// Define Schema
	// Using map schema where "steps" accumulates values
	schema := graph.NewMapSchema()
	schema.RegisterReducer("steps", graph.AppendReducer)
	g.SetSchema(schema)

	g.AddNode("node_a", "node_a", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"steps": "A"}, nil
	})

	g.AddNode("node_b", "node_b", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"steps": "B"}, nil
	})

	g.AddNode("node_c", "node_c", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"steps": "C"}, nil
	})

	// Linear chain
	g.SetEntryPoint("node_a")
	g.AddEdge("node_a", "node_b")
	g.AddEdge("node_b", "node_c")
	g.AddEdge("node_c", graph.END)

	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	res, err := runnable.Invoke(context.Background(), map[string]any{"steps": []string{}})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Steps: %v\n", res["steps"])
}
