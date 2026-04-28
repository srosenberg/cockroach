// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package streamingvalidator

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/kv/kvnemesis/kvnemesisutil"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/stretchr/testify/require"
)

const (
	sessA = "session-a"
	sessB = "session-b"
)

func k(s string) roachpb.Key { return roachpb.Key(s) }

func put(session string, key string, seq kvnemesisutil.Seq, value string) Observation {
	return Observation{
		SessionID: session,
		Kind:      OpPut,
		Key:       k(key),
		Seq:       seq,
		Value:     []byte(value),
		Outcome:   OutcomeCommitted,
	}
}

func del(session string, key string, seq kvnemesisutil.Seq) Observation {
	return Observation{
		SessionID: session,
		Kind:      OpDelete,
		Key:       k(key),
		Seq:       seq,
		Outcome:   OutcomeCommitted,
	}
}

func get(session string, key string, observed ReadKV) Observation {
	return Observation{
		SessionID: session,
		Kind:      OpGet,
		Key:       k(key),
		Reads:     []ReadKV{observed},
		Outcome:   OutcomeCommitted,
	}
}

// TestCleanRun submits a varied stream of valid ops and verifies no
// violations are produced. This is the negative control for the rest
// of the suite: if it fails, the validator is over-reporting.
func TestCleanRun(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	// Single-session: write, read-your-write, overwrite, read again.
	require.Empty(t, v.OnOpResult(put(sessA, "k1", 1, "v1")))
	require.Empty(t, v.OnOpResult(get(sessA, "k1", ReadKV{Key: k("k1"), Value: []byte("v1"), Seq: 1})))
	require.Empty(t, v.OnOpResult(put(sessA, "k1", 2, "v2")))
	require.Empty(t, v.OnOpResult(get(sessA, "k1", ReadKV{Key: k("k1"), Value: []byte("v2"), Seq: 2})))

	// Delete then read absent.
	require.Empty(t, v.OnOpResult(del(sessA, "k1", 3)))
	require.Empty(t, v.OnOpResult(get(sessA, "k1", ReadKV{Key: k("k1")})))

	// Different session, different shard, same flow.
	require.Empty(t, v.OnOpResult(put(sessB, "k2", 4, "x")))
	require.Empty(t, v.OnOpResult(get(sessB, "k2", ReadKV{Key: k("k2"), Value: []byte("x"), Seq: 4})))

	stats := v.Stats()
	require.Equal(t, 2, stats.Sessions)
	require.Equal(t, 2, stats.Keys)
	require.Equal(t, 4, stats.Writes)
}

// TestI1ReadYourWrite_Violation: session writes Seq=2, then a
// subsequent read in the same session returns Seq=1 (the older
// version). I1 must fire.
func TestI1ReadYourWrite_Violation(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	require.Empty(t, v.OnOpResult(put(sessA, "k1", 1, "v1")))
	require.Empty(t, v.OnOpResult(put(sessA, "k1", 2, "v2")))

	violations := v.OnOpResult(get(sessA, "k1",
		ReadKV{Key: k("k1"), Value: []byte("v1"), Seq: 1}))

	require.Len(t, violations, 1)
	require.Equal(t, ReadYourWrite, violations[0].Kind)
	require.Equal(t, kvnemesisutil.Seq(1), violations[0].ObservedSeq)
	require.Equal(t, kvnemesisutil.Seq(2), violations[0].ExpectedSeq)
}

// TestI1ReadYourWrite_AbsentAfterWrite: a session that wrote Seq=1
// then sees the key absent must fire I1 (the write is missing).
func TestI1ReadYourWrite_AbsentAfterWrite(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	require.Empty(t, v.OnOpResult(put(sessA, "k1", 1, "v1")))

	violations := v.OnOpResult(get(sessA, "k1", ReadKV{Key: k("k1")}))

	require.Len(t, violations, 1)
	require.Equal(t, ReadYourWrite, violations[0].Kind)
}

// TestI1ReadYourWrite_DifferentSession: a write in session A is
// invisible to session B's reads (no I1 violation for B).
func TestI1ReadYourWrite_DifferentSession(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	require.Empty(t, v.OnOpResult(put(sessA, "k1", 1, "v1")))
	// Session B has no expectation; an absent read is fine.
	require.Empty(t, v.OnOpResult(get(sessB, "k1", ReadKV{Key: k("k1")})))
}

// TestI2MonotonicReads_Violation: session reads Seq=2 then later sees
// Seq=1 (older). I2 must fire.
func TestI2MonotonicReads_Violation(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	// Use session B as the writer so the reads in session A do not also
	// trip I1.
	require.Empty(t, v.OnOpResult(put(sessB, "k1", 1, "v1")))
	require.Empty(t, v.OnOpResult(put(sessB, "k1", 2, "v2")))

	require.Empty(t, v.OnOpResult(get(sessA, "k1",
		ReadKV{Key: k("k1"), Value: []byte("v2"), Seq: 2})))
	violations := v.OnOpResult(get(sessA, "k1",
		ReadKV{Key: k("k1"), Value: []byte("v1"), Seq: 1}))

	require.Len(t, violations, 1)
	require.Equal(t, MonotonicReads, violations[0].Kind)
	require.Equal(t, kvnemesisutil.Seq(1), violations[0].ObservedSeq)
	require.Equal(t, kvnemesisutil.Seq(2), violations[0].ExpectedSeq)
}

