package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	roachpb "github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/desctestutils"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/testcluster"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/workload/bank"
	"github.com/cockroachdb/cockroach/pkg/workload/workloadsql"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/require"
)

const (
	singleNode                  = 1
	multiNode                   = 3
	backupRestoreDefaultRanges  = 10
	backupRestoreRowPayloadSize = 100
	localFoo                    = "nodelocal://0/foo"
)

func backupRestoreTestSetupWithParams(
	t testing.TB,
	clusterSize int,
	numAccounts int,
	init func(tc *testcluster.TestCluster),
	params base.TestClusterArgs,
) (tc *testcluster.TestCluster, sqlDB *sqlutils.SQLRunner, tempDir string, cleanup func()) {
	__antithesis_instrumentation__.Notify(13563)
	ctx := logtags.AddTag(context.Background(), "backup-restore-test-setup", nil)

	dir, dirCleanupFn := testutils.TempDir(t)
	params.ServerArgs.ExternalIODir = dir
	params.ServerArgs.UseDatabase = "data"
	if len(params.ServerArgsPerNode) > 0 {
		__antithesis_instrumentation__.Notify(13569)
		for i := range params.ServerArgsPerNode {
			__antithesis_instrumentation__.Notify(13570)
			param := params.ServerArgsPerNode[i]
			param.ExternalIODir = dir
			param.UseDatabase = "data"
			params.ServerArgsPerNode[i] = param
		}
	} else {
		__antithesis_instrumentation__.Notify(13571)
	}
	__antithesis_instrumentation__.Notify(13564)

	tc = testcluster.StartTestCluster(t, clusterSize, params)
	init(tc)

	const payloadSize = 100
	splits := 10
	if numAccounts == 0 {
		__antithesis_instrumentation__.Notify(13572)
		splits = 0
	} else {
		__antithesis_instrumentation__.Notify(13573)
	}
	__antithesis_instrumentation__.Notify(13565)
	bankData := bank.FromConfig(numAccounts, numAccounts, payloadSize, splits)

	sqlDB = sqlutils.MakeSQLRunner(tc.Conns[0])

	sqlDB.Exec(t, `SET CLUSTER SETTING bulkio.backup.merge_file_buffer_size = '16MiB'`)

	sqlDB.Exec(t, `CREATE DATABASE data`)
	l := workloadsql.InsertsDataLoader{BatchSize: 1000, Concurrency: 4}
	if _, err := workloadsql.Setup(ctx, sqlDB.DB.(*gosql.DB), bankData, l); err != nil {
		__antithesis_instrumentation__.Notify(13574)
		t.Fatalf("%+v", err)
	} else {
		__antithesis_instrumentation__.Notify(13575)
	}
	__antithesis_instrumentation__.Notify(13566)

	if err := tc.WaitForFullReplication(); err != nil {
		__antithesis_instrumentation__.Notify(13576)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(13577)
	}
	__antithesis_instrumentation__.Notify(13567)

	cleanupFn := func() {
		__antithesis_instrumentation__.Notify(13578)
		tc.Stopper().Stop(ctx)
		dirCleanupFn()
	}
	__antithesis_instrumentation__.Notify(13568)

	return tc, sqlDB, dir, cleanupFn
}

func backupDestinationTestSetup(
	t testing.TB, clusterSize int, numAccounts int, init func(*testcluster.TestCluster),
) (tc *testcluster.TestCluster, sqlDB *sqlutils.SQLRunner, tempDir string, cleanup func()) {
	__antithesis_instrumentation__.Notify(13579)
	return backupRestoreTestSetupWithParams(t, clusterSize, numAccounts, init, base.TestClusterArgs{})
}

func backupRestoreTestSetup(
	t testing.TB, clusterSize int, numAccounts int, init func(*testcluster.TestCluster),
) (tc *testcluster.TestCluster, sqlDB *sqlutils.SQLRunner, tempDir string, cleanup func()) {
	__antithesis_instrumentation__.Notify(13580)
	return backupRestoreTestSetupWithParams(t, clusterSize, numAccounts, init, base.TestClusterArgs{})
}

