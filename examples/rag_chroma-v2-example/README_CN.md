# RAG 与 Chroma v2 API 向量数据库示例

本示例演示如何将 **Chroma v2 API 向量数据库** 与 **LangGraphGo 的 RAG 管道**结合使用，通过原生的 `ChromaV2VectorStore` 实现。

## 什么是 Chroma v2？

Chroma v2 API 引入了相对于 v1 的重大改进：

- **层次化结构** - 将资源组织为 tenant → database → collection 层次结构
- **RESTful API 设计** - 遵循 OpenAPI 3.1.0 规范
- **改进的身份验证** - 更好的身份验证和令牌管理
- **细粒度权限** - 精细的访问控制
- **更好的性能** - 优化的查询和索引操作
- **多租户支持** - 内置支持多租户和数据库

## 前置要求

1. **Go 1.21+**: 构建示例所需
2. **Chroma v2 服务器**: 需要在本地或远程运行
   ```bash
   # 使用 Docker 启动 Chroma v2（最新版本）
   docker run -p 8000:8000 chromadb/chroma

   # 或指定版本
   docker run -p 8000:8000 chromadb/chroma:0.6.0

   # 或本地安装并运行兼容版本
   pip install chromadb==0.4.22
   chroma run --host localhost --port 8000
   ```
3. **OpenAI API Key**: 同时用于 LLM 调用和嵌入向量生成
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```

**注意**: 本示例使用真实的 OpenAI 嵌入来生成文档的向量表示。

## 运行示例

```bash
cd examples/chroma-v2-example
go run main.go
```

**可选**: 设置自定义 Chroma URL
```bash
export CHROMA_URL="http://localhost:8000"
go run main.go
```

## 功能演示

- 使用原生实现创建 Chroma v2 向量存储
- 理解租户/数据库/集合层次结构
- 添加文档并自动生成嵌入向量
- 使用 Chroma v2 作为向量存储构建 RAG 管道
- 使用自然语言问题查询管道
- 通过 Chroma v2 API 进行直接相似度搜索
- 元数据过滤实现定向搜索
- 可视化 RAG 管道图

## 代码亮点

### 创建 OpenAI 嵌入器

```go
import (
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/smallnest/langgraphgo/rag"
)

// 创建 OpenAI LLM
llm, err := openai.New()
if err != nil {
    log.Fatal(err)
}

// 创建 OpenAI 嵌入器
openaiEmbedder, err := embeddings.NewEmbedder(llm)
if err != nil {
    log.Fatal(err)
}

// 使用 LangGraphGo 适配器包装
embedder := rag.NewLangChainEmbedder(openaiEmbedder)
```

### 创建 Chroma v2 存储

```go
import "github.com/smallnest/langgraphgo/rag/store"

// 简单初始化
store, err := store.NewChromaV2VectorStoreSimple(
    "http://localhost:8000",  // Chroma 服务器 URL
    "my_collection",         // 集合名称
    embedder,                // 嵌入器
)

// 或使用完整配置
store, err := store.NewChromaV2VectorStore(store.ChromaV2Config{
    BaseURL:    "http://localhost:8000",
    Tenant:     "default_tenant",
    Database:   "default_database",
    Collection: "my_collection",
    Embedder:   embedder,
})
```

### 添加文档

```go
documents := []rag.Document{
    {
        ID:      "doc1",
        Content: "你的文档内容",
        Metadata: map[string]any{
            "category": "tech",
            "source":   "docs",
        },
    },
}

err = store.Add(ctx, documents)
```

### 相似度搜索

```go
// 生成查询嵌入向量
queryEmbedding, _ := embedder.EmbedDocument(ctx, "搜索查询")

// 基本相似度搜索
results, err := store.Search(ctx, queryEmbedding, k)

// 带元数据过滤的搜索
results, err := store.SearchWithFilter(ctx, queryEmbedding, k,
    map[string]any{"category": "tech"},
)

for _, result := range results {
    fmt.Printf("分数: %.4f - %s\n", result.Score, result.Document.Content)
}
```

### 与 RAG 管道一起使用

```go
import (
    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/retriever"
)

// 创建检索器
retriever := retriever.NewVectorStoreRetriever(store, embedder, topK)

// 配置 RAG 管道
config := rag.DefaultPipelineConfig()
config.Retriever = retriever
config.LLM = llm

// 构建和编译
pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()
runnable, _ := pipeline.Compile()

// 查询
result, _ := runnable.Invoke(ctx, map[string]any{"query": "什么是 Chroma v2？"})
```

## 配置选项

### ChromaV2Config

| 选项 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| `BaseURL` | string | Chroma 服务器的 URL | 必需 |
| `Tenant` | string | 租户名称 | "default_tenant" |
| `Database` | string | 数据库名称 | "default_database" |
| `Collection` | string | 集合名称 | 必需 |
| `CollectionID` | string | 集合 UUID（可选） | 自动生成 |
| `Embedder` | Embedder | 嵌入函数 | 必需 |
| `HTTPClient` | *http.Client | HTTP 客户端 | 默认 30 秒超时 |

## Chroma v2 API 结构

### 层次结构

```
Tenant（租户，例如 "default_tenant"）
  └── Database（数据库，例如 "default_database"）
      └── Collection（集合，例如 "my_collection"）
          └── Records（记录，即带嵌入向量的文档）
