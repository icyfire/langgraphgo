# RAG with Chroma v2 API VectorStore Example

This example demonstrates how to use **Chroma v2 API vector database** with **LangGraphGo's RAG pipeline** through the native `ChromaV2VectorStore` implementation.

## What is Chroma v2?

Chroma v2 API introduces significant improvements over v1:

- **Hierarchical Structure**: Organizes resources into tenant → database → collection hierarchy
- **RESTful API Design**: Follows OpenAPI 3.1.0 specification
- **Improved Authentication**: Better auth and token management
- **Granular Permissions**: Fine-grained access control
- **Better Performance**: Optimized query and indexing operations
- **Multi-tenancy Support**: Built-in support for multiple tenants and databases

## Prerequisites

1. **Go 1.21+**: Required for building the example
2. **Chroma Server v2**: Required to be running locally or remotely
   ```bash
   # Start Chroma v2 with Docker (latest version)
   docker run -p 8000:8000 chromadb/chroma

   # Or specific version
   docker run -p 8000:8000 chromadb/chroma:0.6.0
   ```
3. **OpenAI API Key**: Required for both LLM calls and embeddings
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```

**Note**: This example uses real OpenAI embeddings for generating vector representations of documents.

## Running the Example

```bash
cd examples/chroma-v2-example
go run main.go
```

**Optional**: Set a custom Chroma URL
```bash
export CHROMA_URL="http://localhost:8000"
go run main.go
```

## Features Demonstrated

- Creating a Chroma v2 vector store with native implementation
- Understanding tenant/database/collection hierarchy
- Adding documents with automatic embedding generation
- Building a RAG pipeline with Chroma v2 as the vector store
- Querying the pipeline with natural language questions
- Direct similarity search through Chroma v2 API
- Metadata filtering for targeted searches
- Visualization of the RAG pipeline graph

## Code Highlights

### Creating OpenAI Embedder

```go
import (
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/smallnest/langgraphgo/rag"
)

// Create OpenAI LLM
llm, err := openai.New()
if err != nil {
    log.Fatal(err)
}

// Create OpenAI embedder
openaiEmbedder, err := embeddings.NewEmbedder(llm)
if err != nil {
    log.Fatal(err)
}

// Wrap with LangGraphGo adapter
embedder := rag.NewLangChainEmbedder(openaiEmbedder)
```

### Creating Chroma v2 Store

```go
import "github.com/smallnest/langgraphgo/rag/store"

// Simple initialization
store, err := store.NewChromaV2VectorStoreSimple(
    "http://localhost:8000",  // Chroma server URL
    "my_collection",         // Collection name
    embedder,                // Embedder
)

// Or with full configuration
store, err := store.NewChromaV2VectorStore(store.ChromaV2Config{
    BaseURL:    "http://localhost:8000",
    Tenant:     "default_tenant",
    Database:   "default_database",
    Collection: "my_collection",
    Embedder:   embedder,
})
```

### Adding Documents

```go
documents := []rag.Document{
    {
        ID:      "doc1",
        Content: "Your document content",
        Metadata: map[string]any{
            "category": "tech",
            "source":   "docs",
        },
    },
}

err = store.Add(ctx, documents)
```

### Similarity Search

```go
// Generate query embedding
queryEmbedding, _ := embedder.EmbedDocument(ctx, "search query")

// Basic similarity search
results, err := store.Search(ctx, queryEmbedding, k)

// Search with metadata filtering
results, err := store.SearchWithFilter(ctx, queryEmbedding, k,
    map[string]any{"category": "tech"},
)

for _, result := range results {
    fmt.Printf("Score: %.4f - %s\n", result.Score, result.Document.Content)
}
```

### Using with RAG Pipeline

```go
import (
    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/retriever"
)

// Create retriever
retriever := retriever.NewVectorStoreRetriever(store, embedder, topK)

// Configure RAG pipeline
config := rag.DefaultPipelineConfig()
config.Retriever = retriever
config.LLM = llm

// Build and compile
pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()
runnable, _ := pipeline.Compile()

// Query
result, _ := runnable.Invoke(ctx, map[string]any{"query": "What is Chroma v2?"})
```

## Configuration Options

### ChromaV2Config

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `BaseURL` | string | URL of the Chroma server | Required |
| `Tenant` | string | Tenant name | "default_tenant" |
| `Database` | string | Database name | "default_database" |
| `Collection` | string | Collection name | Required |
| `CollectionID` | string | Collection UUID (optional) | Auto-generated |
| `Embedder` | Embedder | Embedding function | Required |
| `HTTPClient` | *http.Client | HTTP client | Default 30s timeout |

## Chroma v2 API Structure

### Hierarchy

```
Tenant (e.g., "default_tenant")
  └── Database (e.g., "default_database")
      └── Collection (e.g., "my_collection")
          └── Records (documents with embeddings)
