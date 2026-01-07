// Package main demonstrates a request-response API pattern with checkpoint-based
// interrupt handling for conversational agents.
//
// This example shows how to:
// 1. Build an HTTP API that handles conversational flows with interrupts
// 2. Automatically save checkpoints when interrupts occur (Issue #70 fix)
// 3. Detect and resume from interrupted states using checkpoint metadata
// 4. Use thread_id to maintain conversation state across requests
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/store/file"
)

// OrderState represents the state of an order processing conversation
type OrderState struct {
	SessionId   string    `json:"session_id"`
	UserInput   string    `json:"user_input"`
	ProductInfo string    `json:"product_info"`
	OrderId     string    `json:"order_id"`
	Price       float64   `json:"price"`
	OrderStatus string    `json:"order_status"`
	Message     string    `json:"message"`
	UpdateAt    time.Time `json:"update_at"`
	NextNode    string    `json:"next_node,omitempty"` // Tracks where to resume
	IsInterrupt bool      `json:"is_interrupt,omitempty"`
}

// Product catalog
var ProductCatalog = map[string]struct {
	Price float64
	Stock int
}{
	"iPhone 15":   {Price: 7999.00, Stock: 50},
	"MacBook Pro": {Price: 15999.00, Stock: 20},
	"AirPods":     {Price: 1299.00, Stock: 100},
	"iPad Air":    {Price: 4799.00, Stock: 30},
}

// BuildOrderGraph creates the order processing graph with interrupts
func BuildOrderGraph(store graph.CheckpointStore) *graph.CheckpointableRunnable[OrderState] {
	g := graph.NewCheckpointableStateGraphWithConfig[OrderState](graph.CheckpointConfig{
		Store:          store,
		AutoSave:       true,
		MaxCheckpoints: 10,
	})

	// Node 1: Order receive - extract product information
	g.AddNode("order_receive", "Order receive", func(ctx context.Context, state OrderState) (OrderState, error) {
		// If ProductInfo is already set and we're resuming, skip processing
		if state.ProductInfo != "" && state.IsInterrupt {
			// Already have product info, probably resuming - just pass through
			return state, nil
		}

		if state.UserInput == "" {
			state.Message = "请输入您要购买的产品"
			return state, nil
		}

		// String matching to find product
		var foundProduct string
		for productName := range ProductCatalog {
			if strings.Contains(state.UserInput, productName) {
				foundProduct = productName
				state.ProductInfo = productName
				break
			}
		}

		// Product not in catalog
		if foundProduct == "" {
			availableProducts := make([]string, 0, len(ProductCatalog))
			for name := range ProductCatalog {
				availableProducts = append(availableProducts, name)
			}
			state.Message = fmt.Sprintf("抱歉，我们的产品清单中没有您要购买的商品。\n可选产品有：%s",
				strings.Join(availableProducts, "、"))
			return state, nil
		}

		// Generate order ID
		state.OrderId = fmt.Sprintf("ORD%s%d", state.SessionId, time.Now().Unix())
		state.UpdateAt = time.Now()

		return state, nil
	})

	// Node 2: Inventory check
	g.AddNode("inventory_check", "Inventory check", func(ctx context.Context, state OrderState) (OrderState, error) {
		product, exists := ProductCatalog[state.ProductInfo]
		if !exists {
			state.Message = "产品信息异常，请重新下单"
			state.OrderId = ""
			return state, nil
		}

		// Check inventory
		if product.Stock <= 0 {
			state.Message = fmt.Sprintf("抱歉，%s 暂时无货，请选择其他产品", state.ProductInfo)
			state.OrderId = ""
			return state, nil
		}

		state.UpdateAt = time.Now()
		return state, nil
	})

	// Node 3: Price calculation
	g.AddNode("price_calculation", "Price calculation", func(ctx context.Context, state OrderState) (OrderState, error) {
		product, exists := ProductCatalog[state.ProductInfo]
		if !exists {
			return state, fmt.Errorf("产品信息异常")
		}

		state.Price = product.Price
		state.UpdateAt = time.Now()
		return state, nil
	})

	// Node 4: Payment processing with human-in-the-loop
	g.AddNode("payment_processing", "Payment processing", func(ctx context.Context, state OrderState) (OrderState, error) {
		state.OrderStatus = "待支付"

		// Human-in-the-loop: wait for user to confirm payment
		confirmMsg := fmt.Sprintf("您购买的 %s，价格：%.2f 元\n请确认是否支付？（回复`确认`以完成支付）",
			state.ProductInfo, state.Price)

		payInfo, err := graph.Interrupt(ctx, confirmMsg)
		if err != nil {
			// Set NextNode to indicate where to resume
			state.NextNode = "payment_processing"
			state.IsInterrupt = true
			return state, err
		}

		// Clear interrupt flag on resume
		state.IsInterrupt = false
		state.NextNode = ""

		// Check user confirmation
		payInfoStr, ok := payInfo.(string)
		if !ok || !strings.Contains(strings.ToLower(payInfoStr), "确认") {
			state.Message = "您已取消支付，订单已关闭"
			state.OrderStatus = "已取消"
			state.OrderId = ""
			return state, nil
		}

		state.OrderStatus = "已支付"
		state.UpdateAt = time.Now()
		return state, nil
	})

	// Node 5: Warehouse notification
	g.AddNode("warehouse_notify", "Warehouse notify", func(ctx context.Context, state OrderState) (OrderState, error) {
		// In a real application, this would call a warehouse notification API
		state.OrderStatus = "已发货"
		state.Message = fmt.Sprintf("您购买的 %s，价格：%.2f 元，已发货！\n订单号：%s",
			state.ProductInfo, state.Price, state.OrderId)
		state.UpdateAt = time.Now()

		return state, nil
	})

	// Set entry point
	g.SetEntryPoint("order_receive")

	// Define conditional edges
	g.AddConditionalEdge("order_receive", func(ctx context.Context, state OrderState) string {
		if state.ProductInfo == "" || state.OrderId == "" {
			return graph.END
		}
		return "inventory_check"
	})

	g.AddConditionalEdge("inventory_check", func(ctx context.Context, state OrderState) string {
		if state.OrderId == "" {
			return graph.END
		}
		return "price_calculation"
	})

	g.AddEdge("price_calculation", "payment_processing")

	g.AddConditionalEdge("payment_processing", func(ctx context.Context, state OrderState) string {
		if state.OrderId == "" || state.OrderStatus == "已取消" {
			return graph.END
		}
		return "warehouse_notify"
	})

	g.AddEdge("warehouse_notify", graph.END)

	runnable, err := g.CompileCheckpointable()
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	return runnable
}

