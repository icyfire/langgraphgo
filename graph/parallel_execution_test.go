package graph

import (
	"context"
	"maps"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// appendVisitors is a helper function that appends a node name to the visited list in the state.
func appendVisitors(state map[string]any, node string) []string {
	visited, ok := state["visited"].([]string)
	if !ok {
		visited = []string{}
	}
	return append(visited, node)
}

func TestParallelExecution_FanOut(t *testing.T) {
	g := NewStateGraph[map[string]any]()

	// Node A: Entry point
	g.AddNode("A", "A", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		newState := make(map[string]any)
		maps.Copy(newState, state)
		visited := appendVisitors(newState, "A")
		newState["visited"] = visited
		return newState, nil
	})

	// Node B: Branch 1
	g.AddNode("B", "B", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		newState := make(map[string]any)
		maps.Copy(newState, state)
		visited := appendVisitors(newState, "B")
		newState["visited"] = visited
		time.Sleep(10 * time.Millisecond) // Simulate work
		return newState, nil
	})

	// Node C: Branch 2
	g.AddNode("C", "C", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		newState := make(map[string]any)
		maps.Copy(newState, state)
		visited := appendVisitors(newState, "C")
		newState["visited"] = visited
		time.Sleep(10 * time.Millisecond) // Simulate work
		return newState, nil
	})

	// Node D: Join point
	g.AddNode("D", "D", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		newState := make(map[string]any)
		maps.Copy(newState, state)
		visited := appendVisitors(newState, "D")
		newState["visited"] = visited
		return newState, nil
	})

	g.SetEntryPoint("A")

	// Define Fan-Out: A -> B, A -> C
	g.AddEdge("A", "B")
	g.AddEdge("A", "C")

	// Define Fan-In: B -> D, C -> D
	g.AddEdge("B", "D")
	g.AddEdge("C", "D")

	g.AddEdge("D", END)

	// Set state merger for parallel execution
	g.SetStateMerger(func(ctx context.Context, current map[string]any, newStates []map[string]any) (map[string]any, error) {
		// Collect all visited nodes from all states
		visitedSet := make(map[string]bool)
		for _, s := range newStates {
			if v, ok := s["visited"].([]string); ok {
				for _, node := range v {
					visitedSet[node] = true
				}
			}
		}
		// Convert to sorted slice for deterministic output
		visited := make([]string, 0, len(visitedSet))
		for node := range visitedSet {
			visited = append(visited, node)
		}
		result := make(map[string]any)
		maps.Copy(result, current)
		result["visited"] = visited
		return result, nil
	})

	// Compile
	runnable, err := g.Compile()
	assert.NoError(t, err)

	// Execute
	initialState := map[string]any{
		"visited": []string{},
	}

	result, err := runnable.Invoke(context.Background(), initialState)
	assert.NoError(t, err)

	// Check results
	visited := result["visited"].([]string)
	assert.Contains(t, visited, "A")
	assert.Contains(t, visited, "D")

	// Both B and C should be visited
	hasB := false
	hasC := false
	for _, v := range visited {
		if v == "B" {
			hasB = true
		}
		if v == "C" {
			hasC = true
		}
	}

	assert.True(t, hasB, "Node B should be visited")
	assert.True(t, hasC, "Node C should be visited")
}
