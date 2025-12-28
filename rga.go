package gocrdt

import "sync"

// ID represents a unique identifier for an element in the RGA.
// It uses a Lamport Timestamp combined with a unique NodeID to establish
// a "happened-before" relationship and ensure deterministic ordering
// across distributed replicas.
type ID struct {
	Timestamp int64
	NodeID    string
}

// Greater compares two IDs to provide a total ordering of elements.
// This is used to resolve conflicts when two users perform concurrent
// inserts after the same parent node. The sorting prioritizes the
// higher Timestamp, using the NodeID as a tie-breaker.
func (a ID) Greater(b ID) bool {
	if a.Timestamp != b.Timestamp {
		return a.Timestamp > b.Timestamp
	}
	return a.NodeID > b.NodeID
}

// Node represents a single element (typically a character) in the
// replicated sequence. It maintains metadata required for linking
// and conflict resolution.
type Node struct {
	ID       ID    // Unique identifier for this node
	ParentID ID    // The ID of the node this element was inserted after
	Value    rune  // The actual character or data value
	Deleted  bool  // Tombstone flag to mark logical deletion
	Next     *Node // Pointer to the next node in the linearized view
}

// RGA is a Replicated Growable Array CRDT designed for collaborative
// sequence editing.
//
// RGA uses a Linked-List structure to represent the document and a
// Hash Map (registry) to provide O(1) random access to any node by its ID.
// This hybrid approach allows for high-performance insertions and
// deletions in large documents.
type RGA struct {
	mu             sync.RWMutex
	nodeID         string
	clock          int64
	registry       map[ID]*Node
	root           *Node
	pendingOrphans map[ID][]Node // Buffer for causal consistency
}

// NewRGA initializes a new RGA instance for a given node.
// It creates a sentinel "root" node which serves as the anchor
// for the beginning of the sequence.
func NewRGA(nodeID string) *RGA {
	rootID := ID{0, "root"}
	rootNode := &Node{ID: rootID}
	return &RGA{
		nodeID:         nodeID,
		registry:       map[ID]*Node{rootID: rootNode},
		root:           rootNode,
		pendingOrphans: make(map[ID][]Node),
	}
}

// Insert creates a new element in the sequence after the specified
// parentID. It increments the local logical clock and integrates
// the new node into the local state.
func (r *RGA) Insert(val rune, parentID ID) ID {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clock++
	newID := ID{r.clock, r.nodeID}
	newNode := &Node{
		ID:       newID,
		ParentID: parentID,
		Value:    val,
	}

	r.integrate(newNode)
	return newID
}

// Delete marks a node as logically deleted (a "Tombstone").
// Nodes are not physically removed from the registry or linked-list
// to ensure that concurrent operations referencing this node can
// still be resolved correctly.
func (r *RGA) Delete(id ID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if node, exists := r.registry[id]; exists {
		node.Deleted = true
	}
}

// Merge incorporates remote state into the local RGA.
//
// It handles deduplication of nodes and ensures Causal Consistency
// by buffering "orphan" nodes whose parents have not yet arrived
// from the network. Once a missing parent is integrated, its
// buffered children are automatically processed.
func (r *RGA) Merge(remoteNodes []Node) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, n := range remoteNodes {
		if _, exists := r.registry[n.ID]; exists {
			if n.Deleted {
				r.registry[n.ID].Deleted = true
			}
			continue
		}
		r.processNode(n)
	}
}

// processNode handles the causal dependency logic during a merge.
// If a node's parent is missing, the node is moved to the pendingOrphans buffer.
func (r *RGA) processNode(n Node) {
	if _, parentExists := r.registry[n.ParentID]; parentExists {
		newNode := &Node{
			ID:       n.ID,
			ParentID: n.ParentID,
			Value:    n.Value,
			Deleted:  n.Deleted,
		}
		r.integrate(newNode)

		if orphans, ok := r.pendingOrphans[n.ID]; ok {
			for _, child := range orphans {
				r.processNode(child)
			}
			delete(r.pendingOrphans, n.ID)
		}
	} else {
		r.pendingOrphans[n.ParentID] = append(r.pendingOrphans[n.ParentID], n)
	}
}

// integrate executes the deterministic pointer-linking math.
// It ensures that siblings (nodes sharing the same parent) are
// ordered by their IDs, guaranteeing that all replicas converge
// to the same linear sequence.
func (r *RGA) integrate(newNode *Node) {
	parent := r.registry[newNode.ParentID]

	prev := parent
	current := parent.Next
	for current != nil && current.ParentID == newNode.ParentID {
		if newNode.ID.Greater(current.ID) {
			break
		}
		prev = current
		current = current.Next
	}

	newNode.Next = current
	prev.Next = newNode
	r.registry[newNode.ID] = newNode

	if newNode.ID.Timestamp > r.clock {
		r.clock = newNode.ID.Timestamp
	}
}

// Value returns the linearized, visible text of the sequence.
// It traverses the internal linked-list and filters out nodes
// marked as deleted (tombstones). This satisfies the CRDT interface.
func (r *RGA) Value() any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var chars []rune
	curr := r.root.Next
	for curr != nil {
		if !curr.Deleted {
			chars = append(chars, curr.Value)
		}
		curr = curr.Next
	}
	return string(chars)
}
