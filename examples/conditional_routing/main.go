package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
)

type Task struct {
	ID       string
	Priority string
	Content  string
}

func main() {
	// Create graph
	g := graph.NewStateGraph[map[string]any]()

	g.AddNode("router", "router", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return state, nil
	})

	g.AddNode("urgent_handler", "urgent_handler", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"status": "handled_urgent"}, nil
	})

	g.AddNode("normal_handler", "normal_handler", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"status": "handled_normal"}, nil
	})

	g.AddNode("batch_handler", "batch_handler", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		return map[string]any{"status": "handled_batch"}, nil
	})

	g.SetEntryPoint("router")

	// Add conditional edge based on "priority"
	g.AddConditionalEdge("router", func(ctx context.Context, state map[string]any) string {
		priority, _ := state["priority"].(string)
		switch priority {
		case "high":
			return "urgent_handler"
		case "low":
			return "batch_handler"
		default:
			return "normal_handler"
		}
	})

	g.AddEdge("urgent_handler", graph.END)
	g.AddEdge("normal_handler", graph.END)
	g.AddEdge("batch_handler", graph.END)

	// Compile
	runnable, err := g.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// 1. High priority task
	fmt.Println("--- High Priority Task ---")
	task1 := map[string]any{"id": "1", "priority": "high", "content": "System down"}
	result, err := runnable.Invoke(context.Background(), task1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result: %s\n", result["status"])

	// 2. Low priority task
	fmt.Println("\n--- Low Priority Task ---")
	task2 := map[string]any{"id": "2", "priority": "low", "content": "Update docs"}
	result, err = runnable.Invoke(context.Background(), task2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result: %s\n", result["status"])

	// 3. Normal priority task
	fmt.Println("\n--- Normal Priority Task ---")
	task3 := map[string]any{"id": "3", "priority": "medium", "content": "Bug fix"}
	result, err = runnable.Invoke(context.Background(), task3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result: %s\n", result["status"])
}
