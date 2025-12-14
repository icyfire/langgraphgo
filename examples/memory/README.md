# Memory Management Strategies

This example demonstrates all **9 memory management strategies** available in LangGraphGo. Each strategy optimizes conversation context differently, allowing you to choose the best approach for your specific use case.

## Overview

Effective memory management is crucial for AI agents to:
- **Control costs** by minimizing token usage
- **Maintain context** for coherent conversations
- **Scale to long conversations** without losing important information
- **Optimize performance** for real-time applications

## The 9 Memory Strategies

### 1. Sequential Memory - Keep-It-All
**Stores all messages without optimization**

```go
mem := memory.NewSequentialMemory()
```

**Best for:**
- Short conversations (< 20 messages)
- Prototyping and testing
- When cost is not a concern

**Characteristics:**
- ✅ Perfect memory retention
- ✅ Simple implementation
- ❌ Linear token growth
- ❌ May exceed context window

### 2. Sliding Window Memory - Recent Messages Only
**Keeps only the most recent N messages**

```go
mem := memory.NewSlidingWindowMemory(10) // Keep last 10 messages
```

**Best for:**
- Chat applications
- Real-time conversations
- Recent context matters most

**Characteristics:**
- ✅ Predictable memory usage
- ✅ Fast retrieval
- ❌ Loses old important messages
- ❌ Fixed context size

### 3. Buffer Memory - Flexible Limits
**Flexible message or token limits with optional auto-summarization**

```go
mem := memory.NewBufferMemory(&memory.BufferConfig{
    MaxMessages:   50,     // Or MaxTokens: 2000
    AutoSummarize: true,   // Compress when limit exceeded
    Summarizer:    customSummarizer,
})
```

**Best for:**
- General-purpose applications
- When you need flexibility
- Automatic memory management

**Characteristics:**
- ✅ Configurable limits
- ✅ Optional summarization
- ⚠️ Requires summarization strategy
- ⚠️ May lose some context

### 4. Summarization Memory - Smart Compression
**Compresses old messages while preserving recent ones**

```go
mem := memory.NewSummarizationMemory(&memory.SummarizationConfig{
    RecentWindowSize: 5,     // Keep last 5 messages
    SummarizeAfter:   15,    // Start summarizing after 15 messages
    Summarizer:       llmSummarizer,
})
```

**Best for:**
- Long conversations
- When historical context is valuable
- Cost-sensitive applications

**Characteristics:**
- ✅ Maintains key information
- ✅ Controls token growth
- ❌ Requires LLM calls for summarization
- ❌ Processing latency

### 5. Retrieval Memory - Similarity-Based
**Retrieves relevant messages based on similarity**

```go
mem := memory.NewRetrievalMemory(&memory.RetrievalConfig{
    TopK: 5, // Return top 5 most relevant messages
    EmbeddingFunc: func(ctx context.Context, text string) ([]float64, error) {
        return createEmbedding(text), nil
    },
    SimilarityThreshold: 0.7,
})
```

**Best for:**
- Knowledge-intensive applications
- When specific topics matter
- Large conversation histories

**Characteristics:**
- ✅ Precise context retrieval
- ✅ Scales to large histories
- ❌ Requires embedding model
- ❌ Computational overhead

### 6. Hierarchical Memory - Layered Storage
**Separates messages into recent and important layers**

```go
mem := memory.NewHierarchicalMemory(&memory.HierarchicalConfig{
    RecentLimit: 10,   // Recent messages layer
    ImportantLimit: 20, // Important messages layer
    ImportanceScorer: func(msg *memory.Message) float64 {
        score := 0.5
        if msg.Role == "system" {
            score += 0.3
        }
        if containsKeywords(msg.Content, []string{"important", "decision"}) {
            score += 0.3
        }
        return score
    },
})
```

**Best for:**
- Complex conversations
- When message importance varies
- Context preservation

**Characteristics:**
- ✅ Preserves important context
- ✅ Balances recency and importance
- ❌ Requires importance scoring
- ❌ More complex logic

### 7. Graph-Based Memory - Relationship Mapping
**Builds a graph of message relationships**

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

**Best for:**
- Relationship-dense conversations
- Topic-based discussions
- Complex information networks

**Characteristics:**
- ✅ Maintains relationships
- ✅ Topic-based retrieval
- ❌ Higher computational cost
- ❌ Complex implementation

### 8. Compression Memory - Intelligent Compression
**Intelligently compresses message blocks**

```go
mem := memory.NewCompressionMemory(&memory.CompressionConfig{
    CompressionTrigger: 20,    // Trigger after 20 messages
    CompressionRatio:  0.3,    // Target 30% of original size
    ConsolidateAfter: time.Hour, // Merge compressed blocks
    Compressor: func(ctx context.Context, msgs []*memory.Message) (*memory.CompressedBlock, error) {
        return llmCompressor(msgs), nil
    },
})
```

**Best for:**
- Very long conversations
- Cost-critical applications
- When memory efficiency is paramount

**Characteristics:**
- ✅ Maximum compression
- ✅ Configurable compression ratio
- ❌ Processing overhead
- ❌ Potential information loss

### 9. OS-Like Memory - Operating System Management
**Mimics OS memory management with paging**

```go
mem := memory.NewOSLikeMemory(&memory.OSLikeConfig{
    ActiveLimit:  10,            // Active pages
    CacheLimit:   20,            // Cached pages
    AccessWindow: time.Minute * 5, // Access tracking window
    PageStrategy: memory.LRU,    // LRU replacement strategy
})
```

