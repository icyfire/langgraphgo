package ernie

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// getTestModel returns the model name from env or default.
//
// IMPORTANT: Set the ERNIE_MODEL environment variable to your preferred model.
// Common models: ernie-4.5-turbo-128k, ernie-speed-128k, ernie-speed-8k, deepseek-r1
func getTestModel() ModelName {
	return ModelName(getEnvOrDefault("ERNIE_MODEL", "ernie-speed-8k"))
}

// getTestEmbeddingModel returns the embedding model name from env or default.
func getTestEmbeddingModel() ModelName {
	return ModelName(getEnvOrDefault("ERNIE_EMBEDDING_MODEL", "embedding-v1"))
}

// TestLLM_Create tests the LLM creation with various options.
func TestLLM_Create(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "with api key",
			opts: []Option{
				WithAPIKey("test-key"),
			},
			wantErr: false,
		},
		{
			name: "with api key and model",
			opts: []Option{
				WithAPIKey("test-key"),
				WithModel("ernie-speed-8k"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && llm == nil {
				t.Error("New() returned nil LLM")
			}
		})
	}
}

// TestLLM_GenerateContent tests the content generation with real API.
// Skipped if QIANFAN_TOKEN is not set.
func TestLLM_GenerateContent(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	llm, err := New(
		WithAPIKey(apiKey),
		WithModel(getTestModel()),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	ctx := context.Background()
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Hello, how are you?"),
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	content := resp.Choices[0].Content
	if content == "" {
		t.Error("Empty response content")
	}

	t.Logf("Response: %s", content)
	t.Logf("StopReason: %s", resp.Choices[0].StopReason)

	// Check GenerationInfo for token usage
	if resp.Choices[0].GenerationInfo != nil {
		if totalTokens, ok := resp.Choices[0].GenerationInfo["total_tokens"].(int); ok {
			t.Logf("Total tokens: %d", totalTokens)
		}
	}
}

// TestLLM_CreateEmbedding tests the embedding generation with real API.
// Skipped if QIANFAN_TOKEN is not set.
func TestLLM_CreateEmbedding(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	embeddingModel := getTestEmbeddingModel()
	if embeddingModel == "embedding-v1" {
		t.Skip(`ERNIE_EMBEDDING_MODEL not set.
Please set ERNIE_EMBEDDING_MODEL to a valid embedding endpoint ID.
Common models: embedding-v1, bge-large-zh, bge-large-en, tao-8k`)
	}

	llm, err := New(
		WithAPIKey(apiKey),
		WithModel(embeddingModel),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	ctx := context.Background()
	texts := []string{"Hello world"}

	embeddings, err := llm.CreateEmbedding(ctx, texts)
	if err != nil {
		// Check if it's a 404 error (model not found)
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "ResourceNotFound") {
			t.Skipf(`Embedding model '%s' not found. Please set ERNIE_EMBEDDING_MODEL to a valid endpoint ID.
Error: %v`, embeddingModel, err)
		}
		t.Fatalf("Failed to create embedding: %v", err)
	}

	if len(embeddings) != 1 {
		t.Fatalf("Expected 1 embedding, got %d", len(embeddings))
	}

	if len(embeddings[0]) == 0 {
		t.Fatal("Empty embedding")
	}

	t.Logf("Embedding dimension: %d", len(embeddings[0]))
}

// TestLLM_CreateEmbeddingMultiple tests embedding generation for multiple texts.
func TestLLM_CreateEmbeddingMultiple(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	embeddingModel := getTestEmbeddingModel()
	if embeddingModel == "embedding-v1" {
		t.Skip(`ERNIE_EMBEDDING_MODEL not set.
Please set ERNIE_EMBEDDING_MODEL to a valid embedding endpoint ID.`)
	}

	llm, err := New(
		WithAPIKey(apiKey),
		WithModel(embeddingModel),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	ctx := context.Background()
	texts := []string{"Hello", "World"}

	embeddings, err := llm.CreateEmbedding(ctx, texts)
	if err != nil {
		// Check if it's a 404 error (model not found)
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "ResourceNotFound") {
			t.Skipf(`Embedding model '%s' not found. Please set ERNIE_EMBEDDING_MODEL to a valid endpoint ID.`, embeddingModel)
		}
		t.Fatalf("Failed to create embedding: %v", err)
	}

	if len(embeddings) != 2 {
		t.Fatalf("Expected 2 embeddings, got %d", len(embeddings))
	}

	for i, emb := range embeddings {
		if len(emb) == 0 {
			t.Errorf("Empty embedding at index %d", i)
		}
		t.Logf("Embedding %d dimension: %d", i, len(emb))
	}
}

// TestLLM_Call tests the Call method.
func TestLLM_Call(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	llm, err := New(
		WithAPIKey(apiKey),
		WithModel(getTestModel()),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	ctx := context.Background()
	response, err := llm.Call(ctx, "What is 2+2?")
	if err != nil {
		t.Fatalf("Failed to call LLM: %v", err)
	}

	if response == "" {
		t.Error("Empty response")
	}

	t.Logf("Response: %s", response)
}

// TestLLM_ModelString tests model name to string conversion.
func TestLLM_ModelString(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected string
	}{
		// 推荐模型
		{"ERNIE 5.0 Thinking Preview", "ernie-5.0-thinking-preview", "ernie-5.0-thinking-preview"},
		{"ERNIE 4.5 Turbo 128K", "ernie-4.5-turbo-128k", "ernie-4.5-turbo-128k"},
		{"DeepSeek R1", "deepseek-r1", "deepseek-r1"},

		// ERNIE系列
		{"ERNIE Speed 8K", "ernie-speed-8k", "ernie-speed-8k"},
		{"ERNIE Lite 8K", "ernie-lite-8k", "ernie-lite-8k"},
		{"ERNIE Tiny 8K", "ernie-tiny-8k", "ernie-tiny-8k"},

		// DeepSeek系列
		{"DeepSeek V3", "deepseek-v3", "deepseek-v3"},
		{"DeepSeek V3.2", "deepseek-v3.2", "deepseek-v3.2"},

		// Qwen系列
		{"Qwen3 8B", "qwen3-8b", "qwen3-8b"},
		{"Qwen3 32B", "qwen3-32b", "qwen3-32b"},

		// Embedding模型
		{"Embedding V1", "embedding-v1", "embedding-v1"},
		{"BGE Large ZH", "bge-large-zh", "bge-large-zh"},
		{"BGE Large EN", "bge-large-en", "bge-large-en"},
		{"Tao 8k", "tao-8k", "tao-8k"},
		{"Qwen3 Embedding 0.6B", "qwen3-embedding-0.6b", "qwen3-embedding-0.6b"},
		{"Qwen3 Embedding 4B", "qwen3-embedding-4b", "qwen3-embedding-4b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modelToModelString(ModelName(tt.model))
			if result != tt.expected {
				t.Errorf("modelToModelString(%v) = %s, want %s", tt.model, result, tt.expected)
			}
		})
	}
}

