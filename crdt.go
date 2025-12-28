// Package gocrdt provides a suite of Conflict-free Replicated Data Types (CRDTs).
//
// CRDTs are distributed data structures that guarantee convergence: if multiple
// replicas receive the same set of updates, they will eventually reach the
// same state regardless of the order in which updates were processed.
//
// This package implements State-based CRDTs (CvRDTs) including Counters (G, PN)
// and Sequences (RGA).
package gocrdt

// CRDT is the base interface that defines the behavior for all convergent
// data types in this package.
//
// Implementing types must ensure that their internal state can be merged
// commutatively, associatively, and idempotently to satisfy the mathematical
// properties of a Join-Semilattice.
type CRDT interface {
	// Value returns the current consolidated state of the CRDT.
	//
	// For counters, this typically returns a numeric type (int).
	// For sequences like RGA, this returns the linearized view of the
	// data (usually a string or a slice).
	//
	// Note: Because this returns 'any' (interface{}), callers may need
	// to perform a type assertion to use the underlying data.
	Value() any

	// Merge combines the state of a remote CRDT into the local instance.
	//
	// To guarantee convergence across all distributed replicas, the
	// implementation of Merge MUST be:
	//
	// 1. Commutative: The order of merging doesn't matter.
	//    A.Merge(B) results in the same state as B.Merge(A).
	//
	// 2. Associative: The grouping of merges doesn't matter.
	//    (A.Merge(B)).Merge(C) == A.Merge((B.Merge(C))).
	//
	// 3. Idempotent: Merging the same state multiple times has no effect
	//    beyond the first merge. A.Merge(A) == A.
	//
	// Implementations should perform type-assertion on the 'other' parameter
	// and return an error if the types are incompatible (e.g., attempting
	// to merge a GCounter into an RGA).
	Merge(other CRDT) error
}
