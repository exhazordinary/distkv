# distkv — Distributed Key-Value Store from Scratch

## Context

A learning project to deeply understand distributed systems by building a multi-node, Raft-based, sharded key-value store from scratch in Go. The builder is new to Go (experienced in Python, TypeScript, Rust) and has some concurrency exposure. No cloud infrastructure or LLM credits needed — everything runs on a MacBook (4-core, 16GB RAM, macOS).

The project uses a phase-based approach where each phase produces a working system. The builder codes everything; guidance is provided but not implementation.

## Architecture

```
┌─────────────────────────────────────────┐
│           Client CLI / HTTP API          │  Clients interact here
├─────────────────────────────────────────┤
│          Router / Shard Manager          │  Routes keys to correct shard group
├─────────────────────────────────────────┤
│         Raft Consensus (per shard)       │  Ensures all replicas agree on state
├─────────────────────────────────────────┤
│        Storage Engine (per node)         │  In-memory → WAL + snapshots
├─────────────────────────────────────────┤
│       Transport Layer (gRPC/TCP)         │  Inter-node communication
└─────────────────────────────────────────┘
```

Each "node" is a separate Go process on localhost with its own port. Nodes communicate via gRPC (protobuf). Clients interact via HTTP REST API.

## Phase Breakdown

### Phase 0: Go Warmup via Gossip Glomers (3-5 days)

**Goal:** Learn Go fundamentals through distributed systems challenges.

**Tasks:**
- Install Maelstrom (requires Java/JVM)
- Complete Gossip Glomers Challenge 1 (Echo) — learn Go project structure, JSON marshaling, stdin/stdout I/O
- Complete Gossip Glomers Challenge 2 (Unique ID Generation) — learn about coordination-free ID generation

**Go concepts covered:** Variables, structs, interfaces, error handling, goroutines, channels, JSON encoding/decoding, maps, slices, `go mod`, testing basics.

**Reading:** Go Tour (tour.golang.org), Effective Go (go.dev/doc/effective_go)

**Success criteria:** Both challenges pass Maelstrom's verification. Can explain what a goroutine is and when to use channels.

---

### Phase 1: Single-Node In-Memory KV Store (1 week)

**Goal:** Build a working single-server key-value store with HTTP API.

**Features:**
- HTTP REST API: `PUT /keys/{key}`, `GET /keys/{key}`, `DELETE /keys/{key}`, `GET /keys` (list all)
- In-memory storage using Go `map[string][]byte` with `sync.RWMutex` for thread safety
- Proper HTTP status codes (200, 201, 404, 409)
- Basic CLI client or curl-based testing
- Unit tests for storage layer and integration tests for HTTP API

**Architecture:**
```
main.go          — Entry point, flag parsing, server startup
server/
  handler.go     — HTTP handlers (Get, Put, Delete, List)
  router.go      — HTTP router setup
store/
  memory.go      — In-memory KV store with mutex
  store.go       — Store interface definition
```

**Key design decisions for the builder:**
- Store interface design (what methods? what return types?)
- Error handling strategy (custom error types vs. sentinel errors)
- Key/value constraints (max key length? value size limit?)

**Reading:** Go standard library `net/http` docs, `sync` package docs

**Success criteria:** Can store and retrieve values via curl. Tests pass. Concurrent reads don't race (verified with `go test -race`).

---

### Phase 2: Multi-Node Naive Replication (1.5-2 weeks)

**Goal:** Run 3 nodes that replicate data. Discover why naive replication breaks.

**Features:**
- Cluster configuration (node addresses via config file or CLI flags)
- Write forwarding: writes to any node are forwarded to all others
- Synchronous replication first (wait for all ACKs before responding)
- Then async replication (respond after local write, replicate in background)
- Node health checking via periodic heartbeats

**Architecture additions:**
```
cluster/
  config.go      — Cluster topology (node addresses, IDs)
  transport.go   — gRPC client/server for inter-node RPC
  replicator.go  — Replication logic (sync and async modes)
```

**Concepts learned:**
- Network partitions: what happens when node 3 can't reach node 1?
- Split brain: two nodes accept conflicting writes
- Consistency vs availability trade-off (CAP theorem, experienced firsthand)
- Why you need consensus (motivation for Phase 3)

**Deliberate failure exercises:**
1. Kill node 2 mid-write — observe inconsistency
2. Introduce 2-second network delay — observe timeout behavior
3. Partition node 3 from nodes 1-2 — observe split-brain

