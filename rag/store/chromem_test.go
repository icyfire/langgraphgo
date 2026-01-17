package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChromemVectorStore_InMemory tests the in-memory chromem vector store
func TestChromemVectorStore_InMemory(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewChromemVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Add and Search", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "1", Content: "hello", Embedding: []float32{1, 0, 0}},
			{ID: "2", Content: "world", Embedding: []float32{0, 1, 0}},
		}
		err := store.Add(ctx, docs)
		assert.NoError(t, err)

		// Search for something close to "hello"
		results, err := store.Search(ctx, []float32{1, 0.1, 0}, 1)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "1", results[0].Document.ID)
		assert.Greater(t, results[0].Score, 0.9)
	})

	t.Run("Search with Filter", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "3", Content: "filtered", Embedding: []float32{0, 0, 1}, Metadata: map[string]any{"type": "special"}},
		}
		err := store.Add(ctx, docs)
		assert.NoError(t, err)

		results, err := store.SearchWithFilter(ctx, []float32{0, 0, 1}, 10, map[string]any{"type": "special"})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "3", results[0].Document.ID)

		results, err = store.SearchWithFilter(ctx, []float32{0, 0, 1}, 10, map[string]any{"type": "none"})
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("Update and Delete", func(t *testing.T) {
		doc := rag.Document{ID: "1", Content: "updated", Embedding: []float32{1, 1, 1}}
		err := store.Update(ctx, []rag.Document{doc})
		assert.NoError(t, err)

		stats, _ := store.GetStats(ctx)
		countBefore := stats.TotalDocuments

		err = store.Delete(ctx, []string{"1"})
		assert.NoError(t, err)

		stats, _ = store.GetStats(ctx)
		assert.Equal(t, countBefore-1, stats.TotalDocuments)
	})

	t.Run("GetStats", func(t *testing.T) {
		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalDocuments, 0)
		assert.Equal(t, 3, stats.Dimension)
	})

	t.Run("Add without embedding", func(t *testing.T) {
		doc := rag.Document{ID: "4", Content: "no emb"}
		err := store.Add(ctx, []rag.Document{doc})
		assert.NoError(t, err)

		stats, _ := store.GetStats(ctx)
		assert.GreaterOrEqual(t, stats.TotalVectors, 1)
	})
}

// TestChromemVectorStore_Persistent tests the persistent chromem vector store
func TestChromemVectorStore_Persistent(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	// Create a temporary directory for the test
	tempDir := filepath.Join(os.TempDir(), "chromem-test")
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	store1, err := NewChromemVectorStoreSimple(tempDir, embedder)
	require.NoError(t, err)

	t.Run("Add documents and verify persistence", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "1", Content: "persistent doc 1", Embedding: []float32{1, 0, 0}},
			{ID: "2", Content: "persistent doc 2", Embedding: []float32{0, 1, 0}},
		}
		err := store1.Add(ctx, docs)
		assert.NoError(t, err)

		// Verify documents are in the store
		results, err := store1.Search(ctx, []float32{1, 0, 0}, 2)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})

	// Close the first store
	err = store1.Close()
	assert.NoError(t, err)

	// Open a new store with the same persistence directory
	store2, err := NewChromemVectorStoreSimple(tempDir, embedder)
	require.NoError(t, err)
	defer func() {
		_ = store2.Close()
	}()

	t.Run("Verify documents persist across store instances", func(t *testing.T) {
		results, err := store2.Search(ctx, []float32{1, 0, 0}, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		stats, err := store2.GetStats(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalDocuments, 2)
	})
}

// TestChromemVectorStore_WithConfig tests creating a store with full configuration
func TestChromemVectorStore_WithConfig(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 128}

	config := ChromemConfig{
		PersistenceDir: "", // In-memory
		CollectionName: "test_collection",
		Embedder:       embedder,
	}

	store, err := NewChromemVectorStore(config)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Verify collection name", func(t *testing.T) {
		assert.Equal(t, "test_collection", store.GetCollectionName())
	})

	t.Run("Add and search with different dimensions", func(t *testing.T) {
		// Create embedding with 128 dimensions
		embedding := make([]float32, 128)
		for i := range embedding {
			embedding[i] = 0.1
		}

		docs := []rag.Document{
			{ID: "1", Content: "test doc with 128 dims", Embedding: embedding},
		}
		err := store.Add(ctx, docs)
		assert.NoError(t, err)

		results, err := store.Search(ctx, embedding, 1)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "1", results[0].Document.ID)
	})
}

