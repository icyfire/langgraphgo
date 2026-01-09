# LangChainGo Streaming Example

This example demonstrates how to use LangChainGo's streaming functionality with LangGraphGo to build real-time streaming LLM applications.

## Features

- **Basic Streaming**: Stream LLM responses token-by-token using `WithStreamingFunc`
- **Event-Driven Streaming**: Combine streaming with LangGraphGo's event listeners, including `NodeEventProgress` for each chunk
- **Chunk Storage**: Store all streaming chunks in `[][]byte` with thread-safe access for later analysis
- **Multi-Step Streaming**: Stream responses across multiple graph nodes with state passing and checkpointing
- **OpenAI Integration**: Uses LangChainGo's OpenAI client for streaming

## How It Works

### Streaming with LangChainGo

LangChainGo provides streaming support through the `WithStreamingFunc` option:

```go
llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
    // Handle each streaming chunk
    fmt.Print(string(chunk))
    return nil
})
```

### Integration with LangGraphGo

1. **StateGraph**: Holds the streaming callback in the state
2. **ListenableStateGraph**: Emits events during node execution
3. **CheckpointableStateGraph**: Saves state during multi-step streaming workflows

## Examples

### Example 1: Basic Streaming

Demonstrates simple token-by-token streaming from an LLM:

```go
g := graph.NewStateGraph[StreamingState]()

g.AddNode("stream_chat", "stream_chat", func(ctx context.Context, state StreamingState) (StreamingState, error) {
    _, err := llm.GenerateContent(ctx, state.Messages,
        llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
            state.StreamCallback(string(chunk))
            return nil
        }),
    )
    return state, nil
})
```

### Example 2: Streaming with Events

Shows how to combine streaming with event listeners and chunk storage:

```go
// Custom listener with chunk storage
type ProgressListener struct {
    graph.NodeListenerFunc[StreamingState]
    chunkCount int
    chunks     [][]byte
    mu         sync.Mutex
}

// Inside streaming callback - save chunks and emit progress events
llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
    progressListener.mu.Lock()
    chunkCopy := make([]byte, len(chunk))
    copy(chunkCopy, chunk)
    progressListener.chunks = append(progressListener.chunks, chunkCopy)
    progressListener.mu.Unlock()

    progressListener.OnNodeEvent(ctx, graph.NodeEventProgress, nodeName, state, nil)
    state.StreamCallback(string(chunk))
    return nil
})
```

### Example 3: Multi-Step Streaming

Demonstrates streaming across multiple nodes with checkpointing:

```go
g := graph.NewCheckpointableStateGraph[map[string]any]()

g.AddNode("analyze", "analyze", func(ctx context.Context, data map[string]any) (map[string]any, error) {
    var analysisBuilder strings.Builder
    llm.GenerateContent(ctx, messages,
        llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
            fmt.Print(string(chunk))
            analysisBuilder.Write(chunk)
            return nil
        }),
    )
    data["analysis"] = analysisBuilder.String()
    return data, nil
})
```

## Comparison of Streaming Approaches

The three examples demonstrate different streaming patterns with increasing complexity:

### Example 1: Basic Streaming

**Graph Type**: `NewStateGraph[StreamingState]()` - Simplest stateful graph

**Streaming Approach**:
```go
// Streaming callback passed through state
state.StreamCallback = func(chunk string) {
    fmt.Print(chunk)              // Real-time output
    fullResponse.WriteString(chunk) // Accumulate full response
}
```

**Characteristics**:
- **One-way communication**: Output only, doesn't save response to state
- **Simple accumulation**: Uses `strings.Builder` externally
- **Single node**: One node completes the entire flow

**Best for**: Simple one-time queries, no need to preserve conversation history

---

### Example 2: Streaming with Events

**Graph Type**: `NewListenableStateGraph[StreamingState]()` - Listen-enabled stateful graph

**Streaming Approach**:
```go
// Custom listener with chunk storage
type ProgressListener struct {
    graph.NodeListenerFunc[StreamingState]
    chunkCount int
    chunks     [][]byte  // Store all chunks in order
    mu         sync.Mutex // Thread-safe access
}

progressListener := &ProgressListener{}

// Define event handler
progressListener.NodeListenerFunc = graph.NodeListenerFunc[StreamingState](func(...) {
    switch event {
    case graph.NodeEventStart:
        fmt.Printf("[EVENT] Node '%s' started\n", nodeName)
    case graph.NodeEventProgress:
        progressListener.chunkCount++
    case graph.NodeEventComplete:
        // Calculate total bytes from stored chunks
        totalBytes := 0
        for _, chunk := range progressListener.chunks {
            totalBytes += len(chunk)
        }
        fmt.Printf("[EVENT] Completed (chunks: %d, bytes: %d)\n",
            progressListener.chunkCount, totalBytes)

        // Verify chunks are in order by reconstructing
        reconstructed := string(bytes.Join(progressListener.chunks, nil))
        fmt.Printf("[EVENT] Reconstructed: %d chars\n", len(reconstructed))
    }
})

// Inside streaming callback - save chunks and emit events
llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
    // Thread-safe chunk storage
    progressListener.mu.Lock()
    chunkCopy := make([]byte, len(chunk))
    copy(chunkCopy, chunk)
    progressListener.chunks = append(progressListener.chunks, chunkCopy)
    progressListener.mu.Unlock()

    // Emit NodeEventProgress
    progressListener.OnNodeEvent(ctx, graph.NodeEventProgress, nodeName, state, nil)

    // Stream to output
    state.StreamCallback(string(chunk))
    return nil
})
```

