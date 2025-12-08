package graph

import (
	g "github.com/smallnest/langgraphgo/graph"
)

func NewGraph() (*g.StateRunnable, error) {
	workflow := g.NewStateGraph()

	workflow.AddNode("account", "提取用户ID", AccountNode)
	workflow.AddNode("search", "搜索社交资料", SearchNode)
	workflow.AddNode("profile", "生成用户画像", ProfileNode)

	workflow.SetEntryPoint("account")
	workflow.AddEdge("account", "search")
	workflow.AddEdge("search", "profile")
	workflow.AddEdge("profile", g.END)

	return workflow.Compile()
}
