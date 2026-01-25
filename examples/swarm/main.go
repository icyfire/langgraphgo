package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
)

// Swarm style: Multiple specialized agents that can hand off to each other directly.
// This is different from Supervisor style where a central node routes.
// Here, nodes themselves decide next step.

func main() {
	// Define the graph
	workflow := graph.NewStateGraph[map[string]any]()

	// Schema: shared state
	schema := graph.NewMapSchema()
	schema.RegisterReducer("history", graph.AppendReducer)
	workflow.SetSchema(schema)

	// Agent 1: Triage
	workflow.AddNode("Triage", "Triage", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("[Triage] analyzing request...")
		return map[string]any{
				"history": []string{"Triage reviewed request"},
				"intent":  "research", // Simplified logic: always determine research needed
			},
			nil
	})

	// Agent 2: Researcher
	workflow.AddNode("Researcher", "Researcher", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("[Researcher] conducting research...")
		return map[string]any{
				"history": []string{"Researcher gathered data"},
				"data":    "Some facts found",
			},
			nil
	})

	// Agent 3: Writer
	workflow.AddNode("Writer", "Writer", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("[Writer] writing report...")
		data, _ := state["data"].(string)
		return map[string]any{
				"history": []string{"Writer created report"},
				"report":  fmt.Sprintf("Report based on %s", data),
			},
			nil
	})

	// Define Handoffs (Edges)
	workflow.SetEntryPoint("Triage")

	// Triage decides where to go
	workflow.AddConditionalEdge("Triage", func(ctx context.Context, state map[string]any) string {
		intent, _ := state["intent"].(string)
		if intent == "research" {
			return "Researcher"
		}
		if intent == "write" {
			return "Writer"
		}
		return graph.END
	})

	// Researcher hands off to Writer
	workflow.AddEdge("Researcher", "Writer")

	// Writer finishes
	workflow.AddEdge("Writer", graph.END)

	// Compile
	app, err := workflow.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// Execute
	fmt.Println("---" + " Starting Swarm ---")
	initialState := map[string]any{
		"history": []string{},
	}

	result, err := app.Invoke(context.Background(), initialState)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Final State: %v\n", result)
}