func backupRestoreTestSetupEmpty(
	t testing.TB,
	clusterSize int,
	tempDir string,
	init func(*testcluster.TestCluster),
	params base.TestClusterArgs,
) (tc *testcluster.TestCluster, sqlDB *sqlutils.SQLRunner, cleanup func()) {
	__antithesis_instrumentation__.Notify(13581)
	return backupRestoreTestSetupEmptyWithParams(t, clusterSize, tempDir, init, params)
}

func verifyBackupRestoreStatementResult(
	t *testing.T, sqlDB *sqlutils.SQLRunner, query string, args ...interface{},
) error {
	__antithesis_instrumentation__.Notify(13582)
	t.Helper()
	rows := sqlDB.Query(t, query, args...)

	columns, err := rows.Columns()
	if err != nil {
		__antithesis_instrumentation__.Notify(13589)
		return err
	} else {
		__antithesis_instrumentation__.Notify(13590)
	}
	__antithesis_instrumentation__.Notify(13583)
	if e, a := columns, []string{
		"job_id", "status", "fraction_completed", "rows", "index_entries", "bytes",
	}; !reflect.DeepEqual(e, a) {
		__antithesis_instrumentation__.Notify(13591)
		return errors.Errorf("unexpected columns:\n%s", strings.Join(pretty.Diff(e, a), "\n"))
	} else {
		__antithesis_instrumentation__.Notify(13592)
	}
	__antithesis_instrumentation__.Notify(13584)

	type job struct {
		id                int64
		status            string
		fractionCompleted float32
	}

	var expectedJob job
	var actualJob job
	var unused int64

	if !rows.Next() {
		__antithesis_instrumentation__.Notify(13593)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(13595)
			return err
		} else {
			__antithesis_instrumentation__.Notify(13596)
		}
		__antithesis_instrumentation__.Notify(13594)
		return errors.New("zero rows in result")
	} else {
		__antithesis_instrumentation__.Notify(13597)
	}
	__antithesis_instrumentation__.Notify(13585)
	if err := rows.Scan(
		&actualJob.id, &actualJob.status, &actualJob.fractionCompleted, &unused, &unused, &unused,
	); err != nil {
		__antithesis_instrumentation__.Notify(13598)
		return err
	} else {
		__antithesis_instrumentation__.Notify(13599)
	}
	__antithesis_instrumentation__.Notify(13586)
	if rows.Next() {
		__antithesis_instrumentation__.Notify(13600)
		return errors.New("more than one row in result")
	} else {
		__antithesis_instrumentation__.Notify(13601)
	}
	__antithesis_instrumentation__.Notify(13587)

	sqlDB.QueryRow(t,
		`SELECT job_id, status, fraction_completed FROM crdb_internal.jobs WHERE job_id = $1`, actualJob.id,
	).Scan(
		&expectedJob.id, &expectedJob.status, &expectedJob.fractionCompleted,
	)

	if e, a := expectedJob, actualJob; !reflect.DeepEqual(e, a) {
		__antithesis_instrumentation__.Notify(13602)
		return errors.Errorf("result does not match system.jobs:\n%s",
			strings.Join(pretty.Diff(e, a), "\n"))
	} else {
		__antithesis_instrumentation__.Notify(13603)
	}
	__antithesis_instrumentation__.Notify(13588)

	return nil
}

func backupRestoreTestSetupEmptyWithParams(
	t testing.TB,
	clusterSize int,
	dir string,
	init func(tc *testcluster.TestCluster),
	params base.TestClusterArgs,
) (tc *testcluster.TestCluster, sqlDB *sqlutils.SQLRunner, cleanup func()) {
	__antithesis_instrumentation__.Notify(13604)
	ctx := logtags.AddTag(context.Background(), "backup-restore-test-setup-empty", nil)

	params.ServerArgs.ExternalIODir = dir
	if len(params.ServerArgsPerNode) > 0 {
		__antithesis_instrumentation__.Notify(13607)
		for i := range params.ServerArgsPerNode {
			__antithesis_instrumentation__.Notify(13608)
			param := params.ServerArgsPerNode[i]
			param.ExternalIODir = dir
			params.ServerArgsPerNode[i] = param
		}
	} else {
		__antithesis_instrumentation__.Notify(13609)
	}
	__antithesis_instrumentation__.Notify(13605)
	tc = testcluster.StartTestCluster(t, clusterSize, params)
	init(tc)

	sqlDB = sqlutils.MakeSQLRunner(tc.Conns[0])

	cleanupFn := func() {
		__antithesis_instrumentation__.Notify(13610)
		tc.Stopper().Stop(ctx)
	}
	__antithesis_instrumentation__.Notify(13606)

	return tc, sqlDB, cleanupFn
}

