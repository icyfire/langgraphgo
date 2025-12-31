package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a checkpointable graph
	g := graph.NewCheckpointableStateGraph[map[string]any]()

	// 1. Initial State Node
	g.AddNode("A", "A", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("Executing Node A")
		return map[string]any{"trace": []string{"A"}}, nil
	})

	// 2. Second Node
	g.AddNode("B", "B", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("Executing Node B")
		trace := state["trace"].([]string)
		return map[string]any{"trace": append(trace, "B")}, nil
	})

	g.SetEntryPoint("A")
	g.AddEdge("A", "B")
	g.AddEdge("B", graph.END)

	// Configure in-memory store
	store := graph.NewMemoryCheckpointStore()
	g.SetCheckpointConfig(graph.CheckpointConfig{
		Store:    store,
		AutoSave: true,
	})

	runnable, err := g.CompileCheckpointable()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Run first time
	fmt.Println("--- First Run ---")
	res, err := runnable.Invoke(ctx, map[string]any{"input": "start"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result 1: %v\n", res)

	// Wait for async saves
	time.Sleep(100 * time.Millisecond)

	// List checkpoints
	checkpoints, _ := runnable.ListCheckpoints(ctx)
	if len(checkpoints) == 0 {
		log.Fatal("No checkpoints found")
	}

	// "Time Travel": Load a previous checkpoint (e.g. after Node A)
	// We want to find the checkpoint where NodeName is "A"
	var targetCP *graph.Checkpoint
	for _, cp := range checkpoints {
		if cp.NodeName == "A" {
			targetCP = cp
			break
		}
	}

	if targetCP != nil {
		fmt.Println("\n--- Time Travel (Resuming from Node A) ---")
		// Resume execution from this checkpoint
		// ResumeFromCheckpoint loads the state.
		// To continue execution, we usually need to know where to go next.
		// Or if we just want to inspect state.
		// If we want to branch off:
		// We use InvokeWithConfig with ResumeFrom (this logic is app specific usually)

		// But here let's just inspect
		loadedState, err := runnable.LoadCheckpoint(ctx, targetCP.ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Traveled back to state after Node A: %v\n", loadedState.State)

		// Branch off: Run Node B again but with modified state?
		// Or just re-run B.
		config := &graph.Config{
			Configurable: map[string]any{
				"thread_id":     runnable.GetExecutionID(),
				"checkpoint_id": targetCP.ID,
			},
			ResumeFrom: []string{"B"},
		}

		// Let's modify state "in place" conceptually (forking)
		// by passing modified state to Invoke
		forkedState := loadedState.State.(map[string]any)
		forkedState["forked"] = true

		resFork, err := runnable.InvokeWithConfig(ctx, forkedState, config)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Forked Result: %v\n", resFork)
	}
}
