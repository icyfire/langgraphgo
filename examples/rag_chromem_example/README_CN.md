# RAG 与 chromem-go 向量数据库示例

本示例演示如何将 **chromem-go 向量数据库** 与 **LangGraphGo 的 RAG 管道**结合使用。

## 什么是 chromem-go？

chromem-go 是一个受 Chroma 启发的纯 Go 实现的向量数据库。它提供：

- **零外部依赖** - 无需 Docker，无需外部服务
- **嵌入式 SQLite 存储** - 支持持久化存储，也可选内存模式
- **线程安全操作** - 支持多个 goroutine 并发安全访问
- **并发文档处理** - 支持工作池批量操作
- **多种嵌入函数** - 支持各种嵌入提供商
- **余弦相似度搜索** - 高效的向量相似度查询
- **元数据过滤** - 根据文档元数据过滤搜索结果
- **原生 Go 实现** - 完全用 Go 编写，无 CGo 依赖

## 前置要求

1. **Go 1.21+**: 构建示例所需
2. **OpenAI API Key**: 同时用于 LLM 调用和嵌入向量生成
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```

**注意**: 本示例使用真实的 OpenAI 嵌入（默认使用 `text-embedding-3-small`）来生成文档的向量表示。

## 运行示例

```bash
cd examples/rag_chromem_example
go run main.go
```

## 功能演示

- 创建带持久化存储的 chromem-go 向量存储
- 添加文档并自动生成嵌入向量
- 使用 chromem-go 作为向量存储构建 RAG 管道
- 使用自然语言问题查询管道
- 带相关性分数的相似度搜索
- 元数据过滤实现定向搜索
- 跨存储实例的持久化存储验证
- 存储统计信息和集合管理

## 代码亮点

### 创建 OpenAI 嵌入器

```go
import (
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/smallnest/langgraphgo/rag"
)

// 创建用于嵌入的 OpenAI LLM
llm, err := openai.New()
if err != nil {
    log.Fatal(err)
}

// 创建 OpenAI 嵌入器（默认使用 text-embedding-3-small）
openaiEmbedder, err := embeddings.NewEmbedder(llm)
if err != nil {
    log.Fatal(err)
}

// 使用 LangGraphGo 适配器包装
embedder := rag.NewLangChainEmbedder(openaiEmbedder)
```

### 创建 chromem-go 存储

```go
// 使用内存存储的简单初始化
store, err := store.NewChromemVectorStoreSimple("", embedder)

// 或使用完整配置
store, err := store.NewChromemVectorStore(store.ChromemConfig{
    PersistenceDir: "/path/to/storage",  // 空字符串表示内存模式
    CollectionName: "my_collection",     // 默认为 "default"
    Embedder:       embedder,            // 必需的嵌入函数
})
```

### 添加文档

```go
// 文档可以在没有预计算嵌入向量的情况下添加
// 嵌入器将自动生成嵌入向量
documents := []rag.Document{
    {
        ID:      "doc1",
        Content: "你的文档内容",
        Metadata: map[string]any{
            "category": "tech",
            "source":   "docs",
        },
        // 嵌入向量是可选的 - 如果未提供将自动生成
    },
}

err = vectorStore.Add(ctx, documents)
```

### 相似度搜索

```go
// 使用查询嵌入向量进行基本搜索
results, err := vectorStore.Search(ctx, queryEmbedding, k)

// 使用元数据过滤进行搜索
results, err := vectorStore.SearchWithFilter(ctx, queryEmbedding, k, map[string]any{
    "category": "tech",
})

for _, result := range results {
    fmt.Printf("分数: %.4f - %s\n", result.Score, result.Document.Content)
}
```

### 持久化存储

```go
// 创建持久化存储
store, err := store.NewChromemVectorStoreSimple("/path/to/db", embedder)
defer store.Close()

// 添加文档
store.Add(ctx, documents)

// 稍后重新打开相同的存储
store2, err := store.NewChromemVectorStoreSimple("/path/to/db", embedder)
// 所有文档都已持久化！
```

### 与 RAG 管道一起使用

```go
// 创建检索器
retriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, topK)

// 配置 RAG 管道
config := rag.DefaultPipelineConfig()
config.Retriever = retriever
config.LLM = llm

// 构建和编译
pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()
runnable, _ := pipeline.Compile()

// 查询
result, _ := runnable.Invoke(ctx, map[string]any{"query": "什么是 chromem-go？"})
```

## 配置选项

### ChromemConfig

| 选项 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| `PersistenceDir` | string | SQLite 存储目录（空值 = 内存模式） | "" |
| `CollectionName` | string | 集合名称 | "default" |
| `Embedder` | Embedder | 用于生成向量的嵌入函数 | 必需 |

## 存储模式

### 内存存储

```go
// 快速、临时存储
store, err := store.NewChromemVectorStoreSimple("", embedder)
```

**使用场景：**
- 测试和开发
- 临时数据处理
- 缓存层

### 持久化存储

```go
// 数据跨重启持久化
store, err := store.NewChromemVectorStoreSimple("/data/vectors", embedder)
```

**使用场景：**
- 生产应用
- 长期数据保留
- 分布式系统

## 预期输出

```
=== RAG 与 chromem-go 向量存储示例 ===

