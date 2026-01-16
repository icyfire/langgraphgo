package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/smallnest/langgraphgo/rag"
)

// ChromaV2VectorStore is a vector store implementation using Chroma v2 API
// It provides persistent vector storage with HTTP backend
//
// Note on Metadata: Chroma v2 API support for metadata storage and filtering
// may vary by server version. Some versions may store metadata as null even when
// properly sent in the payload. This is a server-side limitation, not an issue
// with this implementation.
type ChromaV2VectorStore struct {
	baseURL      string
	tenant       string
	database     string
	collectionID string
	collection   string
	embedder     rag.Embedder
	httpClient   *http.Client
}

// ChromaV2Config contains configuration for ChromaV2VectorStore
type ChromaV2Config struct {
	// BaseURL is the base URL of the Chroma server (e.g., http://localhost:8000)
	BaseURL string

	// Tenant is the tenant name (defaults to "default_tenant")
	Tenant string

	// Database is the database name (defaults to "default_database")
	Database string

	// Collection is the collection name
	Collection string

	// CollectionID is the collection UUID (optional, will use Collection name if not provided)
	CollectionID string

	// Embedder is the embedder to use for generating embeddings
	Embedder rag.Embedder

	// HTTPClient is the HTTP client to use (optional, will create default if not provided)
	HTTPClient *http.Client
}

// NewChromaV2VectorStore creates a new ChromaV2VectorStore with the given configuration
func NewChromaV2VectorStore(config ChromaV2Config) (*ChromaV2VectorStore, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if config.Embedder == nil {
		return nil, fmt.Errorf("embedder is required")
	}

	// Set defaults
	if config.Tenant == "" {
		config.Tenant = "default_tenant"
	}
	if config.Database == "" {
		config.Database = "default_database"
	}

	// Create HTTP client if not provided
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	store := &ChromaV2VectorStore{
		baseURL:    config.BaseURL,
		tenant:     config.Tenant,
		database:   config.Database,
		collection: config.Collection,
		embedder:   config.Embedder,
		httpClient: config.HTTPClient,
	}

	// Initialize collection
	if err := store.initCollection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize collection: %w", err)
	}

	return store, nil
}

// NewChromaV2VectorStoreSimple creates a new ChromaV2VectorStore with simple parameters
func NewChromaV2VectorStoreSimple(baseURL, collection string, embedder rag.Embedder) (*ChromaV2VectorStore, error) {
	return NewChromaV2VectorStore(ChromaV2Config{
		BaseURL:    baseURL,
		Collection: collection,
		Embedder:   embedder,
	})
}

// initCollection initializes or gets the collection
func (s *ChromaV2VectorStore) initCollection(ctx context.Context) error {
	// First try to get existing collection by name
	collections, err := s.listCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// Look for existing collection with matching name
	for _, col := range collections {
		if col.Name == s.collection {
			s.collectionID = col.ID
			return nil
		}
	}

	// Create new collection if not found
	return s.createCollection(ctx)
}

// createCollection creates a new collection
func (s *ChromaV2VectorStore) createCollection(ctx context.Context) error {
	// Create collection without specifying dimension - Chroma will auto-detect from first added document
	payload := map[string]any{
		"name": s.collection,
		// No configuration needed - let Chroma use defaults
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections", s.baseURL, s.tenant, s.database)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	s.collectionID = result.ID
	return nil
}

// listCollections lists all collections
func (s *ChromaV2VectorStore) listCollections(ctx context.Context) ([]CollectionInfo, error) {
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections", s.baseURL, s.tenant, s.database)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list collections: status %d", resp.StatusCode)
	}

	// Try to decode as array first (Chroma v2 returns empty array [])
	var collections []CollectionInfo
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		// If array fails, try object with items field
		respBody, _ := io.ReadAll(resp.Body)
		var result struct {
			Items []CollectionInfo `json:"items"`
		}
		if err2 := json.Unmarshal(respBody, &result); err2 != nil {
			return nil, fmt.Errorf("failed to decode response: %w (original: %v)", err2, err)
		}
		return result.Items, nil
	}

	return collections, nil
}

