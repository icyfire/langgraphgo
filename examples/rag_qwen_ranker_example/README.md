# RAG with Qwen3-Embedding-4B Reranker Example

This example demonstrates how to build a RAG (Retrieval-Augmented Generation) pipeline using Qwen3-Embedding-4B as both an embedder and reranker with LangGraphGo.

## What is Qwen3-Embedding-4B?

Qwen3-Embedding-4B is Alibaba's state-of-the-art embedding model that provides:

- **High-quality embeddings**: 4 billion parameters for rich semantic representations
- **Multilingual support**: Excellent performance on Chinese and English text
- **Dual capabilities**: Can generate embeddings AND perform reranking
- **4096 dimensions**: High capacity for capturing nuanced meanings

## What is Reranking?

Reranking is a two-stage retrieval technique:

1. **Stage 1 - Fast Retrieval**: Use vector similarity search to fetch a large number of candidates (e.g., 50-100 documents)
2. **Stage 2 - Accurate Reranking**: Use a cross-encoder model to re-score the candidates and return only the top-K most relevant (e.g., top 5-10)

This combines:
- **Speed** of vector search (bi-encoder models)
- **Accuracy** of cross-encoder rerankers

## Features Demonstrated

1. **Qwen3-Embedding-4B for embeddings** - Generate vector representations of documents
2. **Vector similarity search** - Fast initial retrieval using cosine similarity
3. **LLM-based reranking** - Re-score documents using Qwen's reranking capability
4. **Two-stage retrieval** - Fetch many candidates, rerank to find the best
5. **Composite retriever** - Combine vector search and reranking in one pipeline

## Prerequisites

### 1. Install Dependencies

```bash
cd examples/rag_qwen_ranker_example
go mod tidy
```

### 2. Configure API Access

You can use different embedding backends with this example:

#### Option 1: ModelScope (Recommended for Qwen3-Embedding-4B)

```bash
# ModelScope API endpoint
export EMBEDDING_BASE_URL=https://api-inference.modelscope.cn/v1

# Your ModelScope API key
export MODELSCOPE_API_KEY=your-modelscope-api-key

# Embedding model
export OPENAI_EMBEDDING_MODEL=Qwen/Qwen3-Embedding-4B
```

