package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/philippgille/chromem-go"
	"github.com/smallnest/langgraphgo/rag"
)

// ChromemVectorStore is a vector store implementation using chromem-go
// It provides persistent vector storage with SQLite backend
type ChromemVectorStore struct {
	db             *chromem.DB
	collection     *chromem.Collection
	embedder       rag.Embedder
	collectionName string
}

// ChromemConfig contains configuration for ChromemVectorStore
type ChromemConfig struct {
	// PersistenceDir is the directory to store the chromem database
	// If empty, uses in-memory storage
	PersistenceDir string

	// CollectionName is the name of the collection to use
	// If empty, uses "default"
	CollectionName string

	// Embedder is the embedder to use for generating embeddings
	Embedder rag.Embedder
}

// NewChromemVectorStore creates a new ChromemVectorStore with the given configuration
func NewChromemVectorStore(config ChromemConfig) (*ChromemVectorStore, error) {
	var db *chromem.DB
	var err error

	if config.PersistenceDir != "" {
		// Ensure the directory exists
		if err := os.MkdirAll(config.PersistenceDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create persistence directory: %w", err)
		}
		dbPath := filepath.Join(config.PersistenceDir, "chromem.db")
		db, err = chromem.NewPersistentDB(dbPath, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create chromem db: %w", err)
		}
	} else {
		db = chromem.NewDB()
	}

	if config.Embedder == nil {
		return nil, fmt.Errorf("embedder is required")
	}

	collectionName := config.CollectionName
	if collectionName == "" {
		collectionName = "default"
	}

	// Create an embedding function from the embedder
	embeddingFunc := func(ctx context.Context, text string) ([]float32, error) {
		return config.Embedder.EmbedDocument(ctx, text)
	}

	// Try to get existing collection, or create a new one
	collection := db.GetCollection(collectionName, embeddingFunc)
	if collection == nil {
		// Collection doesn't exist, create it
		collection, err = db.CreateCollection(collectionName, nil, embeddingFunc)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}
	}

	return &ChromemVectorStore{
		db:             db,
		collection:     collection,
		embedder:       config.Embedder,
		collectionName: collectionName,
	}, nil
}

// NewChromemVectorStoreSimple creates a new ChromemVectorStore with simple parameters
// For in-memory storage, pass an empty string for persistenceDir
func NewChromemVectorStoreSimple(persistenceDir string, embedder rag.Embedder) (*ChromemVectorStore, error) {
	return NewChromemVectorStore(ChromemConfig{
		PersistenceDir: persistenceDir,
		Embedder:       embedder,
	})
}

// Add adds documents to the chromem vector store
func (s *ChromemVectorStore) Add(ctx context.Context, documents []rag.Document) error {
	if len(documents) == 0 {
		return nil
	}

	// Create embedding function for chromem
	embeddingFunc := func(ctx context.Context, text string) ([]float32, error) {
		return s.embedder.EmbedDocument(ctx, text)
	}

	// Prepare documents for chromem
	chromemDocs := make([]chromem.Document, len(documents))
	for i, doc := range documents {
		// Convert metadata to the format expected by chromem
		metadata := make(map[string]string)
		for k, v := range doc.Metadata {
			metadata[k] = fmt.Sprint(v)
		}

		// If document has embedding, use it; otherwise nil and chromem will generate it
		var embedding []float32
		if len(doc.Embedding) > 0 {
			embedding = doc.Embedding
		}

		chromemDoc, err := chromem.NewDocument(ctx, doc.ID, metadata, embedding, doc.Content, embeddingFunc)
		if err != nil {
			return fmt.Errorf("failed to create chromem document for %s: %w", doc.ID, err)
		}
		chromemDocs[i] = chromemDoc
	}

	// Add documents to the collection in a batch
	return s.collection.AddDocuments(ctx, chromemDocs, runtimeNumWorkers(len(documents)))
}

