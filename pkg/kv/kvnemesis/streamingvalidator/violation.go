// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package streamingvalidator

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv/kvnemesis/kvnemesisutil"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

// ViolationKind enumerates the invariants this validator checks.
type ViolationKind int

const (
	// ReadYourWrite (I1): a session read a key without observing its
	// own most recent committed write to that key.
	ReadYourWrite ViolationKind = iota + 1
	// MonotonicReads (I2): a session's successive reads of a key
	// regressed in the per-key history order.
	MonotonicReads
	// PhantomValue (I3): a read returned a Seq the validator has no
	// record of (no committed or ambiguous write attempt).
	PhantomValue
	// TornWrite (I5): the bytes returned for a Seq did not match the
	// bytes the writer claimed to have stored under that Seq.
	TornWrite
)

// String returns a short identifier for log messages.
func (k ViolationKind) String() string {
	switch k {
	case ReadYourWrite:
		return "I1_read_your_write"
	case MonotonicReads:
		return "I2_monotonic_reads"
	case PhantomValue:
		return "I3_phantom_value"
	case TornWrite:
		return "I5_torn_write"
	default:
		return fmt.Sprintf("unknown(%d)", int(k))
	}
}

// Violation is a single invariant breach the validator surfaced. The
// fields are intentionally flat — Violation is meant to be JSONL'd
// straight to a journal and grepped during triage.
type Violation struct {
	// Kind is the invariant that was breached.
	Kind ViolationKind
	// SessionID identifies the session whose observation surfaced the
	// breach. For PhantomValue and TornWrite this is the reader's
	// session (which may differ from the writer's).
	SessionID string
	// Key is the key whose history was inconsistent.
	Key roachpb.Key
	// ObservedSeq is the Seq the reader saw, when relevant.
	ObservedSeq kvnemesisutil.Seq
	// ExpectedSeq is the Seq the validator believed should be visible
	// (the session's own most recent write for I1; the previous read's
	// Seq for I2). Zero if not applicable.
	ExpectedSeq kvnemesisutil.Seq
	// Detail is a human-readable description that includes whatever
	// context did not fit in the structured fields.
	Detail string
}

func (v Violation) String() string {
	return fmt.Sprintf("[%s] session=%q key=%s observed=%s expected=%s: %s",
		v.Kind, v.SessionID, v.Key, v.ObservedSeq, v.ExpectedSeq, v.Detail)
}
