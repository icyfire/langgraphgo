# memU Integration for LangGraphGo

This package provides a Go client for [memU](https://github.com/NevaMind-AI/memU), an advanced agentic memory framework for LLM and AI agent backends.

## What is memU?

memU is a future-oriented agentic memory system that:

- **Hierarchical Memory Structure**: Organizes memory into three layers - Resource → Item → Category
- **Multimodal Support**: Processes conversations, documents, images, audio, and video
- **Dual Retrieval Methods**:
  - **RAG** (embedding-based): Fast, similarity-based retrieval
  - **LLM** (non-embedding): Deep semantic understanding search
- **Self-Evolving**: Memory structure adapts based on usage patterns

## Features

- ✅ Implements the `memory.Memory` interface from LangGraphGo
- ✅ Persistent, structured memory storage
- ✅ AI-powered memory extraction and organization
- ✅ Supports both cloud API and self-hosted deployments
- ✅ Thread-safe concurrent access
- ✅ Context-aware retrieval

## Installation

```bash
go get github.com/smallnest/langgraphgo/memory/memu
```

## Quick Start

### 1. Get an API Key

Visit [memu.so](https://memu.so) to get your API key, or self-host memU using [Docker](https://github.com/NevaMind-AI/memU#option-2-self-hosted).

### 2. Initialize the Client

```go
package main

import (
    "context"
    "log"
    "github.com/smallnest/langgraphgo/memory/memu"
)

func main() {
    client, err := memu.NewClient(memu.Config{
        BaseURL:        "https://api.memu.so", // or your self-hosted URL
        APIKey:         "your-api-key",
        UserID:         "user-123",            // unique identifier for the user
        RetrieveMethod: "rag",                 // "rag" or "llm"
    })
    if err != nil {
        log.Fatal(err)
    }

    // Use the client...
}
```

### 3. Use with LangGraphGo Agents

```go
import (
    "github.com/smallnest/langgraphgo/graph"
    "github.com/smallnest/langgraphgo/memory/memu"
    "github.com/tmc/langchaingo/llms"
)

// Initialize memU client
memClient, _ := memu.NewClient(memu.Config{
    BaseURL: "https://api.memu.so",
    APIKey:  os.Getenv("MEMU_API_KEY"),
    UserID:  userID,
})

// Define agent node with memory
agentNode := func(ctx context.Context, state MyState) (MyState, error) {
    // Retrieve relevant context
    memories, err := memClient.GetContext(ctx, userQuery)
    if err != nil {
        return state, err
    }

    // Use memories to augment prompt
    for _, mem := range memories {
        // Add memory context to system prompt
    }

    // Generate response...

    // Store conversation in memory
    memClient.AddMessage(ctx, userMessage)

    return state, nil
}
```

## Configuration

### Config Options

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `BaseURL` | string | Yes | memU API base URL (e.g., `https://api.memu.so`) |
| `APIKey` | string | Yes | Your memU API key |
| `UserID` | string | Yes | Unique identifier for memory isolation |
| `RetrieveMethod` | string | No | `"rag"` (default) or `"llm"` |
| `HTTPClient` | *http.Client | No | Custom HTTP client (defaults to 30s timeout) |

### Retrieval Methods

- **RAG** (default): Fast embedding-based search. Best for:
  - Quick retrieval
  - High-volume queries
  - Factual information

- **LLM**: Deep semantic understanding search. Best for:
  - Complex queries
  - Reasoning-based retrieval
  - Understanding context and relationships

## API Reference

### Client Methods

#### `NewClient(cfg Config) (*Client, error)`

Creates a new memU client with the given configuration.

#### `AddMessage(ctx context.Context, msg *Message) error`

Stores a message in memU's memory system. The message will be:
- Extracted into structured memory items
- Organized into appropriate categories
- Made available for future retrieval

#### `GetContext(ctx context.Context, query string) ([]*Message, error)`

Retrieves relevant memories based on the query. Returns:
- Category summaries (high-level context)
- Individual memory items (specific details)
- Associated resource references

#### `GetStats(ctx context.Context) (*Stats, error)`

Returns statistics about memory usage:
- Total messages stored
- Total tokens
- Active categories
- Memory compression rate

## Memory Structure

memU organizes memory in a three-layer hierarchy:

```
┌─────────────────────────────────────────┐
│           Category Layer                 │
│  (e.g., "preferences.md", "work.md")    │
├─────────────────────────────────────────┤
│            Item Layer                    │
│  (e.g., "likes coffee", "early bird")   │
├─────────────────────────────────────────┤
│          Resource Layer                  │
│  (raw conversations, documents, images) │
└─────────────────────────────────────────┘
```

## Examples

See the [examples/memu_agent](../../examples/memu_agent) directory for a complete example demonstrating:
- Multi-turn conversations with persistent memory
- Context-aware response generation
- Memory statistics and retrieval

## Comparison with Other Memory Strategies

| Strategy | Persistence | Structure | AI-Powered | Use Case |
|----------|-------------|-----------|------------|----------|
| **Buffer** | No | Flat | No | Simple history |
| **SlidingWindow** | No | Flat | No | Recent context |
| **Summarization** | No | Flat | Yes | Compressed history |
| **memU** | **Yes** | **Hierarchical** | **Yes** | **Production agents** |

## Environment Variables

```bash
# Required
export MEMU_API_KEY="your-api-key"

# Optional
export MEMU_BASE_URL="https://api.memu.so"  # Default
export MEMU_RETRIEVE_METHOD="rag"           # "rag" or "llm"
```

## Self-Hosting

To self-host memU instead of using the cloud API:

```bash
# Using Docker
docker run -d \
  --name memu \
  -p 8000:8000 \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  ghcr.io/nevamind/memu:latest

# Then configure your client
client, _ := memu.NewClient(memu.Config{
    BaseURL: "http://localhost:8000",
    // ...
})
```

See [memU documentation](https://github.com/NevaMind-AI/memU) for detailed self-hosting instructions.

## Limitations

- `Clear()` operation is not currently supported by memU's API
- Requires network connectivity to memU service
- API rate limits apply (cloud version)

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This package is part of LangGraphGo and follows the same license.

## Links

- [memU GitHub](https://github.com/NevaMind-AI/memU)
- [memU Documentation](https://memu.pro/docs)
- [LangGraphGo](https://github.com/smallnest/langgraphgo)
