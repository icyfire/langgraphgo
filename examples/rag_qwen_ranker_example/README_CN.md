# Qwen3-Embedding-4B 重排器示例

本示例演示如何使用 Qwen3-Embedding-4B 作为嵌入模型和重排器（reranker）构建 RAG（检索增强生成）管道。

## 什么是 Qwen3-Embedding-4B？

Qwen3-Embedding-4B 是阿里发布的最新嵌入模型，具有以下特点：

- **高质量嵌入**：40 亿参数，提供丰富的语义表示
- **多语言支持**：在中英文文本上表现优异
- **双重能力**：可生成嵌入向量 AND 执行重排序
- **4096 维度**：高容量，能捕捉细微的语义差异

## 什么是重排序（Reranking）？

重排序是一种两阶段检索技术：

1. **阶段一 - 快速检索**：使用向量相似度搜索获取大量候选文档（如 50-100 篇）
2. **阶段二 - 精确重排**：使用交叉编码器模型对候选文档重新打分，只返回最相关的前 K 个（如前 5-10 个）

这种方案结合了：
- 向量搜索的**速度**（双编码器模型）
- 重排序器的**准确性**（交叉编码器模型）

## 功能演示

1. **Qwen3-Embedding-4B 生成嵌入** - 为文档生成向量表示
2. **向量相似度搜索** - 使用余弦相似度进行快速初步检索
3. **基于 LLM 的重排序** - 使用 Qwen 的重排序能力重新评分
4. **两阶段检索** - 先获取大量候选，再重排找出最佳
5. **组合检索器** - 在一个管道中结合向量搜索和重排序

## 前置要求

### 1. 安装依赖

```bash
cd examples/rag_qwen_ranker_example
go mod tidy
```

### 2. 配置 API 访问

本示例支持多种嵌入后端：

#### 选项 1：ModelScope（推荐用于 Qwen3-Embedding-4B）

```bash
# ModelScope API 端点
export EMBEDDING_BASE_URL=https://api-inference.modelscope.cn/v1

# 你的 ModelScope API Key
export MODELSCOPE_API_KEY=your-modelscope-api-key

# 嵌入模型
export OPENAI_EMBEDDING_MODEL=Qwen/Qwen3-Embedding-4B
```

