package rag

import (
	"context"
	"time"
)

// ========================================
// Core Types
// ========================================

// Document represents a document or document chunk in the RAG system
type Document struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata"`
	Embedding []float32      `json:"embedding,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Entity represents a knowledge graph entity
type Entity struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Properties map[string]any `json:"properties"`
	Embedding  []float32      `json:"embedding,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// Relationship represents a relationship between entities
type Relationship struct {
	ID         string         `json:"id"`
	Source     string         `json:"source"`
	Target     string         `json:"target"`
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Weight     float64        `json:"weight,omitempty"`
	Confidence float64        `json:"confidence,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// DocumentSearchResult represents a document search result with relevance score
type DocumentSearchResult struct {
	Document Document       `json:"document"`
	Score    float64        `json:"score"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// GraphQuery represents a query to the knowledge graph
type GraphQuery struct {
	EntityTypes   []string       `json:"entity_types,omitempty"`
	Relationships []string       `json:"relationships,omitempty"`
	Filters       map[string]any `json:"filters,omitempty"`
	Limit         int            `json:"limit,omitempty"`
	MaxDepth      int            `json:"max_depth,omitempty"`
	StartEntity   string         `json:"start_entity,omitempty"`
	EntityType    string         `json:"entity_type,omitempty"`
}

// GraphQueryResult represents the result of a graph query
type GraphQueryResult struct {
	Entities      []*Entity       `json:"entities"`
	Relationships []*Relationship `json:"relationships"`
	Paths         [][]*Entity     `json:"paths,omitempty"`
	Score         float64         `json:"score"`
	Scores        []float64       `json:"scores,omitempty"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
}

// ========================================
// Configuration Types
// ========================================

// RetrievalConfig contains configuration for retrieval operations
type RetrievalConfig struct {
	K              int            `json:"k"`
	ScoreThreshold float64        `json:"score_threshold"`
	SearchType     string         `json:"search_type"`
	Filter         map[string]any `json:"filter,omitempty"`
	IncludeScores  bool           `json:"include_scores"`
}

// VectorStoreStats contains statistics about a vector store
type VectorStoreStats struct {
	TotalDocuments int       `json:"total_documents"`
	TotalVectors   int       `json:"total_vectors"`
	Dimension      int       `json:"dimension"`
	LastUpdated    time.Time `json:"last_updated"`
}

// VectorRAGConfig represents configuration for vector-based RAG
type VectorRAGConfig struct {
	EmbeddingModel    string          `json:"embedding_model"`
	VectorStoreType   string          `json:"vector_store_type"`
	VectorStoreConfig map[string]any  `json:"vector_store_config"`
	ChunkSize         int             `json:"chunk_size"`
	ChunkOverlap      int             `json:"chunk_overlap"`
	EnableReranking   bool            `json:"enable_reranking"`
	RetrieverConfig   RetrievalConfig `json:"retriever_config"`
}

// GraphRAGConfig represents configuration for graph-based RAG
type GraphRAGConfig struct {
	DatabaseURL      string              `json:"database_url"`
	ModelProvider    string              `json:"model_provider"`
	EmbeddingModel   string              `json:"embedding_model"`
	ChatModel        string              `json:"chat_model"`
	EntityTypes      []string            `json:"entity_types"`
	Relationships    map[string][]string `json:"relationships"`
	MaxDepth         int                 `json:"max_depth"`
	EnableReasoning  bool                `json:"enable_reasoning"`
	ExtractionPrompt string              `json:"extraction_prompt"`
}

// Config is a generic RAG configuration
type Config struct {
	VectorRAG *VectorRAGConfig `json:"vector_rag,omitempty"`
	GraphRAG  *GraphRAGConfig  `json:"graph_rag,omitempty"`
}

// RAGConfig represents the main RAG configuration
type RAGConfig struct {
	Config        `json:"config"`
	EnableCache   bool          `json:"enable_cache"`
	CacheSize     int           `json:"cache_size"`
	EnableMetrics bool          `json:"enable_metrics"`
	Debug         bool          `json:"debug"`
	Timeout       time.Duration `json:"timeout"`
}

// ========================================
// Core Interfaces
// ========================================

// TextSplitter interface for splitting text into chunks
type TextSplitter interface {
	SplitText(text string) []string
	SplitDocuments(documents []Document) []Document
	JoinText(chunks []string) string
}

// Embedder interface for text embeddings
type Embedder interface {
	EmbedDocument(ctx context.Context, text string) ([]float32, error)
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	GetDimension() int
}

// VectorStore interface for vector storage and retrieval
type VectorStore interface {
	Add(ctx context.Context, documents []Document) error
	Search(ctx context.Context, query []float32, k int) ([]DocumentSearchResult, error)
	SearchWithFilter(ctx context.Context, query []float32, k int, filter map[string]any) ([]DocumentSearchResult, error)
	Delete(ctx context.Context, ids []string) error
	Update(ctx context.Context, documents []Document) error
	GetStats(ctx context.Context) (*VectorStoreStats, error)
}

// Retriever interface for document retrieval
type Retriever interface {
	Retrieve(ctx context.Context, query string) ([]Document, error)
	RetrieveWithK(ctx context.Context, query string, k int) ([]Document, error)
	RetrieveWithConfig(ctx context.Context, query string, config *RetrievalConfig) ([]DocumentSearchResult, error)
}

// Reranker interface for reranking search results
type Reranker interface {
	Rerank(ctx context.Context, query string, documents []DocumentSearchResult) ([]DocumentSearchResult, error)
}

// DocumentLoader interface for loading documents
type DocumentLoader interface {
	Load(ctx context.Context) ([]Document, error)
}

// LLMInterface defines the interface for language models
type LLMInterface interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateWithConfig(ctx context.Context, prompt string, config map[string]any) (string, error)
	GenerateWithSystem(ctx context.Context, system, prompt string) (string, error)
}

// KnowledgeGraph interface for graph-based retrieval
type KnowledgeGraph interface {
	AddEntity(ctx context.Context, entity *Entity) error
	AddRelationship(ctx context.Context, relationship *Relationship) error
	Query(ctx context.Context, query *GraphQuery) (*GraphQueryResult, error)
	GetRelatedEntities(ctx context.Context, entityID string, maxDepth int) ([]*Entity, error)
	GetEntity(ctx context.Context, entityID string) (*Entity, error)
}

// Engine interface for RAG engines
type Engine interface {
	Query(ctx context.Context, query string) (*QueryResult, error)
	QueryWithConfig(ctx context.Context, query string, config *RetrievalConfig) (*QueryResult, error)
	AddDocuments(ctx context.Context, docs []Document) error
	DeleteDocument(ctx context.Context, docID string) error
	UpdateDocument(ctx context.Context, doc Document) error
	SimilaritySearch(ctx context.Context, query string, k int) ([]Document, error)
	SimilaritySearchWithScores(ctx context.Context, query string, k int) ([]DocumentSearchResult, error)
}

// ========================================
// Result Types
// ========================================

// Metrics contains performance metrics for RAG engines
type Metrics struct {
	TotalQueries    int64         `json:"total_queries"`
	TotalDocuments  int64         `json:"total_documents"`
	AverageLatency  time.Duration `json:"average_latency"`
	MinLatency      time.Duration `json:"min_latency"`
	MaxLatency      time.Duration `json:"max_latency"`
	LastQueryTime   time.Time     `json:"last_query_time"`
	CacheHits       int64         `json:"cache_hits"`
	CacheMisses     int64         `json:"cache_misses"`
	IndexingLatency time.Duration `json:"indexing_latency"`
}

// QueryResult represents the result of a RAG query
type QueryResult struct {
	Query        string         `json:"query"`
	Answer       string         `json:"answer"`
	Sources      []Document     `json:"sources"`
	Context      string         `json:"context"`
	Confidence   float64        `json:"confidence"`
	ResponseTime time.Duration  `json:"response_time"`
	Metadata     map[string]any `json:"metadata"`
}
