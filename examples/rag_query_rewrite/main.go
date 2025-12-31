package main

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

func main() {
	g := graph.NewStateGraph[map[string]any]()

	g.AddNode("rewrite_query", "Rewrite user query for better retrieval", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		query, _ := state["query"].(string)
		fmt.Printf("Original query: %s\n", query)
		rewritten := "LangGraph architecture state management" // Simulated rewrite
		fmt.Printf("Rewritten query: %s\n", rewritten)
		return map[string]any{"rewritten_query": rewritten}, nil
	})

	g.AddNode("retrieve", "Retrieve documents", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		query, _ := state["rewritten_query"].(string)
		fmt.Printf("Retrieving documents for: %s\n", query)
		docs := []string{"Doc A: LangGraph manages state...", "Doc B: Graph nodes execution..."}
		return map[string]any{"documents": docs}, nil
	})

	g.AddNode("generate", "Generate Answer", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		docs, _ := state["documents"].([]string)
		fmt.Printf("Generating answer based on %d documents\n", len(docs))
		answer := "LangGraph uses a graph-based approach for state management."
		return map[string]any{"answer": answer}, nil
	})

	g.SetEntryPoint("rewrite_query")
	g.AddEdge("rewrite_query", "retrieve")
	g.AddEdge("retrieve", "generate")
	g.AddEdge("generate", graph.END)

	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	res, err := runnable.Invoke(context.Background(), map[string]any{"query": "How does LangGraph handle state?"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Answer: %s\n", res["answer"])
}
