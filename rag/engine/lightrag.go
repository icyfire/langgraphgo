package engine

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/smallnest/langgraphgo/rag"
)

// LightRAGEngine implements LightRAG functionality
// LightRAG combines low-level semantic chunks with high-level graph structures
// It supports four retrieval modes: naive, local, global, and hybrid
type LightRAGEngine struct {
	config         rag.LightRAGConfig
	knowledgeGraph rag.KnowledgeGraph
	embedder       rag.Embedder
	llm            rag.LLMInterface
	vectorStore    rag.VectorStore
	chunkCache     map[string][]rag.Document
	communityCache map[string]*rag.Community
	cacheMutex     sync.RWMutex
	metrics        *rag.Metrics
	baseEngine     *rag.BaseEngine
}

// NewLightRAGEngine creates a new LightRAG engine
func NewLightRAGEngine(
	config rag.LightRAGConfig,
	llm rag.LLMInterface,
	embedder rag.Embedder,
	kg rag.KnowledgeGraph,
	vectorStore rag.VectorStore,
) (*LightRAGEngine, error) {
	if kg == nil {
		return nil, fmt.Errorf("knowledge graph is required")
	}
	if embedder == nil {
		return nil, fmt.Errorf("embedder is required")
	}
	if llm == nil {
		return nil, fmt.Errorf("llm is required")
	}

	// Set default values
	if config.Mode == "" {
		config.Mode = "hybrid"
	}
	if config.ChunkSize == 0 {
		config.ChunkSize = 512
	}
	if config.ChunkOverlap == 0 {
		config.ChunkOverlap = 50
	}
	if config.LocalConfig.TopK == 0 {
		config.LocalConfig.TopK = 10
	}
	if config.LocalConfig.MaxHops == 0 {
		config.LocalConfig.MaxHops = 2
	}
	if config.GlobalConfig.MaxCommunities == 0 {
		config.GlobalConfig.MaxCommunities = 5
	}
	if config.MaxEntitiesPerChunk == 0 {
		config.MaxEntitiesPerChunk = 20
	}
	if config.HybridConfig.LocalWeight == 0 {
		config.HybridConfig.LocalWeight = 0.5
	}
	if config.HybridConfig.GlobalWeight == 0 {
		config.HybridConfig.GlobalWeight = 0.5
	}
	if config.HybridConfig.RFFK == 0 {
		config.HybridConfig.RFFK = 60
	}

	baseEngine := rag.NewBaseEngine(nil, embedder, &rag.Config{
		LightRAG: &config,
	})

	return &LightRAGEngine{
		config:         config,
		knowledgeGraph: kg,
		embedder:       embedder,
		llm:            llm,
		vectorStore:    vectorStore,
		chunkCache:     make(map[string][]rag.Document),
		communityCache: make(map[string]*rag.Community),
		metrics:        &rag.Metrics{},
		baseEngine:     baseEngine,
	}, nil
}

// Query performs a LightRAG query with the configured mode
func (l *LightRAGEngine) Query(ctx context.Context, query string) (*rag.QueryResult, error) {
	return l.QueryWithConfig(ctx, query, &rag.RetrievalConfig{
		K:              5,
		ScoreThreshold: 0.3,
		SearchType:     l.config.Mode,
		IncludeScores:  true,
	})
}

// QueryWithConfig performs a LightRAG query with custom configuration
func (l *LightRAGEngine) QueryWithConfig(ctx context.Context, query string, config *rag.RetrievalConfig) (*rag.QueryResult, error) {
	startTime := time.Now()
	mode := config.SearchType
	if mode == "" {
		mode = l.config.Mode
	}

	var result *rag.QueryResult
	var err error

	// Route to appropriate retrieval mode
	switch mode {
	case "naive":
		result, err = l.naiveRetrieval(ctx, query, config)
	case "local":
		result, err = l.localRetrieval(ctx, query, config)
	case "global":
		result, err = l.globalRetrieval(ctx, query, config)
	case "hybrid":
		result, err = l.hybridRetrieval(ctx, query, config)
	default:
		return nil, fmt.Errorf("unsupported retrieval mode: %s (supported: naive, local, global, hybrid)", mode)
	}

	if err != nil {
		return nil, err
	}

	// Update metrics
	l.metrics.TotalQueries++
	l.metrics.LastQueryTime = time.Now()
	latency := time.Since(startTime)
	l.metrics.AverageLatency = time.Duration((int64(l.metrics.AverageLatency)*l.metrics.TotalQueries + int64(latency)) / (l.metrics.TotalQueries + 1))

	result.ResponseTime = latency

	return result, nil
}