func createEmptyCluster(
	t testing.TB, clusterSize int,
) (sqlDB *sqlutils.SQLRunner, tempDir string, cleanup func()) {
	__antithesis_instrumentation__.Notify(13611)
	ctx := context.Background()

	dir, dirCleanupFn := testutils.TempDir(t)
	params := base.TestClusterArgs{}
	params.ServerArgs.ExternalIODir = dir
	tc := testcluster.StartTestCluster(t, clusterSize, params)

	sqlDB = sqlutils.MakeSQLRunner(tc.Conns[0])

	cleanupFn := func() {
		__antithesis_instrumentation__.Notify(13613)
		tc.Stopper().Stop(ctx)
		dirCleanupFn()
	}
	__antithesis_instrumentation__.Notify(13612)

	return sqlDB, dir, cleanupFn
}

func getStatsQuery(tableName string) string {
	__antithesis_instrumentation__.Notify(13614)
	return fmt.Sprintf(`SELECT
	  statistics_name,
	  column_names,
	  row_count,
	  distinct_count,
	  null_count
	FROM [SHOW STATISTICS FOR TABLE %s]`, tableName)
}

func injectStats(
	t *testing.T, sqlDB *sqlutils.SQLRunner, tableName string, columnName string,
) [][]string {
	__antithesis_instrumentation__.Notify(13615)
	return injectStatsWithRowCount(t, sqlDB, tableName, columnName, 100)
}

func injectStatsWithRowCount(
	t *testing.T, sqlDB *sqlutils.SQLRunner, tableName string, columnName string, rowCount int,
) [][]string {
	__antithesis_instrumentation__.Notify(13616)
	sqlDB.Exec(t, fmt.Sprintf(`ALTER TABLE %s INJECT STATISTICS '[
	{
		"columns": ["%s"],
		"created_at": "2018-01-01 1:00:00.00000+00:00",
		"row_count": %d,
		"distinct_count": %d
	}
	]'`, tableName, columnName, rowCount, rowCount))
	return sqlDB.QueryStr(t, getStatsQuery(tableName))
}

func makeInsecureHTTPServer(t *testing.T) (string, func()) {
	__antithesis_instrumentation__.Notify(13617)
	t.Helper()

	const badHeadResponse = "bad-head-response"

	tmp, dirCleanup := testutils.TempDir(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		__antithesis_instrumentation__.Notify(13621)
		localfile := filepath.Join(tmp, filepath.Base(r.URL.Path))
		switch r.Method {
		case "PUT":
			__antithesis_instrumentation__.Notify(13622)
			f, err := os.Create(localfile)
			if err != nil {
				__antithesis_instrumentation__.Notify(13630)
				http.Error(w, err.Error(), 500)
				return
			} else {
				__antithesis_instrumentation__.Notify(13631)
			}
			__antithesis_instrumentation__.Notify(13623)
			defer f.Close()
			if _, err := io.Copy(f, r.Body); err != nil {
				__antithesis_instrumentation__.Notify(13632)
				http.Error(w, err.Error(), 500)
				return
			} else {
				__antithesis_instrumentation__.Notify(13633)
			}
			__antithesis_instrumentation__.Notify(13624)
			w.WriteHeader(201)
		case "GET", "HEAD":
			__antithesis_instrumentation__.Notify(13625)
			if filepath.Base(localfile) == badHeadResponse {
				__antithesis_instrumentation__.Notify(13634)
				http.Error(w, "HEAD not implemented", 500)
				return
			} else {
				__antithesis_instrumentation__.Notify(13635)
			}
			__antithesis_instrumentation__.Notify(13626)
			http.ServeFile(w, r, localfile)
		case "DELETE":
			__antithesis_instrumentation__.Notify(13627)
			if err := os.Remove(localfile); err != nil {
				__antithesis_instrumentation__.Notify(13636)
				http.Error(w, err.Error(), 500)
				return
			} else {
				__antithesis_instrumentation__.Notify(13637)
			}
			__antithesis_instrumentation__.Notify(13628)
			w.WriteHeader(204)
		default:
			__antithesis_instrumentation__.Notify(13629)
			http.Error(w, "unsupported method "+r.Method, 400)
		}
	}))
	__antithesis_instrumentation__.Notify(13618)

	cleanup := func() {
		__antithesis_instrumentation__.Notify(13638)
		srv.Close()
		dirCleanup()
	}
	__antithesis_instrumentation__.Notify(13619)

	t.Logf("Mock HTTP Storage %q", srv.URL)
	uri, err := url.Parse(srv.URL)
	if err != nil {
		__antithesis_instrumentation__.Notify(13639)
		srv.Close()
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(13640)
	}
	__antithesis_instrumentation__.Notify(13620)
	uri.Path = filepath.Join(uri.Path, "testing")
	return uri.String(), cleanup
}

