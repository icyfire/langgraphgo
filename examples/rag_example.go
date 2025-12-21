package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/prebuilt"
	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/loaders"
	"github.com/smallnest/langgraphgo/rag/retrievers"
	"github.com/smallnest/langgraphgo/rag/splitters"
	"github.com/tmc/langchaingo/embeddings/openai"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	ctx := context.Background()

	// Example 1: Basic Vector RAG
	fmt.Println("=== Example 1: Basic Vector RAG ===")
	if err := basicVectorRAGExample(ctx); err != nil {
		log.Printf("Vector RAG example failed: %v", err)
	}

	// Example 2: Graph RAG with Memory Graph
	fmt.Println("\n=== Example 2: Graph RAG with Memory Graph ===")
	if err := graphRAGExample(ctx); err != nil {
		log.Printf("Graph RAG example failed: %v", err)
	}

	// Example 3: Hybrid RAG (Vector + Graph)
	fmt.Println("\n=== Example 3: Hybrid RAG ===")
	if err := hybridRAGExample(ctx); err != nil {
		log.Printf("Hybrid RAG example failed: %v", err)
	}

	// Example 4: RAG with LangGraph Agent
	fmt.Println("\n=== Example 4: RAG with LangGraph Agent ===")
	if err := ragAgentExample(ctx); err != nil {
		log.Printf("RAG Agent example failed: %v", err)
	}
}

// basicVectorRAGExample demonstrates basic vector-based RAG
func basicVectorRAGExample(ctx context.Context) error {
	// Initialize OpenAI components
	llm, err := openai.New()
	if err != nil {
		return fmt.Errorf("failed to create LLM: %w", err)
	}

	embedder, err := openai.NewEmbedder()
	if err != nil {
		return fmt.Errorf("failed to create embedder: %w", err)
	}

	// Create a simple in-memory vector store (in practice, use a real vector store)
	vectorStore := &MockVectorStore{}

	// Create vector RAG engine
	vectorRAG, err := rag.NewVectorRAGEngine(llm, embedder, vectorStore, 5)
	if err != nil {
		return fmt.Errorf("failed to create vector RAG: %w", err)
	}

	// Load documents
	loader := loaders.NewTextLoader("testdata/sample.txt", loaders.WithMetadata(map[string]any{
		"category": "sample",
		"author":   "example",
	}))

	docs, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load documents: %w", err)
	}

	// Add documents to RAG engine
	if err := vectorRAG.AddDocuments(ctx, docs); err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}

	// Perform query
	result, err := vectorRAG.Query(ctx, "What is the main topic?")
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}

	fmt.Printf("Query: %s\n", result.Query)
	fmt.Printf("Context: %s\n", result.Context)
	fmt.Printf("Confidence: %.2f\n", result.Confidence)
	fmt.Printf("Sources: %d documents\n", len(result.Sources))

	return nil
}

// graphRAGExample demonstrates GraphRAG with an in-memory knowledge graph
func graphRAGExample(ctx context.Context) error {
	// Initialize LLM for entity extraction
	llm, &MockLLM{})
	if err != nil {
		return fmt.Errorf("failed to create LLM: %w", err)
	}

	// Initialize embedder
	embedder, &MockEmbedder{})

	// Create GraphRAG engine with memory graph
	graphRAG, err := rag.NewGraphRAGEngine(rag.GraphRAGConfig{
		DatabaseURL:     "memory://localhost",
		ModelProvider:   "mock",
		EmbeddingModel:  "mock-embedder",
		ChatModel:       "mock-chat",
		EntityTypes:     []string{"PERSON", "ORGANIZATION", "LOCATION", "CONCEPT"},
		MaxDepth:        3,
		EnableReasoning: true,
	}, llm, embedder)
	if err != nil {
		return fmt.Errorf("failed to create GraphRAG: %w", err)
	}

	// Sample documents with entities
	sampleDocs := []rag.Document{
		{
			ID: "doc1",
			Content: "Apple Inc. is a technology company headquartered in Cupertino, California. " +
				"The company was founded by Steve Jobs and is now led by CEO Tim Cook.",
			Metadata: map[string]any{
				"source": "tech_news",
			},
		},
		{
			ID: "doc2",
			Content: "Tim Cook became the CEO of Apple after Steve Jobs stepped down in 2011. " +
				"Under his leadership, Apple has continued to innovate in the smartphone market.",
			Metadata: map[string]any{
				"source": "business_news",
			},
		},
	}

	// Add documents to extract entities and relationships
	if err := graphRAG.AddDocuments(ctx, sampleDocs); err != nil {
		return fmt.Errorf("failed to add documents to GraphRAG: %w", err)
	}

	// Query about entities
	result, err := graphRAG.Query(ctx, "Who is the CEO of Apple?")
	if err != nil {
		return fmt.Errorf("failed to query GraphRAG: %w", err)
	}

	fmt.Printf("Query: %s\n", result.Query)
	fmt.Printf("Context: %s\n", result.Context)
	fmt.Printf("Confidence: %.2f\n", result.Confidence)

	if result.Metadata != nil {
		fmt.Printf("Entities found: %v\n", result.Metadata["entities_found"])
		fmt.Printf("Relationships: %v\n", result.Metadata["relationships"])
	}

	return nil
}

