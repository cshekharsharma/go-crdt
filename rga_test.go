package gocrdt

import (
	"testing"
)

func TestRGA_FullLifeCycle(t *testing.T) {
	alice := NewRGA("alice")
	bob := NewRGA("bob")
	rootID := ID{0, "root"}

	// 1. Basic Sequential Insert
	idH := alice.Insert('H', rootID)
	idE := alice.Insert('E', idH)

	// Sync Bob
	bob.Merge(getNodes(alice))
	if bob.Value() != "HE" {
		t.Fatalf("Bob sync failed, got: %s", bob.Value())
	}

	// 2. Concurrent Sibling Insert
	// Alice types 'L' after 'E' -> HEL
	alice.Insert('L', idE)
	// Bob types 'Y' after 'E' -> HEY
	bob.Insert('Y', idE)

	// Cross Merge
	aliceState := getNodes(alice)
	bobState := getNodes(bob)

	alice.Merge(bobState)
	bob.Merge(aliceState)

	if alice.Value() != bob.Value() {
		t.Errorf("Divergence! Alice: %s, Bob: %s", alice.Value(), bob.Value())
	}

	// Deterministic order: 'Y' (bob) > 'L' (alice) because NodeID 'bob' > 'alice'
	if alice.Value() != "HEYL" {
		t.Errorf("Expected HEYL, got %s", alice.Value())
	}
}

func TestRGA_CausalOrderFixed(t *testing.T) {
	r := NewRGA("client")
	rootID := ID{0, "root"}

	parentID := ID{Timestamp: 10, NodeID: "server"}
	childID := ID{Timestamp: 11, NodeID: "server"}

	parent := Node{ID: parentID, ParentID: rootID, Value: 'P'}
	child := Node{ID: childID, ParentID: parentID, Value: 'C'}

	// Merge Child FIRST (Parent is missing)
	r.Merge([]Node{child})
	if r.Value() != "" {
		t.Errorf("Should be empty, waiting for parent. Got: %s", r.Value())
	}

	// Merge Parent SECOND
	r.Merge([]Node{parent})

	if r.Value() != "PC" {
		t.Errorf("Causal resolution failed. Expected PC, got: %s", r.Value())
	}
}

func TestRGA_TimestampPriority(t *testing.T) {
	alice := NewRGA("alice")
	bob := NewRGA("bob")
	rootID := ID{0, "root"}

	// 1. Setup: Both have "H"
	idH := alice.Insert('H', rootID)
	bob.Merge([]Node{*alice.registry[idH]})

	// 2. Alice performs TWO operations to push her local clock forward
	// Alice: H -> X -> A (Timestamp for 'A' will be higher)
	_ = alice.Insert('X', idH)
	idA := alice.Insert('A', idH) // Alice's clock is now at 3

	// 3. Bob performs ONE operation after 'H'
	// Bob: H -> B
	// Bob's clock was at 1 (from H), so this insert will be at Timestamp 2
	idB := bob.Insert('B', idH)

	if idA.Timestamp <= idB.Timestamp {
		t.Errorf("Setup failed: Alice's timestamp (%d) should be > Bob's (%d)", idA.Timestamp, idB.Timestamp)
	}

	// 4. Merge
	alice.Merge(getNodes(bob))
	bob.Merge(getNodes(alice))

	// 5. Logic Check:
	// Both 'A', 'X', and 'B' have the same parent 'H'.
	// In the sibling list, they should be ordered by Timestamp DESC.
	// Order should be: H -> (Highest TS: A) -> (Next TS: X/B) ...
	// Since 'A' has T=3 and 'B' has T=2, 'A' MUST come before 'B'.

	text := alice.Value()

	// We check if 'A' appears before 'B' in the string
	foundA := false
	for _, char := range text.(string) {
		if char == 'A' {
			foundA = true
		}
		if char == 'B' && !foundA {
			t.Errorf("Timestamp sorting failed: 'B' appeared before 'A'. Text: %s", text)
		}
	}

	t.Logf("Final Text with TS Priority: %s", text)
}

func TestRGA_Tombstones(t *testing.T) {
	r := NewRGA("alice")
	id1 := r.Insert('A', ID{0, "root"})
	r.Delete(id1)

	if r.Value() != "" {
		t.Errorf("Expected empty string, got %s", r.Value())
	}
	if len(r.registry) != 2 { // root + A
		t.Errorf("Registry should keep tombstones")
	}
}

func TestRGA_RemoteDeletionPropagation(t *testing.T) {
	alice := NewRGA("alice")
	bob := NewRGA("bob")
	rootID := ID{0, "root"}

	// 1. Setup: Alice types "Hi" and syncs with Bob
	idH := alice.Insert('H', rootID)
	idI := alice.Insert('i', idH)

	// Sync: Bob now has "Hi"
	bob.Merge(getNodes(alice))
	if bob.Value() != "Hi" {
		t.Fatalf("Setup failed: Bob should have 'Hi', got %s", bob.Value())
	}

	// 2. Action: Alice deletes 'i' locally
	// This sets Deleted = true for idI in Alice's registry
	alice.Delete(idI)
	if alice.Value() != "H" {
		t.Errorf("Alice local delete failed: expected 'H', got %s", alice.Value())
	}

	// 3. Merge: Bob merges Alice's state again
	// This triggers: if _, exists := r.registry[n.ID]; exists { if n.Deleted { ... } }
	bob.Merge(getNodes(alice))

	if bob.Value() != "H" {
		t.Errorf("Remote deletion failed to propagate: Bob still has %s", bob.Value())
	}

	if node, exists := bob.registry[idI]; !exists || !node.Deleted {
		t.Error("Bob's registry entry for 'i' should exist and be marked as Deleted")
	}
}

func getNodes(r *RGA) []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var nodes []Node
	for _, n := range r.registry {
		if n.ID.NodeID != "root" {
			nodes = append(nodes, *n)
		}
	}
	return nodes
}