// API handlers
type ChatRequest struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
}

type ChatResponse struct {
	Message     string `json:"message"`
	OrderStatus string `json:"order_status,omitempty"`
	IsInterrupt bool   `json:"is_interrupt,omitempty"`
	NeedsResume bool   `json:"needs_resume,omitempty"`
}

// Server holds the graph and store
type Server struct {
	Runnable *graph.CheckpointableRunnable[OrderState]
	Store    graph.CheckpointStore
}

// NewServer creates a new API server
func NewServer() (*Server, error) {
	// Create checkpoint store
	checkpointDir := "./checkpoints_api_demo"
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	store, err := file.NewFileCheckpointStore(checkpointDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint store: %w", err)
	}

	// Build graph with the store
	runnable := BuildOrderGraph(store)

	// Configure the runnable to use the store
	runnable.SetExecutionID("") // Will be set per-request using thread_id

	return &Server{
		Runnable: runnable,
		Store:    store,
	}, nil
}

// HandleChat handles chat requests
func (s *Server) HandleChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use session_id as thread_id
	threadID := req.SessionID
	if threadID == "" {
		threadID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}

	// Check if this thread has any checkpoints (to detect if we're resuming)
	checkpoints, err := s.Store.List(ctx, threadID)
	isResuming := false
	var latestCP *graph.Checkpoint

	if err == nil && len(checkpoints) > 0 {
		// Find the latest checkpoint
		latestCP = checkpoints[len(checkpoints)-1]
		// Check if the latest checkpoint has interrupt metadata
		if event, ok := latestCP.Metadata["event"].(string); ok && event == "step" {
			// The checkpoint was saved after a step completed
			// We need to check if there was an interrupt by looking at the state
			// State might be OrderState or map[string]any (from JSON deserialization)
			if state, ok := latestCP.State.(OrderState); ok && state.IsInterrupt {
				isResuming = true
			} else if m, ok := latestCP.State.(map[string]any); ok {
				// Check for is_interrupt in map
				if isInterrupt, ok := m["is_interrupt"].(bool); ok && isInterrupt {
					isResuming = true
				}
			}
		}
	}

	var config *graph.Config
	var initialState OrderState

	if isResuming && latestCP != nil {
		// RESUMING FROM INTERRUPT
		// Convert checkpoint state to OrderState
		if cpState, ok := latestCP.State.(OrderState); ok {
			initialState = cpState
		} else {
			// Handle case where state is map[string]any
			if m, ok := latestCP.State.(map[string]any); ok {
				// Convert map to OrderState (simplified - in production use proper JSON unmarshaling)
				initialState.SessionId = toString(m["session_id"])
				initialState.UserInput = req.Content
				initialState.ProductInfo = toString(m["product_info"])
				initialState.OrderId = toString(m["order_id"])
				initialState.Price = toFloat64(m["price"])
				initialState.OrderStatus = toString(m["order_status"])
				initialState.Message = toString(m["message"])
				initialState.UpdateAt = toTime(m["update_at"])
				initialState.NextNode = toString(m["next_node"])
				initialState.IsInterrupt = toBool(m["is_interrupt"])
			}
		}

		// Update with new user input
		initialState.UserInput = req.Content

		config = &graph.Config{
			Configurable: map[string]any{
				"thread_id": threadID,
			},
			ResumeValue: req.Content,
			ResumeFrom:  []string{latestCP.NodeName},
		}
	} else {
		// NEW REQUEST
		initialState = OrderState{
			SessionId: req.SessionID,
			UserInput: req.Content,
		}

		config = &graph.Config{
			Configurable: map[string]any{
				"thread_id": threadID,
			},
		}
	}

	// Set the execution ID to match the thread ID for checkpoint storage
	s.Runnable.SetExecutionID(threadID)

	// Execute the graph
	result, err := s.Runnable.InvokeWithConfig(ctx, initialState, config)

	var graphInterrupt *graph.GraphInterrupt
	if errors.As(err, &graphInterrupt) {
		// Graph was interrupted - state has been automatically saved by the fix (Issue #70)
		interruptState, ok := graphInterrupt.State.(OrderState)
		if !ok {
			log.Printf("Warning: Could not convert interrupt state to OrderState, got %T", graphInterrupt.State)
		}

		// Send interrupt response to client
		response := ChatResponse{
			Message:     fmt.Sprintf("%v", graphInterrupt.InterruptValue),
			IsInterrupt: true,
			NeedsResume: true,
		}

		if interruptState.ProductInfo != "" {
			response.OrderStatus = interruptState.OrderStatus
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Send normal response
	response := ChatResponse{
		Message:     result.Message,
		IsInterrupt: false,
		NeedsResume: false,
	}
	if result.OrderStatus != "" {
		response.OrderStatus = result.OrderStatus
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions for type conversion
func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toBool(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func toFloat64(v any) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if f, ok := v.(float32); ok {
		return float64(f)
	}
	return 0
}

func toTime(v any) time.Time {
	if v == nil {
		return time.Time{}
	}
	if s, ok := v.(string); ok {
		t, err := time.Parse(time.RFC3339Nano, s)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup HTTP handlers
	http.HandleFunc("/chat", server.HandleChat)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	port := "8080"
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	fmt.Printf("Try these examples:\n\n")
	fmt.Printf("1. Start new order:\n")
	fmt.Printf("   curl -X POST http://localhost:%s/chat \\\n", port)
	fmt.Printf("     -H 'Content-Type: application/json' \\\n")
	fmt.Printf("     -d '{\"session_id\":\"user123\",\"content\":\"我想买AirPods\"}'\n\n")
	fmt.Printf("2. Confirm payment (resume from interrupt):\n")
	fmt.Printf("   curl -X POST http://localhost:%s/chat \\\n", port)
	fmt.Printf("     -H 'Content-Type: application/json' \\\n")
	fmt.Printf("     -d '{\"session_id\":\"user123\",\"content\":\"确认\"}'\n\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
