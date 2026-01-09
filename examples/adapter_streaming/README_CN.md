# Adapter 流式输出示例

本示例演示如何使用 `adapter` 包中的 `StreamingLLM` 适配器为任何 LLM 添加流式输出功能。

## 概述

本示例展示了如何：
- 使用 `adapter.WrapLLMWithStreaming` 包装 LLM 以支持流式输出
- 在 LangGraphGo 状态图中使用流式 LLM
- 实时处理流式响应
- 在显示数据块的同时捕获完整响应

## 前置条件

1. **OpenAI API Key**: 设置 OpenAI API 密钥作为环境变量：
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

## 运行示例

```bash
cd examples/adapter_streaming
go run main.go
```

使用自定义输入：
```bash
go run main.go "法国的首都是什么？"
```

## 代码结构

```go
// 使用流式输出功能包装 LLM
streamingLLM := adapter.WrapLLMWithStreaming(llm, func(chunk string) {
    fmt.Print(chunk)              // 实时打印每个数据块
    fullResponse.WriteString(chunk) // 累积完整响应
})

// 在图节点中使用
g.AddNode("chat", func(ctx context.Context, state State) (State, error) {
    response, err := streamingLLM.GenerateContent(ctx, state.Messages)
    // ...
})
```

## 核心概念

- **StreamingLLM**: 为任何 `llms.Model` 添加流式回调的包装器
- **实时输出**: 处理并显示从 API 返回的每个数据块
- **响应累积**: 存储完整响应用于后续处理
- **图集成**: 在状态图节点中无缝使用流式 LLM

## 特性

- **实时流式输出**: 逐个 token 显示 LLM 输出
- **灵活回调**: 自定义回调函数处理数据块
- **状态管理**: 与 LangGraphGo 状态管理完全集成
- **错误处理**: 通过图正确传播错误
