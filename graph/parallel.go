package graph

import (
	"context"
	"fmt"
	"sync"
)

// ParallelNode represents a set of nodes that can execute in parallel
type ParallelNode[S any] struct {
	nodes []TypedNode[S]
	name  string
}

// NewParallelNode creates a new parallel node
func NewParallelNode[S any](name string, nodes ...TypedNode[S]) *ParallelNode[S] {
	return &ParallelNode[S]{
		name:  name,
		nodes: nodes,
	}
}

// Execute runs all nodes in parallel and collects results
func (pn *ParallelNode[S]) Execute(ctx context.Context, state S) ([]S, error) {
	// Create channels for results and errors
	type result struct {
		index int
		value S
		err   error
	}

	results := make(chan result, len(pn.nodes))
	var wg sync.WaitGroup

	// Execute all nodes in parallel
	for i, node := range pn.nodes {
		wg.Add(1)
		go func(idx int, n TypedNode[S]) {
			defer wg.Done()

			// Execute with panic recovery
			defer func() {
				if r := recover(); r != nil {
					results <- result{
						index: idx,
						err:   fmt.Errorf("panic in parallel node %s[%d]: %v", pn.name, idx, r),
					}
				}
			}()

			value, err := n.Function(ctx, state)
			results <- result{
				index: idx,
				value: value,
				err:   err,
			}
		}(i, node)
	}

	// Wait for all nodes to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	outputs := make([]S, len(pn.nodes))
	var firstError error

	for res := range results {
		if res.err != nil && firstError == nil {
			firstError = res.err
		}
		outputs[res.index] = res.value
	}

	if firstError != nil {
		return nil, fmt.Errorf("parallel execution failed: %w", firstError)
	}

	// Return collected results
	return outputs, nil
}

// AddParallelNodes adds a set of nodes that execute in parallel.
// merger is used to combine the results from parallel execution into a single state S.
func (g *StateGraph[S]) AddParallelNodes(
	groupName string,
	nodes map[string]func(context.Context, S) (S, error),
	merger func([]S) S,
) {
	// Create parallel node group
	parallelNodes := make([]TypedNode[S], 0, len(nodes))
	for name, fn := range nodes {
		parallelNodes = append(parallelNodes, TypedNode[S]{
			Name:     name,
			Function: fn,
		})
	}

	// Add as a single parallel node
	parallelNode := NewParallelNode(groupName, parallelNodes...)

	// Wrap with merger
	g.AddNode(groupName, "Parallel execution group: "+groupName, func(ctx context.Context, state S) (S, error) {
		results, err := parallelNode.Execute(ctx, state)
		if err != nil {
			var zero S
			return zero, err
		}
		return merger(results), nil
	})
}

// MapReduceNode executes nodes in parallel and reduces results
type MapReduceNode[S any] struct {
	name     string
	mapNodes []TypedNode[S]
	reducer  func([]S) (S, error)
}

// NewMapReduceNode creates a new map-reduce node
func NewMapReduceNode[S any](name string, reducer func([]S) (S, error), mapNodes ...TypedNode[S]) *MapReduceNode[S] {
	return &MapReduceNode[S]{
		name:     name,
		mapNodes: mapNodes,
		reducer:  reducer,
	}
}

// Execute runs map nodes in parallel and reduces results
func (mr *MapReduceNode[S]) Execute(ctx context.Context, state S) (S, error) {
	// Execute map phase in parallel
	pn := NewParallelNode(mr.name+"_map", mr.mapNodes...)
	results, err := pn.Execute(ctx, state)
	if err != nil {
		var zero S
		return zero, fmt.Errorf("map phase failed: %w", err)
	}

	// Execute reduce phase
	if mr.reducer != nil {
		return mr.reducer(results)
	}

	// If no reducer, return zero state (or we should enforce reducer?)
	// In the generic version, we can't return []S as S unless S is []S.
	// So we assume reducer is provided or S can hold the results.
	// But without reducer, we don't know how to combine.
	// For now, return zero if no reducer.
	var zero S
	return zero, nil
}

// AddMapReduceNode adds a map-reduce pattern node
func (g *StateGraph[S]) AddMapReduceNode(
	name string,
	mapFunctions map[string]func(context.Context, S) (S, error),
	reducer func([]S) (S, error),
) {
	// Create map nodes
	mapNodes := make([]TypedNode[S], 0, len(mapFunctions))
	for nodeName, fn := range mapFunctions {
		mapNodes = append(mapNodes, TypedNode[S]{
			Name:     nodeName,
			Function: fn,
		})
	}

	// Create and add map-reduce node
	mrNode := NewMapReduceNode(name, reducer, mapNodes...)
	g.AddNode(name, "Map-reduce node: "+name, mrNode.Execute)
}

// FanOutFanIn creates a fan-out/fan-in pattern.
// aggregator merges worker results into a state S that is passed to the collector.
func (g *StateGraph[S]) FanOutFanIn(
	source string,
	_ []string, // workers parameter kept for API compatibility
	collector string,
	workerFuncs map[string]func(context.Context, S) (S, error),
	aggregator func([]S) S,
	collectFunc func(S) (S, error),
) {
	// Add parallel worker nodes
	g.AddParallelNodes(source+"_workers", workerFuncs, aggregator)

	// Add collector node
	g.AddNode(collector, "Collector node: "+collector, func(ctx context.Context, state S) (S, error) {
		return collectFunc(state)
	})

	// Connect source to workers and workers to collector
	g.AddEdge(source, source+"_workers")
	g.AddEdge(source+"_workers", collector)
}
