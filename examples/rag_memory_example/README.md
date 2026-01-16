# RAG with In-Memory VectorStore Example

This example demonstrates how to build a **RAG (Retrieval-Augmented Generation) pipeline** using **LangGraphGo's in-memory vector store** for quick prototyping and testing without external dependencies.

## What is In-Memory VectorStore?

The in-memory vector store is a lightweight, ephemeral storage solution that:
- **Runs entirely in memory** - No database setup required
- **Perfect for testing** - Ideal for development and experimentation
- **Fast and simple** - Zero configuration, instant startup
- **Embedding-agnostic** - Works with any embedder (OpenAI, mock, custom)
- **Non-persistent** - Data is lost when the program exits

## Prerequisites

1. **Go 1.21+**: Required for building the example
2. **API Key** (Optional): For using real LLM
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```
   **Note**: This example uses a mock embedder by default, so no API key is needed for basic testing.

## Running the Example

```bash
cd examples/rag_memory_example
go run main.go
```

## Features Demonstrated

- Creating an in-memory vector store with mock embeddings
- Adding documents with automatic embedding generation
- Building a complete RAG pipeline with memory
- Querying the pipeline with natural language questions
- Visualizing the RAG pipeline graph structure
- Understanding the retrieval and generation flow

## Code Highlights

### Creating Mock Embedder and In-Memory Store

```go
import (
    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/store"
)

// Create mock embedder (128-dimensional vectors for testing)
embedder := store.NewMockEmbedder(128)

// Create in-memory vector store (ephemeral, no persistence)
vectorStore := store.NewInMemoryVectorStore(embedder)
```

### Adding Documents

```go
// Create sample documents
documents := []rag.Document{
    {
        Content: "Chroma is an open-source vector database...",
        Metadata: map[string]any{"source": "chroma_docs"},
    },
    {
        Content: "LangGraphGo integrates with various vector stores...",
        Metadata: map[string]any{"source": "langgraphgo_docs"},
    },
}

// Add documents with embeddings
vectorStore.Add(ctx, documents)
```

### Building RAG Pipeline

```go
import "github.com/smallnest/langgraphgo/rag/retriever"

// Create retriever
retriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, 2)

// Configure RAG pipeline
config := rag.DefaultPipelineConfig()
config.Retriever = retriever
config.LLM = llm

// Build and compile pipeline
pipeline := rag.NewRAGPipeline(config)
err = pipeline.BuildBasicRAG()
runnable, err := pipeline.Compile()
```

### Querying the Pipeline

```go
result, err := runnable.Invoke(ctx, map[string]any{
    "query": "What is Chroma?",
})

if answer, ok := result["answer"].(string); ok {
    fmt.Printf("Answer: %s\n", answer)
}

if docs, ok := result["documents"].([]rag.RAGDocument); ok {
    for _, doc := range docs {
        fmt.Printf("Retrieved: %s\n", doc.Content)
    }
}
```

## Expected Output

```
    +-------------------------------------------------------------------+
    |                              ____________                          |
    |                             |            |                         |
    |                             |   START    |                         |
    |                             |____________|                         |
    |                                    |                               |
    |                                    v                               |
    |    ___________________________     __________________             |
    |   |                           |   |                  |            |
    |   |        retrieve_docs      |   |       cond       |            |
    |   |___________________________|   |__________________|            |
    |                                    |                               |
    |              _______________________|_____________________         |
    |             |                                           |          |
    |     ________v________                          _________v______    |
    |    |                 |                        |                |   |
    |    |  generate_answer|                        |      end       |   |
    |    |_________________|                        |________________|   |
    |                                                                   |
    +-------------------------------------------------------------------+

Query: What is Chroma?

Retrieved Documents:
  [1] Chroma is an open-source vector database that allows you to store and query embeddings...
  [2] LangGraphGo integrates with various vector stores including Chroma...

