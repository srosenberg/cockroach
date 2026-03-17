// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/roachtestutil"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/roachtestutil/task"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/failureinjection/failures"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/lib/pq"
)

const (
	jepsenBankNumAccounts    = 5
	jepsenBankInitialBalance = 10
	jepsenBankNumWorkers     = 30
	jepsenBankDuration       = 5 * time.Minute
	jepsenBankNumNodes       = 5
)

// jepsenBankOpType is the type of operation (read or transfer).
type jepsenBankOpType string

const (
	jepsenBankOpRead     jepsenBankOpType = "read"
	jepsenBankOpTransfer jepsenBankOpType = "transfer"
)

// jepsenBankOutcome is the Jepsen-style outcome of an operation.
type jepsenBankOutcome string

const (
	jepsenBankOutcomeInvoke jepsenBankOutcome = "invoke"
	jepsenBankOutcomeOk     jepsenBankOutcome = "ok"
	jepsenBankOutcomeFail   jepsenBankOutcome = "fail"
	jepsenBankOutcomeInfo   jepsenBankOutcome = "info"
)

// jepsenBankHistoryEntry is one entry in the Jepsen-style operation history.
type jepsenBankHistoryEntry struct {
	Index   int64                  `json:"index"`
	Type    jepsenBankOutcome      `json:"type"`
	Op      jepsenBankOpType       `json:"op"`
	Process int                    `json:"process"`
	TimeNs  int64                  `json:"time_ns"`
	Value   map[string]interface{} `json:"value"`
	Error   string                 `json:"error,omitempty"`
}

// jepsenBankHistory accumulates history entries in a thread-safe manner.
type jepsenBankHistory struct {
	mu      sync.Mutex
	entries []jepsenBankHistoryEntry
	nextIdx atomic.Int64
}

func (h *jepsenBankHistory) append(entry jepsenBankHistoryEntry) {
	entry.Index = h.nextIdx.Add(1) - 1
	entry.TimeNs = time.Now().UnixNano()
	h.mu.Lock()
	h.entries = append(h.entries, entry)
	h.mu.Unlock()
}

func (h *jepsenBankHistory) writeJSON(path string) error {
	h.mu.Lock()
	entries := make([]jepsenBankHistoryEntry, len(h.entries))
	copy(entries, h.entries)
	h.mu.Unlock()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}

// jepsenBankNemesisType identifies which nemesis to use.
type jepsenBankNemesisType int

const (
	jepsenBankNemesisStartKill2 jepsenBankNemesisType = iota
	jepsenBankNemesisMajorityRing
)

func (n jepsenBankNemesisType) String() string {
	switch n {
	case jepsenBankNemesisStartKill2:
		return "start-kill-2"
	case jepsenBankNemesisMajorityRing:
		return "majority-ring"
	default:
		return "unknown"
	}
}

// classifyError maps a CockroachDB error to a Jepsen outcome.
// "fail" means definitely did NOT commit; "info" means ambiguous.
func classifyError(err error) jepsenBankOutcome {
	if err == nil {
		return jepsenBankOutcomeOk
	}
	pqErr, ok := err.(*pq.Error)
	if ok {
		switch pqErr.Code {
		case "40001": // serialization_failure
			return jepsenBankOutcomeFail
		case "40003": // statement_completion_unknown
			return jepsenBankOutcomeInfo
		case "23505", "23503", "23514", "23502": // constraint violations
			return jepsenBankOutcomeFail
		}
		// Other SQL errors are definite failures.
		return jepsenBankOutcomeFail
	}
	// Connection errors, EOF, context canceled, timeouts are ambiguous.
	return jepsenBankOutcomeInfo
}

func registerJepsenBank(r registry.Registry) {
	nemeses := []jepsenBankNemesisType{
		jepsenBankNemesisStartKill2,
		jepsenBankNemesisMajorityRing,
	}
	for _, nemesis := range nemeses {
		nem := nemesis
		r.Add(registry.TestSpec{
			Name:             fmt.Sprintf("jepsen-bank/%s", nem),
			Owner:            registry.OwnerTestEng,
			Cluster:          r.MakeClusterSpec(jepsenBankNumNodes, spec.ReuseNone()),
			CompatibleClouds: registry.OnlyGCE,
			Suites:           registry.ManualOnly,
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				runJepsenBank(ctx, t, c, nem)
			},
		})
	}
}

