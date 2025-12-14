// LangGraphGo 内存策略演示程序
//
// 这个程序演示了 LangGraphGo 中所有可用的内存管理策略。
// 运行方式: go run memory_examples.go
//
// 演示的内存策略:
// 1. Sequential Memory - 存储所有消息
// 2. Sliding Window Memory - 只保留最近的消息
// 3. Buffer Memory - 灵活的缓冲限制
// 4. Summarization Memory - 压缩旧消息
// 5. Retrieval Memory - 基于相似度检索
// 6. Hierarchical Memory - 分层存储
// 7. Graph-Based Memory - 关系图谱
// 8. Compression Memory - 智能压缩
// 9. OS-Like Memory - 操作系统式管理
package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/smallnest/langgraphgo/memory"
)

// 内存策略演示程序
func main() {
	ctx := context.Background()

	fmt.Println("=== LangGraphGo 内存策略演示 ===\n")

	// 1. Sequential Memory 示例
	fmt.Println("1. Sequential Memory (顺序内存)")
	demonstrateSequentialMemory(ctx)

	// 2. Sliding Window Memory 示例
	fmt.Println("\n2. Sliding Window Memory (滑动窗口)")
	demonstrateSlidingWindowMemory(ctx)

	// 3. Buffer Memory 示例
	fmt.Println("\n3. Buffer Memory (缓冲内存)")
	demonstrateBufferMemory(ctx)

	// 4. Summarization Memory 示例
	fmt.Println("\n4. Summarization Memory (摘要内存)")
	demonstrateSummarizationMemory(ctx)

	// 5. Retrieval Memory 示例
	fmt.Println("\n5. Retrieval Memory (检索内存)")
	demonstrateRetrievalMemory(ctx)

	// 6. Hierarchical Memory 示例
	fmt.Println("\n6. Hierarchical Memory (分层内存)")
	demonstrateHierarchicalMemory(ctx)

	// 7. Graph-Based Memory 示例
	fmt.Println("\n7. Graph-Based Memory (图内存)")
	demonstrateGraphBasedMemory(ctx)

	// 8. Compression Memory 示例
	fmt.Println("\n8. Compression Memory (压缩内存)")
	demonstrateCompressionMemory(ctx)

	// 9. OS-Like Memory 示例
	fmt.Println("\n9. OS-Like Memory (操作系统式内存)")
	demonstrateOSLikeMemory(ctx)

	// 性能对比
	fmt.Println("\n=== 性能对比 ===")
	comparePerformance(ctx)
}

// 1. Sequential Memory 演示
func demonstrateSequentialMemory(ctx context.Context) {
	mem := memory.NewSequentialMemory()

	// 添加消息
	messages := []string{
		"你好，我想了解 LangGraphGo",
		"LangGraphGo 是一个强大的 AI 框架",
		"它有什么特性？",
		"支持并行执行、状态管理、持久化等",
	}

	for _, content := range messages {
		msg := memory.NewMessage("user", content)
		if strings.Contains(content, "框架") || strings.Contains(content, "支持") {
			msg.Role = "assistant"
		}
		mem.AddMessage(ctx, msg)
	}

	// 获取所有消息
	context, _ := mem.GetContext(ctx, "")
	stats, _ := mem.GetStats(ctx)

	fmt.Printf("  存储消息数: %d\n", len(context))
	fmt.Printf("  总 Token 数: %d\n", stats.TotalTokens)
	fmt.Printf("  最新消息: %s\n", context[len(context)-1].Content)
}

// 2. Sliding Window Memory 演示
func demonstrateSlidingWindowMemory(ctx context.Context) {
	mem := memory.NewSlidingWindowMemory(3) // 只保留最近 3 条

	// 添加 5 条消息
	for i := 1; i <= 5; i++ {
		msg := memory.NewMessage("user", fmt.Sprintf("消息 %d", i))
		mem.AddMessage(ctx, msg)
		fmt.Printf("  添加消息 %d\n", i)
	}

	// 只会保留最近 3 条
	context, _ := mem.GetContext(ctx, "")
	fmt.Printf("  实际保留: %d 条消息\n", len(context))
	fmt.Printf("  消息列表: %s\n", formatMessageList(context))
}

// 3. Buffer Memory 演示
func demonstrateBufferMemory(ctx context.Context) {
	mem := memory.NewBufferMemory(&memory.BufferConfig{
		MaxMessages:   5,
		MaxTokens:     200,
		AutoSummarize: true,
		Summarizer:    simpleSummarizer,
	})

	// 添加消息
	longMessage := strings.Repeat("这是一个很长的消息。", 20)
	for i := 1; i <= 6; i++ {
		msg := memory.NewMessage("user", fmt.Sprintf("消息 %d: %s", i, longMessage[:50]))
		mem.AddMessage(ctx, msg)
	}

	stats, _ := mem.GetStats(ctx)
	fmt.Printf("  活跃消息: %d\n", stats.ActiveMessages)
	fmt.Printf("  活跃 Tokens: %d\n", stats.ActiveTokens)
}

