package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/rag"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== RAG Package Demo ===")

	// Create basic RAG components
	embedder := rag.NewMockEmbedder(1536)
	vectorStore := rag.NewInMemoryVectorStore(embedder)

	// Add sample documents
	docs := []rag.Document{
		{
			ID: "doc1",
			Content: "Artificial intelligence (AI) is intelligence demonstrated by machines.",
			Metadata: map[string]any{
				"source": "wikipedia",
				"topic":  "AI",
			},
		},
		{
			ID: "doc2",
			Content: "Machine learning automates analytical model building.",
			Metadata: map[string]any{
				"source": "wikipedia",
				"topic":  "ML",
			},
		},
	}

	// Add documents to vector store
	if err := vectorStore.Add(ctx, docs); err != nil {
		log.Printf("Failed to add documents: %v", err)
		return
	}

	fmt.Println("✓ Successfully added documents to vector store")

	// Create text splitter
	splitter := rag.NewRecursiveCharacterTextSplitter(
		rag.WithChunkSize(100),
		rag.WithChunkOverlap(20),
	)

	// Split documents into chunks
	chunks := splitter.SplitDocuments(docs)
	fmt.Printf("✓ Split documents into %d chunks\n", len(chunks))

	// Create retriever
	retriever := rag.NewVectorStoreRetriever(vectorStore, embedder, 2)

	// Perform similarity search
	results, err := retriever.RetrieveWithK(ctx, "What is AI?", 2)
	if err != nil {
		log.Printf("Failed to retrieve: %v", err)
		return
	}

	fmt.Printf("✓ Retrieved %d documents for query 'What is AI?'\n", len(results))

	// Show search results
	for i, result := range results {
		fmt.Printf("Result %d:\n", i+1)
		fmt.Printf("  ID: %s\n", result.ID)
		fmt.Printf("  Content: %s\n", result.Content)
		fmt.Printf("  Metadata: %v\n", result.Metadata)
	}

	// Create simple reranker
	_ = rag.NewSimpleReranker()
	fmt.Println("✓ Created simple reranker")

	// Create static document loader
	_ = rag.NewStaticDocumentLoader(docs)
	fmt.Println("✓ Created static document loader")

	// Test basic operations
	searchResults, err := retriever.RetrieveWithConfig(ctx, "artificial intelligence", &rag.RetrievalConfig{
		K:              1,
		ScoreThreshold: 0.5,
		SearchType:     "similarity",
	})
	if err != nil {
		log.Printf("Failed to retrieve with config: %v", err)
		return
	}

	fmt.Printf("✓ Retrieved with custom config: %d results\n", len(searchResults))

	// Get vector store stats
	stats, err := vectorStore.GetStats(ctx)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		return
	}

	fmt.Printf("✓ Vector store stats: %d documents, %d vectors, %d dimensions\n",
		stats.TotalDocuments, stats.TotalVectors, stats.Dimension)

	fmt.Println("\n=== RAG Package Migration Complete ===")
	fmt.Println("All RAG functionality is now available in the dedicated rag package!")
	fmt.Println("\nKey features:")
	fmt.Println("- Vector-based retrieval with multiple vector stores")
	fmt.Println("- GraphRAG with knowledge graph extraction")
	fmt.Println("- Hybrid retrieval combining multiple strategies")
	fmt.Println("- Flexible document processing and text splitting")
	fmt.Println("- Seamless integration with LangGraph workflows")
	fmt.Println("- Backward compatibility with prebuilt package")

	fmt.Println("\nFor more examples, see the rag package documentation and tests.")
}