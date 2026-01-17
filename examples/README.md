# LangGraphGo Examples

This directory contains **85+ examples** demonstrating the features of LangGraphGo.

## Table of Contents

- [Basic Concepts](#basic-concepts)
- [State Management](#state-management)
- [Graph Structure & Routing](#graph-structure--routing)
- [Parallel Execution](#parallel-execution)
- [Streaming & Events](#streaming--events)
- [Persistence (Checkpointing)](#persistence-checkpointing)
- [Human-in-the-Loop](#human-in-the-loop)
- [Pre-built Agents](#pre-built-agents)
- [Programmatic Tool Calling (PTC)](#programmatic-tool-calling-ptc)
- [Memory](#memory)
- [RAG (Retrieval Augmented Generation)](#rag-retrieval-augmented-generation)
- [Tools & Integrations](#tools--integrations)
- [Advanced Patterns](#advanced-patterns)

---

## Basic Concepts

- **[Basic Example](basic_example/)** - Simple graph with hardcoded steps
- **[Basic LLM](basic_llm/)** - Integration with LLMs
- **[Arithmetic Example](arith_example/)** - Using LLM to calculate arithmetic expressions
- **[LangChain Integration](langchain_example/)** - Using LangChain tools and models

## State Management

- **[State Schema](state_schema/)** - Managing complex state updates with Schema and Reducers
- **[Custom Reducer](custom_reducer/)** - Defining custom state reducers for complex merge logic
- **[Smart Messages](smart_messages/)** - Intelligent message merging with ID-based upserts
- **[Command API](command_api/)** - Dynamic control flow and state updates from nodes
- **[Configuration](configuration/)** - Using runtime configuration to pass metadata and settings

## Graph Structure & Routing

- **[Conditional Routing](conditional_routing/)** - Dynamic routing based on state
- **[Conditional Edges](conditional_edges_example/)** - Using conditional edges for branching logic
- **[Subgraphs](subgraph/)** - Composing graphs within graphs
- **[Multiple Subgraphs](subgraphs/)** - Managing multiple subgraph compositions
- **[Generic State Graph](generic_state_graph/)** - Using generic types for type-safe state management
- **[Generic State Graph Listenable](generic_state_graph_listenable/)** - Generic state graph with event listening
- **[Generic State Graph ReAct Agent](generic_state_graph_react_agent/)** - ReAct agent using generic types

## Parallel Execution

- **[Parallel Execution](parallel_execution/)** - Fan-out/Fan-in execution with state merging
- **[Complex Parallel Execution](complex_parallel_execution/)** - Advanced parallel execution with branches of varying lengths

## Streaming & Events

- **[Streaming Modes](streaming_modes/)** - Advanced streaming with updates, values, and messages modes
- **[Streaming Pipeline](streaming_pipeline/)** - Building streaming data processing pipelines
- **[Listeners](listeners/)** - Attaching event listeners to the graph
- **[Logger](logger/)** - Logging graph execution events
- **[LLM Streaming](llm_streaming/)** - Basic LLM streaming with callbacks
- **[LangChain Go Streaming](langchaingo_streaming/)** - LangChain Go integration with streaming
- **[Adapter Streaming](adapter_streaming/)** - Custom adapter streaming implementation

## Persistence (Checkpointing)

- **[Memory Basic](memory_basic/)** - In-memory checkpointing for state persistence
- **[Durable Execution](durable_execution/)** - Crash recovery and resuming execution from checkpoints
- **[File Checkpointing](file_checkpointing/)** - Checkpointing to file system
- **[File Checkpointing Resume](file_checkpointing_resume/)** - Resuming execution from file checkpoints

## Human-in-the-Loop

- **[Human in the Loop](human_in_the_loop/)** - Workflow with interrupts and human approval steps
- **[Time Travel](time_travel/)** - Inspecting, modifying state history, and forking execution
- **[Dynamic Interrupt](dynamic_interrupt/)** - Pausing execution from within a node using `graph.Interrupt`
- **[Payment Interrupt](payment_interrupt/)** - Human approval workflow example for payment scenarios
- **[API Interrupt Demo](api_interrupt_demo/)** - API-based interrupt demonstration for approval workflows

## Pre-built Agents

### Core Agent Patterns

- **[Create Agent](create_agent/)** - Easy way to create an agent with options
- **[Dynamic Skill Agent](dynamic_skill_agent/)** - Agent with dynamic skill discovery and selection
- **[ReAct Agent](react_agent/)** - Reason and Action agent using tools

### Planning & Reasoning Agents

- **[Planning Agent](planning_agent/)** - Intelligent agent that creates workflow plans based on user requests
- **[PEV Agent](pev_agent/)** - Plan-Execute-Verify agent with self-correction and error recovery
- **[Reflection Agent](reflection_agent/)** - Iterative improvement agent that refines responses through self-reflection
- **[Tree of Thoughts](tree_of_thoughts/)** - Search-based reasoning agent exploring multiple solution paths
- **[Mental Loop](mental_loop/)** - Simulator-in-the-Loop agent that tests actions in a sandbox before execution
- **[Reflexive Metacognitive Agent](reflexive_metacognitive/)** - Self-aware agent with explicit self-model of capabilities
- **[Reflexive Metacognitive Agent CN](reflexive_metacognitive_cn/)** - Chinese version of the reflexive metacognitive agent
- **[Manus Agent](manus_agent/)** - Advanced agent for document processing workflows

### Multi-Agent Systems

- **[Supervisor](supervisor/)** - Multi-agent orchestration using a supervisor pattern
- **[Swarm](swarm/)** - Multi-agent collaboration using handoffs

### Chat Agents

- **[Chat Agent](chat_agent/)** - Multi-turn conversation agent with automatic session management
- **[Chat Agent Async](chat_agent_async/)** - Asynchronous streaming chat agent with real-time LLM response streaming
- **[Chat Agent Dynamic Tools](chat_agent_dynamic_tools/)** - Chat agent with runtime tool management

## Programmatic Tool Calling (PTC)

Programmatic Tool Calling reduces latency by ~10x by having the LLM generate code instead of structured tool outputs.

- **[PTC Basic](ptc_basic/)** - Introduction to Programmatic Tool Calling
- **[PTC Simple](ptc_simple/)** - Simple PTC example with calculator and weather tools
- **[PTC Expense Analysis](ptc_expense_analysis/)** - Complex expense analysis scenario
- **[PTC + GoSkills](ptc_goskills/)** - Integration of PTC with GoSkills for local tool execution

## Memory

Conversation memory strategies for maintaining context across sessions.

- **[Memory Basic](memory_basic/)** - Basic usage of LangChain memory adapters
- **[Memory Chatbot](memory_chatbot/)** - Chatbot with LangChain memory integration
- **[Memory Strategies](memory_strategies/)** - Comprehensive guide to all 9 memory management strategies
- **[Memory Agent](memory_agent/)** - Real-world agents using different memory strategies
- **[Memory Graph Integration](memory_graph_integration/)** - State-based memory integration in LangGraph workflows
- **[Memory](memory/)** - Additional memory examples

## RAG (Retrieval Augmented Generation)

### Basic RAG

- **[RAG Basic](rag_basic/)** - Basic RAG implementation
- **[RAG Pipeline](rag_pipeline/)** - Complete RAG pipeline
- **[RAG with LangChain](rag_with_langchain/)** - RAG using LangChain components

### Advanced RAG

- **[RAG Advanced](rag_advanced/)** - Advanced RAG techniques
- **[RAG Conditional](rag_conditional/)** - Conditional RAG workflow
- **[RAG with Embeddings](rag_with_embeddings/)** - RAG using embeddings
- **[RAG Query Rewrite](rag_query_rewrite/)** - RAG with query rewriting for better retrieval
- **[LightRAG Simple](lightrag_simple/)** - Basic LightRAG usage
- **[LightRAG Advanced](lightrag_advanced/)** - Advanced LightRAG patterns

### RAG with Vector Stores

- **[RAG with VectorStores](rag_langchain_vectorstore_example/)** - RAG using LangChain VectorStores
- **[RAG with Chroma](rag_chroma_example/)** - RAG using Chroma database
- **[RAG with Chroma v2](rag_chroma-v2-example/)** - RAG using Chroma v2 database
- **[RAG with Chromem](rag_chromem_example/)** - RAG using chromem-go in-memory store
- **[RAG with Milvus](rag_milvus_example/)** - RAG using Milvus vector database
- **[RAG with Qwen Ranker](rag_qwen_ranker_example/)** - RAG with Qwen-based reranking
- **[RAG with Memory](rag_memory_example/)** - RAG with conversation memory integration

### GraphRAG (Knowledge Graph)

- **[RAG with FalkorDB Graph](rag_falkordb_graph/)** - RAG using FalkorDB knowledge graph with automatic entity extraction
- **[RAG with FalkorDB Simple](rag_falkordb_simple/)** - Simple RAG with FalkorDB using manual entity/relationship creation
- **[RAG with FalkorDB Fast](rag_falkordb_fast/)** - Optimized RAG with FalkorDB for fast queries
- **[RAG with FalkorDB Debug](rag_falkordb_debug/)** - Debug version with detailed logging
- **[RAG with FalkorDB Debug Query](rag_falkordb_debug_query/)** - Query debugging for FalkorDB RAG

## Tools & Integrations

### Search Tools

- **[Tavily Search](tool_tavily/)** - Using Tavily search tool with ReAct agent
- **[Exa Search](tool_exa/)** - Using Exa search tool with ReAct agent
- **[Brave Search](tool_brave/)** - Using Brave search API with agents

### Other Integrations

- **[GoSkills Integration](goskills_example/)** - Integrating GoSkills as tools for agents
- **[MCP Agent](mcp_agent/)** - Using Model Context Protocol (MCP) tools with agents
- **[Context Store](context_store/)** - Managing context with external stores

## Advanced Patterns

- **[Visualization](visualization/)** - Generating Mermaid diagrams for graphs

---

## Running Examples

Most examples can be run directly:

```bash
cd examples/<example_name>
go run main.go
```

Some examples may require additional setup:

1. **API Keys**: Search tools (Tavily, Exa, Brave) require API keys
2. **Databases**: FalkorDB examples require FalkorDB to be running
3. **Vector Stores**: Chroma examples require Chroma DB to be running
4. **Redis**: Some checkpointing examples require Redis

## Example Categories

| Category | Count | Description |
|----------|-------|-------------|
| Basic Concepts | 3 | Getting started examples |
| State Management | 5 | State schema and reducers |
| Graph Structure | 7 | Routing, subgraphs, generics |
| Parallel Execution | 2 | Concurrent execution patterns |
| Streaming & Events | 7 | Real-time data flow |
| Persistence | 4 | Checkpointing and recovery |
| Human-in-the-Loop | 5 | Interactive workflows |
| Pre-built Agents | 18 | Ready-to-use agent patterns |
| PTC | 4 | Programmatic tool calling |
| Memory | 6 | Conversation memory strategies |
| RAG | 22 | Retrieval-augmented generation |
| Tools & Integrations | 6 | External service integrations |
| Advanced | 1 | Visualization and debugging |
| **Total** | **85+** | **Comprehensive examples** |
