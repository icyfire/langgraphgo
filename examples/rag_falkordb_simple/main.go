package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/store"
)

func main() {
	ctx := context.Background()

	// Create FalkorDB knowledge graph
	fmt.Println("Initializing FalkorDB knowledge graph...")
	falkorDBConnStr := "falkordb://localhost:6379/simple_rag_graph"
	kg, err := store.NewFalkorDBGraph(falkorDBConnStr)
	if err != nil {
		log.Fatalf("Failed to create FalkorDB knowledge graph: %v", err)
	}
	// Close the connection when done (type assert to access Close method)
	defer func() {
		if falkorDB, ok := kg.(*store.FalkorDBGraph); ok {
			falkorDB.Close()
		}
	}()

	// Add entities manually to the knowledge graph (simplified approach)
	fmt.Println("Adding entities and relationships to knowledge graph...")

	// Add entities manually to the knowledge graph
	entities := []*rag.Entity{
		{
			ID:   "john_smith",
			Name: "John Smith",
			Type: "PERSON",
			Properties: map[string]any{
				"role":      "senior software engineer",
				"specialty": "machine learning and artificial intelligence",
				"company":   "Google",
			},
		},
		{
			ID:   "sarah_johnson",
			Name: "Sarah Johnson",
			Type: "PERSON",
			Properties: map[string]any{
				"role":    "CEO",
				"company": "TechStart Inc.",
			},
		},
		{
			ID:   "google",
			Name: "Google",
			Type: "ORGANIZATION",
			Properties: map[string]any{
				"industry": "technology",
				"location": "Mountain View, California",
			},
		},
		{
			ID:   "techstart",
			Name: "TechStart Inc.",
			Type: "ORGANIZATION",
			Properties: map[string]any{
				"specialty": "blockchain technology",
				"location":  "San Francisco",
			},
		},
		{
			ID:   "python",
			Name: "Python",
			Type: "TECHNOLOGY",
			Properties: map[string]any{
				"type":      "programming language",
				"uses":      "machine learning, web development, data science",
				"libraries": "TensorFlow, PyTorch",
			},
		},
		{
			ID:   "machine_learning",
			Name: "Machine Learning",
			Type: "CONCEPT",
			Properties: map[string]any{
				"category":    "subset of artificial intelligence",
				"description": "enables computers to learn from data",
				"algorithms":  "neural networks, decision trees",
			},
		},
	}

	// Add entities manually to the knowledge graph
	fmt.Println("Adding entities...")
	for _, entity := range entities {
		err := kg.AddEntity(ctx, entity)
		if err != nil {
			log.Printf("Failed to add entity %s: %v", entity.ID, err)
		} else {
			fmt.Printf("  ✓ Added entity: %s (%s)\n", entity.Name, entity.Type)
		}
	}

	// Add relationships manually to the knowledge graph
	fmt.Println("\nAdding relationships...")
	relationships := []*rag.Relationship{
		{
			ID:     "john_works_at_google",
			Source: "john_smith",
			Target: "google",
			Type:   "WORKS_AT",
		},
		{
			ID:     "sarah_ceo_of_techstart",
			Source: "sarah_johnson",
			Target: "techstart",
			Type:   "CEO_OF",
		},
		{
			ID:     "john_specializes_ml",
			Source: "john_smith",
			Target: "machine_learning",
			Type:   "SPECIALIZES_IN",
		},
		{
			ID:     "python_used_for_ml",
			Source: "python",
			Target: "machine_learning",
			Type:   "USED_FOR",
		},
	}

	for _, rel := range relationships {
		err := kg.AddRelationship(ctx, rel)
		if err != nil {
			log.Printf("Failed to add relationship %s: %v", rel.ID, err)
		} else {
			fmt.Printf("  ✓ Added relationship: %s -> %s (%s)\n", rel.Source, rel.Target, rel.Type)
		}
	}

	fmt.Println("\nKnowledge graph populated successfully!\n")

	// Use a custom implementation for better querying
	fmt.Println("=== Fixed Graph Query Examples ===\n")

	// Create a Redis client for direct queries
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	// Query examples using direct Redis commands
	queryExamples := []struct {
		description string
		cypherQuery string
	}{
		{
			description: "Find all PERSON entities",
			cypherQuery: "MATCH (n:PERSON) RETURN n.id, n.name, n.role, n.company",
		},
		{
			description: "Find all ORGANIZATION entities",
			cypherQuery: "MATCH (n:ORGANIZATION) RETURN n.id, n.name, n.industry",
		},
		{
			description: "Find all TECHNOLOGY entities",
			cypherQuery: "MATCH (n:TECHNOLOGY) RETURN n.id, n.name, n.type, n.uses",
		},
		{
			description: "Find all CONCEPT entities",
			cypherQuery: "MATCH (n:CONCEPT) RETURN n.id, n.name, n.description",
		},
	}

	for i, example := range queryExamples {
		fmt.Printf("=== Query %d ===\n", i+1)
		fmt.Printf("Description: %s\n\n", example.description)

		// Execute query
		result, err := client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", example.cypherQuery).Result()
		if err != nil {
			log.Printf("Failed to execute query: %v", err)
			continue
		}

		// Parse the result (simplified parsing for non-compact mode)
		if r, ok := result.([]interface{}); ok && len(r) > 1 {
			if rows, ok := r[1].([]interface{}); ok && len(rows) > 0 {
				fmt.Printf("Results:\n")
				for j, row := range rows {
					if rowArr, ok := row.([]interface{}); ok {
						fmt.Printf("  [%d] ", j+1)
						for k, item := range rowArr {
							if str, ok := item.(string); ok {
								fmt.Printf("%s", str)
								if k < len(rowArr)-1 {
									fmt.Printf(", ")
								}
							}
						}
						fmt.Println()
					}
				}
			} else {
				fmt.Println("  No results found")
			}
		}
		fmt.Println(strings.Repeat("-", 60))
	}

	// Demonstrate relationship queries
	fmt.Println("\n=== Relationship Queries ===\n")

	relationshipQueries := []struct {
		description string
		cypherQuery string
	}{
		{
			description: "Find all relationships",
			cypherQuery: "MATCH (a)-[r]->(b) RETURN a.name, type(r), b.name",
		},
		{
			description: "Find who works where",
			cypherQuery: "MATCH (p:PERSON)-[r:WORKS_AT]->(o:ORGANIZATION) RETURN p.name, r, o.name",
		},
		{
			description: "Find what John Smith specializes in",
			cypherQuery: "MATCH (p {name: 'John Smith'})-[r:SPECIALIZES_IN]->(c) RETURN p.name, type(r), c.name",
		},
	}

	for i, example := range relationshipQueries {
		fmt.Printf("=== Relationship Query %d ===\n", i+1)
		fmt.Printf("Description: %s\n\n", example.description)

		result, err := client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", example.cypherQuery).Result()
		if err != nil {
			log.Printf("Failed to execute relationship query: %v", err)
			continue
		}

		if r, ok := result.([]interface{}); ok && len(r) > 1 {
			if rows, ok := r[1].([]interface{}); ok && len(rows) > 0 {
				fmt.Printf("Results:\n")
				for j, row := range rows {
					if rowArr, ok := row.([]interface{}); ok {
						fmt.Printf("  [%d] ", j+1)
						for k, item := range rowArr {
							if str, ok := item.(string); ok {
								fmt.Printf("%s", str)
								if k < len(rowArr)-1 {
									fmt.Printf(", ")
								}
							}
						}
						fmt.Println()
					}
				}
			} else {
				fmt.Println("  No results found")
			}
		}
		fmt.Println(strings.Repeat("-", 60))
	}

	// Show statistics
	fmt.Println("\n=== Knowledge Graph Statistics ===\n")

	statsQueries := []struct {
		name  string
		query string
	}{
		{"Total nodes", "MATCH (n) RETURN count(n) as count"},
		{"Total relationships", "MATCH ()-[r]->() RETURN count(r) as count"},
		{"Person count", "MATCH (p:PERSON) RETURN count(p) as count"},
		{"Organization count", "MATCH (o:ORGANIZATION) RETURN count(o) as count"},
	}

	for _, stat := range statsQueries {
		result, err := client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", stat.query).Result()
		if err != nil {
			log.Printf("Failed to get %s: %v", stat.name, err)
			continue
		}

		if r, ok := result.([]interface{}); ok && len(r) > 1 {
			if rows, ok := r[1].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].([]interface{}); ok && len(row) > 0 {
					if count, ok := row[0].(int64); ok {
						fmt.Printf("%s: %d\n", stat.name, count)
					}
				}
			}
		}
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("This fixed example demonstrates:")
	fmt.Println("✅ Proper entity and relationship storage")
	fmt.Println("✅ Correct data retrieval and display")
	fmt.Println("✅ Working relationship queries")
	fmt.Println("✅ Knowledge graph statistics")
	fmt.Println("✅ Ready for integration with RAG applications")
}
