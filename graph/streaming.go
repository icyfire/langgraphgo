package graph

import (
	"context"
	"sync"
	"time"
)

// StreamMode defines the mode of streaming
type StreamMode string

const (
	// StreamModeValues emits the full state after each step
	StreamModeValues StreamMode = "values"
	// StreamModeUpdates emits the updates (deltas) from each node
	StreamModeUpdates StreamMode = "updates"
	// StreamModeMessages emits LLM messages/tokens (if available)
	StreamModeMessages StreamMode = "messages"
	// StreamModeDebug emits all events (default)
	StreamModeDebug StreamMode = "debug"
)

// StreamConfig configures streaming behavior
type StreamConfig struct {
	// BufferSize is the size of the event channel buffer
	BufferSize int

	// EnableBackpressure determines if backpressure handling is enabled
	EnableBackpressure bool

	// MaxDroppedEvents is the maximum number of events to drop before logging
	MaxDroppedEvents int

	// Mode specifies what kind of events to stream
	Mode StreamMode
}

// DefaultStreamConfig returns the default streaming configuration
func DefaultStreamConfig() StreamConfig {
	return StreamConfig{
		BufferSize:         1000,
		EnableBackpressure: true,
		MaxDroppedEvents:   100,
		Mode:               StreamModeDebug,
	}
}

// StreamResult contains the channels returned by streaming execution
type StreamResult[S any] struct {
	// Events channel receives StreamEvent objects in real-time
	Events <-chan StreamEvent[S]

	// Result channel receives the final result when execution completes
	Result <-chan S

	// Errors channel receives any errors that occur during execution
	Errors <-chan error

	// Done channel is closed when streaming is complete
	Done <-chan struct{}

	// Cancel function can be called to stop streaming
	Cancel context.CancelFunc
}

// StreamingListener implements NodeListener for streaming events
type StreamingListener[S any] struct {
	eventChan chan<- StreamEvent[S]
	config    StreamConfig
	mutex     sync.RWMutex

	droppedEvents int
	closed        bool
}

// NewStreamingListener creates a new streaming listener
func NewStreamingListener[S any](eventChan chan<- StreamEvent[S], config StreamConfig) *StreamingListener[S] {
	return &StreamingListener[S]{
		eventChan: eventChan,
		config:    config,
	}
}

// emitEvent sends an event to the channel handling backpressure
func (sl *StreamingListener[S]) emitEvent(event StreamEvent[S]) {
	// Check if listener is closed
	sl.mutex.RLock()
	if sl.closed {
		sl.mutex.RUnlock()
		return
	}
	sl.mutex.RUnlock()

	// Filter based on Mode
	if !sl.shouldEmit(event) {
		return
	}

	// Try to send event without blocking
	select {
	case sl.eventChan <- event:
		// Event sent successfully
	default:
		// Channel is full
		if sl.config.EnableBackpressure {
			sl.handleBackpressure()
		}
		// Drop the event if backpressure handling is disabled or channel is still full
	}
}

func (sl *StreamingListener[S]) shouldEmit(event StreamEvent[S]) bool {
	switch sl.config.Mode {
	case StreamModeDebug:
		return true
	case StreamModeValues:
		// Only emit OnGraphStep events (which contain full state)
		// We expect a custom event type or we rely on node complete if it returns full state?
		// For now, emit everything that looks like a state update
		return event.Event == NodeEventComplete || event.Event == EventChainEnd
	case StreamModeUpdates:
		// Emit node outputs
		return event.Event == NodeEventComplete || event.Event == EventChainEnd
	case StreamModeMessages:
		// Emit LLM events - this is tricky because generic S doesn't imply LLM events
		// But if the event metadata says it's LLM...
		return event.Event == EventLLMEnd || event.Event == EventLLMStart
	default:
		return true
	}
}

// OnNodeEvent implements the NodeListener interface
func (sl *StreamingListener[S]) OnNodeEvent(ctx context.Context, event NodeEvent, nodeName string, state S, err error) {
	streamEvent := StreamEvent[S]{
		Timestamp: time.Now(),
		NodeName:  nodeName,
		Event:     event,
		State:     state,
		Error:     err,
		Metadata:  make(map[string]any),
	}
	sl.emitEvent(streamEvent)
}

// Close marks the listener as closed to prevent sending to closed channels
func (sl *StreamingListener[S]) Close() {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	sl.closed = true
}

// handleBackpressure manages channel backpressure
func (sl *StreamingListener[S]) handleBackpressure() {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	sl.droppedEvents++
}

// GetDroppedEventsCount returns the number of dropped events
func (sl *StreamingListener[S]) GetDroppedEventsCount() int {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()
	return sl.droppedEvents
}

// StreamingRunnable wraps a ListenableRunnable with streaming capabilities
type StreamingRunnable[S any] struct {
	runnable *ListenableRunnable[S]
	config   StreamConfig
}

// NewStreamingRunnable creates a new streaming runnable
func NewStreamingRunnable[S any](runnable *ListenableRunnable[S], config StreamConfig) *StreamingRunnable[S] {
	return &StreamingRunnable[S]{
		runnable: runnable,
		config:   config,
	}
}

// NewStreamingRunnableWithDefaults creates a streaming runnable with default config
func NewStreamingRunnableWithDefaults[S any](runnable *ListenableRunnable[S]) *StreamingRunnable[S] {
	return NewStreamingRunnable(runnable, DefaultStreamConfig())
}