type thresholdBlocker struct {
	threshold        int
	reachedThreshold chan struct{}
	canProceed       chan struct{}
}

func (t thresholdBlocker) maybeBlock(count int) {
	__antithesis_instrumentation__.Notify(13641)
	if count == t.threshold {
		__antithesis_instrumentation__.Notify(13642)
		close(t.reachedThreshold)
		<-t.canProceed
	} else {
		__antithesis_instrumentation__.Notify(13643)
	}
}

func (t thresholdBlocker) waitUntilBlocked() {
	__antithesis_instrumentation__.Notify(13644)
	<-t.reachedThreshold
}

func (t thresholdBlocker) allowToProceed() {
	__antithesis_instrumentation__.Notify(13645)
	close(t.canProceed)
}

func makeThresholdBlocker(threshold int) thresholdBlocker {
	__antithesis_instrumentation__.Notify(13646)
	return thresholdBlocker{
		threshold:        threshold,
		reachedThreshold: make(chan struct{}),
		canProceed:       make(chan struct{}),
	}
}

func getSpansFromManifest(ctx context.Context, t *testing.T, backupPath string) roachpb.Spans {
	__antithesis_instrumentation__.Notify(13647)
	backupManifestBytes, err := ioutil.ReadFile(backupPath + "/" + backupManifestName)
	require.NoError(t, err)
	var backupManifest BackupManifest
	decompressedBytes, err := decompressData(ctx, nil, backupManifestBytes)
	require.NoError(t, err)
	require.NoError(t, protoutil.Unmarshal(decompressedBytes, &backupManifest))
	spans := make([]roachpb.Span, 0, len(backupManifest.Files))
	for _, file := range backupManifest.Files {
		__antithesis_instrumentation__.Notify(13649)
		spans = append(spans, file.Span)
	}
	__antithesis_instrumentation__.Notify(13648)
	mergedSpans, _ := roachpb.MergeSpans(&spans)
	return mergedSpans
}

func getKVCount(ctx context.Context, kvDB *kv.DB, dbName, tableName string) (int, error) {
	__antithesis_instrumentation__.Notify(13650)
	tableDesc := desctestutils.TestingGetPublicTableDescriptor(kvDB, keys.SystemSQLCodec, dbName, tableName)
	tablePrefix := keys.SystemSQLCodec.TablePrefix(uint32(tableDesc.GetID()))
	tableEnd := tablePrefix.PrefixEnd()
	kvs, err := kvDB.Scan(ctx, tablePrefix, tableEnd, 0)
	return len(kvs), err
}

func uriFmtStringAndArgs(uris []string) (string, []interface{}) {
	__antithesis_instrumentation__.Notify(13651)
	urisForFormat := make([]interface{}, len(uris))
	var fmtString strings.Builder
	if len(uris) > 1 {
		__antithesis_instrumentation__.Notify(13655)
		fmtString.WriteString("(")
	} else {
		__antithesis_instrumentation__.Notify(13656)
	}
	__antithesis_instrumentation__.Notify(13652)
	for i, uri := range uris {
		__antithesis_instrumentation__.Notify(13657)
		if i > 0 {
			__antithesis_instrumentation__.Notify(13659)
			fmtString.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(13660)
		}
		__antithesis_instrumentation__.Notify(13658)
		fmtString.WriteString(fmt.Sprintf("$%d", i+1))
		urisForFormat[i] = uri
	}
	__antithesis_instrumentation__.Notify(13653)
	if len(uris) > 1 {
		__antithesis_instrumentation__.Notify(13661)
		fmtString.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(13662)
	}
	__antithesis_instrumentation__.Notify(13654)
	return fmtString.String(), urisForFormat
}