// naiveRetrieval performs simple retrieval without graph structure
func (l *LightRAGEngine) naiveRetrieval(ctx context.Context, query string, config *rag.RetrievalConfig) (*rag.QueryResult, error) {
	if l.vectorStore == nil {
		return nil, fmt.Errorf("vector store is required for naive retrieval")
	}

	// Generate query embedding
	queryEmbedding, err := l.embedder.EmbedDocument(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search vector store
	searchResults, err := l.vectorStore.Search(ctx, queryEmbedding, config.K)
	if err != nil {
		return nil, fmt.Errorf("failed to search vector store: %w", err)
	}

	// Convert to documents
	docs := make([]rag.Document, len(searchResults))
	for i, result := range searchResults {
		docs[i] = result.Document
	}

	// Build context
	contextStr := l.buildNaiveContext(searchResults)

	return &rag.QueryResult{
		Query:      query,
		Sources:    docs,
		Context:    contextStr,
		Confidence: l.calculateNaiveConfidence(searchResults),
		Metadata: map[string]any{
			"mode":        "naive",
			"num_results": len(searchResults),
			"avg_score":   l.avgScore(searchResults),
		},
	}, nil
}

// localRetrieval performs local mode retrieval
// Local mode retrieves relevant entities and their relationships within a localized context
func (l *LightRAGEngine) localRetrieval(ctx context.Context, query string, config *rag.RetrievalConfig) (*rag.QueryResult, error) {
	// Extract entities from query
	queryEntities, err := l.extractEntities(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to extract entities: %w", err)
	}

	// Build entity context
	entityDocs := make([]rag.Document, 0)
	seenEntities := make(map[string]bool)

	for _, queryEntity := range queryEntities {
		// Traverse the knowledge graph to find related entities
		relatedDocs, err := l.traverseEntities(ctx, queryEntity.ID, l.config.LocalConfig.MaxHops, seenEntities)
		if err != nil {
			continue
		}
		entityDocs = append(entityDocs, relatedDocs...)
	}

	// If we have a vector store, supplement with vector search
	if l.vectorStore != nil && len(entityDocs) < config.K {
		queryEmbedding, err := l.embedder.EmbedDocument(ctx, query)
		if err == nil {
			vectorResults, _ := l.vectorStore.Search(ctx, queryEmbedding, config.K-len(entityDocs))
			for _, result := range vectorResults {
				if len(entityDocs) >= config.K {
					break
				}
				entityDocs = append(entityDocs, result.Document)
			}
		}
	}

	// Limit results
	if len(entityDocs) > config.K {
		entityDocs = entityDocs[:config.K]
	}

	// Build context
	contextStr := l.buildLocalContext(queryEntities, entityDocs)

	return &rag.QueryResult{
		Query:      query,
		Sources:    entityDocs,
		Context:    contextStr,
		Confidence: l.calculateLocalConfidence(queryEntities, entityDocs),
		Metadata: map[string]any{
			"mode":               "local",
			"query_entities":     len(queryEntities),
			"retrieved_entities": len(entityDocs),
			"max_hops":           l.config.LocalConfig.MaxHops,
		},
	}, nil
}

// globalRetrieval performs global mode retrieval
// Global mode retrieves information from community-level summaries
func (l *LightRAGEngine) globalRetrieval(ctx context.Context, query string, config *rag.RetrievalConfig) (*rag.QueryResult, error) {
	// Extract entities from query
	queryEntities, err := l.extractEntities(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to extract entities: %w", err)
	}

	// Find relevant communities
	communities, err := l.findRelevantCommunities(ctx, query, queryEntities)
	if err != nil {
		return nil, fmt.Errorf("failed to find relevant communities: %w", err)
	}

	// Limit to max communities
	if len(communities) > l.config.GlobalConfig.MaxCommunities {
		communities = communities[:l.config.GlobalConfig.MaxCommunities]
	}

	// Build community documents
	communityDocs := make([]rag.Document, len(communities))
	for i, community := range communities {
		content := fmt.Sprintf("Community: %s\nSummary: %s\nEntities: %s",
			community.Title,
			community.Summary,
			strings.Join(community.Entities, ", "))
		communityDocs[i] = rag.Document{
			ID:      community.ID,
			Content: content,
			Metadata: map[string]any{
				"community_level": community.Level,
				"num_entities":    len(community.Entities),
				"score":           community.Score,
			},
		}
	}

	// Build context
	contextStr := l.buildGlobalContext(communities)

	return &rag.QueryResult{
		Query:      query,
		Sources:    communityDocs,
		Context:    contextStr,
		Confidence: l.calculateGlobalConfidence(communities),
		Metadata: map[string]any{
			"mode":              "global",
			"num_communities":   len(communities),
			"query_entities":    len(queryEntities),
			"include_hierarchy": l.config.GlobalConfig.IncludeHierarchy,
		},
	}, nil
}

// hybridRetrieval combines local and global retrieval results
func (l *LightRAGEngine) hybridRetrieval(ctx context.Context, query string, config *rag.RetrievalConfig) (*rag.QueryResult, error) {
	// Perform local and global retrieval
	localResult, err := l.localRetrieval(ctx, query, config)
	if err != nil {
		return nil, fmt.Errorf("local retrieval failed: %w", err)
	}

	globalResult, err := l.globalRetrieval(ctx, query, config)
	if err != nil {
		return nil, fmt.Errorf("global retrieval failed: %w", err)
	}

	// Fuse results based on fusion method
	var fusedDocs []rag.Document
	var fusedScores []float64

	switch l.config.HybridConfig.FusionMethod {
	case "rrf":
		fusedDocs, fusedScores = l.reciprocalRankFusion(localResult, globalResult)
	case "weighted":
		fusedDocs, fusedScores = l.weightedFusion(localResult, globalResult)
	default:
		fusedDocs, fusedScores = l.reciprocalRankFusion(localResult, globalResult)
	}

	// Limit results
	if len(fusedDocs) > config.K {
		fusedDocs = fusedDocs[:config.K]
		fusedScores = fusedScores[:config.K]
	}

	// Build context
	contextStr := l.buildHybridContext(localResult, globalResult, fusedDocs)

	// Calculate combined confidence
	combinedConfidence := (localResult.Confidence * l.config.HybridConfig.LocalWeight) +
		(globalResult.Confidence * l.config.HybridConfig.GlobalWeight)

	// Build metadata with scores
	metadata := map[string]any{
		"mode":              "hybrid",
		"fusion_method":     l.config.HybridConfig.FusionMethod,
		"local_weight":      l.config.HybridConfig.LocalWeight,
		"global_weight":     l.config.HybridConfig.GlobalWeight,
		"local_confidence":  localResult.Confidence,
		"global_confidence": globalResult.Confidence,
		"local_count":       len(localResult.Sources),
		"global_count":      len(globalResult.Sources),
	}
	// Add fused scores to metadata if available
	if fusedScores != nil {
		metadata["fused_scores"] = fusedScores
	}

	return &rag.QueryResult{
		Query:      query,
		Sources:    fusedDocs,
		Context:    contextStr,
		Confidence: combinedConfidence,
		Metadata:   metadata,
	}, nil
}

// AddDocuments adds documents to the LightRAG system
func (l *LightRAGEngine) AddDocuments(ctx context.Context, docs []rag.Document) error {
	startTime := time.Now()

	for _, doc := range docs {
		// Split document into chunks
		chunks := l.splitDocument(doc)

		// Cache chunks
		l.cacheMutex.Lock()
		l.chunkCache[doc.ID] = chunks
		l.cacheMutex.Unlock()

		// Add chunks to vector store if available
		if l.vectorStore != nil {
			if err := l.vectorStore.Add(ctx, chunks); err != nil {
				return fmt.Errorf("failed to add chunks to vector store: %w", err)
			}
		}

		// Extract entities and relationships from each chunk
		for _, chunk := range chunks {
			entities, err := l.extractEntities(ctx, chunk.Content)
			if err != nil {
				continue
			}

			// Add entities to knowledge graph
			for _, entity := range entities {
				if err := l.knowledgeGraph.AddEntity(ctx, entity); err != nil {
					return fmt.Errorf("failed to add entity: %w", err)
				}
			}

			// Extract and add relationships
			relationships, err := l.extractRelationships(ctx, chunk.Content, entities)
			if err != nil {
				continue
			}

			for _, rel := range relationships {
				if err := l.knowledgeGraph.AddRelationship(ctx, rel); err != nil {
					return fmt.Errorf("failed to add relationship: %w", err)
				}
			}
		}
	}

	// Build communities if enabled
	if l.config.EnableCommunityDetection {
		if err := l.buildCommunities(ctx); err != nil {
			return fmt.Errorf("failed to build communities: %w", err)
		}
	}

	l.metrics.TotalDocuments += int64(len(docs))
	l.metrics.IndexingLatency = time.Since(startTime)

	return nil
}

// DeleteDocument removes a document from the system
func (l *LightRAGEngine) DeleteDocument(ctx context.Context, docID string) error {
	l.cacheMutex.Lock()
	delete(l.chunkCache, docID)
	l.cacheMutex.Unlock()

	if l.vectorStore != nil {
		return l.vectorStore.Delete(ctx, []string{docID})
	}

	return fmt.Errorf("document deletion not fully implemented for LightRAG")
}

// UpdateDocument updates a document in the system
func (l *LightRAGEngine) UpdateDocument(ctx context.Context, doc rag.Document) error {
	if err := l.DeleteDocument(ctx, doc.ID); err != nil {
		return err
	}
	return l.AddDocuments(ctx, []rag.Document{doc})
}

// SimilaritySearch performs similarity search
func (l *LightRAGEngine) SimilaritySearch(ctx context.Context, query string, k int) ([]rag.Document, error) {
	result, err := l.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return result.Sources, nil
}

// SimilaritySearchWithScores performs similarity search with scores
func (l *LightRAGEngine) SimilaritySearchWithScores(ctx context.Context, query string, k int) ([]rag.DocumentSearchResult, error) {
	docs, err := l.SimilaritySearch(ctx, query, k)
	if err != nil {
		return nil, err
	}

	results := make([]rag.DocumentSearchResult, len(docs))
	for i, doc := range docs {
		results[i] = rag.DocumentSearchResult{
			Document: doc,
			Score:    1.0 - float64(i)/float64(len(docs)), // Simple ranking
		}
	}

	return results, nil
}

// splitDocument splits a document into chunks
func (l *LightRAGEngine) splitDocument(doc rag.Document) []rag.Document {
	chunks := make([]rag.Document, 0)
	content := doc.Content
	chunkSize := l.config.ChunkSize
	overlap := l.config.ChunkOverlap

	for i := 0; i < len(content); {
		end := min(i+chunkSize, len(content))

		chunk := rag.Document{
			ID:      fmt.Sprintf("%s_chunk_%d", doc.ID, len(chunks)),
			Content: content[i:end],
			Metadata: map[string]any{
				"source_doc":  doc.ID,
				"chunk_index": len(chunks),
				"metadata":    doc.Metadata,
			},
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		}

		chunks = append(chunks, chunk)

		i += (chunkSize - overlap)
	}

	return chunks
}

// extractEntities extracts entities from text
func (l *LightRAGEngine) extractEntities(ctx context.Context, text string) ([]*rag.Entity, error) {
	prompt := l.getEntityExtractionPrompt(text)

	response, err := l.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return l.parseEntityExtraction(response)
}

// extractRelationships extracts relationships between entities
func (l *LightRAGEngine) extractRelationships(ctx context.Context, text string, entities []*rag.Entity) ([]*rag.Relationship, error) {
	if len(entities) < 2 {
		return nil, nil
	}

	prompt := l.getRelationshipExtractionPrompt(text, entities)

	response, err := l.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return l.parseRelationshipExtraction(response)
}

// traverseEntities traverses the knowledge graph to find related entities
func (l *LightRAGEngine) traverseEntities(ctx context.Context, entityID string, maxHops int, seen map[string]bool) ([]rag.Document, error) {
	docs := make([]rag.Document, 0)

	if maxHops <= 0 || seen[entityID] {
		return docs, nil
	}

	seen[entityID] = true

	// Get related entities
	relatedEntities, err := l.knowledgeGraph.GetRelatedEntities(ctx, entityID, 1)
	if err != nil {
		return docs, err
	}

	for _, entity := range relatedEntities {
		if !seen[entity.ID] {
			// Create document from entity
			content := fmt.Sprintf("Entity: %s\nType: %s", entity.Name, entity.Type)
			if l.config.LocalConfig.IncludeDescriptions {
				if desc, ok := entity.Properties["description"]; ok {
					content += fmt.Sprintf("\nDescription: %v", desc)
				}
			}

			doc := rag.Document{
				ID:      entity.ID,
				Content: content,
				Metadata: map[string]any{
					"entity_type": entity.Type,
					"properties":  entity.Properties,
					"source":      "local_traversal",
				},
			}

			docs = append(docs, doc)

			// Recursively traverse
			moreDocs, _ := l.traverseEntities(ctx, entity.ID, maxHops-1, seen)
			docs = append(docs, moreDocs...)
		}
	}

	return docs, nil
}

// findRelevantCommunities finds communities relevant to the query
func (l *LightRAGEngine) findRelevantCommunities(ctx context.Context, query string, queryEntities []*rag.Entity) ([]*rag.Community, error) {
	communities := make([]*rag.Community, 0)

	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()

	// For each query entity, find its community
	for _, entity := range queryEntities {
		for _, community := range l.communityCache {
			// Check if entity is in this community
			if slices.Contains(community.Entities, entity.ID) {
				communities = append(communities, community)
			}
		}
	}

	return communities, nil
}

// buildCommunities builds communities using community detection
func (l *LightRAGEngine) buildCommunities(ctx context.Context) error {
	// This is a simplified implementation
	// In a production system, you would use proper community detection algorithms
	// like Louvain or Leiden

	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()

	// Create a default community
	community := &rag.Community{
		ID:         "community_0",
		Level:      0,
		Title:      "Default Community",
		Summary:    "All entities grouped together",
		Entities:   make([]string, 0),
		Properties: make(map[string]any),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Get all entities from the knowledge graph
	// This is a placeholder - in real implementation, you'd query the KG
	// to get all entities and group them properly

	l.communityCache["community_0"] = community

	return nil
}

// reciprocalRankFusion fuses results using RRF
func (l *LightRAGEngine) reciprocalRankFusion(localResult, globalResult *rag.QueryResult) ([]rag.Document, []float64) {
	k := float64(l.config.HybridConfig.RFFK)
	scores := make(map[string]float64)
	docs := make(map[string]rag.Document)

	// Process local results
	for i, doc := range localResult.Sources {
		score := 1.0 / (k + float64(i+1))
		if s, ok := scores[doc.ID]; ok {
			scores[doc.ID] = s + score
		} else {
			scores[doc.ID] = score
			docs[doc.ID] = doc
		}
	}

	// Process global results
	for i, doc := range globalResult.Sources {
		score := 1.0 / (k + float64(i+1))
		if s, ok := scores[doc.ID]; ok {
			scores[doc.ID] = s + score
		} else {
			scores[doc.ID] = score
			docs[doc.ID] = doc
		}
	}

	// Sort by score
	type docScore struct {
		doc   rag.Document
		score float64
	}
	sortedDocs := make([]docScore, 0, len(scores))
	for id, score := range scores {
		sortedDocs = append(sortedDocs, docScore{doc: docs[id], score: score})
	}

	// Simple sort
	for i := 0; i < len(sortedDocs); i++ {
		for j := i + 1; j < len(sortedDocs); j++ {
			if sortedDocs[j].score > sortedDocs[i].score {
				sortedDocs[i], sortedDocs[j] = sortedDocs[j], sortedDocs[i]
			}
		}
	}

	resultDocs := make([]rag.Document, len(sortedDocs))
	resultScores := make([]float64, len(sortedDocs))
	for i, ds := range sortedDocs {
		resultDocs[i] = ds.doc
		resultScores[i] = ds.score
	}

	return resultDocs, resultScores
}

// weightedFusion fuses results using weighted scores
func (l *LightRAGEngine) weightedFusion(localResult, globalResult *rag.QueryResult) ([]rag.Document, []float64) {
	scores := make(map[string]float64)
	docs := make(map[string]rag.Document)

	// Process local results
	for i, doc := range localResult.Sources {
		score := (1.0 - float64(i)/float64(len(localResult.Sources))) * l.config.HybridConfig.LocalWeight
		if s, ok := scores[doc.ID]; ok {
			scores[doc.ID] = s + score
		} else {
			scores[doc.ID] = score
			docs[doc.ID] = doc
		}
	}

	// Process global results
	for i, doc := range globalResult.Sources {
		score := (1.0 - float64(i)/float64(len(globalResult.Sources))) * l.config.HybridConfig.GlobalWeight
		if s, ok := scores[doc.ID]; ok {
			scores[doc.ID] = s + score
		} else {
			scores[doc.ID] = score
			docs[doc.ID] = doc
		}
	}

	// Sort by score
	type docScore struct {
		doc   rag.Document
		score float64
	}
	sortedDocs := make([]docScore, 0, len(scores))
	for id, score := range scores {
		sortedDocs = append(sortedDocs, docScore{doc: docs[id], score: score})
	}

	// Simple sort
	for i := 0; i < len(sortedDocs); i++ {
		for j := i + 1; j < len(sortedDocs); j++ {
			if sortedDocs[j].score > sortedDocs[i].score {
				sortedDocs[i], sortedDocs[j] = sortedDocs[j], sortedDocs[i]
			}
		}
	}

	resultDocs := make([]rag.Document, len(sortedDocs))
	resultScores := make([]float64, len(sortedDocs))
	for i, ds := range sortedDocs {
		resultDocs[i] = ds.doc
		resultScores[i] = ds.score
	}

	return resultDocs, resultScores
}

// Context building methods

func (l *LightRAGEngine) buildNaiveContext(results []rag.DocumentSearchResult) string {
	var sb strings.Builder
	sb.WriteString("Retrieved Context:\n\n")
	for i, result := range results {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, result.Document.Content))
		sb.WriteString(fmt.Sprintf("   Score: %.4f\n\n", result.Score))
	}
	return sb.String()
}

