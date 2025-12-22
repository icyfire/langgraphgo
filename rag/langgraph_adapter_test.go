package rag

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockEngine is a mock implementation of Engine for testing
type mockEngine struct {
	queryResult *QueryResult
	queryError  error
	queryCalled bool
	queryInput  string
}

func (m *mockEngine) Query(ctx context.Context, query string) (*QueryResult, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.queryCalled = true
	m.queryInput = query

	if m.queryError != nil {
		return nil, m.queryError
	}

	if m.queryResult != nil {
		return m.queryResult, nil
	}

	// Default result
	return &QueryResult{
		Query:      query,
		Context:    "Default context for: " + query,
		Answer:     "Default answer",
		Confidence: 0.9,
	}, nil
}

func (m *mockEngine) QueryWithConfig(ctx context.Context, query string, config *RetrievalConfig) (*QueryResult, error) {
	return m.Query(ctx, query)
}

func (m *mockEngine) AddDocuments(ctx context.Context, docs []Document) error {
	return nil
}

func (m *mockEngine) DeleteDocument(ctx context.Context, docID string) error {
	return nil
}

func (m *mockEngine) UpdateDocument(ctx context.Context, doc Document) error {
	return nil
}

func (m *mockEngine) SimilaritySearch(ctx context.Context, query string, k int) ([]Document, error) {
	return []Document{}, nil
}

func (m *mockEngine) SimilaritySearchWithScores(ctx context.Context, query string, k int) ([]DocumentSearchResult, error) {
	return []DocumentSearchResult{}, nil
}

// ========================================
// NewRetrievalNode Tests
// ========================================

func TestNewRetrievalNode(t *testing.T) {
	engine := &mockEngine{
		queryResult: &QueryResult{
			Query:   "test query",
			Context: "retrieved context",
		},
	}

	node := NewRetrievalNode(engine, "question", "context")

	if node == nil {
		t.Fatal("NewRetrievalNode returned nil")
	}
}

