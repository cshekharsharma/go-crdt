# Go-CRDT: Conflict-Free Replicated Data Types

[![Go Report Card](https://goreportcard.com/badge/github.com/cshekharsharma/go-crdt)](https://goreportcard.com/report/github.com/cshekharsharma/go-crdt)
[![Go Reference](https://pkg.go.dev/badge/github.com/cshekharsharma/go-crdt.svg)](https://pkg.go.dev/github.com/cshekharsharma/go-crdt)

A production-grade implementation of distributed, conflict-free data structures in Go. This library provides the building blocks for creating highly available, collaborative, and distributed systems where data consistency is guaranteed without the need for central locking.

---

## What is a CRDT?

A **Conflict-free Replicated Data Type (CRDT)** is a data structure that can be replicated across multiple nodes in a network. Replicas can be updated independently and concurrently without coordination, and they are mathematically guaranteed to converge to the exact same state once all updates are received.

### The Problems It Solves
1. **Concurrency Conflicts:** Unlike traditional databases that require row-locking or "Last-Write-Wins" (which causes data loss), CRDTs use commutative logic to merge changes.
2. **Network Partitions:** Nodes can continue to work while offline or during a "split-brain" scenario and sync seamlessly when connectivity returns.
3. **High Latency:** By allowing local-first updates, users see zero latency; the "consensus" happens in the background.

---

## 🛠 Features Implemented

This repository implements three fundamental CRDT variants in a modular, thread-safe, and documented manner:

### 1. G-Counter (Grow-only Counter)
A state-based counter that only moves forward.
* **Mechanism:** Maintains a vector of counts per node.
* **Merge Rule:** `Max(local_slot, remote_slot)` per index.
* **Use Case:** Distributed page views, total "likes" across regions.

### 2. PN-Counter (Positive-Negative Counter)
A counter that supports both increments and decrements.
* **Mechanism:** Composed of two G-Counters: one for additions (P) and one for subtractions (N).
* **Formula:** `Value = Sum(P) - Sum(N)`.
* **Use Case:** Live user counts, inventory management.

### 3. RGA (Replicated Growable Array)
A sophisticated sequence CRDT designed for collaborative text editing (similar to the logic used in Figma or Apple Notes).
* **Architecture:** Hybrid **Linked-List + Hash Map** for $O(1)$ node lookup and efficient insertion.
* **Conflict Resolution:** Uses Lamport Timestamps and NodeID tie-breaking to sort siblings deterministically.
* **Causal Stability:** Features a `pendingOrphans` buffer to handle out-of-order network packets (children arriving before parents).
* **Tombstones:** Supports deletions by marking nodes as "deleted" rather than removing them, preserving the integrity of the causal chain.

---

## 🚀 Getting Started

### Installation
```bash
go get [github.com/cshekharsharma/go-crdt](https://github.com/cshekharsharma/go-crdt)