func (l *LightRAGEngine) buildLocalContext(entities []*rag.Entity, docs []rag.Document) string {
	var sb strings.Builder
	sb.WriteString("Local Retrieval Context:\n\n")

	sb.WriteString("Query Entities:\n")
	for _, entity := range entities {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", entity.Name, entity.Type))
	}
	sb.WriteString("\n")

	sb.WriteString("Related Information:\n")
	for i, doc := range docs {
		sb.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, doc.Content))
	}

	return sb.String()
}

func (l *LightRAGEngine) buildGlobalContext(communities []*rag.Community) string {
	var sb strings.Builder
	sb.WriteString("Global Retrieval Context:\n\n")

	for i, community := range communities {
		sb.WriteString(fmt.Sprintf("Community %d: %s\n", i+1, community.Title))
		sb.WriteString(fmt.Sprintf("Summary: %s\n", community.Summary))
		sb.WriteString(fmt.Sprintf("Entities: %s\n", strings.Join(community.Entities, ", ")))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (l *LightRAGEngine) buildHybridContext(localResult, globalResult *rag.QueryResult, fusedDocs []rag.Document) string {
	var sb strings.Builder
	sb.WriteString("Hybrid Retrieval Context:\n\n")

	sb.WriteString("=== Local Results ===\n")
	sb.WriteString(localResult.Context)
	sb.WriteString("\n")

	sb.WriteString("=== Global Results ===\n")
	sb.WriteString(globalResult.Context)
	sb.WriteString("\n")

	sb.WriteString("=== Fused Results ===\n")
	for i, doc := range fusedDocs {
		sb.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, doc.Content))
	}

	return sb.String()
}

