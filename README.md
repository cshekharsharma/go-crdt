# Go-CRDT: Conflict-Free Replicated Data Types

[![Go Report Card](https://goreportcard.com/badge/github.com/cshekharsharma/go-crdt)](https://goreportcard.com/report/github.com/cshekharsharma/go-crdt)
[![Go Reference](https://pkg.go.dev/badge/github.com/cshekharsharma/go-crdt.svg)](https://pkg.go.dev/github.com/cshekharsharma/go-crdt)

An educational implementation of distributed, conflict-free data structures in Go. This library provides the building blocks for creating highly available, collaborative, and distributed systems where data consistency is guaranteed without the need for central locking.

---

## What is a CRDT?

A **Conflict-free Replicated Data Type (CRDT)** is a data structure that can be replicated across multiple nodes in a network. Replicas can be updated independently and concurrently without coordination, and they are mathematically guaranteed to converge to the exact same state once all updates are received.


### The Problems It Solves
1. **Concurrency Conflicts:** Unlike traditional databases that require row-locking or "Last-Write-Wins" (which causes data loss), CRDTs use commutative logic to merge changes.
2. **Network Partitions:** Nodes can continue to work while offline or during a "split-brain" scenario and sync seamlessly when connectivity returns.
3. **High Latency:** By allowing local-first updates, users see zero latency; the "consensus" happens in the background.

---

## Features Implemented

This repository implements three fundamental CRDT variants in a modular, thread-safe, and documented manner:

### 1. G-Counter (Grow-only Counter)
A state-based counter that only moves forward.
* **Mechanism:** Maintains a vector of counts per node.
* **Merge Rule:** `Max(local_slot, remote_slot)` per index.

### 2. PN-Counter (Positive-Negative Counter)
A counter that supports both increments and decrements.
* **Mechanism:** Composed of two G-Counters: one for additions (P) and one for subtractions (N).

### 3. RGA (Replicated Growable Array)
A sophisticated sequence CRDT designed for collaborative text editing.
* **Architecture:** Hybrid **Linked-List + Hash Map** for $O(1)$ node lookup.
* **Conflict Resolution:** Uses Lamport Timestamps and NodeID tie-breaking.
* **Causal Stability:** Features a `pendingOrphans` buffer to handle out-of-order network packets.

---

## Getting Started

### Installation
```bash
go get [github.com/cshekharsharma/go-crdt](https://github.com/cshekharsharma/go-crdt)
```

### Working Example: Collaborative Text Editing
The following example simulates two users, Alice and Bob, editing the same document concurrently.

```go
package main

import (
	"fmt"
	"[github.com/cshekharsharma/go-crdt](https://github.com/cshekharsharma/go-crdt)"
)

func main() {
	// 1. Setup two independent replicas
	aliceDoc := gocrdt.NewRGA("alice")
	bobDoc := gocrdt.NewRGA("bob")
	rootID := gocrdt.ID{Timestamp: 0, NodeID: "root"}

	// 2. Alice types "Hi"
	idH := aliceDoc.Insert('H', rootID)
	aliceDoc.Insert('i', idH)

	// 3. Sync Bob with Alice (Manual merge simulation)
	bobDoc.Merge(aliceDoc)
	fmt.Printf("Bob's initial view: %v\n", bobDoc.Value()) // "Hi"

	// 4. CONCURRENT EDIT: Alice and Bob type at the same time
	// Alice types '!' after 'i'
	aliceDoc.Insert('!', gocrdt.ID{Timestamp: 2, NodeID: "alice"})
	// Bob types '?' after 'i'
	bobDoc.Insert('?', gocrdt.ID{Timestamp: 2, NodeID: "alice"})

	// 5. Final Sync
	aliceDoc.Merge(bobDoc)
	bobDoc.Merge(aliceDoc)

	// Both converged to the same string deterministically
	fmt.Printf("Alice final: %v\n", aliceDoc.Value())
	fmt.Printf("Bob final:   %v\n", bobDoc.Value())
}
```

### Development Commands
This project includes a Makefile to simplify common development tasks.

```bash
make test # Runs the full unit test suite with verbose output.
make testcoverage # Runs tests and generates an HTML coverage report.
make lint # Runs golangci-lint (if installed)
```

## Contributing
We welcome contributions! To contribute:

1. Fork the repository.
2. Create a Feature Branch (git checkout -b feature/AmazingFeature).
3. Commit your Changes (git commit -m 'Add some AmazingFeature').
4. Ensure Tests Pass: Run make test to verify your changes don't break existing CRDT logic.
5. Push to the Branch (git push origin feature/AmazingFeature).
6. Open a Pull Request.

Please ensure your code follows the existing style, includes unit tests, and has updated GoDoc comments.

## License
Distributed under the MIT License. See [LICENSE](./LICENSE) for more information.