// hybridRAGExample demonstrates hybrid RAG combining vector and graph retrieval
func hybridRAGExample(ctx context.Context) error {
	// Initialize components
	llm, &MockLLM{})
	if err != nil {
		return fmt.Errorf("failed to create LLM: %w", err)
	}

	embedder, &MockEmbedder{})

	// Create individual retrievers
	vectorRetriever := retrievers.NewVectorRetriever(&MockVectorStore{}, embedder, rag.RetrievalConfig{
		K:              3,
		ScoreThreshold: 0.5,
		SearchType:     "similarity",
	})

	// Create knowledge graph
	graphStore, err := rag.NewKnowledgeGraph("memory://localhost")
	if err != nil {
		return fmt.Errorf("failed to create knowledge graph: %w", err)
	}

	graphRetriever := retrievers.NewGraphRetriever(graphStore, embedder, rag.RetrievalConfig{
		K:              2,
		ScoreThreshold: 0.3,
		SearchType:     "graph",
	})

	// Create hybrid retriever with weights
	hybridRetriever := retrievers.NewHybridRetriever(
		[]rag.Retriever{vectorRetriever, graphRetriever},
		[]float64{0.6, 0.4}, // Give more weight to vector search
		rag.RetrievalConfig{
			K:              5,
			ScoreThreshold: 0.4,
			SearchType:     "hybrid",
		},
	)

	// Create RAG engine with hybrid retriever
	hybridRAG := rag.NewBaseEngine(hybridRetriever, embedder, &rag.Config{
		RetrieverK:      5,
		ScoreThreshold:  0.4,
		SearchType:      "hybrid",
		EnableReranking: true,
	})

	// Query using hybrid approach
	result, err := hybridRAG.Query(ctx, "What are the latest developments in AI technology?")
	if err != nil {
		return fmt.Errorf("failed to query hybrid RAG: %w", err)
	}

	fmt.Printf("Query: %s\n", result.Query)
	fmt.Printf("Sources: %d documents retrieved\n", len(result.Sources))
	fmt.Printf("Confidence: %.2f\n", result.Confidence)
	fmt.Printf("Context length: %d characters\n", len(result.Context))

	return nil
}

