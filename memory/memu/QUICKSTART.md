# memU Integration - Quick Start Guide

This guide will help you quickly integrate memU into your LangGraphGo agent for advanced memory management.

## What is memU?

[memU](https://github.com/NevaMind-AI/memU) is an agentic memory framework that:
- **Extracts structured memory** from conversations, documents, and images
- **Organizes memory hierarchically** (Resource → Item → Category)
- **Provides dual retrieval**: RAG (fast) or LLM (deep understanding)
- **Supports multimodal inputs**: text, images, audio, video

## Prerequisites

1. Get a memU API key at [memu.so](https://memu.so)
2. Or self-host memU following [their guide](https://github.com/NevaMind-AI/memU#option-2-self-hosted)

## Installation

```bash
go get github.com/smallnest/langgraphgo/memory/memu
```

## Basic Usage

### 1. Initialize the Client

```go
import (
    "context"
    "os"
    "github.com/smallnest/langgraphgo/memory/memu"
)

func main() {
    client, err := memu.NewClient(memu.Config{
        BaseURL:        "https://api.memu.so",
        APIKey:         os.Getenv("MEMU_API_KEY"),
        UserID:         "user-123",  // Unique per user
        RetrieveMethod: "rag",       // "rag" or "llm"
    })
    if err != nil {
        panic(err)
    }
}
```

### 2. Store Messages in Memory

```go
ctx := context.Background()

// User shares a preference
msg := memory.NewMessage("user", "I prefer working in the morning and love coffee")
if err := client.AddMessage(ctx, msg); err != nil {
    log.Printf("Failed to store: %v", err)
}
```

### 3. Retrieve Relevant Context

```go
// Query for relevant memories
memories, err := client.GetContext(ctx, "What are my work habits?")
if err != nil {
    log.Printf("Failed to retrieve: %v", err)
}

// Use retrieved memories to personalize responses
for _, mem := range memories {
    fmt.Printf("[%s] %s\n", mem.Metadata["source"], mem.Content)
}

// Output:
// [memu_category] [work_life] The user prefers working in morning hours...
// [memu_item] Prefers working in morning
// [memu_item] Loves coffee
```

## Integration with LangGraphGo Agents

### Complete Agent Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/smallnest/langgraphgo/graph"
    "github.com/smallnest/langgraphgo/memory"
    "github.com/smallnest/langgraphgo/memory/memu"
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
)

type AgentState struct {
    Messages []llms.MessageContent `json:"messages"`
}

func main() {
    ctx := context.Background()

    // Initialize LLM
    llm, _ := openai.New(openai.WithToken(os.Getenv("OPENAI_API_KEY")))

    // Initialize memU
    memClient, _ := memu.NewClient(memu.Config{
        BaseURL: "https://api.memu.so",
        APIKey:  os.Getenv("MEMU_API_KEY"),
        UserID:  "demo-user",
    })

    // Build agent graph
    builder := graph.NewStateGraph(graph.StateGraphOpts{})

    // Define agent node
    agentNode := func(ctx context.Context, state AgentState) (AgentState, error) {
        // Get last user message
        lastMsg := getLastUserMessage(state.Messages)

        // Retrieve relevant memories
        memories, _ := memClient.GetContext(ctx, lastMsg)

        // Build prompt with memory context
        systemPrompt := "You are a helpful assistant.\n\n"
        if len(memories) > 0 {
            systemPrompt += "User memories:\n"
            for _, mem := range memories {
                systemPrompt += fmt.Sprintf("- %s\n", mem.Content)
            }
        }

        // Generate response
        response, _ := llm.Generate(ctx, append(
            []llms.MessageContent{llms.TextParts(llms.SystemMessageType, systemPrompt)},
            state.Messages...,
        ))

        // Store conversation
        memClient.AddMessage(ctx, memory.NewMessage("user", lastMsg))
        memClient.AddMessage(ctx, memory.NewMessage("assistant", response))

        // Append response
        state.Messages = append(state.Messages,
            llms.TextParts(llms.AIMessageType, response))

        return state, nil
    }

    builder.AddNode("agent", agentNode)
    builder.AddEdge(graph.START, "agent")
    builder.AddEdge("agent", graph.END)

    runnable := builder.Compile()

    // Run the agent
    state := AgentState{
        Messages: []llms.MessageContent{
            llms.TextParts(llms.HumanMessageType,
                "My name is Alice and I love coffee"),
        },
    }

    result, _ := runnable.Invoke(ctx, state)
    fmt.Println(result)
}
```

## Retrieval Methods

### RAG (Default)
```go
client, _ := memu.NewClient(memu.Config{
    RetrieveMethod: "rag",  // Fast embedding-based search
    // ...
})
```
- ✅ Fast response times
- ✅ Best for factual queries
- ✅ Lower API costs

### LLM
```go
client, _ := memu.NewClient(memu.Config{
    RetrieveMethod: "llm",  // Deep semantic understanding
    // ...
})
```
- ✅ Deep semantic understanding
- ✅ Better for complex queries
- ✅ Understands context and relationships

## Memory Structure

memU organizes memory in three layers:

```
Resource (Raw Data)
    ↓
Item (Extracted Facts)
    ↓
Category (Organized Topics)
```

**Example:**
- **Resource**: Conversation JSON file
- **Items**: "Prefers coffee", "Works in morning", "Named Alice"
- **Categories**: "preferences.md", "work_life.md", "identity.md"

## Environment Variables

```bash
# Required
export MEMU_API_KEY="your-api-key-here"

# Optional
export MEMU_BASE_URL="https://api.memu.so"  # Cloud API (default)
export MEMU_RETRIEVE_METHOD="rag"           # "rag" or "llm"
```

## Complete Example

See [examples/memu_agent](../../examples/memu_agent) for a working example demonstrating:
- Multi-turn conversations
- Memory persistence across sessions
- Context-aware responses

## Common Patterns

### 1. User-Specific Memory

```go
func getUserMemoryClient(userID string) (*memu.Client, error) {
    return memu.NewClient(memu.Config{
        BaseURL: "https://api.memu.so",
        APIKey:  os.Getenv("MEMU_API_KEY"),
        UserID:  userID,  // Unique per user
    })
}
```

### 2. Error Handling

```go
memories, err := client.GetContext(ctx, query)
if err != nil {
    // Fall back to basic conversation
    log.Printf("memU unavailable: %v", err)
    memories = []*memory.Message{}
}
```

### 3. Memory Statistics

```go
stats, err := client.GetStats(ctx)
if err == nil {
    log.Printf("Categories: %d, Items: %d",
        stats.ActiveMessages, stats.TotalMessages)
}
```

## Troubleshooting

### "API key required"
Ensure `MEMU_API_KEY` environment variable is set or pass APIKey in config.

### "clear operation not supported"
memU doesn't support clearing memory via API. This is by design for data integrity.

### Slow responses
- Use `"rag"` retrieval method for faster responses
- Consider caching frequently accessed memories
- Use self-hosted memU to reduce network latency

## Next Steps

1. Read the [full API documentation](README.md)
2. Explore the [example agent](../../examples/memu_agent)
3. Visit [memU documentation](https://memu.pro/docs)
4. Join [memU Discord](https://discord.gg/memu) for community support

## License

This integration is part of LangGraphGo and follows the same license.
