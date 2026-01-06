# RAG Reranker Implementation Plan

## Overview

This document describes the implementation plan for providing reranking capabilities in the LangGraphGo RAG system.

## Background

Reranking is a crucial step in RAG pipelines that improves retrieval quality by re-scoring retrieved documents based on their relevance to the query. While the initial retrieval uses bi-encoder models (fast but less accurate), reranking can use more sophisticated models to capture fine-grained query-document interactions.

## Current State

The project already has:
- `Reranker` interface defined in `rag/types.go`
- `SimpleReranker` implementation in `rag/retriever/reranker.go` (keyword-based)
- Pipeline support via `BuildAdvancedRAG()` and `BuildConditionalRAG()`
- Example usage in `examples/rag_advanced/main.go`

**Limitation**: `SimpleReranker` is too basic for production use.

## Implementation Plan

### 1. LLMReranker

**File**: `rag/retriever/llm_reranker.go`

Uses an LLM to score query-document pairs. The LLM is prompted to rate relevance on a scale.

**Pros**:
- No additional dependencies
- Good semantic understanding
- Works with any LLM

**Cons**:
- Slower than dedicated models
- Higher API costs

### 2. CrossEncoderReranker

**File**: `rag/retriever/cross_encoder_reranker.go`

Uses local cross-encoder models (BERT, MiniLM, etc.) for scoring.

**Pros**:
- Fast after model loaded
- No API costs
- Good accuracy

**Cons**:
- Requires additional dependencies (huggingface.co/go-transformers)
- Model download on first use

### 3. CohereReranker

**File**: `rag/retriever/cohere_reranker.go`

Uses Cohere's Rerank API.

**Pros**:
- High quality results
- Easy to use
- Fast

**Cons**:
- API costs
- Requires API key

### 4. JinaReranker

**File**: `rag/retriever/jina_reranker.go`

Uses Jina AI's Rerank API.

**Pros**:
- High quality results
- Competitive pricing
- Fast

**Cons**:
- API costs
- Requires API key

## File Structure

```
rag/
├── types.go                      # Core types and interfaces
├── retriever/
│   ├── reranker.go              # Existing SimpleReranker
│   ├── llm_reranker.go          # NEW: LLM-based reranker
│   ├── cross_encoder_reranker.go # NEW: Cross-encoder reranker
│   ├── cohere_reranker.go       # NEW: Cohere API reranker
│   ├── jina_reranker.go         # NEW: Jina API reranker
│   └── reranker_test.go         # Tests
```

## Usage Example

```go
import (
    "github.com/smallnest/langgraphgo/rag/retriever"
)

// Using LLM Reranker
llmReranker := retriever.NewLLMReranker(llm, retriever.LLMRerankerConfig{
    TopK: 5,
    ScoreThreshold: 0.5,
})

// Using Cohere Reranker
cohereReranker := retriever.NewCohereReranker(apiKey, retriever.CohereRerankerConfig{
    Model: "rerank-v3.5",
    TopK: 5,
})

// In RAG Pipeline
config := rag.DefaultPipelineConfig()
config.Reranker = cohereReranker
config.UseReranking = true
```

## Testing Strategy

1. Unit tests for each reranker implementation
2. Integration tests with mock APIs
3. End-to-end tests in RAG pipeline
4. Example demonstrating usage

## API Design

All rerankers follow the same interface:

```go
type Reranker interface {
    Rerank(ctx context.Context, query string, documents []rag.DocumentSearchResult) ([]rag.DocumentSearchResult, error)
}
```

Each reranker has its own config struct:

```go
type LLMRerankerConfig struct {
    TopK             int     // Number of documents to return
    ScoreThreshold   float64 // Minimum relevance score (0-1)
    SystemPrompt     string  // Custom system prompt
}

type CrossEncoderRerankerConfig struct {
    ModelName      string  // Model name (e.g., "cross-encoder/ms-marco-MiniLM-L-6-v2")
    TopK           int     // Number of documents to return
    ScoreThreshold float64 // Minimum relevance score
    Device         string  // "cpu" or "cuda"
}

type CohereRerankerConfig struct {
    Model    string  // "rerank-v3.5", "rerank-english-v3.0", etc.
    TopK     int     // Number of documents to return
    APIBase  string  // Custom API base URL (optional)
}

type JinaRerankerConfig struct {
    Model    string  // "jina-reranker-v1-base-en", "jina-reranker-v2-base-multilingual"
    TopK     int     // Number of documents to return
    APIBase  string  // Custom API base URL (optional)
}
```

## Dependencies

| Reranker | New Dependencies |
|----------|-----------------|
| LLMReranker | None (uses existing llms.Model) |
| CrossEncoderReranker | `github.com/huggingface/go-transformers` (optional) |
| CohereReranker | `github.com/cohere/cohere-go` |
| JinaReranker | Standard `net/http` (no SDK) |

## Implementation Priority

1. **Phase 1 (High Priority)**: LLMReranker + CohereReranker
   - These provide the most value with minimal complexity

2. **Phase 2 (Medium Priority)**: JinaReranker
   - Alternative to Cohere with competitive pricing

3. **Phase 3 (Optional)**: CrossEncoderReranker
   - For users who prefer local models
   - Requires more complex setup

## Notes

- All rerankers preserve the `Reranker` interface for drop-in replacement
- Default configs are provided for ease of use
- Errors are propagated to the caller for proper handling
- Context is respected for cancellation and timeouts
