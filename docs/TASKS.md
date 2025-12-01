# LangGraphGo Implementation Roadmap

This document outlines the step-by-step plan to implement the features listed in `TODOs.md`.

## Phase 1: Core Runtime Enhancements
Focus on the fundamental execution model and configuration to support complex graphs.

- [x] **Parallel Execution (Fan-out / Fan-in)**
    - [x] Design concurrent node execution model in `graph.go`.
    - [x] Implement `ExecuteParallel` in `pregel` or `graph` package.
    - [x] Add synchronization mechanism (WaitGroups/Channels) for fan-in.
    - [x] Ensure thread-safe state merging.
    - [x] Add unit tests for parallel execution scenarios.

- [x] **Runtime Configuration**
    - [x] Define `RunnableConfig` struct (similar to Python's `configurable`).
    - [x] Update `Invoke` and `Stream` signatures to accept `RunnableConfig`.
    - [x] Propagate config to `Node` context.
    - [x] Add helper function to retrieve config from context.

## Phase 2: Persistence & Reliability
Enable durable execution and state recovery.

- [x] **Persistent Checkpoint Interface**
    - [x] Refine `CheckpointSaver` interface in `checkpoint` package.
    - [x] Define serialization format for State (JSON/Gob).

- [x] **Redis Checkpointer**
    - [x] Create `checkpoint/redis` package.
    - [x] Implement `Put` and `Get` using Redis client.
    - [x] Add integration tests with Redis (mock or container).

- [x] **Postgres Checkpointer**
    - [x] Create `checkpoint/postgres` package.
    - [x] Design schema for checkpoints.
    - [x] Implement `Put` and `Get` using `database/sql` or `pgx`.

- [x] **SQLite Checkpointer**
    - [x] Create `checkpoint/sqlite` package.
    - [x] Implement file-based persistence for local development.

## Phase 3: Advanced Features
Enhance the capabilities of the graph for complex agentic behaviors.

- [x] **Advanced State Management**
    - [x] Design `Schema` interface for state validation.
    - [x] Implement `Annotated` style reducers (e.g., `AppendMessages`).
    - [x] Refactor `MessageGraph` to use the new state system.

- [x] **Enhanced Streaming**
    - [x] Define `StreamEvent` types (NodeStart, NodeEnd, Token, etc.).
    - [x] Implement `CallbackHandler` interface for granular events.
    - [x] Update `StreamingExecutor` to emit typed events.o Node context.

- [x] **Pre-built Agentic Components**
    - [x] Implement `ToolExecutor` node.
    - [x] Implement `ReAct` agent factory.
    - [x] Implement `Supervisor` agent factory.

## Phase 4: Developer Experience & Human-in-the-loop
Tools for debugging, visualizing, and controlling execution.

- [x] **Visualization Improvements**
    - [x] Update `Exporter` to traverse and render conditional edges.
    - [x] Add styling options to `DrawMermaid`.

- [x] **Human-in-the-loop (HITL)**
    - [x] Implement `Interrupt` mechanism in graph execution.
    - [x] Add `Command` support to resume/update state.
    - [x] Create an example of a "Human Approval" workflow.

## Phase 5: Future & Research
Long-term architectural improvements.

- [x] **Multi-Agent Collaboration**
    - [x] Prototype "Swarm" patterns using `Subgraph`.
    - [x] Created `examples/swarm/main.go`.

- [x] **Channels Architecture**
    - [x] Research Python's Channel implementation.
    - [x] Propose RFC for Go implementation if beneficial.
    - [x] Created `RFC_CHANNELS.md`.
