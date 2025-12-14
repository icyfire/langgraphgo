# 带事件监听的泛型 StateGraph

本示例展示了如何在 LangGraphGo 中使用**类型安全的泛型 StateGraph**配合**事件监听器**。可监听图（Listenable Graph）为你的工作流提供实时监控和事件处理能力。

## 概述

可监听 StateGraph 扩展了泛型 StateGraph，增加了以下功能：
- **事件驱动架构** - 监听图执行过程中的各种事件
- **实时监控** - 跟踪节点执行、错误和状态变化
- **流式支持** - 通过流式传输实时事件
- **灵活的监听器类型** - 支持全局监听器和节点特定监听器

## 核心特性

✅ **类型安全** - 泛型提供完整的编译时类型安全
✅ **事件监听** - 实时监控图执行过程
✅ **多重监听器** - 可添加全局或节点特定的监听器
✅ **事件流式传输** - 通过通道流式传输执行事件
✅ **自定义事件** - 创建自定义事件处理器
✅ **零运行时开销** - 泛型仅在编译时存在

## 示例：带事件监控的计数器

示例演示了一个计数器工作流，它能够：
- 将计数器递增 5 次
- 记录每个节点的执行
- 跟踪进度
- 实时流式传输事件

### 状态定义

```go
type CounterState struct {
    Count int      `json:"count"`     // 当前计数器值
    Name  string   `json:"name"`      // 工作流名称
    Logs  []string `json:"logs"`      // 执行日志
}
```

### 事件监听器

#### 1. 全局事件记录器

```go
type EventLogger struct{}

func (l *EventLogger) OnNodeEvent(ctx context.Context, event graph.NodeEvent,
                                 nodeName string, state CounterState, err error) {
    switch event {
    case graph.NodeEventStart:
        fmt.Printf("🔵 节点 '%s' 开始执行 (count=%d)\n", nodeName, state.Count)
    case graph.NodeEventComplete:
        fmt.Printf("🟢 节点 '%s' 执行完成 (count=%d)\n", nodeName, state.Count)
    case graph.NodeEventError:
        fmt.Printf("🔴 节点 '%s' 执行失败: %v\n", nodeName, err)
    }
}
```

#### 2. 进度跟踪器

```go
type ProgressTracker struct {
    totalNodes int
    completed  int
}

func (p *ProgressTracker) OnNodeEvent(ctx context.Context, event graph.NodeEvent,
                                     nodeName string, state CounterState, err error) {
    if event == graph.NodeEventComplete {
        p.completed++
        progress := float64(p.completed) / float64(p.totalNodes) * 100
        fmt.Printf("📊 进度: %.1f%% (%d/%d)\n", progress, p.completed, p.totalNodes)
    }
}
```

#### 3. 节点特定监听器

```go
// 为 increment 节点添加特殊监听器
incrementListener := graph.NodeListenerTypedFunc[CounterState](
    func(ctx context.Context, event graph.NodeEvent, nodeName string,
         state CounterState, err error) {
        if event == graph.NodeEventComplete {
            fmt.Printf("✨ 特殊通知：计数器现在是 %d！\n", state.Count)
        }
    },
)
```

### 创建可监听图

```go
// 创建类型化的可监听状态图
workflow := graph.NewListenableStateGraphTyped[CounterState]()

// 添加全局监听器
workflow.AddGlobalListener(&EventLogger{})
workflow.AddGlobalListener(&ProgressTracker{})

// 添加节点
incrementNode := workflow.AddNode("increment", "递增计数器",
    func(ctx context.Context, state CounterState) (CounterState, error) {
        state.Count++
        logMsg := fmt.Sprintf("计数器递增到 %d", state.Count)
        state.Logs = append(state.Logs, logMsg)
        time.Sleep(500 * time.Millisecond) // 模拟工作
        return state, nil
    })

// 添加节点特定监听器
incrementNode.AddListener(incrementListener)
```

### 编译和执行

```go
// 编译可监听图
runnable, err := workflow.CompileListenable()
if err != nil {
    log.Fatalf("编译图失败: %v", err)
}

// 执行图
finalState, err := runnable.Invoke(context.Background(), initialState)
if err != nil {
    log.Fatalf("图执行失败: %v", err)
}
```

### 流式执行

```go
// 创建流式监听器
streamingListener := &StreamingCounterListener{}
workflow.AddGlobalListener(streamingListener)

// 编译为流式执行
streamingRunnable, err := workflow.CompileListenable()

// 流式传输事件
eventChan := streamingRunnable.Stream(context.Background(), initialState)

// 处理事件
for event := range eventChan {
    switch event.Event {
    case graph.EventChainStart:
        fmt.Printf("🟢 流: 链开始\n")
    case graph.NodeEventStart:
        fmt.Printf("🔵 流: 节点 '%s' 开始\n", event.NodeName)
    case graph.NodeEventComplete:
        fmt.Printf("🟢 流: 节点 '%s' 完成\n", event.NodeName)
    case graph.EventChainEnd:
        fmt.Printf("🔴 流: 链结束\n")
    }
}
```

## 运行示例

```bash
cd examples/generic_state_graph_listenable
go run listenable_example.go
```

## 预期输出

