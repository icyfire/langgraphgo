# 内存管理策略

本示例演示了 LangGraphGo 中所有可用的 **9 种内存管理策略**。每种策略以不同的方式优化对话上下文，让你能够为特定用例选择最佳方法。

## 概述

有效的内存管理对于 AI Agent 至关重要，可以：
- **控制成本** - 通过最小化 token 使用
- **保持上下文** - 维持连贯的对话
- **扩展到长对话** - 不丢失重要信息
- **优化性能** - 实现实时应用程序

## 9 种内存策略

### 1. Sequential Memory - 顺序内存（保留所有）
**存储所有消息，不进行优化**

```go
mem := memory.NewSequentialMemory()
```

**适用于：**
- 短对话（< 20 条消息）
- 原型开发和测试
- 当成本不是问题时

**特点：**
- ✅ 完美的记忆保留
- ✅ 实现简单
- ❌ Token 线性增长
- ❌ 可能超出上下文窗口

### 2. Sliding Window Memory - 滑动窗口
**只保留最近的 N 条消息**

```go
mem := memory.NewSlidingWindowMemory(10) // 保留最近 10 条消息
```

**适用于：**
- 聊天应用程序
- 实时对话
- 最近上下文最重要

**特点：**
- ✅ 可预测的内存使用
- ✅ 快速检索
- ❌ 丢失旧的的重要消息
- ❌ 固定的上下文大小

### 3. Buffer Memory - 缓冲内存
**灵活的消息或 token 限制，可选自动摘要**

```go
mem := memory.NewBufferMemory(&memory.BufferConfig{
    MaxMessages:   50,     // 或 MaxTokens: 2000
    AutoSummarize: true,   // 超限时压缩
    Summarizer:    customSummarizer,
})
```

**适用于：**
- 通用应用程序
- 需要灵活性时
- 自动内存管理

**特点：**
- ✅ 可配置的限制
- ✅ 可选摘要
- ⚠️ 需要摘要策略
- ⚠️ 可能丢失一些上下文

### 4. Summarization Memory - 摘要内存
**压缩旧消息同时保留最近的**

```go
mem := memory.NewSummarizationMemory(&memory.SummarizationConfig{
    RecentWindowSize: 5,     // 保留最近 5 条消息
    SummarizeAfter:   15,    // 15 条消息后开始摘要
    Summarizer:       llmSummarizer,
})
```

**适用于：**
- 长对话
- 当历史上下文有价值时
- 成本敏感的应用程序

**特点：**
- ✅ 维护关键信息
- ✅ 控制 token 增长
- ❌ 需要 LLM 调用进行摘要
- ❌ 处理延迟

### 5. Retrieval Memory - 检索内存
**基于相似度检索相关消息**

```go
mem := memory.NewRetrievalMemory(&memory.RetrievalConfig{
    TopK: 5, // 返回最相关的 5 条消息
    EmbeddingFunc: func(ctx context.Context, text string) ([]float64, error) {
        return createEmbedding(text), nil
    },
    SimilarityThreshold: 0.7,
})
```

**适用于：**
- 知识密集型应用程序
- 当特定主题重要时
- 大型对话历史

**特点：**
- ✅ 精确的上下文检索
- ✅ 扩展到大型历史
- ❌ 需要嵌入模型
- ❌ 计算开销

### 6. Hierarchical Memory - 分层内存
**将消息分为最近和重要两层**

```go
mem := memory.NewHierarchicalMemory(&memory.HierarchicalConfig{
    RecentLimit: 10,   // 最近消息层
    ImportantLimit: 20, // 重要消息层
    ImportanceScorer: func(msg *memory.Message) float64 {
        score := 0.5
        if msg.Role == "system" {
            score += 0.3
        }
        if containsKeywords(msg.Content, []string{"重要", "决策"}) {
            score += 0.3
        }
        return score
    },
})
```

**适用于：**
- 复杂对话
- 当消息重要性不同时
- 上下文保留

**特点：**
- ✅ 保留重要上下文
- ✅ 平衡时效性和重要性
- ❌ 需要重要性评分
- ❌ 更复杂的逻辑

### 7. Graph-Based Memory - 图内存
**构建消息关系图**

```go
mem := memory.NewGraphBasedMemory(&memory.GraphConfig{
    TopK: 5,
    RelationExtractor: func(msg *memory.Message) []string {
        return extractTopicsAndEntities(msg.Content)
    },
    RelationScorer: func(topic1, topic2 string) float64 {
        return calculateTopicSimilarity(topic1, topic2)
    },
})
```