// Confidence calculation methods

func (l *LightRAGEngine) calculateNaiveConfidence(results []rag.DocumentSearchResult) float64 {
	if len(results) == 0 {
		return 0.0
	}
	return l.avgScore(results)
}

func (l *LightRAGEngine) calculateLocalConfidence(entities []*rag.Entity, docs []rag.Document) float64 {
	if len(entities) == 0 {
		return 0.0
	}
	entityFactor := float64(len(entities)) / 10.0
	if entityFactor > 1.0 {
		entityFactor = 1.0
	}
	docFactor := float64(len(docs)) / 20.0
	if docFactor > 1.0 {
		docFactor = 1.0
	}
	return (entityFactor + docFactor) / 2.0
}

func (l *LightRAGEngine) calculateGlobalConfidence(communities []*rag.Community) float64 {
	if len(communities) == 0 {
		return 0.0
	}
	return float64(len(communities)) / float64(l.config.GlobalConfig.MaxCommunities)
}

// Helper methods

func (l *LightRAGEngine) avgScore(results []rag.DocumentSearchResult) float64 {
	if len(results) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, result := range results {
		sum += result.Score
	}
	return sum / float64(len(results))
}

// getEntityExtractionPrompt returns the prompt for entity extraction
func (l *LightRAGEngine) getEntityExtractionPrompt(text string) string {
	if customPrompt, ok := l.config.PromptTemplates["entity_extraction"]; ok {
		return fmt.Sprintf(customPrompt, text)
	}

	return fmt.Sprintf(`Extract entities from the following text. Focus on important entities like:
- People (PERSON)
- Organizations (ORGANIZATION)
- Locations (LOCATION)
- Products/Technologies (PRODUCT)
- Concepts (CONCEPT)

Return a JSON response with this structure:
{
  "entities": [
    {
      "id": "unique_id",
      "name": "entity_name",
      "type": "entity_type",
      "description": "brief_description",
      "properties": {}
    }
  ]
}

Limit to %d most important entities.

Text: %s`, l.config.MaxEntitiesPerChunk, text)
}

