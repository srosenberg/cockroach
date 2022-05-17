package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cloudstorage "github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/cloud/amazon"
	"github.com/cockroachdb/cockroach/pkg/cloud/gcp"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

const (
	KMSRegionAEnvVar  = "AWS_KMS_REGION_A"
	KMSRegionBEnvVar  = "AWS_KMS_REGION_B"
	KMSKeyARNAEnvVar  = "AWS_KMS_KEY_ARN_A"
	KMSKeyARNBEnvVar  = "AWS_KMS_KEY_ARN_B"
	KMSKeyNameAEnvVar = "GOOGLE_KMS_KEY_A"
	KMSKeyNameBEnvVar = "GOOGLE_KMS_KEY_B"
	KMSGCSCredentials = "GOOGLE_EPHEMERAL_CREDENTIALS"

	rows2TiB   = 65_104_166
	rows100GiB = rows2TiB / 20
	rows30GiB  = rows2TiB / 66
	rows15GiB  = rows30GiB / 2
	rows5GiB   = rows100GiB / 20
	rows3GiB   = rows30GiB / 10
)

func destinationName(c cluster.Cluster) string {
	__antithesis_instrumentation__.Notify(45620)
	dest := c.Name()
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(45622)
		dest += fmt.Sprintf("%d", timeutil.Now().UnixNano())
	} else {
		__antithesis_instrumentation__.Notify(45623)
	}
	__antithesis_instrumentation__.Notify(45621)
	return dest
}

func importBankDataSplit(
	ctx context.Context, rows, ranges int, t test.Test, c cluster.Cluster,
) string {
	__antithesis_instrumentation__.Notify(45624)
	dest := destinationName(c)

	c.Put(ctx, t.DeprecatedWorkload(), "./workload")
	c.Put(ctx, t.Cockroach(), "./cockroach")

	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
	runImportBankDataSplit(ctx, rows, ranges, t, c)
	return dest
}

func runImportBankDataSplit(ctx context.Context, rows, ranges int, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(45625)
	c.Run(ctx, c.All(), `./workload csv-server --port=8081 &> logs/workload-csv-server.log < /dev/null &`)
	time.Sleep(time.Second)
	importArgs := []string{
		"./workload", "fixtures", "import", "bank",
		"--db=bank",
		"--payload-bytes=10240",
		"--csv-server", "http://localhost:8081",
		"--seed=1",
		fmt.Sprintf("--ranges=%d", ranges),
		fmt.Sprintf("--rows=%d", rows),
		"{pgurl:1}",
	}
	c.Run(ctx, c.Node(1), importArgs...)
}

func importBankData(ctx context.Context, rows int, t test.Test, c cluster.Cluster) string {
	__antithesis_instrumentation__.Notify(45626)
	return importBankDataSplit(ctx, rows, 0, t, c)
}

func registerBackupNodeShutdown(r registry.Registry) {
	__antithesis_instrumentation__.Notify(45627)

	backupNodeRestartSpec := r.MakeClusterSpec(4)
	loadBackupData := func(ctx context.Context, t test.Test, c cluster.Cluster) string {
		__antithesis_instrumentation__.Notify(45630)

		rows := rows15GiB
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(45632)

			rows = rows5GiB
		} else {
			__antithesis_instrumentation__.Notify(45633)
		}
		__antithesis_instrumentation__.Notify(45631)
		return importBankData(ctx, rows, t, c)
	}
	__antithesis_instrumentation__.Notify(45628)

	r.Add(registry.TestSpec{
		Name:            fmt.Sprintf("backup/nodeShutdown/worker/%s", backupNodeRestartSpec),
		Owner:           registry.OwnerBulkIO,
		Cluster:         backupNodeRestartSpec,
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45634)
			gatewayNode := 2
			nodeToShutdown := 3
			dest := loadBackupData(ctx, t, c)
			backupQuery := `BACKUP bank.bank TO 'nodelocal://1/` + dest + `' WITH DETACHED`
			startBackup := func(c cluster.Cluster, t test.Test) (jobID string, err error) {
				__antithesis_instrumentation__.Notify(45636)
				gatewayDB := c.Conn(ctx, t.L(), gatewayNode)
				defer gatewayDB.Close()

				err = gatewayDB.QueryRowContext(ctx, backupQuery).Scan(&jobID)
				return
			}
			__antithesis_instrumentation__.Notify(45635)

			jobSurvivesNodeShutdown(ctx, t, c, nodeToShutdown, startBackup)
		},
	})
	__antithesis_instrumentation__.Notify(45629)
	r.Add(registry.TestSpec{
		Name:            fmt.Sprintf("backup/nodeShutdown/coordinator/%s", backupNodeRestartSpec),
		Owner:           registry.OwnerBulkIO,
		Cluster:         backupNodeRestartSpec,
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45637)
			gatewayNode := 2
			nodeToShutdown := 2
			dest := loadBackupData(ctx, t, c)
			backupQuery := `BACKUP bank.bank TO 'nodelocal://1/` + dest + `' WITH DETACHED`
			startBackup := func(c cluster.Cluster, t test.Test) (jobID string, err error) {
				__antithesis_instrumentation__.Notify(45639)
				gatewayDB := c.Conn(ctx, t.L(), gatewayNode)
				defer gatewayDB.Close()

				err = gatewayDB.QueryRowContext(ctx, backupQuery).Scan(&jobID)
				return
			}
			__antithesis_instrumentation__.Notify(45638)

			jobSurvivesNodeShutdown(ctx, t, c, nodeToShutdown, startBackup)
		},
	})

}

