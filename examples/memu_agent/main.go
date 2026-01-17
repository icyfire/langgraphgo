// Package main demonstrates how to use memU with a simple agent
// for advanced memory management in multi-turn conversations.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/memory"
	"github.com/smallnest/langgraphgo/memory/memu"
)

// AgentState represents the state of our chat agent
type AgentState struct {
	Messages    []string
	UserInput   string
	Response    string
	MemoryStats *memory.Stats
}

func main() {
	ctx := context.Background()

	// Get API key from environment
	memuAPIKey := os.Getenv("MEMU_API_KEY")
	if memuAPIKey == "" {
		log.Println("MEMU_API_KEY environment variable not set")
		log.Println("Get your API key at https://memu.so")
		log.Println("\nThen run:")
		log.Println("  export MEMU_API_KEY='your-api-key'")
		log.Println("  go run main.go")
		os.Exit(1)
	}

	memuBaseURL := os.Getenv("MEMU_BASE_URL")
	if memuBaseURL == "" {
		// Default to cloud API
		memuBaseURL = "https://127.0.0.1:8000"
	}

	// Initialize memU client
	// memU provides advanced memory management with:
	// - Hierarchical memory structure (Resource -> Item -> Category)
	// - Multimodal input support (conversations, documents, images)
	// - Dual retrieval methods (RAG for speed, LLM for deep understanding)
	memClient, err := memu.NewClient(memu.Config{
		BaseURL:        memuBaseURL,
		APIKey:         memuAPIKey,
		UserID:         "demo-user", // In production, use unique user IDs
		RetrieveMethod: "rag",       // Use "rag" for fast retrieval or "llm" for deep semantic search
	})
	if err != nil {
		log.Fatalf("Failed to create memU client: %v", err)
	}
	log.Println("memU client initialized successfully")

	// Create a simple agent graph with memory support
	// The agent will remember user preferences across conversations
	g := graph.NewStateGraph[AgentState]()

	// Agent node: processes user input and generates response
	g.AddNode("agent", "agent", func(ctx context.Context, state AgentState) (AgentState, error) {
		userMsg := state.UserInput

		// Add user message to memory
		msg := memory.NewMessage("user", userMsg)
		if err := memClient.AddMessage(ctx, msg); err != nil {
			log.Printf("Warning: failed to store message in memU: %v", err)
		}

		// Retrieve relevant context from memU
		memories, err := memClient.GetContext(ctx, userMsg)
		if err != nil {
			log.Printf("Warning: failed to get context from memU: %v", err)
			// Continue without memories
		} else if len(memories) > 0 {
			log.Printf("Retrieved %d memories from memU", len(memories))
		}

		// Generate response based on user input and memories
		response := generateResponse(userMsg, memories)

		// Add assistant response to memory
		respMsg := memory.NewMessage("assistant", response)
		if err := memClient.AddMessage(ctx, respMsg); err != nil {
			log.Printf("Warning: failed to store response in memU: %v", err)
		}

		state.Messages = append(state.Messages, userMsg, response)
		state.Response = response
		return state, nil
	})

	g.AddEdge("agent", graph.END)
	g.SetEntryPoint("agent")

	runnable, err := g.Compile()
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	// Run the agent with a multi-turn conversation
	fmt.Println("=== memU-Powered Agent Demo ===")
	fmt.Println("This agent remembers your preferences across conversations!")
	fmt.Println("Get your API key at https://memu.so")
	fmt.Println("\nCommands:")
	fmt.Println("  'stats' - Show memory statistics")
	fmt.Println("  'quit' - Exit the demo")
	fmt.Println("\nLet's chat! (Try telling me your name, preferences, etc.)\n")

	// Simulate a conversation to demonstrate memU capabilities
	conversations := []string{
		"My name is Alice and I love drinking coffee in the morning",
		"What's my name?",
		"What do I like to drink in the morning?",
		"I also enjoy working out in the evenings",
		"Tell me about my evening habits",
		"stats",
	}

	for _, userMsg := range conversations {
		// Handle special commands
		if userMsg == "quit" {
			fmt.Println("\nGoodbye!")
			break
		}

		if userMsg == "stats" {
			stats, err := memClient.GetStats(ctx)
			if err != nil {
				log.Printf("Failed to get stats: %v", err)
			} else {
				fmt.Printf("\n=== Memory Statistics ===\n")
				fmt.Printf("Categories: %d\n", stats.ActiveMessages)
				fmt.Printf("Total Items: %d\n", stats.TotalMessages)
				fmt.Printf("Total Tokens: %d\n", stats.TotalTokens)
			}
			fmt.Println()
			continue
		}

		// Process user message
		fmt.Printf("User: %s\n", userMsg)

		state := AgentState{
			UserInput: userMsg,
		}

		result, err := runnable.Invoke(ctx, state)
		if err != nil {
			log.Printf("Agent invocation failed: %v", err)
			continue
		}

		fmt.Printf("Agent: %s\n\n", result.Response)
	}

	fmt.Println("=== Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✅ Persistent memory across conversation turns")
	fmt.Println("✅ AI-powered memory extraction and organization")
	fmt.Println("✅ Context-aware retrieval based on queries")
	fmt.Println("✅ Hierarchical memory structure (Categories, Items)")
	fmt.Println("\nTry It Yourself:")
	fmt.Println("1. Get an API key at https://memu.so")
	fmt.Println("2. Set MEMU_API_KEY environment variable")
	fmt.Println("3. Run this example multiple times with different users")
	fmt.Println("4. See how memories persist and evolve!")
}

// generateResponse creates a response based on user input and retrieved memories
func generateResponse(userMsg string, memories []*memory.Message) string {
	var memoryContext strings.Builder
	if len(memories) > 0 {
		memoryContext.WriteString("\n[Based on my memory of you:]")
		for _, mem := range memories {
			if source, ok := mem.Metadata["source"]; ok {
				memoryContext.WriteString(fmt.Sprintf("\n  - %s: %s", source, mem.Content))
			}
		}
	}

	// Simple pattern-based responses (in production, use an LLM here)
	userMsgLower := strings.ToLower(userMsg)

	switch {
	case strings.Contains(userMsgLower, "name") && strings.Contains(userMsgLower, "my"):
		return "Based on our conversation history, I remember your name!" + memoryContext.String()

	case strings.Contains(userMsgLower, "prefer") || strings.Contains(userMsgLower, "like") || strings.Contains(userMsgLower, "love"):
		return fmt.Sprintf("I recall your preferences from our previous conversations:%s\nI'll keep this in mind for future interactions.", memoryContext.String())

	case strings.Contains(userMsgLower, "remember") || strings.Contains(userMsgLower, "memory"):
		return "I'm using memU to remember our conversations. I can store and retrieve information about you across sessions!" + memoryContext.String()

	case strings.Contains(userMsgLower, "hello") || strings.Contains(userMsgLower, "hi"):
		return "Hello! I'm a memU-powered agent. I can remember our conversations across sessions. Try telling me your preferences!"

	default:
		if len(memories) > 0 {
			return fmt.Sprintf("I remember:%s\n\nBased on this context, thanks for sharing: %s",
				memoryContext.String(), userMsg)
		}
		return fmt.Sprintf("I heard: %s. Tell me more about yourself so I can remember!", userMsg)
	}
}
