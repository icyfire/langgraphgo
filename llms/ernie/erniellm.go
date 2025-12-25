package ernie

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/smallnest/langgraphgo/llms/ernie/client"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrCodeResponse  = errors.New("has error code")
)

// LLM is a client for Baidu Qianfan (Ernie) LLM.
// It uses OpenAI-compatible API for chat completions and custom client for embeddings.
type LLM struct {
	chatLLM          *openai.LLM
	embeddingClient  *client.Client
	model            ModelName
	CallbacksHandler callbacks.Handler
}

var _ llms.Model = (*LLM)(nil)

// New returns a new Ernie LLM client using API Key authentication.
//
// Authentication options:
// 1. WithAPIKey(apiKey) - pass API key directly
// 2. Set ERNIE_API_KEY environment variable
//
// Model options:
// - WithModel(modelName) - set model name (default: ernie-speed-8k)
//
// Common models: ernie-4.5-turbo-128k, ernie-speed-128k, ernie-speed-8k, deepseek-r1
// For the complete model list, visit: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Nlks5zkzu
//
// Example:
//
//	llm, err := ernie.New(
//		ernie.WithAPIKey("your-api-key"),
//		ernie.WithModel("ernie-4.5-turbo-128k"),
//	)
func New(opts ...Option) (*LLM, error) {
	options := &options{
		apiKey:    getEnvOrDefault("ERNIE_API_KEY", ""),
		modelName: "ernie-speed-8k", // 默认使用 ERNIE Speed 8K
		baseURL:   "https://qianfan.baidubce.com/v2",
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.apiKey == "" {
		return nil, fmt.Errorf(`%w
You can pass auth info by using ernie.New(ernie.WithAPIKey("{API Key}"))
or
export ERNIE_API_KEY={API Key}
doc: https://cloud.baidu.com/doc/qianfan-api/s/3m9b5lqft`, client.ErrNotSetAuth)
	}

	// Create OpenAI-compatible chat client
	chatOpts := []openai.Option{
		openai.WithToken(options.apiKey),
		openai.WithBaseURL(options.baseURL),
	}

	// Apply callback handler if provided
	if options.callbacksHandler != nil {
		chatOpts = append(chatOpts, openai.WithCallback(options.callbacksHandler))
	}

	// Apply custom HTTP client if provided
	if options.httpClient != nil {
		chatOpts = append(chatOpts, openai.WithHTTPClient(options.httpClient))
	}

	chatLLM, err := openai.New(chatOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat client: %w", err)
	}

	// Create embedding client (uses custom API endpoint)
	embeddingClientOpts := []client.Option{
		client.WithAPIKey(options.apiKey),
		client.WithBaseURL(options.baseURL),
	}
	if options.httpClient != nil {
		embeddingClientOpts = append(embeddingClientOpts, client.WithHTTPClient(options.httpClient))
	}

	embeddingClient, err := client.New(embeddingClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding client: %w", err)
	}

	return &LLM{
		chatLLM:         chatLLM,
		embeddingClient: embeddingClient,
		model:           options.modelName,
		CallbacksHandler: options.callbacksHandler,
	}, nil
}

// Call generates a response from the LLM for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
// Uses OpenAI-compatible API for chat completions.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// Parse call options to check if temperature is set
	callOpts := llms.CallOptions{}
	for _, opt := range options {
		opt(&callOpts)
	}

	// Prepend model option and default temperature if not set
	opts := make([]llms.CallOption, 0, len(options)+2)
	if o.model != "" {
		opts = append(opts, llms.WithModel(string(o.model)))
	}
	// Baidu Qianfan API requires temperature in (0, 1.0] range
	// Set default temperature to 0.1 if not specified
	if callOpts.Temperature == 0 {
		opts = append(opts, llms.WithTemperature(0.1))
	}
	opts = append(opts, options...)

	return o.chatLLM.GenerateContent(ctx, messages, opts...)
}

// CreateEmbedding generates embeddings for the given texts using Ernie embedding models.
//
// The embedding model has the following limitations:
//   - Embedding-V1: token count <= 384, text length <= 1000 characters
//   - bge-large-zh or bge-large-en: token count <= 512, text length <= 2000 characters
//   - tao-8k: token count <= 8192, text length <= 28000 characters
//   - Qwen3-Embedding-0.6B or Qwen3-Embedding-4B: max 8k tokens per text
//
// Text count limits:
//   - tao-8k: only 1 text
//   - Others: max 16 texts
//
// API documentation: https://cloud.baidu.com/doc/qianfan-api/s/Fm7u3ropn
func (o *LLM) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := o.embeddingClient.CreateEmbedding(ctx, o.getModelString(llms.CallOptions{}), texts)
	if err != nil {
		return nil, err
	}

	if resp.ErrorCode > 0 {
		return nil, fmt.Errorf("%w, error_code:%v, error_msg:%v, id:%v",
			ErrCodeResponse, resp.ErrorCode, resp.ErrorMsg, resp.ID)
	}

	emb := make([][]float32, 0, len(resp.Data))
	for i := range resp.Data {
		emb = append(emb, resp.Data[i].Embedding)
	}

	return emb, nil
}

func (o *LLM) getModelString(opts llms.CallOptions) string {
	model := o.model

	if model == "" {
		model = ModelName(opts.Model)
	}

	return modelToModelString(model)
}

// modelToModelString returns the model string for API calls.
// Since ModelName constants already use the correct API IDs, this just returns the string value.
func modelToModelString(model ModelName) string {
	return string(model)
}
