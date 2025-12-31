package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	// Create a state graph with map state
	g := graph.NewStateGraph[map[string]any]()

	g.AddNode("ask_name", "ask_name", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		// This simulates an interrupt.
		// graph.Interrupt pauses execution and waits for input.
		// When execution resumes, it returns the provided value.
		answer, err := graph.Interrupt(ctx, "What is your name?")
		if err != nil {
			return nil, err
		}

		// Use the answer
		return map[string]any{
				"name":    answer,
				"message": fmt.Sprintf("Hello, %s!", answer),
			},
			nil
	})

	g.SetEntryPoint("ask_name")
	g.AddEdge("ask_name", graph.END)

	runnable, err := g.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// 1. Initial Run
	fmt.Println("--- 1. Initial Execution ---")
	// We pass empty map as initial state
	_, err = runnable.Invoke(context.Background(), map[string]any{})

	// Check if the execution was interrupted
	var graphInterrupt *graph.GraphInterrupt
	if errors.As(err, &graphInterrupt) {
		fmt.Printf("Graph interrupted at node: %s\n", graphInterrupt.Node)
		fmt.Printf("Interrupt Value (Query): %v\n", graphInterrupt.InterruptValue)

		// Simulate getting input from a user
		userInput := "Alice"
		fmt.Printf("\n[User Input]: %s\n", userInput)

		// 2. Resume Execution
		fmt.Println("\n--- 2. Resuming Execution ---")

		// We provide the user input as ResumeValue in the config
		config := &graph.Config{
			ResumeValue: userInput,
		}

		// Re-run the graph. The 'ask_name' node will run again,
		// but this time graph.Interrupt() will return 'userInput' immediately.
		// Note: We need to pass the same initial state (or the state at interruption if we had it, but here it's stateless start)
		res, err := runnable.InvokeWithConfig(context.Background(), map[string]any{}, config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Final Result: %v\n", res)

	} else if err != nil {
		log.Fatalf("Execution failed: %v", err)
	} else {
		fmt.Println("Execution finished without interrupt.")
	}
}
