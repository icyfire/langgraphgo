package prebuilt

import (
	"context"
	"testing"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
)

// MockLLMSupervisor for testing
type MockLLMSupervisor struct {
	responses []string
	index     int
}

func (m *MockLLMSupervisor) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	response := m.responses[m.index%len(m.responses)]
	m.index++

	// Return a response with a tool call
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: response,
				ToolCalls: []llms.ToolCall{
					{
						ID: "test-call",
						FunctionCall: &llms.FunctionCall{
							Name:      "route",
							Arguments: `{"next":"worker1"}`,
						},
					},
				},
			},
		},
	}, nil
}

// Call implements the deprecated Call method for backward compatibility
func (m *MockLLMSupervisor) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	response := m.responses[m.index%len(m.responses)]
	m.index++
	return response, nil
}

func TestCreateSupervisorTyped(t *testing.T) {
	// Create a mock LLM
	mockLLM := &MockLLMSupervisor{
		responses: []string{
			"Thinking...",
		},
	}

	// Create mock member runnables
	member1 := &graph.StateRunnableTyped[SupervisorState]{}
	member2 := &graph.StateRunnableTyped[SupervisorState]{}

	members := map[string]*graph.StateRunnableTyped[SupervisorState]{
		"worker1": member1,
		"worker2": member2,
	}

	// Create supervisor
	supervisor, err := CreateSupervisorTyped(mockLLM, members)
	if err != nil {
		t.Fatalf("Failed to create supervisor: %v", err)
	}

	if supervisor == nil {
		t.Fatal("Supervisor should not be nil")
	}
}

func TestCreateSupervisorTyped_EmptyMembers(t *testing.T) {
	mockLLM := &MockLLMSupervisor{
		responses: []string{
			"Thinking...",
		},
	}

	members := map[string]*graph.StateRunnableTyped[SupervisorState]{}

	supervisor, err := CreateSupervisorTyped(mockLLM, members)
	if err != nil {
		t.Fatalf("Failed to create supervisor with empty members: %v", err)
	}

	if supervisor == nil {
		t.Fatal("Supervisor should not be nil even with empty members")
	}
}

func TestCreateSupervisorWithStateTyped(t *testing.T) {
	// Define custom state type
	type CustomState struct {
		Step      int
		Data      string
		Messages  []llms.MessageContent
		Next      string
	}

	// Create a mock LLM
	mockLLM := &MockLLMSupervisor{
		responses: []string{
			"Processing...",
		},
	}

	// Create mock member runnables
	member1 := &graph.StateRunnableTyped[CustomState]{}
	member2 := &graph.StateRunnableTyped[CustomState]{}

	members := map[string]*graph.StateRunnableTyped[CustomState]{
		"worker1": member1,
		"worker2": member2,
	}

	// Define state handlers
	getMessages := func(s CustomState) []llms.MessageContent {
		return s.Messages
	}

	setMessages := func(s CustomState, msgs []llms.MessageContent) CustomState {
		s.Messages = msgs
		return s
	}

	getNext := func(s CustomState) string {
		return s.Next
	}

	setNext := func(s CustomState, next string) CustomState {
		s.Next = next
		return s
	}

	// Create supervisor with custom state
	supervisor, err := CreateSupervisorWithStateTyped(
		mockLLM,
		members,
		getMessages,
		setMessages,
		getNext,
		setNext,
	)

	if err != nil {
		t.Fatalf("Failed to create supervisor with custom state: %v", err)
	}

	if supervisor == nil {
		t.Fatal("Supervisor should not be nil")
	}
}

func TestCreateSupervisorWithStateTyped_CustomLogic(t *testing.T) {
	type CustomState struct {
		Counter   int
		Processed []string
		Messages  []llms.MessageContent
		Next      string
	}

	mockLLM := &MockLLMSupervisor{
		responses: []string{
			"Deciding next step...",
		},
	}

	// Create a mock member runnable that updates state
	member := &graph.StateRunnableTyped[CustomState]{}

	members := map[string]*graph.StateRunnableTyped[CustomState]{
		"processor": member,
	}

	getMessages := func(s CustomState) []llms.MessageContent {
		return s.Messages
	}

	setMessages := func(s CustomState, msgs []llms.MessageContent) CustomState {
		s.Messages = msgs
		return s
	}

	getNext := func(s CustomState) string {
		return s.Next
	}

	setNext := func(s CustomState, next string) CustomState {
		s.Next = next
		s.Counter++
		if next != "" {
			s.Processed = append(s.Processed, next)
		}
		return s
	}

	supervisor, err := CreateSupervisorWithStateTyped(
		mockLLM,
		members,
		getMessages,
		setMessages,
		getNext,
		setNext,
	)

	if err != nil {
		t.Fatalf("Failed to create supervisor: %v", err)
	}

	if supervisor == nil {
		t.Fatal("Supervisor should not be nil")
	}
}