**Characteristics**:
- **Chunk storage**: All chunks stored in `[][]byte` in original order
- **Thread-safe**: Uses `sync.Mutex` for concurrent access protection
- **Progress tracking**: Can count and track each chunk received
- **Event monitoring**: Monitors node start/progress/complete/error lifecycle events
- **State persistence**: Response added to `Messages` array for multi-turn conversations
- **Verification**: Can reconstruct full response by joining chunks to verify order

**Best for**: Scenarios requiring detailed progress tracking, chunk-by-chunk monitoring, chunk storage/analysis, and conversation history

---

### Example 3: Multi-Step Streaming

**Graph Type**: `NewCheckpointableStateGraph[map[string]any]()` - Checkpoint-enabled stateful graph

**Streaming Approach**:
```go
// Each node handles streaming independently and accumulates in state
g.AddNode("analyze", "analyze", func(ctx context.Context, data map[string]any) (map[string]any, error) {
    fmt.Println("\n[Step 1] Analysis:")
    fmt.Print("  ")

    var analysisBuilder strings.Builder
    _, err := llm.GenerateContent(ctx, messages,
        llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
            fmt.Print(string(chunk))      // Real-time output
            analysisBuilder.Write(chunk)   // Accumulate in builder
            return nil
        }),
        llms.WithMaxTokens(100),
    )

    // Save to state for next node
    data["analysis"] = analysisBuilder.String()
    data["step1_completed"] = true
    return data, nil
})
```

**Characteristics**:
- **Multi-node workflow**: analyze → expand executed serially
- **State passing**: Each node accumulates streaming output and saves to `map[string]any` for the next node to use
- **Checkpoint support**: Automatically saves state after each node, can resume execution
- **Progressive enhancement**: Each step builds upon the previous step's output

**Best for**: Complex multi-step processes requiring fault tolerance and state recovery

---

### Summary Comparison

| Feature | Basic | Events | Multi-Step |
|---------|-------|--------|------------|
| **Graph Type** | StateGraph | ListenableStateGraph | CheckpointableStateGraph |
| **Streaming Method** | Callback function | Callback + Progress Events | Multiple independent callbacks |
| **State Management** | External accumulation | Save to Messages | Accumulate & pass via map |
| **Event Monitoring** | ❌ | ✅ (Start/Progress/Complete) | ✅ (via checkpoint) |
| **Chunk Storage** | ❌ | ✅ ([][]byte in order) | ❌ |
| **Thread-Safe** | N/A | ✅ (sync.Mutex) | N/A |
| **Checkpoints** | ❌ | ❌ | ✅ |
| **Node Count** | 1 | 1 | 2+ |
| **Complexity** | Low | Medium | High |

**Which to choose**:
- Simple output → Example 1
- Need event notifications/save conversations → Example 2
- Complex workflows/need fault tolerance → Example 3

## Running the Example

### Prerequisites

Set the OpenAI API key environment variable:

```bash
export OPENAI_API_KEY="your-openai-api-key"
```

### Run

```bash
cd examples/langchaingo_streaming
go run main.go
```

## Expected Output

```
🦜🔗 LangChainGo Streaming Examples for LangGraphGo
====================================================

=== Example 1: Basic Streaming ===

Streaming response:
-------------------
Go's concurrency model is based on goroutines...
-------------------
Total characters received: 250

=== Example 2: Streaming with Events ===

[EVENT] Node 'stream_with_events' started
Streaming response with progress events:
-----------------------------------------
[EVENT] Node 'stream_with_events' progress: chunk #1 received
Code flows like water,
bugs hide in the logic stream,
[EVENT] Node 'stream_with_events' progress: chunk #11 received
coffee keeps it real.
-----------------------------------------
[EVENT] Node 'stream_with_events' completed (chunks: 25, bytes: 145)
[EVENT] Reconstructed response length: 145 chars

=== Example 3: Multi-Step Streaming ===

Multi-step streaming response:
-------------------------------
[Step 1] Analysis:
  Go is a statically typed language...
[Step 2] Expansion:
  Go was created at Google...
-------------------------------
Steps completed: step1=true, step2=true
Analysis length: 150 chars
Expansion length: 200 chars

✅ All examples completed!
```

## Use Cases

- **Chat Applications**: Real-time streaming of AI responses
- **Code Generation**: Stream generated code as it's produced
- **Data Analysis**: Stream analysis results progressively
- **Multi-Agent Workflows**: Coordinate streaming across multiple agents

## Notes

- Streaming is handled by the LangChainGo LLM client, not LangGraphGo directly
- LangGraphGo provides the framework to orchestrate streaming workflows
- The `StreamingState` type demonstrates a pattern for passing streaming callbacks through the graph
- For production use, consider error handling, context cancellation, and rate limiting

## See Also

- [LangChainGo Documentation](https://github.com/tmc/langchaingo)
- [LangGraphGo Documentation](https://github.com/smallnest/langgraphgo)
- [Streaming Modes Example](../streaming_modes/)
- [Listeners Example](../listeners/)
