package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/smallnest/langgraphgo/graph"
)

type StreamState struct {
	StreamCallback func(sseResponse openai.ChatCompletionStreamChoice)
}

func main() {

	g := graph.NewStateGraph[StreamState]()
	g.AddNode("stream", "stream", func(ctx context.Context, state StreamState) (StreamState, error) {
		apiKey := os.Getenv("DEEPSEEK_API_KEY")
		if apiKey == "" {
			log.Fatal("Please set the DEEPSEEK_API_KEY environment variable.")
		}
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = "https://api.deepseek.com/v1"
		client := openai.NewClientWithConfig(config)

		req := openai.ChatCompletionRequest{
			Model: "deepseek-chat",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "你是一个有用的助手，请用中文回答。",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "请详细解释一下 Go 语言的并发模型。",
				},
			},
			Stream: true,
		}

		stream, err := client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			return state, err
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				break
			}

			if len(response.Choices) > 0 {
				state.StreamCallback(response.Choices[0])
			}
		}
		return state, nil
	})
	g.SetEntryPoint("stream")
	compile, _ := g.Compile()
	answer := ""
	state := StreamState{
		StreamCallback: func(sseResp openai.ChatCompletionStreamChoice) {
			// Control the flow direction by yourself: for example, sending to the frontend via SSE
			fmt.Print(sseResp.Delta.Content)
			answer += sseResp.Delta.Content
		},
	}
	_, _ = compile.Invoke(context.Background(), state)

	fmt.Println("\nAnswer: ", answer)
}
