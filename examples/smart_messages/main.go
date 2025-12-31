package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	g := graph.NewStateGraph[map[string]any]()

	g.AddNode("user_input", "user_input", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		// In a real app, this would get input from UI
		// Here we simulate it from initial state or hardcode
		return map[string]any{"user_query": "Hello"}, nil
	})

	g.AddNode("ai_response", "ai_response", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		query, _ := state["user_query"].(string)
		// Simulate smart message generation
		return map[string]any{"response": fmt.Sprintf("Echo: %s", query)}, nil
	})

	// Hypothetical "Smart Messages" logic where we might update previous messages in UI
	// This usually involves state management where messages have IDs
	g.AddNode("ai_update", "ai_update", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"response": "Updated: Echo Hello"}, nil
	})

	g.SetEntryPoint("user_input")
	g.AddEdge("user_input", "ai_response")
	g.AddEdge("ai_response", "ai_update")
	g.AddEdge("ai_update", graph.END)

	runnable, _ := g.Compile()
	res, _ := runnable.Invoke(context.Background(), map[string]any{})

	fmt.Printf("Final: %v\n", res)
}