**适用于：**
- 关系密集的对话
- 基于主题的讨论
- 复杂的信息网络

**特点：**
- ✅ 维护关系
- ✅ 基于主题的检索
- ❌ 更高的计算成本
- ❌ 复杂的实现

### 8. Compression Memory - 压缩内存
**智能压缩消息块**

```go
mem := memory.NewCompressionMemory(&memory.CompressionConfig{
    CompressionTrigger: 20,    // 20 条消息后触发
    CompressionRatio:  0.3,    // 目标压缩到 30%
    ConsolidateAfter: time.Hour, // 1 小时后合并压缩块
    Compressor: func(ctx context.Context, msgs []*memory.Message) (*memory.CompressedBlock, error) {
        return llmCompressor(msgs), nil
    },
})
```

**适用于：**
- 非常长的对话
- 成本关键的应用程序
- 当内存效率至关重要时

**特点：**
- ✅ 最大压缩
- ✅ 可配置的压缩比
- ❌ 处理开销
- ❌ 可能的信息丢失

### 9. OS-Like Memory - 操作系统式内存
**模拟操作系统的内存管理**

```go
mem := memory.NewOSLikeMemory(&memory.OSLikeConfig{
    ActiveLimit:  10,            // 活跃页面
    CacheLimit:   20,            // 缓存页面
    AccessWindow: time.Minute * 5, // 访问跟踪窗口
    PageStrategy: memory.LRU,    // LRU 替换策略
})
```

**适用于：**
- 企业应用程序
- 复杂的内存管理需求
- 需要细粒度控制时

**特点：**
- ✅ 高级内存管理
- ✅ 多种策略可用
- ❌ 最复杂的实现
- ❌ 最高的开销

## 运行示例

```bash
cd examples/memory
go run memory_examples.go
```

## 预期输出

```
=== LangGraphGo 内存策略演示 ===

1. Sequential Memory (顺序内存)
  存储消息数: 4
  总 Token 数: 65
  最新消息: 支持并行执行、状态管理、持久化等

2. Sliding Window Memory (滑动窗口)
  添加消息 1
  添加消息 2
  添加消息 3
  添加消息 4
  添加消息 5
  实际保留: 3 条消息
  消息列表: [消息 3, 消息 4, 消息 5]

3. Buffer Memory (缓冲内存)
  活跃消息: 5
  活跃 Tokens: 195

4. Summarization Memory (摘要内存)
  总消息数: 6
  压缩率: 50.00%
  上下文构成: 1 条摘要 + 3 条最近消息

5. Retrieval Memory (检索内存)
  查询 '价格信息': 找到 3 条相关消息
  查询 '技术架构': 找到 3 条相关消息
  查询 '团队规模': 找到 3 条相关消息

6. Hierarchical Memory (分层内存)
  活跃消息: 6 条
  总消息: 6 条
  总 Tokens: 84

7. Graph-Based Memory (图内存)
  查询 '产品': 找到 3 条相关消息

8. Compression Memory (压缩内存)
  活跃消息: 5
  压缩率: 80.00%

9. OS-Like Memory (操作系统式内存)
  活跃消息: 3
  总消息: 10
  压缩率: 70.00%

=== 性能对比 ===
测试场景: 50 条消息的性能对比

策略            存储消息      活跃消息      Token 效率    响应时间
----------------------------------------------------------------------
Sequential      50           50           1.00          1.2ms
SlidingWindow    50           10           0.20          0.8ms
Buffer          50           20           0.40          1.0ms
Hierarchical     50           15           0.30          1.1ms
```

## 性能对比

示例包含了一个性能对比，用 50 条消息测试每种策略：

| 策略 | 存储消息 | 活跃消息 | Token 效率 | 响应时间 |
|------|----------|----------|------------|----------|
| Sequential | 50 | 50 | 1.00 | 1.2ms |
| SlidingWindow | 50 | 10 | 0.20 | 0.8ms |
| Buffer | 50 | 20 | 0.40 | 1.0ms |
| Hierarchical | 50 | 15 | 0.30 | 1.1ms |

## 选择合适的策略

### 快速决策指南