```
🔧 编译图中...

🚀 开始图执行...
[17:36:33] 🔵 节点 'increment' 开始执行 (count=0)
  ✨ 特殊通知：计数器现在是 1！
[17:36:33] 🟢 节点 'increment' 执行完成 (count=1)
📊 进度: 20.0% (1/5)
[17:36:34] 🔵 节点 'increment' 开始执行 (count=1)
  ✨ 特殊通知：计数器现在是 2！
[17:36:34] 🟢 节点 'increment' 执行完成 (count=2)
📊 进度: 40.0% (2/5)
...
[17:36:36] 🟢 节点 'print' 执行完成 (count=5)

✅ 执行成功完成！
最终状态: {Count:5 Name:TypedCounter Logs:[计数器递增到 1 ...]}

--- 流式示例 ---
🎬 开始流式执行...
📡 接收事件:
[17:36:33.684] 🟢 流: 链开始
[17:36:33.684] 🔵 流: 节点 'increment' 开始
[17:36:34.185] 🟢 流: 节点 'increment' 完成 (count=1)
...
[17:36:36.189] 🔴 流: 链结束
```

## 事件类型

### 节点事件
- `NodeEventStart` - 节点开始执行
- `NodeEventComplete` - 节点成功完成执行
- `NodeEventError` - 节点执行失败

### 链事件
- `EventChainStart` - 图开始执行
- `EventChainEnd` - 图执行完成

### 流式事件
流式事件包含额外的元数据：
```go
type StreamEventTyped[S any] struct {
    Timestamp time.Time        // 事件发生时间
    Event     graph.NodeEvent  // 事件类型
    NodeName  string           // 节点名称
    State     S                // 当前状态
    Error     error            // 错误信息（如果有）
}
```

## API 参考

### 创建可监听图

```go
workflow := graph.NewListenableStateGraphTyped[你的状态类型]()
```

### 添加监听器

#### 全局监听器
```go
// 全局监听器接收所有节点的事件
workflow.AddGlobalListener(listener)
```

#### 节点特定监听器
```go
// 节点监听器只接收特定节点的事件
node := workflow.AddNode("nodeName", "描述", handler)
node.AddListener(specificListener)
```

#### 监听器接口
```go
type NodeListenerTyped[T any] interface {
    OnNodeEvent(ctx context.Context, event NodeEvent, nodeName string, state T, err error)
}

// 快速创建监听器，使用函数适配器
workflow.AddGlobalListener(
    graph.NodeListenerTypedFunc[你的状态类型](
        func(ctx context.Context, event NodeEvent, nodeName string, state 你的状态类型, err error) {
            // 处理事件
        },
    ),
)
```

### 编译方法

```go
// 编译为普通执行
runnable, err := workflow.CompileListenable()

// 使用配置编译
config := &graph.Config{ThreadID: "thread-123"}
runnable, err := workflow.CompileListenableWithConfig(config)
```

### 执行方法

```go
// 执行一次
finalState, err := runnable.Invoke(ctx, initialState)

// 流式传输事件
eventChan := runnable.Stream(ctx, initialState)
for event := range eventChan {
    // 处理事件
}

// 使用配置执行
finalState, err := runnable.InvokeWithConfig(ctx, initialState, config)
```

## 使用场景

### 1. **实时监控**

```go
type Monitor struct {
    metrics prometheus.Counter
}

func (m *Monitor) OnNodeEvent(ctx context.Context, event NodeEvent,
                             nodeName string, state MyState, err error) {
    switch event {
    case NodeEventComplete:
        m.metrics.Inc()
    case NodeEventError:
        log.Printf("节点 %s 失败: %v", nodeName, err)
    }
}
```

### 2. **调试和日志记录**

```go
type Debugger struct {
    logger *log.Logger
}

func (d *Debugger) OnNodeEvent(ctx context.Context, event NodeEvent,
                             nodeName string, state MyState, err error) {
    d.logger.Printf("[%s] %s: %+v", time.Now().Format(time.RFC3339), event, state)
}
```

### 3. **进度报告**

```go
type ProgressReporter struct {
    websocket *websocket.Conn
}

func (p *ProgressReporter) OnNodeEvent(ctx context.Context, event NodeEvent,
                                      nodeName string, state MyState, err error) {
    progress := map[string]any{
        "node": nodeName,
        "event": event,
        "timestamp": time.Now(),
    }
    p.websocket.WriteJSON(progress)
}
```

### 4. **条件逻辑**

```go
type ConditionalStopper struct {
    stopThreshold int
    stopChan      chan struct{}
}

func (c *ConditionalStopper) OnNodeEvent(ctx context.Context, event NodeEvent,
                                        nodeName string, state MyState, err error) {
    if state.Count >= c.stopThreshold {
        close(c.stopChan)  // 发送停止执行信号
    }
}
```

## 最佳实践

1. **保持监听器轻量** - 避免在监听器中进行阻塞操作
2. **使用缓冲通道** - 对于流式传输，使用适当大小的缓冲区
3. **优雅处理错误** - 不要让监听器错误导致图崩溃
4. **使用 Context** - 在监听器中尊重取消信号
5. **避免无限循环** - 谨慎使用修改状态的监听器

## 性能考虑

- 监听器对图执行的开销很小
- 每个监听器在与节点相同的 goroutine 中运行
- 对于高性能场景，考虑：
  - 使用通道进行异步处理
  - 实现监听器池
  - 采样事件而不是处理所有事件

## 相关示例

- [泛型 StateGraph](../generic_state_graph/) - 不带监听器的基础类型安全图
- [流式管道](../streaming_pipeline/) - 高级流式模式
- [检查点](../checkpointing/) - 持久化图状态

## 了解更多

- [LangGraphGo 文档](../../README.md)
- [Go 泛型](https://go.dev/blog/generics)
- [观察者模式](https://zh.wikipedia.org/wiki/观察者模式)