// Stream executes the graph with real-time event streaming
func (sr *StreamingRunnable[S]) Stream(ctx context.Context, initialState S) *StreamResult[S] {
	// Create channels
	eventChan := make(chan StreamEvent[S], sr.config.BufferSize)
	resultChan := make(chan S, 1)
	errorChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// Create cancellable context
	streamCtx, cancel := context.WithCancel(ctx)

	// Create streaming listener
	streamingListener := NewStreamingListener(eventChan, sr.config)

	// Add the streaming listener to all nodes
	// We add it globally using the graph
	sr.runnable.GetListenableGraph().AddGlobalListener(streamingListener)

	// Execute in goroutine
	go func() {
		defer func() {
			// First, close the streaming listener to prevent new events
			streamingListener.Close()

			// Clean up: remove listener
			sr.runnable.GetListenableGraph().RemoveGlobalListener(streamingListener)

			// Give a small delay for any in-flight listener calls to complete
			time.Sleep(10 * time.Millisecond)

			// Now safe to close channels
			close(eventChan)
			close(resultChan)
			close(errorChan)
			close(doneChan)
		}()

		// Execute the runnable
		result, err := sr.runnable.Invoke(streamCtx, initialState)

		// Send result or error
		if err != nil {
			select {
			case errorChan <- err:
			case <-streamCtx.Done():
			}
		} else {
			select {
			case resultChan <- result:
			case <-streamCtx.Done():
			}
		}
	}()

	return &StreamResult[S]{
		Events: eventChan,
		Result: resultChan,
		Errors: errorChan,
		Done:   doneChan,
		Cancel: cancel,
	}
}

// StreamingStateGraph[S any] extends ListenableStateGraph[S] with streaming capabilities
type StreamingStateGraph[S any] struct {
	*ListenableStateGraph[S]
	config StreamConfig
}

// NewStreamingStateGraph creates a new streaming state graph with type parameter
func NewStreamingStateGraph[S any]() *StreamingStateGraph[S] {
	baseGraph := NewListenableStateGraph[S]()
	return &StreamingStateGraph[S]{
		ListenableStateGraph: baseGraph,
		config:               DefaultStreamConfig(),
	}
}

// NewStreamingStateGraphWithConfig creates a streaming graph with custom config
func NewStreamingStateGraphWithConfig[S any](config StreamConfig) *StreamingStateGraph[S] {
	baseGraph := NewListenableStateGraph[S]()
	return &StreamingStateGraph[S]{
		ListenableStateGraph: baseGraph,
		config:               config,
	}
}

// CompileStreaming compiles the graph into a streaming runnable
func (g *StreamingStateGraph[S]) CompileStreaming() (*StreamingRunnable[S], error) {
	listenableRunnable, err := g.CompileListenable()
	if err != nil {
		return nil, err
	}

	return NewStreamingRunnable(listenableRunnable, g.config), nil
}

// SetStreamConfig updates the streaming configuration
func (g *StreamingStateGraph[S]) SetStreamConfig(config StreamConfig) {
	g.config = config
}

// GetStreamConfig returns the current streaming configuration
func (g *StreamingStateGraph[S]) GetStreamConfig() StreamConfig {
	return g.config
}

// StreamingExecutor[S] provides a high-level interface for streaming execution
type StreamingExecutor[S any] struct {
	runnable *StreamingRunnable[S]
}

// NewStreamingExecutor creates a new streaming executor
func NewStreamingExecutor[S any](runnable *StreamingRunnable[S]) *StreamingExecutor[S] {
	return &StreamingExecutor[S]{
		runnable: runnable,
	}
}

// ExecuteWithCallback executes the graph and calls the callback for each event
func (se *StreamingExecutor[S]) ExecuteWithCallback(
	ctx context.Context,
	initialState S,
	eventCallback func(event StreamEvent[S]),
	resultCallback func(result S, err error),
) error {

	streamResult := se.runnable.Stream(ctx, initialState)
	defer streamResult.Cancel()

	var finalResult S
	var finalError error
	resultReceived := false

	for {
		select {
		case event, ok := <-streamResult.Events:
			if !ok {
				// Events channel closed
				if resultReceived && resultCallback != nil {
					resultCallback(finalResult, finalError)
				}
				return finalError
			}
			if eventCallback != nil {
				eventCallback(event)
			}

		case result := <-streamResult.Result:
			finalResult = result
			resultReceived = true
			// Don't return immediately, wait for events channel to close

		case err := <-streamResult.Errors:
			finalError = err
			resultReceived = true
			// Don't return immediately, wait for events channel to close

		case <-streamResult.Done:
			if resultReceived && resultCallback != nil {
				resultCallback(finalResult, finalError)
			}
			return finalError

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// ExecuteAsync executes the graph asynchronously and returns immediately
func (se *StreamingExecutor[S]) ExecuteAsync(ctx context.Context, initialState S) *StreamResult[S] {
	return se.runnable.Stream(ctx, initialState)
}

// GetGraph returns a Exporter for the streaming runnable
func (sr *StreamingRunnable[S]) GetGraph() *Exporter[S] {
	return sr.runnable.GetGraph()
}

// GetTracer returns the tracer from the underlying runnable
func (sr *StreamingRunnable[S]) GetTracer() *Tracer {
	return sr.runnable.GetTracer()
}

// SetTracer sets the tracer on the underlying runnable
func (sr *StreamingRunnable[S]) SetTracer(tracer *Tracer) {
	sr.runnable.SetTracer(tracer)
}

// WithTracer returns a new StreamingRunnable with the given tracer
func (sr *StreamingRunnable[S]) WithTracer(tracer *Tracer) *StreamingRunnable[S] {
	newRunnable := sr.runnable.WithTracer(tracer)
	return &StreamingRunnable[S]{
		runnable: newRunnable,
		config:   sr.config,
	}
}
