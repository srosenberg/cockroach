// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package streamingvalidator

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvnemesis/kvnemesisutil"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

// OpKind identifies the primitive op that produced an Observation.
type OpKind int

const (
	// OpGet is a single-key point read.
	OpGet OpKind = iota + 1
	// OpPut is a single-key write of a value.
	OpPut
	// OpDelete is a single-key tombstone write. Tombstones still carry
	// a Seq, encoded into the MVCC value the same way a Put's Seq is.
	OpDelete
	// OpScan is a range read returning zero or more KVs.
	OpScan
)

// Outcome classifies what the worker observed when its op completed.
//
// OutcomeAmbiguous deserves special handling: an ambiguous result means
// the write may or may not have committed, so the validator must be
// prepared for that Seq to surface in subsequent reads — but its absence
// is also valid. We track ambiguous writes so I3 (no phantom values)
// does not fire on a successful-but-reported-ambiguous commit.
type Outcome int

const (
	// OutcomeCommitted means the op succeeded and (for writes) is durable.
	OutcomeCommitted Outcome = iota + 1
	// OutcomeFailed means the op cleanly failed and (for writes) did not
	// commit.
	OutcomeFailed
	// OutcomeAmbiguous means the op may or may not have committed.
	OutcomeAmbiguous
)

// ReadKV is one key-value pair observed by a read op. For deletes the
// value is nil and Seq carries the tombstone's Seq if the engine
// surfaced it; for missing keys, Seq is zero.
type ReadKV struct {
	// Key is the read key.
	Key roachpb.Key
	// Value is the bytes the read observed, or nil if the key is absent
	// or covered by a tombstone.
	Value []byte
	// Seq is the kvnemesisutil.Seq decoded from the MVCC value's
	// KVNemesisSeq tag. Zero means "no Seq" (e.g. a missing key, or a
	// value not written by this workload).
	Seq kvnemesisutil.Seq
}

// Observation is one atomic unit's worth of result that the workload
// hands to the Validator. For the MVP, batches and txns are flattened
// into one Observation per primitive op, in causal order within a
// session.
type Observation struct {
	// SessionID identifies the causal stream this observation belongs
	// to. Within a session, observations must be submitted in the order
	// they happened-before. Across sessions, no ordering is required.
	SessionID string

	// Kind is the op type.
	Kind OpKind

	// Key is the operand key (for Get/Put/Delete) or the start key (for
	// Scan).
	Key roachpb.Key
	// EndKey is the exclusive end of the scanned span; only set for
	// Scan.
	EndKey roachpb.Key

	// Seq is set for write ops (Put, Delete) and identifies the
	// kvnemesisutil.Seq embedded in the MVCC value.
	Seq kvnemesisutil.Seq
	// Value is set for Put ops; it is the bytes the writer claimed to
	// store. Used by I5 (torn writes) to compare against subsequent
	// reads.
	Value []byte

	// Reads is set for read ops (Get, Scan) and lists the KVs the read
	// observed. A Get returns at most one entry; a Scan returns zero or
	// more.
	Reads []ReadKV

	// Outcome classifies the result; see Outcome's docs.
	Outcome Outcome
	// SubmittedAt is the wall-clock time at which the workload finished
	// the op. Used for retention/trimming.
	SubmittedAt time.Time
}

// IsWrite reports whether op is a write op (Put or Delete).
func (k OpKind) IsWrite() bool {
	return k == OpPut || k == OpDelete
}

// IsRead reports whether op is a read op (Get or Scan).
func (k OpKind) IsRead() bool {
	return k == OpGet || k == OpScan
}
