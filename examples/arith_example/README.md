# Arithmetic Example (arith_example)

This example demonstrates how to use LangGraphGo with an LLM to perform arithmetic calculations.

## Overview

A simple graph with one `arith` node that:
1. Receives an arithmetic expression in the state
2. Calls LLM to calculate the result
3. Updates the state with the result

## Code

```go
// Add arith node that calls LLM to calculate expression
g.AddNode("arith", "arith", func(ctx context.Context, state map[string]any) (map[string]any, error) {
    expression := state["expression"].(string)

    // Call LLM to calculate
    prompt := fmt.Sprintf("Calculate: %s. Only return the number.", expression)
    result, err := model.Call(ctx, prompt)
    if err != nil {
        return nil, err
    }

    // Update state with result
    state["result"] = result
    return state, nil
})
```

## Running the Example

```bash
cd examples/arith_example
export OPENAI_API_KEY=your_key
go run main.go
```

## Expected Output

```
123 + 456 = 579
```
