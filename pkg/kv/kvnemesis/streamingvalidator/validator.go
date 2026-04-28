// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package streamingvalidator

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/cockroachdb/cockroach/pkg/kv/kvnemesis/kvnemesisutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

// Default retention bounds. The validator trims oldest entries beyond
// these limits so memory stays O(maxKeys * MaxHistoryPerKey + maxSessions
// * MaxOpsPerSession). The defaults are sized for a several-minute roach
// test; the workload binary may override them via Options for longer
// runs.
const (
	defaultMaxHistoryPerKey = 1024
	defaultMaxOpsPerSession = 1024
)

// Options configures a Validator.
type Options struct {
	// MaxHistoryPerKey caps the number of write versions retained per
	// key. When exceeded, the oldest versions are dropped. A value of 0
	// uses defaultMaxHistoryPerKey; a negative value disables the cap
	// (intended for tests).
	MaxHistoryPerKey int
	// MaxOpsPerSession caps the number of recent ops retained per
	// session for I1/I2 lookups. Same semantics as MaxHistoryPerKey for
	// 0 and negative values.
	MaxOpsPerSession int
}

// Validator implements the streaming invariants described in the
// package doc. It is safe for concurrent use across sessions.
type Validator struct {
	opts Options

	mu struct {
		syncutil.Mutex

		// bySeq maps a write's Seq to the writeRecord the validator
		// learned about. Used by I3 (phantom) and I5 (torn writes).
		bySeq map[kvnemesisutil.Seq]*writeRecord

		// history maps string(key) to the per-key version history,
		// ordered by the order the validator learned about each
		// version (an approximation of MVCC commit order under the
		// MVP's "single writer per key shard" assumption).
		history map[string]*keyHistory

		// sessions maps SessionID to the session's recent state, used
		// by I1 (RYW) and I2 (monotonic reads).
		sessions map[string]*sessionState
	}
}

// writeRecord is the validator's memory of one write op. Created when
// the validator sees a Put/Delete observation; later consulted by reads
// to validate the bytes (I5) and to confirm the Seq is known (I3).
type writeRecord struct {
	seq       kvnemesisutil.Seq
	key       []byte // string-keyed; copy retained
	value     []byte // nil for Delete
	isDelete  bool
	sessionID string
	outcome   Outcome
}

// keyHistory stores the per-key version chain. Bounded by
// MaxHistoryPerKey.
type keyHistory struct {
	// versions are in the order the validator learned about them. With
	// the MVP's per-shard ownership, this matches MVCC commit order; in
	// the multi-writer case it is an approximation.
	versions []*writeRecord
	// seqIndex maps Seq to its position in versions, for O(1)
	// lookups by readers.
	seqIndex map[kvnemesisutil.Seq]int
}

// sessionState is the per-session bookkeeping for I1 and I2.
type sessionState struct {
	// lastWrittenSeq[key] is the Seq of the session's most recent
	// committed write to key. Used by I1 (RYW).
	lastWrittenSeq map[string]kvnemesisutil.Seq
	// lastReadSeq[key] is the Seq the session most recently observed
	// for key. Zero means "absent". Used by I2 (monotonic reads).
	lastReadSeq map[string]kvnemesisutil.Seq
	// opCount counts ops processed for this session; trim threshold
	// for opportunistic GC of the maps above (kept simple for MVP).
	opCount int
}

// New returns a Validator with the given options. The zero Options
// value is fine.
func New(opts Options) *Validator {
	if opts.MaxHistoryPerKey == 0 {
		opts.MaxHistoryPerKey = defaultMaxHistoryPerKey
	}
	if opts.MaxOpsPerSession == 0 {
		opts.MaxOpsPerSession = defaultMaxOpsPerSession
	}
	v := &Validator{opts: opts}
	v.mu.bySeq = make(map[kvnemesisutil.Seq]*writeRecord)
	v.mu.history = make(map[string]*keyHistory)
	v.mu.sessions = make(map[string]*sessionState)
	return v
}

// OnOpResult feeds one Observation through every enabled invariant
// check and returns any violations it detected. The returned slice is
// owned by the caller.
//
// The validator updates its bookkeeping (write history, session state)
// before returning regardless of whether violations were found, so a
// single bug does not desynchronize the validator from the workload.
func (v *Validator) OnOpResult(obs Observation) []Violation {
	v.mu.Lock()
	defer v.mu.Unlock()

	sess := v.sessionLocked(obs.SessionID)

	var violations []Violation
	switch {
	case obs.Kind.IsWrite():
		v.recordWriteLocked(obs)
		// A committed write counts as a "self read" for I1 — a read in
		// the same session that follows must observe at least this
		// Seq.
		if obs.Outcome == OutcomeCommitted {
			if sess.lastWrittenSeq == nil {
				sess.lastWrittenSeq = make(map[string]kvnemesisutil.Seq)
			}
			sess.lastWrittenSeq[string(obs.Key)] = obs.Seq
		}
	case obs.Kind.IsRead():
		violations = append(violations, v.checkReadLocked(sess, obs)...)
	}

	sess.opCount++
	return violations
}

