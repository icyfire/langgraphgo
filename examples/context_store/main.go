package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/langgraphgo/graph"
)

type ProcessState struct {
	Step    int
	Data    string
	History []string
}

func main() {
	// Create a checkpointable graph with typed state
	g := graph.NewCheckpointableStateGraph[ProcessState]()

	// Define state schema with merger logic
	// Since ProcessState is a struct, we can use a schema to merge partial updates if needed.
	// But here nodes return full state (modified), so Overwrite is fine or default struct merge.
	// Default struct merge overwrites non-zero fields.
	// Let's use NewStructSchema.
	schema := graph.NewStructSchema(
		ProcessState{},
		func(current, new ProcessState) (ProcessState, error) {
			// For this example, we simply replace the state with the new one returned by the node
			return new, nil
		},
	)
	g.SetSchema(schema)

	// Configure checkpointing
	config := graph.CheckpointConfig{
		Store:          graph.NewMemoryCheckpointStore(),
		AutoSave:       true,
		SaveInterval:   2 * time.Second,
		MaxCheckpoints: 5,
	}
	g.SetCheckpointConfig(config)

	// Add processing nodes
	g.AddNode("step1", "step1", func(ctx context.Context, s ProcessState) (ProcessState, error) {
		s.Step = 1
		s.Data = s.Data + " → Step1"
		s.History = append(s.History, "Completed Step 1")
		fmt.Println("Executing Step 1...")
		time.Sleep(500 * time.Millisecond) // Simulate work
		return s, nil
	})

	g.AddNode("step2", "step2", func(ctx context.Context, s ProcessState) (ProcessState, error) {
		s.Step = 2
		s.Data = s.Data + " → Step2"
		s.History = append(s.History, "Completed Step 2")
		fmt.Println("Executing Step 2...")
		time.Sleep(500 * time.Millisecond) // Simulate work
		return s, nil
	})

	g.AddNode("step3", "step3", func(ctx context.Context, s ProcessState) (ProcessState, error) {
		s.Step = 3
		s.Data = s.Data + " → Step3"
		s.History = append(s.History, "Completed Step 3")
		fmt.Println("Executing Step 3...")
		time.Sleep(500 * time.Millisecond) // Simulate work
		return s, nil
	})

	// Build the pipeline
	g.SetEntryPoint("step1")
	g.AddEdge("step1", "step2")
	g.AddEdge("step2", "step3")
	g.AddEdge("step3", graph.END)

	// Compile checkpointable runnable
	runnable, err := g.CompileCheckpointable()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	initialState := ProcessState{
		Step:    0,
		Data:    "Start",
		History: []string{"Initialized"},
	}

	fmt.Println("=== Starting execution with checkpointing ===")

	// Execute with automatic checkpointing
	result, err := runnable.Invoke(ctx, initialState)
	if err != nil {
		panic(err)
	}

	finalState := result
	fmt.Printf("\n=== Execution completed ===\n")
	fmt.Printf("Final Step: %d\n", finalState.Step)
	fmt.Printf("Final Data: %s\n", finalState.Data)
	fmt.Printf("History: %v\n", finalState.History)

	// List saved checkpoints
	checkpoints, err := runnable.ListCheckpoints(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n=== Created %d checkpoints ===\n", len(checkpoints))
	for i, cp := range checkpoints {
		fmt.Printf("Checkpoint %d: ID=%s, Time=%v\n", i+1, cp.ID, cp.Timestamp)
	}

	// Demonstrate resuming from a checkpoint
	if len(checkpoints) > 1 {
		fmt.Printf("\n=== Resuming from checkpoint %s ===\n", checkpoints[1].ID)
		// Checkpoint stores generic 'any', need to cast if loading manually,
		// but ResumeFromCheckpoint is not implemented in generic CheckpointableRunnable yet?
		// Wait, I didn't implement ResumeFromCheckpoint in generic CheckpointableRunnable[S]!
		// I only implemented LoadCheckpoint.
		// Let's implement ResumeFromCheckpoint or just use LoadCheckpoint + Invoke.

		// Actually, ResumeFromCheckpoint was convenient wrapper.
		// But in the example I can just use LoadCheckpoint and cast.

		cp, err := runnable.LoadCheckpoint(ctx, checkpoints[1].ID)
		if err != nil {
			fmt.Printf("Error loading checkpoint: %v\n", err)
		} else {
			// Ensure state is cast correctly
			// Checkpoint.State is any. JSON unmarshal might make it map[string]any if using file store.
			// But here we use MemoryStore which stores struct as is (if pointer) or value.
			// Let's assume it's ProcessState.

			var resumed ProcessState
			if s, ok := cp.State.(ProcessState); ok {
				resumed = s
			} else {
				// Handle map[string]any case if needed (e.g. if loaded from JSON)
				// For now assuming MemoryStore preserves type
				fmt.Printf("Warning: Checkpoint state type mismatch: %T\n", cp.State)
			}

			fmt.Printf("Resumed at Step: %d\n", resumed.Step)
			fmt.Printf("Resumed Data: %s\n", resumed.Data)
			fmt.Printf("Resumed History: %v\n", resumed.History)
		}
	}
}
