package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/store"
)

// --- Simple File-based Checkpoint Store for Demo ---

type DiskStore struct {
	FilePath string
}

func NewDiskStore(path string) *DiskStore {
	return &DiskStore{FilePath: path}
}

func (s *DiskStore) loadAll() map[string]*graph.Checkpoint {
	data, err := os.ReadFile(s.FilePath)
	if err != nil {
		return make(map[string]*graph.Checkpoint)
	}
	var checkpoints map[string]*graph.Checkpoint
	if err := json.Unmarshal(data, &checkpoints); err != nil {
		return make(map[string]*graph.Checkpoint)
	}
	return checkpoints
}

func (s *DiskStore) saveAll(cps map[string]*graph.Checkpoint) error {
	data, err := json.MarshalIndent(cps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.FilePath, data, 0600)
}

func (s *DiskStore) Save(ctx context.Context, cp *graph.Checkpoint) error {
	cps := s.loadAll()
	cps[cp.ID] = cp
	return s.saveAll(cps)
}

func (s *DiskStore) Load(ctx context.Context, id string) (*graph.Checkpoint, error) {
	cps := s.loadAll()
	if cp, ok := cps[id]; ok {
		return cp, nil
	}
	return nil, fmt.Errorf("checkpoint not found")
}

func (s *DiskStore) List(ctx context.Context, threadID string) ([]*graph.Checkpoint, error) {
	cps := s.loadAll()
	var result []*graph.Checkpoint
	for _, cp := range cps {
		// Check metadata for thread_id
		if tid, ok := cp.Metadata["thread_id"].(string); ok && tid == threadID {
			result = append(result, cp)
		}
	}
	// Sort by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})
	return result, nil
}

func (s *DiskStore) Delete(ctx context.Context, id string) error {
	cps := s.loadAll()
	delete(cps, id)
	return s.saveAll(cps)
}

func (s *DiskStore) Clear(ctx context.Context, threadID string) error {
	cps := s.loadAll()
	for id, cp := range cps {
		if tid, ok := cp.Metadata["thread_id"].(string); ok && tid == threadID {
			delete(cps, id)
		}
	}
	return s.saveAll(cps)
}

// ListByThread returns all checkpoints for a specific thread_id
func (s *DiskStore) ListByThread(ctx context.Context, threadID string) ([]*store.Checkpoint, error) {
	cps := s.loadAll()
	var result []*store.Checkpoint
	for _, cp := range cps {
		// Check metadata for thread_id
		if tid, ok := cp.Metadata["thread_id"].(string); ok && tid == threadID {
			// Convert graph.Checkpoint to store.Checkpoint
			result = append(result, &store.Checkpoint{
				ID:        cp.ID,
				NodeName:  cp.NodeName,
				State:     cp.State,
				Metadata:  cp.Metadata,
				Timestamp: cp.Timestamp,
				Version:   cp.Version,
			})
		}
	}
	// Sort by version ascending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})
	return result, nil
}

// GetLatestByThread returns the latest checkpoint for a thread_id
func (s *DiskStore) GetLatestByThread(ctx context.Context, threadID string) (*store.Checkpoint, error) {
	checkpoints, err := s.ListByThread(ctx, threadID)
	if err != nil {
		return nil, err
	}

	if len(checkpoints) == 0 {
		return nil, fmt.Errorf("no checkpoints found for thread: %s", threadID)
	}

	// Return the last one (highest version due to sorting)
	return checkpoints[len(checkpoints)-1], nil
}

// --- Main Logic ---