func waitForTableSplit(t *testing.T, conn *gosql.DB, tableName, dbName string) {
	__antithesis_instrumentation__.Notify(13663)
	t.Helper()
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(13664)
		count := 0
		if err := conn.QueryRow(
			"SELECT count(*) "+
				"FROM crdb_internal.ranges_no_leases "+
				"WHERE table_name = $1 "+
				"AND database_name = $2",
			tableName, dbName).Scan(&count); err != nil {
			__antithesis_instrumentation__.Notify(13667)
			return err
		} else {
			__antithesis_instrumentation__.Notify(13668)
		}
		__antithesis_instrumentation__.Notify(13665)
		if count == 0 {
			__antithesis_instrumentation__.Notify(13669)
			return errors.New("waiting for table split")
		} else {
			__antithesis_instrumentation__.Notify(13670)
		}
		__antithesis_instrumentation__.Notify(13666)
		return nil
	})
}

func getTableStartKey(t *testing.T, conn *gosql.DB, tableName, dbName string) roachpb.Key {
	__antithesis_instrumentation__.Notify(13671)
	t.Helper()
	row := conn.QueryRow(
		"SELECT start_key "+
			"FROM crdb_internal.ranges_no_leases "+
			"WHERE table_name = $1 "+
			"AND database_name = $2 "+
			"ORDER BY start_key ASC "+
			"LIMIT 1",
		tableName, dbName)
	var startKey roachpb.Key
	require.NoError(t, row.Scan(&startKey))
	return startKey
}

func getFirstStoreReplica(
	t *testing.T, s serverutils.TestServerInterface, key roachpb.Key,
) (*kvserver.Store, *kvserver.Replica) {
	__antithesis_instrumentation__.Notify(13672)
	t.Helper()
	store, err := s.GetStores().(*kvserver.Stores).GetStore(s.GetFirstStoreID())
	require.NoError(t, err)
	var repl *kvserver.Replica
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(13674)
		repl = store.LookupReplica(roachpb.RKey(key))
		if repl == nil {
			__antithesis_instrumentation__.Notify(13676)
			return errors.New(`could not find replica`)
		} else {
			__antithesis_instrumentation__.Notify(13677)
		}
		__antithesis_instrumentation__.Notify(13675)
		return nil
	})
	__antithesis_instrumentation__.Notify(13673)
	return store, repl
}

func getStoreAndReplica(
	t *testing.T, tc *testcluster.TestCluster, conn *gosql.DB, tableName, dbName string,
) (*kvserver.Store, *kvserver.Replica) {
	__antithesis_instrumentation__.Notify(13678)
	t.Helper()
	startKey := getTableStartKey(t, conn, tableName, dbName)

	r := tc.LookupRangeOrFatal(t, startKey)
	l, _, err := tc.FindRangeLease(r, nil)
	require.NoError(t, err)

	lhServer := tc.Server(int(l.Replica.NodeID) - 1)
	return getFirstStoreReplica(t, lhServer, startKey)
}

func waitForReplicaFieldToBeSet(
	t *testing.T,
	tc *testcluster.TestCluster,
	conn *gosql.DB,
	tableName, dbName string,
	isReplicaFieldSet func(r *kvserver.Replica) (bool, error),
) {
	__antithesis_instrumentation__.Notify(13679)
	t.Helper()
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(13680)
		_, r := getStoreAndReplica(t, tc, conn, tableName, dbName)
		if isSet, err := isReplicaFieldSet(r); !isSet {
			__antithesis_instrumentation__.Notify(13682)
			return err
		} else {
			__antithesis_instrumentation__.Notify(13683)
		}
		__antithesis_instrumentation__.Notify(13681)
		return nil
	})
}

func thresholdFromTrace(t *testing.T, traceString string) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(13684)
	t.Helper()
	thresholdRE := regexp.MustCompile(`(?s).*Threshold:(?P<threshold>[^\s]*)`)
	threshStr := string(thresholdRE.ExpandString(nil, "$threshold",
		traceString, thresholdRE.FindStringSubmatchIndex(traceString)))
	thresh, err := hlc.ParseTimestamp(threshStr)
	require.NoError(t, err)
	return thresh
}

