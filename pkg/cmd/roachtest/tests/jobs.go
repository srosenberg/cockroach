package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

type jobStarter func(c cluster.Cluster, t test.Test) (string, error)

func jobSurvivesNodeShutdown(
	ctx context.Context, t test.Test, c cluster.Cluster, nodeToShutdown int, startJob jobStarter,
) {
	__antithesis_instrumentation__.Notify(48796)
	watcherNode := 1 + (nodeToShutdown)%c.Spec().NodeCount
	target := c.Node(nodeToShutdown)
	t.L().Printf("test has chosen shutdown target node %d, and watcher node %d",
		nodeToShutdown, watcherNode)

	jobIDCh := make(chan string, 1)

	m := c.NewMonitor(ctx)
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(48799)
		defer close(jobIDCh)

		watcherDB := c.Conn(ctx, t.L(), watcherNode)
		defer watcherDB.Close()

		t.Status("waiting for cluster to be 3x replicated")
		err := WaitFor3XReplication(ctx, t, watcherDB)
		require.NoError(t, err)

		t.Status("running job")
		var jobID string
		jobID, err = startJob(c, t)
		if err != nil {
			__antithesis_instrumentation__.Notify(48801)
			return errors.Wrap(err, "starting the job")
		} else {
			__antithesis_instrumentation__.Notify(48802)
		}
		__antithesis_instrumentation__.Notify(48800)
		t.L().Printf("started running job with ID %s", jobID)
		jobIDCh <- jobID

		pollInterval := 5 * time.Second
		ticker := time.NewTicker(pollInterval)
		var status string
		for {
			__antithesis_instrumentation__.Notify(48803)
			select {
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(48804)
				err := watcherDB.QueryRowContext(ctx, `SELECT status FROM [SHOW JOBS] WHERE job_id=$1`, jobID).Scan(&status)
				if err != nil {
					__antithesis_instrumentation__.Notify(48807)
					return errors.Wrap(err, "getting the job status")
				} else {
					__antithesis_instrumentation__.Notify(48808)
				}
				__antithesis_instrumentation__.Notify(48805)
				jobStatus := jobs.Status(status)
				switch jobStatus {
				case jobs.StatusSucceeded:
					__antithesis_instrumentation__.Notify(48809)
					t.Status("job completed")
					return nil
				case jobs.StatusRunning:
					__antithesis_instrumentation__.Notify(48810)
					t.L().Printf("job %s still running, waiting to succeed", jobID)
				default:
					__antithesis_instrumentation__.Notify(48811)

					return errors.Newf("unexpectedly found job %s in state %s", jobID, status)
				}
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(48806)
				return errors.Wrap(ctx.Err(), "context canceled while waiting for job to finish")
			}
		}
	})
	__antithesis_instrumentation__.Notify(48797)

	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(48812)
		jobID, ok := <-jobIDCh
		if !ok {
			__antithesis_instrumentation__.Notify(48816)
			return errors.New("job never created")
		} else {
			__antithesis_instrumentation__.Notify(48817)
		}
		__antithesis_instrumentation__.Notify(48813)

		watcherDB := c.Conn(ctx, t.L(), watcherNode)
		defer watcherDB.Close()
		timeToWait := time.Second
		timer := timeutil.Timer{}
		jobRunning := false
		for {
			__antithesis_instrumentation__.Notify(48818)
			var status string
			err := watcherDB.QueryRowContext(ctx, `SELECT status FROM [SHOW JOBS] WHERE job_id=$1`, jobID).Scan(&status)
			if err != nil {
				__antithesis_instrumentation__.Notify(48822)
				return errors.Wrap(err, "getting the job status")
			} else {
				__antithesis_instrumentation__.Notify(48823)
			}
			__antithesis_instrumentation__.Notify(48819)
			switch jobs.Status(status) {
			case jobs.StatusPending:
				__antithesis_instrumentation__.Notify(48824)
			case jobs.StatusRunning:
				__antithesis_instrumentation__.Notify(48825)
				jobRunning = true
			default:
				__antithesis_instrumentation__.Notify(48826)
				return errors.Newf("job too fast! job got to state %s before the target node could be shutdown",
					status)
			}
			__antithesis_instrumentation__.Notify(48820)
			t.L().Printf(`status %s`, status)
			timer.Reset(timeToWait)
			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(48827)
				return errors.Wrapf(ctx.Err(), "stopping test, did not shutdown node")
			case <-timer.C:
				__antithesis_instrumentation__.Notify(48828)
				timer.Read = true
			}
			__antithesis_instrumentation__.Notify(48821)

			if jobRunning {
				__antithesis_instrumentation__.Notify(48829)
				break
			} else {
				__antithesis_instrumentation__.Notify(48830)
			}
		}
		__antithesis_instrumentation__.Notify(48814)

		t.L().Printf(`stopping node %s`, target)
		if err := c.StopE(ctx, t.L(), option.DefaultStopOpts(), target); err != nil {
			__antithesis_instrumentation__.Notify(48831)
			return errors.Wrapf(err, "could not stop node %s", target)
		} else {
			__antithesis_instrumentation__.Notify(48832)
		}
		__antithesis_instrumentation__.Notify(48815)
		t.L().Printf("stopped node %s", target)

		return nil
	})
	__antithesis_instrumentation__.Notify(48798)

	m.ExpectDeath()
	m.Wait()

	t.Status(fmt.Sprintf("restarting %s (node restart test is done)\n", target))
	if err := c.StartE(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), target); err != nil {
		__antithesis_instrumentation__.Notify(48833)
		t.Fatal(errors.Wrapf(err, "could not restart node %s", target))
	} else {
		__antithesis_instrumentation__.Notify(48834)
	}
}
