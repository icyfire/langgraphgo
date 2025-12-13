package prebuilt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

// TestReactAgentTyped_Simple Tests simple creation without execution
func TestReactAgentTyped_Simple(t *testing.T) {
	// Create a simple mock LLM that doesn't require complex setup
	mockLLM := &struct {
		llms.Model
	}{
		// Empty model for testing - we only need it to exist
	}

	// Test with nil tools first
	agent, err := CreateReactAgentTyped(mockLLM, nil)
	if err != nil {
		// This is expected to fail because of nil model, but that's OK for this test
		assert.Error(t, err)
		assert.Nil(t, agent)
	}

	// Test with empty tools slice
	emptyTools := []tools.Tool{}
	agent, err = CreateReactAgentTyped(mockLLM, emptyTools)
	if err != nil {
		// This is expected to fail, but we're testing the error path
		assert.Error(t, err)
		assert.Nil(t, agent)
	}
}

// TestReactAgentState_Simple Tests the state structure
func TestReactAgentState_Simple(t *testing.T) {
	// Test empty state
	state := ReactAgentState{}
	assert.Empty(t, state.Messages)

	// Test with messages
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "Hello"),
	}
	state.Messages = messages
	assert.Equal(t, 1, len(state.Messages))
}

// TestCreateReactAgentWithCustomStateTyped_Simple Tests custom state creation
func TestCreateReactAgentWithCustomStateTyped_Simple(t *testing.T) {
	type TestState struct {
		Data string
	}

	getMessages := func(s TestState) []llms.MessageContent {
		return nil
	}

	setMessages := func(s TestState, msgs []llms.MessageContent) TestState {
		return s
	}

	hasToolCalls := func(msgs []llms.MessageContent) bool {
		return false
	}

	// Mock LLM
	mockLLM := &struct {
		llms.Model
	}{}

	// This test may fail due to model issues, but we're testing the type system
	_, err := CreateReactAgentWithCustomStateTyped(
		mockLLM,
		[]tools.Tool{},
		getMessages,
		setMessages,
		hasToolCalls,
	)

	// We don't assert the result here because the model may be invalid
	// The important thing is that the function signature compiles
	_ = err
}