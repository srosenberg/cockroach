package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type UpdateFn func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error

type RunStats struct {
	LastRun time.Time
	NumRuns int
}

type JobMetadata struct {
	ID       jobspb.JobID
	Status   Status
	Payload  *jobspb.Payload
	Progress *jobspb.Progress
	RunStats *RunStats
}

func (md *JobMetadata) CheckRunningOrReverting() error {
	__antithesis_instrumentation__.Notify(84952)
	if md.Status != StatusRunning && func() bool {
		__antithesis_instrumentation__.Notify(84954)
		return md.Status != StatusReverting == true
	}() == true {
		__antithesis_instrumentation__.Notify(84955)
		return &InvalidStatusError{md.ID, md.Status, "update progress on", md.Payload.Error}
	} else {
		__antithesis_instrumentation__.Notify(84956)
	}
	__antithesis_instrumentation__.Notify(84953)
	return nil
}

type JobUpdater struct {
	md JobMetadata
}

func (ju *JobUpdater) UpdateStatus(status Status) {
	__antithesis_instrumentation__.Notify(84957)
	ju.md.Status = status
}

func (ju *JobUpdater) UpdatePayload(payload *jobspb.Payload) {
	__antithesis_instrumentation__.Notify(84958)
	ju.md.Payload = payload
}

func (ju *JobUpdater) UpdateProgress(progress *jobspb.Progress) {
	__antithesis_instrumentation__.Notify(84959)
	ju.md.Progress = progress
}

func (ju *JobUpdater) hasUpdates() bool {
	__antithesis_instrumentation__.Notify(84960)
	return ju.md != JobMetadata{}
}

func (ju *JobUpdater) UpdateRunStats(numRuns int, lastRun time.Time) {
	__antithesis_instrumentation__.Notify(84961)
	ju.md.RunStats = &RunStats{
		NumRuns: numRuns,
		LastRun: lastRun,
	}
}

func UpdateHighwaterProgressed(highWater hlc.Timestamp, md JobMetadata, ju *JobUpdater) error {
	__antithesis_instrumentation__.Notify(84962)
	if err := md.CheckRunningOrReverting(); err != nil {
		__antithesis_instrumentation__.Notify(84965)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84966)
	}
	__antithesis_instrumentation__.Notify(84963)

	if highWater.Less(hlc.Timestamp{}) {
		__antithesis_instrumentation__.Notify(84967)
		return errors.Errorf("high-water %s is outside allowable range > 0.0", highWater)
	} else {
		__antithesis_instrumentation__.Notify(84968)
	}
	__antithesis_instrumentation__.Notify(84964)
	md.Progress.Progress = &jobspb.Progress_HighWater{
		HighWater: &highWater,
	}
	ju.UpdateProgress(md.Progress)
	return nil
}

func (j *Job) Update(ctx context.Context, txn *kv.Txn, updateFn UpdateFn) error {
	__antithesis_instrumentation__.Notify(84969)
	const useReadLock = false
	return j.update(ctx, txn, useReadLock, updateFn)
}