// ragAgentExample demonstrates integrating RAG with a LangGraph agent
func ragAgentExample(ctx context.Context) error {
	// Initialize LLM
	llm, &MockLLM{})
	if err != nil {
		return fmt.Errorf("failed to create LLM: %w", err)
	}

	// Create RAG tool
	embedder, &MockEmbedder{})
	vectorStore := &MockVectorStore{}
	ragEngine, err := rag.NewVectorRAGEngine(llm, embedder, vectorStore, 3)
	if err != nil {
		return fmt.Errorf("failed to create RAG engine: %w", err)
	}

	// Add some sample documents
	sampleDocs := []rag.Document{
		{
			ID: "rag_doc_1",
			Content: "Retrieval-Augmented Generation (RAG) combines the benefits of parametric " +
				"and non-parametric approaches to language model generation.",
			Metadata: map[string]any{
				"category": "technical",
				"topic":    "RAG",
			},
		},
		{
			ID: "rag_doc_2",
			Content: "LangGraph provides a framework for building stateful, multi-agent applications " +
				"with language models using graph-based workflows.",
			Metadata: map[string]any{
				"category": "technical",
				"topic":    "LangGraph",
			},
		},
	}

	if err := ragEngine.AddDocuments(ctx, sampleDocs); err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}

	// Create a simple state graph with RAG integration
	g := graph.NewMessageGraph()

	// Define RAG-enhanced generation node
	g.AddNode("rag_enhanced_generation", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		// Extract query from messages
		messages, ok := state["messages"].([]any)
		if !ok || len(messages) == 0 {
			return state, fmt.Errorf("no messages in state")
		}

		// For simplicity, extract text from first message
		var query string
		if msg, ok := messages[0].(map[string]any); ok {
			if text, ok := msg["content"].(string); ok {
				query = text
			}
		}

		if query == "" {
			return state, fmt.Errorf("no query found")
		}

		// Use RAG to get context
		ragResult, err := ragEngine.Query(ctx, query)
		if err != nil {
			return state, fmt.Errorf("RAG query failed: %w", err)
		}

		// Generate response using RAG context
		response := fmt.Sprintf("Based on the retrieved information: %s\n\n"+
			"Answer: This demonstrates how RAG enhances generation with relevant context.",
			ragResult.Context)

		// Add response to messages
		responseMsg := map[string]any{
			"role":    "assistant",
			"content": response,
		}
		state["messages"] = append(messages, responseMsg)
		state["rag_context"] = ragResult.Context
		state["rag_sources"] = ragResult.Sources

		return state, nil
	})

	// Set up graph flow
	g.SetEntry("rag_enhanced_generation")
	g.AddEdge("rag_enhanced_generation", graph.END)

	// Compile and run
	runnable, err := g.Compile()
	if err != nil {
		return fmt.Errorf("failed to compile graph: %w", err)
	}

	// Run the RAG-enhanced agent
	inputState := map[string]any{
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": "What is Retrieval-Augmented Generation and how does it work?",
			},
		},
	}

	result, err := runnable.Invoke(ctx, inputState)
	if err != nil {
		return fmt.Errorf("failed to run RAG agent: %w", err)
	}

	fmt.Printf("RAG-enhanced agent response generated successfully\n")
	if messages, ok := result["messages"].([]map[string]any); ok && len(messages) > 1 {
		if response, ok := messages[len(messages)-1]["content"].(string); ok {
			fmt.Printf("Response: %s\n", response)
		}
	}

	return nil
}

// Mock implementations for demonstration purposes
// In practice, you would use real implementations

type MockVectorStore struct{}

func (m *MockVectorStore) Add(ctx context.Context, docs []rag.Document) error { return nil }
func (m *MockVectorStore) Search(ctx context.Context, query []float32, k int) ([]rag.DocumentSearchResult, error) {
	return []rag.DocumentSearchResult{
		{
			Document: rag.Document{
				ID:      "mock_doc_1",
				Content: "This is a mock document for demonstration purposes.",
				Metadata: map[string]any{
					"source": "mock",
				},
			},
			Score: 0.8,
		},
	}, nil
}
func (m *MockVectorStore) SearchWithFilter(ctx context.Context, query []float32, k int, filter map[string]any) ([]rag.DocumentSearchResult, error) {
	return m.Search(ctx, query, k)
}
func (m *MockVectorStore) Delete(ctx context.Context, ids []string) error { return nil }
func (m *MockVectorStore) Update(ctx context.Context, docs []rag.Document) error { return nil }
func (m *MockVectorStore) GetStats(ctx context.Context) (*rag.VectorStoreStats, error) {
	return &rag.VectorStoreStats{
		TotalDocuments: 1,
		TotalVectors:   1,
		Dimension:      1536,
		LastUpdated:    time.Now(),
	}, nil
}

type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, prompt string) (string, error) {
	return "This is a mock LLM response for demonstration purposes.", nil
}
func (m *MockLLM) GenerateWithSystem(ctx context.Context, system, prompt string) (string, error) {
	return m.Generate(ctx, prompt)
}

type MockEmbedder struct{}

func (m *MockEmbedder) EmbedDocument(ctx context.Context, text string) ([]float32, error) {
	// Return a mock embedding of dimension 1536
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = 0.1 // Mock value
	}
	return embedding, nil
}
func (m *MockEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := m.EmbedDocument(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}
func (m *MockEmbedder) GetDimension() int { return 1536 }