package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
)

// State represents the workflow state
type State struct {
	Input    string
	Approved bool
	Output   string
}

func main() {
	// Create a new graph with typed state
	g := graph.NewStateGraph[State]()

	// Define nodes
	g.AddNode("process_request", "process_request", func(ctx context.Context, state State) (State, error) {
		fmt.Printf("[Process] Processing request: %s\n", state.Input)
		state.Output = "Processed: " + state.Input
		return state, nil
	})

	g.AddNode("human_approval", "human_approval", func(ctx context.Context, state State) (State, error) {
		if state.Approved {
			fmt.Println("[Human] Request APPROVED.")
			state.Output += " (Approved)"
		} else {
			fmt.Println("[Human] Request REJECTED.")
			state.Output += " (Rejected)"
		}
		return state, nil
	})

	g.AddNode("finalize", "finalize", func(ctx context.Context, state State) (State, error) {
		fmt.Printf("[Finalize] Final output: %s\n", state.Output)
		return state, nil
	})

	// Define edges
	g.SetEntryPoint("process_request")
	g.AddEdge("process_request", "human_approval")
	g.AddEdge("human_approval", "finalize")
	g.AddEdge("finalize", graph.END)

	// Compile the graph
	runnable, err := g.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// Initial state
	initialState := State{
		Input:    "Deploy to Production",
		Approved: false,
	}

	// 1. Run with InterruptBefore "human_approval"
	fmt.Println("=== Starting Workflow (Phase 1) ===")
	config := &graph.Config{
		InterruptBefore: []string{"human_approval"},
	}

	// Note: InvokeWithConfig returns (S, error) where S is State struct.
	// But interrupt error is separate.
	// If interrupted, it returns the state at interrupt and the GraphInterrupt error.
	res, err := runnable.InvokeWithConfig(context.Background(), initialState, config)

	// We expect an interrupt error
	if err != nil {
		var interrupt *graph.GraphInterrupt
		if errors.As(err, &interrupt) {
			fmt.Printf("Workflow interrupted at node: %s\n", interrupt.Node)
			fmt.Printf("Current State: %+v\n", interrupt.State)
		} else {
			log.Fatalf("Unexpected error: %v", err)
		}
	} else {
		// If it didn't interrupt, that's unexpected for this example
		fmt.Printf("Workflow completed without interrupt: %+v\n", res)
		return
	}

	// Simulate Human Interaction
	fmt.Println("\n=== Human Interaction ===")
	fmt.Println("Reviewing request...")
	fmt.Println("Approving request...")

	// Update state to reflect approval
	var interrupt *graph.GraphInterrupt
	errors.As(err, &interrupt)

	// Since we use typed graph, interrupt.State is 'any' but contains 'State' struct.
	currentState := interrupt.State.(State)
	currentState.Approved = true // Human approves

	// 2. Resume execution
	fmt.Println("\n=== Resuming Workflow (Phase 2) ===")
	resumeConfig := &graph.Config{
		ResumeFrom: []string{interrupt.Node}, // Resume from the interrupted node
	}

	finalRes, err := runnable.InvokeWithConfig(context.Background(), currentState, resumeConfig)
	if err != nil {
		log.Fatalf("Error resuming workflow: %v", err)
	}

	fmt.Printf("Workflow completed successfully.\n")
	fmt.Printf("Final Result: %+v\n", finalRes)
}