// TestLLM_Conversation tests a multi-turn conversation.
func TestLLM_Conversation(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	llm, err := New(
		WithAPIKey(apiKey),
		WithModel(getTestModel()),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	ctx := context.Background()
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("My name is Alice"),
			},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPart("Hello Alice! Nice to meet you."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What's my name?"),
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	content := resp.Choices[0].Content
	t.Logf("Response: %s", content)

	// Check if the model remembers the name
	containsName := false
	lowerContent := content
	if len(lowerContent) > 0 {
		// Simple check for name presence
		for i := 0; i <= len(lowerContent)-5; i++ {
			if lowerContent[i:i+5] == "Alice" {
				containsName = true
				break
			}
		}
	}
	t.Logf("Model remembers name: %v", containsName)
}

// TestLLM_DifferentModels tests different model types.
func TestLLM_DifferentModels(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	models := []struct {
		name ModelName
		desc string
	}{
		{"ernie-speed-8k", "ERNIE Speed 8K - fast response"},
		{"ernie-tiny-8k", "ERNIE Tiny 8K - lightweight"},
		{"ernie-lite-8k", "ERNIE Lite 8K - basic"},
	}

	for _, m := range models {
		t.Run(m.desc, func(t *testing.T) {
			llm, err := New(
				WithAPIKey(apiKey),
				WithModel(m.name),
			)
			if err != nil {
				t.Fatalf("Failed to create LLM: %v", err)
			}

			ctx := context.Background()
			response, err := llm.Call(ctx, "Say hello")
			if err != nil {
				t.Logf("Model %s error: %v", m.name, err)
				return
			}

			if response == "" {
				t.Errorf("Model %s returned empty response", m.name)
			}

			t.Logf("Model %s response: %s", m.name, response)
		})
	}
}

// TestLLM_EmbeddingModels tests different embedding models.
func TestLLM_EmbeddingModels(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	models := []struct {
		name        ModelName
		expectedDim int
		description string
	}{
		{"embedding-v1", 384, "Embedding V1 - 384 dimensions"},
		{"bge-large-zh", 1024, "BGE Large ZH - 1024 dimensions"},
		{"tao-8k", 1024, "Tao 8k - 1024 dimensions"},
	}

	for _, m := range models {
		t.Run(m.description, func(t *testing.T) {
			llm, err := New(
				WithAPIKey(apiKey),
				WithModel(m.name),
			)
			if err != nil {
				t.Fatalf("Failed to create LLM: %v", err)
			}

			ctx := context.Background()
			embeddings, err := llm.CreateEmbedding(ctx, []string{"test text"})
			if err != nil {
				t.Logf("Model %s error: %v", m.name, err)
				return
			}

			if len(embeddings) != 1 {
				t.Fatalf("Expected 1 embedding, got %d", len(embeddings))
			}

			dim := len(embeddings[0])
			if dim != m.expectedDim {
				t.Errorf("Model %s: expected dimension %d, got %d", m.name, m.expectedDim, dim)
			}

			t.Logf("Model %s: dimension = %d", m.name, dim)
		})
	}
}

// TestLLM_ToolCall tests tool call functionality.
// Skipped if QIANFAN_TOKEN is not set.
func TestLLM_ToolCall(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	llm, err := New(
		WithAPIKey(apiKey),
		WithModel(getTestModel()),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	ctx := context.Background()

	// Define a simple tool for getting weather
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]any{
							"type":        "string",
							"enum":        []string{"celsius", "fahrenheit"},
							"description": "The temperature unit to use",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What's the weather like in San Francisco, CA?"),
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages, llms.WithTools(tools))
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	choice := resp.Choices[0]
	t.Logf("Content: %s", choice.Content)
	t.Logf("StopReason: %s", choice.StopReason)

	// Check for tool calls
	if len(choice.ToolCalls) > 0 {
		t.Logf("Tool calls returned: %d", len(choice.ToolCalls))
		for i, tc := range choice.ToolCalls {
			t.Logf("  ToolCall[%d]: ID=%s, Type=%s, Function.Name=%s, Function.Arguments=%s",
				i, tc.ID, tc.Type, tc.FunctionCall.Name, tc.FunctionCall.Arguments)
		}
	} else if choice.FuncCall != nil {
		t.Logf("FuncCall returned: Name=%s, Arguments=%s",
			choice.FuncCall.Name, choice.FuncCall.Arguments)
	} else {
		t.Log("No tool calls or function call in response")
	}
}

// TestLLM_ToolChoice tests tool choice option.
// Skipped if QIANFAN_TOKEN is not set.
func TestLLM_ToolChoice(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	llm, err := New(
		WithAPIKey(apiKey),
		WithModel(getTestModel()),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	ctx := context.Background()

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_current_time",
				Description: "Get the current time",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}

	tests := []struct {
		name       string
		toolChoice any
		desc       string
	}{
		{
			name:       "tool choice auto",
			toolChoice: "auto",
			desc:       "Model decides whether to call tools",
		},
		{
			name:       "tool choice none",
			toolChoice: "none",
			desc:       "Model will not call tools",
		},
		{
			name: "tool choice required",
			toolChoice: llms.ToolChoice{
				Type: "required",
			},
			desc: "Model must call a tool",
		},
		{
			name: "tool choice specific function",
			toolChoice: llms.ToolChoice{
				Type: "function",
				Function: &llms.FunctionReference{
					Name: "get_current_time",
				},
			},
			desc: "Model must call the specific function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextPart("What time is it?"),
					},
				},
			}

			resp, err := llm.GenerateContent(ctx, messages,
				llms.WithTools(tools),
				llms.WithToolChoice(tt.toolChoice),
			)
			if err != nil {
				t.Fatalf("Failed to generate content: %v", err)
			}

			if len(resp.Choices) == 0 {
				t.Fatal("No choices in response")
			}

			choice := resp.Choices[0]
			t.Logf("Test: %s", tt.desc)
			t.Logf("Content: %s", choice.Content)
			t.Logf("StopReason: %s", choice.StopReason)

			if len(choice.ToolCalls) > 0 {
				t.Logf("Tool calls: %d", len(choice.ToolCalls))
			}
		})
	}
}