func removeJobClaimsForNodes(
	ctx context.Context, t test.Test, db *gosql.DB, nodes option.NodeListOption, jobID jobspb.JobID,
) {
	__antithesis_instrumentation__.Notify(45640)
	if len(nodes) == 0 {
		__antithesis_instrumentation__.Notify(45643)
		return
	} else {
		__antithesis_instrumentation__.Notify(45644)
	}
	__antithesis_instrumentation__.Notify(45641)

	n := make([]string, 0)
	for _, node := range nodes {
		__antithesis_instrumentation__.Notify(45645)
		n = append(n, strconv.Itoa(node))
	}
	__antithesis_instrumentation__.Notify(45642)
	nodesStr := strings.Join(n, ",")

	removeClaimQuery := `
UPDATE system.jobs
   SET claim_session_id = NULL
WHERE claim_instance_id IN (%s)
AND id = $1
`
	_, err := db.ExecContext(ctx, fmt.Sprintf(removeClaimQuery, nodesStr), jobID)
	require.NoError(t, err)
}

func waitForJobToHaveStatus(
	ctx context.Context,
	t test.Test,
	db *gosql.DB,
	jobID jobspb.JobID,
	expectedStatus jobs.Status,
	nodesWithAdoptionDisabled option.NodeListOption,
) {
	__antithesis_instrumentation__.Notify(45646)
	if err := retry.ForDuration(time.Minute*2, func() error {
		__antithesis_instrumentation__.Notify(45647)

		removeJobClaimsForNodes(ctx, t, db, nodesWithAdoptionDisabled, jobID)

		var status string
		var payloadBytes []byte
		err := db.QueryRow(`SELECT status, payload FROM system.jobs WHERE id = $1`, jobID).Scan(&status, &payloadBytes)
		require.NoError(t, err)
		if jobs.Status(status) == jobs.StatusFailed {
			__antithesis_instrumentation__.Notify(45650)
			payload := &jobspb.Payload{}
			if err := protoutil.Unmarshal(payloadBytes, payload); err == nil {
				__antithesis_instrumentation__.Notify(45652)
				t.Fatalf("job failed: %s", payload.Error)
			} else {
				__antithesis_instrumentation__.Notify(45653)
			}
			__antithesis_instrumentation__.Notify(45651)
			t.Fatalf("job failed")
		} else {
			__antithesis_instrumentation__.Notify(45654)
		}
		__antithesis_instrumentation__.Notify(45648)
		if e, a := expectedStatus, jobs.Status(status); e != a {
			__antithesis_instrumentation__.Notify(45655)
			return errors.Errorf("expected job status %s, but got %s", e, a)
		} else {
			__antithesis_instrumentation__.Notify(45656)
		}
		__antithesis_instrumentation__.Notify(45649)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(45657)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(45658)
	}
}

func fingerprint(ctx context.Context, conn *gosql.DB, db, table string) (string, error) {
	__antithesis_instrumentation__.Notify(45659)
	var b strings.Builder

	query := fmt.Sprintf("SHOW EXPERIMENTAL_FINGERPRINTS FROM TABLE %s.%s", db, table)
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		__antithesis_instrumentation__.Notify(45662)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(45663)
	}
	__antithesis_instrumentation__.Notify(45660)
	defer rows.Close()
	for rows.Next() {
		__antithesis_instrumentation__.Notify(45664)
		var name, fp string
		if err := rows.Scan(&name, &fp); err != nil {
			__antithesis_instrumentation__.Notify(45666)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(45667)
		}
		__antithesis_instrumentation__.Notify(45665)
		fmt.Fprintf(&b, "%s: %s\n", name, fp)
	}
	__antithesis_instrumentation__.Notify(45661)

	return b.String(), rows.Err()
}

