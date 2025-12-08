package prebuilt

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

// PEVMockTool for testing
type PEVMockTool struct {
	name        string
	description string
	response    string
}

func (m PEVMockTool) Name() string {
	return m.name
}

func (m PEVMockTool) Description() string {
	return m.description
}

func (m PEVMockTool) Call(ctx context.Context, input string) (string, error) {
	return m.response, nil
}

// PEVMockLLM for testing
type PEVMockLLM struct {
	responses []string
	callCount int
}

func (m *PEVMockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if m.callCount >= len(m.responses) {
		m.callCount++
		return &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{Content: "Default response"},
			},
		}, nil
	}

	response := m.responses[m.callCount]
	m.callCount++

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: response},
		},
	}, nil
}

func (m *PEVMockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "mock response", nil
}

func TestCreatePEVAgent(t *testing.T) {
	mockLLM := &PEVMockLLM{
		responses: []string{
			"1. Calculate 2 + 2",
			`{"is_successful": true, "reasoning": "Calculation completed successfully"}`,
			"The result of 2 + 2 is 4",
		},
	}

	mockTool := PEVMockTool{
		name:        "calculator",
		description: "Performs calculations",
		response:    "4",
	}

	config := PEVAgentConfig{
		Model:      mockLLM,
		Tools:      []tools.Tool{mockTool},
		MaxRetries: 3,
		Verbose:    false,
	}

	agent, err := CreatePEVAgent(config)
	if err != nil {
		t.Fatalf("Failed to create PEV agent: %v", err)
	}

	if agent == nil {
		t.Fatal("Expected non-nil agent")
	}
}

func TestPEVAgentRequiresModel(t *testing.T) {
	config := PEVAgentConfig{
		Tools:      []tools.Tool{},
		MaxRetries: 3,
	}

	_, err := CreatePEVAgent(config)
	if err == nil {
		t.Fatal("Expected error when model is nil")
	}
}

func TestParsePlanSteps(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "numbered list",
			input: `1. First step
2. Second step
3. Third step`,
			expected: []string{"First step", "Second step", "Third step"},
		},
		{
			name: "bullet points",
			input: `- First step
- Second step
- Third step`,
			expected: []string{"First step", "Second step", "Third step"},
		},
		{
			name: "mixed format",
			input: `1. First step
- Second step
* Third step`,
			expected: []string{"First step", "Second step", "Third step"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePlanSteps(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d steps, got %d", len(tt.expected), len(result))
			}
			for i, step := range result {
				if i < len(tt.expected) && step != tt.expected[i] {
					t.Errorf("Step %d: expected %q, got %q", i, tt.expected[i], step)
				}
			}
		})
	}
}

func TestParseVerificationResult(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    VerificationResult
	}{
		{
			name:        "valid JSON",
			input:       `{"is_successful": true, "reasoning": "All good"}`,
			expectError: false,
			expected:    VerificationResult{IsSuccessful: true, Reasoning: "All good"},
		},
		{
			name:        "JSON with surrounding text",
			input:       `Here is the result: {"is_successful": false, "reasoning": "Failed"} Done.`,
			expectError: false,
			expected:    VerificationResult{IsSuccessful: false, Reasoning: "Failed"},
		},
		{
			name:        "no JSON",
			input:       "This is not JSON",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result VerificationResult
			err := parseVerificationResult(tt.input, &result)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.IsSuccessful != tt.expected.IsSuccessful {
					t.Errorf("Expected IsSuccessful=%v, got %v", tt.expected.IsSuccessful, result.IsSuccessful)
				}
				if result.Reasoning != tt.expected.Reasoning {
					t.Errorf("Expected Reasoning=%q, got %q", tt.expected.Reasoning, result.Reasoning)
				}
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "exact length",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "long string",
			input:    "This is a very long string",
			maxLen:   10,
			expected: "This is a ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
