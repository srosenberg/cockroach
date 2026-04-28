// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

// Package streamingvalidator implements a small set of MVCC-history
// invariants that can be checked online (no end-of-run batch pass) against
// observations emitted by a kvnemesis-style workload.
//
// # Motivation
//
// The original kvnemesis validator (pkg/kv/kvnemesis) ingests the entire
// MVCC history via RangeFeed and performs an Elle-style serializability
// check at the end of the run. That is well-suited to short, in-process
// tests but does not fit a long-running workload that needs to assert
// invariants continuously and surface violations promptly.
//
// This package is intentionally narrow. It checks only:
//
//   - I1 (read-your-writes): within a session, a read of K must observe
//     a value at least as new as the most recent write to K by the same
//     session.
//   - I2 (monotonic reads): within a session, successive reads of K must
//     observe non-regressing positions in K's per-key history.
//   - I3 (no phantom values): every Seq returned by a read must
//     correspond to a write the validator has been told about.
//   - I5 (no torn writes): the bytes returned for a Seq must equal the
//     bytes the corresponding writer claimed to store.
//
// Invariants are deliberately per-key and per-session — they do not
// require any global serializability check, and so cost O(1) work per
// observation. See `kvnemesis_drt_workload_plan.md` for the broader
// design and the invariants that are intentionally deferred (I4
// rolling-window survival, I6 SI repeatable reads, I7 closed-ts safety).
//
// # Threading and ordering
//
// The Validator is safe for concurrent use. Within any one session,
// observations must be submitted in causal order — i.e. the worker
// goroutine for a session is expected to be single-threaded. Across
// sessions, no ordering is required, but observations that reference a
// Seq written by another session must arrive after the corresponding
// write observation, otherwise I3 will fire spuriously. The MVP
// workload sidesteps this by giving each session a disjoint key shard.
package streamingvalidator