func TestNewRetrievalNode_Success(t *testing.T) {
	tests := []struct {
		name           string
		inputState     map[string]any
		inputKey       string
		outputKey      string
		queryResult    *QueryResult
		expectedOutput map[string]any
	}{
		{
			name: "successful retrieval",
			inputState: map[string]any{
				"question": "What is the capital of France?",
			},
			inputKey:  "question",
			outputKey: "context",
			queryResult: &QueryResult{
				Query:   "What is the capital of France?",
				Context: "Paris is the capital of France.",
				Answer:  "Paris",
			},
			expectedOutput: map[string]any{
				"context": "Paris is the capital of France.",
			},
		},
		{
			name: "custom input and output keys",
			inputState: map[string]any{
				"user_query": "How does RAG work?",
			},
			inputKey:  "user_query",
			outputKey: "retrieved_docs",
			queryResult: &QueryResult{
				Query:   "How does RAG work?",
				Context: "RAG combines retrieval and generation.",
			},
			expectedOutput: map[string]any{
				"retrieved_docs": "RAG combines retrieval and generation.",
			},
		},
		{
			name: "empty query",
			inputState: map[string]any{
				"question": "",
			},
			inputKey:  "question",
			outputKey: "context",
			queryResult: &QueryResult{
				Query:   "",
				Context: "empty result",
			},
			expectedOutput: map[string]any{
				"context": "empty result",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &mockEngine{queryResult: tt.queryResult}
			node := NewRetrievalNode(engine, tt.inputKey, tt.outputKey)

			ctx := context.Background()
			result, err := node(ctx, tt.inputState)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Fatalf("expected map[string]any, got %T", result)
			}

			for key, expectedValue := range tt.expectedOutput {
				actualValue, ok := resultMap[key]
				if !ok {
					t.Errorf("missing key %q in result", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("for key %q: expected %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestNewRetrievalNode_Errors(t *testing.T) {
	tests := []struct {
		name        string
		inputState  any
		inputKey    string
		outputKey   string
		queryError  error
		expectedErr string
	}{
		{
			name:        "state is not a map",
			inputState:  "not a map",
			inputKey:    "question",
			outputKey:   "context",
			expectedErr: "state is not a map[string]any, got string",
		},
		{
			name:        "input key not found",
			inputState:  map[string]any{"other_key": "value"},
			inputKey:    "question",
			outputKey:   "context",
			expectedErr: "input key 'question' not found or not a string",
		},
		{
			name:        "input key is not a string",
			inputState:  map[string]any{"question": 123},
			inputKey:    "question",
			outputKey:   "context",
			expectedErr: "input key 'question' not found or not a string",
		},
		{
			name:        "engine query error",
			inputState:  map[string]any{"question": "test"},
			inputKey:    "question",
			outputKey:   "context",
			queryError:  errors.New("query failed"),
			expectedErr: "query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &mockEngine{queryError: tt.queryError}
			node := NewRetrievalNode(engine, tt.inputKey, tt.outputKey)

			ctx := context.Background()
			_, err := node(ctx, tt.inputState)

			if err == nil {
				t.Error("expected error but got nil")
			}
			if err != nil && tt.expectedErr != "" && err.Error() != tt.expectedErr {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestNewRetrievalNode_WithContext(t *testing.T) {
	engine := &mockEngine{
		queryResult: &QueryResult{
			Query:   "test query",
			Context: "test context",
		},
	}

	node := NewRetrievalNode(engine, "question", "context")

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := node(ctx, map[string]any{"question": "test"})
	if err == nil {
		t.Error("expected error due to context cancellation")
	}
}

// ========================================
// RetrieverTool Tests
// ========================================

func TestNewRetrieverTool(t *testing.T) {
	engine := &mockEngine{}

	t.Run("with all parameters", func(t *testing.T) {
		tool := NewRetrieverTool(engine, "my_tool", "My custom tool description")

		if tool == nil {
			t.Fatal("NewRetrieverTool returned nil")
		}
		if tool.NameVal != "my_tool" {
			t.Errorf("expected name 'my_tool', got %q", tool.NameVal)
		}
		if tool.DescVal != "My custom tool description" {
			t.Errorf("expected description 'My custom tool description', got %q", tool.DescVal)
		}
	})

	t.Run("with empty name", func(t *testing.T) {
		tool := NewRetrieverTool(engine, "", "")

		if tool.NameVal != "knowledge_base" {
			t.Errorf("expected default name 'knowledge_base', got %q", tool.NameVal)
		}
		if tool.DescVal != "A knowledge base tool. Use this to search for information to answer questions." {
			t.Errorf("expected default description, got %q", tool.DescVal)
		}
	})

	t.Run("with empty description", func(t *testing.T) {
		tool := NewRetrieverTool(engine, "custom_name", "")

		if tool.NameVal != "custom_name" {
			t.Errorf("expected name 'custom_name', got %q", tool.NameVal)
		}
		if tool.DescVal != "A knowledge base tool. Use this to search for information to answer questions." {
			t.Errorf("expected default description, got %q", tool.DescVal)
		}
	})
}

func TestRetrieverTool_Name(t *testing.T) {
	engine := &mockEngine{}
	tool := NewRetrieverTool(engine, "test_tool", "test description")

	if tool.Name() != "test_tool" {
		t.Errorf("expected 'test_tool', got %q", tool.Name())
	}
}

func TestRetrieverTool_Description(t *testing.T) {
	engine := &mockEngine{}
	tool := NewRetrieverTool(engine, "test_tool", "test description")

	if tool.Description() != "test description" {
		t.Errorf("expected 'test description', got %q", tool.Description())
	}
}

func TestRetrieverTool_Call(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		queryResult    *QueryResult
		queryError     error
		expectedResult string
		expectedErr    bool
	}{
		{
			name:  "successful call",
			input: "What is RAG?",
			queryResult: &QueryResult{
				Query:   "What is RAG?",
				Context: "RAG stands for Retrieval Augmented Generation.",
				Answer:  "It's a technique...",
			},
			expectedResult: "RAG stands for Retrieval Augmented Generation.",
			expectedErr:    false,
		},
		{
			name:           "empty input",
			input:          "",
			queryResult:    &QueryResult{Context: "empty result"},
			expectedResult: "empty result",
			expectedErr:    false,
		},
		{
			name:        "engine error",
			input:       "test",
			queryError:  errors.New("engine failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &mockEngine{queryResult: tt.queryResult, queryError: tt.queryError}
			tool := NewRetrieverTool(engine, "", "")

			ctx := context.Background()
			result, err := tool.Call(ctx, tt.input)

			if tt.expectedErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expectedResult {
				t.Errorf("expected %q, got %q", tt.expectedResult, result)
			}
		})
	}
}

func TestRetrieverTool_WithEngine(t *testing.T) {
	// Test that the tool properly uses the engine
	engine := &mockEngine{
		queryResult: &QueryResult{
			Query:   "test query",
			Context: "test context",
		},
	}

	tool := NewRetrieverTool(engine, "test", "test tool")

	ctx := context.Background()
	result, err := tool.Call(ctx, "test query")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !engine.queryCalled {
		t.Error("engine.Query was not called")
	}

	if engine.queryInput != "test query" {
		t.Errorf("expected query input 'test query', got %q", engine.queryInput)
	}

	if result != "test context" {
		t.Errorf("expected 'test context', got %q", result)
	}
}

func TestRetrieverTool_CallWithContextCancellation(t *testing.T) {
	engine := &mockEngine{
		queryResult: &QueryResult{Context: "result"},
	}
	tool := NewRetrieverTool(engine, "", "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tool.Call(ctx, "test")
	if err == nil {
		t.Error("expected error due to context cancellation")
	}
}

// ========================================
// Interface Compliance Tests
// ========================================

func TestRetrieverTool_InterfaceCompliance(t *testing.T) {
	// Ensure RetrieverTool implements the expected interface
	engine := &mockEngine{}
	tool := NewRetrieverTool(engine, "", "")

	// Check that Name and Description methods exist
	if tool.Name() == "" {
		t.Error("Name() should not return empty string")
	}

	if tool.Description() == "" {
		t.Error("Description() should not return empty string")
	}
}

// ========================================
// Edge Cases and Integration Tests
// ========================================

func TestNewRetrievalNode_PreservesOtherState(t *testing.T) {
	// Test that the node only updates the output key and doesn't affect other state
	engine := &mockEngine{
		queryResult: &QueryResult{
			Context: "retrieved context",
		},
	}

	node := NewRetrievalNode(engine, "question", "context")

	inputState := map[string]any{
		"question":  "test query",
		"other_key": "other_value",
		"number":    42,
	}

	ctx := context.Background()
	result, err := node(ctx, inputState)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultMap := result.(map[string]any)

	// Should only contain the output key
	if len(resultMap) != 1 {
		t.Errorf("expected result map with 1 key, got %d keys", len(resultMap))
	}

	if resultMap["context"] != "retrieved context" {
		t.Errorf("expected context 'retrieved context', got %v", resultMap["context"])
	}
}

func TestRetrieverTool_WithComplexQueryResult(t *testing.T) {
	engine := &mockEngine{
		queryResult: &QueryResult{
			Query:        "complex query",
			Context:      "complex context with newlines\nand multiple\nlines",
			Answer:       "complex answer",
			Confidence:   0.95,
			ResponseTime: 100 * time.Millisecond,
			Metadata: map[string]any{
				"source": "test",
				"count":  5,
			},
		},
	}

	tool := NewRetrieverTool(engine, "", "")

	ctx := context.Background()
	result, err := tool.Call(ctx, "complex query")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedResult := "complex context with newlines\nand multiple\nlines"
	if result != expectedResult {
		t.Errorf("expected %q, got %q", expectedResult, result)
	}
}
