package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
)

// AgentState for the example
type AgentState struct {
	Query    string
	Intent   string
	Info     string
	Response string
}

func classifyIntent(ctx context.Context, state map[string]any) (map[string]any, error) {
	agentState := state["agent_state"].(*AgentState)
	fmt.Println("[Memory Integration] Classifying intent...")
	// Simulate classification
	agentState.Intent = "search"
	return state, nil
}

func retrieveInformation(ctx context.Context, state map[string]any) (map[string]any, error) {
	agentState := state["agent_state"].(*AgentState)
	fmt.Println("[Memory Integration] Retrieving information...")
	// Simulate retrieval
	agentState.Info = "Information about " + agentState.Query
	return state, nil
}

func generateResponse(ctx context.Context, state map[string]any) (map[string]any, error) {
	agentState := state["agent_state"].(*AgentState)
	fmt.Println("[Memory Integration] Generating response...")
	agentState.Response = "Here is what I found: " + agentState.Info
	return state, nil
}

func main() {
	// Create a state graph with map state
	g := graph.NewStateGraph[map[string]any]()

	g.AddNode("classify", "classify", classifyIntent)
	g.AddNode("retrieve", "retrieve", retrieveInformation)
	g.AddNode("generate", "generate", generateResponse)

	g.SetEntryPoint("classify")
	g.AddEdge("classify", "retrieve")
	g.AddEdge("retrieve", "generate")
	g.AddEdge("generate", graph.END)

	runnable, err := g.Compile()
	if err != nil {
		log.Fatal(err)
	}

	agentState := &AgentState{Query: "Go Generics"}
	input := map[string]any{
		"agent_state": agentState,
	}

	result, err := runnable.Invoke(context.Background(), input)
	if err != nil {
		log.Fatal(err)
	}

	finalState := result["agent_state"].(*AgentState)
	fmt.Printf("Response: %s\n", finalState.Response)
}
