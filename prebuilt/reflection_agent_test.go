package prebuilt

import (
	"testing"
)

func TestCreateReflectionAgentMap(t *testing.T) {
	mockLLM := &MockReflectionLLM{
		responses: []string{
			"Initial response",
			"**Strengths:** Good. **Weaknesses:** None.",
		},
	}
	config := ReflectionAgentConfig{Model: mockLLM, MaxIterations: 2}
	agent, err := CreateReflectionAgentMap(config)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if agent == nil {
		t.Fatal("Agent is nil")
	}
}
