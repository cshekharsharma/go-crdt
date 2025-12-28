package gocrdt

import "sync"

// GCounter is a state-based Grow-only Counter CRDT.
//
// It is a distributed counter where the value only increases (increments).
// To prevent double-counting across different nodes, it maintains a vector
// (map) of counts, where each node is responsible for updating its own slot.
//
// The total value is derived by summing all slots in the map.
type GCounter struct {
	mu     sync.RWMutex
	nodeID string
	// slots maps NodeID -> Current Count for that node
	slots map[string]int
}

// NewGCounter initializes a GCounter for a specific node.
// The nodeID must be unique across the entire distributed system to ensure
// that increments from different sources do not overwrite each other.
func NewGCounter(nodeID string) *GCounter {
	return &GCounter{
		nodeID: nodeID,
		slots:  make(map[string]int),
	}
}

// Increment adds 1 to the local node's slot in the counter.
// This operation is thread-safe and affects only the entry corresponding
// to the nodeID provided during initialization.
func (c *GCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.slots[c.nodeID]++
}

// Value returns the sum of all slots, representing the global total count.
// This method satisfies the CRDT interface. Even if the network is partitioned,
// this returns the most complete count currently known by the local node.
func (c *GCounter) Value() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sum := 0
	for _, value := range c.slots {
		sum += value
	}
	return sum
}

// Merge combines the state of another GCounter into this one.
//
// It implements the Join-Semilattice "join" operation by taking the maximum
// value for each node ID found in either counter. This ensures that the
// merge is:
//   - Commutative: A merged with B is the same as B merged with A.
//   - Associative: (A merged with B) merged with C is the same as A merged with (B merged with C).
//   - Idempotent: Merging the same counter multiple times does not change the result.
func (c *GCounter) Merge(other *GCounter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	for id, value := range other.slots {
		if value > c.slots[id] {
			c.slots[id] = value
		}
	}
}