获取 ModelScope API Key：
1. 访问 [ModelScope](https://www.modelscope.cn/)
2. 注册或登录账号
3. 进入账户设置获取 API 密钥

#### 选项 2：灵积 DashScope（阿里云）

```bash
# 灵积 API 端点
export EMBEDDING_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

# 你的灵积 API Key
export OPENAI_API_KEY=your-dashscope-api-key

# 嵌入模型
export OPENAI_EMBEDDING_MODEL=text-embedding-v3
```

获取灵积 API Key：
1. 访问 [灵积控制台](https://dashscope.console.aliyun.com/)
2. 使用阿里云账号登录或注册
3. 进入 API Key 管理页面
4. 创建新的 API Key

#### 选项 3：OpenAI

```bash
# OpenAI API 端点（默认）
export EMBEDDING_BASE_URL=https://api.openai.com/v1

# 你的 OpenAI API Key
export OPENAI_API_KEY=your-openai-api-key

# 嵌入模型
export OPENAI_EMBEDDING_MODEL=text-embedding-3-small
```

## 运行示例

```bash
go run main.go
```

## 示例输出

```
=== RAG with Qwen3-Embedding-4B Reranker Example ===

初始化 Qwen3-Embedding-4B 向量存储...
添加文档到向量存储...
成功添加 6 个文档

创建了组合检索器：向量搜索 + LLM 重排序

构建 RAG 管道...
管道图：
          +--------+
          | __start__ |
          +--------+
            |
            v
       +----------+
       | retrieve |
       +----------+
         |
         v
     +---------+
     | generate |
     +---------+
       |
       v
   +--------+
   | __end__ |
   +--------+

================================================================================
查询 1: 什么是 Qwen3-Embedding-4B？

回答：
Qwen3-Embedding-4B 是阿里云 Qwen 团队发布的最新嵌入模型，拥有 40 亿参数...
它为多语言文本提供高质量的向量表示...

检索到 3 个文档：
  [1] 评分: 0.9234 - Qwen3-Embedding-4B 是最先进的嵌入模型...
  [2] 评分: 0.8876 - Qwen3-Embedding-4B 模型支持嵌入生成和重排序...
  [3] 评分: 0.8234 - 重排序是一种两阶段技术...

================================================================================
重排序演示

查询：Qwen 嵌入模型有哪些特性？

1. 向量搜索结果（无重排序）：
   [1] 评分: 0.8756 - Qwen3-Embedding-4B 是最先进的嵌入模型...
   [2] 评分: 0.8432 - Qwen3-Embedding-4B 模型支持嵌入生成和重排序...
   [3] 评分: 0.7891 - Milvus 和 chromem-go 等向量数据库存储嵌入...
   [4] 评分: 0.7654 - 重排序是一种两阶段技术...
   [5] 评分: 0.7432 - LangGraphGo 提供灵活的 RAG 管道...

2. 使用 Qwen3-Embedding-4B 重排序后：
   [1] 评分: 0.9456 - Qwen3-Embedding-4B 模型支持嵌入生成和重排序...
   [2] 评分: 0.9123 - Qwen3-Embedding-4B 是最先进的嵌入模型...
   [3] 评分: 0.8765 - Qwen3-Embedding-4B 模型使用 4096 维向量...
```

## 架构

```
用户查询
    |
    v
+-------------------+
| 向量搜索          |  <-- 快速检索大量候选
| (双编码器)        |      使用嵌入向量的余弦相似度
+-------------------+
    |
    | 返回前 10-50 个候选
    v
+-------------------+
| 重排序            |  <-- 精确重新评分
| (交叉编码器)      |      使用 Qwen3-Embedding-4B 重排器
+-------------------+
    |
    | 返回前 3-5 个最相关的
    v
+-------------------+
| LLM 生成回答      |  <-- 生成最终答案
| (Qwen 聊天模型)   |      使用检索到的上下文
+-------------------+
    |
    v
最终答案
```

## 配置说明

### 嵌入模型

```go
// 使用 Qwen3-Embedding-4B
embeddingModel := "text-embedding-v3"

llmForEmbeddings, err := openai.New(
    openai.WithEmbeddingModel(embeddingModel),
)
```

### 重排器

```go
// 创建基于 LLM 的重排器
rerankerConfig := retriever.DefaultLLMRerankerConfig()
rerankerConfig.TopK = 3 // 返回前 3 个结果
rerankerConfig.SystemPrompt = "自定义评分提示..."

reranker := retriever.NewLLMReranker(llm, rerankerConfig)
```

### 向量检索器

```go
// 初始检索更多候选用于重排序
vectorRetriever := retriever.NewVectorStoreRetriever(
    vectorStore,
    embedder,
    10, // 检索 10 个候选用于重排序
)
```

### 自定义重排检索器

示例包含一个自定义的 `RerankingRetriever`，结合了向量搜索和 LLM 重排序：

```go
import "github.com/smallnest/langgraphgo/llms/qwen"

// 创建支持 encoding_format 的 Qwen 嵌入器
embedder := qwen.NewEmbedder(
    "https://api-inference.modelscope.cn/v1",
    apiKey,
    "Qwen/Qwen3-Embedding-4B",
)

type RerankingRetriever struct {
    vectorStore rag.VectorStore
    embedder    rag.Embedder
    reranker    rag.Reranker
    fetchK      int // 获取候选数量用于重排序
}

func (r *RerankingRetriever) RetrieveWithConfig(ctx context.Context, query string, config *rag.RetrievalConfig) ([]rag.DocumentSearchResult, error) {
    // 步骤 1：使用向量搜索获取更多候选
    queryEmbedding, _ := r.embedder.EmbedDocument(ctx, query)
    candidates, _ := r.vectorStore.Search(ctx, queryEmbedding, r.fetchK)

    // 步骤 2：对候选进行重排序
    reranked, _ := r.reranker.Rerank(ctx, query, candidates)

    return reranked, nil
}
```

**注意**：Qwen 嵌入器现在作为可重用包提供，位于 `github.com/smallnest/langgraphgo/llms/qwen`。你可以在自己的项目中使用它。

## 性能考虑

### 何时使用重排序

**使用重排序：**
- 准确性比延迟更重要
- 文档语料库较大（> 1 万篇文档）
- 查询复杂，需要深度理解
- 需要最相关的结果

**跳过重排序：**
- 延迟很关键
- 文档语料库较小（< 1000 篇文档）
- 查询简单，只需关键词匹配
- 向量搜索已经提供良好结果

### 检索策略

| 阶段 | 文档数 | 模型 | 延迟 | 准确性 |
|------|--------|------|------|--------|
| 向量搜索 | 10-50 | 双编码器 | ~10ms | 良好 |
| 重排序 | 3-10 | 交叉编码器 | ~100ms | 优秀 |

### 成本优化

1. **调整检索数量**：如果重排序开销大，减少候选数量
2. **缓存嵌入**：预计算并缓存文档嵌入
3. **批量请求**：并行处理多个查询
4. **使用更小模型**：初步检索可考虑更小的嵌入模型

## 高级用法

### 自定义重排器

```go
type CustomReranker struct {
    llm *openai.LLM
}

func (r *CustomReranker) Retrieve(ctx context.Context, query string, k int) ([]rag.RAGDocument, error) {
    // 初步检索
    candidates, _ := vectorRetriever.Retrieve(ctx, query, k*5)

    // 使用自定义逻辑重排
    for _, doc := range candidates {
        // 自定义重排序逻辑
        doc.Score = calculateRelevance(query, doc)
    }

    // 排序并返回前 k 个
    return topK(candidates, k), nil
}
```

### 混合搜索

```go
// 结合稠密和稀疏检索
hybridRetriever := retriever.NewHybridRetriever(
    retriever.NewVectorStoreRetriever(vectorStore, embedder, 10),
    retriever.NewBM25Retriever(documentStore, 10),
    0.7, // 向量搜索权重
    0.3, // BM25 权重
)
```

### 元数据过滤

```go
// 重排序前按元数据过滤
filteredRetriever := retriever.NewFilteredRetriever(
    vectorRetriever,
    map[string]any{
        "category": "technical",
        "priority": 1,
    },
)
```

## 故障排查

### API 错误

**错误**：`API returned unexpected status code: 401`

**解决方案**：检查 API Key 是否正确：
```bash
echo $OPENAI_API_KEY
```

**错误**：`API name not exist`

**解决方案**：验证嵌入模型名称：
```bash
export OPENAI_EMBEDDING_MODEL=text-embedding-v3
```

**错误**：`encoding_format must be 'float' or 'base64'`

**解决方案**：本示例中的自定义 `QwenEmbedder` 已自动处理此问题，通过设置 `encoding_format: "float"`。如果你实现自己的嵌入器，请确保使用这些有效值之一。

**错误**：`API returned status 429: We have to rate limit you`

**解决方案**：ModelScope 的免费 API 层有速率限制。你可以：
1. 在请求之间等待几秒
2. 使用 DashScope（阿里云商业 API）获得更高限额
3. 使用其他嵌入提供商（如 OpenAI）

### 重排序不工作

**错误**：`Failed to create LLM reranker`

**解决方案**：示例会退回到仅向量检索。检查：
1. LLM 配置正确
2. 模型支持重排序能力
3. API 凭证有效

### 检索质量差

**可能原因**：
1. 文档太短或缺少上下文
2. 查询模糊
3. 嵌入模型不适合该领域

**解决方案**：
1. 将文档分块为更小、更聚焦的片段
2. 添加查询扩展或重写
3. 使用领域特定的嵌入模型

## 与其他方法对比

| 方法 | 延迟 | 准确性 | 复杂度 |
|------|------|--------|--------|
| 仅向量搜索 | 低 | 良好 | 低 |
| 向量 + 重排 | 中 | 优秀 | 中 |
| 纯 LLM 检索 | 高 | 优秀 | 低 |
| 混合（稠密 + 稀疏） | 中 | 优秀 | 高 |

## 参考资源

- [Qwen 文档](https://qwen.readthedocs.io/)
- [灵积 API 文档](https://help.aliyun.com/zh/dashscope/)
- [LangGraphGo RAG 文档](../../rag/README.md)
- [检索器实现](../../rag/retriever/)
- [向量存储选项](../../rag/store/)
- [llms/qwen 包](../../llms/qwen/) - 可重用的 Qwen 嵌入器包

## 许可证

本示例遵循 LangGraphGo 项目许可证。
