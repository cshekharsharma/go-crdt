package gocrdt

import "testing"

func TestGCounter_Convergence(t *testing.T) {
	nodeA := NewGCounter("node-a")
	nodeB := NewGCounter("node-b")

	nodeA.Increment()
	nodeA.Increment()
	nodeB.Increment()

	// Cross-merge
	nodeA.Merge(nodeB)
	nodeB.Merge(nodeA)

	if nodeA.Value() != 3 || nodeB.Value() != 3 {
		t.Errorf("Expected convergence at 3, got A=%d, B=%d", nodeA.Value(), nodeB.Value())
	}

	nodeA.Merge(nodeB)
	if nodeA.Value() != 3 {
		t.Errorf("Idempotency failed: expected 3, got %d", nodeA.Value())
	}
}