// TestChromemVectorStore_ConcurrentOperations tests concurrent operations
func TestChromemVectorStore_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewChromemVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Concurrent adds", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := range 10 {
			go func(idx int) {
				docs := []rag.Document{
					{ID: fmt.Sprintf("concurrent-%d", idx), Content: fmt.Sprintf("content %d", idx), Embedding: []float32{float32(idx) * 0.1, 0, 0}},
				}
				_ = store.Add(ctx, docs)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for range 10 {
			<-done
		}

		// Verify all documents were added
		stats, _ := store.GetStats(ctx)
		assert.GreaterOrEqual(t, stats.TotalDocuments, 10)
	})
}

// TestChromemVectorStore_EmbeddingGeneration tests automatic embedding generation
func TestChromemVectorStore_EmbeddingGeneration(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewChromemVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Add document without embedding", func(t *testing.T) {
		doc := rag.Document{
			ID:      "auto-embed",
			Content: "this should be auto-embedded",
		}
		err := store.Add(ctx, []rag.Document{doc})
		assert.NoError(t, err)

		// Search with a query embedding
		results, err := store.Search(ctx, []float32{0.1, 0.1, 0.1}, 1)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})
}

// TestChromemVectorStore_MetadataFiltering tests metadata filtering
func TestChromemVectorStore_MetadataFiltering(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewChromemVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Multiple metadata filters", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "1", Content: "doc 1", Embedding: []float32{1, 0, 0}, Metadata: map[string]any{"category": "tech", "year": "2023"}},
			{ID: "2", Content: "doc 2", Embedding: []float32{0, 1, 0}, Metadata: map[string]any{"category": "news", "year": "2023"}},
			{ID: "3", Content: "doc 3", Embedding: []float32{0, 0, 1}, Metadata: map[string]any{"category": "tech", "year": "2024"}},
		}
		err := store.Add(ctx, docs)
		assert.NoError(t, err)

		// Filter by category
		results, err := store.SearchWithFilter(ctx, []float32{1, 0, 0}, 10, map[string]any{"category": "tech"})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)

		// Filter by both category and year
		results, err = store.SearchWithFilter(ctx, []float32{1, 0, 0}, 10, map[string]any{"category": "tech", "year": "2023"})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "1", results[0].Document.ID)
	})
}

// TestChromemVectorStore_EdgeCases tests edge cases
func TestChromemVectorStore_EdgeCases(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewChromemVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Search with k=0", func(t *testing.T) {
		_, err := store.Search(ctx, []float32{1, 0, 0}, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "k must be positive")
	})

	t.Run("Search on empty store", func(t *testing.T) {
		emptyStore, err := NewChromemVectorStoreSimple("", embedder)
		require.NoError(t, err)
		defer func() {
			_ = emptyStore.Close()
		}()

		results, err := emptyStore.Search(ctx, []float32{1, 0, 0}, 5)
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("Delete non-existent documents", func(t *testing.T) {
		err := store.Delete(ctx, []string{"non-existent-id"})
		assert.NoError(t, err) // Should not error
	})

	t.Run("Update non-existent document", func(t *testing.T) {
		doc := rag.Document{ID: "non-existent", Content: "updated", Embedding: []float32{1, 1, 1}}
		err := store.Update(ctx, []rag.Document{doc})
		assert.NoError(t, err) // Should add the document
	})

	t.Run("Empty document list", func(t *testing.T) {
		err := store.Add(ctx, []rag.Document{})
		assert.NoError(t, err)

		err = store.Update(ctx, []rag.Document{})
		assert.NoError(t, err)

		err = store.Delete(ctx, []string{})
		assert.NoError(t, err)
	})
}

// TestConvertStringMapToAnyMap tests the helper function
func TestConvertStringMapToAnyMap(t *testing.T) {
	t.Run("Non-nil map", func(t *testing.T) {
		input := map[string]string{"key1": "value1", "key2": "value2"}
		result := convertStringMapToAnyMap(input)
		assert.Len(t, result, 2)
		assert.Equal(t, "value1", result["key1"])
		assert.Equal(t, "value2", result["key2"])
	})

	t.Run("Nil map", func(t *testing.T) {
		result := convertStringMapToAnyMap(nil)
		assert.Nil(t, result)
	})
}

// TestRuntimeNumWorkers tests the worker count calculation
func TestRuntimeNumWorkers(t *testing.T) {
	tests := []struct {
		numDocs     int
		expectedMin int
		expectedMax int
	}{
		{5, 1, 1},
		{10, 1, 2},
		{50, 2, 4},
		{100, 4, 4},
		{500, 4, 8},
		{1000, 8, 8},
		{5000, 8, 8},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("numDocs=%d", tt.numDocs), func(t *testing.T) {
			workers := runtimeNumWorkers(tt.numDocs)
			assert.GreaterOrEqual(t, workers, tt.expectedMin)
			assert.LessOrEqual(t, workers, tt.expectedMax)
		})
	}
}
