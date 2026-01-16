# RAG with chromem-go VectorStore Example

This example demonstrates how to use **chromem-go vector database** with **LangGraphGo's RAG pipeline**.

## What is chromem-go?

chromem-go is a pure Go implementation of a vector database inspired by Chroma. It provides:

- **Zero external dependencies** - No Docker, no external services required
- **Embedded SQLite storage** - Persistent vector storage with optional in-memory mode
- **Thread-safe operations** - Safe concurrent access from multiple goroutines
- **Concurrent document processing** - Batch operations with worker pools
- **Multiple embedding functions** - Support for various embedding providers
- **Cosine similarity search** - Efficient vector similarity queries
- **Metadata filtering** - Filter search results by document metadata
- **Native Go implementation** - Written entirely in Go, no CGo dependencies

## Prerequisites

1. **Go 1.21+**: Required for building the example
2. **OpenAI API Key**: Required for both LLM calls and embeddings
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```

**Note**: This example uses real OpenAI embeddings (`text-embedding-3-small` by default) for generating vector representations of documents.

## Running the Example

```bash
cd examples/rag_chromem_example
go run main.go
```

## Features Demonstrated

- Creating a chromem-go vector store with persistent storage
- Adding documents with automatic embedding generation
- Building a RAG pipeline with chromem-go as the vector store
- Querying the pipeline with natural language questions
- Similarity search with relevance scores
- Metadata filtering for targeted searches
- Persistent storage verification across store instances
- Store statistics and collection management

## Code Highlights

### Creating OpenAI Embedder

```go
import (
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/smallnest/langgraphgo/rag"
)

// Create OpenAI LLM for embeddings
llm, err := openai.New()
if err != nil {
    log.Fatal(err)
}

// Create OpenAI embedder (uses text-embedding-3-small by default)
openaiEmbedder, err := embeddings.NewEmbedder(llm)
if err != nil {
    log.Fatal(err)
}

// Wrap with LangGraphGo adapter
embedder := rag.NewLangChainEmbedder(openaiEmbedder)
```

### Creating chromem-go Store

```go
// Simple initialization with in-memory storage
store, err := store.NewChromemVectorStoreSimple("", embedder)

// Or with full configuration
store, err := store.NewChromemVectorStore(store.ChromemConfig{
    PersistenceDir: "/path/to/storage",  // Empty for in-memory
    CollectionName: "my_collection",     // Defaults to "default"
    Embedder:       embedder,            // Required embedding function
})
```

### Adding Documents

```go
// Documents can be added without pre-computed embeddings
// The embedder will automatically generate embeddings
documents := []rag.Document{
    {
        ID:      "doc1",
        Content: "Your document content here",
        Metadata: map[string]any{
            "category": "tech",
            "source":   "docs",
        },
        // Embedding is optional - will be auto-generated if not provided
    },
}

err = vectorStore.Add(ctx, documents)
```

### Similarity Search

```go
// Basic search with query embedding
results, err := vectorStore.Search(ctx, queryEmbedding, k)

// Search with metadata filtering
results, err := vectorStore.SearchWithFilter(ctx, queryEmbedding, k, map[string]any{
    "category": "tech",
})

for _, result := range results {
    fmt.Printf("Score: %.4f - %s\n", result.Score, result.Document.Content)
}
```

### Persistent Storage

```go
// Create a persistent store
store, err := store.NewChromemVectorStoreSimple("/path/to/db", embedder)
defer store.Close()

// Add documents
store.Add(ctx, documents)

// Later, reopen the same store
store2, err := store.NewChromemVectorStoreSimple("/path/to/db", embedder)
// All documents are persisted!
```

### Using with RAG Pipeline

```go
// Create retriever
retriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, topK)

// Configure RAG pipeline
config := rag.DefaultPipelineConfig()
config.Retriever = retriever
config.LLM = llm

// Build and compile
pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()
runnable, _ := pipeline.Compile()

// Query
result, _ := runnable.Invoke(ctx, map[string]any{"query": "What is chromem-go?"})
```

## Configuration Options

### ChromemConfig

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `PersistenceDir` | string | Directory for SQLite storage (empty = in-memory) | "" |
| `CollectionName` | string | Name of the collection | "default" |
| `Embedder` | Embedder | Embedding function for generating vectors | Required |

## Storage Modes

### In-Memory Storage

```go
// Fast, temporary storage
store, err := store.NewChromemVectorStoreSimple("", embedder)
```

**Use cases:**
- Testing and development
- Temporary data processing
- Caching layer

### Persistent Storage

```go
// Data persists across restarts
store, err := store.NewChromemVectorStoreSimple("/data/vectors", embedder)
```

**Use cases:**
- Production applications
- Long-term data retention
- Distributed systems

## Expected Output

```
=== RAG with chromem-go VectorStore Example ===