func runJepsenBank(
	ctx context.Context, t test.Test, c cluster.Cluster, nemesis jepsenBankNemesisType,
) {
	nodes := c.CRDBNodes()

	// Start the cluster.
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), nodes)

	// Create schema and seed data.
	setupJepsenBankSchema(ctx, t, c)

	history := &jepsenBankHistory{}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Launch workers.
	for i := 0; i < jepsenBankNumWorkers; i++ {
		workerID := i
		t.Go(func(goCtx context.Context, l *logger.Logger) error {
			runJepsenBankWorker(goCtx, l, c, nodes, workerID, history, rng.Int63())
			return nil
		}, task.Name(fmt.Sprintf("bank-worker-%d", workerID)))
	}

	// Launch nemesis.
	cancelNemesis := t.GoWithCancel(func(goCtx context.Context, l *logger.Logger) error {
		return runJepsenBankNemesis(goCtx, l, t, c, nodes, nemesis, rng.Int63())
	}, task.Name("bank-nemesis"))

	// Wait for the workload duration, then stop.
	select {
	case <-time.After(jepsenBankDuration):
	case <-ctx.Done():
	}

	// Cancel nemesis first to allow recovery before workers stop.
	cancelNemesis()

	// Write history to artifacts.
	historyPath := filepath.Join(t.ArtifactsDir(), "jepsen-bank-history.json")
	if err := history.writeJSON(historyPath); err != nil {
		t.L().Printf("failed to write history: %v", err)
	} else {
		t.L().Printf("wrote %d history entries to %s", len(history.entries), historyPath)
	}
}

func setupJepsenBankSchema(ctx context.Context, t test.Test, c cluster.Cluster) {
	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	if _, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS accounts (id INT PRIMARY KEY, balance INT NOT NULL)`,
	); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < jepsenBankNumAccounts; i++ {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO accounts (id, balance) VALUES ($1, $2)`, i, jepsenBankInitialBalance,
		); err != nil {
			t.Fatal(err)
		}
	}
}

// runJepsenBankWorker runs a single bank worker that issues reads and transfers
// in a loop until the context is cancelled. Each worker uses its own DB
// connection and reconnects to a different random node on connection failure.
func runJepsenBankWorker(
	ctx context.Context,
	l *logger.Logger,
	c cluster.Cluster,
	nodes option.NodeListOption,
	workerID int,
	history *jepsenBankHistory,
	seed int64,
) {
	rng := rand.New(rand.NewSource(seed))

	connect := func() *sql.DB {
		node := nodes[rng.Intn(len(nodes))]
		db, err := c.ConnE(ctx, l, node)
		if err != nil {
			l.Printf("worker %d: connect to n%d failed: %v", workerID, node, err)
			return nil
		}
		return db
	}

	db := connect()
	for ctx.Err() == nil {
		if db == nil {
			time.Sleep(500 * time.Millisecond)
			db = connect()
			continue
		}

		if rng.Intn(2) == 0 {
			doJepsenBankRead(ctx, db, workerID, history)
		} else {
			from := rng.Intn(jepsenBankNumAccounts)
			to := rng.Intn(jepsenBankNumAccounts - 1)
			if to >= from {
				to++
			}
			amount := rng.Intn(5) + 1
			doJepsenBankTransfer(ctx, db, workerID, from, to, amount, history)
		}

		// Check if connection is still alive; reconnect if not.
		if err := db.PingContext(ctx); err != nil && ctx.Err() == nil {
			l.Printf("worker %d: ping failed, reconnecting: %v", workerID, err)
			db.Close()
			db = connect()
		}
	}
	if db != nil {
		db.Close()
	}
}

