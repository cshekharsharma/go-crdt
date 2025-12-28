package gocrdt

// PNCounter is a Positive-Negative Counter CRDT.
//
// Unlike a GCounter, which is increment-only, a PNCounter allows for both
// increments and decrements. It achieves this by internally managing two
// independent G-Counters:
//   - The "P" counter tracks the sum of all increments.
//   - The "N" counter tracks the sum of all decrements.
//
// This structure ensures that even when nodes decrement values, the underlying
// state remains monotonic (always growing), which is a requirement for
// successful merging in distributed systems.
type PNCounter struct {
	pCounter *GCounter // Increments
	nCounter *GCounter // Decrements
}

// NewPNCounter initializes a PNCounter for a specific node.
// It creates two underlying GCounters, both sharing the same nodeID to
// track that node's specific contribution to the global sum and delta.
func NewPNCounter(nodeID string) *PNCounter {
	return &PNCounter{
		pCounter: NewGCounter(nodeID),
		nCounter: NewGCounter(nodeID),
	}
}

// Increment adds 1 to the counter.
// Internally, this increases the value in the positive GCounter.
func (c *PNCounter) Increment() {
	c.pCounter.Increment()
}

// Decrement subtracts 1 from the counter.
// Internally, this increases the value in the negative GCounter.
// Note: We "increment" the negative state to represent a "decrement"
// of the total value.
func (c *PNCounter) Decrement() {
	c.nCounter.Increment()
}

// Value calculates the current total by subtracting the negative GCounter sum
// from the positive GCounter sum.
//
// This represents the "drift" between all additions and all subtractions
// known by the node. This method satisfies the CRDT interface.
func (c *PNCounter) Value() int {
	return c.pCounter.Value() - c.nCounter.Value()
}

// Merge combines the state of another PNCounter into this one.
//
// The merge is performed by independently merging the underlying positive
// and negative GCounters. Since both underlying counters satisfy the
// properties of a Join-Semilattice, the PNCounter merge is also commutative,
// associative, and idempotent.
func (c *PNCounter) Merge(other *PNCounter) {
	c.pCounter.Merge(other.pCounter)
	c.nCounter.Merge(other.nCounter)
}
