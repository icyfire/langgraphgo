# 使用内存向量存储的 RAG 示例

本示例演示如何使用 **LangGraphGo 的内存向量存储**构建 **RAG（检索增强生成）管道**，用于快速原型开发和测试，无需外部依赖。

## 什么是内存向量存储？

内存向量存储是一个轻量级、临时的存储解决方案，具有以下特点：
- **完全在内存中运行** - 无需数据库设置
- **非常适合测试** - 理想的开发和实验工具
- **快速且简单** - 零配置，即时启动
- **嵌入无关** - 适用于任何嵌入器（OpenAI、mock、自定义）
- **非持久化** - 程序退出时数据丢失

## 前置要求

1. **Go 1.21+**：构建示例所必需
2. **API 密钥**（可选）：用于使用真实的 LLM
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```
   **注意**：此示例默认使用 mock 嵌入器，因此基本测试不需要 API 密钥。

## 运行示例

```bash
cd examples/rag_memory_example
go run main.go
```

## 功能演示

- 使用 mock 嵌入创建内存向量存储
- 添加文档并自动生成嵌入
- 构建完整的带记忆的 RAG 管道
- 使用自然语言问题查询管道
- 可视化 RAG 管道图结构
- 理解检索和生成流程

## 代码要点

### 创建 Mock 嵌入器和内存存储

```go
import (
    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/store"
)

// 创建 mock 嵌入器（128 维向量用于测试）
embedder := store.NewMockEmbedder(128)

// 创建内存向量存储（临时的，无持久化）
vectorStore := store.NewInMemoryVectorStore(embedder)
```

### 添加文档

```go
// 创建示例文档
documents := []rag.Document{
    {
        Content: "Chroma 是一个开源向量数据库...",
        Metadata: map[string]any{"source": "chroma_docs"},
    },
    {
        Content: "LangGraphGo 集成了各种向量存储...",
        Metadata: map[string]any{"source": "langgraphgo_docs"},
    },
}

// 添加文档和嵌入
vectorStore.Add(ctx, documents)
```

### 构建 RAG 管道

```go
import "github.com/smallnest/langgraphgo/rag/retriever"

// 创建检索器
retriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, 2)

// 配置 RAG 管道
config := rag.DefaultPipelineConfig()
config.Retriever = retriever
config.LLM = llm

// 构建并编译管道
pipeline := rag.NewRAGPipeline(config)
err = pipeline.BuildBasicRAG()
runnable, err := pipeline.Compile()
```

### 查询管道

```go
result, err := runnable.Invoke(ctx, map[string]any{
    "query": "什么是 Chroma？",
})

if answer, ok := result["answer"].(string); ok {
    fmt.Printf("回答: %s\n", answer)
}

if docs, ok := result["documents"].([]rag.RAGDocument); ok {
    for _, doc := range docs {
        fmt.Printf("检索到: %s\n", doc.Content)
    }
}
```

## 预期输出

```
    +-------------------------------------------------------------------+
    |                              ____________                          |
    |                             |            |                         |
    |                             |   开始     |                         |
    |                             |____________|                         |
    |                                    |                               |
    |                                    v                               |
    |    ___________________________     __________________             |
    |   |                           |   |                  |            |
    |   |      检索文档             |   |      条件       |            |
    |   |___________________________|   |__________________|            |
    |                                    |                               |
    |              _______________________|_____________________         |
    |             |                                           |          |
    |     ________v________                          _________v______    |
    |    |                 |                        |                |   |
    |    |    生成回答     |                        |      结束      |   |
    |    |_________________|                        |________________|   |
    |                                                                   |
    +-------------------------------------------------------------------+

查询: 什么是 Chroma？

检索到的文档:
  [1] Chroma 是一个开源向量数据库，允许您存储和查询嵌入...
  [2] LangGraphGo 集成了各种向量存储，包括 Chroma...

