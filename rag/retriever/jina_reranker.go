package retriever

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"time"

	"github.com/smallnest/langgraphgo/rag"
)

// JinaRerankerConfig configures the Jina reranker
type JinaRerankerConfig struct {
	// Model is the Jina rerank model to use
	// Options: "jina-reranker-v1-base-en", "jina-reranker-v2-base-multilingual"
	Model string
	// TopK is the number of documents to return
	TopK int
	// APIBase is the custom API base URL (optional)
	APIBase string
	// Timeout is the HTTP request timeout
	Timeout time.Duration
}

// DefaultJinaRerankerConfig returns the default configuration for Jina reranker
func DefaultJinaRerankerConfig() JinaRerankerConfig {
	return JinaRerankerConfig{
		Model:   "jina-reranker-v2-base-multilingual",
		TopK:    5,
		APIBase: "https://api.jina.ai/v1/rerank",
		Timeout: 30 * time.Second,
	}
}

// JinaReranker uses Jina AI's Rerank API to rerank documents
type JinaReranker struct {
	apiKey string
	client *http.Client
	config JinaRerankerConfig
}

// NewJinaReranker creates a new Jina reranker
// The API key can be provided via the apiKey parameter or JINA_API_KEY environment variable
func NewJinaReranker(apiKey string, config JinaRerankerConfig) *JinaReranker {
	if apiKey == "" {
		apiKey = os.Getenv("JINA_API_KEY")
	}
	if config.Model == "" {
		config = DefaultJinaRerankerConfig()
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.APIBase == "" {
		config.APIBase = "https://api.jina.ai/v1/rerank"
	}

	return &JinaReranker{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// jinaRerankRequest represents the request body for Jina Rerank API
type jinaRerankRequest struct {
	Query     string         `json:"query"`
	Documents []jinaDocument `json:"documents"`
	TopN      int            `json:"top_n,omitempty"`
	Model     string         `json:"model,omitempty"`
}

// jinaDocument represents a document in the Jina API
type jinaDocument struct {
	Text  string `json:"text"`
	Title string `json:"title,omitempty"`
}

// jinaRerankResponse represents the response from Jina Rerank API
type jinaRerankResponse struct {
	Model   string             `json:"model"`
	Results []jinaRerankResult `json:"results"`
	Usage   jinaUsage          `json:"usage"`
}

// jinaRerankResult represents a single rerank result
type jinaRerankResult struct {
	Index          int          `json:"index"`
	Document       jinaDocument `json:"document"`
	RelevanceScore float64      `json:"relevance_score"`
}

// jinaUsage represents token usage in the response
type jinaUsage struct {
	TotalTokens int `json:"total_tokens"`
}

// Rerank reranks documents based on query relevance using Jina's Rerank API
func (r *JinaReranker) Rerank(ctx context.Context, query string, documents []rag.DocumentSearchResult) ([]rag.DocumentSearchResult, error) {
	if len(documents) == 0 {
		return []rag.DocumentSearchResult{}, nil
	}

	if r.apiKey == "" {
		return nil, fmt.Errorf("Jina API key is required. Set JINA_API_KEY environment variable or pass apiKey parameter")
	}

	// Prepare request body
	reqDocs := make([]jinaDocument, len(documents))
	for i, doc := range documents {
		title := ""
		if t, ok := doc.Document.Metadata["title"].(string); ok {
			title = t
		}
		reqDocs[i] = jinaDocument{
			Text:  doc.Document.Content,
			Title: title,
		}
	}

	reqBody := jinaRerankRequest{
		Query:     query,
		Documents: reqDocs,
		TopN:      r.config.TopK,
		Model:     r.config.Model,
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
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

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
		return nil, fmt.Errorf("Jina API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var rerankResp jinaRerankResponse
	if err := json.Unmarshal(body, &rerankResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map results back to documents
	results := make([]rag.DocumentSearchResult, len(rerankResp.Results))
	for i, result := range rerankResp.Results {
		originalDoc := documents[result.Index]
		results[i] = rag.DocumentSearchResult{
			Document: originalDoc.Document,
			Score:    result.RelevanceScore,
			Metadata: r.mergeMetadata(originalDoc.Metadata, map[string]any{
				"jina_rerank_score": result.RelevanceScore,
				"original_score":    originalDoc.Score,
				"original_index":    result.Index,
				"reranking_method":  "jina",
				"rerank_model":      rerankResp.Model,
			}),
		}
	}

	return results, nil
}

// mergeMetadata merges two metadata maps
func (r *JinaReranker) mergeMetadata(m1, m2 map[string]any) map[string]any {
	result := make(map[string]any)
	maps.Copy(result, m1)
	maps.Copy(result, m2)
	return result
}
