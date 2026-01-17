# memU-Powered Agent Example

This example demonstrates how to integrate [memU](https://memu.so) with LangGraphGo for advanced, persistent memory management in AI agents.

## Overview

A simple chat agent that uses memU to remember user information and preferences across multi-turn conversations. Unlike basic in-memory storage, memU provides a cloud-based persistent memory that can extract and organize information automatically.

## Key Features

- **Persistent Memory**: Memories persist across different sessions and even different runs of the application.
- **AI-Powered Extraction**: memU automatically identifies and extracts important information (like names and preferences) from the conversation.
- **Context-Aware Retrieval**: The agent retrieves relevant memories based on the current user input to provide personalized responses.
- **Hierarchical Organization**: Information is organized into Categories and Items for better management.

## Prerequisites

1. Get an API key from [memu.so](https://memu.so).
2. Set the `MEMU_API_KEY` environment variable.

```bash
export MEMU_API_KEY='your-api-key'
```

## Running the Example

```bash
cd examples/memu_agent
go run main.go
```

The example will run a simulated conversation:
1. Alice introduces herself and her preferences.
2. The agent stores this in memU.
3. The agent recalls Alice's name and favorite drink in subsequent turns.
4. Finally, it displays memory statistics.

## How It Works

The example follows this pattern:

1. **Initialize memU Client**:
   ```go
   memClient, err := memu.NewClient(memu.Config{
       APIKey: os.Getenv("MEMU_API_KEY"),
       UserID: "demo-user",
       RetrieveMethod: "rag",
   })
   ```

2. **Add Messages**:
   Every user and assistant message is added to memU.
   ```go
   msg := memory.NewMessage("user", userMsg)
   memClient.AddMessage(ctx, msg)
   ```

3. **Retrieve Context**:
   Before generating a response, the agent asks memU for relevant context.
   ```go
   memories, err := memClient.GetContext(ctx, userMsg)
   ```

4. **Generate Personalized Response**:
   The agent uses the retrieved memories to craft a response that knows about the user.

## Related Examples

- [memory_agent](../memory_agent/) - Generic memory agent with various strategies
- [memory_chatbot](../memory_chatbot/) - Basic chatbot with memory
- [memory_basic](../memory_basic/) - Simple memory usage