**Reading:** DDIA Chapter 5 (Replication)

**Success criteria:** 3-node cluster replicates data. Builder can articulate *why* naive replication is insufficient and what problems Raft solves.

---

### Phase 3: Raft Consensus (3-4 weeks)

**Goal:** Implement the Raft consensus protocol from the paper. The core of the project.

This phase is subdivided into 4 sub-phases:

#### 3a: Leader Election (1 week)
- Nodes start as followers
- If a follower doesn't hear from a leader within a randomized timeout, it becomes a candidate
- Candidates request votes from all nodes; majority wins
- Only one leader per term
- Implement `RequestVote` RPC

#### 3b: Log Replication (1-1.5 weeks)
- Leader receives client writes, appends to its log
- Leader sends `AppendEntries` RPC to replicate log entries to followers
- Entry is "committed" once majority of nodes have it
- Committed entries are applied to the KV state machine
- Followers reject entries that don't match their log (consistency check)

#### 3c: Safety & Raft State Persistence (0.5-1 week)
- Election restriction: only candidates with up-to-date logs can win
- Commitment rules: leader only commits entries from its own term
- Persist Raft-specific state (`currentTerm`, `votedFor`, `log[]`) to disk — this is distinct from Phase 4's KV data persistence. Here we only persist what Raft needs to recover its protocol state after a restart.

#### 3d: Log Compaction / Snapshotting (0.5 week)
- Periodically snapshot the KV state
- Discard log entries before the snapshot
- `InstallSnapshot` RPC for slow followers

**Architecture additions:**
```
raft/
  raft.go        — Core Raft state machine (follower/candidate/leader)
  log.go         — Log entry storage and indexing
  rpc.go         — RequestVote and AppendEntries RPC definitions
  state.go       — Persistent state management
  snapshot.go    — Snapshot creation and installation
```

**Reading:** "In Search of an Understandable Consensus Algorithm" (Raft paper) — read Figure 2 carefully, it's the complete spec.

**Success criteria:**
- Leader election works with 3 and 5 nodes
- Writes survive leader failure (kill leader, new leader elected, data intact)
- Raft tests pass: election, basic agreement, fail agree, rejoin, backup, persistence, snapshot

---

### Phase 4: Persistent Storage Engine (1-2 weeks)

**Goal:** Make the KV data itself durable. Phase 3c persisted Raft protocol state (term, votes, log); this phase persists the actual key-value data so it survives node restarts.

**Features:**
- Write-Ahead Log (WAL): every mutation appended to disk before applied to memory
- Crash recovery: replay WAL on startup to rebuild state
- Periodic snapshots to bound WAL size
- Simple file-based storage (no need for LSM trees at this stage)

**Architecture additions:**
```
storage/
  wal.go         — Append-only write-ahead log
  snapshot.go    — Point-in-time state serialization
  engine.go      — Storage engine combining WAL + snapshots + in-memory index
```

**Reading:** DDIA Chapter 3 (Storage and Retrieval) — focus on WAL and log-structured storage

**Success criteria:** Kill a node (`kill -9`), restart it, all committed data is recovered. No data loss on clean or unclean shutdown.

---

### Phase 5: Sharding with Consistent Hashing (2 weeks)

**Goal:** Distribute keys across multiple Raft groups for horizontal scalability.

**Features:**
- Consistent hashing ring to map keys to shard groups
- Each shard is an independent Raft group (3 replicas)
- Router layer that directs client requests to the correct shard
- Shard rebalancing when adding/removing groups
- Cross-shard key listing

**Architecture additions:**
```
shard/
  ring.go        — Consistent hashing ring implementation
  router.go      — Request routing based on key → shard mapping
  manager.go     — Shard assignment and rebalancing logic
```

**Configuration example (6 nodes, 2 shards):**
```
Shard 1 (keys hash 0-127):   Node 1 (leader), Node 2, Node 3
Shard 2 (keys hash 128-255): Node 4 (leader), Node 5, Node 6
```

**Reading:** DDIA Chapter 6 (Partitioning)

**Success criteria:** Data is distributed across 2+ shards. Adding a new shard group triggers rebalancing. Each shard independently handles leader failure.

---

### Phase 6: Fault Injection & Chaos Testing (1 week)

**Goal:** Prove the system works under failure. Build confidence in correctness.

