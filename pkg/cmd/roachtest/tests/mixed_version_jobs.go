package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/testutils"
)

type backgroundFn func(ctx context.Context, u *versionUpgradeTest) error

type backgroundStepper struct {
	run backgroundFn

	onStop func(context.Context, test.Test, *versionUpgradeTest, error)
	nodes  option.NodeListOption

	m cluster.Monitor
}

func (s *backgroundStepper) launch(ctx context.Context, t test.Test, u *versionUpgradeTest) {
	__antithesis_instrumentation__.Notify(49322)
	nodes := s.nodes
	if nodes == nil {
		__antithesis_instrumentation__.Notify(49324)
		nodes = u.c.All()
	} else {
		__antithesis_instrumentation__.Notify(49325)
	}
	__antithesis_instrumentation__.Notify(49323)
	s.m = u.c.NewMonitor(ctx, nodes)
	s.m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(49326)
		return s.run(ctx, u)
	})
}

func (s *backgroundStepper) wait(ctx context.Context, t test.Test, u *versionUpgradeTest) {
	__antithesis_instrumentation__.Notify(49327)

	err := s.m.WaitE()
	if s.onStop != nil {
		__antithesis_instrumentation__.Notify(49328)
		s.onStop(ctx, t, u, err)
	} else {
		__antithesis_instrumentation__.Notify(49329)
		if err != nil {
			__antithesis_instrumentation__.Notify(49330)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49331)
		}
	}
}

func overrideErrorFromJobsTable(ctx context.Context, t test.Test, u *versionUpgradeTest, _ error) {
	__antithesis_instrumentation__.Notify(49332)
	db := u.conn(ctx, t, 1)
	t.L().Printf("Resuming any paused jobs left")
	for {
		__antithesis_instrumentation__.Notify(49335)
		_, err := db.ExecContext(
			ctx,
			`RESUME JOBS (SELECT job_id FROM [SHOW JOBS] WHERE status = $1);`,
			jobs.StatusPaused,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(49339)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49340)
		}
		__antithesis_instrumentation__.Notify(49336)
		row := db.QueryRow(
			"SELECT count(*) FROM [SHOW JOBS] WHERE status = $1",
			jobs.StatusPauseRequested,
		)
		var nNotYetPaused int
		if err = row.Scan(&nNotYetPaused); err != nil {
			__antithesis_instrumentation__.Notify(49341)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49342)
		}
		__antithesis_instrumentation__.Notify(49337)
		if nNotYetPaused <= 0 {
			__antithesis_instrumentation__.Notify(49343)
			break
		} else {
			__antithesis_instrumentation__.Notify(49344)
		}
		__antithesis_instrumentation__.Notify(49338)

		time.Sleep(10 * time.Second)
		t.L().Printf("Waiting for %d jobs to pause", nNotYetPaused)
	}
	__antithesis_instrumentation__.Notify(49333)

	t.L().Printf("Waiting for jobs to complete...")
	var err error
	for {
		__antithesis_instrumentation__.Notify(49345)
		q := "SHOW JOBS WHEN COMPLETE (SELECT job_id FROM [SHOW JOBS]);"
		_, err = db.ExecContext(ctx, q)
		if testutils.IsError(err, "pq: restart transaction:.*") {
			__antithesis_instrumentation__.Notify(49347)
			t.L().Printf("SHOW JOBS WHEN COMPLETE returned %s, retrying", err.Error())
			time.Sleep(10 * time.Second)
			continue
		} else {
			__antithesis_instrumentation__.Notify(49348)
		}
		__antithesis_instrumentation__.Notify(49346)
		break
	}
	__antithesis_instrumentation__.Notify(49334)
	if err != nil {
		__antithesis_instrumentation__.Notify(49349)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49350)
	}
}

func backgroundJobsTestTPCCImport(t test.Test, warehouses int) backgroundStepper {
	__antithesis_instrumentation__.Notify(49351)
	return backgroundStepper{run: func(ctx context.Context, u *versionUpgradeTest) error {
		__antithesis_instrumentation__.Notify(49352)

		err := u.c.RunE(ctx, u.c.Node(1), tpccImportCmd(warehouses))
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(49354)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(49355)
		}
		__antithesis_instrumentation__.Notify(49353)
		return err
	},
		onStop: overrideErrorFromJobsTable,
	}
}

func pauseAllJobsStep() versionStep {
	__antithesis_instrumentation__.Notify(49356)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49357)
		db := u.conn(ctx, t, 1)
		_, err := db.ExecContext(
			ctx,
			`PAUSE JOBS (SELECT job_id FROM [SHOW JOBS] WHERE status = $1);`,
			jobs.StatusRunning,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(49360)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49361)
		}
		__antithesis_instrumentation__.Notify(49358)

		row := db.QueryRow("SELECT count(*) FROM [SHOW JOBS] WHERE status LIKE 'pause%'")
		var nPaused int
		if err := row.Scan(&nPaused); err != nil {
			__antithesis_instrumentation__.Notify(49362)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49363)
		}
		__antithesis_instrumentation__.Notify(49359)
		t.L().Printf("Paused %d jobs", nPaused)
		time.Sleep(time.Second)
	}
}

