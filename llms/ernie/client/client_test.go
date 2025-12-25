package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestClientNew tests the Client creation with various options.
func TestClientNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "no api key",
			opts:    []Option{},
			wantErr: true,
		},
		{
			name: "with api key",
			opts: []Option{
				WithAPIKey("test-key"),
			},
			wantErr: false,
		},
		{
			name: "with api key and base url",
			opts: []Option{
				WithAPIKey("test-key"),
				WithBaseURL("https://custom.example.com"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("New() returned nil client")
			}
		})
	}
}

// TestClientHeaders tests that the correct headers are set.
func TestClientHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("Expected Authorization header to start with 'Bearer ', got: %s", auth)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got: %s", contentType)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"test","object":"list","data":[{"object":"embedding","embedding":[0.1,0.2,0.3],"index":0}],"model":"embedding-v1","usage":{"prompt_tokens":2,"total_tokens":2}}`))
	}))
	defer server.Close()

	client, err := New(WithAPIKey("test-key"), WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.CreateEmbedding(context.Background(), "test-model", []string{"Hello"})
	if err != nil {
		t.Fatalf("Failed to create embedding: %v", err)
	}
}

// TestClientCreateEmbedding_RealAPI tests embedding generation with real API.
// Skipped if QIANFAN_TOKEN is not set.
func TestClientCreateEmbedding_RealAPI(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	client, err := New(WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("single text", func(t *testing.T) {
		resp, err := client.CreateEmbedding(context.Background(), "embedding-v1", []string{"Hello world"})
		if err != nil {
			t.Fatalf("Failed to create embedding: %v", err)
		}
		if len(resp.Data) != 1 {
			t.Fatalf("Expected 1 embedding, got %d", len(resp.Data))
		}
		if len(resp.Data[0].Embedding) == 0 {
			t.Fatal("Empty embedding")
		}
		t.Logf("Embedding dimension: %d", len(resp.Data[0].Embedding))
	})

	t.Run("multiple texts", func(t *testing.T) {
		resp, err := client.CreateEmbedding(context.Background(), "embedding-v1", []string{"Hello", "World"})
		if err != nil {
			t.Fatalf("Failed to create embedding: %v", err)
		}
		if len(resp.Data) != 2 {
			t.Fatalf("Expected 2 embeddings, got %d", len(resp.Data))
		}
		for i, data := range resp.Data {
			if len(data.Embedding) == 0 {
				t.Errorf("Empty embedding at index %d", i)
			}
			t.Logf("Embedding %d dimension: %d", i, len(data.Embedding))
		}
	})
}

// TestClient_EmptyInput tests error handling for empty input.
func TestClient_EmptyInput(t *testing.T) {
	client, err := New(WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("CreateEmbedding with empty texts", func(t *testing.T) {
		_, err := client.CreateEmbedding(context.Background(), "embedding-v1", []string{})
		if err == nil {
			t.Error("Expected error for empty texts, got nil")
		}
	})
}
