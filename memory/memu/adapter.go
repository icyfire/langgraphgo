package memu

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/langgraphgo/memory"
)

// Ensure Client implements the memory.Memory interface
var _ memory.Memory = (*Client)(nil)

// AddMessage adds a new message to memory
// This implements the memory.Memory interface
func (c *Client) AddMessage(ctx context.Context, msg *memory.Message) error {
	queries := []Query{
		{
			Role:    msg.Role,
			Content: map[string]any{"text": msg.Content},
		},
	}

	return c.memorize(ctx, queries)
}

// GetContext retrieves relevant context for the current conversation
// This implements the memory.Memory interface
func (c *Client) GetContext(ctx context.Context, query string) ([]*memory.Message, error) {
	result, err := c.retrieve(ctx, query)
	if err != nil {
		return nil, err
	}

	// Convert memU results back to memory.Messages
	messages := make([]*memory.Message, 0)

	// Add category summaries as system messages
	for _, cat := range result.Categories {
		summary := cat.Summary
		if summary == "" {
			summary = cat.Description
		}
		if summary != "" {
			messages = append(messages, &memory.Message{
				ID:        cat.ID,
				Role:      "system",
				Content:   fmt.Sprintf("[%s] %s", cat.Name, summary),
				Timestamp: time.Now(),
				Metadata: map[string]any{
					"source":     "memu_category",
					"name":       cat.Name,
					"item_count": len(cat.ItemIDs),
				},
				TokenCount: estimateTokens(summary),
			})
		}
	}

	// Add item summaries as context
	for _, item := range result.Items {
		if item.Summary != "" {
			role := "system"
			if item.Type == "preference" || item.Type == "opinion" {
				role = "user"
			}

			messages = append(messages, &memory.Message{
				ID:        item.ID,
				Role:      role,
				Content:   item.Summary,
				Timestamp: time.Now(),
				Metadata: map[string]any{
					"source":     "memu_item",
					"type":       item.Type,
					"category":   item.Category,
					"importance": item.Importance,
				},
				TokenCount: estimateTokens(item.Summary),
			})
		}
	}

	return messages, nil
}

// Clear removes all messages from memory for the current user
// This implements the memory.Memory interface
// Note: memU doesn't have a direct "clear" API, so this returns an error
func (c *Client) Clear(ctx context.Context) error {
	return fmt.Errorf("clear operation not supported by memU API")
}

// GetStats returns statistics about the current memory state
// This implements the memory.Memory interface
func (c *Client) GetStats(ctx context.Context) (*memory.Stats, error) {
	result, err := c.getCategories(ctx)
	if err != nil {
		return nil, err
	}

	stats := &memory.Stats{
		TotalMessages:  0,
		TotalTokens:    0,
		ActiveMessages: len(result.Categories),
		ActiveTokens:   0,
	}

	// Count items across all categories
	for _, cat := range result.Categories {
		stats.TotalMessages += len(cat.ItemIDs)
		stats.TotalTokens += estimateTokens(cat.Summary)
	}

	return stats, nil
}
