# LangChainGo 流式输出示例

本示例演示如何将 LangChainGo 的流式输出功能与 LangGraphGo 结合使用，构建实时流式 LLM 应用。

## 特性

- **基础流式输出**：使用 `WithStreamingFunc` 逐个 token 流式输出 LLM 响应
- **事件驱动流式输出**：将流式输出与 LangGraphGo 的事件监听器结合，包括每个 chunk 的 `NodeEventProgress`
- **Chunk 存储**：在 `[][]byte` 中存储所有流式输出的 chunks，线程安全访问，用于后续分析
- **多步流式输出**：在多个图节点间进行流式输出，支持状态传递和检查点
- **OpenAI 集成**：使用 LangChainGo 的 OpenAI 客户端进行流式输出

## 工作原理

### LangChainGo 流式输出

LangChainGo 通过 `WithStreamingFunc` 选项提供流式输出支持：

```go
llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
    // 处理每个流式数据块
    fmt.Print(string(chunk))
    return nil
})
```

### 与 LangGraphGo 集成

1. **StateGraph**：在状态中保存流式回调函数
2. **ListenableStateGraph**：在节点执行期间发出事件
3. **CheckpointableStateGraph**：在多步流式工作流中保存状态

## 示例

### 示例 1：基础流式输出

演示简单的 LLM 逐 token 流式输出：

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

### 示例 2：带事件的流式输出

展示如何将流式输出与事件监听器结合，以及 chunk 存储：

```go
// 自定义监听器，带 chunk 存储
type ProgressListener struct {
    graph.NodeListenerFunc[StreamingState]
    chunkCount int
    chunks     [][]byte
    mu         sync.Mutex
}

// 在流式回调内部 - 保存 chunks 并发出进度事件
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

### 示例 3：多步流式输出

演示在多个节点间进行带检查点的流式输出：

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

## 流式输出方式对比

三个示例展示了不同复杂度的流式输出模式：

### 示例 1：基础流式输出

**图类型**: `NewStateGraph[StreamingState]()` - 最简单的有状态图

**流式输出方式**:
```go
// 通过 state 传递流式回调
state.StreamCallback = func(chunk string) {
    fmt.Print(chunk)              // 实时输出
    fullResponse.WriteString(chunk) // 累积完整响应
}
```

**特点**:
- **单向通信**：只负责输出，不保存响应到状态
- **简单累积**：使用 `strings.Builder` 在外部累积完整响应
- **单节点**：一个节点完成整个流程

**适用场景**: 简单的一次性查询，不需要保存对话历史

---

### 示例 2：带事件的流式输出

**图类型**: `NewListenableStateGraph[StreamingState]()` - 可监听的有状态图

**流式输出方式**:
```go
// 自定义监听器，带 chunk 存储
type ProgressListener struct {
    graph.NodeListenerFunc[StreamingState]
    chunkCount int
    chunks     [][]byte  // 按顺序存储所有 chunks
    mu         sync.Mutex // 线程安全访问
}

progressListener := &ProgressListener{}

// 定义事件处理器
progressListener.NodeListenerFunc = graph.NodeListenerFunc[StreamingState](func(...) {
    switch event {
    case graph.NodeEventStart:
        fmt.Printf("[EVENT] 节点 '%s' 开始\n", nodeName)
    case graph.NodeEventProgress:
        progressListener.chunkCount++
    case graph.NodeEventComplete:
        // 从存储的 chunks 计算总字节数
        totalBytes := 0
        for _, chunk := range progressListener.chunks {
            totalBytes += len(chunk)
        }
        fmt.Printf("[EVENT] 完成 (chunks: %d, bytes: %d)\n",
            progressListener.chunkCount, totalBytes)

        // 通过拼接验证 chunks 顺序
        reconstructed := string(bytes.Join(progressListener.chunks, nil))
        fmt.Printf("[EVENT] 重构后长度: %d 字符\n", len(reconstructed))
    }
})