```

### Key Endpoints

- `POST /api/v2/tenants/{tenant}/databases/{database}/collections` - Create collection
- `POST /api/v2/tenants/{tenant}/databases/{database}/collections/{id}/add` - Add records
- `POST /api/v2/tenants/{tenant}/databases/{database}/collections/{id}/search` - Search
- `POST /api/v2/tenants/{tenant}/databases/{database}/collections/{id}/delete` - Delete records
- `POST /api/v2/tenants/{tenant}/databases/{database}/collections/{id}/upsert` - Update records

## Expected Output

```
=== RAG with Chroma v2 API VectorStore Example ===

Connecting to Chroma v2 at: http://localhost:8000

Store created with collection: langgraphgo_example (ID: 12345678-1234-1234-1234-123456789abc)

Adding documents to Chroma v2...
Successfully added 5 documents

Store Statistics:
  Total Documents: 5
  Vector Dimension: 1536

Building RAG pipeline...

Pipeline Graph:
┌───────────────────────────────────────────────┐
│                   RAG Pipeline                 │
└───────────────────────────────────────────────┘

================================================================================
Query 1: What is new in Chroma v2?
--------------------------------------------------------------------------------

Direct Chroma v2 similarity search results:
  [1] Score: 0.8542
      Chroma v2 API introduces a new hierarchical structure with tenants...
      Metadata: map[category:architecture source:chroma_v2_docs]
  [2] Score: 0.7821
      LangGraphGo provides native Chroma v2 support through ChromaV2VectorStore...
      Metadata: map[category:integration source:langgraphgo_docs]

RAG Answer:
Chroma v2 introduces a hierarchical structure...

================================================================================
Metadata Filtering Example (Chroma v2 Native)
--------------------------------------------------------------------------------

Found 2 documents with category='architecture':
  [1] Score: 0.8542
      Chroma v2 API introduces a new hierarchical structure...
      Category: architecture
  [2] Score: 0.7234
      Chroma v2 separates concepts into tenant, database, and collection...
      Category: architecture

=== Example completed successfully! ===

Note: This example uses Chroma v2 API.
Chroma v2 is the default in latest Chroma versions.
Start Chroma with: docker run -p 8000:8000 chromadb/chroma
```

## Troubleshooting

### Connection refused

**Error**: `Failed to create Chroma v2 store: connection refused`

**Solution**: Ensure Chroma server is running:
```bash
# Check if Chroma is running
curl http://localhost:8000/api/v2/heartbeat

# Start Chroma if not running
docker run -p 8000:8000 chromadb/chroma
```

### OpenAI API key not found

**Error**: `Failed to create embedder: OPENAI_API_KEY not found`

**Solution**: Set your OpenAI API key:
```bash
export OPENAI_API_KEY="your-api-key"
```

### Collection not found

**Error**: `Failed to search: status 404`

**Solution**: The collection will be auto-created on first use. If you see this error, check:
1. The collection name is correct
2. The tenant and database exist
3. You have proper permissions

### Metadata filtering returns no results

**Issue**: Metadata filtering may return 0 results even when documents with matching metadata exist.

**Explanation**: Some Chroma v2 server versions may not fully support metadata storage and filtering. The `ChromaV2VectorStore` implementation correctly sends metadata in the API payload, but the server may store it as `null` regardless.

**Status**: This is a known limitation of certain Chroma v2 server versions. The basic vector search functionality works correctly - only metadata filtering may be affected.

**Workaround**:
- Use Chroma v2 server versions that fully support metadata
- Store filtering criteria as part of the document content instead
- Use multiple collections for different categories instead of metadata filtering

## Chroma v2 vs v1

| Feature | v1 API | v2 API |
|---------|-------|-------|
| URL Structure | `/api/v1/...` | `/api/v2/tenants/{t}/databases/{d}/...` |
| Organization | Flat (collections only) | Hierarchical (tenant/db/collection) |
| Auth | Basic token | Enhanced auth with identity |
| Specification | Custom | OpenAPI 3.1.0 |
| Multi-tenancy | Limited | Full support |
| Distance Function | Config option | Part of collection config |

## Advantages of Chroma v2

1. **Better Organization** - Hierarchical structure for resource management
2. **Improved Security** - Enhanced authentication and authorization
3. **Standards Compliant** - OpenAPI specification for better integration
4. **Performance** - Optimized query execution
5. **Scalability** - Better support for large-scale deployments
6. **Developer Experience** - Clear API structure and documentation

## Next Steps

- Experiment with different tenant/database configurations
- Try metadata filtering for advanced search scenarios
- Implement custom HTTP clients with retry logic
- Use Chroma Cloud for production deployments
- Explore the full Chroma v2 API specification

## References

- [Chroma Documentation](https://docs.trychroma.com/)
- [Chroma v2 API Spec](http://localhost:8000/openapi.json) (when server is running)
- [LangGraphGo Documentation](../../docs/RAG/RAG.md)
- [RAG Architecture](../../docs/RAG/Architecture.md)
- [Vector Store Interface](../../rag/types.go)
