# RAG Package - Retrieval-Augmented Generation for LangGraph Go

The `rag` package provides comprehensive RAG (Retrieval-Augmented Generation) capabilities for the LangGraph Go framework. It supports multiple retrieval strategies including traditional vector-based search and advanced GraphRAG with knowledge graphs.

## Features

- **Multiple Retrieval Strategies**: Vector similarity, graph-based, and hybrid approaches
- **Knowledge Graph Integration**: Automatic entity and relationship extraction
- **Document Processing**: Various loaders and intelligent text splitters
- **Flexible Storage**: Support for multiple vector stores and graph databases
- **LangGraph Integration**: Seamless integration with agents and workflows via StateGraph
- **Scalable Architecture**: Modular engines, retrievers, and stores

## Quick Start

### Basic Vector RAG

```go
package main

import (
    "context"
    "fmt"

    "github.com/smallnest/langgraphgo/rag/engine"
    "github.com/tmc/langchaingo/embeddings/openai"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/vectorstores/pgvector"
)

func main() {
    ctx := context.Background()

    // Initialize components
    llm, _ := openai.New()
    embedder, _ := openai.NewEmbedder()

    // Setup vector store (using adapter if needed)
    store, _ := pgvector.New(ctx, pgvector.WithConnString("postgres://localhost/postgres"))

    // Create vector RAG engine
    vectorRAG, _ := engine.NewVectorRAGEngine(llm, embedder, store, 5)

    // Query (Retrieves relevant context)
    result, _ := vectorRAG.Query(ctx, "What is quantum computing?")
    fmt.Printf("Context: %s\nSources: %d\n", result.Context, len(result.Sources))
}
```

### GraphRAG with Knowledge Graph

```go
import (
    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/engine"
)

// Create GraphRAG engine
graphRAG, _ := engine.NewGraphRAGEngine(rag.GraphRAGConfig{
    DatabaseURL:     "falkordb://localhost:6379",
    ModelProvider:   "openai",
    EntityTypes:     []string{"PERSON", "ORGANIZATION", "LOCATION"},
    MaxDepth:        3,
    EnableReasoning: true,
}, llm, embedder, knowledgeGraph)

// Add documents to extract entities and relationships
docs := []rag.Document{{
    ID: "doc1",
    Content: "Apple Inc. is headquartered in Cupertino, California and led by CEO Tim Cook.",
}}
graphRAG.AddDocuments(ctx, docs)

// Query using graph-enhanced retrieval
result, _ := graphRAG.Query(ctx, "Who is the CEO of Apple?")
fmt.Printf("Graph Context: %s\n", result.Context)
```

### Hybrid RAG

```go
import (
    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/retriever"
)

// Create individual retrievers
vectorRetriever := retriever.NewVectorRetriever(vectorStore, embedder, 3)
graphRetriever := retriever.NewGraphRetriever(knowledgeGraph, 2)

// Combine with weighted approach
hybridRetriever := retriever.NewHybridRetriever(
    []rag.Retriever{vectorRetriever, graphRetriever},
    []float64{0.6, 0.4}, // 60% vector, 40% graph
    &rag.RetrievalConfig{K: 5},
)

// Create hybrid RAG engine using base engine
hybridRAG := rag.NewBaseEngine(hybridRetriever, embedder, &rag.Config{
    VectorRAG: &rag.VectorRAGConfig{
        EnableReranking: true,
    },
})
```

## Architecture

### Core Components

#### Engines (rag/engine/)
- `VectorRAGEngine`: Traditional vector-based retrieval
- `GraphRAGEngine`: Knowledge graph-based retrieval with entity extraction
- `CompositeEngine`: Combines multiple retrieval strategies

#### Retrievers (rag/retriever/)
- `VectorRetriever`: Vector similarity search
- `GraphRetriever`: Entity-based traversal
- `HybridRetriever`: Weighted combination of multiple retrievers

#### Document Processing
- **Loaders** (`rag/loader/`): `TextLoader`, `StaticLoader`
- **Splitters** (`rag/splitter/`): `RecursiveCharacterTextSplitter`, `SimpleTextSplitter`
- **Adapters** (`rag/adapters.go`): Integration with `langchaingo` components

#### Storage (rag/store/)
- **Vector Stores**: `VectorStore` interface with various implementations
- **Knowledge Graphs**: `KnowledgeGraph` interface for graph databases

## Pipeline Usage

For a full RAG experience (Retrieve + Generate), use the `RAGPipeline`:

```go
import "github.com/smallnest/langgraphgo/rag"

config := rag.DefaultPipelineConfig()
config.Retriever = myRetriever
config.LLM = myLLM

pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()

runnable, _ := pipeline.Compile()
result, _ := runnable.Invoke(ctx, rag.RAGState{
    Query: "Explain the benefits of RAG",
})

fmt.Printf("Answer: %s\n", result.(rag.RAGState).Answer)
```

## Examples

See the root `examples/` directory for comprehensive demonstrations of:
- Basic vector RAG
- GraphRAG with entity extraction
- Hybrid retrieval strategies
- Integration with LangGraph agents

## License

This package is part of the LangGraph Go project and is licensed under the MIT License.