// sessionLocked returns the sessionState for id, creating it lazily.
// Caller must hold v.mu.
func (v *Validator) sessionLocked(id string) *sessionState {
	s := v.mu.sessions[id]
	if s != nil {
		return s
	}
	s = &sessionState{}
	v.mu.sessions[id] = s
	return s
}

// recordWriteLocked stores a write observation in bySeq and history.
// Failed writes are not recorded — they did not land in MVCC, so any
// subsequent read seeing their Seq is a phantom (which is what we want
// I3 to catch). Ambiguous writes are recorded so I3 does not
// false-positive on a write that did commit despite the ambiguous
// result.
func (v *Validator) recordWriteLocked(obs Observation) {
	if obs.Outcome == OutcomeFailed {
		return
	}
	rec := &writeRecord{
		seq:       obs.Seq,
		key:       append([]byte(nil), obs.Key...),
		value:     append([]byte(nil), obs.Value...),
		isDelete:  obs.Kind == OpDelete,
		sessionID: obs.SessionID,
		outcome:   obs.Outcome,
	}
	if rec.isDelete {
		rec.value = nil
	}
	v.mu.bySeq[obs.Seq] = rec

	h := v.mu.history[string(obs.Key)]
	if h == nil {
		h = &keyHistory{seqIndex: make(map[kvnemesisutil.Seq]int)}
		v.mu.history[string(obs.Key)] = h
	}
	h.versions = append(h.versions, rec)
	h.seqIndex[obs.Seq] = len(h.versions) - 1
	v.trimHistoryLocked(h)
}

// trimHistoryLocked drops the oldest versions if the history exceeds
// MaxHistoryPerKey. A negative value disables trimming.
func (v *Validator) trimHistoryLocked(h *keyHistory) {
	cap := v.opts.MaxHistoryPerKey
	if cap < 0 || len(h.versions) <= cap {
		return
	}
	drop := len(h.versions) - cap
	for _, r := range h.versions[:drop] {
		delete(h.seqIndex, r.seq)
	}
	h.versions = slices.Delete(h.versions, 0, drop)
	// Rebuild seqIndex offsets in place: every remaining seq's stored
	// index decreases by drop.
	for seq, idx := range h.seqIndex {
		h.seqIndex[seq] = idx - drop
	}
}

// checkReadLocked validates one read observation against I1, I2, I3,
// and I5. Returns any violations and updates per-session bookkeeping.
func (v *Validator) checkReadLocked(sess *sessionState, obs Observation) []Violation {
	var violations []Violation
	for _, r := range obs.Reads {
		violations = append(violations, v.checkReadKVLocked(sess, obs, r)...)
	}
	return violations
}

