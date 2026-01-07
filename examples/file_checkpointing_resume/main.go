package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a temporary directory for checkpoints
	checkpointDir := "./checkpoints_resume"
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(checkpointDir) // Cleanup after run

	fmt.Printf("Using checkpoint directory: %s\n", checkpointDir)

	// Initialize FileCheckpointStore
	store, err := graph.NewFileCheckpointStore(checkpointDir)
	if err != nil {
		log.Fatalf("Failed to create checkpoint store: %v", err)
	}

	// Define a simplified setup function to create the graph logic
	createGraph := func() *graph.CheckpointableStateGraph[map[string]any] {
		g := graph.NewCheckpointableStateGraph[map[string]any]()

		g.AddNode("step1", "step1", func(ctx context.Context, state map[string]any) (map[string]any, error) {
			fmt.Println("  [EXEC] Running Step 1")
			state["step1"] = "done"
			return state, nil
		})

		g.AddNode("step2", "step2", func(ctx context.Context, state map[string]any) (map[string]any, error) {
			fmt.Println("  [EXEC] Running Step 2")
			state["step2"] = "done"
			return state, nil
		})

		g.AddNode("step3", "step3", func(ctx context.Context, state map[string]any) (map[string]any, error) {
			fmt.Println("  [EXEC] Running Step 3")
			state["step3"] = "done"
			return state, nil
		})

		g.AddEdge("step1", "step2")
		g.AddEdge("step2", "step3")
		g.AddEdge("step3", graph.END)
		g.SetEntryPoint("step1")
		return g
	}

	// define common config
	threadID := "resume_thread"
	baseConfig := graph.CheckpointConfig{
		Store:    store,
		AutoSave: true,
	}

	// ---------------------------------------------------------
	// PHASE 1: Run until interrupted (after Step 2)
	// ---------------------------------------------------------
	fmt.Println("\n--- PHASE 1: Running until interruption after Step 2 ---")

	g1 := createGraph()
	g1.SetCheckpointConfig(baseConfig)
	runnable1, err := g1.CompileCheckpointable()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	initialState := map[string]any{
		"input": "start",
	}

	// Config with interrupt
	config1 := &graph.Config{
		Configurable: map[string]any{"thread_id": threadID},
		// We interrupt AFTER step 2 runs.
		// The graph will stop before executing step 3.
		InterruptAfter: []string{"step2"},
	}

	res1, err := runnable1.InvokeWithConfig(ctx, initialState, config1)
	if err != nil {
		// We expect an interrupt error or a GraphInterrupt return if treated as error?
		if _, ok := err.(*graph.GraphInterrupt); ok {
			fmt.Printf("  [INFO] Graph interrupted as expected: %v\n", err)
		} else {
			log.Fatalf("Unexpected error in Phase 1: %v", err)
		}
	} else {
		// If it didn't return an error/interrupt, maybe it finished?
		fmt.Printf("  [WARN] Phase 1 finished without interrupt? Result: %v\n", res1)
	}

	// ---------------------------------------------------------
	// PHASE 2: Resume from the interrupted state
	// ---------------------------------------------------------
	fmt.Println("\n--- PHASE 2: Resuming from checkpoint ---")

	// 1. List checkpoints to find the latest state
	checkpoints, err := store.List(ctx, threadID)
	if err != nil {
		log.Fatal(err)
	}

	if len(checkpoints) == 0 {
		log.Fatal("No checkpoints found!")
	}

	// Sort by version (List implementation handles this but good to be sure or verify)
	// Get the latest checkpoint
	latestCP := checkpoints[len(checkpoints)-1]
	fmt.Printf("  [INFO] Resuming from checkpoint: ID=%s, Node=%s, Version=%d\n", latestCP.ID, latestCP.NodeName, latestCP.Version)
	fmt.Printf("  [INFO] State at checkpoint: %v\n", latestCP.State)

	// 2. Prepare for resume
	// We need to know where to resume FROM.
	// Since we interrupted AFTER step 2, we want to proceed to step 3.
	// Or, more accurately, we start execution.
	// We must pass the LAST state as initial state.
	// And we must tell the graph where to begin execution using `ResumeFrom`.
	// Since we interrupted after `step2`, the next logical step strictly defined by the graph is `step3`.
	// `ResumeFrom` overrides the entry point.

	g2 := createGraph()
	g2.SetCheckpointConfig(baseConfig)
	runnable2, err := g2.CompileCheckpointable()
	if err != nil {
		log.Fatal(err)
	}

	config2 := &graph.Config{
		Configurable: map[string]any{"thread_id": threadID},
		ResumeFrom:   []string{"step3"}, // Start directly at step 3
	}

	// Use the state from the checkpoint
	// The state in checkpoint is generic 'any'. We cast it to map[string]any.
	resumedState, ok := latestCP.State.(map[string]any)
	if !ok {
		// Try to handle map[string]any from JSON unmarshal
		if m, ok := latestCP.State.(map[string]any); ok {
			resumedState = m
		} else {
			log.Fatalf("Failed to cast checkpoint state to map[string]any: %T", latestCP.State)
		}
	}

	// Invoke
	res2, err := runnable2.InvokeWithConfig(ctx, resumedState, config2)
	if err != nil {
		log.Fatalf("Execution failed in Phase 2: %v", err)
	}

	fmt.Printf("  [INFO] Final Result: %v\n", res2)

	// Verify complete execution state
	finalMap := res2
	if finalMap["step1"] == "done" && finalMap["step2"] == "done" && finalMap["step3"] == "done" {
		fmt.Println("  [SUCCESS] Graph successfully resumed and completed all steps.")
	} else {
		fmt.Println("  [FAILURE] Final state missing steps.")
	}
}
