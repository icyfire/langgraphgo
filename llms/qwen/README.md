# llms/qwen

Qwen embedder implementation for LangGraphGo.

This package provides a Qwen embedding model client that supports the `encoding_format` parameter required by Qwen3-Embedding-4B and other Qwen models on ModelScope API.

## Features

- **encoding_format support**: Supports `float` and `base64` encoding formats (not `float32`)
- **Automatic retry**: Built-in retry logic with exponential backoff for rate limiting (429) and server errors (5xx)
- **Batch processing**: Efficient batch embedding of multiple texts
- **LangChain compatibility**: Implements both `rag.Embedder` and `langchaingo embeddings.Embedder` interfaces

## Installation

```bash
go get github.com/smallnest/langgraphgo/llms/qwen
```

## Usage

### Basic Usage

```go
import "github.com/smallnest/langgraphgo/llms/qwen"

// Create embedder with ModelScope API
embedder := qwen.NewEmbedder(
    "https://api-inference.modelscope.cn/v1",
    "your-api-key",
    "Qwen/Qwen3-Embedding-4B",
)

// Embed a single text
embedding, err := embedder.EmbedDocument(ctx, "Hello, world!")

// Embed multiple texts
embeddings, err := embedder.EmbedDocuments(ctx, []string{"text1", "text2"})
```

### With Options

```go
import "github.com/smallnest/langgraphgo/llms/qwen"

embedder := qwen.NewEmbedderWithOptions(
    qwen.WithBaseURL("https://api-inference.modelscope.cn/v1"),
    qwen.WithAPIKey("your-api-key"),
    qwen.WithModel("Qwen/Qwen3-Embedding-4B"),
)
```

### With RAG Pipeline

```go
import (
    "github.com/smallnest/langgraphgo/llms/qwen"
    "github.com/smallnest/langgraphgo/rag/store"
)

// Create embedder
embedder := qwen.NewEmbedder(
    "https://api-inference.modelscope.cn/v1",
    os.Getenv("MODELSCOPE_API_KEY"),
    "Qwen/Qwen3-Embedding-4B",
)

// Create vector store
vectorStore := store.NewInMemoryVectorStore(embedder)

// Add documents
err := vectorStore.Add(ctx, documents)
```

## API Configuration

### ModelScope (Recommended for Qwen Models)

```bash
export EMBEDDING_BASE_URL=https://api-inference.modelscope.cn/v1
export MODELSCOPE_API_KEY=your-modelscope-api-key
```

### DashScope (Alibaba Cloud)

```bash
export EMBEDDING_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
export OPENAI_API_KEY=your-dashscope-api-key
```

## Retry Configuration

The embedder automatically retries on:
- **429 Rate Limit**: Maximum 5 retries with exponential backoff (2s, 4s, 8s, 16s, 32s)
- **5xx Server Errors**: Same retry logic

Other errors (400, 401, etc.) are not retried.

## encoding_format Parameter

Qwen embedding models require the `encoding_format` parameter in the API request:

- **"float"**: Returns float32 arrays (default in this package)
- **"base64"**: Returns base64-encoded strings

**Note**: The value must be `"float"` or `"base64"`, **not** `"float32"`. Using `"float32"` will result in a 400 error.

## Supported Models

- `Qwen/Qwen3-Embedding-4B` - 4 billion parameters, 1536 dimensions
- `Qwen/Qwen2.5-72B-Instruct` - Chat model (not for embeddings)
- Other Qwen models available on ModelScope

## Error Handling

```go
embedding, err := embedder.EmbedDocument(ctx, "text")
if err != nil {
    if strings.Contains(err.Error(), "rate limit") {
        // Handle rate limiting
    } else if strings.Contains(err.Error(), "401") {
        // Handle authentication error
    }
}
```

## See Also

- [RAG with Qwen Example](../../examples/rag_qwen_ranker_example/)
- [Qwen Documentation](https://qwen.readthedocs.io/)
- [ModelScope API](https://www.modelscope.cn/)
