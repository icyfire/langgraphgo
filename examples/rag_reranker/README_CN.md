# RAG 重排序器对比示例

本示例展示了 LangGraphGo RAG 系统中可用的不同重排序策略。重排序是提高检索质量的关键步骤，通过根据文档与查询的相关性重新评分来优化结果。

## 概述

本示例对比了四种不同的重排序器实现：

| 重排序器 | 描述 | 优点 | 缺点 |
|----------|-------------|------|------|
| **SimpleReranker** | 基于关键词的重排序 | 快速，无需 API 调用，始终可用 | 准确度有限，仅支持关键词 |
| **LLMReranker** | 使用 LLM 对文档评分 | 良好的语义理解，无需新依赖 | 较慢，API 成本较高 |
| **CohereReranker** | 使用 Cohere 的 Rerank API | 高质量结果，快速 | 需要 API 费用和密钥 |
| **JinaReranker** | 使用 Jina AI 的 Rerank API | 高质量，支持多语言 | 需要 API 费用和密钥 |

## 前置要求

```bash
# 必需
OPENAI_API_KEY=your_key_here

# 可选（用于 Cohere 重排序器）
COHERE_API_KEY=your_key_here

# 可选（用于 Jina 重排序器）
JINA_API_KEY=your_key_here
```

## 运行示例

### 基础使用（基于 LLM 的重排序）

```bash
OPENAI_API_KEY=sk-xxx go run main.go
```

这将使用 SimpleReranker 和 LLMReranker 运行示例。

### 使用 Cohere 重排序器

```bash
COHERE_API_KEY=your_key OPENAI_API_KEY=sk-xxx go run main.go
```

### 使用 Jina 重排序器

```bash
JINA_API_KEY=your_key OPENAI_API_KEY=sk-xxx go run main.go
```

### 使用所有重排序器

```bash
COHERE_API_KEY=your_key JINA_API_KEY=your_key OPENAI_API_KEY=sk-xxx go run main.go
```

## 输出

示例将显示：

1. **检索到的文档**：被检索和重排序的文档
2. **相关性评分**：每个重排序器分配的相关性评分
3. **生成的答案**：基于重排序文档的最终答案

示例输出：
```
--- LLMReranker ---
检索到的文档：
  [1] langgraph_intro.txt (主题: LangGraph)
      LangGraph is a library for building stateful, multi-actor applications...
  [2] multi_agent.txt (主题: Multi-Agent)
      Multi-agent systems in LangGraph enable multiple AI agents to work...

相关性评分：
  [1] 评分: 0.8234 (方法: llm)
  [2] 评分: 0.7891 (方法: llm)

答案: LangGraph is a library designed for building stateful, multi-actor applications...
```

## 代码结构

```go
// 创建基础检索器
baseRetriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, 5)

// 创建重排序器
llmReranker := retriever.NewLLMReranker(llm, retriever.DefaultLLMRerankerConfig())

// 配置带重排序的 RAG 管道
config := rag.DefaultPipelineConfig()
config.Retriever = baseRetriever
config.Reranker = llmReranker
config.UseReranking = true

// 构建并运行管道
pipeline := rag.NewRAGPipeline(config)
pipeline.BuildAdvancedRAG()
runnable, _ := pipeline.Compile()
result, _ := runnable.Invoke(ctx, map[string]any{"query": query})
```

## 自定义配置

### 调整 TopK（返回结果数量）

```go
config := retriever.DefaultLLMRerankerConfig()
config.TopK = 10  // 返回前 10 个文档而不是 5 个

llmReranker := retriever.NewLLMReranker(llm, config)
```

### 设置评分阈值

```go
config := retriever.DefaultCohereRerankerConfig()
config.TopK = 5

cohereReranker := retriever.NewCohereReranker(apiKey, config)

// 在管道中，重排序后按阈值过滤
```

### 自定义系统提示词（LLMReranker）

```go
config := retriever.LLMRerankerConfig{
    TopK: 5,
    ScoreThreshold: 0.5,
    SystemPrompt: "你是一个专业的技术文档评分员。请根据技术准确性和完整性进行评分。",
    BatchSize: 5,
}

llmReranker := retriever.NewLLMReranker(llm, config)
```

### 使用不同的 Cohere 模型

```go
config := retriever.CohereRerankerConfig{
    Model: "rerank-english-v3.0",  // 英语优化
    // Model: "rerank-multilingual-v3.0",  // 多语言
    TopK: 5,
}

cohereReranker := retriever.NewCohereReranker(apiKey, config)
```

### 使用不同的 Jina 模型

```go
config := retriever.JinaRerankerConfig{
    Model: "jina-reranker-v1-base-en",  // 仅英语
    // Model: "jina-reranker-v2-base-multilingual",  // 多语言
    TopK: 5,
}

jinaReranker := retriever.NewJinaReranker(apiKey, config)
```

## Cross-Encoder 重排序

对于本地、隐私保护的重排序（无需 API 调用），你可以使用 CrossEncoderReranker 配合本地服务：

### 设置 Python 服务

```bash
# 安装依赖
pip install sentence-transformers flask flask-cors

# 启动服务
python ../../scripts/cross_encoder_server.py --port 8000
```

### 在 Go 代码中使用

```go
config := retriever.CrossEncoderRerankerConfig{
    APIBase:   "http://localhost:8000/rerank",
    ModelName: "cross-encoder/ms-marco-MiniLM-L-6-v2",
    TopK:      5,
}

ceReranker := retriever.NewCrossEncoderReranker(config)

// 在管道中使用
config := rag.DefaultPipelineConfig()
config.Reranker = ceReranker
```

## 性能对比

| 重排序器 | 速度 | 质量 | 成本 | 隐私 |
|----------|-------|---------|------|---------|
| SimpleReranker | ⚡⚡⚡ 最快 | ⭐⭐ 基础 | 免费 | 本地 |
| LLMReranker | ⚡⚡ 中等 | ⭐⭐⭐⭐ 良好 | 按token计费 | 取决于 LLM |
| CohereReranker | ⚡⚡⚡ 快速 | ⭐⭐⭐⭐⭐ 优秀 | 按请求计费 | 云 API |
| JinaReranker | ⚡⚡⚡ 快速 | ⭐⭐⭐⭐⭐ 优秀 | 按请求计费 | 云 API |
| CrossEncoder | ⚡⚡ 中等 | ⭐⭐⭐⭐ 很好 | 免费（本地） | 本地 |

## 何时使用哪种重排序器

- **SimpleReranker**：快速原型开发，无需外部依赖
- **LLMReranker**：已有 LLM，想要语义理解而无需新服务
- **CohereReranker**：生产环境，英语为主的应用
- **JinaReranker**：生产环境，多语言应用
- **CrossEncoder**：隐私敏感的应用，成本敏感的部署

## 相关文档

- [../../rag/RERANKER.md](../../rag/RERANKER.md) - 详细实现方案
- [../../rag/](../../rag/) - RAG 包文档
- [../rag_advanced/](../rag_advanced/) - 带重排序的高级 RAG 示例