// Search performs similarity search in the chromem vector store
func (s *ChromemVectorStore) Search(ctx context.Context, query []float32, k int) ([]rag.DocumentSearchResult, error) {
	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	// Get the count to limit k appropriately
	count := s.collection.Count()
	if k > count {
		k = count
	}

	// If no documents, return empty results
	if k == 0 {
		return []rag.DocumentSearchResult{}, nil
	}

	// Perform the search with embedding
	results, err := s.collection.QueryEmbedding(ctx, query, k, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query collection: %w", err)
	}

	// Convert results to our format
	searchResults := make([]rag.DocumentSearchResult, len(results))
	for i, result := range results {
		searchResults[i] = rag.DocumentSearchResult{
			Document: rag.Document{
				ID:        result.ID,
				Content:   result.Content,
				Metadata:  convertStringMapToAnyMap(result.Metadata),
				CreatedAt: time.Now(), // chromem doesn't store creation time
				UpdatedAt: time.Now(),
			},
			Score: float64(result.Similarity),
		}
	}

	return searchResults, nil
}

// SearchWithFilter performs similarity search with metadata filters
func (s *ChromemVectorStore) SearchWithFilter(ctx context.Context, query []float32, k int, filter map[string]any) ([]rag.DocumentSearchResult, error) {
	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	// Convert filter to chromem format (string map)
	where := make(map[string]string)
	for k, v := range filter {
		where[k] = fmt.Sprint(v)
	}

	// Get the count to limit k appropriately
	count := s.collection.Count()
	if k > count {
		k = count
	}

	// If no documents, return empty results
	if k == 0 {
		return []rag.DocumentSearchResult{}, nil
	}

	// Perform the search with filters
	results, err := s.collection.QueryEmbedding(ctx, query, k, where, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query collection with filter: %w", err)
	}

	// Convert results to our format
	searchResults := make([]rag.DocumentSearchResult, len(results))
	for i, result := range results {
		searchResults[i] = rag.DocumentSearchResult{
			Document: rag.Document{
				ID:        result.ID,
				Content:   result.Content,
				Metadata:  convertStringMapToAnyMap(result.Metadata),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Score: float64(result.Similarity),
		}
	}

	return searchResults, nil
}

// Delete removes documents by their IDs
func (s *ChromemVectorStore) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return s.collection.Delete(ctx, nil, nil, ids...)
}

// Update updates documents in the chromem vector store
func (s *ChromemVectorStore) Update(ctx context.Context, documents []rag.Document) error {
	if len(documents) == 0 {
		return nil
	}

	// For each document, we need to delete the old one and add the new one
	// This is because chromem doesn't have a native update operation
	idsToDelete := make([]string, len(documents))
	for i, doc := range documents {
		idsToDelete[i] = doc.ID
	}

	// Delete old documents (ignore errors if documents don't exist)
	_ = s.collection.Delete(ctx, nil, nil, idsToDelete...)

	// Add updated documents
	return s.Add(ctx, documents)
}

// GetStats returns statistics about the chromem vector store
func (s *ChromemVectorStore) GetStats(ctx context.Context) (*rag.VectorStoreStats, error) {
	count := s.collection.Count()

	dimension := s.embedder.GetDimension()

	return &rag.VectorStoreStats{
		TotalDocuments: count,
		TotalVectors:   count,
		Dimension:      dimension,
		LastUpdated:    time.Now(),
	}, nil
}

// Close closes the chromem vector store and releases resources
func (s *ChromemVectorStore) Close() error {
	// chromem-go doesn't require explicit cleanup for in-memory DB
	// For persistent DB, it will be properly closed when the DB object is garbage collected
	// or we can rely on Go's finalizer
	return nil
}

// GetCollectionName returns the name of the collection
func (s *ChromemVectorStore) GetCollectionName() string {
	return s.collectionName
}

// convertStringMapToAnyMap converts a map[string]string to map[string]any
func convertStringMapToAnyMap(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// runtimeNumWorkers returns the number of workers to use for parallel operations
// based on the number of documents
func runtimeNumWorkers(numDocuments int) int {
	if numDocuments < 10 {
		return 1
	}
	if numDocuments < 100 {
		return 2
	}
	if numDocuments < 1000 {
		return 4
	}
	return 8
}
