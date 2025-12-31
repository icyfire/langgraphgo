package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// 1. Create a child graph
	child := graph.NewStateGraph[map[string]any]()
	child.AddNode("child_process", "child_process", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["child_visited"] = true
		return state, nil
	})
	child.SetEntryPoint("child_process")
	child.AddEdge("child_process", graph.END)

	// 2. Create parent graph
	parent := graph.NewStateGraph[map[string]any]()

	parent.AddNode("start", "start", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["parent_start"] = true
		return state, nil
	})

	// Add child graph as a node
	// Note: Generic AddSubgraph requires converters.
	// Since both are map[string]any, we use identity.
	err := graph.AddSubgraph(parent, "child_graph", child,
		func(s map[string]any) map[string]any { return s },
		func(s map[string]any) map[string]any { return s })
	if err != nil {
		panic(err)
	}

	parent.AddNode("end", "end", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		state["parent_end"] = true
		return state, nil
	})

	parent.SetEntryPoint("start")
	parent.AddEdge("start", "child_graph")
	parent.AddEdge("child_graph", "end")
	parent.AddEdge("end", graph.END)

	runnable, err := parent.Compile()
	if err != nil {
		panic(err)
	}

	res, err := runnable.Invoke(context.Background(), map[string]any{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result: %v\n", res)
}
