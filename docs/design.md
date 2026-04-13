# distkv — Learning Distributed Systems

## Context

A learning project to deeply understand distributed systems. The builder knows Python, TypeScript, and Rust, has some concurrency exposure, and 10-20 hours/week. No cloud infrastructure or LLM credits needed — everything runs on a MacBook (4-core, 16GB RAM, macOS).

The builder codes everything; guidance is provided but not implementation.

## Two-Track Strategy

**Track A (Rust — PRIMARY, do first):** Structured learning with fast feedback loops, leveraging existing Rust skills. Leads directly to OSS contributions.

**Track B (Go — SECONDARY, do after):** Build a distributed KV store from scratch in Go. Deeper ownership, learns a new language, reinforces concepts from Track A.

---

# Track A: Rust Path (~12 weeks)

## Phase A1: Gossip Glomers — All 6 Challenges in Rust (Weeks 1-3)

**Goal:** Breadth across distributed systems concepts with fast feedback loops.

Challenges:
1. Echo (already done in Go — redo in Rust for Maelstrom Rust setup)
2. Unique ID Generation — coordination-free distributed IDs
3. Broadcast — gossip protocol implementation
4. Grow-Only Counter — CRDTs / eventually consistent counters
5. Kafka-Style Log — replicated log service
6. Totally-Available Transactions — transactional key/value store

**Why this first:**
- 6 small victories for confidence building
- Each challenge is self-contained (hours, not weeks)
- Covers breadth: gossip, CRDTs, replication, transactions
- Maelstrom provides automated correctness verification

**Reading:**
- Challenge 2: DDIA Ch. 5 (Replication) — unique IDs across replicas
- Challenge 3: DDIA Ch. 5 (Replication) — gossip protocols
- Challenge 4: DDIA Ch. 5 (Replication) — CRDTs, eventual consistency
- Challenge 5: DDIA Ch. 3 (Storage) + Ch. 5 — log-structured storage
- Challenge 6: DDIA Ch. 7 (Transactions)

**Success criteria:** All 6 challenges pass Maelstrom verification.

---

## Phase A2: PingCAP Talent Plan TP 202 — Distributed Systems in Rust (Weeks 4-8)

**Goal:** Deep dive into Raft consensus with proper test infrastructure.

MIT 6.5840 adapted for Rust by the TiKV team. Includes:
- Raft leader election
- Raft log replication
- Raft persistence and snapshotting
- Test harnesses provided (no need to build your own)
- Percolator distributed transaction protocol as bonus

**Why this second:**
- Raft is the hardest and most valuable concept — do it with scaffolding
- Test harnesses catch bugs that would take days to find manually
- Same conceptual rigor as MIT 6.824
- Pipelines directly into TiKV contribution (Phase A3)

**Reading:** Raft extended paper ("In Search of an Understandable Consensus Algorithm") — Figure 2 is the complete spec.

**Success criteria:** All Talent Plan test suites pass. Can explain leader election, log replication, and safety properties.

---

## Phase A3: First Real OSS Contributions (Weeks 9-12)

**Goal:** Actual merged PRs in production distributed systems projects.

**Target projects (in order of approachability):**
1. **OpenRaft** (github.com/databendlabs/openraft) — Rust Raft library, smaller codebase, actively maintained
2. **TiKV** (github.com/tikv/tikv) — Production distributed KV store in Rust, "good first issue" labels
3. **etcd** or **NATS** (in Go) — If ready to pick up Go through bounded contributions

**Approach:**
- Start with "good first issue" labels
- Begin with tests, docs, or small bug fixes to learn the codebase
- Graduate to feature work as comfort grows

**Success criteria:** At least 1 merged PR in a distributed systems project.

---

# Track B: Go Path (~14-20 weeks, more realistic timeline)