// 4. Summarization Memory 演示
func demonstrateSummarizationMemory(ctx context.Context) {
	mem := memory.NewSummarizationMemory(&memory.SummarizationConfig{
		RecentWindowSize: 3,
		SummarizeAfter:   5,
		Summarizer:       simpleSummarizer,
	})

	// 添加消息
	topics := []string{"产品介绍", "价格讨论", "技术细节", "交付时间", "售后服务", "合同条款"}
	for i, topic := range topics {
		msg := memory.NewMessage("user", fmt.Sprintf("讨论 %s: 详细信息...", topic))
		mem.AddMessage(ctx, msg)
		if i == 2 { // 添加一个系统消息
			sysMsg := memory.NewMessage("system", "重要：记录所有决策")
			mem.AddMessage(ctx, sysMsg)
		}
	}

	context, _ := mem.GetContext(ctx, "")
	stats, _ := mem.GetStats(ctx)

	fmt.Printf("  总消息数: %d\n", stats.TotalMessages)
	fmt.Printf("  压缩率: %.2f%%\n", stats.CompressionRate*100)
	fmt.Printf("  上下文构成: %d 条摘要 + %d 条最近消息\n",
		countSummaries(context), countRecent(context))
}

// 5. Retrieval Memory 演示
func demonstrateRetrievalMemory(ctx context.Context) {
	mem := memory.NewRetrievalMemory(&memory.RetrievalConfig{
		TopK: 3,
		EmbeddingFunc: func(ctx context.Context, text string) ([]float64, error) {
			// 简化的嵌入函数（实际应使用真实的嵌入模型）
			return simpleEmbedding(text), nil
		},
	})

	// 添加不同主题的消息
	topics := []string{
		"产品价格是 1000 元",
		"技术栈使用 React 和 Go",
		"团队有 10 名开发者",
		"交付周期是 3 个月",
		"支持 24/7 客服",
	}

	for _, content := range topics {
		msg := memory.NewMessage("user", content)
		mem.AddMessage(ctx, msg)
	}

	// 查询相关问题
	queries := []string{"价格信息", "技术架构", "团队规模"}
	for _, query := range queries {
		context, _ := mem.GetContext(ctx, query)
		fmt.Printf("  查询 '%s': 找到 %d 条相关消息\n", query, len(context))
	}
}

// 6. Hierarchical Memory 演示
func demonstrateHierarchicalMemory(ctx context.Context) {
	mem := memory.NewHierarchicalMemory(&memory.HierarchicalConfig{
		RecentLimit:    3,
		ImportantLimit: 5,
		ImportanceScorer: func(msg *memory.Message) float64 {
			score := 0.5

			if msg.Role == "system" {
				score += 0.3
			}

			if strings.Contains(msg.Content, "重要") ||
				strings.Contains(msg.Content, "决策") {
				score += 0.3
			}

			return math.Min(score, 1.0)
		},
	})

	// 添加混合重要性的消息
	messages := []struct {
		role    string
		content string
	}{
		{"user", "普通消息 1"},
		{"system", "重要系统通知"},
		{"user", "普通消息 2"},
		{"user", "重要决策：选择方案 A"},
		{"assistant", "普通消息 3"},
		{"user", "普通消息 4"},
	}

	for _, m := range messages {
		msg := memory.NewMessage(m.role, m.content)
		mem.AddMessage(ctx, msg)
	}

	stats, _ := mem.GetStats(ctx)
	fmt.Printf("  活跃消息: %d 条\n", stats.ActiveMessages)
	fmt.Printf("  总消息: %d 条\n", stats.TotalMessages)
	fmt.Printf("  总 Tokens: %d\n", stats.TotalTokens)
}

// 7. Graph-Based Memory 演示
func demonstrateGraphBasedMemory(ctx context.Context) {
	mem := memory.NewGraphBasedMemory(&memory.GraphConfig{
		TopK: 3,
		RelationExtractor: func(msg *memory.Message) []string {
			// 简单的主题提取
			topics := []string{}
			if strings.Contains(msg.Content, "产品") {
				topics = append(topics, "产品")
			}
			if strings.Contains(msg.Content, "价格") {
				topics = append(topics, "价格")
			}
			if strings.Contains(msg.Content, "技术") {
				topics = append(topics, "技术")
			}
			return topics
		},
	})

	// 添加相关消息
	messages := []string{
		"产品功能介绍",
		"价格策略制定",
		"技术架构设计",
		"产品路线图",
		"技术选型讨论",
	}

	for _, content := range messages {
		msg := memory.NewMessage("user", content)
		mem.AddMessage(ctx, msg)
	}

	// 查询相关主题
	context, _ := mem.GetContext(ctx, "产品")
	fmt.Printf("  查询 '产品': 找到 %d 条相关消息\n", len(context))
}

