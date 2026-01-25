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

// State defines the schema for the graph.
type State struct {
	History []string
	Intent  string
	Data    string
	Report  string
}

func main() {
	// Define the graph with typed State
	workflow := graph.NewStateGraph[State]()

	// Agent 1: Triage
	workflow.AddNode("Triage", "Triage", func(ctx context.Context, state State) (State, error) {
		fmt.Println("[Triage] analyzing request...")
		state.History = append(state.History, "Triage reviewed request")
		state.Intent = "research" // Simplified logic: always determine research needed
		return state, nil
	})

	// Agent 2: Researcher
	workflow.AddNode("Researcher", "Researcher", func(ctx context.Context, state State) (State, error) {
		fmt.Println("[Researcher] conducting research...")
		state.History = append(state.History, "Researcher gathered data")
		state.Data = "Some facts found"
		return state, nil
	})

	// Agent 3: Writer
	workflow.AddNode("Writer", "Writer", func(ctx context.Context, state State) (State, error) {
		fmt.Println("[Writer] writing report...")
		state.History = append(state.History, "Writer created report")
		state.Report = fmt.Sprintf("Report based on %s", state.Data)
		return state, nil
	})

	// Define Handoffs (Edges)
	workflow.SetEntryPoint("Triage")

	// Triage decides where to go
	workflow.AddConditionalEdge("Triage", func(ctx context.Context, state State) string {
		if state.Intent == "research" {
			return "Researcher"
		}
		if state.Intent == "write" {
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
	initialState := State{
		History: []string{},
	}

	result, err := app.Invoke(context.Background(), initialState)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Final State: %+v\n", result)
}
