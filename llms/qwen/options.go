package qwen

// Option is a function that configures an Embedder.
type Option func(*Embedder)

// WithBaseURL sets the base URL for the Qwen API.
func WithBaseURL(baseURL string) Option {
	return func(e *Embedder) {
		e.baseURL = baseURL
	}
}

// WithAPIKey sets the API key for the Qwen API.
func WithAPIKey(apiKey string) Option {
	return func(e *Embedder) {
		e.apiKey = apiKey
	}
}

// WithModel sets the model name for the Qwen API.
func WithModel(model string) Option {
	return func(e *Embedder) {
		e.model = model
	}
}

// NewEmbedderWithOptions creates a new Qwen embedder with the given options.
func NewEmbedderWithOptions(opts ...Option) *Embedder {
	e := &Embedder{
		baseURL: "https://api-inference.modelscope.cn/v1",
		model:   "Qwen/Qwen3-Embedding-4B",
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}