func registerBackupMixedVersion(r registry.Registry) {
	__antithesis_instrumentation__.Notify(45668)

	setShortJobIntervalsStep := func(node int) versionStep {
		__antithesis_instrumentation__.Notify(45677)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(45678)
			db := u.conn(ctx, t, node)
			_, err := db.ExecContext(ctx, `SET CLUSTER SETTING jobs.registry.interval.cancel = '1s'`)
			if err != nil {
				__antithesis_instrumentation__.Notify(45680)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45681)
			}
			__antithesis_instrumentation__.Notify(45679)

			_, err = db.ExecContext(ctx, `SET CLUSTER SETTING jobs.registry.interval.adopt = '1s'`)
			if err != nil {
				__antithesis_instrumentation__.Notify(45682)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45683)
			}
		}
	}
	__antithesis_instrumentation__.Notify(45669)

	loadBackupDataStep := func(c cluster.Cluster) versionStep {
		__antithesis_instrumentation__.Notify(45684)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(45685)
			rows := rows3GiB
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(45687)
				rows = 100
			} else {
				__antithesis_instrumentation__.Notify(45688)
			}
			__antithesis_instrumentation__.Notify(45686)
			runImportBankDataSplit(ctx, rows, 0, t, u.c)
		}
	}
	__antithesis_instrumentation__.Notify(45670)

	disableJobAdoptionStep := func(c cluster.Cluster, nodeIDs option.NodeListOption) versionStep {
		__antithesis_instrumentation__.Notify(45689)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(45690)
			for _, nodeID := range nodeIDs {
				__antithesis_instrumentation__.Notify(45692)
				result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(nodeID), "echo", "-n", "{store-dir}")
				if err != nil {
					__antithesis_instrumentation__.Notify(45694)
					t.L().Printf("Failed to retrieve store directory from node %d: %v\n", nodeID, err.Error())
				} else {
					__antithesis_instrumentation__.Notify(45695)
				}
				__antithesis_instrumentation__.Notify(45693)
				storeDirectory := result.Stdout
				disableJobAdoptionSentinelFilePath := filepath.Join(storeDirectory, jobs.PreventAdoptionFile)
				c.Run(ctx, nodeIDs, fmt.Sprintf("touch %s", disableJobAdoptionSentinelFilePath))

				testutils.SucceedsSoon(t, func() error {
					__antithesis_instrumentation__.Notify(45696)
					gatewayDB := c.Conn(ctx, t.L(), nodeID)
					defer gatewayDB.Close()

					row := gatewayDB.QueryRow(`SELECT count(*) FROM [SHOW JOBS] WHERE status = 'running'`)
					var count int
					require.NoError(t, row.Scan(&count))
					if count != 0 {
						__antithesis_instrumentation__.Notify(45698)
						return errors.Newf("node is still running %d jobs", count)
					} else {
						__antithesis_instrumentation__.Notify(45699)
					}
					__antithesis_instrumentation__.Notify(45697)
					return nil
				})
			}
			__antithesis_instrumentation__.Notify(45691)

			_, err := c.RunWithDetails(ctx, t.L(), nodeIDs, "export COCKROACH_JOB_ADOPTIONS_PER_PERIOD=0")
			require.NoError(t, err)
		}
	}
	__antithesis_instrumentation__.Notify(45671)

	enableJobAdoptionStep := func(c cluster.Cluster,
		nodeIDs option.NodeListOption) versionStep {
		__antithesis_instrumentation__.Notify(45700)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(45701)
			for _, nodeID := range nodeIDs {
				__antithesis_instrumentation__.Notify(45703)
				result, err := c.RunWithDetailsSingleNode(ctx, t.L(),
					c.Node(nodeID), "echo", "-n", "{store-dir}")
				if err != nil {
					__antithesis_instrumentation__.Notify(45705)
					t.L().Printf("Failed to retrieve store directory from node %d: %v\n", nodeID, err.Error())
				} else {
					__antithesis_instrumentation__.Notify(45706)
				}
				__antithesis_instrumentation__.Notify(45704)
				storeDirectory := result.Stdout
				disableJobAdoptionSentinelFilePath := filepath.Join(storeDirectory, jobs.PreventAdoptionFile)
				c.Run(ctx, nodeIDs, fmt.Sprintf("rm -f %s", disableJobAdoptionSentinelFilePath))
			}
			__antithesis_instrumentation__.Notify(45702)

			_, err := c.RunWithDetails(ctx, t.L(), nodeIDs, "export COCKROACH_JOB_ADOPTIONS_PER_PERIOD=10")
			require.NoError(t, err)
		}
	}
	__antithesis_instrumentation__.Notify(45672)

	planAndRunBackup := func(t test.Test, c cluster.Cluster, nodeToPlanBackup option.NodeListOption,
		nodesWithAdoptionDisabled option.NodeListOption, backupStmt string) versionStep {
		__antithesis_instrumentation__.Notify(45707)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(45708)
			gatewayDB := c.Conn(ctx, t.L(), nodeToPlanBackup[0])
			defer gatewayDB.Close()
			t.Status("Running: ", backupStmt)
			var jobID jobspb.JobID
			err := gatewayDB.QueryRow(backupStmt).Scan(&jobID)
			require.NoError(t, err)
			waitForJobToHaveStatus(ctx, t, gatewayDB, jobID, jobs.StatusSucceeded, nodesWithAdoptionDisabled)
		}
	}
	__antithesis_instrumentation__.Notify(45673)

	writeToBankStep := func(node int) versionStep {
		__antithesis_instrumentation__.Notify(45709)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(45710)
			db := u.conn(ctx, t, node)
			_, err := db.ExecContext(ctx, `UPSERT INTO bank.bank (id,balance) SELECT generate_series(1,100), random()*100;`)
			if err != nil {
				__antithesis_instrumentation__.Notify(45711)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45712)
			}
		}
	}
	__antithesis_instrumentation__.Notify(45674)

	saveFingerprintStep := func(node int, fingerprints map[string]string, key string) versionStep {
		__antithesis_instrumentation__.Notify(45713)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(45714)
			db := u.conn(ctx, t, node)
			f, err := fingerprint(ctx, db, "bank", "bank")
			if err != nil {
				__antithesis_instrumentation__.Notify(45716)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45717)
			}
			__antithesis_instrumentation__.Notify(45715)
			fingerprints[key] = f
		}
	}
	__antithesis_instrumentation__.Notify(45675)

	verifyBackupStep := func(
		node option.NodeListOption,
		backupLoc string,
		dbName, tableName, intoDB string,
		fingerprints map[string]string,
	) versionStep {
		__antithesis_instrumentation__.Notify(45718)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(45719)
			db := u.conn(ctx, t, node[0])

			_, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s`, intoDB))
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, fmt.Sprintf(`RESTORE TABLE %s.%s FROM LATEST IN '%s' WITH into_db = '%s'`,
				dbName, tableName, backupLoc, intoDB))
			require.NoError(t, err)

			restoredFingerPrint, err := fingerprint(ctx, db, intoDB, tableName)
			require.NoError(t, err)
			if fingerprints[backupLoc] != restoredFingerPrint {
				__antithesis_instrumentation__.Notify(45720)
				log.Infof(ctx, "original %s \n\n restored %s", fingerprints[backupLoc],
					restoredFingerPrint)
				t.Fatal("expected backup and restore fingerprints to match")
			} else {
				__antithesis_instrumentation__.Notify(45721)
			}
		}
	}
	__antithesis_instrumentation__.Notify(45676)

	r.Add(registry.TestSpec{
		Name:            "backup/mixed-version-basic",
		Owner:           registry.OwnerBulkIO,
		Cluster:         r.MakeClusterSpec(4),
		EncryptAtRandom: true,
		RequiresLicense: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45722)

			const mainVersion = ""
			roachNodes := c.All()
			upgradedNodes := c.Nodes(1, 2)
			oldNodes := c.Nodes(3, 4)
			predV, err := PredecessorVersion(*t.BuildVersion())
			require.NoError(t, err)
			c.Put(ctx, t.DeprecatedWorkload(), "./workload")

			fingerprints := make(map[string]string)
			u := newVersionUpgradeTest(c,
				uploadAndStartFromCheckpointFixture(roachNodes, predV),
				waitForUpgradeStep(roachNodes),
				preventAutoUpgradeStep(1),
				setShortJobIntervalsStep(1),
				loadBackupDataStep(c),

				binaryUpgradeStep(upgradedNodes, mainVersion),

				disableJobAdoptionStep(c, oldNodes),

				planAndRunBackup(t, c, oldNodes.RandNode(), oldNodes,
					`BACKUP TABLE bank.bank INTO 'nodelocal://1/plan-old-resume-new' WITH detached`),

				writeToBankStep(1),

				planAndRunBackup(t, c, oldNodes.RandNode(), oldNodes,
					`BACKUP TABLE bank.bank INTO LATEST IN 'nodelocal://1/plan-old-resume-new' WITH detached`),

				saveFingerprintStep(1, fingerprints, "nodelocal://1/plan-old-resume-new"),

				enableJobAdoptionStep(c, oldNodes),

				disableJobAdoptionStep(c, upgradedNodes),

				planAndRunBackup(t, c, upgradedNodes.RandNode(), upgradedNodes,
					`BACKUP TABLE bank.bank INTO 'nodelocal://1/plan-new-resume-old' WITH detached`),

				writeToBankStep(1),

				planAndRunBackup(t, c, upgradedNodes.RandNode(), upgradedNodes,
					`BACKUP TABLE bank.bank INTO LATEST IN 'nodelocal://1/plan-new-resume-old' WITH detached`),

				saveFingerprintStep(1, fingerprints, "nodelocal://1/plan-new-resume-old"),

				enableJobAdoptionStep(c, upgradedNodes),

				disableJobAdoptionStep(c, oldNodes),

				planAndRunBackup(t, c, upgradedNodes.RandNode(), oldNodes,
					`BACKUP TABLE bank.bank INTO 'nodelocal://1/new-node-full-backup' WITH detached`),

				writeToBankStep(1),

				enableJobAdoptionStep(c, oldNodes),
				disableJobAdoptionStep(c, upgradedNodes),

				planAndRunBackup(t, c, oldNodes.RandNode(), upgradedNodes,
					`BACKUP TABLE bank.bank INTO LATEST IN 'nodelocal://1/new-node-full-backup' WITH detached`),

				saveFingerprintStep(1, fingerprints, "nodelocal://1/new-node-full-backup"),

				planAndRunBackup(t, c, oldNodes.RandNode(), upgradedNodes,
					`BACKUP TABLE bank.bank INTO 'nodelocal://1/old-node-full-backup' WITH detached`),

				writeToBankStep(1),

				enableJobAdoptionStep(c, upgradedNodes),

				binaryUpgradeStep(oldNodes, mainVersion),
				allowAutoUpgradeStep(1),
				waitForUpgradeStep(roachNodes),

				planAndRunBackup(t, c, roachNodes.RandNode(), nil,
					`BACKUP TABLE bank.bank INTO LATEST IN 'nodelocal://1/old-node-full-backup' WITH detached`),

				saveFingerprintStep(1, fingerprints, "nodelocal://1/old-node-full-backup"),

				verifyBackupStep(roachNodes.RandNode(), "nodelocal://1/plan-old-resume-new",
					"bank", "bank", "bank1", fingerprints),
				verifyBackupStep(roachNodes.RandNode(), "nodelocal://1/plan-new-resume-old",
					"bank", "bank", "bank2", fingerprints),
				verifyBackupStep(roachNodes.RandNode(), "nodelocal://1/new-node-full-backup",
					"bank", "bank", "bank3", fingerprints),
				verifyBackupStep(roachNodes.RandNode(), "nodelocal://1/old-node-full-backup",
					"bank", "bank", "bank4", fingerprints),
			)
			u.run(ctx, t)
		},
	})
}

func initBulkJobPerfArtifacts(testName string, timeout time.Duration) (func(), *bytes.Buffer) {
	__antithesis_instrumentation__.Notify(45723)

	reg := histogram.NewRegistry(
		timeout,
		histogram.MockWorkloadName,
	)
	reg.GetHandle().Get(testName)

	bytesBuf := bytes.NewBuffer([]byte{})
	jsonEnc := json.NewEncoder(bytesBuf)
	tick := func() {
		__antithesis_instrumentation__.Notify(45725)
		reg.Tick(func(tick histogram.Tick) {
			__antithesis_instrumentation__.Notify(45726)
			_ = jsonEnc.Encode(tick.Snapshot())
		})
	}
	__antithesis_instrumentation__.Notify(45724)

	return tick, bytesBuf
}

func registerBackup(r registry.Registry) {
	__antithesis_instrumentation__.Notify(45727)

	backup2TBSpec := r.MakeClusterSpec(10)
	r.Add(registry.TestSpec{
		Name:            fmt.Sprintf("backup/2TB/%s", backup2TBSpec),
		Owner:           registry.OwnerBulkIO,
		Cluster:         backup2TBSpec,
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45730)
			rows := rows2TiB
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(45733)
				rows = 100
			} else {
				__antithesis_instrumentation__.Notify(45734)
			}
			__antithesis_instrumentation__.Notify(45731)
			dest := importBankData(ctx, rows, t, c)
			tick, perfBuf := initBulkJobPerfArtifacts("backup/2TB", 2*time.Hour)

			m := c.NewMonitor(ctx)
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(45735)
				t.Status(`running backup`)

				tick()
				c.Run(ctx, c.Node(1), `./cockroach sql --insecure -e "
				BACKUP bank.bank TO 'gs://cockroachdb-backup-testing/`+dest+`?AUTH=implicit'"`)
				tick()

				dest := filepath.Join(t.PerfArtifactsDir(), "stats.json")
				if err := c.RunE(ctx, c.Node(1), "mkdir -p "+filepath.Dir(dest)); err != nil {
					__antithesis_instrumentation__.Notify(45738)
					log.Errorf(ctx, "failed to create perf dir: %+v", err)
				} else {
					__antithesis_instrumentation__.Notify(45739)
				}
				__antithesis_instrumentation__.Notify(45736)
				if err := c.PutString(ctx, perfBuf.String(), dest, 0755, c.Node(1)); err != nil {
					__antithesis_instrumentation__.Notify(45740)
					log.Errorf(ctx, "failed to upload perf artifacts to node: %s", err.Error())
				} else {
					__antithesis_instrumentation__.Notify(45741)
				}
				__antithesis_instrumentation__.Notify(45737)
				return nil
			})
			__antithesis_instrumentation__.Notify(45732)
			m.Wait()
		},
	})
	__antithesis_instrumentation__.Notify(45728)

	KMSSpec := r.MakeClusterSpec(3)
	for _, item := range []struct {
		kmsProvider string
		machine     string
	}{
		{kmsProvider: "GCS", machine: spec.GCE},
		{kmsProvider: "AWS", machine: spec.AWS},
	} {
		__antithesis_instrumentation__.Notify(45742)
		item := item
		r.Add(registry.TestSpec{
			Name:            fmt.Sprintf("backup/KMS/%s/%s", item.kmsProvider, KMSSpec.String()),
			Owner:           registry.OwnerBulkIO,
			Cluster:         KMSSpec,
			EncryptAtRandom: true,
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(45743)
				if c.Spec().Cloud != item.machine {
					__antithesis_instrumentation__.Notify(45749)
					t.Skip("backupKMS roachtest is only configured to run on "+item.machine, "")
				} else {
					__antithesis_instrumentation__.Notify(45750)
				}
				__antithesis_instrumentation__.Notify(45744)

				rows := rows30GiB
				if c.IsLocal() {
					__antithesis_instrumentation__.Notify(45751)
					rows = 100
				} else {
					__antithesis_instrumentation__.Notify(45752)
				}
				__antithesis_instrumentation__.Notify(45745)
				dest := importBankData(ctx, rows, t, c)

				conn := c.Conn(ctx, t.L(), 1)
				m := c.NewMonitor(ctx)
				m.Go(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(45753)
					_, err := conn.ExecContext(ctx, `
					CREATE DATABASE restoreA;
					CREATE DATABASE restoreB;
				`)
					return err
				})
				__antithesis_instrumentation__.Notify(45746)
				m.Wait()
				var kmsURIA, kmsURIB string
				var err error
				backupPath := fmt.Sprintf("nodelocal://1/kmsbackup/%s/%s", item.kmsProvider, dest)

				m = c.NewMonitor(ctx)
				m.Go(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(45754)
					switch item.kmsProvider {
					case "AWS":
						__antithesis_instrumentation__.Notify(45756)
						t.Status(`running encrypted backup with AWS KMS`)
						kmsURIA, err = getAWSKMSURI(KMSRegionAEnvVar, KMSKeyARNAEnvVar)
						if err != nil {
							__antithesis_instrumentation__.Notify(45761)
							return err
						} else {
							__antithesis_instrumentation__.Notify(45762)
						}
						__antithesis_instrumentation__.Notify(45757)

						kmsURIB, err = getAWSKMSURI(KMSRegionBEnvVar, KMSKeyARNBEnvVar)
						if err != nil {
							__antithesis_instrumentation__.Notify(45763)
							return err
						} else {
							__antithesis_instrumentation__.Notify(45764)
						}
					case "GCS":
						__antithesis_instrumentation__.Notify(45758)
						t.Status(`running encrypted backup with GCS KMS`)
						kmsURIA, err = getGCSKMSURI(KMSKeyNameAEnvVar)
						if err != nil {
							__antithesis_instrumentation__.Notify(45765)
							return err
						} else {
							__antithesis_instrumentation__.Notify(45766)
						}
						__antithesis_instrumentation__.Notify(45759)

						kmsURIB, err = getGCSKMSURI(KMSKeyNameBEnvVar)
						if err != nil {
							__antithesis_instrumentation__.Notify(45767)
							return err
						} else {
							__antithesis_instrumentation__.Notify(45768)
						}
					default:
						__antithesis_instrumentation__.Notify(45760)
					}
					__antithesis_instrumentation__.Notify(45755)

					kmsOptions := fmt.Sprintf("KMS=('%s', '%s')", kmsURIA, kmsURIB)
					_, err := conn.ExecContext(ctx, `BACKUP bank.bank TO '`+backupPath+`' WITH `+kmsOptions)
					return err
				})
				__antithesis_instrumentation__.Notify(45747)
				m.Wait()

				m = c.NewMonitor(ctx)
				m.Go(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(45769)
					t.Status(`restore using KMSURIA`)
					if _, err := conn.ExecContext(ctx,
						`RESTORE bank.bank FROM $1 WITH into_db=restoreA, kms=$2`,
						backupPath, kmsURIA,
					); err != nil {
						__antithesis_instrumentation__.Notify(45778)
						return err
					} else {
						__antithesis_instrumentation__.Notify(45779)
					}
					__antithesis_instrumentation__.Notify(45770)

					t.Status(`restore using KMSURIB`)
					if _, err := conn.ExecContext(ctx,
						`RESTORE bank.bank FROM $1 WITH into_db=restoreB, kms=$2`,
						backupPath, kmsURIB,
					); err != nil {
						__antithesis_instrumentation__.Notify(45780)
						return err
					} else {
						__antithesis_instrumentation__.Notify(45781)
					}
					__antithesis_instrumentation__.Notify(45771)

					t.Status(`fingerprint`)
					fingerprint := func(db string) (string, error) {
						__antithesis_instrumentation__.Notify(45782)
						var b strings.Builder

						query := fmt.Sprintf("SHOW EXPERIMENTAL_FINGERPRINTS FROM TABLE %s.%s", db, "bank")
						rows, err := conn.QueryContext(ctx, query)
						if err != nil {
							__antithesis_instrumentation__.Notify(45785)
							return "", err
						} else {
							__antithesis_instrumentation__.Notify(45786)
						}
						__antithesis_instrumentation__.Notify(45783)
						defer rows.Close()
						for rows.Next() {
							__antithesis_instrumentation__.Notify(45787)
							var name, fp string
							if err := rows.Scan(&name, &fp); err != nil {
								__antithesis_instrumentation__.Notify(45789)
								return "", err
							} else {
								__antithesis_instrumentation__.Notify(45790)
							}
							__antithesis_instrumentation__.Notify(45788)
							fmt.Fprintf(&b, "%s: %s\n", name, fp)
						}
						__antithesis_instrumentation__.Notify(45784)

						return b.String(), rows.Err()
					}
					__antithesis_instrumentation__.Notify(45772)

					originalBank, err := fingerprint("bank")
					if err != nil {
						__antithesis_instrumentation__.Notify(45791)
						return err
					} else {
						__antithesis_instrumentation__.Notify(45792)
					}
					__antithesis_instrumentation__.Notify(45773)
					restoreA, err := fingerprint("restoreA")
					if err != nil {
						__antithesis_instrumentation__.Notify(45793)
						return err
					} else {
						__antithesis_instrumentation__.Notify(45794)
					}
					__antithesis_instrumentation__.Notify(45774)
					restoreB, err := fingerprint("restoreB")
					if err != nil {
						__antithesis_instrumentation__.Notify(45795)
						return err
					} else {
						__antithesis_instrumentation__.Notify(45796)
					}
					__antithesis_instrumentation__.Notify(45775)

					if originalBank != restoreA {
						__antithesis_instrumentation__.Notify(45797)
						return errors.Errorf("got %s, expected %s while comparing restoreA with originalBank", restoreA, originalBank)
					} else {
						__antithesis_instrumentation__.Notify(45798)
					}
					__antithesis_instrumentation__.Notify(45776)
					if originalBank != restoreB {
						__antithesis_instrumentation__.Notify(45799)
						return errors.Errorf("got %s, expected %s while comparing restoreB with originalBank", restoreB, originalBank)
					} else {
						__antithesis_instrumentation__.Notify(45800)
					}
					__antithesis_instrumentation__.Notify(45777)
					return nil
				})
				__antithesis_instrumentation__.Notify(45748)
				m.Wait()
			},
		})
	}
	__antithesis_instrumentation__.Notify(45729)

	r.Add(registry.TestSpec{
		Name:            `backupTPCC`,
		Owner:           registry.OwnerBulkIO,
		Cluster:         r.MakeClusterSpec(3),
		Timeout:         1 * time.Hour,
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45801)
			c.Put(ctx, t.Cockroach(), "./cockroach")
			c.Put(ctx, t.DeprecatedWorkload(), "./workload")
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
			conn := c.Conn(ctx, t.L(), 1)

			duration := 5 * time.Minute
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(45812)
				duration = 5 * time.Second
			} else {
				__antithesis_instrumentation__.Notify(45813)
			}
			__antithesis_instrumentation__.Notify(45802)
			warehouses := 10

			backupDir := "gs://cockroachdb-backup-testing/" + c.Name() + "?AUTH=implicit"

			if t.BuildVersion().AtLeast(version.MustParse(`v20.1.0-0`)) {
				__antithesis_instrumentation__.Notify(45814)
				backupDir = "nodelocal://1/" + c.Name()
			} else {
				__antithesis_instrumentation__.Notify(45815)
			}
			__antithesis_instrumentation__.Notify(45803)
			fullDir := backupDir + "/full"
			incDir := backupDir + "/inc"

			t.Status(`workload initialization`)
			cmd := []string{fmt.Sprintf(
				"./workload init tpcc --warehouses=%d {pgurl:1-%d}",
				warehouses, c.Spec().NodeCount,
			)}
			if !t.BuildVersion().AtLeast(version.MustParse("v20.2.0")) {
				__antithesis_instrumentation__.Notify(45816)
				cmd = append(cmd, "--deprecated-fk-indexes")
			} else {
				__antithesis_instrumentation__.Notify(45817)
			}
			__antithesis_instrumentation__.Notify(45804)
			c.Run(ctx, c.Node(1), cmd...)

			m := c.NewMonitor(ctx)
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(45818)
				_, err := conn.ExecContext(ctx, `
					CREATE DATABASE restore_full;
					CREATE DATABASE restore_inc;
				`)
				return err
			})
			__antithesis_instrumentation__.Notify(45805)
			m.Wait()

			t.Status(`run tpcc`)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			cmdDone := make(chan error)
			go func() {
				__antithesis_instrumentation__.Notify(45819)
				cmd := fmt.Sprintf(
					"./workload run tpcc --warehouses=%d {pgurl:1-%d}",
					warehouses, c.Spec().NodeCount,
				)

				cmdDone <- c.RunE(ctx, c.Node(1), cmd)
			}()
			__antithesis_instrumentation__.Notify(45806)

			select {
			case <-time.After(duration):
				__antithesis_instrumentation__.Notify(45820)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(45821)
				return
			}
			__antithesis_instrumentation__.Notify(45807)

			tFull := fmt.Sprint(timeutil.Now().Add(time.Second * -2).UnixNano())
			m = c.NewMonitor(ctx)
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(45822)
				t.Status(`full backup`)
				_, err := conn.ExecContext(ctx,
					`BACKUP tpcc.* TO $1 AS OF SYSTEM TIME `+tFull,
					fullDir,
				)
				return err
			})
			__antithesis_instrumentation__.Notify(45808)
			m.Wait()

			t.Status(`continue tpcc`)
			select {
			case <-time.After(duration):
				__antithesis_instrumentation__.Notify(45823)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(45824)
				return
			}
			__antithesis_instrumentation__.Notify(45809)

			tInc := fmt.Sprint(timeutil.Now().Add(time.Second * -2).UnixNano())
			m = c.NewMonitor(ctx)
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(45825)
				t.Status(`incremental backup`)
				_, err := conn.ExecContext(ctx,
					`BACKUP tpcc.* TO $1 AS OF SYSTEM TIME `+tInc+` INCREMENTAL FROM $2`,
					incDir,
					fullDir,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(45827)
					return err
				} else {
					__antithesis_instrumentation__.Notify(45828)
				}
				__antithesis_instrumentation__.Notify(45826)

				select {
				case err := <-cmdDone:
					__antithesis_instrumentation__.Notify(45829)

					return err
				default:
					__antithesis_instrumentation__.Notify(45830)
					return nil
				}
			})
			__antithesis_instrumentation__.Notify(45810)
			m.Wait()

			m = c.NewMonitor(ctx)
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(45831)
				t.Status(`restore full`)
				if _, err := conn.ExecContext(ctx,
					`RESTORE tpcc.* FROM $1 WITH into_db='restore_full'`,
					fullDir,
				); err != nil {
					__antithesis_instrumentation__.Notify(45841)
					return err
				} else {
					__antithesis_instrumentation__.Notify(45842)
				}
				__antithesis_instrumentation__.Notify(45832)

				t.Status(`restore incremental`)
				if _, err := conn.ExecContext(ctx,
					`RESTORE tpcc.* FROM $1, $2 WITH into_db='restore_inc'`,
					fullDir,
					incDir,
				); err != nil {
					__antithesis_instrumentation__.Notify(45843)
					return err
				} else {
					__antithesis_instrumentation__.Notify(45844)
				}
				__antithesis_instrumentation__.Notify(45833)

				t.Status(`fingerprint`)

				fingerprint := func(db string, asof string) (string, error) {
					__antithesis_instrumentation__.Notify(45845)
					var b strings.Builder

					var tables []string
					rows, err := conn.QueryContext(
						ctx,
						fmt.Sprintf("SELECT table_name FROM [SHOW TABLES FROM %s] ORDER BY table_name", db),
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(45849)
						return "", err
					} else {
						__antithesis_instrumentation__.Notify(45850)
					}
					__antithesis_instrumentation__.Notify(45846)
					defer rows.Close()
					for rows.Next() {
						__antithesis_instrumentation__.Notify(45851)
						var name string
						if err := rows.Scan(&name); err != nil {
							__antithesis_instrumentation__.Notify(45853)
							return "", err
						} else {
							__antithesis_instrumentation__.Notify(45854)
						}
						__antithesis_instrumentation__.Notify(45852)
						tables = append(tables, name)
					}
					__antithesis_instrumentation__.Notify(45847)

					for _, table := range tables {
						__antithesis_instrumentation__.Notify(45855)
						fmt.Fprintf(&b, "table %s\n", table)
						query := fmt.Sprintf("SHOW EXPERIMENTAL_FINGERPRINTS FROM TABLE %s.%s", db, table)
						if asof != "" {
							__antithesis_instrumentation__.Notify(45858)
							query = fmt.Sprintf("SELECT * FROM [%s] AS OF SYSTEM TIME %s", query, asof)
						} else {
							__antithesis_instrumentation__.Notify(45859)
						}
						__antithesis_instrumentation__.Notify(45856)
						rows, err = conn.QueryContext(ctx, query)
						if err != nil {
							__antithesis_instrumentation__.Notify(45860)
							return "", err
						} else {
							__antithesis_instrumentation__.Notify(45861)
						}
						__antithesis_instrumentation__.Notify(45857)
						defer rows.Close()
						for rows.Next() {
							__antithesis_instrumentation__.Notify(45862)
							var name, fp string
							if err := rows.Scan(&name, &fp); err != nil {
								__antithesis_instrumentation__.Notify(45864)
								return "", err
							} else {
								__antithesis_instrumentation__.Notify(45865)
							}
							__antithesis_instrumentation__.Notify(45863)
							fmt.Fprintf(&b, "%s: %s\n", name, fp)
						}
					}
					__antithesis_instrumentation__.Notify(45848)

					return b.String(), rows.Err()
				}
				__antithesis_instrumentation__.Notify(45834)

				tpccFull, err := fingerprint("tpcc", tFull)
				if err != nil {
					__antithesis_instrumentation__.Notify(45866)
					return err
				} else {
					__antithesis_instrumentation__.Notify(45867)
				}
				__antithesis_instrumentation__.Notify(45835)
				tpccInc, err := fingerprint("tpcc", tInc)
				if err != nil {
					__antithesis_instrumentation__.Notify(45868)
					return err
				} else {
					__antithesis_instrumentation__.Notify(45869)
				}
				__antithesis_instrumentation__.Notify(45836)
				restoreFull, err := fingerprint("restore_full", "")
				if err != nil {
					__antithesis_instrumentation__.Notify(45870)
					return err
				} else {
					__antithesis_instrumentation__.Notify(45871)
				}
				__antithesis_instrumentation__.Notify(45837)
				restoreInc, err := fingerprint("restore_inc", "")
				if err != nil {
					__antithesis_instrumentation__.Notify(45872)
					return err
				} else {
					__antithesis_instrumentation__.Notify(45873)
				}
				__antithesis_instrumentation__.Notify(45838)

				if tpccFull != restoreFull {
					__antithesis_instrumentation__.Notify(45874)
					return errors.Errorf("got %s, expected %s", restoreFull, tpccFull)
				} else {
					__antithesis_instrumentation__.Notify(45875)
				}
				__antithesis_instrumentation__.Notify(45839)
				if tpccInc != restoreInc {
					__antithesis_instrumentation__.Notify(45876)
					return errors.Errorf("got %s, expected %s", restoreInc, tpccInc)
				} else {
					__antithesis_instrumentation__.Notify(45877)
				}
				__antithesis_instrumentation__.Notify(45840)

				return nil
			})
			__antithesis_instrumentation__.Notify(45811)
			m.Wait()
		},
	})

}