func doJepsenBankRead(ctx context.Context, db *sql.DB, workerID int, history *jepsenBankHistory) {
	history.append(jepsenBankHistoryEntry{
		Type:    jepsenBankOutcomeInvoke,
		Op:      jepsenBankOpRead,
		Process: workerID,
	})

	rows, err := db.QueryContext(ctx, `SELECT id, balance FROM accounts ORDER BY id`)
	if err != nil {
		history.append(jepsenBankHistoryEntry{
			Type:    classifyError(err),
			Op:      jepsenBankOpRead,
			Process: workerID,
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	balances := make(map[string]interface{})
	for rows.Next() {
		var id, balance int
		if err := rows.Scan(&id, &balance); err != nil {
			history.append(jepsenBankHistoryEntry{
				Type:    classifyError(err),
				Op:      jepsenBankOpRead,
				Process: workerID,
				Error:   err.Error(),
			})
			return
		}
		balances[fmt.Sprintf("%d", id)] = balance
	}
	if err := rows.Err(); err != nil {
		history.append(jepsenBankHistoryEntry{
			Type:    classifyError(err),
			Op:      jepsenBankOpRead,
			Process: workerID,
			Error:   err.Error(),
		})
		return
	}

	history.append(jepsenBankHistoryEntry{
		Type:    jepsenBankOutcomeOk,
		Op:      jepsenBankOpRead,
		Process: workerID,
		Value:   balances,
	})
}

func doJepsenBankTransfer(
	ctx context.Context, db *sql.DB, workerID, from, to, amount int, history *jepsenBankHistory,
) {
	value := map[string]interface{}{
		"from":   from,
		"to":     to,
		"amount": amount,
	}
	history.append(jepsenBankHistoryEntry{
		Type:    jepsenBankOutcomeInvoke,
		Op:      jepsenBankOpTransfer,
		Process: workerID,
		Value:   value,
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		history.append(jepsenBankHistoryEntry{
			Type:    classifyError(err),
			Op:      jepsenBankOpTransfer,
			Process: workerID,
			Value:   value,
			Error:   err.Error(),
		})
		return
	}

	// Debit the source account, only if balance is sufficient.
	res, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1`,
		amount, from,
	)
	if err != nil {
		_ = tx.Rollback()
		history.append(jepsenBankHistoryEntry{
			Type:    classifyError(err),
			Op:      jepsenBankOpTransfer,
			Process: workerID,
			Value:   value,
			Error:   err.Error(),
		})
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		history.append(jepsenBankHistoryEntry{
			Type:    classifyError(err),
			Op:      jepsenBankOpTransfer,
			Process: workerID,
			Value:   value,
			Error:   err.Error(),
		})
		return
	}
	if affected == 0 {
		// Insufficient balance; abort.
		_ = tx.Rollback()
		history.append(jepsenBankHistoryEntry{
			Type:    jepsenBankOutcomeFail,
			Op:      jepsenBankOpTransfer,
			Process: workerID,
			Value:   value,
			Error:   "insufficient balance",
		})
		return
	}

	// Credit the destination account.
	if _, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
		amount, to,
	); err != nil {
		_ = tx.Rollback()
		history.append(jepsenBankHistoryEntry{
			Type:    classifyError(err),
			Op:      jepsenBankOpTransfer,
			Process: workerID,
			Value:   value,
			Error:   err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		history.append(jepsenBankHistoryEntry{
			Type:    classifyError(err),
			Op:      jepsenBankOpTransfer,
			Process: workerID,
			Value:   value,
			Error:   err.Error(),
		})
		return
	}

	history.append(jepsenBankHistoryEntry{
		Type:    jepsenBankOutcomeOk,
		Op:      jepsenBankOpTransfer,
		Process: workerID,
		Value:   value,
	})
}

// runJepsenBankNemesis runs the nemesis loop: wait, inject, hold, recover,
// repeat. On context cancellation, it performs a final recovery.
func runJepsenBankNemesis(
	ctx context.Context,
	l *logger.Logger,
	t test.Test,
	c cluster.Cluster,
	nodes option.NodeListOption,
	nemesis jepsenBankNemesisType,
	seed int64,
) error {
	rng := rand.New(rand.NewSource(seed))

	switch nemesis {
	case jepsenBankNemesisStartKill2:
		return runStartKill2Nemesis(ctx, l, t, c, nodes, rng)
	case jepsenBankNemesisMajorityRing:
		return runMajorityRingNemesis(ctx, l, t, c, nodes, rng)
	default:
		return fmt.Errorf("unknown nemesis: %d", nemesis)
	}
}

func runStartKill2Nemesis(
	ctx context.Context,
	l *logger.Logger,
	t test.Test,
	c cluster.Cluster,
	nodes option.NodeListOption,
	rng *rand.Rand,
) error {
	// Create the failer once; we'll pass different args on each inject round.
	// Use a dummy node set for initial creation; the actual targets are
	// specified per-Inject call.
	failer, _, err := roachtestutil.MakeProcessKillFailer(
		l, c, nodes, false /* graceful */, 0,
	)
	if err != nil {
		return err
	}

	failerLogger := newFailerLogger(l, nodes, "jepsen-bank-start-kill-2")
	defer func() {
		if cleanupErr := failer.Cleanup(context.Background(), failerLogger); cleanupErr != nil {
			l.Printf("nemesis cleanup failed: %v", cleanupErr)
		}
	}()

	if err := failer.Setup(ctx, failerLogger, failures.ProcessKillArgs{
		Nodes: nodes.InstallNodes(),
	}); err != nil {
		return err
	}

	for ctx.Err() == nil {
		// Wait 5-15s before injecting.
		if !sleepWithContext(ctx, time.Duration(5+rng.Intn(11))*time.Second) {
			break
		}

		// Pick 2 random nodes to kill.
		targets, err := nodes.SeededRandList(rng, 2)
		if err != nil {
			return err
		}
		l.Printf("nemesis: killing nodes %s", targets)

		killArgs := failures.ProcessKillArgs{
			Nodes: targets.InstallNodes(),
		}
		if err := failer.Inject(ctx, failerLogger, killArgs); err != nil {
			l.Printf("nemesis: inject failed: %v", err)
			break
		}

		if err := failer.WaitForFailureToPropagate(ctx, failerLogger); err != nil {
			l.Printf("nemesis: wait for propagation failed: %v", err)
		}

		// Hold failure for 15-30s.
		if !sleepWithContext(ctx, time.Duration(15+rng.Intn(16))*time.Second) {
			// Context cancelled; recover before exiting.
			if recoverErr := failer.Recover(context.Background(), failerLogger); recoverErr != nil {
				l.Printf("nemesis: final recover failed: %v", recoverErr)
			}
			break
		}

		l.Printf("nemesis: restarting nodes %s", targets)
		if err := failer.Recover(ctx, failerLogger); err != nil {
			l.Printf("nemesis: recover failed: %v", err)
			break
		}
		if err := failer.WaitForFailureToRecover(ctx, failerLogger); err != nil {
			l.Printf("nemesis: wait for recovery failed: %v", err)
		}
	}
	return nil
}

func runMajorityRingNemesis(
	ctx context.Context,
	l *logger.Logger,
	t test.Test,
	c cluster.Cluster,
	nodes option.NodeListOption,
	rng *rand.Rand,
) error {
	// We'll create a new failer for each partition round because the
	// MakeNetworkPartitionFailer helper takes source/dest at creation time,
	// and the Failer stores setupArgs. Instead, we create one per round.
	for ctx.Err() == nil {
		// Wait 5-15s before injecting.
		if !sleepWithContext(ctx, time.Duration(5+rng.Intn(11))*time.Second) {
			break
		}

		// Pick a random 2-node minority.
		minority, err := nodes.SeededRandList(rng, 2)
		if err != nil {
			return err
		}
		// Compute the 3-node majority as the complement.
		var majority option.NodeListOption
		for _, n := range nodes {
			if !minority.Contains(n) {
				majority = append(majority, n)
			}
		}
		l.Printf("nemesis: partitioning minority %s from majority %s", minority, majority)

		failer, partArgs, err := roachtestutil.MakeBidirectionalPartitionFailer(
			l, c, minority, majority,
		)
		if err != nil {
			l.Printf("nemesis: create failer failed: %v", err)
			break
		}
		failerLogger := newFailerLogger(l, nodes, "jepsen-bank-majority-ring")

		if err := failer.Setup(ctx, failerLogger, partArgs); err != nil {
			l.Printf("nemesis: setup failed: %v", err)
			break
		}

		if err := failer.Inject(ctx, failerLogger, partArgs); err != nil {
			l.Printf("nemesis: inject failed: %v", err)
			_ = failer.Cleanup(context.Background(), failerLogger)
			break
		}

		// Hold partition for 15-30s.
		if !sleepWithContext(ctx, time.Duration(15+rng.Intn(16))*time.Second) {
			// Context cancelled; recover before exiting.
			if recoverErr := failer.Recover(context.Background(), failerLogger); recoverErr != nil {
				l.Printf("nemesis: final recover failed: %v", recoverErr)
			}
			_ = failer.Cleanup(context.Background(), failerLogger)
			break
		}

		l.Printf("nemesis: healing partition")
		if err := failer.Recover(ctx, failerLogger); err != nil {
			l.Printf("nemesis: recover failed: %v", err)
			_ = failer.Cleanup(context.Background(), failerLogger)
			break
		}

		if err := failer.Cleanup(ctx, failerLogger); err != nil {
			l.Printf("nemesis: cleanup failed: %v", err)
			break
		}
	}
	return nil
}

// sleepWithContext sleeps for the given duration or until the context is
// cancelled. Returns true if the sleep completed, false if cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}