回答: Chroma 是一个开源向量数据库，专为存储和查询嵌入而设计...
```

## 内存存储 vs 持久化向量存储

| 特性 | 内存存储 | 持久化存储（Chroma 等） |
|---------|----------------|----------------------------------|
| **设置** | 零配置 | 需要服务器/集群 |
| **持久化** | 退出时丢失 | 持久存储 |
| **用例** | 测试、原型开发 | 生产环境 |
| **性能** | 快速（本地） | 网络延迟 |
| **可扩展性** | 受内存限制 | 水平扩展 |
| **并发性** | 单进程 | 多客户端 |

## 何时使用内存存储

✅ **适合**：
- 快速原型开发和实验
- RAG 管道的单元测试
- 学习 RAG 概念
- 无外部依赖的开发
- 基准测试和性能测试

❌ **不适合**：
- 生产应用程序
- 大型文档集合
- 多用户场景
- 长期数据保留

## 自定义示例

### 使用真实的嵌入

将 mock 嵌入器替换为 OpenAI：

```go
import (
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/openai"
)

// 创建 OpenAI LLM 和嵌入器
llm, _ := openai.New()
openaiEmbedder, _ := embeddings.NewEmbedder(llm)
embedder := rag.NewLangChainEmbedder(openaiEmbedder)

// 将真实嵌入器与内存存储一起使用
vectorStore := store.NewInMemoryVectorStore(embedder)
```

### 调整检索参数

```go
// 检索更多文档
retriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, 5)

// 或使用自定义检索配置
retriever := retriever.NewVectorStoreRetrieverWithConfig(
    vectorStore,
    embedder,
    &retriever.Config{
        K: 3,
        ScoreThreshold: 0.7,
    },
)
```

### 添加更多文档

```go
documents := []rag.Document{
    {
        Content: "您的自定义文档在这里...",
        Metadata: map[string]any{
            "source": "custom",
            "category": "testing",
        },
    },
    // 添加更多文档...
}
vectorStore.Add(ctx, documents)
```

## 故障排除

### 缺少 API 密钥错误

**错误**: `Failed to create LLM: OPENAI_API_KEY not found`

**解决方案**：选择以下之一：
1. 设置您的 OpenAI API 密钥：`export OPENAI_API_KEY="your-key"`
2. 使用 mock LLM 进行测试（修改代码）

### 检索结果为空

**问题**: 检索到的文档列表为空

**解决方案**：
- 确保在查询之前添加了文档
- 检查嵌入是否成功生成
- 验证查询文本与存储的文档相关

### 构建错误

**错误**: `package github.com/smallnest/langgraphgo/... not found`

**解决方案**：从 langgraphgo 根目录运行：
```bash
cd /path/to/langgraphgo
go run examples/rag_memory_example/main.go
```

## 下一步

- **尝试持久化存储**：将内存存储替换为 Chroma、Pinecone 或 Weaviate
- **添加元数据过滤**：在检索期间按元数据过滤文档
- **实现混合搜索**：结合关键词和向量搜索
- **添加流式输出**：使用流式响应进行实时反馈
- **多轮对话**：为 RAG 管道添加对话记忆

## 高级 RAG 模式

### 带元数据过滤的检索

```go
// 存储带元数据的文档
documents := []rag.Document{
    {
        Content: "...",
        Metadata: map[string]any{
            "category": "technical",
            "language": "zh",
        },
    },
}

// 稍后，使用元数据过滤器搜索（如果存储支持）
results, _ := vectorStore.SearchWithFilter(ctx, queryEmbedding, k, map[string]any{
    "category": "technical",
})
```

### 自定义检索策略

```go
// 实现自定义检索逻辑
type CustomRetriever struct {
    vectorStore *store.InMemoryVectorStore
    embedder    rag.Embedder
}

func (r *CustomRetriever) Retrieve(ctx context.Context, query string, k int) ([]rag.Document, error) {
    // 自定义检索逻辑在这里
    // 例如：重新排序、过滤、转换
}
```

## 参考资料

- [LangGraphGo RAG 文档](../../docs/RAG/RAG_CN.md)
- [向量存储接口](../../rag/types.go)
- [RAG 管道配置](../../rag/pipeline.go)
- [检索器模式](../../rag/retriever/)
- [Chroma 示例](../chroma-v2-example/README_CN.md)
- [本地文件示例](../rag_local_files_example/README_CN.md)
