# LLM 流式输出示例

本示例演示如何在 LangGraphGo 中实现 LLM 的流式响应。

## 概述

展示如何：
- 使用 LangGraphGo 设置流式 LLM 响应
- 在图节点中处理实时流式输出
- 使用带有回调函数的自定义状态
- 处理来自 DeepSeek API 的流式数据块

## 展示的功能

### 流式集成
- 创建带有自定义基础 URL 的 OpenAI 兼容客户端
- 配置流式请求
- 处理流式响应块

### 带有回调的自定义状态
- 使用带有自定义状态结构体的 `StateGraphTyped`
- 在状态中嵌入回调函数
- 通过回调进行实时数据流式传输

### 图工作流
- 单节点流式处理
- 类型化状态管理
- 处理流式生命周期

## 运行示例

```bash
cd examples/llm_streaming
export DEEPSEEK_API_KEY="your-api-key-here"
go run main.go
```

## 代码结构

```go
// 带有流式回调的自定义状态
type StreamState struct {
    StreamCallback func(sseResponse openai.ChatCompletionStreamChoice)
}

// 带有流式功能的图节点
g.AddNode("stream", func(ctx context.Context, state StreamState) (StreamState, error) {
    // 配置流式请求
    req := openai.ChatCompletionRequest{
        Model: "deepseek-chat",
        Messages: []openai.ChatCompletionMessage{...},
        Stream: true,  // 启用流式传输
    }

    // 处理流式块
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

## 预期输出

```
Go语言的并发模型是其最核心和最强大的特性之一...

Answer: Go语言的并发模型是其最核心和最强大的特性之一...
```

## 核心概念

- **流式响应**: 以实时块的形式处理 LLM 输出
- **自定义状态**: 使用带有嵌入回调的类型化状态
- **事件驱动架构**: 基于回调的流式模式
- **OpenAI 兼容 API**: 使用 go-openai 客户端访问自定义端点

## 扩展示例

您可以通过以下方式扩展此示例：
- 通过 SSE (Server-Sent Events) 将流式输出发送到前端
- 为复杂工作流添加多个流式节点
- 实现错误处理和重连逻辑
- 与 WebSocket 集成以实现实时 Web 应用
- 添加检查点以支持可恢复的流式会话