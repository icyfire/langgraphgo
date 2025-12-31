package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a new state graph with typed state map[string]any
	g := graph.NewStateGraph[map[string]any]()

	// Define schema with custom reducer for "tags"
	schema := graph.NewMapSchema()
	// Using generic AppendReducer
	schema.RegisterReducer("tags", graph.AppendReducer)
	g.SetSchema(schema)

	g.AddNode("start", "start", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"tags": []string{"initial"}}, nil
	})

	g.AddNode("tagger_a", "tagger_a", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"tags": []string{"A"}}, nil
	})

	g.AddNode("tagger_b", "tagger_b", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"tags": []string{"B"}}, nil
	})

	// Parallel execution for taggers
	g.SetEntryPoint("start")
	g.AddEdge("start", "tagger_a")
	g.AddEdge("start", "tagger_b")
	g.AddEdge("tagger_a", graph.END)
	g.AddEdge("tagger_b", graph.END)

	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	// Execute
	res, err := runnable.Invoke(context.Background(), map[string]any{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result tags: %v\n", res["tags"])
}
