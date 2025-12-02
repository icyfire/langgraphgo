package main

import (
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	// Check for API key
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("Please set OPENAI_API_KEY environment variable")
	}
	// Check for API Base if using DeepSeek (optional but recommended for non-OpenAI)
	if os.Getenv("OPENAI_API_BASE") == "" {
		fmt.Println("Warning: OPENAI_API_BASE not set. Defaulting to OpenAI. If using DeepSeek, set this to their API URL.")
	}

	query := "What are the latest advancements in solid state batteries?"
	if len(os.Args) > 1 {
		query = os.Args[1]
	}

	fmt.Printf("Starting Deer-Flow Research Agent with query: %s\n", query)

	graph, err := NewGraph()
	if err != nil {
		log.Fatalf("Failed to create graph: %v", err)
	}

	initialState := &State{
		Request: Request{
			Query: query,
		},
	}

	result, err := graph.Invoke(context.Background(), initialState)
	if err != nil {
		log.Fatalf("Graph execution failed: %v", err)
	}

	finalState := result.(*State)
	fmt.Println("\n=== Final Report ===")
	fmt.Println(finalState.FinalReport)
}
