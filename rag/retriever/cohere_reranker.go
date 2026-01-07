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

// CohereRerankerConfig configures the Cohere reranker
type CohereRerankerConfig struct {
	// Model is the Cohere rerank model to use
	// Options: "rerank-v3.5", "rerank-english-v3.0", "rerank-multilingual-v3.0"
	Model string
	// TopK is the number of documents to return
	TopK int
	// APIBase is the custom API base URL (optional)
	APIBase string
	// Timeout is the HTTP request timeout
	Timeout time.Duration
}

// DefaultCohereRerankerConfig returns the default configuration for Cohere reranker
func DefaultCohereRerankerConfig() CohereRerankerConfig {
	return CohereRerankerConfig{
		Model:   "rerank-v3.5",
		TopK:    5,
		APIBase: "https://api.cohere.ai/v1/rerank",
		Timeout: 30 * time.Second,
	}
}

// CohereReranker uses Cohere's Rerank API to rerank documents
type CohereReranker struct {
	apiKey string
	client *http.Client
	config CohereRerankerConfig
}

// NewCohereReranker creates a new Cohere reranker
// The API key can be provided via the apiKey parameter or COHERE_API_KEY environment variable
func NewCohereReranker(apiKey string, config CohereRerankerConfig) *CohereReranker {
	if apiKey == "" {
		apiKey = os.Getenv("COHERE_API_KEY")
	}
	if config.Model == "" {
		config = DefaultCohereRerankerConfig()
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.APIBase == "" {
		config.APIBase = "https://api.cohere.ai/v1/rerank"
	}

	return &CohereReranker{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// cohereRerankRequest represents the request body for Cohere Rerank API
type cohereRerankRequest struct {
	Query      string           `json:"query"`
	Documents  []cohereDocument `json:"documents"`
	TopN       int              `json:"top_n,omitempty"`
	Model      string           `json:"model,omitempty"`
	RankFields []string         `json:"rank_fields,omitempty"`
}

// cohereDocument represents a document in the Cohere API
type cohereDocument struct {
	Text  string `json:"text"`
	Title string `json:"title,omitempty"`
}

// cohereRerankResponse represents the response from Cohere Rerank API
type cohereRerankResponse struct {
	Results []cohereRerankResult `json:"results"`
	Meta    cohereMeta           `json:"meta"`
}

// cohereRerankResult represents a single rerank result
type cohereRerankResult struct {
	Index          int            `json:"index"`
	RelevanceScore float64        `json:"relevance_score"`
	Document       cohereDocument `json:"document"`
}

// cohereMeta represents metadata in the response
type cohereMeta struct {
	APIVersion struct {
		Version string `json:"version"`
	} `json:"api_version"`
}

// Rerank reranks documents based on query relevance using Cohere's Rerank API
func (r *CohereReranker) Rerank(ctx context.Context, query string, documents []rag.DocumentSearchResult) ([]rag.DocumentSearchResult, error) {
	if len(documents) == 0 {
		return []rag.DocumentSearchResult{}, nil
	}

	if r.apiKey == "" {
		return nil, fmt.Errorf("Cohere API key is required. Set COHERE_API_KEY environment variable or pass apiKey parameter")
	}

	// Prepare request body
	reqDocs := make([]cohereDocument, len(documents))
	for i, doc := range documents {
		title := ""
		if t, ok := doc.Document.Metadata["title"].(string); ok {
			title = t
		}
		reqDocs[i] = cohereDocument{
			Text:  doc.Document.Content,
			Title: title,
		}
	}

	reqBody := cohereRerankRequest{
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
	req.Header.Set("X-Client-Name", "langgraphgo")

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
		return nil, fmt.Errorf("Cohere API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var rerankResp cohereRerankResponse
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
				"cohere_rerank_score": result.RelevanceScore,
				"original_score":      originalDoc.Score,
				"original_index":      result.Index,
				"reranking_method":    "cohere",
				"rerank_model":        r.config.Model,
			}),
		}
	}

	return results, nil
}

// mergeMetadata merges two metadata maps
func (r *CohereReranker) mergeMetadata(m1, m2 map[string]any) map[string]any {
	result := make(map[string]any)
	maps.Copy(result, m1)
	maps.Copy(result, m2)
	return result
}
