# 算术计算示例 (arith_example)

本示例演示如何使用 LangGraphGo 配合 LLM 进行算术计算。

## 概述

一个简单的图，包含一个 `arith` 节点：
1. 从状态中接收算术表达式
2. 调用 LLM 计算结果
3. 将结果更新到状态中

## 代码

```go
// 添加 arith 节点调用 LLM 计算表达式
g.AddNode("arith", "arith", func(ctx context.Context, state map[string]any) (map[string]any, error) {
    expression := state["expression"].(string)

    // 调用 LLM 计算
    prompt := fmt.Sprintf("Calculate: %s. Only return the number.", expression)
    result, err := model.Call(ctx, prompt)
    if err != nil {
        return nil, err
    }

    // 更新状态中的结果
    state["result"] = result
    return state, nil
})
```

## 运行示例

```bash
cd examples/arith_example
export OPENAI_API_KEY=your_key
go run main.go
```

## 预期输出

```
123 + 456 = 579
```
