package memu

import "time"

// Message represents a single message compatible with the memory package interface
type Message struct {
	ID         string
	Role       string // "user", "assistant", "system"
	Content    string
	Timestamp  time.Time
	Metadata   map[string]any
	TokenCount int
}

// Stats contains statistics about memory usage (compatible with memory package)
type Stats struct {
	TotalMessages   int
	TotalTokens     int
	ActiveMessages  int
	ActiveTokens    int
	CompressionRate float64
}

// Query represents a single query in the conversation
type Query struct {
	Role    string         `json:"role"`
	Content map[string]any `json:"content"`
}

// MemorizeRequest is the request body for the memorize endpoint
type MemorizeRequest struct {
	Queries  []Query        `json:"queries"`
	User     map[string]any `json:"user"`
	Modality string         `json:"modality"`
	Async    bool           `json:"async,omitempty"`
}

// MemorizeResponse is the response from the memorize endpoint
type MemorizeResponse struct {
	TaskID     string     `json:"task_id"`
	Status     string     `json:"status"`
	Categories []Category `json:"categories"`
	Items      []Item     `json:"items"`
	Resources  []Resource `json:"resources"`
	Message    string     `json:"message,omitempty"`
}

// RetrieveRequest is the request body for the retrieve endpoint
type RetrieveRequest struct {
	Queries []Query        `json:"queries"`
	Where   map[string]any `json:"where,omitempty"`
	Method  string         `json:"method,omitempty"` // "rag" or "llm"
}

// RetrieveResponse is the response from the retrieve endpoint
type RetrieveResponse struct {
	Categories []Category `json:"categories"`
	Items      []Item     `json:"items"`
	Resources  []Resource `json:"resources"`
	Message    string     `json:"message,omitempty"`
}

// CategoriesRequest is the request body for the categories endpoint
type CategoriesRequest struct {
	Where map[string]any `json:"where,omitempty"`
}

// CategoriesResponse is the response from the categories endpoint
type CategoriesResponse struct {
	Categories []Category `json:"categories"`
	Total      int        `json:"total,omitempty"`
}

// Category represents a memory category in memU
type Category struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Summary     string   `json:"summary"`
	ItemIDs     []string `json:"item_ids"`
	ResourceIDs []string `json:"resource_ids"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// Item represents a memory item in memU
type Item struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"` // e.g., "preference", "skill", "opinion", "habit"
	Summary    string         `json:"summary"`
	Content    string         `json:"content"`
	Category   string         `json:"category"`
	Importance float64        `json:"importance"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

// Resource represents a raw resource in memU
type Resource struct {
	ID        string         `json:"id"`
	Modality  string         `json:"modality"` // e.g., "conversation", "document", "image"
	URL       string         `json:"url"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt string         `json:"created_at"`
}
