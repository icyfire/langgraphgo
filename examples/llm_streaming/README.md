# LLM Streaming Example

This example demonstrates how to implement streaming responses from LLMs in LangGraphGo.

## Overview

Shows how to:
- Set up streaming LLM responses using LangGraphGo
- Handle real-time streaming output in graph nodes
- Use custom state with callback functions
- Process streaming chunks from DeepSeek API

## Features Demonstrated

### Streaming Integration
- Creating OpenAI-compatible client with custom base URL
- Configuring streaming requests
- Processing streaming response chunks

### Custom State with Callbacks
- Using `StateGraphTyped` with custom state structs
- Embedding callback functions in state
- Real-time data streaming via callbacks

### Graph Workflow
- Single-node streaming processing
- Typed state management
- Handling streaming lifecycle

## Running the Example

```bash
cd examples/llm_streaming
export DEEPSEEK_API_KEY="your-api-key-here"
go run main.go
```

## Code Structure

```go
// Custom state with streaming callback
type StreamState struct {
    StreamCallback func(sseResponse openai.ChatCompletionStreamChoice)
}

// Graph node with streaming
g.AddNode("stream", func(ctx context.Context, state StreamState) (StreamState, error) {
    // Configure streaming request
    req := openai.ChatCompletionRequest{
        Model: "deepseek-chat",
        Messages: []openai.ChatCompletionMessage{...},
        Stream: true,  // Enable streaming
    }

    // Process streaming chunks
    stream, err := client.CreateChatCompletionStream(ctx, req)
    for {
        response, err := stream.Recv()
        if len(response.Choices) > 0 {
            state.StreamCallback(response.Choices[0])
        }
    }
    return state, nil
})
```

## Expected Output

```
Go语言的并发模型是其最核心和最强大的特性之一...

Answer: Go语言的并发模型是其最核心和最强大的特性之一...
```

## Key Concepts

- **Streaming Responses**: Processing LLM output in real-time chunks
- **Custom State**: Using typed state with embedded callbacks
- **Event-Driven Architecture**: Callback-based streaming pattern
- **OpenAI-Compatible APIs**: Using go-openai client with custom endpoints

## Extensions

You can extend this example by:
- Sending streaming output to frontend via SSE (Server-Sent Events)
- Adding multiple streaming nodes for complex workflows
- Implementing error handling and reconnection logic
- Integrating with WebSocket for real-time web applications
- Adding checkpointing for resumable streaming sessions