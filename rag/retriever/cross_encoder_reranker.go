package retriever

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"time"

	"github.com/smallnest/langgraphgo/rag"
)

// CrossEncoderRerankerConfig configures the Cross-Encoder reranker
type CrossEncoderRerankerConfig struct {
	// ModelName is the name of the cross-encoder model
	// Common models:
	// - "cross-encoder/ms-marco-MiniLM-L-6-v2"
	// - "cross-encoder/ms-marco-MiniLM-L-12-v2"
	// - "cross-encoder/mmarco-mMiniLMv2-L12-H384-v1"
	ModelName string
	// TopK is the number of documents to return
	TopK int
	// APIBase is the URL of the cross-encoder service
	// This can be a local service (e.g., http://localhost:8000/rerank)
	// or a remote service
	APIBase string
	// Timeout is the HTTP request timeout
	Timeout time.Duration
}

// DefaultCrossEncoderRerankerConfig returns the default configuration
func DefaultCrossEncoderRerankerConfig() CrossEncoderRerankerConfig {
	return CrossEncoderRerankerConfig{
		ModelName: "cross-encoder/ms-marco-MiniLM-L-6-v2",
		TopK:      5,
		APIBase:   "http://localhost:8000/rerank",
		Timeout:   30 * time.Second,
	}
}

// CrossEncoderReranker uses a cross-encoder model service for reranking
//
// This reranker expects an HTTP service that accepts POST requests with the following JSON format:
//
//	{
//	  "query": "search query",
//	  "documents": ["document 1", "document 2", ...],
//	  "top_n": 5,
//	  "model": "model-name"
//	}
//
// And returns:
//
//	{
//	  "scores": [0.95, 0.87, ...],
//	  "indices": [0, 2, ...]
//	}
//
// You can set up a local service using Python with the sentence-transformers library.
// See the RERANKER.md file for an example setup script.
type CrossEncoderReranker struct {
	client *http.Client
	config CrossEncoderRerankerConfig
}

// NewCrossEncoderReranker creates a new cross-encoder reranker
func NewCrossEncoderReranker(config CrossEncoderRerankerConfig) *CrossEncoderReranker {
	if config.ModelName == "" {
		config = DefaultCrossEncoderRerankerConfig()
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.APIBase == "" {
		config.APIBase = "http://localhost:8000/rerank"
	}

	return &CrossEncoderReranker{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// crossEncoderRequest represents the request body for the cross-encoder service
type crossEncoderRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n,omitempty"`
	Model     string   `json:"model,omitempty"`
}

// crossEncoderResponse represents the response from the cross-encoder service
type crossEncoderResponse struct {
	Scores  []float64 `json:"scores"`
	Indices []int     `json:"indices"`
}

// Rerank reranks documents based on query relevance using cross-encoder scoring
func (r *CrossEncoderReranker) Rerank(ctx context.Context, query string, documents []rag.DocumentSearchResult) ([]rag.DocumentSearchResult, error) {
	if len(documents) == 0 {
		return []rag.DocumentSearchResult{}, nil
	}

	// Prepare request body
	docTexts := make([]string, len(documents))
	for i, doc := range documents {
		docTexts[i] = doc.Document.Content
	}

	reqBody := crossEncoderRequest{
		Query:     query,
		Documents: docTexts,
		TopN:      r.config.TopK,
		Model:     r.config.ModelName,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", r.config.APIBase, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cross-encoder service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ceResp crossEncoderResponse
	if err := json.Unmarshal(body, &ceResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map results back to documents
	results := make([]rag.DocumentSearchResult, len(ceResp.Indices))
	for i, idx := range ceResp.Indices {
		if idx >= 0 && idx < len(documents) {
			originalDoc := documents[idx]
			results[i] = rag.DocumentSearchResult{
				Document: originalDoc.Document,
				Score:    ceResp.Scores[i],
				Metadata: r.mergeMetadata(originalDoc.Metadata, map[string]any{
					"cross_encoder_score": ceResp.Scores[i],
					"original_score":      originalDoc.Score,
					"original_index":      idx,
					"reranking_method":    "cross_encoder",
					"rerank_model":        r.config.ModelName,
				}),
			}
		}
	}

	return results, nil
}

// mergeMetadata merges two metadata maps
func (r *CrossEncoderReranker) mergeMetadata(m1, m2 map[string]any) map[string]any {
	result := make(map[string]any)
	maps.Copy(result, m1)
	maps.Copy(result, m2)
	return result
}
