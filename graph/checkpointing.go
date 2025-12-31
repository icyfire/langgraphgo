package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smallnest/langgraphgo/store"
	"github.com/smallnest/langgraphgo/store/file"
	"github.com/smallnest/langgraphgo/store/memory"
)

// Checkpoint represents a saved state at a specific point in execution
type Checkpoint = store.Checkpoint

// CheckpointStore defines the interface for checkpoint persistence
type CheckpointStore = store.CheckpointStore

// NewMemoryCheckpointStore creates a new in-memory checkpoint store
func NewMemoryCheckpointStore() CheckpointStore {
	return memory.NewMemoryCheckpointStore()
}

// NewFileCheckpointStore creates a new file-based checkpoint store
func NewFileCheckpointStore(path string) (CheckpointStore, error) {
	return file.NewFileCheckpointStore(path)
}

// CheckpointConfig configures checkpointing behavior
type CheckpointConfig struct {
	// Store is the checkpoint storage backend
	Store CheckpointStore

	// AutoSave enables automatic checkpointing after each node
	AutoSave bool

	// SaveInterval specifies how often to save (when AutoSave is false)
	SaveInterval time.Duration

	// MaxCheckpoints limits the number of checkpoints to keep
	MaxCheckpoints int
}

// DefaultCheckpointConfig returns a default checkpoint configuration
func DefaultCheckpointConfig() CheckpointConfig {
	return CheckpointConfig{
		Store:          NewMemoryCheckpointStore(),
		AutoSave:       true,
		SaveInterval:   30 * time.Second,
		MaxCheckpoints: 10,
	}
}

// CheckpointListener automatically creates checkpoints during execution
type CheckpointListener[S any] struct {
	store       CheckpointStore
	executionID string
	threadID    string
	autoSave    bool
}

// OnGraphStep is called after a step in the graph has completed and the state has been merged.
func (cl *CheckpointListener[S]) OnGraphStep(ctx context.Context, nodeName string, state any) {
	if cl.autoSave {
		if s, ok := state.(S); ok {
			cl.saveCheckpoint(ctx, nodeName, s)
		}
	}
}

// Implement other methods of CallbackHandler as no-ops
func (cl *CheckpointListener[S]) OnChainStart(context.Context, map[string]any, map[string]any, string, *string, []string, map[string]any) {
}
func (cl *CheckpointListener[S]) OnChainEnd(context.Context, map[string]any, string) {}
func (cl *CheckpointListener[S]) OnChainError(context.Context, error, string)        {}
func (cl *CheckpointListener[S]) OnToolStart(context.Context, map[string]any, string, string, *string, []string, map[string]any) {
}
func (cl *CheckpointListener[S]) OnToolEnd(context.Context, string, string)  {}
func (cl *CheckpointListener[S]) OnToolError(context.Context, error, string) {}
func (cl *CheckpointListener[S]) OnLLMStart(context.Context, map[string]any, []string, string, *string, []string, map[string]any) {
}
func (cl *CheckpointListener[S]) OnLLMEnd(context.Context, any, string)     {}
func (cl *CheckpointListener[S]) OnLLMError(context.Context, error, string) {}
func (cl *CheckpointListener[S]) OnRetrieverStart(context.Context, map[string]any, string, string, *string, []string, map[string]any) {
}
func (cl *CheckpointListener[S]) OnRetrieverEnd(context.Context, []any, string)   {}
func (cl *CheckpointListener[S]) OnRetrieverError(context.Context, error, string) {}

func (cl *CheckpointListener[S]) saveCheckpoint(ctx context.Context, nodeName string, state S) {
	// Get current version from existing checkpoints
	checkpoints, err := cl.store.List(ctx, cl.executionID)
	version := 1
	if err == nil && len(checkpoints) > 0 {
		// Get the latest version
		latest := checkpoints[len(checkpoints)-1]
		version = latest.Version + 1
	}

	metadata := map[string]any{
		"execution_id": cl.executionID,
		"event":        "step",
	}
	if cl.threadID != "" {
		metadata["thread_id"] = cl.threadID
	}

	checkpoint := &Checkpoint{
		ID:        generateCheckpointID(),
		NodeName:  nodeName,
		State:     state,
		Timestamp: time.Now(),
		Version:   version,
		Metadata:  metadata,
	}

	// Save checkpoint synchronously
	_ = cl.store.Save(ctx, checkpoint)
}

