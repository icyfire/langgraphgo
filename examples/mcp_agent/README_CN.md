# MCP Agent 示例

本示例演示了如何在 LangGraphGo 代理中使用 MCP (Model Context Protocol) 工具。

## 前置条件

1. **配置 MCP 服务器**: 创建 `~/.claude.json` 文件并配置您的 MCP 服务器：

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

2. **安装 MCP 服务器**: 本示例使用 npm 包，请确保已安装 Node.js：

```bash
npm install -g @modelcontextprotocol/server-filesystem
npm install -g @modelcontextprotocol/server-brave-search
```

3. **设置 OpenAI API Key**:

```bash
export OPENAI_API_KEY="your-openai-api-key"
```

## 运行示例

```bash
cd examples/mcp_agent
go run main.go
```

## 功能说明

1. **加载 MCP 配置**: 读取 `~/.claude.json` 以发现可用的 MCP 服务器
2. **连接服务器**: 与所有配置的 MCP 服务器建立连接
3. **获取工具**: 从每个服务器获取可用工具
4. **创建代理**: 设置带有 MCP 工具的 LangGraphGo 代理
5. **处理查询**: 运行可能使用 MCP 工具的示例查询
6. **显示结果**: 显示代理的响应和工具使用情况

## 可用的 MCP 服务器

以下是一些常用的 MCP 服务器：

### 文件系统 (Filesystem)
```json
{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/directory"]
}
```

### Brave 搜索 (Brave Search)
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

## 自定义

您可以修改示例以：

- 使用不同的查询
- 添加自定义 MCP 服务器
- 实现流式响应
- 添加记忆/状态管理
- 使用不同的 LLM 模型

## 故障排除

### 连接失败

如果看到 "Failed to connect to MCP server" 错误：
- 检查 MCP 服务器命令是否已安装且可访问
- 验证环境变量是否正确设置
- 确保配置中的路径是绝对路径且存在

### 未找到工具

如果工具未显示：
- 直接运行 `npx <server-package>` 测试服务器
- 检查 stderr 中的服务器日志
- 验证服务器是否支持 MCP 协议版本

### API Key 问题

确保设置了所有必需的 API Key：
- LLM 的 `OPENAI_API_KEY`
- 特定服务器的 Key (BRAVE_API_KEY, GITHUB_TOKEN 等)
