package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// PlannerNode generates a research plan based on the query.
func PlannerNode(ctx context.Context, state interface{}) (interface{}, error) {
	s := state.(*State)
	fmt.Printf("--- Planner Node: Planning for query '%s' ---\n", s.Request.Query)

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf("You are a research planner. Create a step-by-step research plan for the following query: %s. Return ONLY the plan as a numbered list.", s.Request.Query)
	completion, err := llm.Call(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Simple parsing of the plan (splitting by newlines)
	lines := strings.Split(completion, "\n")
	var plan []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			plan = append(plan, trimmed)
		}
	}
	s.Plan = plan
	fmt.Printf("Generated Plan: %v\n", s.Plan)

	return s, nil
}

// ResearcherNode executes the research plan.
func ResearcherNode(ctx context.Context, state interface{}) (interface{}, error) {
	s := state.(*State)
	fmt.Printf("--- Researcher Node: Executing plan ---\n")

	// In a real implementation, we would use a search tool here.
	// For this showcase, we will simulate research or use the LLM to "hallucinate"/generate info if no tool is available.
	// Or better, let's use the LLM to simulate finding information for each step.

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	var results []string
	for _, step := range s.Plan {
		fmt.Printf("Researching step: %s\n", step)
		prompt := fmt.Sprintf("You are a researcher. Find detailed information for this research step: %s. Provide a summary of findings.", step)
		completion, err := llm.Call(ctx, prompt)
		if err != nil {
			return nil, err
		}
		results = append(results, fmt.Sprintf("Step: %s\nFindings: %s", step, completion))
	}

	s.ResearchResults = results
	return s, nil
}

// ReporterNode compiles the final report.
func ReporterNode(ctx context.Context, state interface{}) (interface{}, error) {
	s := state.(*State)
	fmt.Printf("--- Reporter Node: Generating final report ---\n")

	llm, err := getLLM()
	if err != nil {
		return nil, err
	}

	researchData := strings.Join(s.ResearchResults, "\n\n")
	prompt := fmt.Sprintf("You are a reporter. Write a comprehensive final report based on the following research findings:\n\n%s\n\nThe original query was: %s", researchData, s.Request.Query)

	completion, err := llm.Call(ctx, prompt)
	if err != nil {
		return nil, err
	}

	s.FinalReport = completion
	fmt.Printf("Final Report Generated.\n")
	return s, nil
}

func getLLM() (llms.Model, error) {
	// Use DeepSeek as per user preference
	// Ensure OPENAI_API_KEY and OPENAI_API_BASE are set in the environment
	return openai.New(
		openai.WithModel("deepseek-chat"), // or deepseek-v3 if available/mapped
		// openai.WithBaseURL is picked up from OPENAI_API_BASE env var by langchaingo usually,
		// or we can explicitly set it if we knew it.
		// The user said "llm = ChatOpenAI(model="deepseek-v3", temperature=0)".
		// Let's try to respect that.
		openai.WithModel("deepseek-v3"),
	)
}