// 8. Compression Memory 演示
func demonstrateCompressionMemory(ctx context.Context) {
	mem := memory.NewCompressionMemory(&memory.CompressionConfig{
		CompressionTrigger: 5,
		Compressor: func(ctx context.Context, msgs []*memory.Message) (*memory.CompressedBlock, error) {
			content := fmt.Sprintf("压缩块 (%d-%d): %d 条消息",
				msgs[0].Timestamp.Format("01-02"),
				msgs[len(msgs)-1].Timestamp.Format("01-02"),
				len(msgs))

			return &memory.CompressedBlock{
				ID:               fmt.Sprintf("block-%d", time.Now().Unix()),
				Summary:          content,
				OriginalCount:    len(msgs),
				OriginalTokens:   calculateTokens(msgs),
				CompressedTokens: len(content) / 4, // 简单估算
				TimeRange: memory.TimeRange{
					Start: msgs[0].Timestamp,
					End:   msgs[len(msgs)-1].Timestamp,
				},
			}, nil
		},
	})

	// 添加消息触发压缩
	for i := 1; i <= 10; i++ {
		msg := memory.NewMessage("user", fmt.Sprintf("消息 %d: 详细内容...", i))
		mem.AddMessage(ctx, msg)
	}

	stats, _ := mem.GetStats(ctx)
	fmt.Printf("  活跃消息: %d\n", stats.ActiveMessages)
	fmt.Printf("  压缩率: %.2f%%\n", stats.CompressionRate*100)
}

// 9. OS-Like Memory 演示
func demonstrateOSLikeMemory(ctx context.Context) {
	mem := memory.NewOSLikeMemory(&memory.OSLikeConfig{
		ActiveLimit:  3,
		CacheLimit:   5,
		AccessWindow: time.Minute * 5,
	})

	// 添加消息并模拟访问
	for i := 1; i <= 10; i++ {
		msg := memory.NewMessage("user", fmt.Sprintf("消息 %d", i))
		mem.AddMessage(ctx, msg)

		// 模拟随机访问
		if i%3 == 0 {
			mem.GetContext(ctx, fmt.Sprintf("消息 %d", i-1))
		}
	}

	stats, _ := mem.GetStats(ctx)
	fmt.Printf("  活跃消息: %d\n", stats.ActiveMessages)
	fmt.Printf("  总消息: %d\n", stats.TotalMessages)
	fmt.Printf("  压缩率: %.2f%%\n", stats.CompressionRate*100)
}

// 性能对比
func comparePerformance(ctx context.Context) {
	strategies := map[string]memory.Memory{
		"Sequential":    memory.NewSequentialMemory(),
		"SlidingWindow": memory.NewSlidingWindowMemory(10),
		"Buffer":        memory.NewBufferMemory(&memory.BufferConfig{MaxMessages: 20}),
		"Hierarchical": memory.NewHierarchicalMemory(&memory.HierarchicalConfig{
			RecentLimit: 5, ImportantLimit: 10,
			ImportanceScorer: func(msg *memory.Message) float64 { return 0.5 },
		}),
	}

	messageCount := 50

	fmt.Printf("测试场景: %d 条消息的性能对比\n\n", messageCount)
	fmt.Printf("%-15s %-12s %-12s %-15s %-12s\n", "策略", "存储消息", "活跃消息", "Token 效率", "响应时间")
	fmt.Println(strings.Repeat("-", 70))

	for name, mem := range strategies {
		// 清空并添加测试消息
		mem.Clear(ctx)

		start := time.Now()
		for i := 0; i < messageCount; i++ {
			msg := memory.NewMessage("user", fmt.Sprintf("测试消息 %d: %s", i, strings.Repeat("内容", 10)))
			mem.AddMessage(ctx, msg)
		}

		// 获取上下文
		mem.GetContext(ctx, "测试查询")
		duration := time.Since(start)

		stats, _ := mem.GetStats(ctx)
		efficiency := float64(stats.ActiveTokens) / float64(stats.TotalTokens)

		fmt.Printf("%-15s %-12d %-12d %-15.2f %-12s\n",
			name,
			stats.TotalMessages,
			stats.ActiveMessages,
			efficiency,
			duration.String())
	}
}

// 辅助函数

func formatMessageList(messages []*memory.Message) string {
	contents := make([]string, len(messages))
	for i, msg := range messages {
		contents[i] = fmt.Sprintf("%s", msg.Content[:min(20, len(msg.Content))])
	}
	return fmt.Sprintf("[%s]", strings.Join(contents, ", "))
}

func simpleSummarizer(ctx context.Context, messages []*memory.Message) (string, error) {
	return fmt.Sprintf("摘要: %d 条消息已压缩", len(messages)), nil
}

func simpleEmbedding(text string) []float64 {
	// 简化的嵌入：基于字符串哈希
	hash := 0.0
	for i, c := range text {
		hash += float64(int(c) * (i + 1))
	}

	embedding := make([]float64, 10)
	for i := range embedding {
		embedding[i] = math.Sin(hash + float64(i))
	}
	return embedding
}

func countSummaries(messages []*memory.Message) int {
	count := 0
	for _, msg := range messages {
		if strings.Contains(msg.Content, "摘要") {
			count++
		}
	}
	return count
}

func countRecent(messages []*memory.Message) int {
	return len(messages) - countSummaries(messages)
}

func calculateTokens(messages []*memory.Message) int {
	total := 0
	for _, msg := range messages {
		total += msg.TokenCount
	}
	return total
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
