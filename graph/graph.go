package graph

import (
	"context"
	"errors"
	"fmt"
)

// END is a special constant used to represent the end node in the graph.
const END = "END"

var (
	// ErrEntryPointNotSet is returned when the entry point of the graph is not set.
	ErrEntryPointNotSet = errors.New("entry point not set")

	// ErrNodeNotFound is returned when a node is not found in the graph.
	ErrNodeNotFound = errors.New("node not found")

	// ErrNoOutgoingEdge is returned when no outgoing edge is found for a node.
	ErrNoOutgoingEdge = errors.New("no outgoing edge found for node")
)

// GraphInterrupt is returned when execution is interrupted by configuration or dynamic interrupt
type GraphInterrupt struct {
	// Node that caused the interruption
	Node string
	// State at the time of interruption
	State any
	// NextNodes that would have been executed if not interrupted
	NextNodes []string
	// InterruptValue is the value provided by the dynamic interrupt (if any)
	InterruptValue any
}

func (e *GraphInterrupt) Error() string {
	if e.InterruptValue != nil {
		return fmt.Sprintf("graph interrupted at node %s with value: %v", e.Node, e.InterruptValue)
	}
	return fmt.Sprintf("graph interrupted at node %s", e.Node)
}

// Interrupt pauses execution and waits for input.
// If resuming, it returns the value provided in the resume command.
func Interrupt(ctx context.Context, value any) (any, error) {
	if resumeVal := GetResumeValue(ctx); resumeVal != nil {
		return resumeVal, nil
	}
	return nil, &NodeInterrupt{Value: value}
}

// Edge represents an edge in the graph.
type Edge struct {
	// From is the name of the node from which the edge originates.
	From string

	// To is the name of the node to which the edge points.
	To string
}

// RetryPolicy defines how to handle node failures
type RetryPolicy struct {
	MaxRetries      int
	BackoffStrategy BackoffStrategy
	RetryableErrors []string
}

// BackoffStrategy defines different backoff strategies
type BackoffStrategy int

const (
	FixedBackoff BackoffStrategy = iota
	ExponentialBackoff
	LinearBackoff
)

// Runnable is an alias for StateRunnable[map[string]any] for convenience.
type Runnable = StateRunnable[map[string]any]

// StateGraphMap is an alias for StateGraph[map[string]any] for convenience.
// Use NewStateGraph[map[string]any]() or NewStateGraph[S]() for other types.
type StateGraphMap = StateGraph[map[string]any]

// ListenableStateGraphMap is an alias for ListenableStateGraph[map[string]any].
type ListenableStateGraphMap = ListenableStateGraph[map[string]any]

// ListenableRunnableMap is an alias for ListenableRunnable[map[string]any].
type ListenableRunnableMap = ListenableRunnable[map[string]any]

// NewMessageGraph creates a new instance of StateGraph[map[string]any] with a default schema
// that handles "messages" using the AddMessages reducer.
// This is the recommended constructor for chat-based agents that use
// map[string]any as state with a "messages" key.
//
// Deprecated: Use NewStateGraph[MessageState]() for type-safe state management.
func NewMessageGraph() *StateGraph[map[string]any] {
	g := NewStateGraph[map[string]any]()

	// Initialize default schema for message handling
	schema := NewMapSchema()
	schema.RegisterReducer("messages", AddMessages)

	g.SetSchema(schema)

	return g
}
