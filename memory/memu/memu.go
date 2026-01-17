// Package memu provides a Go client for memU (https://github.com/NevaMind-AI/memU),
// an agentic memory framework for LLM and AI agent backends.
//
// memU receives multimodal inputs (conversations, documents, images), extracts them
// into structured memory, and organizes them into a hierarchical file system that
// supports both embedding-based (RAG) and non-embedding (LLM) retrieval.
//
// This package implements the Memory interface from the parent memory package,
// allowing LangGraphGo agents to use memU as a backend for advanced memory management.
package memu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a memU client that implements the Memory interface
type Client struct {
	baseURL        string
	apiKey         string
	userID         string
	httpClient     *http.Client
	retrieveMethod string // "rag" or "llm"
}

// Config holds the configuration for the memU client
type Config struct {
	// BaseURL is the base URL of the memU API
	// For cloud: "https://api.memu.so"
	// For self-hosted: e.g., "http://localhost:8000"
	BaseURL string

	// APIKey is the authentication key for the memU API
	APIKey string

	// UserID is the user identifier for memory isolation
	UserID string

	// RetrieveMethod specifies the retrieval method: "rag" (default) or "llm"
	// "rag" - Fast embedding-based search
	// "llm" - Deep semantic understanding search
	RetrieveMethod string

	// HTTPClient is the HTTP client to use for API requests
	// If nil, a default client with 30s timeout will be used
	HTTPClient *http.Client
}

// NewClient creates a new memU client with the given configuration
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if cfg.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	retrieveMethod := cfg.RetrieveMethod
	if retrieveMethod == "" {
		retrieveMethod = "rag" // Default to RAG for faster retrieval
	}

	return &Client{
		baseURL:        cfg.BaseURL,
		apiKey:         cfg.APIKey,
		userID:         cfg.UserID,
		httpClient:     httpClient,
		retrieveMethod: retrieveMethod,
	}, nil
}

// doRequest performs an HTTP request with proper headers and error handling
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader, result any) error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// memorize sends a memorization request to memU API
func (c *Client) memorize(ctx context.Context, queries []Query) error {
	req := MemorizeRequest{
		Queries:  queries,
		User:     map[string]any{"user_id": c.userID},
		Modality: "conversation",
		Async:    false,
	}

	var result MemorizeResponse
	if err := c.doRequest(ctx, "POST", "/api/v3/memory/memorize", requestBody(req), &result); err != nil {
		return fmt.Errorf("failed to memorize: %w", err)
	}

	return nil
}

// retrieve sends a retrieval request to memU API
func (c *Client) retrieve(ctx context.Context, query string) (*RetrieveResponse, error) {
	queries := []Query{
		{
			Role:    "user",
			Content: map[string]any{"text": query},
		},
	}

	req := RetrieveRequest{
		Queries: queries,
		Where:   map[string]any{"user_id": c.userID},
		Method:  c.retrieveMethod,
	}

	var result RetrieveResponse
	if err := c.doRequest(ctx, "POST", "/api/v3/memory/retrieve", requestBody(req), &result); err != nil {
		return nil, fmt.Errorf("failed to retrieve memory: %w", err)
	}

	return &result, nil
}

// getCategories sends a categories request to memU API
func (c *Client) getCategories(ctx context.Context) (*CategoriesResponse, error) {
	req := CategoriesRequest{
		Where: map[string]any{"user_id": c.userID},
	}

	var result CategoriesResponse
	if err := c.doRequest(ctx, "POST", "/api/v3/memory/categories", requestBody(req), &result); err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return &result, nil
}

// requestBody helper to encode request body
func requestBody(v any) io.Reader {
	data, _ := json.Marshal(v)
	return &readerWithCloser{Reader: &reader{data: data}}
}

// reader implements io.Reader
type reader struct {
	data []byte
	pos  int
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// readerWithCloser wraps an io.Reader to provide an io.ReadCloser
type readerWithCloser struct {
	io.Reader
}

func (rc *readerWithCloser) Close() error {
	return nil
}

// estimateTokens provides a rough estimate of token count
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	// Rough approximation: ~4 characters per token
	return len(text) / 4
}