func getAWSKMSURI(regionEnvVariable, keyIDEnvVariable string) (string, error) {
	__antithesis_instrumentation__.Notify(45878)
	q := make(url.Values)
	expect := map[string]string{
		"AWS_ACCESS_KEY_ID":     amazon.AWSAccessKeyParam,
		"AWS_SECRET_ACCESS_KEY": amazon.AWSSecretParam,
		regionEnvVariable:       amazon.KMSRegionParam,
	}
	for env, param := range expect {
		__antithesis_instrumentation__.Notify(45881)
		v := os.Getenv(env)
		if v == "" {
			__antithesis_instrumentation__.Notify(45883)
			return "", errors.Newf("env variable %s must be present to run the KMS test", env)
		} else {
			__antithesis_instrumentation__.Notify(45884)
		}
		__antithesis_instrumentation__.Notify(45882)
		q.Add(param, v)
	}
	__antithesis_instrumentation__.Notify(45879)

	keyARN := os.Getenv(keyIDEnvVariable)
	if keyARN == "" {
		__antithesis_instrumentation__.Notify(45885)
		return "", errors.Newf("env variable %s must be present to run the KMS test", keyIDEnvVariable)
	} else {
		__antithesis_instrumentation__.Notify(45886)
	}
	__antithesis_instrumentation__.Notify(45880)

	q.Add(cloudstorage.AuthParam, cloudstorage.AuthParamSpecified)
	correctURI := fmt.Sprintf("aws:///%s?%s", keyARN, q.Encode())

	return correctURI, nil
}

