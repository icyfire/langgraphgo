# MCP Agent Example

This example demonstrates how to use MCP (Model Context Protocol) tools with LangGraphGo agents.

## Prerequisites

1. **Configure MCP Servers**: Create a `~/.claude.json` file with your MCP server configurations:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/username/Documents"]
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "your-brave-api-key"
      }
    }
  }
}
```

2. **Install MCP Servers**: The example uses npm packages, so make sure you have Node.js installed:

```bash
npm install -g @modelcontextprotocol/server-filesystem
npm install -g @modelcontextprotocol/server-brave-search
```

3. **Set OpenAI API Key**:

```bash
export OPENAI_API_KEY="your-openai-api-key"
```

## Running the Example

```bash
cd examples/mcp_agent
go run main.go
```

## What It Does

1. **Loads MCP Configuration**: Reads `~/.claude.json` to discover available MCP servers
2. **Connects to Servers**: Establishes connections to all configured MCP servers
3. **Retrieves Tools**: Fetches available tools from each server
4. **Creates Agent**: Sets up a LangGraphGo agent with MCP tools
5. **Processes Query**: Runs a sample query that may use MCP tools
6. **Displays Results**: Shows the agent's response and tool usage

## Available MCP Servers

Here are some popular MCP servers you can use:

### Filesystem
```json
{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/directory"]
}
```

### Brave Search
```json
{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-brave-search"],
  "env": {
    "BRAVE_API_KEY": "your-api-key"
  }
}
```

### PostgreSQL
```json
{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost/mydb"]
}
```

### GitHub
```json
{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-github"],
  "env": {
    "GITHUB_PERSONAL_ACCESS_TOKEN": "your-token"
  }
}
```

### Slack
```json
{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-slack"],
  "env": {
    "SLACK_BOT_TOKEN": "your-token"
  }
}
```

## Customization

You can modify the example to:

- Use different queries
- Add custom MCP servers
- Implement streaming responses
- Add memory/state management
- Use different LLM models

## Troubleshooting

### Connection Failures

If you see "Failed to connect to MCP server" errors:
- Check that the MCP server command is installed and accessible
- Verify environment variables are set correctly
- Ensure paths in the config are absolute and exist

### Tool Not Found

If tools aren't appearing:
- Run `npx <server-package>` directly to test the server
- Check server logs in stderr
- Verify the server supports the MCP protocol version

### API Key Issues

Make sure all required API keys are set:
- `OPENAI_API_KEY` for the LLM
- Server-specific keys (BRAVE_API_KEY, GITHUB_TOKEN, etc.)
