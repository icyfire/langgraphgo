package qwen

import (
	"context"
	"os"
	"testing"
)

func TestEmbedder(t *testing.T) {
	t.Run("NewEmbedder creates valid embedder", func(t *testing.T) {
		embedder := NewEmbedder("https://api.example.com", "test-key", "test-model")
		if embedder == nil {
			t.Fatal("NewEmbedder returned nil")
		}
		if embedder.baseURL != "https://api.example.com" {
			t.Errorf("expected baseURL https://api.example.com, got %s", embedder.baseURL)
		}
		if embedder.apiKey != "test-key" {
			t.Errorf("expected apiKey test-key, got %s", embedder.apiKey)
		}
		if embedder.model != "test-model" {
			t.Errorf("expected model test-model, got %s", embedder.model)
		}
	})

	t.Run("NewEmbedderWithOptions creates valid embedder", func(t *testing.T) {
		embedder := NewEmbedderWithOptions(
			WithBaseURL("https://api.example.com"),
			WithAPIKey("test-key"),
			WithModel("test-model"),
		)
		if embedder == nil {
			t.Fatal("NewEmbedderWithOptions returned nil")
		}
		if embedder.baseURL != "https://api.example.com" {
			t.Errorf("expected baseURL https://api.example.com, got %s", embedder.baseURL)
		}
		if embedder.apiKey != "test-key" {
			t.Errorf("expected apiKey test-key, got %s", embedder.apiKey)
		}
		if embedder.model != "test-model" {
			t.Errorf("expected model test-model, got %s", embedder.model)
		}
	})

	t.Run("NewEmbedderWithOptions uses defaults", func(t *testing.T) {
		embedder := NewEmbedderWithOptions()
		if embedder.baseURL != "https://api-inference.modelscope.cn/v1" {
			t.Errorf("expected default baseURL, got %s", embedder.baseURL)
		}
		if embedder.model != "Qwen/Qwen3-Embedding-4B" {
			t.Errorf("expected default model, got %s", embedder.model)
		}
	})

	t.Run("GetDimension returns correct dimension", func(t *testing.T) {
		embedder := NewEmbedder("https://api.example.com", "test-key", "test-model")
		if dim := embedder.GetDimension(); dim != 2560 {
			t.Errorf("expected dimension 2560, got %d", dim)
		}
	})

	t.Run("Dimension returns correct dimension", func(t *testing.T) {
		embedder := NewEmbedder("https://api.example.com", "test-key", "test-model")
		if dim := embedder.Dimension(); dim != 2560 {
			t.Errorf("expected dimension 2560, got %d", dim)
		}
	})

	t.Run("EmbedDocuments with empty input", func(t *testing.T) {
		embedder := NewEmbedder("https://api.example.com", "test-key", "test-model")
		result, err := embedder.EmbedDocuments(context.Background(), []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}

// Integration test - run with MODELSCOPE_API_KEY set
func TestEmbedderIntegration(t *testing.T) {
	apiKey := os.Getenv("MODELSCOPE_API_KEY")
	if apiKey == "" {
		t.Skip("MODELSCOPE_API_KEY not set, skipping integration test")
	}

	embedder := NewEmbedder(
		"https://api-inference.modelscope.cn/v1",
		apiKey,
		"Qwen/Qwen3-Embedding-4B",
	)

	t.Run("EmbedDocument", func(t *testing.T) {
		ctx := context.Background()
		embedding, err := embedder.EmbedDocument(ctx, "Hello, world!")
		if err != nil {
			t.Fatalf("EmbedDocument failed: %v", err)
		}
		if len(embedding) != 2560 {
			t.Errorf("expected embedding length 2560, got %d", len(embedding))
		}
	})

	t.Run("EmbedDocuments", func(t *testing.T) {
		ctx := context.Background()
		embeddings, err := embedder.EmbedDocuments(ctx, []string{"text1", "text2"})
		if err != nil {
			t.Fatalf("EmbedDocuments failed: %v", err)
		}
		if len(embeddings) != 2 {
			t.Errorf("expected 2 embeddings, got %d", len(embeddings))
		}
		for i, emb := range embeddings {
			if len(emb) != 2560 {
				t.Errorf("embedding %d: expected length 2560, got %d", i, len(emb))
			}
		}
	})
}