func makeResumeAllJobsAndWaitStep(d time.Duration) versionStep {
	__antithesis_instrumentation__.Notify(49364)
	var numResumes int
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49365)
		numResumes++
		t.L().Printf("Resume all jobs number: %d", numResumes)
		db := u.conn(ctx, t, 1)
		_, err := db.ExecContext(
			ctx,
			`RESUME JOBS (SELECT job_id FROM [SHOW JOBS] WHERE status = $1);`,
			jobs.StatusPaused,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(49368)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49369)
		}
		__antithesis_instrumentation__.Notify(49366)

		row := db.QueryRow(
			"SELECT count(*) FROM [SHOW JOBS] WHERE status = $1",
			jobs.StatusRunning,
		)
		var nRunning int
		if err := row.Scan(&nRunning); err != nil {
			__antithesis_instrumentation__.Notify(49370)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49371)
		}
		__antithesis_instrumentation__.Notify(49367)
		t.L().Printf("Resumed %d jobs", nRunning)
		time.Sleep(d)
	}
}

func checkForFailedJobsStep(ctx context.Context, t test.Test, u *versionUpgradeTest) {
	__antithesis_instrumentation__.Notify(49372)
	t.L().Printf("Checking for failed jobs.")

	db := u.conn(ctx, t, 1)

	rows, err := db.Query(`
SELECT job_id, job_type, description, status, error, ifnull(coordinator_id, 0)
FROM [SHOW JOBS] WHERE status = $1 OR status = $2`,
		jobs.StatusFailed, jobs.StatusReverting,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(49375)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49376)
	}
	__antithesis_instrumentation__.Notify(49373)
	var jobType, desc, status, jobError string
	var jobID jobspb.JobID
	var coordinatorID int64
	var errMsg string
	for rows.Next() {
		__antithesis_instrumentation__.Notify(49377)
		err := rows.Scan(&jobID, &jobType, &desc, &status, &jobError, &coordinatorID)
		if err != nil {
			__antithesis_instrumentation__.Notify(49379)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49380)
		}
		__antithesis_instrumentation__.Notify(49378)

		errMsg = fmt.Sprintf(
			"%sUnsuccessful job %d of type %s, description %s, status %s, error %s, coordinator %d\n",
			errMsg, jobID, jobType, desc, status, jobError, coordinatorID,
		)
	}
	__antithesis_instrumentation__.Notify(49374)
	if errMsg != "" {
		__antithesis_instrumentation__.Notify(49381)
		nodeInfo := "Cluster info\n"
		for i := range u.c.All() {
			__antithesis_instrumentation__.Notify(49383)
			nodeInfo = fmt.Sprintf(
				"%sNode %d: %s\n", nodeInfo, i+1, u.binaryVersion(ctx, t, i+1))
		}
		__antithesis_instrumentation__.Notify(49382)
		t.Fatalf("%s\n%s", nodeInfo, errMsg)
	} else {
		__antithesis_instrumentation__.Notify(49384)
	}
}

func runJobsMixedVersions(
	ctx context.Context, t test.Test, c cluster.Cluster, warehouses int, predecessorVersion string,
) {
	__antithesis_instrumentation__.Notify(49385)

	const mainVersion = ""
	roachNodes := c.All()
	backgroundTPCC := backgroundJobsTestTPCCImport(t, warehouses)
	resumeAllJobsAndWaitStep := makeResumeAllJobsAndWaitStep(10 * time.Second)
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(1))

	u := newVersionUpgradeTest(c,
		uploadAndStartFromCheckpointFixture(roachNodes, predecessorVersion),
		waitForUpgradeStep(roachNodes),
		preventAutoUpgradeStep(1),

		backgroundTPCC.launch,
		func(ctx context.Context, _ test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(49387)
			time.Sleep(10 * time.Second)
		},
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(3), mainVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(2), mainVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(1), mainVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(4), mainVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(2), predecessorVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(4), predecessorVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(3), predecessorVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(1), predecessorVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(4), mainVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(3), mainVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(1), mainVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		binaryUpgradeStep(c.Node(2), mainVersion),
		resumeAllJobsAndWaitStep,
		checkForFailedJobsStep,
		pauseAllJobsStep(),

		allowAutoUpgradeStep(1),
		waitForUpgradeStep(roachNodes),
		resumeAllJobsAndWaitStep,
		backgroundTPCC.wait,
		checkForFailedJobsStep,
	)
	__antithesis_instrumentation__.Notify(49386)
	u.run(ctx, t)
}

func registerJobsMixedVersions(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49388)
	r.Add(registry.TestSpec{
		Name:  "jobs/mixed-versions",
		Owner: registry.OwnerBulkIO,
		Skip:  "#67587",

		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49389)
			predV, err := PredecessorVersion(*t.BuildVersion())
			if err != nil {
				__antithesis_instrumentation__.Notify(49391)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49392)
			}
			__antithesis_instrumentation__.Notify(49390)
			warehouses := 10
			runJobsMixedVersions(ctx, t, c, warehouses, predV)
		},
	})
}
