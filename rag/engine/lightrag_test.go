package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/store"
)

// MockLLM implements the LLMInterface for testing
type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, prompt string) (string, error) {
	return `{
		"entities": [
			{
				"id": "test_entity_1",
				"name": "Test Entity",
				"type": "TEST",
				"description": "A test entity",
				"properties": {}
			}
		]
	}`, nil
}

func (m *MockLLM) GenerateWithConfig(ctx context.Context, prompt string, config map[string]any) (string, error) {
	return m.Generate(ctx, prompt)
}

func (m *MockLLM) GenerateWithSystem(ctx context.Context, system, prompt string) (string, error) {
	return m.Generate(ctx, prompt)
}

func TestNewLightRAGEngine(t *testing.T) {
	_ = context.Background()

	config := rag.LightRAGConfig{
		Mode:                      "hybrid",
		ChunkSize:                 512,
		ChunkOverlap:              50,
		MaxEntitiesPerChunk:       20,
		EntityExtractionThreshold: 0.5,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		t.Fatalf("Failed to create knowledge graph: %v", err)
	}

	// Test without vector store
	engine, err := NewLightRAGEngine(config, llm, embedder, kg, nil)
	if err != nil {
		t.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}

	// Test default values
	if engine.config.Mode != "hybrid" {
		t.Errorf("Expected mode 'hybrid', got '%s'", engine.config.Mode)
	}

	if engine.config.ChunkSize != 512 {
		t.Errorf("Expected chunk size 512, got %d", engine.config.ChunkSize)
	}
}