```go
func selectMemoryStrategy(conversationLength int, hasKnowledgeBase bool, costSensitive bool) string {
    switch {
    case conversationLength < 10:
        return "sequential"      // 短对话
    case conversationLength < 30 && !costSensitive:
        return "sliding_window" // 中等对话
    case hasKnowledgeBase:
        return "retrieval"       // 知识密集型
    case costSensitive && conversationLength > 100:
        return "compression"     // 成本关键
    default:
        return "hierarchical"    // 通用目的
    }
}
```

### 策略选择矩阵

| 用例 | 推荐策略 | 原因 |
|------|----------|------|
| 简单聊天 | Sliding Window | 最近上下文最重要 |
| 客户支持 | Hierarchical | 保留重要交互 |
| 知识库问答 | Retrieval | 精确的主题检索 |
| 长文档 | Summarization | 压缩同时保留 |
| 成本敏感应用 | Compression | 最大效率 |
| 企业系统 | OS-Like | 高级控制 |

## 实现技巧

### 自定义摘要器

```go
func customSummarizer(ctx context.Context, msgs []*memory.Message) (string, error) {
    // 提取关键主题
    topics := extractTopics(msgs)
    decisions := extractDecisions(msgs)

    // 构建结构化摘要
    summary := fmt.Sprintf("摘要 (%d 条消息):\n", len(msgs))
    summary += fmt.Sprintf("主题: %s\n", strings.Join(topics, ", "))
    if len(decisions) > 0 {
        summary += fmt.Sprintf("决策: %s", strings.Join(decisions, "; "))
    }

    return summary, nil
}
```

### 自定义重要性评分器

```go
func customImportanceScorer(msg *memory.Message) float64 {
    score := 0.5

    // 基于角色的评分
    if msg.Role == "system" {
        score += 0.3
    }

    // 基于关键词的评分
    importantKeywords := []string{"决策", "行动", "截止日期", "错误"}
    for _, kw := range importantKeywords {
        if strings.Contains(strings.ToLower(msg.Content), kw) {
            score += 0.1
        }
    }

    // 基于长度的评分（较长的消息可能更重要）
    if msg.TokenCount > 100 {
        score += 0.1
    }

    return math.Min(score, 1.0)
}
```

### 性能优化

```go
// 使用预训练的嵌入进行检索
func optimizedEmbedding(text string) ([]float64, error) {
    // 缓存嵌入以避免重新计算
    if cached, ok := embeddingCache.Get(text); ok {
        return cached, nil
    }

    embedding, err := embeddingModel.Create(text)
    if err == nil {
        embeddingCache.Set(text, embedding)
    }

    return embedding, err
}
```

## 最佳实践

1. **先分析配置**：在选择之前测量实际使用模式
2. **从简单开始**：从 SlidingWindow 或 Hierarchical 开始
3. **监控性能**：跟踪 token 使用和响应时间
4. **测试组合**：为复杂需求混合策略
5. **考虑 LLM 成本**：考虑摘要的 API 成本

## 高级用法

### 多策略组合

```go
type HybridMemory struct {
    recent    memory.Memory
    important memory.Memory
    search    memory.Memory
}

func (h *HybridMemory) GetContext(ctx context.Context, query string) ([]*memory.Message, error) {
    // 获取最近的消息
    recent, _ := h.recent.GetContext(ctx, query)

    // 获取重要的消息
    important, _ := h.important.GetContext(ctx, query)

    // 获取相关的消息
    relevant, _ := h.search.GetContext(ctx, query)

    // 合并和去重
    return mergeAndDeduplicate(recent, important, relevant), nil
}
```

### 动态策略切换

```go
type AdaptiveMemory struct {
    current memory.Memory
    metrics  *MemoryMetrics
}

func (a *AdaptiveMemory) AddMessage(ctx context.Context, msg *memory.Message) error {
    // 检查是否需要切换策略
    if a.metrics.AvgResponseTime > time.Second*2 {
        a.switchStrategy("sliding_window")
    }

    return a.current.AddMessage(ctx, msg)
}
```

## 相关示例

- [内存集成](../memory_graph_integration/) - LangGraph 工作流中的内存
- [带内存的聊天 Agent](../memory_agent/) - 完整的聊天应用
- [上下文存储](../context_store/) - 高级上下文管理

## 了解更多

- [内存包文档](../../memory/README.md)
- [LangGraphGo 文档](../../README.md)
- [Agent 内存策略](../../tutorial/001-agent-memory-strategies.md)