// TestI3PhantomValue: a read returns a Seq the validator never saw.
// I3 must fire.
func TestI3PhantomValue(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	violations := v.OnOpResult(get(sessA, "k1",
		ReadKV{Key: k("k1"), Value: []byte("ghost"), Seq: 999}))

	require.Len(t, violations, 1)
	require.Equal(t, PhantomValue, violations[0].Kind)
	require.Equal(t, kvnemesisutil.Seq(999), violations[0].ObservedSeq)
}

// TestI3PhantomValue_AmbiguousIsNotPhantom: a write that returned an
// ambiguous result is still recorded by the validator, so a subsequent
// read seeing its Seq must NOT fire I3.
func TestI3PhantomValue_AmbiguousIsNotPhantom(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	ambiguous := put(sessA, "k1", 1, "v1")
	ambiguous.Outcome = OutcomeAmbiguous
	require.Empty(t, v.OnOpResult(ambiguous))

	require.Empty(t, v.OnOpResult(get(sessA, "k1",
		ReadKV{Key: k("k1"), Value: []byte("v1"), Seq: 1})))
}

// TestI3PhantomValue_FailedIsPhantom: a write reported as Failed is
// not recorded; a subsequent read seeing its Seq must fire I3 (this
// is a real bug — the engine surfaced a Seq that the writer was told
// did not commit).
func TestI3PhantomValue_FailedIsPhantom(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	failed := put(sessA, "k1", 1, "v1")
	failed.Outcome = OutcomeFailed
	require.Empty(t, v.OnOpResult(failed))

	violations := v.OnOpResult(get(sessB, "k1",
		ReadKV{Key: k("k1"), Value: []byte("v1"), Seq: 1}))

	require.Len(t, violations, 1)
	require.Equal(t, PhantomValue, violations[0].Kind)
}

// TestI5TornWrite: a write of "v1" is followed by a read that returns
// the right Seq but the wrong bytes. I5 must fire.
func TestI5TornWrite(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	require.Empty(t, v.OnOpResult(put(sessA, "k1", 1, "v1")))

	violations := v.OnOpResult(get(sessB, "k1",
		ReadKV{Key: k("k1"), Value: []byte("CORRUPT"), Seq: 1}))

	require.Len(t, violations, 1)
	require.Equal(t, TornWrite, violations[0].Kind)
	require.Equal(t, kvnemesisutil.Seq(1), violations[0].ObservedSeq)
}

// TestI5TornWrite_DeleteResurrected: a Delete with Seq=1 is later
// reported by a read as carrying that Seq but with a non-nil value.
// I5 must fire.
func TestI5TornWrite_DeleteResurrected(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	require.Empty(t, v.OnOpResult(del(sessA, "k1", 1)))

	violations := v.OnOpResult(get(sessB, "k1",
		ReadKV{Key: k("k1"), Value: []byte("ghost"), Seq: 1}))

	require.Len(t, violations, 1)
	require.Equal(t, TornWrite, violations[0].Kind)
}

// TestScanReports multiple ReadKVs in one observation; a torn entry
// among clean entries must fire only one violation.
func TestScanReportsPerKVViolations(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{})

	require.Empty(t, v.OnOpResult(put(sessA, "k1", 1, "v1")))
	require.Empty(t, v.OnOpResult(put(sessA, "k2", 2, "v2")))

	scan := Observation{
		SessionID: sessB,
		Kind:      OpScan,
		Key:       k("k1"),
		EndKey:    k("k3"),
		Reads: []ReadKV{
			{Key: k("k1"), Value: []byte("v1"), Seq: 1},
			{Key: k("k2"), Value: []byte("CORRUPT"), Seq: 2},
		},
		Outcome: OutcomeCommitted,
	}
	violations := v.OnOpResult(scan)

	require.Len(t, violations, 1)
	require.Equal(t, TornWrite, violations[0].Kind)
	require.Equal(t, k("k2"), violations[0].Key)
}

// TestRetentionTrim: with a tight per-key cap, the oldest version is
// dropped from history. The writer is sessA; reads come from sessB so
// I1 cannot also fire and conflate the assertion. After trimming:
//
//   - bySeq still holds Seq=1, so I3 does not fire on a stale read of
//     Seq=1 (and I5 still validates the bytes if anyone reads it).
//   - I2 between Seq=2 and Seq=3 still works because both remain in
//     history.
func TestRetentionTrim(t *testing.T) {
	defer leaktest.AfterTest(t)()

	v := New(Options{MaxHistoryPerKey: 2})

	require.Empty(t, v.OnOpResult(put(sessA, "k1", 1, "v1")))
	require.Empty(t, v.OnOpResult(put(sessA, "k1", 2, "v2")))
	require.Empty(t, v.OnOpResult(put(sessA, "k1", 3, "v3")))

	stats := v.Stats()
	require.Equal(t, 3, stats.Writes)

	// I1 against a trimmed expected falls back to "cannot prove": sessB
	// reading the trimmed Seq=1 is not flagged.
	require.Empty(t, v.OnOpResult(get(sessB, "k1",
		ReadKV{Key: k("k1"), Value: []byte("v1"), Seq: 1})))

	// I2 between Seq=3 and Seq=2 (both retained) still fires.
	require.Empty(t, v.OnOpResult(get(sessB, "k1",
		ReadKV{Key: k("k1"), Value: []byte("v3"), Seq: 3})))
	violations := v.OnOpResult(get(sessB, "k1",
		ReadKV{Key: k("k1"), Value: []byte("v2"), Seq: 2}))
	require.Len(t, violations, 1)
	require.Equal(t, MonotonicReads, violations[0].Kind)
}
