# Listenable StateGraph with Generics

This example demonstrates how to use **type-safe generic StateGraph** with **event listeners** in LangGraphGo. Listenable graphs provide real-time monitoring and event handling capabilities for your workflows.

## Overview

A Listenable StateGraph extends the generic StateGraph by adding:
- **Event-driven architecture** - Listen to graph execution events
- **Real-time monitoring** - Track node execution, errors, and state changes
- **Streaming support** - Stream execution events in real-time
- **Flexible listener types** - Global listeners and node-specific listeners

## Key Features

✅ **Type Safety** - Full compile-time type safety with generics
✅ **Event Listening** - Monitor graph execution in real-time
✅ **Multiple Listeners** - Add global or node-specific listeners
✅ **Streaming Events** - Stream execution events via channels
✅ **Custom Events** - Create custom event handlers
✅ **Zero Runtime Overhead** - Generics are compile-time only

## Example: Counter with Event Monitoring

The example demonstrates a counter workflow that:
- Increments a counter 5 times
- Logs each node execution
- Tracks progress
- Streams events in real-time

### State Definition

```go
type CounterState struct {
    Count int      `json:"count"`     // Current counter value
    Name  string   `json:"name"`      // Workflow name
    Logs  []string `json:"logs"`      // Execution logs
}
```

### Event Listeners

#### 1. Global Event Logger

```go
type EventLogger struct{}

func (l *EventLogger) OnNodeEvent(ctx context.Context, event graph.NodeEvent,
                                 nodeName string, state CounterState, err error) {
    switch event {
    case graph.NodeEventStart:
        fmt.Printf("🔵 Node '%s' started (count=%d)\n", nodeName, state.Count)
    case graph.NodeEventComplete:
        fmt.Printf("🟢 Node '%s' completed (count=%d)\n", nodeName, state.Count)
    case graph.NodeEventError:
        fmt.Printf("🔴 Node '%s' failed: %v\n", nodeName, err)
    }
}
```

#### 2. Progress Tracker

```go
type ProgressTracker struct {
    totalNodes int
    completed  int
}

func (p *ProgressTracker) OnNodeEvent(ctx context.Context, event graph.NodeEvent,
                                     nodeName string, state CounterState, err error) {
    if event == graph.NodeEventComplete {
        p.completed++
        progress := float64(p.completed) / float64(p.totalNodes) * 100
        fmt.Printf("📊 Progress: %.1f%% (%d/%d)\n", progress, p.completed, p.totalNodes)
    }
}
```

#### 3. Node-Specific Listener

```go
// Special listener for the increment node
incrementListener := graph.NodeListenerTypedFunc[CounterState](
    func(ctx context.Context, event graph.NodeEvent, nodeName string,
         state CounterState, err error) {
        if event == graph.NodeEventComplete {
            fmt.Printf("✨ Special notification: Count is now %d!\n", state.Count)
        }
    },
)
```

### Creating the Listenable Graph

```go
// Create a typed listenable state graph
workflow := graph.NewListenableStateGraphTyped[CounterState]()

// Add global listeners
workflow.AddGlobalListener(&EventLogger{})
workflow.AddGlobalListener(&ProgressTracker{})

// Add nodes
incrementNode := workflow.AddNode("increment", "Increment counter",
    func(ctx context.Context, state CounterState) (CounterState, error) {
        state.Count++
        logMsg := fmt.Sprintf("Incremented count to %d", state.Count)
        state.Logs = append(state.Logs, logMsg)
        time.Sleep(500 * time.Millisecond) // Simulate work
        return state, nil
    })

// Add node-specific listener
incrementNode.AddListener(incrementListener)
```

### Compilation and Execution

```go
// Compile the listenable graph
runnable, err := workflow.CompileListenable()
if err != nil {
    log.Fatalf("Failed to compile graph: %v", err)
}

// Execute the graph
finalState, err := runnable.Invoke(context.Background(), initialState)
if err != nil {
    log.Fatalf("Graph execution failed: %v", err)
}
```

### Streaming Execution

```go
// Create streaming listener
streamingListener := &StreamingCounterListener{}
workflow.AddGlobalListener(streamingListener)

// Compile for streaming
streamingRunnable, err := workflow.CompileListenable()

// Stream events
eventChan := streamingRunnable.Stream(context.Background(), initialState)

// Process events
for event := range eventChan {
    switch event.Event {
    case graph.EventChainStart:
        fmt.Printf("🟢 Stream: Chain started\n")
    case graph.NodeEventStart:
        fmt.Printf("🔵 Stream: Node '%s' started\n", event.NodeName)
    case graph.NodeEventComplete:
        fmt.Printf("🟢 Stream: Node '%s' completed\n", event.NodeName)
    case graph.EventChainEnd:
        fmt.Printf("🔴 Stream: Chain ended\n")
    }
}
```

## Running the Example

```bash
cd examples/generic_state_graph_listenable
go run listenable_example.go
```

## Expected Output