// getRelationshipExtractionPrompt returns the prompt for relationship extraction
func (l *LightRAGEngine) getRelationshipExtractionPrompt(text string, entities []*rag.Entity) string {
	if customPrompt, ok := l.config.PromptTemplates["relationship_extraction"]; ok {
		entityList := make([]string, len(entities))
		for i, e := range entities {
			entityList[i] = fmt.Sprintf("%s (%s)", e.Name, e.Type)
		}
		return fmt.Sprintf(customPrompt, text, strings.Join(entityList, ", "))
	}

	entityList := make([]string, len(entities))
	for i, e := range entities {
		entityList[i] = fmt.Sprintf("%s (%s)", e.Name, e.Type)
	}

	return fmt.Sprintf(`Extract relationships between the following entities from the text.
Consider relationship types like: RELATED_TO, PART_OF, WORKS_WITH, LOCATED_IN, CREATED_BY, etc.

Entities: %s

Return a JSON response with this structure:
{
  "relationships": [
    {
      "source": "entity1_name",
      "target": "entity2_name",
      "type": "relationship_type",
      "properties": {},
      "confidence": 0.9
    }
  ]
}

Text: %s`, strings.Join(entityList, ", "), text)
}

// parseEntityExtraction parses the entity extraction response
func (l *LightRAGEngine) parseEntityExtraction(response string) ([]*rag.Entity, error) {
	// Simplified parsing - in production, use proper JSON parsing
	entities := make([]*rag.Entity, 0)

	// For now, return a simple default entity
	// In production, parse the JSON response properly
	entity := &rag.Entity{
		ID:   "entity_1",
		Type: "UNKNOWN",
		Name: "Extracted Entity",
		Properties: map[string]any{
			"description": "Entity extracted from text",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	entities = append(entities, entity)

	return entities, nil
}

// parseRelationshipExtraction parses the relationship extraction response
func (l *LightRAGEngine) parseRelationshipExtraction(response string) ([]*rag.Relationship, error) {
	// Simplified parsing
	relationships := make([]*rag.Relationship, 0)

	rel := &rag.Relationship{
		ID:         "rel_1",
		Source:     "entity_1",
		Target:     "entity_2",
		Type:       "RELATED_TO",
		Properties: make(map[string]any),
		Confidence: 0.8,
		CreatedAt:  time.Now(),
	}
	relationships = append(relationships, rel)

	return relationships, nil
}

// GetMetrics returns the current metrics
func (l *LightRAGEngine) GetMetrics() *rag.Metrics {
	return l.metrics
}

// GetConfig returns the current configuration
func (l *LightRAGEngine) GetConfig() rag.LightRAGConfig {
	return l.config
}

// GetKnowledgeGraph returns the underlying knowledge graph
func (l *LightRAGEngine) GetKnowledgeGraph() rag.KnowledgeGraph {
	return l.knowledgeGraph
}