func (j *Job) update(ctx context.Context, txn *kv.Txn, useReadLock bool, updateFn UpdateFn) error {
	__antithesis_instrumentation__.Notify(84970)
	var payload *jobspb.Payload
	var progress *jobspb.Progress
	var status Status
	var runStats *RunStats

	if err := j.runInTxn(ctx, txn, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(84973)
		payload, progress, runStats = nil, nil, nil
		var err error
		var row tree.Datums
		row, err = j.registry.ex.QueryRowEx(
			ctx, "log-job", txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			getSelectStmtForJobUpdate(j.session != nil, useReadLock), j.ID(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(84993)
			return err
		} else {
			__antithesis_instrumentation__.Notify(84994)
		}
		__antithesis_instrumentation__.Notify(84974)
		if row == nil {
			__antithesis_instrumentation__.Notify(84995)
			return errors.Errorf("not found in system.jobs table")
		} else {
			__antithesis_instrumentation__.Notify(84996)
		}
		__antithesis_instrumentation__.Notify(84975)

		if status, err = unmarshalStatus(row[0]); err != nil {
			__antithesis_instrumentation__.Notify(84997)
			return err
		} else {
			__antithesis_instrumentation__.Notify(84998)
		}
		__antithesis_instrumentation__.Notify(84976)
		if payload, err = UnmarshalPayload(row[1]); err != nil {
			__antithesis_instrumentation__.Notify(84999)
			return err
		} else {
			__antithesis_instrumentation__.Notify(85000)
		}
		__antithesis_instrumentation__.Notify(84977)
		if progress, err = UnmarshalProgress(row[2]); err != nil {
			__antithesis_instrumentation__.Notify(85001)
			return err
		} else {
			__antithesis_instrumentation__.Notify(85002)
		}
		__antithesis_instrumentation__.Notify(84978)
		if j.session != nil {
			__antithesis_instrumentation__.Notify(85003)
			if row[3] == tree.DNull {
				__antithesis_instrumentation__.Notify(85005)
				return errors.Errorf(
					"with status %q: expected session %q but found NULL",
					status, j.session.ID())
			} else {
				__antithesis_instrumentation__.Notify(85006)
			}
			__antithesis_instrumentation__.Notify(85004)
			storedSession := []byte(*row[3].(*tree.DBytes))
			if !bytes.Equal(storedSession, j.session.ID().UnsafeBytes()) {
				__antithesis_instrumentation__.Notify(85007)
				return errors.Errorf(
					"with status %q: expected session %q but found %q",
					status, j.session.ID(), sqlliveness.SessionID(storedSession))
			} else {
				__antithesis_instrumentation__.Notify(85008)
			}
		} else {
			__antithesis_instrumentation__.Notify(85009)
			log.VInfof(ctx, 1, "job %d: update called with no session ID", j.ID())
		}
		__antithesis_instrumentation__.Notify(84979)

		md := JobMetadata{
			ID:       j.ID(),
			Status:   status,
			Payload:  payload,
			Progress: progress,
		}

		offset := 0
		if j.session != nil {
			__antithesis_instrumentation__.Notify(85010)
			offset = 1
		} else {
			__antithesis_instrumentation__.Notify(85011)
		}
		__antithesis_instrumentation__.Notify(84980)
		var lastRun *tree.DTimestamp
		var ok bool
		lastRun, ok = row[3+offset].(*tree.DTimestamp)
		if !ok {
			__antithesis_instrumentation__.Notify(85012)
			return errors.AssertionFailedf("expected timestamp last_run, but got %T", lastRun)
		} else {
			__antithesis_instrumentation__.Notify(85013)
		}
		__antithesis_instrumentation__.Notify(84981)
		var numRuns *tree.DInt
		numRuns, ok = row[4+offset].(*tree.DInt)
		if !ok {
			__antithesis_instrumentation__.Notify(85014)
			return errors.AssertionFailedf("expected int num_runs, but got %T", numRuns)
		} else {
			__antithesis_instrumentation__.Notify(85015)
		}
		__antithesis_instrumentation__.Notify(84982)
		md.RunStats = &RunStats{
			NumRuns: int(*numRuns),
			LastRun: lastRun.Time,
		}

		var ju JobUpdater
		if err := updateFn(txn, md, &ju); err != nil {
			__antithesis_instrumentation__.Notify(85016)
			return err
		} else {
			__antithesis_instrumentation__.Notify(85017)
		}
		__antithesis_instrumentation__.Notify(84983)
		if j.registry.knobs.BeforeUpdate != nil {
			__antithesis_instrumentation__.Notify(85018)
			if err := j.registry.knobs.BeforeUpdate(md, ju.md); err != nil {
				__antithesis_instrumentation__.Notify(85019)
				return err
			} else {
				__antithesis_instrumentation__.Notify(85020)
			}
		} else {
			__antithesis_instrumentation__.Notify(85021)
		}
		__antithesis_instrumentation__.Notify(84984)

		if !ju.hasUpdates() {
			__antithesis_instrumentation__.Notify(85022)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(85023)
		}
		__antithesis_instrumentation__.Notify(84985)

		var setters []string
		params := []interface{}{j.ID()}
		addSetter := func(column string, value interface{}) {
			__antithesis_instrumentation__.Notify(85024)
			params = append(params, value)
			setters = append(setters, fmt.Sprintf("%s = $%d", column, len(params)))
		}
		__antithesis_instrumentation__.Notify(84986)

		if ju.md.Status != "" {
			__antithesis_instrumentation__.Notify(85025)
			addSetter("status", ju.md.Status)
		} else {
			__antithesis_instrumentation__.Notify(85026)
		}
		__antithesis_instrumentation__.Notify(84987)

		if ju.md.Payload != nil {
			__antithesis_instrumentation__.Notify(85027)
			payload = ju.md.Payload
			payloadBytes, err := protoutil.Marshal(payload)
			if err != nil {
				__antithesis_instrumentation__.Notify(85029)
				return err
			} else {
				__antithesis_instrumentation__.Notify(85030)
			}
			__antithesis_instrumentation__.Notify(85028)
			addSetter("payload", payloadBytes)
		} else {
			__antithesis_instrumentation__.Notify(85031)
		}
		__antithesis_instrumentation__.Notify(84988)

		if ju.md.Progress != nil {
			__antithesis_instrumentation__.Notify(85032)
			progress = ju.md.Progress
			progress.ModifiedMicros = timeutil.ToUnixMicros(txn.ReadTimestamp().GoTime())
			progressBytes, err := protoutil.Marshal(progress)
			if err != nil {
				__antithesis_instrumentation__.Notify(85034)
				return err
			} else {
				__antithesis_instrumentation__.Notify(85035)
			}
			__antithesis_instrumentation__.Notify(85033)
			addSetter("progress", progressBytes)
		} else {
			__antithesis_instrumentation__.Notify(85036)
		}
		__antithesis_instrumentation__.Notify(84989)

		if ju.md.RunStats != nil {
			__antithesis_instrumentation__.Notify(85037)
			runStats = ju.md.RunStats
			addSetter("last_run", ju.md.RunStats.LastRun)
			addSetter("num_runs", ju.md.RunStats.NumRuns)
		} else {
			__antithesis_instrumentation__.Notify(85038)
		}
		__antithesis_instrumentation__.Notify(84990)

		updateStmt := fmt.Sprintf(
			"UPDATE system.jobs SET %s WHERE id = $1",
			strings.Join(setters, ", "),
		)
		n, err := j.registry.ex.Exec(ctx, "job-update", txn, updateStmt, params...)
		if err != nil {
			__antithesis_instrumentation__.Notify(85039)
			return err
		} else {
			__antithesis_instrumentation__.Notify(85040)
		}
		__antithesis_instrumentation__.Notify(84991)
		if n != 1 {
			__antithesis_instrumentation__.Notify(85041)
			return errors.Errorf(
				"expected exactly one row affected, but %d rows affected by job update", n,
			)
		} else {
			__antithesis_instrumentation__.Notify(85042)
		}
		__antithesis_instrumentation__.Notify(84992)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(85043)
		return errors.Wrapf(err, "job %d", j.id)
	} else {
		__antithesis_instrumentation__.Notify(85044)
	}
	__antithesis_instrumentation__.Notify(84971)
	func() {
		__antithesis_instrumentation__.Notify(85045)
		j.mu.Lock()
		defer j.mu.Unlock()
		if payload != nil {
			__antithesis_instrumentation__.Notify(85049)
			j.mu.payload = *payload
		} else {
			__antithesis_instrumentation__.Notify(85050)
		}
		__antithesis_instrumentation__.Notify(85046)
		if progress != nil {
			__antithesis_instrumentation__.Notify(85051)
			j.mu.progress = *progress
		} else {
			__antithesis_instrumentation__.Notify(85052)
		}
		__antithesis_instrumentation__.Notify(85047)
		if runStats != nil {
			__antithesis_instrumentation__.Notify(85053)
			j.mu.runStats = runStats
		} else {
			__antithesis_instrumentation__.Notify(85054)
		}
		__antithesis_instrumentation__.Notify(85048)
		if status != "" {
			__antithesis_instrumentation__.Notify(85055)
			j.mu.status = status
		} else {
			__antithesis_instrumentation__.Notify(85056)
		}
	}()
	__antithesis_instrumentation__.Notify(84972)
	return nil
}

func getSelectStmtForJobUpdate(hasSession, useReadLock bool) string {
	__antithesis_instrumentation__.Notify(85057)
	const (
		selectWithoutSession = `SELECT status, payload, progress`
		selectWithSession    = selectWithoutSession + `, claim_session_id`
		from                 = ` FROM system.jobs WHERE id = $1`
		fromForUpdate        = from + ` FOR UPDATE`
		backoffColumns       = ", COALESCE(last_run, created), COALESCE(num_runs, 0)"
	)
	stmt := selectWithoutSession
	if hasSession {
		__antithesis_instrumentation__.Notify(85060)
		stmt = selectWithSession
	} else {
		__antithesis_instrumentation__.Notify(85061)
	}
	__antithesis_instrumentation__.Notify(85058)
	stmt = stmt + backoffColumns
	if useReadLock {
		__antithesis_instrumentation__.Notify(85062)
		return stmt + fromForUpdate
	} else {
		__antithesis_instrumentation__.Notify(85063)
	}
	__antithesis_instrumentation__.Notify(85059)
	return stmt + from
}