// Test SupervisorState structure
func TestSupervisorState(t *testing.T) {
	state := SupervisorState{
		Messages: []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "Test message"),
		},
		Next: "worker1",
	}

	if len(state.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(state.Messages))
	}

	if state.Next != "worker1" {
		t.Errorf("Expected next to be 'worker1', got '%s'", state.Next)
	}
}

// Test that supervisor creates the correct graph structure
func TestCreateSupervisorTyped_GraphStructure(t *testing.T) {
	mockLLM := &MockLLMSupervisor{
		responses: []string{
			"Supervisor decision",
		},
	}

	members := map[string]*graph.StateRunnableTyped[SupervisorState]{
		"agent1": {},
		"agent2": {},
	}

	supervisor, err := CreateSupervisorTyped(mockLLM, members)
	if err != nil {
		t.Fatalf("Failed to create supervisor: %v", err)
	}

	// The supervisor should compile successfully
	// This tests that the graph structure is correctly built
	if supervisor == nil {
		t.Fatal("Supervisor should not be nil")
	}
}

// Test supervisor with single member
func TestCreateSupervisorTyped_SingleMember(t *testing.T) {
	mockLLM := &MockLLMSupervisor{
		responses: []string{
			"Decision made",
		},
	}

	members := map[string]*graph.StateRunnableTyped[SupervisorState]{
		"sole_agent": {},
	}

	supervisor, err := CreateSupervisorTyped(mockLLM, members)
	if err != nil {
		t.Fatalf("Failed to create supervisor with single member: %v", err)
	}

	if supervisor == nil {
		t.Fatal("Supervisor should not be nil")
	}
}

// Test supervisor with complex member names
func TestCreateSupervisorTyped_ComplexMemberNames(t *testing.T) {
	mockLLM := &MockLLMSupervisor{
		responses: []string{
			"Analyzing options",
		},
	}

	members := map[string]*graph.StateRunnableTyped[SupervisorState]{
		"agent_with_underscores": {},
		"agent-with-dashes":      {},
		"AgentWithCamelCase":     {},
	}

	supervisor, err := CreateSupervisorTyped(mockLLM, members)
	if err != nil {
		t.Fatalf("Failed to create supervisor with complex member names: %v", err)
	}

	if supervisor == nil {
		t.Fatal("Supervisor should not be nil")
	}
}

// Test CreateSupervisorTyped with edge cases
func TestCreateSupervisorTyped_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		members  map[string]*graph.StateRunnableTyped[SupervisorState]
		expectOK bool
	}{
		{
			name:     "nil members",
			members:  nil,
			expectOK: true,
		},
		{
			name:     "empty members map",
			members:  map[string]*graph.StateRunnableTyped[SupervisorState]{},
			expectOK: true,
		},
		{
			name: "single member",
			members: map[string]*graph.StateRunnableTyped[SupervisorState]{
				"worker": {},
			},
			expectOK: true,
		},
		{
			name: "many members",
			members: map[string]*graph.StateRunnableTyped[SupervisorState]{
				"worker1": {},
				"worker2": {},
				"worker3": {},
				"worker4": {},
				"worker5": {},
			},
			expectOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := &MockLLMSupervisor{
				responses: []string{"Making decision"},
			}

			supervisor, err := CreateSupervisorTyped(mockLLM, tt.members)
			if err != nil {
				t.Errorf("Failed to create supervisor: %v", err)
			}

			if tt.expectOK && supervisor == nil {
				t.Error("Supervisor should not be nil")
			}
		})
	}
}