**Tests to implement:**
- Kill leader during active writes — verify new leader elected, no data loss
- Network partition (isolate minority) — verify majority continues serving
- Network partition (isolate leader) — verify leader steps down, new election
- Slow node (artificial 500ms delay) — verify system doesn't stall
- Concurrent writes to same key — verify linearizability
- Node restart with corrupted WAL — verify graceful degradation

**Tools:**
- Go test framework with helper functions to kill/restart nodes
- Artificial network delays via goroutine-level simulation
- Toxiproxy (optional) for process-level network simulation

**Reading:** Jepsen blog posts (aphyr.com), "Simple Testing Can Prevent Most Critical Failures" paper

**Success criteria:** All chaos tests pass consistently (run 100x with no failures). Can explain each failure mode and how the system handles it.

---

## Technology Choices

| Component | Choice | Rationale |
|---|---|---|
| Language | Go 1.25 | Best distributed systems ecosystem, goroutines for node simulation |
| Inter-node RPC | gRPC (protobuf) | Industry standard, strongly typed, good Go support |
| Client API | HTTP REST | Simple, curl-testable, familiar |
| Serialization | Protocol Buffers | For inter-node; JSON for client API |
| Testing | Go `testing` + `testify` | Standard, good assertion library |
| Build | Go modules | Standard Go dependency management |

## What This Project Does NOT Include (YAGNI)

- No distributed transactions (2PC/3PC)
- No Byzantine fault tolerance (assumes nodes are honest, just crash-faulty)
- No authentication/authorization
- No production-grade performance optimization
- No UI/dashboard
- No Docker/Kubernetes deployment
- No cloud infrastructure

## Running Locally

```bash
# Start a 3-node cluster
./distkv --id=1 --port=8001 --peers=localhost:8002,localhost:8003
./distkv --id=2 --port=8002 --peers=localhost:8001,localhost:8003
./distkv --id=3 --port=8003 --peers=localhost:8001,localhost:8002

# Client operations
curl -X PUT localhost:8001/keys/name -d '{"value":"alice"}'
curl localhost:8002/keys/name  # reads from any node
```

Resource usage: ~20-50MB per node. 3-node cluster uses <200MB. 16GB machine can run 6+ nodes with multiple shard groups.

## Verification Plan

Each phase has its own verification:

- **Phase 0:** Maelstrom verification passes for challenges 1-2
- **Phase 1:** `go test -race ./...` passes; manual curl testing
- **Phase 2:** 3-node cluster replicates; deliberate failure exercises completed
- **Phase 3:** Raft test suite (election, agreement, persistence, snapshot) passes 100/100 runs
- **Phase 4:** `kill -9` + restart recovers all committed data
- **Phase 5:** Multi-shard cluster with rebalancing works; shard-level leader failure handled
- **Phase 6:** All chaos tests pass 100x consecutively

## Companion Reading Schedule

| Phase | Reading |
|---|---|
| 0 | Go Tour, Effective Go |
| 1 | Go `net/http`, `sync` docs |
| 2 | DDIA Ch. 5 (Replication) |
| 3 | Raft paper (Figure 2 is the spec) |
| 4 | DDIA Ch. 3 (Storage & Retrieval) |
| 5 | DDIA Ch. 6 (Partitioning) |
| 6 | Jepsen posts, "Simple Testing" paper |

## Timeline

At 10-20 hours/week:

| Phase | Duration |
|---|---|
| Phase 0: Go Warmup | 3-5 days |
| Phase 1: Single-Node KV | 1 week |
| Phase 2: Naive Replication | 1.5-2 weeks |
| Phase 3: Raft Consensus | 3-4 weeks |
| Phase 4: Persistent Storage | 1-2 weeks |
| Phase 5: Sharding | 2 weeks |
| Phase 6: Chaos Testing | 1 week |
| **Total** | **10-14 weeks** |

## OSS Contribution Pathway

After completing this project, you'll have hands-on experience with the exact concepts used in production distributed systems. Natural OSS targets:

- **NATS** (nats-io/nats-server) — messaging system, simpler codebase
- **etcd** (etcd-io/etcd) — distributed KV store (you'll have built a smaller version!)
- **Dragonboat** (lni/dragonboat) — multi-group Raft library (directly relevant to Phase 3-5)

Start with documentation, test improvements, or "good first issue" labels. Your distkv project demonstrates you understand the domain — maintainers notice that.
