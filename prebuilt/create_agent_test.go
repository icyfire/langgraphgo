package prebuilt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

func TestCreateAgentMap(t *testing.T) {
	mockLLM := &MockLLM{}
	inputTools := []tools.Tool{}
	systemMessage := "You are a helpful assistant."

	t.Run("Basic Agent Creation", func(t *testing.T) {
		agent, err := CreateAgentMap(mockLLM, inputTools, WithSystemMessage(systemMessage))
		assert.NoError(t, err)
		assert.NotNil(t, agent)
	})

	t.Run("Agent with State Modifier", func(t *testing.T) {
		mockLLM := &MockLLMWithInputCapture{}
		modifier := func(messages []llms.MessageContent) []llms.MessageContent {
			return append(messages, llms.TextParts(llms.ChatMessageTypeHuman, "Modified"))
		}

		agent, err := CreateAgentMap(mockLLM, inputTools, WithStateModifier(modifier))
		assert.NoError(t, err)

		_, err = agent.Invoke(context.Background(), map[string]any{"messages": []llms.MessageContent{}})
		assert.NoError(t, err)

		// Verify modifier was called (last message should be "Modified")
		assert.True(t, len(mockLLM.lastMessages) > 0)
		lastMsg := mockLLM.lastMessages[len(mockLLM.lastMessages)-1]
		assert.Equal(t, "Modified", lastMsg.Parts[0].(llms.TextContent).Text)
	})

	t.Run("Agent with System Message", func(t *testing.T) {
		mockLLM := &MockLLMWithInputCapture{}
		systemMsg := "You are a specialized bot."

		agent, err := CreateAgentMap(mockLLM, inputTools, WithSystemMessage(systemMsg))
		assert.NoError(t, err)

		_, err = agent.Invoke(context.Background(), map[string]any{"messages": []llms.MessageContent{}})
		assert.NoError(t, err)

		// Verify system message was prepended
		assert.True(t, len(mockLLM.lastMessages) > 0)
		firstMsg := mockLLM.lastMessages[0]
		assert.Equal(t, llms.ChatMessageTypeSystem, firstMsg.Role)
		assert.Equal(t, systemMsg, firstMsg.Parts[0].(llms.TextContent).Text)
	})
}

// Mock structures for testing
type MockLLM struct {
	llms.Model
}

func (m *MockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "Hello! I'm a mock AI.",
			},
		},
	}, nil
}

type MockLLMWithInputCapture struct {
	llms.Model
	lastMessages []llms.MessageContent
}

func (m *MockLLMWithInputCapture) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	m.lastMessages = messages
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "Response",
			},
		},
	}, nil
}