**Best for:**
- Enterprise applications
- Complex memory management needs
- When fine-grained control is required

**Characteristics:**
- ✅ Advanced memory management
- ✅ Multiple strategies available
- ❌ Most complex implementation
- ❌ Highest overhead

## Running the Example

```bash
cd examples/memory
go run memory_examples.go
```

## Expected Output

```
=== LangGraphGo Memory Strategies Demo ===

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

=== Performance Comparison ===
测试场景: 50 条消息的性能对比

策略            存储消息      活跃消息      Token 效率    响应时间
----------------------------------------------------------------------
Sequential      50           50           1.00          1.2ms
SlidingWindow    50           10           0.20          0.8ms
Buffer          50           20           0.40          1.0ms
Hierarchical     50           15           0.30          1.1ms
```

## Performance Comparison

The example includes a performance comparison that tests each strategy with 50 messages:

| Strategy | Messages Stored | Active Messages | Token Efficiency | Response Time |
|----------|----------------|-----------------|-----------------|---------------|
| Sequential | 50 | 50 | 1.00 | 1.2ms |
| SlidingWindow | 50 | 10 | 0.20 | 0.8ms |
| Buffer | 50 | 20 | 0.40 | 1.0ms |
| Hierarchical | 50 | 15 | 0.30 | 1.1ms |

## Choosing the Right Strategy

### Quick Decision Guide

```go
func selectMemoryStrategy(conversationLength int, hasKnowledgeBase bool, costSensitive bool) string {
    switch {
    case conversationLength < 10:
        return "sequential"      // Short conversations
    case conversationLength < 30 && !costSensitive:
        return "sliding_window" // Medium conversations
    case hasKnowledgeBase:
        return "retrieval"       // Knowledge-intensive
    case costSensitive && conversationLength > 100:
        return "compression"     // Cost-critical
    default:
        return "hierarchical"    // General purpose
    }
}
```

### Strategy Selection Matrix

| Use Case | Recommended Strategy | Reason |
|----------|---------------------|---------|
| Simple chat | Sliding Window | Recent context matters most |
| Customer support | Hierarchical | Preserves important interactions |
| Knowledge base Q&A | Retrieval | Precise topic retrieval |
| Long documentation | Summarization | Compresses while preserving |
| Cost-sensitive app | Compression | Maximum efficiency |
| Enterprise system | OS-Like | Advanced control |

## Implementation Tips

### Custom Summarizer

```go
func customSummarizer(ctx context.Context, msgs []*memory.Message) (string, error) {
    // Extract key topics
    topics := extractTopics(msgs)
    decisions := extractDecisions(msgs)

    // Build structured summary
    summary := fmt.Sprintf("Summary (%d msgs):\n", len(msgs))
    summary += fmt.Sprintf("Topics: %s\n", strings.Join(topics, ", "))
    if len(decisions) > 0 {
        summary += fmt.Sprintf("Decisions: %s", strings.Join(decisions, "; "))
    }

    return summary, nil
}
```

### Custom Importance Scorer

```go
func customImportanceScorer(msg *memory.Message) float64 {
    score := 0.5

    // Role-based scoring
    if msg.Role == "system" {
        score += 0.3
    }

    // Keyword-based scoring
    importantKeywords := []string{"decision", "action", "deadline", "bug"}
    for _, kw := range importantKeywords {
        if strings.Contains(strings.ToLower(msg.Content), kw) {
            score += 0.1
        }
    }

    // Length-based scoring (longer messages might be more important)
    if msg.TokenCount > 100 {
        score += 0.1
    }

    return math.Min(score, 1.0)
}
```

### Performance Optimization

```go
// Use pre-trained embeddings for retrieval
func optimizedEmbedding(text string) ([]float64, error) {
    // Cache embeddings to avoid recomputation
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

## Best Practices

1. **Profile First**: Measure actual usage patterns before choosing
2. **Start Simple**: Begin with SlidingWindow or Hierarchical
3. **Monitor Performance**: Track token usage and response times
4. **Test Combinations**: Mix strategies for complex needs
5. **Consider LLM Costs**: Factor in API costs for summarization

## Advanced Usage

### Multi-Strategy Composition

```go
type HybridMemory struct {
    recent    memory.Memory
    important memory.Memory
    search    memory.Memory
}

func (h *HybridMemory) GetContext(ctx context.Context, query string) ([]*memory.Message, error) {
    // Get recent messages
    recent, _ := h.recent.GetContext(ctx, query)

    // Get important messages
    important, _ := h.important.GetContext(ctx, query)

    // Get relevant messages
    relevant, _ := h.search.GetContext(ctx, query)

    // Merge and deduplicate
    return mergeAndDeduplicate(recent, important, relevant), nil
}
```

### Dynamic Strategy Switching

```go
type AdaptiveMemory struct {
    current memory.Memory
    metrics  *MemoryMetrics
}

func (a *AdaptiveMemory) AddMessage(ctx context.Context, msg *memory.Message) error {
    // Check if we need to switch strategies
    if a.metrics.AvgResponseTime > time.Second*2 {
        a.switchStrategy("sliding_window")
    }

    return a.current.AddMessage(ctx, msg)
}
```

## Related Examples

- [Memory Integration](../memory_graph_integration/) - Memory in LangGraph workflows
- [Chat Agent with Memory](../memory_agent/) - Complete chat application
- [Context Store](../context_store/) - Advanced context management

## Learn More

- [Memory Package Documentation](../../memory/README.md)
- [LangGraphGo Documentation](../../README.md)
- [Agent Memory Strategies](../../tutorial/001-agent-memory-strategies.md)