初始化 chromem-go 向量存储...
集合创建成功: langgraphgo_example
存储位置: /tmp/chromem_example

正在添加文档到 chromem-go...
成功添加 5 个文档

存储统计信息:
  总文档数: 5
  向量维度: 128

正在构建 RAG 管道...

管道图:
┌───────────────────────────────────────────────┐
│                   RAG Pipeline                 │
└───────────────────────────────────────────────┘

================================================================================
查询 1: 什么是 chromem-go？
--------------------------------------------------------------------------------

检索到 2 个文档:
  [1] 分数: 0.8542
      chromem-go 是一个受 Chroma 启发的纯 Go 实现的向量数据库...
      元数据: map[category:introduction source:chromem_docs]
  [2] 分数: 0.7821
      LangGraphGo 与 chromem-go 无缝集成，提供原生 Go 向量存储...
      元数据: map[category:integration source:langgraphgo_docs]

回答:
chromem-go 是一个受 Chroma 启发的纯 Go 实现的向量数据库...

================================================================================
元数据过滤示例
--------------------------------------------------------------------------------

找到 1 个 category='features' 的文档:
  [1] chromem-go 的主要特性包括：零外部依赖、线程安全操作...
      分类: features

================================================================================
持久化存储验证
--------------------------------------------------------------------------------

重新打开存储 - 文档已持久化: 5 个文档
数据成功跨存储实例持久化！

=== 示例成功完成！ ===
```

## 高级用法

### 自定义嵌入函数

```go
// 使用 OpenAI 嵌入（如本例所示）
llm, _ := openai.New()
openaiEmbedder, _ := embeddings.NewEmbedder(llm)
embedder := rag.NewLangChainEmbedder(openaiEmbedder)

// 或使用其他嵌入提供商：
// - Cohere: embeddings.NewEmbedder(cohere.New())
// - Jina: embeddings.NewEmbedder(jina.New())
// - Ollama: embeddings.NewEmbedder(ollama.New())
// 等等
```

### 批量操作

```go
// 高效添加大量文档
documents := make([]rag.Document, 1000)
// ... 填充文档
err = vectorStore.Add(ctx, documents)  // 自动并行处理
```

### 统计和监控

```go
stats, err := vectorStore.GetStats(ctx)
fmt.Printf("文档数: %d\n", stats.TotalDocuments)
fmt.Printf("向量数: %d\n", stats.TotalVectors)
fmt.Printf("维度: %d\n", stats.Dimension)
```

## 与其他向量存储的比较

| 特性 | chromem-go | Chroma | Pinecone | Weaviate |
|------|------------|--------|----------|----------|
| 外部服务 | 否 | 是 (Docker) | 是 | 是 |
| 依赖 | 零 | Docker | 无 | Docker |
| 持久化存储 | SQLite | SQLite/ClickHouse | 云 | 云 |
| 语言 | 纯 Go | Python | SDK | Go/Python |
| 成本 | 免费 | 免费 | 付费 | 免费/付费 |
| 最适合 | Go 应用 | Python 应用 | 生产环境 | 企业级 |

## chromem-go 的优势

1. **无基础设施开销** - 任何能运行 Go 的地方都能运行
2. **简单部署** - 单个二进制文件，无需 Docker
3. **快速启动** - 嵌入式数据库，无网络延迟
4. **类型安全** - 原生 Go 类型和接口
5. **易于测试** - 内存模式用于测试
6. **低资源使用** - 最小的内存和 CPU 占用

## 故障排除

### 创建存储目录时权限被拒绝

**错误**: `failed to create persistence directory: permission denied`

**解决方案**: 确保应用对存储目录有写权限，或使用不同的位置：
```go
store, err := store.NewChromemVectorStoreSimple(os.TempDir(), embedder)
```

### 嵌入维度不匹配

**错误**: `embedding dimension mismatch`

**解决方案**: 确保所有嵌入具有相同的维度：
```go
embedder := store.NewMockEmbedder(128)  // 固定维度
```

## 下一步

- 尝试不同的嵌入函数
- 使用元数据过滤进行高级搜索场景
- 实现结合向量和关键词的混合搜索
- 添加文档更新和删除操作
- 探索并发批量操作

## 参考

- [chromem-go GitHub](https://github.com/philippgille/chromem-go)
- [LangGraphGo 文档](../../docs/RAG/RAG.md)
- [RAG 架构](../../docs/RAG/Architecture.md)
- [向量存储接口](../../rag/types.go)