func main() {
	storeFile := "checkpoints.json"
	store := NewDiskStore(storeFile)
	threadID := "durable-job-1"

	// 1. Define Graph
	g := graph.NewCheckpointableStateGraph[map[string]any]()
	// Use MapSchema for state
	schema := graph.NewMapSchema()
	schema.RegisterReducer("steps", graph.AppendReducer)
	g.SetSchema(schema)

	// Configure Checkpointing
	g.SetCheckpointConfig(graph.CheckpointConfig{
		Store:    store,
		AutoSave: true,
	})

	// Step 1
	g.AddNode("step_1", "step_1", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		steps, ok := state["steps"].([]string)
		if !ok {
			return nil, fmt.Errorf("steps key not found")
		}
		// Append new step to existing steps
		steps = append(steps, "Step 1 Completed")
		state["steps"] = steps
		return state, nil
	})

	// Step 2 (Simulate Crash)
	g.AddNode("step_2", "step_2", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("Executing Step 2...")
		time.Sleep(500 * time.Millisecond)

		// Check if we should crash
		if os.Getenv("CRASH") == "true" {
			fmt.Println("!!! CRASHING AT STEP 2 !!!")
			fmt.Println("(Run again without CRASH=true to recover)")
			os.Exit(1)
		}

		steps, ok := state["steps"].([]string)
		if !ok {
			return nil, fmt.Errorf("steps key not found")
		}
		steps = append(steps, "Step 2 Completed")
		state["steps"] = steps
		return state, nil
	})

	// Step 3
	g.AddNode("step_3", "step_3", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		fmt.Println("Executing Step 3...")
		time.Sleep(500 * time.Millisecond)

		steps, ok := state["steps"].([]string)
		if !ok {
			return nil, fmt.Errorf("steps key not found")
		}
		steps = append(steps, "Step 3 Completed")
		state["steps"] = steps
		return state, nil
	})

	g.SetEntryPoint("step_1")
	g.AddEdge("step_1", "step_2")
	g.AddEdge("step_2", "step_3")
	g.AddEdge("step_3", graph.END)

	runnable, err := g.CompileCheckpointable()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Check for existing checkpoints to resume
	ctx := context.Background()
	checkpoints, _ := store.List(ctx, threadID)

	var config *graph.Config
	if len(checkpoints) > 0 {
		latest := checkpoints[len(checkpoints)-1]
		fmt.Printf("Found existing checkpoint: %s (Node: %s)\n", latest.ID, latest.NodeName)
		fmt.Println("Resuming execution...")

		var nextNode string
		if latest.NodeName == "step_1" {
			nextNode = "step_2"
		} else if latest.NodeName == "step_2" {
			nextNode = "step_3"
		} else {
			// Finished - show the final result from checkpoint
			stateMap, ok := latest.State.(map[string]any)
			if ok {
				// Normalize types after JSON unmarshaling
				if stepsAny, ok := stateMap["steps"].([]any); ok {
					steps := make([]string, len(stepsAny))
					for i, v := range stepsAny {
						s, _ := v.(string)
						steps[i] = s
					}
					stateMap["steps"] = steps
				}
				fmt.Printf("Job already finished.\nFinal Result: %v\n", stateMap)
			} else {
				fmt.Println("Job already finished or unknown state.")
			}
			return
		}

		config = &graph.Config{
			Configurable: map[string]any{
				"thread_id":     threadID,
				"checkpoint_id": latest.ID,
			},
			ResumeFrom: []string{nextNode},
		}

		fmt.Printf("Continuing from %s...\n", nextNode)
		// We need to cast state to map[string]any
		stateMap, ok := latest.State.(map[string]any)
		if !ok {
			// handle parsing if loaded as generic any (unmarshalled from JSON)
			// JSON unmarshal to any makes maps map[string]any
			// So simple cast might work, or we need more robust handling.
			// For now, let's assume it works or we re-marshal.
			// Actually, NewDiskStore uses json.Unmarshal into *graph.Checkpoint.
			// graph.Checkpoint.State is `any`.
			// If we save map[string]any, JSON marshals it.
			// Unmarshal will give map[string]any.
			// So the cast should be fine.
			stateMap = latest.State.(map[string]any)
		}

		// Normalize types after JSON unmarshaling
		// JSON unmarshal converts []string to []any
		if stepsAny, ok := stateMap["steps"].([]any); ok {
			steps := make([]string, len(stepsAny))
			for i, v := range stepsAny {
				s, ok := v.(string)
				if !ok {
					return
				}
				steps[i] = s
			}
			stateMap["steps"] = steps
		}

		res, err := runnable.InvokeWithConfig(ctx, stateMap, config)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Final Result: %v\n", res)

	} else {
		fmt.Println("Starting new execution...")
		config = &graph.Config{
			Configurable: map[string]any{
				"thread_id": threadID,
			},
		}
		initialState := map[string]any{"steps": []string{"Start"}}
		res, err := runnable.InvokeWithConfig(ctx, initialState, config)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Final Result: %v\n", res)
	}
}