func getGCSKMSURI(keyIDEnvVariable string) (string, error) {
	__antithesis_instrumentation__.Notify(45887)
	q := make(url.Values)
	expect := map[string]string{
		KMSGCSCredentials: gcp.CredentialsParam,
	}
	for env, param := range expect {
		__antithesis_instrumentation__.Notify(45890)
		v := os.Getenv(env)
		if v == "" {
			__antithesis_instrumentation__.Notify(45892)
			return "", errors.Newf("env variable %s must be present to run the KMS test", env)
		} else {
			__antithesis_instrumentation__.Notify(45893)
		}
		__antithesis_instrumentation__.Notify(45891)

		q.Add(param, base64.StdEncoding.EncodeToString([]byte(v)))
	}
	__antithesis_instrumentation__.Notify(45888)

	keyID := os.Getenv(keyIDEnvVariable)
	if keyID == "" {
		__antithesis_instrumentation__.Notify(45894)
		return "", errors.Newf("", "%s env var must be set", keyIDEnvVariable)
	} else {
		__antithesis_instrumentation__.Notify(45895)
	}
	__antithesis_instrumentation__.Notify(45889)

	q.Set(cloudstorage.AuthParam, cloudstorage.AuthParamSpecified)
	correctURI := fmt.Sprintf("gs:///%s?%s", keyID, q.Encode())

	return correctURI, nil
}
