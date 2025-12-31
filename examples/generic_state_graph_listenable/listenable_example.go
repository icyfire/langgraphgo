package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

// TypedState represents a strongly typed state
type TypedState struct {
	Count int
	Log   []string
}

func main() {
	// Create a typed graph
	g := graph.NewListenableStateGraph[TypedState]()

	// Add a node
	node := g.AddNode("increment", "Increment counter", func(ctx context.Context, state TypedState) (TypedState, error) {
		state.Count++
		state.Log = append(state.Log, "Incremented")
		return state, nil
	})

	// Add a typed listener using NodeListenerFunc
	listener := graph.NodeListenerFunc[TypedState](
		func(ctx context.Context, event graph.NodeEvent, nodeName string, state TypedState, err error) {
			fmt.Printf("[Listener] Event: %s, Node: %s, Count: %d\n", event, nodeName, state.Count)
		},
	)

	node.AddListener(listener)

	g.SetEntryPoint("increment")
	g.AddEdge("increment", graph.END)

	runnable, err := g.CompileListenable()
	if err != nil {
		panic(err)
	}

	initialState := TypedState{Count: 0, Log: []string{}}
	result, err := runnable.Invoke(context.Background(), initialState)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Final State: %+v\n", result)
}