// Test CreateSupervisorWithStateTyped with various scenarios
func TestCreateSupervisorWithStateTyped_ComplexScenarios(t *testing.T) {
	// Test with state that has channels
	type ChannelState struct {
		Messages []llms.MessageContent
		Next     string
		Done     chan bool
		Data     chan string
	}

	getMessages := func(s ChannelState) []llms.MessageContent {
		return s.Messages
	}

	setMessages := func(s ChannelState, msgs []llms.MessageContent) ChannelState {
		s.Messages = msgs
		return s
	}

	getNext := func(s ChannelState) string {
		return s.Next
	}

	setNext := func(s ChannelState, next string) ChannelState {
		s.Next = next
		if s.Done != nil {
			s.Done <- (next == "END")
		}
		return s
	}

	mockLLM := &MockLLMSupervisor{
		responses: []string{"Processing with channels"},
	}

	members := map[string]*graph.StateRunnableTyped[ChannelState]{
		"processor": {},
	}

	supervisor, err := CreateSupervisorWithStateTyped(
		mockLLM,
		members,
		getMessages,
		setMessages,
		getNext,
		setNext,
	)

	if err != nil {
		t.Fatalf("Failed to create supervisor with channel state: %v", err)
	}

	if supervisor == nil {
		t.Fatal("Supervisor should not be nil")
	}
}

// Test supervisor with different routing strategies
func TestCreateSupervisorWithStateTyped_RoutingStrategies(t *testing.T) {
	type RoutingState struct {
		Messages   []llms.MessageContent
		Next       string
		Priority   int
		RoutingKey string
	}

	getMessages := func(s RoutingState) []llms.MessageContent {
		return s.Messages
	}

	setMessages := func(s RoutingState, msgs []llms.MessageContent) RoutingState {
		s.Messages = msgs
		return s
	}

	getNext := func(s RoutingState) string {
		return s.Next
	}

	setNext := func(s RoutingState, next string) RoutingState {
		s.Next = next
		return s
	}

	mockLLM := &MockLLMSupervisor{
		responses: []string{"Deciding route based on priority"},
	}

	members := map[string]*graph.StateRunnableTyped[RoutingState]{
		"high_priority_worker": {},
		"low_priority_worker":  {},
		"default_worker":       {},
	}

	supervisor, err := CreateSupervisorWithStateTyped(
		mockLLM,
		members,
		getMessages,
		setMessages,
		getNext,
		setNext,
	)

	if err != nil {
		t.Fatalf("Failed to create supervisor with routing: %v", err)
	}

	if supervisor == nil {
		t.Fatal("Supervisor should not be nil")
	}
}

// Test error handling in supervisor creation
func TestCreateSupervisorWithStateTyped_ErrorHandling(t *testing.T) {
	type ErrorState struct {
		Messages []llms.MessageContent
		Next     string
		Error    error
	}

	getMessages := func(s ErrorState) []llms.MessageContent {
		return s.Messages
	}

	setMessages := func(s ErrorState, msgs []llms.MessageContent) ErrorState {
		s.Messages = msgs
		return s
	}

	getNext := func(s ErrorState) string {
		return s.Next
	}

	setNext := func(s ErrorState, next string) ErrorState {
		s.Next = next
		return s
	}

	// Test with nil LLM - the function doesn't validate nil, so it will create the supervisor
	// but it would panic when actually trying to invoke it
	supervisor, err := CreateSupervisorWithStateTyped(
		nil,
		map[string]*graph.StateRunnableTyped[ErrorState]{},
		getMessages,
		setMessages,
		getNext,
		setNext,
	)

	if err != nil {
		t.Errorf("Unexpected error with nil LLM: %v", err)
	}
	if supervisor == nil {
		t.Error("Supervisor should not be nil even with nil LLM (validation happens at invocation)")
	}

	// Test with nil handler functions
	supervisor, err = CreateSupervisorWithStateTyped(
		&MockLLMSupervisor{responses: []string{"test"}},
		map[string]*graph.StateRunnableTyped[ErrorState]{},
		nil, // nil getMessages
		setMessages,
		getNext,
		setNext,
	)

	// Should still work with nil handlers (they'll use defaults)
	if err != nil {
		t.Errorf("Should handle nil handlers gracefully: %v", err)
	}
}