Initializing chromem-go vector store...
Store created with collection: langgraphgo_example
Storage location: /tmp/chromem_example

Adding documents to chromem-go...
Successfully added 5 documents

Store Statistics:
  Total Documents: 5
  Vector Dimension: 128

Building RAG pipeline...

Pipeline Graph:
┌───────────────────────────────────────────────┐
│                   RAG Pipeline                 │
└───────────────────────────────────────────────┘

================================================================================
Query 1: What is chromem-go?
--------------------------------------------------------------------------------

Retrieved 2 documents:
  [1] Score: 0.8542
      chromem-go is a pure Go implementation of a vector database inspired by Chroma...
      Metadata: map[category:introduction source:chromem_docs]
  [2] Score: 0.7821
      LangGraphGo integrates seamlessly with chromem-go, providing a native Go vector store...
      Metadata: map[category:integration source:langgraphgo_docs]

Answer:
chromem-go is a pure Go implementation of a vector database...

================================================================================
Metadata Filtering Example
--------------------------------------------------------------------------------

Found 1 documents with category='features':
  [1] Key features of chromem-go include: zero external dependencies, thread-safe operations...
      Category: features

================================================================================
Persistent Storage Verification
--------------------------------------------------------------------------------

Reopened store - Documents persist: 5 documents
Data successfully persisted across store instances!

=== Example completed successfully! ===
```

## Advanced Usage

### Custom Embedding Functions

```go
// Use OpenAI embeddings (as shown in this example)
llm, _ := openai.New()
openaiEmbedder, _ := embeddings.NewEmbedder(llm)
embedder := rag.NewLangChainEmbedder(openaiEmbedder)

// Or use other embedding providers:
// - Cohere: embeddings.NewEmbedder(cohere.New())
// - Jina: embeddings.NewEmbedder(jina.New())
// - Ollama: embeddings.NewEmbedder(ollama.New())
// etc.
```

### Batch Operations

```go
// Adding many documents efficiently
documents := make([]rag.Document, 1000)
// ... populate documents
err = vectorStore.Add(ctx, documents)  // Automatic parallel processing
```

### Statistics and Monitoring

```go
stats, err := vectorStore.GetStats(ctx)
fmt.Printf("Documents: %d\n", stats.TotalDocuments)
fmt.Printf("Vectors: %d\n", stats.TotalVectors)
fmt.Printf("Dimension: %d\n", stats.Dimension)
```

## Comparison with Other Vector Stores

| Feature | chromem-go | Chroma | Pinecone | Weaviate |
|---------|------------|--------|----------|----------|
| External Service | No | Yes (Docker) | Yes | Yes |
| Dependencies | Zero | Docker | None | Docker |
| Persistent Storage | SQLite | SQLite/ClickHouse | Cloud | Cloud |
| Language | Pure Go | Python | SDK | Go/Python |
| Cost | Free | Free | Paid | Free/Paid |
| Best For | Go apps | Python apps | Production | Enterprise |

## Advantages of chromem-go

1. **No infrastructure overhead** - Run anywhere Go runs
2. **Simple deployment** - Single binary, no Docker required
3. **Fast startup** - Embedded database, no network latency
4. **Type-safe** - Native Go types and interfaces
5. **Easy testing** - In-memory mode for tests
6. **Low resource usage** - Minimal memory and CPU footprint

## Troubleshooting

### Permission denied when creating storage directory

**Error**: `failed to create persistence directory: permission denied`

**Solution**: Ensure the application has write permissions to the storage directory, or use a different location:
```go
store, err := store.NewChromemVectorStoreSimple(os.TempDir(), embedder)
```

### Embedder dimension mismatch

**Error**: `embedding dimension mismatch`

**Solution**: Ensure all embeddings have the same dimension:
```go
embedder := store.NewMockEmbedder(128)  // Fixed dimension
```

## Next Steps

- Experiment with different embedding functions
- Try metadata filtering for advanced search scenarios
- Implement hybrid search combining vector and keyword search
- Add document updates and deletions
- Explore concurrent batch operations

## References

- [chromem-go GitHub](https://github.com/philippgille/chromem-go)
- [LangGraphGo Documentation](../../docs/RAG/RAG.md)
- [RAG Architecture](../../docs/RAG/Architecture.md)
- [Vector Store Interface](../../rag/types.go)
