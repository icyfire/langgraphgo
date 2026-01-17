package memu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/memory"
)

// TestNewClient tests the NewClient function
func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				BaseURL: "https://api.memu.so",
				APIKey:  "test-key",
				UserID:  "user-123",
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			cfg: Config{
				APIKey: "test-key",
				UserID: "user-123",
			},
			wantErr: true,
		},
		{
			name: "missing API key",
			cfg: Config{
				BaseURL: "https://api.memu.so",
				UserID:  "user-123",
			},
			wantErr: true,
		},
		{
			name: "missing user ID",
			cfg: Config{
				BaseURL: "https://api.memu.so",
				APIKey:  "test-key",
			},
			wantErr: true,
		},
		{
			name: "custom HTTP client",
			cfg: Config{
				BaseURL:    "https://api.memu.so",
				APIKey:     "test-key",
				UserID:     "user-123",
				HTTPClient: &http.Client{Timeout: 60 * time.Second},
			},
			wantErr: false,
		},
		{
			name: "default retrieve method",
			cfg: Config{
				BaseURL: "https://api.memu.so",
				APIKey:  "test-key",
				UserID:  "user-123",
			},
			wantErr: false,
		},
		{
			name: "custom retrieve method llm",
			cfg: Config{
				BaseURL:        "https://api.memu.so",
				APIKey:         "test-key",
				UserID:         "user-123",
				RetrieveMethod: "llm",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Errorf("NewClient() returned nil client")
				return
			}

			if client.baseURL != tt.cfg.BaseURL {
				t.Errorf("baseURL mismatch: got %s, want %s", client.baseURL, tt.cfg.BaseURL)
			}

			if client.apiKey != tt.cfg.APIKey {
				t.Errorf("apiKey mismatch: got %s, want %s", client.apiKey, tt.cfg.APIKey)
			}

			if client.userID != tt.cfg.UserID {
				t.Errorf("userID mismatch: got %s, want %s", client.userID, tt.cfg.UserID)
			}

			// Check default retrieve method
			if tt.cfg.RetrieveMethod == "" && client.retrieveMethod != "rag" {
				t.Errorf("default retrieve method should be 'rag', got %s", client.retrieveMethod)
			}

			// Check custom retrieve method
			if tt.cfg.RetrieveMethod != "" && client.retrieveMethod != tt.cfg.RetrieveMethod {
				t.Errorf("retrieve method mismatch: got %s, want %s", client.retrieveMethod, tt.cfg.RetrieveMethod)
			}

			// Check custom HTTP client
			if tt.cfg.HTTPClient != nil && client.httpClient != tt.cfg.HTTPClient {
				t.Errorf("httpClient mismatch")
			}
		})
	}
}

// TestAddMessage tests the AddMessage method with a mock server
func TestAddMessage(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v3/memory/memorize" {
			t.Errorf("expected path /api/v3/memory/memorize, got %s", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("expected Authorization header 'Bearer test-key', got %s", auth)
		}

		// Send mock response
		response := MemorizeResponse{
			TaskID:  "task-123",
			Status:  "completed",
			Message: "Memory stored successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		UserID:  "user-123",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test AddMessage
	msg := &memory.Message{
		ID:      "msg-123",
		Role:    "user",
		Content: "Hello, I prefer coffee over tea",
	}

	ctx := context.Background()
	if err := client.AddMessage(ctx, msg); err != nil {
		t.Errorf("AddMessage() failed: %v", err)
	}
}

// TestGetContext tests the GetContext method with a mock server
func TestGetContext(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v3/memory/retrieve" {
			t.Errorf("expected path /api/v3/memory/retrieve, got %s", r.URL.Path)
		}

		// Send mock response
		response := RetrieveResponse{
			Categories: []Category{
				{
					ID:          "cat-1",
					Name:        "preferences",
					Description: "User preferences",
					Summary:     "The user prefers coffee over tea and enjoys working in the morning",
					ItemIDs:     []string{"item-1", "item-2"},
				},
			},
			Items: []Item{
				{
					ID:         "item-1",
					Type:       "preference",
					Summary:    "Prefers coffee over tea",
					Importance: 0.9,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		UserID:  "user-123",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test GetContext
	ctx := context.Background()
	messages, err := client.GetContext(ctx, "What are my preferences?")
	if err != nil {
		t.Errorf("GetContext() failed: %v", err)
		return
	}

	if len(messages) == 0 {
		t.Errorf("GetContext() returned no messages")
		return
	}

	// Verify we got both category and item messages
	foundCategory := false
	foundItem := false
	for _, msg := range messages {
		if msg.Metadata["source"] == "memu_category" {
			foundCategory = true
			if msg.Role != "system" {
				t.Errorf("category message should have role 'system', got %s", msg.Role)
			}
		}
		if msg.Metadata["source"] == "memu_item" {
			foundItem = true
		}
	}

	if !foundCategory {
		t.Errorf("expected to find category message")
	}

	if !foundItem {
		t.Errorf("expected to find item message")
	}
}

// TestGetStats tests the GetStats method with a mock server
func TestGetStats(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v3/memory/categories" {
			t.Errorf("expected path /api/v3/memory/categories, got %s", r.URL.Path)
		}

		// Send mock response
		response := CategoriesResponse{
			Categories: []Category{
				{
					ID:      "cat-1",
					Name:    "preferences",
					Summary: "User preferences and habits",
					ItemIDs: []string{"item-1", "item-2", "item-3"},
				},
				{
					ID:      "cat-2",
					Name:    "work",
					Summary: "Work-related information",
					ItemIDs: []string{"item-4"},
				},
			},
			Total: 2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		UserID:  "user-123",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test GetStats
	ctx := context.Background()
	stats, err := client.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats() failed: %v", err)
		return
	}

	if stats == nil {
		t.Errorf("GetStats() returned nil stats")
		return
	}

	if stats.ActiveMessages != 2 {
		t.Errorf("expected 2 active messages (categories), got %d", stats.ActiveMessages)
	}

	if stats.TotalMessages != 4 {
		t.Errorf("expected 4 total messages (items), got %d", stats.TotalMessages)
	}
}

// TestClear tests the Clear method
func TestClear(t *testing.T) {
	client, err := NewClient(Config{
		BaseURL: "https://api.memu.so",
		APIKey:  "test-key",
		UserID:  "user-123",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Clear(ctx)
	if err == nil {
		t.Errorf("Clear() should return an error (not supported)")
	}
}

// TestDoRequestErrorHandling tests error handling in doRequest
func TestDoRequestErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedErr    string
		expectResponse bool
	}{
		{
			name: "HTTP error response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
			},
			expectedErr: "API request failed with status 500",
		},
		{
			name: "HTTP 401 unauthorized",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("unauthorized"))
			},
			expectedErr: "API request failed with status 401",
		},
		{
			name: "successful response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			},
			expectResponse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, err := NewClient(Config{
				BaseURL: server.URL,
				APIKey:  "test-key",
				UserID:  "user-123",
			})
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			ctx := context.Background()
			var result map[string]string

			err = client.doRequest(ctx, "GET", "/test", nil, &result)

			if tt.expectResponse {
				if err != nil {
					t.Errorf("doRequest() unexpected error: %v", err)
				}
				if result["status"] != "ok" {
					t.Errorf("expected status 'ok', got %s", result["status"])
				}
			} else if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("doRequest() expected error containing %q, got nil", tt.expectedErr)
				}
			}
		})
	}
}