// CallbackHandler implementation for CheckpointListener is removed because CallbackHandler is untyped/legacy.
// We rely on NodeListener[S].

// CheckpointableStateGraph[S any] extends ListenableStateGraph[S] with checkpointing
type CheckpointableStateGraph[S any] struct {
	*ListenableStateGraph[S]
	config CheckpointConfig
}

// NewCheckpointableStateGraph creates a new checkpointable state graph with type parameter
func NewCheckpointableStateGraph[S any]() *CheckpointableStateGraph[S] {
	baseGraph := NewListenableStateGraph[S]()
	return &CheckpointableStateGraph[S]{
		ListenableStateGraph: baseGraph,
		config:               DefaultCheckpointConfig(),
	}
}

// NewCheckpointableStateGraphWithConfig creates a checkpointable graph with custom config
func NewCheckpointableStateGraphWithConfig[S any](config CheckpointConfig) *CheckpointableStateGraph[S] {
	baseGraph := NewListenableStateGraph[S]()
	return &CheckpointableStateGraph[S]{
		ListenableStateGraph: baseGraph,
		config:               config,
	}
}

// CompileCheckpointable compiles the graph into a checkpointable runnable
func (g *CheckpointableStateGraph[S]) CompileCheckpointable() (*CheckpointableRunnable[S], error) {
	listenableRunnable, err := g.CompileListenable()
	if err != nil {
		return nil, err
	}

	return NewCheckpointableRunnable(listenableRunnable, g.config), nil
}

// SetCheckpointConfig updates the checkpointing configuration
func (g *CheckpointableStateGraph[S]) SetCheckpointConfig(config CheckpointConfig) {
	g.config = config
}

// GetCheckpointConfig returns the current checkpointing configuration
func (g *CheckpointableStateGraph[S]) GetCheckpointConfig() CheckpointConfig {
	return g.config
}

// CheckpointableRunnable[S] wraps a ListenableRunnable[S] with checkpointing capabilities
type CheckpointableRunnable[S any] struct {
	runnable    *ListenableRunnable[S]
	config      CheckpointConfig
	executionID string
	listener    *CheckpointListener[S]
}

// NewCheckpointableRunnable creates a new checkpointable runnable from a listenable runnable
func NewCheckpointableRunnable[S any](runnable *ListenableRunnable[S], config CheckpointConfig) *CheckpointableRunnable[S] {
	executionID := generateExecutionID()
	cr := &CheckpointableRunnable[S]{
		runnable:    runnable,
		config:      config,
		executionID: executionID,
	}

	// Create checkpoint listener
	cr.listener = &CheckpointListener[S]{
		store:       cr.config.Store,
		executionID: executionID,
		threadID:    "",
		autoSave:    true,
	}

	// The listener will be added to config callbacks during invocation.

	return cr
}

// Invoke executes the graph with checkpointing support
func (cr *CheckpointableRunnable[S]) Invoke(ctx context.Context, initialState S) (S, error) {
	return cr.InvokeWithConfig(ctx, initialState, nil)
}

// InvokeWithConfig executes the graph with checkpointing support and config
func (cr *CheckpointableRunnable[S]) InvokeWithConfig(ctx context.Context, initialState S, config *Config) (S, error) {
	// Extract thread_id from config if present
	var threadID string
	if config != nil && config.Configurable != nil {
		if tid, ok := config.Configurable["thread_id"].(string); ok {
			threadID = tid
		}
	}

	// Update checkpoint listener with thread_id
	if cr.listener != nil {
		cr.listener.threadID = threadID
		cr.listener.autoSave = cr.config.AutoSave
	}

	// Add the listener to config callbacks
	if config == nil {
		config = &Config{}
	}
	config.Callbacks = append(config.Callbacks, cr.listener)

	return cr.runnable.InvokeWithConfig(ctx, initialState, config)
}