*Do this after Track A. By then you'll understand the distributed systems concepts deeply — you'll only be learning Go, not Go + distributed systems simultaneously.*

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
│       Transport Layer (gRPC)             │  Inter-node communication
└─────────────────────────────────────────┘
```

Each "node" is a separate Go process on localhost. Nodes communicate via gRPC (protobuf). Clients interact via HTTP REST API.

## Phase B0: Go Warmup via Gossip Glomers (1 week)

**Goal:** Learn Go fundamentals. Already completed Challenge 1 (Echo).

- Complete remaining Gossip Glomers challenges in Go (or at minimum challenges 2-3)
- **Reading:** Go Tour (tour.golang.org), Effective Go

**Success criteria:** Comfortable with Go syntax, goroutines, channels, JSON, error handling.

## Phase B1: Single-Node In-Memory KV Store (1-2 weeks)

**Goal:** Build a working single-server key-value store with HTTP API.

- HTTP REST API: `PUT /keys/{key}`, `GET /keys/{key}`, `DELETE /keys/{key}`
- In-memory storage using Go `map[string][]byte` with `sync.RWMutex`
- Unit tests + integration tests
- **Reading:** Go `net/http`, `sync` docs

**Success criteria:** Store/retrieve values via curl. `go test -race ./...` passes.

## Phase B2: Multi-Node Naive Replication (2 weeks)

**Goal:** Run 3 nodes that replicate data. Experience why naive replication breaks.

- Write forwarding to all peers (sync then async)
- Heartbeat-based health checking
- Deliberate failure exercises: kill nodes, introduce delays, observe split-brain
- **Reading:** DDIA Chapter 5 (Replication)

**Success criteria:** Can articulate why naive replication is insufficient and what Raft solves.

## Phase B3: Raft Consensus (6-8 weeks)

**Goal:** Implement Raft from the paper. You'll already understand Raft deeply from Track A — this is about building it yourself from zero.

- 3a: Leader Election (1.5 weeks)
- 3b: Log Replication (2 weeks)
- 3c: Safety & Raft State Persistence (1-1.5 weeks)
- 3d: Log Compaction / Snapshotting (1 week)
- **Reading:** Raft paper (re-read with implementation lens)

**Success criteria:** Leader election + log replication works. Writes survive leader failure.

## Phase B4: Persistent Storage Engine (1-2 weeks)

**Goal:** WAL + crash recovery for the KV data layer.

- Write-Ahead Log, crash recovery via replay, periodic snapshots
- **Reading:** DDIA Chapter 3

**Success criteria:** `kill -9` + restart recovers all committed data.

## Phase B5: Sharding with Consistent Hashing (2-3 weeks)

**Goal:** Distribute keys across multiple Raft groups.

- Consistent hashing ring, shard routing, rebalancing
- **Reading:** DDIA Chapter 6

**Success criteria:** Multi-shard cluster with independent leader failure handling per shard.

## Phase B6: Fault Injection & Chaos Testing (1-2 weeks)

**Goal:** Prove correctness under failure.

- Kill leader during writes, network partitions, slow nodes, concurrent writes
- **Reading:** Jepsen blog posts

**Success criteria:** All chaos tests pass 100x consistently.

---

## Running Track B Locally

```bash
# Start a 3-node cluster
./distkv --id=1 --port=8001 --peers=localhost:8002,localhost:8003
./distkv --id=2 --port=8002 --peers=localhost:8001,localhost:8003
./distkv --id=3 --port=8003 --peers=localhost:8001,localhost:8002

# Client operations
curl -X PUT localhost:8001/keys/name -d '{"value":"alice"}'
curl localhost:8002/keys/name  # reads from any node
```

Resource usage: ~20-50MB per node. 16GB machine handles 6+ nodes easily.

---

## Combined Timeline

| Track | Phase | Duration |
|---|---|---|
| **A (Rust)** | A1: Gossip Glomers | 3 weeks |
| | A2: Talent Plan Raft | 5 weeks |
| | A3: OSS Contributions | 4 weeks |
| | **Track A Total** | **~12 weeks** |
| **B (Go)** | B0-B1: Go + Single KV | 2-3 weeks |
| | B2: Naive Replication | 2 weeks |
| | B3: Raft from Scratch | 6-8 weeks |
| | B4-B6: Storage + Sharding + Chaos | 4-7 weeks |
| | **Track B Total** | **~14-20 weeks** |
| | **Grand Total** | **~26-32 weeks** |

## What This Does NOT Include (YAGNI)

- No distributed transactions (2PC/3PC)
- No Byzantine fault tolerance
- No authentication/authorization
- No UI/dashboard
- No Docker/Kubernetes deployment
- No cloud infrastructure

## Companion Reading (Both Tracks)

The bible: **Designing Data-Intensive Applications** by Martin Kleppmann

| Concept | DDIA Chapter |
|---|---|
| Replication, gossip, CRDTs | Ch. 5 |
| Storage engines, WAL | Ch. 3 |
| Partitioning / sharding | Ch. 6 |
| Transactions | Ch. 7 |
| Consensus (Raft) | Ch. 9 |

Other key readings:
- Raft paper: "In Search of an Understandable Consensus Algorithm"
- Jepsen blog posts (aphyr.com)
- "Simple Testing Can Prevent Most Critical Failures" paper

## OSS Contribution Targets

**Rust ecosystem:**
- OpenRaft (databendlabs/openraft) — Rust Raft library
- TiKV (tikv/tikv) — distributed KV store

**Go ecosystem (after Track B):**
- NATS (nats-io/nats-server) — messaging system
- etcd (etcd-io/etcd) — distributed KV store
- Dragonboat (lni/dragonboat) — multi-group Raft library

## Previous Work

- Gossip Glomers Challenge 1 (Echo) completed in Go — validates Maelstrom setup
- Gossip Glomers Challenge 1 (Echo) completed in Rust (2026-04-02) — validates Rust+Maelstrom toolchain
  - Rust workspace at `rust/`, challenge at `rust/c01-echo/`
  - Dependencies: serde, serde_json
  - Maelstrom verification: 50/50, zero failures