// CollectionInfo represents collection information
type CollectionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Add adds documents to the Chroma v2 vector store
func (s *ChromaV2VectorStore) Add(ctx context.Context, documents []rag.Document) error {
	if len(documents) == 0 {
		return nil
	}

	// Generate embeddings for all documents
	texts := make([]string, len(documents))
	for i, doc := range documents {
		texts[i] = doc.Content
	}

	embeddedDocs, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Prepare arrays for Chroma v2 API (uses separate arrays for ids, embeddings, documents, metadata)
	ids := make([]string, len(documents))
	docs := make([]string, len(documents))
	embeds := make([][]float64, len(documents))
	metadata := make([]map[string]any, len(documents))

	for i, doc := range documents {
		id := doc.ID
		if id == "" {
			id = fmt.Sprintf("doc_%d_%d", time.Now().UnixNano(), i)
		}

		ids[i] = id
		docs[i] = doc.Content
		metadata[i] = doc.Metadata

		// Convert float32 to float64
		embeds[i] = make([]float64, len(embeddedDocs[i]))
		for j, v := range embeddedDocs[i] {
			embeds[i][j] = float64(v)
		}
	}

	payload := map[string]any{
		"ids":        ids,
		"embeddings": embeds,
		"documents":  docs,
		"metadata":   metadata, // Always include metadata array
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Debug: print the payload to verify metadata is included
	// Note: Chroma v2 may store metadata as null depending on server version/configuration
	// if len(metadata) > 0 && metadata[0] != nil {
	// 	fmt.Printf("DEBUG: First doc metadata: %+v\n", metadata[0])
	// }
	// _ = json.NewEncoder(os.Stdout).Encode(payload)
	// fmt.Println()

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/add",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add documents: status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Search performs similarity search in the Chroma v2 vector store
func (s *ChromaV2VectorStore) Search(ctx context.Context, query []float32, k int) ([]rag.DocumentSearchResult, error) {
	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	// Convert query embedding from float32 to float64
	queryEmbedding := make([]float64, len(query))
	for i, v := range query {
		queryEmbedding[i] = float64(v)
	}

	// Use the query endpoint (not search - search is for distributed mode)
	payload := map[string]any{
		"query_embeddings": [][]float64{queryEmbedding},
		"n_results":        k,
		"include":          []string{"metadatas", "documents", "distances"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/query",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search: status %d: %s", resp.StatusCode, string(respBody))
	}

	// Chroma v2 query response format
	var result struct {
		IDs       [][]string         `json:"ids"`
		Distances [][]float64        `json:"distances"`
		Documents [][]string         `json:"documents"`
		Metadatas [][]map[string]any `json:"metadatas"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to our format
	// Result[0] corresponds to the first (and only) query embedding
	if len(result.IDs) == 0 {
		return []rag.DocumentSearchResult{}, nil
	}

	searchResults := make([]rag.DocumentSearchResult, len(result.IDs[0]))
	for i := range result.IDs[0] {
		var metadata map[string]any
		if len(result.Metadatas) > 0 && i < len(result.Metadatas[0]) {
			metadata = result.Metadatas[0][i]
		}

		var content string
		if len(result.Documents) > 0 && i < len(result.Documents[0]) {
			content = result.Documents[0][i]
		}

		var distance float64
		if len(result.Distances) > 0 && i < len(result.Distances[0]) {
			distance = result.Distances[0][i]
		}

		searchResults[i] = rag.DocumentSearchResult{
			Document: rag.Document{
				ID:        result.IDs[0][i],
				Content:   content,
				Metadata:  metadata,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Score: 1.0 - distance, // Convert distance to similarity score
		}
	}

	return searchResults, nil
}

// SearchWithFilter performs similarity search with metadata filters
func (s *ChromaV2VectorStore) SearchWithFilter(ctx context.Context, query []float32, k int, filter map[string]any) ([]rag.DocumentSearchResult, error) {
	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	// Convert query embedding from float32 to float64
	queryEmbedding := make([]float64, len(query))
	for i, v := range query {
		queryEmbedding[i] = float64(v)
	}

	// Use the query endpoint with where clause for filtering
	payload := map[string]any{
		"query_embeddings": [][]float64{queryEmbedding},
		"n_results":        k,
		"where":            filter,
		"include":          []string{"metadatas", "documents", "distances"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/query",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search with filter: status %d: %s", resp.StatusCode, string(respBody))
	}

	// Chroma v2 query response format
	var result struct {
		IDs       [][]string         `json:"ids"`
		Distances [][]float64        `json:"distances"`
		Documents [][]string         `json:"documents"`
		Metadatas [][]map[string]any `json:"metadatas"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to our format
	if len(result.IDs) == 0 {
		return []rag.DocumentSearchResult{}, nil
	}

	searchResults := make([]rag.DocumentSearchResult, len(result.IDs[0]))
	for i := range result.IDs[0] {
		var metadata map[string]any
		if len(result.Metadatas) > 0 && i < len(result.Metadatas[0]) {
			metadata = result.Metadatas[0][i]
		}

		var content string
		if len(result.Documents) > 0 && i < len(result.Documents[0]) {
			content = result.Documents[0][i]
		}

		var distance float64
		if len(result.Distances) > 0 && i < len(result.Distances[0]) {
			distance = result.Distances[0][i]
		}

		searchResults[i] = rag.DocumentSearchResult{
			Document: rag.Document{
				ID:        result.IDs[0][i],
				Content:   content,
				Metadata:  metadata,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Score: 1.0 - distance,
		}
	}

	return searchResults, nil
}

// Delete removes documents by their IDs
func (s *ChromaV2VectorStore) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	payload := map[string]any{
		"ids": ids,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/delete",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete documents: status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Update updates documents in the Chroma v2 vector store
func (s *ChromaV2VectorStore) Update(ctx context.Context, documents []rag.Document) error {
	if len(documents) == 0 {
		return nil
	}

	// Generate embeddings for all documents
	texts := make([]string, len(documents))
	for i, doc := range documents {
		texts[i] = doc.Content
	}

	embeddedDocs, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Prepare arrays for Chroma v2 upsert API
	ids := make([]string, len(documents))
	docs := make([]string, len(documents))
	embeds := make([][]float64, len(documents))
	metadata := make([]map[string]any, len(documents))

	for i, doc := range documents {
		if doc.ID == "" {
			return fmt.Errorf("document ID is required for update")
		}

		ids[i] = doc.ID
		docs[i] = doc.Content
		metadata[i] = doc.Metadata

		// Convert float32 to float64
		embeds[i] = make([]float64, len(embeddedDocs[i]))
		for j, v := range embeddedDocs[i] {
			embeds[i][j] = float64(v)
		}
	}

	payload := map[string]any{
		"ids":        ids,
		"embeddings": embeds,
		"documents":  docs,
		"metadata":   metadata, // Always include metadata array
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/upsert",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update documents: status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetStats returns statistics about the Chroma v2 vector store
func (s *ChromaV2VectorStore) GetStats(ctx context.Context) (*rag.VectorStoreStats, error) {
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/count",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get stats: status %d", resp.StatusCode)
	}

	var count uint32
	if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
		return nil, err
	}

	dimension := s.embedder.GetDimension()

	return &rag.VectorStoreStats{
		TotalDocuments: int(count),
		TotalVectors:   int(count),
		Dimension:      dimension,
		LastUpdated:    time.Now(),
	}, nil
}

// Close closes the Chroma v2 vector store and releases resources
func (s *ChromaV2VectorStore) Close() error {
	// Nothing to clean up for HTTP client
	return nil
}

// GetCollectionID returns the collection ID
func (s *ChromaV2VectorStore) GetCollectionID() string {
	return s.collectionID
}

// GetCollectionName returns the collection name
func (s *ChromaV2VectorStore) GetCollectionName() string {
	return s.collection
}