// 在流式回调内部 - 保存 chunks 并发出事件
llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
    // 线程安全的 chunk 存储
    progressListener.mu.Lock()
    chunkCopy := make([]byte, len(chunk))
    copy(chunkCopy, chunk)
    progressListener.chunks = append(progressListener.chunks, chunkCopy)
    progressListener.mu.Unlock()

    // 发出 NodeEventProgress
    progressListener.OnNodeEvent(ctx, graph.NodeEventProgress, nodeName, state, nil)

    // 输出到控制台
    state.StreamCallback(string(chunk))
    return nil
})
```

**特点**:
- **Chunk 存储**：所有 chunks 按原始顺序存储在 `[][]byte` 中
- **线程安全**：使用 `sync.Mutex` 保护并发访问
- **进度跟踪**：可以计数和跟踪接收到的每个 chunk
- **事件监听**：监听节点开始/进度/完成/错误生命周期事件
- **状态持久化**：响应被添加到 `Messages` 数组，可用于多轮对话
- **顺序验证**：可以通过拼接 chunks 来重构完整响应以验证顺序

**适用场景**: 需要详细进度跟踪、逐 chunk 监控、chunk 存储/分析和保存对话历史的场景

---

### 示例 3：多步流式输出

**图类型**: `NewCheckpointableStateGraph[map[string]any]()` - 可检查点的有状态图

**流式输出方式**:
```go
// 每个节点独立处理流式输出并在状态中累积
g.AddNode("analyze", "analyze", func(ctx context.Context, data map[string]any) (map[string]any, error) {
    fmt.Println("\n[步骤 1] 分析:")
    fmt.Print("  ")

    var analysisBuilder strings.Builder
    _, err := llm.GenerateContent(ctx, messages,
        llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
            fmt.Print(string(chunk))      // 实时输出
            analysisBuilder.Write(chunk)   // 在 builder 中累积
            return nil
        }),
        llms.WithMaxTokens(100),
    )

    // 保存到状态供下一个节点使用
    data["analysis"] = analysisBuilder.String()
    data["step1_completed"] = true
    return data, nil
})
```

**特点**:
- **多节点工作流**：analyze → expand 串行执行
- **状态传递**：每个节点累积流式输出并保存到 `map[string]any` 供下一个节点使用
- **检查点支持**：每个节点完成后自动保存状态，可恢复执行
- **渐进式增强**：每一步都基于前一步的输出进行构建

**适用场景**: 复杂的多步骤处理流程，需要容错和状态恢复

---

### 总结对比

| 特性 | 基础 | 事件 | 多步 |
|------|-------|--------|------------|
| **图类型** | StateGraph | ListenableStateGraph | CheckpointableStateGraph |
| **流式输出方式** | 回调函数 | 回调 + 进度事件 | 多个独立回调 |
| **状态管理** | 外部累积 | 保存到 Messages | 累积并通过 map 传递 |
| **事件监听** | ❌ | ✅ (开始/进度/完成) | ✅ (通过 checkpoint) |
| **Chunk 存储** | ❌ | ✅ ([][]byte 按顺序) | ❌ |
| **线程安全** | N/A | ✅ (sync.Mutex) | N/A |
| **检查点** | ❌ | ❌ | ✅ |
| **节点数** | 1 | 1 | 2+ |
| **复杂度** | 低 | 中 | 高 |

**选择建议**:
- 简单输出 → 示例 1
- 需要事件通知/保存对话 → 示例 2
- 复杂工作流/需要容错 → 示例 3

## 运行示例

### 前置要求

设置 OpenAI API key 环境变量：

```bash
export OPENAI_API_KEY="your-openai-api-key"
```

### 运行

```bash
cd examples/langchaingo_streaming
go run main.go
```

## 预期输出

```
🦜🔗 LangChainGo 流式输出示例 for LangGraphGo
====================================================

=== 示例 1：基础流式输出 ===

流式响应：
-------------------
Go 的并发模型基于 goroutines...
-------------------
接收到的总字符数：250

=== 示例 2：带事件的流式输出 ===

[EVENT] 节点 'stream_with_events' 开始
带进度事件的流式响应：
-----------------------------------------
[EVENT] 节点 'stream_with_events' 进度: 收到第 1 个 chunk
代码如流水般流淌，
bug 藏于逻辑之中，
[EVENT] 节点 'stream_with_events' 进度: 收到第 11 个 chunk
咖啡让一切继续。
-----------------------------------------
[EVENT] 节点 'stream_with_events' 完成 (chunks: 25, bytes: 145)
[EVENT] 重构后响应长度: 145 字符

=== 示例 3：多步流式输出 ===

多步流式响应：
-------------------------------
[步骤 1] 分析：
  Go 是一种静态类型语言...
[步骤 2] 扩展：
  Go 由 Google 创建...
-------------------------------
步骤完成：step1=true, step2=true
分析长度：150 字符
扩展长度：200 字符

✅ 所有示例已完成！
```

## 使用场景

- **聊天应用**：实时流式输出 AI 响应
- **代码生成**：流式输出生成的代码
- **数据分析**：渐进式流式输出分析结果
- **多 Agent 工作流**：在多个 agent 间协调流式输出

## 注意事项

- 流式输出由 LangChainGo LLM 客户端处理，而非 LangGraphGo 直接处理
- LangGraphGo 提供编排流式工作流的框架
- `StreamingState` 类型演示了在图中传递流式回调的模式
- 生产环境使用时，请考虑错误处理、上下文取消和速率限制

## 另请参阅

- [LangChainGo 文档](https://github.com/tmc/langchaingo)
- [LangGraphGo 文档](https://github.com/smallnest/langgraphgo)
- [流式模式示例](../streaming_modes/)
- [监听器示例](../listeners/)
