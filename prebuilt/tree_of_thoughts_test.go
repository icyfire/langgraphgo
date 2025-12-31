package prebuilt

import (
	"testing"
)

func TestCreateTreeOfThoughtsAgentMap(t *testing.T) {
	config := TreeOfThoughtsConfig{
		Generator:    &SimpleThoughtGenerator{},
		Evaluator:    &SimpleThoughtEvaluator{},
		InitialState: &SimpleThoughtState{isGoal: true, isValid: true, desc: "Goal"},
	}
	agent, err := CreateTreeOfThoughtsAgentMap(config)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if agent == nil {
		t.Fatal("Agent is nil")
	}
}