// TestLLM_ToolResponse tests tool response handling.
// Skipped if QIANFAN_TOKEN is not set.
//
// NOTE: Baidu Qianfan API may not support tool role messages directly.
// The tool responses are typically handled by the OpenAI client wrapper.
func TestLLM_ToolResponse(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	llm, err := New(
		WithAPIKey(apiKey),
		WithModel(getTestModel()),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	ctx := context.Background()

	// Test with a simple user message first
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Hello, please respond"),
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	content := resp.Choices[0].Content
	t.Logf("Response: %s", content)

	if content == "" {
		t.Error("Empty response")
	}
}

// TestLLM_MultiToolConversation tests a multi-turn conversation with tool calls.
// Skipped if QIANFAN_TOKEN is not set.
//
// NOTE: This test demonstrates the basic tool call flow. The actual tool response
// handling is done by the OpenAI client wrapper.
func TestLLM_MultiToolConversation(t *testing.T) {
	apiKey := os.Getenv("QIANFAN_TOKEN")
	if apiKey == "" {
		t.Skip("QIANFAN_TOKEN not set")
	}

	llm, err := New(
		WithAPIKey(apiKey),
		WithModel(getTestModel()),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	ctx := context.Background()

	// Define tools
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city name",
						},
					},
					"required": []string{"location"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_time",
				Description: "Get the current time",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}

	// First message - user asks for weather
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What's the weather in Shanghai?"),
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages, llms.WithTools(tools))
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	choice := resp.Choices[0]
	t.Logf("Response - Content: %s", choice.Content)
	t.Logf("Response - StopReason: %s", choice.StopReason)
	t.Logf("Response - ToolCalls count: %d", len(choice.ToolCalls))

	// Check if tool calls were returned
	if len(choice.ToolCalls) > 0 {
		for i, tc := range choice.ToolCalls {
			t.Logf("ToolCall[%d]: ID=%s, Function=%s", i, tc.ID, tc.FunctionCall.Name)
		}
	}
}