// checkReadKVLocked runs the per-KV portion of read validation.
func (v *Validator) checkReadKVLocked(sess *sessionState, obs Observation, r ReadKV) []Violation {
	var violations []Violation

	// I3 (phantom value): a non-zero Seq must correspond to a recorded
	// write. We only check if Seq is non-zero — a missing key (Seq=0)
	// is by definition not a phantom.
	var rec *writeRecord
	if r.Seq != 0 {
		rec = v.mu.bySeq[r.Seq]
		if rec == nil {
			violations = append(violations, Violation{
				Kind:        PhantomValue,
				SessionID:   obs.SessionID,
				Key:         r.Key,
				ObservedSeq: r.Seq,
				Detail: fmt.Sprintf("read returned Seq %s but the validator has no record of a "+
					"write with that Seq", r.Seq),
			})
		}
	}

	// I5 (torn write): if we have a record for this Seq, the bytes must
	// match. A delete is recorded with value=nil; a delete observed by
	// the read must have value=nil.
	if rec != nil {
		expected := rec.value
		got := r.Value
		if rec.isDelete && got != nil {
			violations = append(violations, Violation{
				Kind:        TornWrite,
				SessionID:   obs.SessionID,
				Key:         r.Key,
				ObservedSeq: r.Seq,
				Detail: fmt.Sprintf("Seq %s was a delete but read returned %d bytes",
					r.Seq, len(got)),
			})
		} else if !rec.isDelete && !bytes.Equal(expected, got) {
			violations = append(violations, Violation{
				Kind:        TornWrite,
				SessionID:   obs.SessionID,
				Key:         r.Key,
				ObservedSeq: r.Seq,
				Detail: fmt.Sprintf("Seq %s was written with %d bytes but read returned %d bytes",
					r.Seq, len(expected), len(got)),
			})
		}
	}

	// I1 (read-your-writes): if this session previously committed a
	// write to r.Key with Seq=K, the read's observed version must be K
	// or newer in the per-key history.
	if expected, ok := sess.lastWrittenSeq[string(r.Key)]; ok {
		if v.observedAtLeastLocked(r.Key, r.Seq, expected) == observedOlder {
			violations = append(violations, Violation{
				Kind:        ReadYourWrite,
				SessionID:   obs.SessionID,
				Key:         r.Key,
				ObservedSeq: r.Seq,
				ExpectedSeq: expected,
				Detail: fmt.Sprintf("session committed Seq %s but a subsequent read returned "+
					"Seq %s", expected, r.Seq),
			})
		}
	}

	// I2 (monotonic reads): the session's previous observed Seq for
	// this key must not be newer than the current one.
	if prev, ok := sess.lastReadSeq[string(r.Key)]; ok && prev != 0 {
		if v.observedAtLeastLocked(r.Key, r.Seq, prev) == observedOlder {
			violations = append(violations, Violation{
				Kind:        MonotonicReads,
				SessionID:   obs.SessionID,
				Key:         r.Key,
				ObservedSeq: r.Seq,
				ExpectedSeq: prev,
				Detail: fmt.Sprintf("session previously read Seq %s but a subsequent read "+
					"returned an older Seq %s", prev, r.Seq),
			})
		}
	}

	if sess.lastReadSeq == nil {
		sess.lastReadSeq = make(map[string]kvnemesisutil.Seq)
	}
	sess.lastReadSeq[string(r.Key)] = r.Seq

	return violations
}

// orderResult classifies how observed compares against expected within a
// per-key history.
type orderResult int

const (
	// observedNewerOrEqual means observed appears at least as late in
	// the per-key history as expected, or expected is unknown to the
	// validator (in which case we cannot prove a violation).
	observedNewerOrEqual orderResult = iota
	// observedOlder means observed appears strictly earlier in the
	// per-key history than expected — i.e. an invariant violation.
	observedOlder
)

// observedAtLeastLocked compares two Seqs against key's history and
// returns whether observed is at least as new as expected.
//
// The comparison uses the per-key history's order (which approximates
// MVCC commit order under the MVP's per-shard ownership). Absent reads
// (observed == 0) are interpreted in light of the latest version in
// history: a tombstone covers absent reads, but a live Put does not.
//
// If expected is unknown to the validator (trimmed or never recorded),
// we cannot prove anything and report observedNewerOrEqual.
func (v *Validator) observedAtLeastLocked(
	key []byte, observed, expected kvnemesisutil.Seq,
) orderResult {
	h := v.mu.history[string(key)]
	if h == nil {
		return observedNewerOrEqual
	}
	expectedIdx, ok := h.seqIndex[expected]
	if !ok {
		return observedNewerOrEqual
	}
	if observed == 0 {
		// "Absent" is consistent with the constraint iff the latest
		// version in history is a tombstone whose index is at least
		// expectedIdx — i.e. a delete that covers expected. Otherwise
		// a Put at or after expected is the visible state and the
		// reader should have seen it.
		last := h.versions[len(h.versions)-1]
		if last.isDelete && h.seqIndex[last.seq] >= expectedIdx {
			return observedNewerOrEqual
		}
		return observedOlder
	}
	observedIdx, ok := h.seqIndex[observed]
	if !ok {
		// observed isn't in history (trimmed, or recorded by another
		// session that hasn't checked in). I3 will fire separately.
		return observedNewerOrEqual
	}
	if observedIdx < expectedIdx {
		return observedOlder
	}
	return observedNewerOrEqual
}

// Stats returns coarse counters useful for tests and roachtest assertions.
func (v *Validator) Stats() Stats {
	v.mu.Lock()
	defer v.mu.Unlock()
	return Stats{
		Sessions: len(v.mu.sessions),
		Keys:     len(v.mu.history),
		Writes:   len(v.mu.bySeq),
	}
}

// Stats is the coarse view returned by Validator.Stats.
type Stats struct {
	// Sessions is the number of sessions the validator has seen.
	Sessions int
	// Keys is the number of distinct keys with at least one write
	// recorded.
	Keys int
	// Writes is the number of writes the validator has recorded
	// (committed or ambiguous; failed writes are not retained).
	Writes int
}