Get ModelScope API key:
1. Visit [ModelScope](https://www.modelscope.cn/)
2. Sign up or log in
3. Navigate to your account settings for API keys

#### Option 2: DashScope (Alibaba Cloud)

```bash
# DashScope API endpoint
export EMBEDDING_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

# Your DashScope API key
export OPENAI_API_KEY=your-dashscope-api-key

# Embedding model
export OPENAI_EMBEDDING_MODEL=text-embedding-v3
```

Get DashScope API key:
1. Visit [DashScope Console](https://dashscope.console.aliyun.com/)
2. Sign up or log in with your Alibaba account
3. Navigate to API Key Management
4. Create a new API key

#### Option 3: OpenAI

```bash
# OpenAI API endpoint (default)
export EMBEDDING_BASE_URL=https://api.openai.com/v1

# Your OpenAI API key
export OPENAI_API_KEY=your-openai-api-key

# Embedding model
export OPENAI_EMBEDDING_MODEL=text-embedding-3-small
```

## Running the Example

```bash
go run main.go
```

## Example Output

```
=== RAG with Qwen3-Embedding-4B Reranker Example ===

Initializing vector store with Qwen3-Embedding-4B...
Adding documents to vector store...
Successfully added 6 documents

Created composite retriever with vector search + LLM reranking

Building RAG pipeline...
Pipeline Graph:
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
Query 1: What is Qwen3-Embedding-4B?

Answer:
Qwen3-Embedding-4B is a state-of-the-art embedding model released by Alibaba's
Qwen team with 4 billion parameters. It provides high-quality vector
representations for text in multiple languages...

Retrieved 3 documents:
  [1] Score: 0.9234 - Qwen3-Embedding-4B is a state-of-the-art embedding model...
  [2] Score: 0.8876 - The Qwen3-Embedding-4B model supports both embedding generation...
  [3] Score: 0.8234 - Reranking is a two-stage technique where...

================================================================================
Reranking Demonstration

Query: What are the features of Qwen embedding models?

1. Vector Search Results (without reranking):
   [1] Score: 0.8756 - Qwen3-Embedding-4B is a state-of-the-art embedding model...
   [2] Score: 0.8432 - The Qwen3-Embedding-4B model supports both embedding generation...
   [3] Score: 0.7891 - Vector databases like Milvus and chromem-go store embeddings...
   [4] Score: 0.7654 - Reranking is a two-stage technique...
   [5] Score: 0.7432 - LangGraphGo provides a flexible RAG pipeline...

2. After Reranking with Qwen3-Embedding-4B:
   [1] Score: 0.9456 - The Qwen3-Embedding-4B model supports both embedding generation...
   [2] Score: 0.9123 - Qwen3-Embedding-4B is a state-of-the-art embedding model...
   [3] Score: 0.8765 - The Qwen3-Embedding-4B model uses 4096-dimensional vectors...
```

## Architecture

```
User Query
    |
    v
+-------------------+
| Vector Search     |  <-- Fast retrieval of many candidates
| (bi-encoder)      |      Uses cosine similarity on embeddings
+-------------------+
    |
    | Returns top 10-50 candidates
    v
+-------------------+
| Reranking         |  <-- Accurate re-scoring
| (cross-encoder)   |      Uses Qwen3-Embedding-4B reranker
+-------------------+
    |
    | Returns top 3-5 most relevant
    v
+-------------------+
| LLM Generation    |  <-- Generate final answer
| (Qwen Chat)       |      Uses retrieved context
+-------------------+
    |
    v
Final Answer
```

## Configuration

### Embedding Model

```go
// Use Qwen3-Embedding-4B
embeddingModel := "text-embedding-v3"

llmForEmbeddings, err := openai.New(
    openai.WithEmbeddingModel(embeddingModel),
)
```

### Reranker

```go
// Create LLM-based reranker
rerankerConfig := retriever.DefaultLLMRerankerConfig()
rerankerConfig.TopK = 3 // Return top 3 results
rerankerConfig.SystemPrompt = "Custom prompt for scoring..."

reranker := retriever.NewLLMReranker(llm, rerankerConfig)
```

### Vector Retriever

```go
// Fetch more candidates initially for reranking
vectorRetriever := retriever.NewVectorStoreRetriever(
    vectorStore,
    embedder,
    10, // Fetch 10 for reranking
)
```

### Custom Reranking Retriever

The example includes a custom `RerankingRetriever` that combines vector search with LLM reranking:

```go
import "github.com/smallnest/langgraphgo/llms/qwen"

// Create Qwen embedder with encoding_format support
embedder := qwen.NewEmbedder(
    "https://api-inference.modelscope.cn/v1",
    apiKey,
    "Qwen/Qwen3-Embedding-4B",
)

type RerankingRetriever struct {
    vectorStore rag.VectorStore
    embedder    rag.Embedder
    reranker    rag.Reranker
    fetchK      int // Number of candidates to fetch for reranking
}

func (r *RerankingRetriever) RetrieveWithConfig(ctx context.Context, query string, config *rag.RetrievalConfig) ([]rag.DocumentSearchResult, error) {
    // Step 1: Fetch more candidates using vector search
    queryEmbedding, _ := r.embedder.EmbedDocument(ctx, query)
    candidates, _ := r.vectorStore.Search(ctx, queryEmbedding, r.fetchK)

    // Step 2: Rerank the candidates
    reranked, _ := r.reranker.Rerank(ctx, query, candidates)

    return reranked, nil
}
```

**Note**: The Qwen embedder is now available as a reusable package at `github.com/smallnest/langgraphgo/llms/qwen`. You can use it in your own projects.

## Performance Considerations

### When to Use Reranking

**Use reranking when:**
- Accuracy is more important than latency
- You have a large document corpus (> 10K documents)
- Queries are complex and require deep understanding
- You need the most relevant results

**Skip reranking when:**
- Latency is critical
- Document corpus is small (< 1K documents)
- Queries are simple keyword matches
- Vector search already provides good results

### Retrieval Strategy

| Stage | Documents | Model | Latency | Accuracy |
|-------|-----------|-------|---------|----------|
| Vector Search | 10-50 | bi-encoder | ~10ms | Good |
| Reranking | 3-10 | cross-encoder | ~100ms | Excellent |

### Cost Optimization

1. **Tune fetch count**: Fetch fewer candidates if reranking is expensive
2. **Cache embeddings**: Pre-compute and cache document embeddings
3. **Batch requests**: Process multiple queries in parallel
4. **Use smaller models**: Consider smaller embedding models for initial retrieval

## Advanced Usage

### Custom Reranker

```go
type CustomReranker struct {
    llm *openai.LLM
}

func (r *CustomReranker) Retrieve(ctx context.Context, query string, k int) ([]rag.RAGDocument, error) {
    // Initial retrieval
    candidates, _ := vectorRetriever.Retrieve(ctx, query, k*5)

    // Rerank with custom logic
    for _, doc := range candidates {
        // Custom reranking logic
        doc.Score = calculateRelevance(query, doc)
    }

    // Sort and return top k
    return topK(candidates, k), nil
}
```

### Hybrid Search

```go
// Combine dense and sparse retrieval
hybridRetriever := retriever.NewHybridRetriever(
    retriever.NewVectorStoreRetriever(vectorStore, embedder, 10),
    retriever.NewBM25Retriever(documentStore, 10),
    0.7, // Weight for vector search
    0.3, // Weight for BM25
)
```

### Metadata Filtering

```go
// Filter by metadata before reranking
filteredRetriever := retriever.NewFilteredRetriever(
    vectorRetriever,
    map[string]any{
        "category": "technical",
        "priority": 1,
    },
)
```

## Troubleshooting

### API Errors

**Error**: `API returned unexpected status code: 401`

**Solution**: Check your API key is correct:
```bash
echo $OPENAI_API_KEY
```

**Error**: `API name not exist`

**Solution**: Verify the embedding model name:
```bash
export OPENAI_EMBEDDING_MODEL=text-embedding-v3
```

**Error**: `encoding_format must be 'float' or 'base64'`

**Solution**: The custom `QwenEmbedder` in this example handles this automatically by setting `encoding_format: "float"`. If you're implementing your own embedder, ensure you use one of these valid values.

**Error**: `API returned status 429: We have to rate limit you`

**Solution**: ModelScope's free API tier has rate limits. You can:
1. Wait a few seconds between requests
2. Use DashScope (Alibaba Cloud's commercial API) for higher limits
3. Use a different embedding provider like OpenAI

### Reranking Not Working

**Error**: `Failed to create LLM reranker`

**Solution**: The example will fall back to vector-only retrieval. Check:
1. LLM is properly configured
2. Model supports reranking capability
3. API credentials are valid

### Poor Retrieval Quality

**Possible causes**:
1. Documents are too short or lack context
2. Query is ambiguous
3. Embedding model not suitable for domain

**Solutions**:
1. Chunk documents into smaller, focused pieces
2. Add query expansion or rewriting
3. Use domain-specific embedding models

## Comparison with Other Approaches

| Approach | Latency | Accuracy | Complexity |
|----------|---------|----------|------------|
| Vector search only | Low | Good | Low |
| Vector + Reranker | Medium | Excellent | Medium |
| Pure LLM retrieval | High | Excellent | Low |
| Hybrid (dense + sparse) | Medium | Excellent | High |

## References

- [Qwen Documentation](https://qwen.readthedocs.io/)
- [DashScope API](https://help.aliyun.com/zh/dashscope/)
- [LangGraphGo RAG Documentation](../../rag/README.md)
- [Retriever Implementation](../../rag/retriever/)
- [Vector Store Options](../../rag/store/)
- [llms/qwen Package](../../llms/qwen/) - Reusable Qwen embedder package

## License

This example follows the LangGraphGo project license.