// Stream executes the graph with checkpointing and streaming support
func (cr *CheckpointableRunnable[S]) Stream(ctx context.Context, initialState S) <-chan StreamEvent[S] {
	return cr.runnable.Stream(ctx, initialState)
}

// StateSnapshot represents a snapshot of the graph state
type StateSnapshot struct {
	Values    any
	Next      []string
	Config    Config
	Metadata  map[string]any
	CreatedAt time.Time
	ParentID  string
}

// GetState retrieves the state for the given config
func (cr *CheckpointableRunnable[S]) GetState(ctx context.Context, config *Config) (*StateSnapshot, error) {
	var threadID string
	var checkpointID string

	if config != nil && config.Configurable != nil {
		if tid, ok := config.Configurable["thread_id"].(string); ok {
			threadID = tid
		}
		if cid, ok := config.Configurable["checkpoint_id"].(string); ok {
			checkpointID = cid
		}
	}

	// Default to current execution ID if thread_id not provided
	if threadID == "" {
		threadID = cr.executionID
	}

	var checkpoint *Checkpoint
	var err error

	if checkpointID != "" {
		checkpoint, err = cr.config.Store.Load(ctx, checkpointID)
	} else {
		// Get latest checkpoint for the thread
		checkpoints, err := cr.config.Store.List(ctx, threadID)
		if err != nil {
			return nil, fmt.Errorf("failed to list checkpoints: %w", err)
		}

		if len(checkpoints) == 0 {
			return nil, fmt.Errorf("no checkpoints found for thread %s", threadID)
		}

		// Get the latest checkpoint (highest version)
		checkpoint = checkpoints[0]
		for _, cp := range checkpoints {
			if cp.Version > checkpoint.Version {
				checkpoint = cp
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	if checkpoint == nil {
		return nil, fmt.Errorf("checkpoint not found")
	}

	// Return state snapshot
	next := []string{checkpoint.NodeName}
	if checkpoint.NodeName == "" {
		next = []string{}
	}
	return &StateSnapshot{
		Values: checkpoint.State,
		Next:   next,
		Config: Config{
			Configurable: map[string]any{
				"thread_id":     threadID,
				"checkpoint_id": checkpoint.ID,
			},
		},
		Metadata:  checkpoint.Metadata,
		CreatedAt: checkpoint.Timestamp,
	}, nil
}

// SaveCheckpoint manually saves a checkpoint at the current state
func (cr *CheckpointableRunnable[S]) SaveCheckpoint(ctx context.Context, nodeName string, state S) error {
	// Get current version to increment
	checkpoints, _ := cr.config.Store.List(ctx, cr.executionID)
	version := 1
	if len(checkpoints) > 0 {
		for _, cp := range checkpoints {
			if cp.Version >= version {
				version = cp.Version + 1
			}
		}
	}

	checkpoint := &Checkpoint{
		ID:        generateCheckpointID(),
		NodeName:  nodeName,
		State:     state,
		Timestamp: time.Now(),
		Version:   version,
		Metadata: map[string]any{
			"execution_id": cr.executionID,
			"source":       "manual_save",
			"saved_by":     nodeName,
		},
	}

	return cr.config.Store.Save(ctx, checkpoint)
}

// ListCheckpoints lists all checkpoints for the current execution
func (cr *CheckpointableRunnable[S]) ListCheckpoints(ctx context.Context) ([]*Checkpoint, error) {
	return cr.config.Store.List(ctx, cr.executionID)
}

// LoadCheckpoint loads a specific checkpoint
func (cr *CheckpointableRunnable[S]) LoadCheckpoint(ctx context.Context, checkpointID string) (*Checkpoint, error) {
	return cr.config.Store.Load(ctx, checkpointID)
}

// ClearCheckpoints removes all checkpoints for this execution
func (cr *CheckpointableRunnable[S]) ClearCheckpoints(ctx context.Context) error {
	return cr.config.Store.Clear(ctx, cr.executionID)
}

// UpdateState updates the state and saves a checkpoint.
func (cr *CheckpointableRunnable[S]) UpdateState(ctx context.Context, config *Config, asNode string, values S) (*Config, error) {
	var threadID string

	if config != nil && config.Configurable != nil {
		if tid, ok := config.Configurable["thread_id"].(string); ok {
			threadID = tid
		}
	}

	if threadID == "" {
		threadID = cr.executionID
	}

	// Get current state from config if available
	var currentState S

	if config != nil {
		snapshot, err := cr.GetState(ctx, config)
		if err == nil && snapshot != nil {
			if s, ok := snapshot.Values.(S); ok {
				currentState = s
			}
		}
	}

	// If current state is still nil (e.g., no checkpoints), initialize from schema
	if any(currentState) == nil && cr.runnable.graph.Schema != nil {
		currentState = cr.runnable.graph.Schema.Init()
	}

	// Apply update using Schema if available
	var newState S
	if cr.runnable.graph.Schema != nil {
		var err error
		newState, err = cr.runnable.graph.Schema.Update(currentState, values)
		if err != nil {
			return nil, fmt.Errorf("failed to update state with schema: %w", err)
		}
	} else {
		// Default: Replace
		newState = values
	}

	// Get max version
	checkpoints, _ := cr.config.Store.List(ctx, threadID)
	version := 1
	for _, cp := range checkpoints {
		if cp.Version >= version {
			version = cp.Version + 1
		}
	}

	// Create new checkpoint
	checkpoint := &Checkpoint{
		ID:        generateCheckpointID(),
		NodeName:  asNode,
		State:     newState,
		Timestamp: time.Now(),
		Version:   version,
		Metadata: map[string]any{
			"execution_id": threadID,
			"source":       "update_state",
			"updated_by":   asNode,
		},
	}

	if err := cr.config.Store.Save(ctx, checkpoint); err != nil {
		return nil, err
	}

	return &Config{
		Configurable: map[string]any{
			"thread_id":     threadID,
			"checkpoint_id": checkpoint.ID,
		},
	}, nil
}

// GetExecutionID returns the current execution ID
func (cr *CheckpointableRunnable[S]) GetExecutionID() string {
	return cr.executionID
}

// SetExecutionID sets a new execution ID
func (cr *CheckpointableRunnable[S]) SetExecutionID(executionID string) {
	cr.executionID = executionID
	if cr.listener != nil {
		cr.listener.executionID = executionID
	}
}

// GetTracer returns the tracer from the underlying runnable
func (cr *CheckpointableRunnable[S]) GetTracer() *Tracer {
	return cr.runnable.GetTracer()
}

// SetTracer sets the tracer on the underlying runnable
func (cr *CheckpointableRunnable[S]) SetTracer(tracer *Tracer) {
	cr.runnable.SetTracer(tracer)
}

// WithTracer returns a new CheckpointableRunnable with the given tracer
func (cr *CheckpointableRunnable[S]) WithTracer(tracer *Tracer) *CheckpointableRunnable[S] {
	newRunnable := cr.runnable.WithTracer(tracer)
	return &CheckpointableRunnable[S]{
		runnable:    newRunnable,
		config:      cr.config,
		executionID: cr.executionID,
		listener:    cr.listener,
	}
}

// GetGraph returns the underlying graph
func (cr *CheckpointableRunnable[S]) GetGraph() *ListenableStateGraph[S] {
	return cr.runnable.GetListenableGraph()
}

// Helper functions
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

func generateCheckpointID() string {
	return fmt.Sprintf("checkpoint_%s", uuid.New().String())
}
