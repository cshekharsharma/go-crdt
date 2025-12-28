package gocrdt

import "testing"

func TestPNCounter_Basic(t *testing.T) {
	counter := NewPNCounter("node-a")

	counter.Increment()
	counter.Increment()
	counter.Decrement()

	if counter.Value() != 1 {
		t.Errorf("Expected 1, got %d", counter.Value())
	}
}

func TestPNCounter_Merge(t *testing.T) {
	nodeA := NewPNCounter("node-a")
	nodeB := NewPNCounter("node-b")

	nodeA.Increment() // A = 1
	nodeB.Decrement() // B = -1

	nodeA.Merge(nodeB)
	nodeB.Merge(nodeA)

	if nodeA.Value() != 0 || nodeB.Value() != 0 {
		t.Errorf("Expected convergence at 0, got A=%d, B=%d", nodeA.Value(), nodeB.Value())
	}
}