func TestLightRAGEngine_NaiveRetrieval(t *testing.T) {
	ctx := context.Background()

	config := rag.LightRAGConfig{
		Mode:         "naive",
		ChunkSize:    512,
		ChunkOverlap: 50,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	vectorStore := store.NewInMemoryVectorStore(embedder)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		t.Fatalf("Failed to create knowledge graph: %v", err)
	}

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, vectorStore)
	if err != nil {
		t.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	// Add test documents
	docs := []rag.Document{
		{
			ID:      "doc1",
			Content: "This is a test document about artificial intelligence and machine learning.",
			Metadata: map[string]any{
				"source": "test.txt",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:      "doc2",
			Content: "This document discusses neural networks and deep learning algorithms.",
			Metadata: map[string]any{
				"source": "test2.txt",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err = engine.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Test naive retrieval
	result, err := engine.Query(ctx, "artificial intelligence")
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Query != "artificial intelligence" {
		t.Errorf("Expected query 'artificial intelligence', got '%s'", result.Query)
	}

	if len(result.Sources) == 0 {
		t.Error("Expected at least one source")
	}

	if result.Metadata["mode"] != "naive" {
		t.Errorf("Expected mode 'naive', got '%v'", result.Metadata["mode"])
	}
}

func TestLightRAGEngine_LocalRetrieval(t *testing.T) {
	ctx := context.Background()

	config := rag.LightRAGConfig{
		Mode: "local",
		LocalConfig: rag.LocalRetrievalConfig{
			TopK:                10,
			MaxHops:             2,
			IncludeDescriptions: true,
		},
		ChunkSize:    512,
		ChunkOverlap: 50,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		t.Fatalf("Failed to create knowledge graph: %v", err)
	}

	vectorStore := store.NewInMemoryVectorStore(embedder)

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, vectorStore)
	if err != nil {
		t.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	// Add test documents
	docs := []rag.Document{
		{
			ID:      "doc1",
			Content: "Elon Musk is the CEO of Tesla and SpaceX. He is known for his work in electric vehicles and space exploration.",
			Metadata: map[string]any{
				"source": "biography.txt",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err = engine.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Test local retrieval
	result, err := engine.Query(ctx, "Who is Elon Musk?")
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Metadata["mode"] != "local" {
		t.Errorf("Expected mode 'local', got '%v'", result.Metadata["mode"])
	}
}

func TestLightRAGEngine_GlobalRetrieval(t *testing.T) {
	ctx := context.Background()

	config := rag.LightRAGConfig{
		Mode: "global",
		GlobalConfig: rag.GlobalRetrievalConfig{
			MaxCommunities:    5,
			IncludeHierarchy:  false,
			MaxHierarchyDepth: 3,
		},
		EnableCommunityDetection: true,
		ChunkSize:                512,
		ChunkOverlap:             50,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		t.Fatalf("Failed to create knowledge graph: %v", err)
	}

	vectorStore := store.NewInMemoryVectorStore(embedder)

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, vectorStore)
	if err != nil {
		t.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	// Add test documents
	docs := []rag.Document{
		{
			ID:      "doc1",
			Content: "Python is a programming language used for web development, data science, and machine learning.",
			Metadata: map[string]any{
				"source": "python.txt",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err = engine.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Test global retrieval
	result, err := engine.Query(ctx, "What is Python used for?")
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Metadata["mode"] != "global" {
		t.Errorf("Expected mode 'global', got '%v'", result.Metadata["mode"])
	}
}

func TestLightRAGEngine_HybridRetrieval(t *testing.T) {
	ctx := context.Background()

	config := rag.LightRAGConfig{
		Mode: "hybrid",
		HybridConfig: rag.HybridRetrievalConfig{
			LocalWeight:  0.5,
			GlobalWeight: 0.5,
			FusionMethod: "rrf",
			RFFK:         60,
		},
		LocalConfig: rag.LocalRetrievalConfig{
			TopK:    10,
			MaxHops: 2,
		},
		GlobalConfig: rag.GlobalRetrievalConfig{
			MaxCommunities: 5,
		},
		ChunkSize:                512,
		ChunkOverlap:             50,
		EnableCommunityDetection: true,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		t.Fatalf("Failed to create knowledge graph: %v", err)
	}

	vectorStore := store.NewInMemoryVectorStore(embedder)

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, vectorStore)
	if err != nil {
		t.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	// Add test documents
	docs := []rag.Document{
		{
			ID:      "doc1",
			Content: "Go is a statically typed, compiled programming language designed at Google.",
			Metadata: map[string]any{
				"source": "go.txt",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err = engine.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Test hybrid retrieval
	result, err := engine.Query(ctx, "Tell me about Go programming language")
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Metadata["mode"] != "hybrid" {
		t.Errorf("Expected mode 'hybrid', got '%v'", result.Metadata["mode"])
	}

	if result.Metadata["fusion_method"] != "rrf" {
		t.Errorf("Expected fusion method 'rrf', got '%v'", result.Metadata["fusion_method"])
	}
}

func TestLightRAGEngine_SplitDocument(t *testing.T) {
	_ = context.Background()

	config := rag.LightRAGConfig{
		Mode:         "hybrid",
		ChunkSize:    100,
		ChunkOverlap: 20,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		t.Fatalf("Failed to create knowledge graph: %v", err)
	}

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, nil)
	if err != nil {
		t.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	// Create a document that will be split into multiple chunks
	doc := rag.Document{
		ID:        "test_doc",
		Content:   strings.Repeat("This is a test sentence. ", 20),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	chunks := engine.splitDocument(doc)

	if len(chunks) < 2 {
		t.Errorf("Expected at least 2 chunks, got %d", len(chunks))
	}

	// Verify chunk IDs are unique
	ids := make(map[string]bool)
	for _, chunk := range chunks {
		if ids[chunk.ID] {
			t.Errorf("Duplicate chunk ID: %s", chunk.ID)
		}
		ids[chunk.ID] = true

		// Verify metadata
		if chunk.Metadata["source_doc"] != "test_doc" {
			t.Errorf("Expected source_doc 'test_doc', got '%v'", chunk.Metadata["source_doc"])
		}
	}
}

func TestLightRAGEngine_SimilaritySearch(t *testing.T) {
	ctx := context.Background()

	config := rag.LightRAGConfig{
		Mode:         "naive",
		ChunkSize:    512,
		ChunkOverlap: 50,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	vectorStore := store.NewInMemoryVectorStore(embedder)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		t.Fatalf("Failed to create knowledge graph: %v", err)
	}

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, vectorStore)
	if err != nil {
		t.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	docs := []rag.Document{
		{
			ID:        "doc1",
			Content:   "Test document one",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err = engine.AddDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	// Test similarity search
	results, err := engine.SimilaritySearch(ctx, "test", 5)
	if err != nil {
		t.Fatalf("Failed to perform similarity search: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}

	// Test similarity search with scores
	scoreResults, err := engine.SimilaritySearchWithScores(ctx, "test", 5)
	if err != nil {
		t.Fatalf("Failed to perform similarity search with scores: %v", err)
	}

	if len(scoreResults) == 0 {
		t.Error("Expected at least one scored result")
	}
}

func TestLightRAGEngine_GetMetrics(t *testing.T) {
	_ = context.Background()

	config := rag.LightRAGConfig{
		Mode: "hybrid",
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		t.Fatalf("Failed to create knowledge graph: %v", err)
	}

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, nil)
	if err != nil {
		t.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	metrics := engine.GetMetrics()

	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	// Initial metrics should be zero
	if metrics.TotalQueries != 0 {
		t.Errorf("Expected 0 total queries, got %d", metrics.TotalQueries)
	}

	if metrics.TotalDocuments != 0 {
		t.Errorf("Expected 0 total documents, got %d", metrics.TotalDocuments)
	}
}

func TestLightRAGEngine_GetConfig(t *testing.T) {
	_ = context.Background()

	config := rag.LightRAGConfig{
		Mode:                      "hybrid",
		ChunkSize:                 1024,
		ChunkOverlap:              100,
		MaxEntitiesPerChunk:       30,
		EntityExtractionThreshold: 0.7,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		t.Fatalf("Failed to create knowledge graph: %v", err)
	}

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, nil)
	if err != nil {
		t.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	retrievedConfig := engine.GetConfig()

	if retrievedConfig.Mode != config.Mode {
		t.Errorf("Expected mode '%s', got '%s'", config.Mode, retrievedConfig.Mode)
	}

	if retrievedConfig.ChunkSize != config.ChunkSize {
		t.Errorf("Expected chunk size %d, got %d", config.ChunkSize, retrievedConfig.ChunkSize)
	}
}

// Benchmark tests
func BenchmarkLightRAGEngine_AddDocuments(b *testing.B) {
	ctx := context.Background()

	config := rag.LightRAGConfig{
		Mode:         "hybrid",
		ChunkSize:    512,
		ChunkOverlap: 50,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	vectorStore := store.NewInMemoryVectorStore(embedder)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		b.Fatalf("Failed to create knowledge graph: %v", err)
	}

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, vectorStore)
	if err != nil {
		b.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	docs := make([]rag.Document, 100)
	for i := range 100 {
		docs[i] = rag.Document{
			ID:        fmt.Sprintf("doc%d", i),
			Content:   strings.Repeat("Test content ", 100),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	for b.Loop() {
		_ = engine.AddDocuments(ctx, docs)
	}
}

func BenchmarkLightRAGEngine_Query(b *testing.B) {
	ctx := context.Background()

	config := rag.LightRAGConfig{
		Mode:         "hybrid",
		ChunkSize:    512,
		ChunkOverlap: 50,
	}

	llm := &MockLLM{}
	embedder := store.NewMockEmbedder(128)
	vectorStore := store.NewInMemoryVectorStore(embedder)
	kg, err := store.NewKnowledgeGraph("memory://")
	if err != nil {
		b.Fatalf("Failed to create knowledge graph: %v", err)
	}

	engine, err := NewLightRAGEngine(config, llm, embedder, kg, vectorStore)
	if err != nil {
		b.Fatalf("Failed to create LightRAG engine: %v", err)
	}

	docs := make([]rag.Document, 100)
	for i := range 100 {
		docs[i] = rag.Document{
			ID:        fmt.Sprintf("doc%d", i),
			Content:   strings.Repeat("Test content ", 100),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	_ = engine.AddDocuments(ctx, docs)

	for b.Loop() {
		_, _ = engine.Query(ctx, "test query")
	}
}