```

### 主要端点

- `POST /api/v2/tenants/{tenant}/databases/{database}/collections` - 创建集合
- `POST /api/v2/tenants/{tenant}/databases/{database}/collections/{id}/add` - 添加记录
- `POST /api/v2/tenants/{tenant}/databases/{database}/collections/{id}/search` - 搜索
- `POST /api/v2/tenants/{tenant}/databases/{database}/collections/{id}/delete` - 删除记录
- `POST /api/v2/tenants/{tenant}/databases/{database}/collections/{id}/upsert` - 更新记录

## 预期输出

```
=== RAG 与 Chroma v2 API 向量数据库示例 ===

正在连接到 Chroma v2: http://localhost:8000

创建的存储集合: langgraphgo_example (ID: 12345678-1234-1234-1234-123456789abc)

正在添加文档到 Chroma v2...
成功添加 5 个文档

存储统计信息:
  总文档数: 5
  向量维度: 1536

正在构建 RAG 管道...

管道图:
┌───────────────────────────────────────────────┐
│                   RAG Pipeline                 │
└───────────────────────────────────────────────┘

================================================================================
查询 1: Chroma v2 有什么新特性？
--------------------------------------------------------------------------------

直接 Chroma v2 相似度搜索结果:
  [1] 分数: 0.8542
      Chroma v2 API 引入了新的层次化结构...
      元数据: map[category:architecture source:chroma_v2_docs]
  [2] 分数: 0.7821
      LangGraphGo 通过 ChromaV2VectorStore 提供原生 Chroma v2 支持...
      元数据: map[category:integration source:langgraphgo_docs]

RAG 回答:
Chroma v2 引入了层次化结构...

================================================================================
元数据过滤示例（Chroma v2 原生）
--------------------------------------------------------------------------------

找到 2 个 category='architecture' 的文档:
  [1] 分数: 0.8542
      Chroma v2 API 引入了新的层次化结构...
      分类: architecture
  [2] 分数: 0.7234
      Chroma v2 将概念分离为租户、数据库和集合...
      分类: architecture

=== 示例成功完成！ ===

注意: 此示例使用 Chroma v2 API。
Chroma v2 是最新 Chroma 版本的默认设置。
启动 Chroma: docker run -p 8000:8000 chromadb/chroma
```

## 故障排除

### 连接被拒绝

**错误**: `Failed to create Chroma v2 store: connection refused`

**解决方案**: 确保 Chroma 服务器正在运行：
```bash
# 检查 Chroma 是否正在运行
curl http://localhost:8000/api/v2/heartbeat

# 如果未运行则启动 Chroma
docker run -p 8000:8000 chromadb/chroma
```

### 找不到 OpenAI API 密钥

**错误**: `Failed to create embedder: OPENAI_API_KEY not found`

**解决方案**: 设置您的 OpenAI API 密钥：
```bash
export OPENAI_API_KEY="your-api-key"
```

### 集合未找到

**错误**: `Failed to search: status 404`

**解决方案**: 集合将在首次使用时自动创建。如果看到此错误，请检查：
1. 集合名称正确
2. 租户和数据库存在
3. 您有适当的权限

### 元数据过滤返回无结果

**问题**: 即使存在具有匹配元数据的文档，元数据过滤也可能返回 0 个结果。

**说明**: 某些 Chroma v2 服务器版本可能不完全支持元数据存储和过滤。`ChromaV2VectorStore` 实现正确地在 API 负载中发送了元数据，但服务器可能无论如何都会将其存储为 `null`。

**状态**: 这是某些 Chroma v2 服务器版本的已知限制。基本的向量搜索功能正常工作 - 只有元数据过滤可能受到影响。

**解决方法**:
- 使用完全支持元数据的 Chroma v2 服务器版本
- 将过滤条件作为文档内容的一部分存储
- 使用多个集合来区分不同类别，而不是使用元数据过滤

## Chroma v2 与 v1 对比

| 特性 | v1 API | v2 API |
|------|-------|-------|
| URL 结构 | `/api/v1/...` | `/api/v2/tenants/{t}/databases/{d}/...` |
| 组织方式 | 平面（仅集合） | 层次化（租户/数据库/集合） |
| 身份验证 | 基本令牌 | 增强的身份验证和身份 |
| 规范 | 自定义 | OpenAPI 3.1.0 |
| 多租户 | 有限 | 完全支持 |
| 距离函数 | 配置选项 | 集合配置的一部分 |

## Chroma v2 的优势

1. **更好的组织** - 用于资源管理的层次化结构
2. **改进的安全性** - 增强的身份验证和授权
3. **标准合规** - OpenAPI 规范便于集成
4. **性能** - 优化的查询执行
5. **可扩展性** - 更好地支持大规模部署
6. **开发体验** - 清晰的 API 结构和文档

## 下一步

- 尝试不同的租户/数据库配置
- 使用元数据过滤进行高级搜索场景
- 实现带重试逻辑的自定义 HTTP 客户端
- 使用 Chroma Cloud 进行生产部署
- 探索完整的 Chroma v2 API 规范

## 参考

- [Chroma 文档](https://docs.trychroma.com/)
- [Chroma v2 API 规范](http://localhost:8000/openapi.json)（服务器运行时）
- [LangGraphGo 文档](../../docs/RAG/RAG.md)
- [RAG 架构](../../docs/RAG/Architecture.md)
- [向量存储接口](../../rag/types.go)
