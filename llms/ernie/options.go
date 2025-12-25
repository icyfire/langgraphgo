package ernie

import (
	"net/http"
	"os"

	"github.com/tmc/langchaingo/callbacks"
)

// ModelName represents the model identifier for Baidu Qianfan (Ernie) API.
//
// IMPORTANT: You should use your model name as a string directly.
// Common model names include:
//   - ernie-4.5-turbo-128k (推荐)
//   - ernie-speed-128k
//   - ernie-speed-8k
//   - deepseek-r1
//
// For the complete model list, visit: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Nlks5zkzu
//
// Example usage:
//
//	llm, err := ernie.New(
//		ernie.WithAPIKey("your-api-key"),
//		ernie.WithModel("ernie-4.5-turbo-128k"), // Use model name directly as string
//	)
type ModelName string

type options struct {
	apiKey           string
	modelName        ModelName
	httpClient       *http.Client
	callbacksHandler callbacks.Handler
	baseURL          string
}

// Option is a function that configures an LLM.
type Option func(*options)

// WithAPIKey sets the API key for the LLM.
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithModel sets the model name for the LLM.
// You should use the model name directly as a string.
// Common models: ernie-4.5-turbo-128k, ernie-speed-128k, deepseek-r1, etc.
// For the complete model list, visit: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Nlks5zkzu
func WithModel(model ModelName) Option {
	return func(opts *options) {
		opts.modelName = model
	}
}

// WithHTTPClient sets the HTTP client for the LLM.
func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}

// WithCallbacks sets the callbacks handler for the LLM.
func WithCallbacks(handler callbacks.Handler) Option {
	return func(opts *options) {
		opts.callbacksHandler = handler
	}
}

// WithBaseURL sets the base URL for the LLM API.
// Default is "https://qianfan.baidubce.com".
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = baseURL
	}
}

// getEnvOrDefault retrieves an environment variable or returns the default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
