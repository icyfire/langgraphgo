# MCP Adapter for LangGraphGo

This package provides an adapter to use MCP (Model Context Protocol) tools with LangGraphGo and LangChainGo.

## Features

- **Seamless Integration**: Convert MCP tools to `langchaingo/tools.Tool` interface
- **Multiple Server Support**: Connect to multiple MCP servers simultaneously
- **Configuration Loading**: Load MCP server configs from Claude's standard config file
- **Type Safety**: Full Go type safety and error handling

## Installation

This package is part of the LangGraphGo repository and uses `github.com/smallnest/goskills/mcp` for MCP protocol support.

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/smallnest/langgraphgo/adapter/mcp"
    "github.com/smallnest/langgraphgo/prebuilt"
)

func main() {
    ctx := context.Background()

    // Create MCP client from Claude's config file
    mcpClient, err := mcp.NewClientFromConfig(ctx, "~/.claude.json")
    if err != nil {
        panic(err)
    }
    defer mcpClient.Close()

    // Convert MCP tools to langchaingo tools
    tools, err := mcp.MCPToTools(ctx, mcpClient)
    if err != nil {
        panic(err)
    }

    // Use with LangGraphGo prebuilt agents
    agent, err := prebuilt.CreateAgent(
        ctx,
        llm,                    // Your LLM instance
        tools,                  // MCP tools
        nil,                    // No memory
        "You are a helpful assistant",
        nil,                    // Default options
    )
    if err != nil {
        panic(err)
    }

    // Run the agent
    result, err := agent.Invoke(ctx, map[string]any{
        "input": "Search for information about Go",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result["output"])
}
```

### Advanced Usage

#### Using Specific MCP Servers

```go
import (
    "context"

    mcpclient "github.com/smallnest/goskills/mcp"
    "github.com/smallnest/langgraphgo/adapter/mcp"
)

func main() {
    ctx := context.Background()

    // Load config
    config, err := mcpclient.LoadConfig("~/.claude.json")
    if err != nil {
        panic(err)
    }

    // Create client
    client, err := mcpclient.NewClient(ctx, config)
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Convert to tools
    tools, err := mcp.MCPToTools(ctx, client)
    if err != nil {
        panic(err)
    }

    // Use tools...
}
```

#### Direct OpenAI Tool Format

If you need OpenAI tool definitions directly:

```go
openaiTools, err := mcp.MCPToolsToOpenAI(ctx, mcpClient)
if err != nil {
    panic(err)
}

// Use with OpenAI API directly
```

#### Inspecting Tool Schema

```go
for _, tool := range tools {
    if schema, ok := mcp.GetToolSchema(tool); ok {
        fmt.Printf("Tool %s has schema: %+v\n", tool.Name(), schema)
    }
}
```

## Configuration

MCP tools are configured in `~/.claude.json` (or any other path you specify). Example configuration:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"]
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost/mydb"]
    },
    "custom-sse-server": {
      "type": "sse",
      "url": "http://localhost:3000/sse",
      "headers": {
        "Authorization": "Bearer your-token"
      }
    }
  }
}
```

### Configuration Options

Each MCP server can have the following fields:

- `command`: The command to execute (for stdio transport)
- `args`: Arguments for the command (for stdio transport)
- `env`: Environment variables to set (optional)
- `type`: Transport type - "stdio" (default) or "sse"
- `url`: The SSE endpoint URL (for sse transport)
- `headers`: HTTP headers for SSE requests (optional)

## Tool Naming

MCP tools are automatically prefixed with their server name to avoid conflicts. For example, a tool named `read_file` from the `filesystem` server will be available as `filesystem__read_file`.

## Error Handling

The adapter provides detailed error messages for common issues:

- Configuration loading errors
- Connection failures to MCP servers
- Tool invocation errors
- Invalid input/output format errors

All errors are wrapped with context for easy debugging.

## Integration with LangGraphGo

MCP tools work seamlessly with all LangGraphGo features:

- **Prebuilt Agents**: Use with `CreateAgent`, `CreateReactAgent`, etc.
- **Tool Nodes**: Use in custom graph workflows
- **Tool Execution**: Automatic serialization/deserialization
- **State Management**: Full state tracking for tool calls

## Examples

See the `examples/` directory for complete examples:

- `examples/mcp_basic/` - Basic MCP tool usage
- `examples/mcp_agent/` - Using MCP with prebuilt agents
- `examples/mcp_custom/` - Custom MCP server integration

## Dependencies

- `github.com/smallnest/goskills/mcp` - MCP protocol implementation
- `github.com/tmc/langchaingo/tools` - LangChain tool interface
- `github.com/sashabaranov/go-openai` - OpenAI types

## License

Same as LangGraphGo main repository.