```
🔧 Compiling the graph...

🚀 Starting graph execution...
[17:36:33] 🔵 Node 'increment' started (count=0)
  ✨ Special notification: Count is now 1!
[17:36:33] 🟢 Node 'increment' completed (count=1)
📊 Progress: 20.0% (1/5)
[17:36:34] 🔵 Node 'increment' started (count=1)
  ✨ Special notification: Count is now 2!
[17:36:34] 🟢 Node 'increment' completed (count=2)
📊 Progress: 40.0% (2/5)
...
[17:36:36] 🟢 Node 'print' completed (count=5)

✅ Execution completed successfully!
Final state: {Count:5 Name:TypedCounter Logs:[Incremented count to 1 ...]}

--- Streaming Example ---
🎬 Starting streaming execution...
📡 Receiving events:
[17:36:33.684] 🟢 Stream: Chain started
[17:36:33.684] 🔵 Stream: Node 'increment' started
[17:36:34.185] 🟢 Stream: Node 'increment' completed (count=1)
...
[17:36:36.189] 🔴 Stream: Chain ended
```

## Event Types

### Node Events
- `NodeEventStart` - Node execution started
- `NodeEventComplete` - Node execution completed successfully
- `NodeEventError` - Node execution failed with error

### Chain Events
- `EventChainStart` - Graph execution started
- `EventChainEnd` - Graph execution completed

### Stream Events
Stream events contain additional metadata:
```go
type StreamEventTyped[S any] struct {
    Timestamp time.Time        // When the event occurred
    Event     graph.NodeEvent  // Event type
    NodeName  string           // Name of the node
    State     S                // Current state
    Error     error            // Error if any
}
```

## API Reference

### Creating a Listenable Graph

```go
workflow := graph.NewListenableStateGraphTyped[YourStateType]()
```

### Adding Listeners

#### Global Listeners
```go
// Global listener receives events from all nodes
workflow.AddGlobalListener(listener)
```

#### Node-Specific Listeners
```go
// Node listener receives events from a specific node only
node := workflow.AddNode("nodeName", "Description", handler)
node.AddListener(specificListener)
```

#### Listener Interface
```go
type NodeListenerTyped[T any] interface {
    OnNodeEvent(ctx context.Context, event NodeEvent, nodeName string, state T, err error)
}

// For quick listeners, use the function adapter
workflow.AddGlobalListener(
    graph.NodeListenerTypedFunc[YourState](
        func(ctx context.Context, event NodeEvent, nodeName string, state YourState, err error) {
            // Handle event
        },
    ),
)
```

### Compilation Methods

```go
// Compile for normal execution
runnable, err := workflow.CompileListenable()

// Compile with configuration
config := &graph.Config{ThreadID: "thread-123"}
runnable, err := workflow.CompileListenableWithConfig(config)
```

### Execution Methods

```go
// Execute once
finalState, err := runnable.Invoke(ctx, initialState)

// Stream events
eventChan := runnable.Stream(ctx, initialState)
for event := range eventChan {
    // Process events
}

// Execute with config
finalState, err := runnable.InvokeWithConfig(ctx, initialState, config)
```

## Use Cases

### 1. **Real-time Monitoring**
```go
type Monitor struct {
    metrics prometheus.Counter
}

func (m *Monitor) OnNodeEvent(ctx context.Context, event NodeEvent,
                             nodeName string, state MyState, err error) {
    switch event {
    case NodeEventComplete:
        m.metrics.Inc()
    case NodeEventError:
        log.Printf("Node %s failed: %v", nodeName, err)
    }
}
```

### 2. **Debugging and Logging**
```go
type Debugger struct {
    logger *log.Logger
}

func (d *Debugger) OnNodeEvent(ctx context.Context, event NodeEvent,
                             nodeName string, state MyState, err error) {
    d.logger.Printf("[%s] %s: %+v", time.Now().Format(time.RFC3339), event, state)
}
```

### 3. **Progress Reporting**
```go
type ProgressReporter struct {
    websocket *websocket.Conn
}

func (p *ProgressReporter) OnNodeEvent(ctx context.Context, event NodeEvent,
                                      nodeName string, state MyState, err error) {
    progress := map[string]any{
        "node": nodeName,
        "event": event,
        "timestamp": time.Now(),
    }
    p.websocket.WriteJSON(progress)
}
```

### 4. **Conditional Logic**
```go
type ConditionalStopper struct {
    stopThreshold int
    stopChan      chan struct{}
}

func (c *ConditionalStopper) OnNodeEvent(ctx context.Context, event NodeEvent,
                                        nodeName string, state MyState, err error) {
    if state.Count >= c.stopThreshold {
        close(c.stopChan)  // Signal to stop execution
    }
}
```

## Best Practices

1. **Keep Listeners Lightweight** - Avoid blocking operations in listeners
2. **Use Buffered Channels** - For streaming, use appropriately sized buffers
3. **Handle Errors Gracefully** - Don't let listener errors crash the graph
4. **Use Context** - Respect cancellation signals in listeners
5. **Avoid Infinite Loops** - Be careful with listeners that modify state

## Performance Considerations

- Listeners add minimal overhead to graph execution
- Each listener runs in the same goroutine as the node execution
- For high-performance scenarios, consider:
  - Using channels for async processing
  - Implementing listener pools
  - Sampling events instead of processing all

## Related Examples

- [Generic StateGraph](../generic_state_graph/) - Basic type-safe graphs without listeners
- [Streaming Pipeline](../streaming_pipeline/) - Advanced streaming patterns
- [Checkpointing](../checkpointing/) - Persisting graph state

## Learn More

- [LangGraphGo Documentation](../../README.md)
- [Generics in Go](https://go.dev/blog/generics)
- [Observer Pattern](https://en.wikipedia.org/wiki/Observer_pattern)