Answer: Chroma is an open-source vector database designed for storing and querying embeddings...
```

## In-Memory vs Persistent Vector Stores

| Feature | In-Memory Store | Persistent Stores (Chroma, etc.) |
|---------|----------------|----------------------------------|
| **Setup** | Zero config | Requires server/cluster |
| **Persistence** | Lost on exit | Durable storage |
| **Use Case** | Testing, prototyping | Production |
| **Performance** | Fast (local) | Network latency |
| **Scalability** | Memory limited | Horizontal scaling |
| **Concurrency** | Single process | Multi-client |

## When to Use In-Memory Store

✅ **Good for**:
- Quick prototyping and experimentation
- Unit testing RAG pipelines
- Learning RAG concepts
- Development without external dependencies
- Benchmarking and performance testing

❌ **Not for**:
- Production applications
- Large document collections
- Multi-user scenarios
- Long-term data retention

## Customizing the Example

### Using Real Embeddings

Replace the mock embedder with OpenAI:

```go
import (
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/openai"
)

// Create OpenAI LLM and embedder
llm, _ := openai.New()
openaiEmbedder, _ := embeddings.NewEmbedder(llm)
embedder := rag.NewLangChainEmbedder(openaiEmbedder)

// Use real embedder with in-memory store
vectorStore := store.NewInMemoryVectorStore(embedder)
```

### Adjusting Retrieval Parameters

```go
// Retrieve more documents
retriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, 5)

// Or use custom retrieval config
retriever := retriever.NewVectorStoreRetrieverWithConfig(
    vectorStore,
    embedder,
    &retriever.Config{
        K: 3,
        ScoreThreshold: 0.7,
    },
)
```

### Adding More Documents

```go
documents := []rag.Document{
    {
        Content: "Your custom document here...",
        Metadata: map[string]any{
            "source": "custom",
            "category": "testing",
        },
    },
    // Add more documents...
}
vectorStore.Add(ctx, documents)
```

## Troubleshooting

### No API key error

**Error**: `Failed to create LLM: OPENAI_API_KEY not found`

**Solution**: Either:
1. Set your OpenAI API key: `export OPENAI_API_KEY="your-key"`
2. Use a mock LLM for testing (modify the code)

### Empty retrieval results

**Issue**: Retrieved documents list is empty

**Solution**:
- Ensure documents are added before querying
- Check that embeddings are generated successfully
- Verify the query text is relevant to stored documents

### Build errors

**Error**: `package github.com/smallnest/langgraphgo/... not found`

**Solution**: Run from the langgraphgo root directory:
```bash
cd /path/to/langgraphgo
go run examples/rag_memory_example/main.go
```

## Next Steps

- **Try persistent stores**: Replace in-memory store with Chroma, Pinecone, or Weaviate
- **Add metadata filtering**: Filter documents by metadata during retrieval
- **Implement hybrid search**: Combine keyword and vector search
- **Add streaming**: Use streaming responses for real-time feedback
- **Multi-turn conversations**: Add conversation memory to the RAG pipeline

## Advanced RAG Patterns

### Retrieval with Metadata Filtering

```go
// Store documents with metadata
documents := []rag.Document{
    {
        Content: "...",
        Metadata: map[string]any{
            "category": "technical",
            "language": "en",
        },
    },
}

// Later, search with metadata filter (if supported by store)
results, _ := vectorStore.SearchWithFilter(ctx, queryEmbedding, k, map[string]any{
    "category": "technical",
})
```

### Custom Retrieval Strategy

```go
// Implement custom retrieval logic
type CustomRetriever struct {
    vectorStore *store.InMemoryVectorStore
    embedder    rag.Embedder
}

func (r *CustomRetriever) Retrieve(ctx context.Context, query string, k int) ([]rag.Document, error) {
    // Custom retrieval logic here
    // e.g., reranking, filtering, transformation
}
```

## References

- [LangGraphGo RAG Documentation](../../docs/RAG/RAG.md)
- [Vector Store Interface](../../rag/types.go)
- [RAG Pipeline Configuration](../../rag/pipeline.go)
- [Retriever Patterns](../../rag/retriever/)
- [Example with Chroma](../chroma-v2-example/README.md)
- [Example with Local Files](../rag_local_files_example/README.md)