func setAndWaitForTenantReadOnlyClusterSetting(
	t *testing.T,
	setting string,
	systemTenantRunner *sqlutils.SQLRunner,
	tenantRunner *sqlutils.SQLRunner,
	tenantID roachpb.TenantID,
	val string,
) {
	__antithesis_instrumentation__.Notify(13685)
	t.Helper()
	systemTenantRunner.Exec(
		t,
		fmt.Sprintf(
			"ALTER TENANT $1 SET CLUSTER	SETTING %s = '%s'",
			setting,
			val,
		),
		tenantID.ToUint64(),
	)

	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(13686)
		var currentVal string
		tenantRunner.QueryRow(t,
			fmt.Sprintf(
				"SHOW CLUSTER SETTING %s", setting,
			),
		).Scan(&currentVal)

		if currentVal != val {
			__antithesis_instrumentation__.Notify(13688)
			return errors.Newf("waiting for cluster setting to be set to %q", val)
		} else {
			__antithesis_instrumentation__.Notify(13689)
		}
		__antithesis_instrumentation__.Notify(13687)
		return nil
	})
}

func runGCAndCheckTraceForSecondaryTenant(
	ctx context.Context,
	t *testing.T,
	tc *testcluster.TestCluster,
	runner *sqlutils.SQLRunner,
	skipShouldQueue bool,
	startPretty string,
	checkGCTrace func(traceStr string) error,
) {
	__antithesis_instrumentation__.Notify(13690)
	t.Helper()
	var startKey roachpb.Key
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(13692)
		err := runner.DB.QueryRowContext(ctx, fmt.Sprintf(`
SELECT start_key FROM crdb_internal.ranges_no_leases
WHERE start_pretty LIKE '%s' ORDER BY start_key ASC`, startPretty)).Scan(&startKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(13694)
			return errors.Wrap(err, "failed to query start_key")
		} else {
			__antithesis_instrumentation__.Notify(13695)
		}
		__antithesis_instrumentation__.Notify(13693)
		return nil
	})
	__antithesis_instrumentation__.Notify(13691)

	r := tc.LookupRangeOrFatal(t, startKey)
	l, _, err := tc.FindRangeLease(r, nil)
	require.NoError(t, err)
	lhServer := tc.Server(int(l.Replica.NodeID) - 1)
	s, repl := getFirstStoreReplica(t, lhServer, startKey)
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(13696)
		trace, _, err := s.ManuallyEnqueue(ctx, "mvccGC", repl, skipShouldQueue)
		require.NoError(t, err)
		return checkGCTrace(trace.String())
	})
}

func runGCAndCheckTrace(
	ctx context.Context,
	t *testing.T,
	tc *testcluster.TestCluster,
	runner *sqlutils.SQLRunner,
	skipShouldQueue bool,
	databaseName, tableName string,
	checkGCTrace func(traceStr string) error,
) {
	__antithesis_instrumentation__.Notify(13697)
	t.Helper()
	var startKey roachpb.Key
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(13699)
		err := runner.DB.QueryRowContext(ctx, `
SELECT start_key FROM crdb_internal.ranges_no_leases
WHERE table_name = $1 AND database_name = $2
ORDER BY start_key ASC`, tableName, databaseName).Scan(&startKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(13701)
			return errors.Wrap(err, "failed to query start_key ")
		} else {
			__antithesis_instrumentation__.Notify(13702)
		}
		__antithesis_instrumentation__.Notify(13700)
		return nil
	})
	__antithesis_instrumentation__.Notify(13698)
	r := tc.LookupRangeOrFatal(t, startKey)
	l, _, err := tc.FindRangeLease(r, nil)
	require.NoError(t, err)
	lhServer := tc.Server(int(l.Replica.NodeID) - 1)
	s, repl := getFirstStoreReplica(t, lhServer, startKey)
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(13703)
		trace, _, err := s.ManuallyEnqueue(ctx, "mvccGC", repl, skipShouldQueue)
		require.NoError(t, err)
		return checkGCTrace(trace.String())
	})
}
