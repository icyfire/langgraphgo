# Adapter Streaming Example

This example demonstrates how to use the `StreamingLLM` adapter from the `adapter` package to add streaming capabilities to any LLM.

## Overview

This example shows how to:
- Wrap an LLM with streaming support using `adapter.WrapLLMWithStreaming`
- Use streaming LLM in a LangGraphGo state graph
- Handle streaming responses in real-time
- Capture complete responses while displaying chunks as they arrive

## Prerequisites

1. **OpenAI API Key**: Set your OpenAI API key as an environment variable:
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

## Running the Example

```bash
cd examples/adapter_streaming
go run main.go
```

With custom input:
```bash
go run main.go "What is the capital of France?"
```

## Code Structure

```go
// Wrap LLM with streaming capability
streamingLLM := adapter.WrapLLMWithStreaming(llm, func(chunk string) {
    fmt.Print(chunk)              // Print each chunk in real-time
    fullResponse.WriteString(chunk) // Accumulate full response
})

// Use in graph node
g.AddNode("chat", func(ctx context.Context, state State) (State, error) {
    response, err := streamingLLM.GenerateContent(ctx, state.Messages)
    // ...
})
```

## Key Concepts

- **StreamingLLM**: A wrapper that adds streaming callback to any `llms.Model`
- **Real-time Output**: Process and display chunks as they arrive from the API
- **Response Accumulation**: Store complete response for further processing
- **Graph Integration**: Use streaming LLM seamlessly in state graph nodes

## Features

- **Real-time Streaming**: Display LLM output token-by-token
- **Flexible Callback**: Custom callback function for handling chunks
- **State Management**: Full integration with LangGraphGo state management
- **Error Handling**: Proper error propagation through the graph
