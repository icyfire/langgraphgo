# RAG Reranker Comparison Example

This example demonstrates different reranking strategies available in LangGraphGo's RAG system. Reranking is a crucial step that improves retrieval quality by re-scoring documents based on their relevance to the query.

## Overview

The example compares four different reranker implementations:

| Reranker | Description | Pros | Cons |
|----------|-------------|------|------|
| **SimpleReranker** | Keyword-based reranking | Fast, no API calls, always available | Limited accuracy, keyword-only |
| **LLMReranker** | Uses LLM to score documents | Good semantic understanding, no new dependencies | Slower, higher API costs |
| **CohereReranker** | Uses Cohere's Rerank API | High quality results, fast | API costs, requires API key |
| **JinaReranker** | Uses Jina AI's Rerank API | High quality, multilingual support | API costs, requires API key |

## Prerequisites

```bash
# Required
OPENAI_API_KEY=your_key_here

# Optional (for Cohere reranker)
COHERE_API_KEY=your_key_here

# Optional (for Jina reranker)
JINA_API_KEY=your_key_here
```

## Running the Example

### Basic Usage (LLM-based Reranking)

```bash
OPENAI_API_KEY=sk-xxx go run main.go
```

This will run the example with SimpleReranker and LLMReranker.

### With Cohere Reranker

```bash
COHERE_API_KEY=your_key OPENAI_API_KEY=sk-xxx go run main.go
```

### With Jina Reranker

```bash
JINA_API_KEY=your_key OPENAI_API_KEY=sk-xxx go run main.go
```

### With All Rerankers

```bash
COHERE_API_KEY=your_key JINA_API_KEY=your_key OPENAI_API_KEY=sk-xxx go run main.go
```

## Output

The example will display:

1. **Retrieved Documents**: The documents retrieved and reranked
2. **Relevance Scores**: The relevance scores assigned by each reranker
3. **Generated Answer**: The final answer based on reranked documents

Example output:
```
--- LLMReranker ---
Top Retrieved Documents:
  [1] langgraph_intro.txt (Topic: LangGraph)
      LangGraph is a library for building stateful, multi-actor applications...
  [2] multi_agent.txt (Topic: Multi-Agent)
      Multi-agent systems in LangGraph enable multiple AI agents to work...

Relevance Scores:
  [1] Score: 0.8234 (Method: llm)
  [2] Score: 0.7891 (Method: llm)

Answer: LangGraph is a library designed for building stateful, multi-actor applications...
```

## Code Structure

```go
// Create base retriever
baseRetriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, 5)

// Create reranker
llmReranker := retriever.NewLLMReranker(llm, retriever.DefaultLLMRerankerConfig())

// Configure RAG pipeline with reranking
config := rag.DefaultPipelineConfig()
config.Retriever = baseRetriever
config.Reranker = llmReranker
config.UseReranking = true

// Build and run pipeline
pipeline := rag.NewRAGPipeline(config)
pipeline.BuildAdvancedRAG()
runnable, _ := pipeline.Compile()
result, _ := runnable.Invoke(ctx, map[string]any{"query": query})
```

## Customization

### Adjust TopK (number of results)

```go
config := retriever.DefaultLLMRerankerConfig()
config.TopK = 10  // Return top 10 documents instead of 5

llmReranker := retriever.NewLLMReranker(llm, config)
```

### Set Score Threshold

```go
config := retriever.DefaultCohereRerankerConfig()
config.TopK = 5

cohereReranker := retriever.NewCohereReranker(apiKey, config)

// In pipeline, filter by threshold after reranking
```

### Custom System Prompt (LLMReranker)

```go
config := retriever.LLMRerankerConfig{
    TopK: 5,
    ScoreThreshold: 0.5,
    SystemPrompt: "You are an expert technical document rater. Rate based on technical accuracy and completeness.",
    BatchSize: 5,
}

llmReranker := retriever.NewLLMReranker(llm, config)
```

### Different Cohere Models

```go
config := retriever.CohereRerankerConfig{
    Model: "rerank-english-v3.0",  // English-optimized
    // Model: "rerank-multilingual-v3.0",  // Multilingual
    TopK: 5,
}

cohereReranker := retriever.NewCohereReranker(apiKey, config)
```

### Different Jina Models

```go
config := retriever.JinaRerankerConfig{
    Model: "jina-reranker-v1-base-en",  // English only
    // Model: "jina-reranker-v2-base-multilingual",  // Multilingual
    TopK: 5,
}

jinaReranker := retriever.NewJinaReranker(apiKey, config)
```

## Cross-Encoder Reranking

For local, privacy-preserving reranking without API calls, you can use the CrossEncoderReranker with a local service:

### Setup Python Server

```bash
# Install dependencies
pip install sentence-transformers flask flask-cors

# Start the server
python ../../scripts/cross_encoder_server.py --port 8000
```

### Use in Go Code

```go
config := retriever.CrossEncoderRerankerConfig{
    APIBase:   "http://localhost:8000/rerank",
    ModelName: "cross-encoder/ms-marco-MiniLM-L-6-v2",
    TopK:      5,
}

ceReranker := retriever.NewCrossEncoderReranker(config)

// Use in pipeline
config := rag.DefaultPipelineConfig()
config.Reranker = ceReranker
```

## Performance Comparison

| Reranker | Speed | Quality | Cost | Privacy |
|----------|-------|---------|------|---------|
| SimpleReranker | ⚡⚡⚡ Fastest | ⭐⭐ Basic | Free | On-device |
| LLMReranker | ⚡⚡ Medium | ⭐⭐⭐⭐ Good | Per-token cost | Depends on LLM |
| CohereReranker | ⚡⚡⚡ Fast | ⭐⭐⭐⭐⭐ Excellent | Per-request cost | Cloud API |
| JinaReranker | ⚡⚡⚡ Fast | ⭐⭐⭐⭐⭐ Excellent | Per-request cost | Cloud API |
| CrossEncoder | ⚡⚡ Medium | ⭐⭐⭐⭐ Very Good | Free (local) | On-device |

## When to Use Each Reranker

- **SimpleReranker**: Quick prototyping, no external dependencies
- **LLMReranker**: When you already have an LLM, want semantic understanding without new services
- **CohereReranker**: Production use, English-focused applications
- **JinaReranker**: Production use, multilingual applications
- **CrossEncoder**: Privacy-sensitive applications, cost-sensitive deployments

## See Also

- [../../rag/RERANKER.md](../../rag/RERANKER.md) - Detailed implementation plan
- [../../rag/](../../rag/) - RAG package documentation
- [../rag_advanced/](../rag_advanced/) - Advanced RAG example with reranking
