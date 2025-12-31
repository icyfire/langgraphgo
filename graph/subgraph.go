package graph

import (
	"context"
	"fmt"
)

// Subgraph represents a nested graph that can be used as a node
type Subgraph[S any] struct {
	name     string
	graph    *StateGraph[S]
	runnable *StateRunnable[S]
}

// NewSubgraph creates a new generic subgraph
func NewSubgraph[S any](name string, graph *StateGraph[S]) (*Subgraph[S], error) {
	runnable, err := graph.Compile()
	if err != nil {
		return nil, fmt.Errorf("failed to compile subgraph %s: %w", name, err)
	}

	return &Subgraph[S]{
		name:     name,
		graph:    graph,
		runnable: runnable,
	}, nil
}

// Execute runs the subgraph as a node
func (s *Subgraph[S]) Execute(ctx context.Context, state S) (S, error) {
	result, err := s.runnable.Invoke(ctx, state)
	if err != nil {
		var zero S
		return zero, fmt.Errorf("subgraph %s execution failed: %w", s.name, err)
	}
	return result, nil
}

// AddSubgraph adds a subgraph as a node in the parent graph
func AddSubgraph[S, SubS any](g *StateGraph[S], name string, subgraph *StateGraph[SubS], converter func(S) SubS, resultConverter func(SubS) S) error {
	sg, err := NewSubgraph(name, subgraph)
	if err != nil {
		return err
	}

	// Wrap the execute function to match the state type
	wrappedFn := func(ctx context.Context, state S) (S, error) {
		// Convert S to SubS
		subState := converter(state)
		result, err := sg.Execute(ctx, subState)
		if err != nil {
			var zero S
			return zero, err
		}
		// Convert result back to S
		return resultConverter(result), nil
	}

	g.AddNode(name, "Subgraph: "+name, wrappedFn)
	return nil
}

// CreateSubgraph creates and adds a subgraph using a builder function
func CreateSubgraph[S, SubS any](g *StateGraph[S], name string, builder func(*StateGraph[SubS]) error, converter func(S) SubS, resultConverter func(SubS) S) error {
	subgraph := NewStateGraph[SubS]()
	if err := builder(subgraph); err != nil {
		return err
	}
	return AddSubgraph(g, name, subgraph, converter, resultConverter)
}

// CompositeGraph allows composing multiple graphs together
type CompositeGraph[S any] struct {
	graphs map[string]*StateGraph[S]
	main   *StateGraph[S]
}

// NewCompositeGraph creates a new composite graph
func NewCompositeGraph[S any]() *CompositeGraph[S] {
	return &CompositeGraph[S]{
		graphs: make(map[string]*StateGraph[S]),
		main:   NewStateGraph[S](),
	}
}

// AddGraph adds a named graph to the composite
func (cg *CompositeGraph[S]) AddGraph(name string, graph *StateGraph[S]) {
	cg.graphs[name] = graph
}

// Connect connects two graphs with a transformation function
func (cg *CompositeGraph[S]) Connect(
	fromGraph string,
	fromNode string,
	toGraph string,
	toNode string,
	transform func(S) S,
) error {
	// Create a bridge node that transforms state between graphs
	bridgeName := fmt.Sprintf("%s_%s_to_%s_%s", fromGraph, fromNode, toGraph, toNode)

	cg.main.AddNode(bridgeName, "Bridge: "+bridgeName, func(_ context.Context, state S) (S, error) {
		if transform != nil {
			return transform(state), nil
		}
		return state, nil
	})

	return nil
}

// Compile compiles the composite graph into a single runnable
func (cg *CompositeGraph[S]) Compile() (*StateRunnable[S], error) {
	// Add all subgraphs to the main graph
	for name, graph := range cg.graphs {
		if err := AddSubgraph(cg.main, name, graph,
			func(s S) S { return s },
			func(s S) S { return s }); err != nil {
			return nil, fmt.Errorf("failed to add subgraph %s: %w", name, err)
		}
	}

	return cg.main.Compile()
}

// RecursiveSubgraph allows a subgraph to call itself recursively
type RecursiveSubgraph[S any] struct {
	name      string
	graph     *StateGraph[S]
	maxDepth  int
	condition func(S, int) bool // Should continue recursion?
}

// NewRecursiveSubgraph creates a new recursive subgraph
func NewRecursiveSubgraph[S any](
	name string,
	maxDepth int,
	condition func(S, int) bool,
) *RecursiveSubgraph[S] {
	return &RecursiveSubgraph[S]{
		name:      name,
		graph:     NewStateGraph[S](),
		maxDepth:  maxDepth,
		condition: condition,
	}
}

// Execute runs the recursive subgraph
func (rs *RecursiveSubgraph[S]) Execute(ctx context.Context, state S) (S, error) {
	return rs.executeRecursive(ctx, state, 0)
}

func (rs *RecursiveSubgraph[S]) executeRecursive(ctx context.Context, state S, depth int) (S, error) {
	// Check max depth
	if depth >= rs.maxDepth {
		return state, nil
	}

	// Check condition
	if !rs.condition(state, depth) {
		return state, nil
	}

	// Compile and execute the graph
	runnable, err := rs.graph.Compile()
	if err != nil {
		var zero S
		return zero, fmt.Errorf("failed to compile recursive subgraph at depth %d: %w", depth, err)
	}

	result, err := runnable.Invoke(ctx, state)
	if err != nil {
		var zero S
		return zero, fmt.Errorf("recursive execution failed at depth %d: %w", depth, err)
	}

	// Recurse with the result
	return rs.executeRecursive(ctx, result, depth+1)
}

// AddRecursiveSubgraph adds a recursive subgraph to the parent graph
func AddRecursiveSubgraph[S, SubS any](
	g *StateGraph[S],
	name string,
	maxDepth int,
	condition func(SubS, int) bool,
	builder func(*StateGraph[SubS]) error,
	converter func(S) SubS,
	resultConverter func(SubS) S,
) error {
	rs := NewRecursiveSubgraph(name, maxDepth, condition)
	if err := builder(rs.graph); err != nil {
		return err
	}

	wrappedFn := func(ctx context.Context, state S) (S, error) {
		subState := converter(state)
		result, err := rs.Execute(ctx, subState)
		if err != nil {
			var zero S
			return zero, err
		}
		return resultConverter(result), nil
	}

	g.AddNode(name, "Recursive subgraph: "+name, wrappedFn)
	return nil
}

// AddNestedConditionalSubgraph creates a subgraph with its own conditional routing
func AddNestedConditionalSubgraph[S, SubS any](
	g *StateGraph[S],
	name string,
	router func(S) string,
	subgraphs map[string]*StateGraph[SubS],
	converter func(S) SubS,
	resultConverter func(SubS) S,
) error {
	// Create a wrapper node that routes to different subgraphs
	wrappedFn := func(ctx context.Context, state S) (S, error) {
		// Determine which subgraph to use
		subgraphName := router(state)

		subgraph, exists := subgraphs[subgraphName]
		if !exists {
			var zero S
			return zero, fmt.Errorf("subgraph %s not found", subgraphName)
		}

		// Convert state to SubS
		subState := converter(state)

		// Compile and execute the selected subgraph
		runnable, err := subgraph.Compile()
		if err != nil {
			var zero S
			return zero, fmt.Errorf("failed to compile subgraph %s: %w", subgraphName, err)
		}

		result, err := runnable.Invoke(ctx, subState)
		if err != nil {
			var zero S
			return zero, err
		}

		// Convert result back to S
		return resultConverter(result), nil
	}

	g.AddNode(name, "Nested conditional subgraph: "+name, wrappedFn)
	return nil
}
