package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// Initialize LLM
	model, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	g := graph.NewStateGraph[map[string]any]()

	g.AddNode("generate", "generate", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		input, ok := state["input"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid input")
		}

		response, err := model.Call(ctx, input)
		if err != nil {
			return nil, err
		}

		return map[string]any{"output": response}, nil
	})

	g.AddEdge("generate", graph.END)
	g.SetEntryPoint("generate")

	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// Invoke with map state
	res, err := runnable.Invoke(ctx, map[string]any{"input": "What is 1 + 1?"})
	if err != nil {
		panic(err)
	}

	fmt.Println("AI Response:", res["output"])
}
