package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/stretchr/testify/require"
)

func registerMultiTenantUpgrade(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49420)
	r.Add(registry.TestSpec{
		Name:              "multitenant-upgrade",
		Cluster:           r.MakeClusterSpec(2),
		Owner:             registry.OwnerKV,
		NonReleaseBlocker: false,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49421)
			runMultiTenantUpgrade(ctx, t, c, *t.BuildVersion())
		},
	})
}

type tenantNode struct {
	tenantID          int
	httpPort, sqlPort int
	kvAddrs           []string
	pgURL             string

	binary string
	errCh  chan error
	node   int
}

func createTenantNode(kvAddrs []string, tenantID, node, httpPort, sqlPort int) *tenantNode {
	__antithesis_instrumentation__.Notify(49422)
	tn := &tenantNode{
		tenantID: tenantID,
		httpPort: httpPort,
		kvAddrs:  kvAddrs,
		node:     node,
		sqlPort:  sqlPort,
	}
	return tn
}

func (tn *tenantNode) stop(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(49423)
	if tn.errCh == nil {
		__antithesis_instrumentation__.Notify(49425)
		return
	} else {
		__antithesis_instrumentation__.Notify(49426)
	}
	__antithesis_instrumentation__.Notify(49424)

	c.Run(ctx, c.Node(tn.node),
		fmt.Sprintf("pkill -o -f '^%s mt start.*tenant-id=%d'", tn.binary, tn.tenantID))
	t.L().Printf("mt cluster exited: %v", <-tn.errCh)
	tn.errCh = nil
}

func (tn *tenantNode) logDir() string {
	__antithesis_instrumentation__.Notify(49427)
	return fmt.Sprintf("logs/mt-%d", tn.tenantID)
}

func (tn *tenantNode) storeDir() string {
	__antithesis_instrumentation__.Notify(49428)
	return fmt.Sprintf("cockroach-data-mt-%d", tn.tenantID)
}

func (tn *tenantNode) start(ctx context.Context, t test.Test, c cluster.Cluster, binary string) {
	__antithesis_instrumentation__.Notify(49429)
	tn.binary = binary
	extraArgs := []string{"--log-dir=" + tn.logDir(), "--store=" + tn.storeDir()}
	tn.errCh = startTenantServer(
		ctx, c, c.Node(tn.node), binary, tn.kvAddrs, tn.tenantID,
		tn.httpPort, tn.sqlPort,
		extraArgs...,
	)
	externalUrls, err := c.ExternalPGUrl(ctx, t.L(), c.Node(tn.node))
	require.NoError(t, err)
	u, err := url.Parse(externalUrls[0])
	require.NoError(t, err)
	internalUrls, err := c.ExternalIP(ctx, t.L(), c.Node(tn.node))
	require.NoError(t, err)
	u.Host = internalUrls[0] + ":" + strconv.Itoa(tn.sqlPort)
	tn.pgURL = u.String()

	if err := retry.ForDuration(45*time.Second, func() error {
		__antithesis_instrumentation__.Notify(49431)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(49434)
			t.Fatal(ctx.Err())
		case err := <-tn.errCh:
			__antithesis_instrumentation__.Notify(49435)
			t.Fatal(err)
		default:
			__antithesis_instrumentation__.Notify(49436)
		}
		__antithesis_instrumentation__.Notify(49432)

		db, err := gosql.Open("postgres", tn.pgURL)
		if err != nil {
			__antithesis_instrumentation__.Notify(49437)
			return err
		} else {
			__antithesis_instrumentation__.Notify(49438)
		}
		__antithesis_instrumentation__.Notify(49433)
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		_, err = db.ExecContext(ctx, `SELECT 1`)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(49439)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49440)
	}
	__antithesis_instrumentation__.Notify(49430)

	t.L().Printf("sql server for tenant %d running at %s", tn.tenantID, tn.pgURL)
}

func runMultiTenantUpgrade(ctx context.Context, t test.Test, c cluster.Cluster, v version.Version) {
	__antithesis_instrumentation__.Notify(49441)
	predecessor, err := PredecessorVersion(v)
	require.NoError(t, err)

	currentBinary := uploadVersion(ctx, t, c, c.All(), "")
	predecessorBinary := uploadVersion(ctx, t, c, c.All(), predecessor)

	kvNodes := c.Node(1)

	settings := install.MakeClusterSettings(install.BinaryOption(predecessorBinary))
	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, kvNodes)

	kvAddrs, err := c.ExternalAddr(ctx, t.L(), kvNodes)
	require.NoError(t, err)

	const tenant11HTTPPort, tenant11SQLPort = 8011, 20011
	const tenant11ID = 11
	runner := sqlutils.MakeSQLRunner(c.Conn(ctx, t.L(), 1))

	runner.SucceedsSoonDuration = 5 * time.Minute
	runner.Exec(t, `SELECT crdb_internal.create_tenant($1)`, tenant11ID)

	var initialVersion string
	runner.QueryRow(t, "SHOW CLUSTER SETTING version").Scan(&initialVersion)

	const tenantNode = 2
	tenant11 := createTenantNode(kvAddrs, tenant11ID, tenantNode, tenant11HTTPPort, tenant11SQLPort)
	tenant11.start(ctx, t, c, predecessorBinary)
	defer tenant11.stop(ctx, t, c)

	t.Status("checking that a client can connect to the tenant 11 server")
	verifySQL(t, tenant11.pgURL,
		mkStmt(`CREATE TABLE foo (id INT PRIMARY KEY, v STRING)`),
		mkStmt(`INSERT INTO foo VALUES($1, $2)`, 1, "bar"),
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}))

	verifySQL(t, tenant11.pgURL,
		mkStmt("SHOW CLUSTER SETTING version").
			withResults([][]string{{initialVersion}}),
	)

	t.Status("preserving downgrade option on host server")
	{
		__antithesis_instrumentation__.Notify(49444)
		s := runner.QueryStr(t, `SHOW CLUSTER SETTING version`)
		runner.Exec(
			t,
			`SET CLUSTER SETTING cluster.preserve_downgrade_option = $1`, s[0][0],
		)
	}
	__antithesis_instrumentation__.Notify(49442)

	t.Status("upgrading host server")
	c.Stop(ctx, t.L(), option.DefaultStopOpts(), kvNodes)
	settings.Binary = currentBinary
	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, kvNodes)
	time.Sleep(time.Second)

	t.Status("checking the pre-upgrade sql server still works after the KV binary upgrade")

	verifySQL(t, tenant11.pgURL,
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}))

	t.Status("creating a new tenant 12")

	const tenant12HTTPPort, tenant12SQLPort = 8012, 20012
	const tenant12ID = 12
	runner.Exec(t, `SELECT crdb_internal.create_tenant($1)`, tenant12ID)

	t.Status("starting tenant 12 server with older binary")
	tenant12 := createTenantNode(kvAddrs, tenant12ID, tenantNode, tenant12HTTPPort, tenant12SQLPort)
	tenant12.start(ctx, t, c, predecessorBinary)
	defer tenant12.stop(ctx, t, c)

	t.Status("verifying that the tenant 12 server works and is at the earlier version")

	verifySQL(t, tenant12.pgURL,
		mkStmt(`CREATE TABLE foo (id INT PRIMARY KEY, v STRING)`),
		mkStmt(`INSERT INTO foo VALUES($1, $2)`, 1, "bar"),
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}),
		mkStmt("SHOW CLUSTER SETTING version").
			withResults([][]string{{initialVersion}}),
	)

	t.Status("creating a new tenant 13")

	const tenant13HTTPPort, tenant13SQLPort = 8013, 20013
	const tenant13ID = 13
	runner.Exec(t, `SELECT crdb_internal.create_tenant($1)`, tenant13ID)

	t.Status("starting tenant 13 server with new binary")
	tenant13 := createTenantNode(kvAddrs, tenant13ID, tenantNode, tenant13HTTPPort, tenant13SQLPort)
	tenant13.start(ctx, t, c, currentBinary)
	defer tenant13.stop(ctx, t, c)

	t.Status("verifying that the tenant 13 server works and is at the earlier version")

	verifySQL(t, tenant13.pgURL,
		mkStmt(`CREATE TABLE foo (id INT PRIMARY KEY, v STRING)`),
		mkStmt(`INSERT INTO foo VALUES($1, $2)`, 1, "bar"),
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}),
		mkStmt("SHOW CLUSTER SETTING version").
			withResults([][]string{{initialVersion}}),
	)

	t.Status("stopping the tenant 11 server ahead of upgrading")
	tenant11.stop(ctx, t, c)

	t.Status("starting the tenant 11 server with the current binary")
	tenant11.start(ctx, t, c, currentBinary)

	t.Status("verify tenant 11 server works with the new binary")
	{
		__antithesis_instrumentation__.Notify(49445)
		verifySQL(t, tenant11.pgURL,
			mkStmt(`SELECT * FROM foo LIMIT 1`).
				withResults([][]string{{"1", "bar"}}),
			mkStmt("SHOW CLUSTER SETTING version").
				withResults([][]string{{initialVersion}}))
	}
	__antithesis_instrumentation__.Notify(49443)

	t.Status("migrating the tenant 11 to the current version before kv is finalized")

	verifySQL(t, tenant11.pgURL,
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}),
		mkStmt("SHOW CLUSTER SETTING version").
			withResults([][]string{{initialVersion}}),
		mkStmt("SET CLUSTER SETTING version = crdb_internal.node_executable_version()"),
		mkStmt("SELECT version = crdb_internal.node_executable_version() FROM [SHOW CLUSTER SETTING version]").
			withResults([][]string{{"true"}}),
	)

	t.Status("finalizing the kv server")
	runner.Exec(t, `SET CLUSTER SETTING cluster.preserve_downgrade_option = DEFAULT`)
	runner.CheckQueryResultsRetry(t,
		"SELECT version = crdb_internal.node_executable_version() FROM [SHOW CLUSTER SETTING version]",
		[][]string{{"true"}})

	t.Status("stopping the tenant 12 server ahead of upgrading")
	tenant12.stop(ctx, t, c)

	t.Status("starting the tenant 12 server with the current binary")
	tenant12.start(ctx, t, c, currentBinary)

	t.Status("verify tenant 12 server works with the new binary")
	verifySQL(t, tenant12.pgURL,
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}),
		mkStmt("SHOW CLUSTER SETTING version").
			withResults([][]string{{initialVersion}}))

	t.Status("migrating tenant 12 to the current version")
	verifySQL(t, tenant12.pgURL,
		mkStmt("SET CLUSTER SETTING version = crdb_internal.node_executable_version()"),
		mkStmt("SELECT version = crdb_internal.node_executable_version() FROM [SHOW CLUSTER SETTING version]").
			withResults([][]string{{"true"}}))

	t.Status("restarting the tenant 12 server to check it works after a restart")
	tenant12.stop(ctx, t, c)
	tenant12.start(ctx, t, c, currentBinary)

	t.Status("verify tenant 12 server works with the new binary after restart")
	verifySQL(t, tenant12.pgURL,
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}),
		mkStmt("SELECT version = crdb_internal.node_executable_version() FROM [SHOW CLUSTER SETTING version]").
			withResults([][]string{{"true"}}))

	t.Status("migrating tenant 13 to the current version")
	verifySQL(t, tenant13.pgURL,
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}),
		mkStmt("SET CLUSTER SETTING version = crdb_internal.node_executable_version()"),
		mkStmt("SELECT version = crdb_internal.node_executable_version() FROM [SHOW CLUSTER SETTING version]").
			withResults([][]string{{"true"}}),
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}))

	t.Status("restarting the tenant 13 server to check it works after a restart")
	tenant13.stop(ctx, t, c)
	tenant13.start(ctx, t, c, currentBinary)

	t.Status("verify tenant 13 server works with the new binary after restart")
	verifySQL(t, tenant13.pgURL,
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}),
		mkStmt("SELECT version = crdb_internal.node_executable_version() FROM [SHOW CLUSTER SETTING version]").
			withResults([][]string{{"true"}}))

	t.Status("creating tenant 14 at the new version")

	const tenant14HTTPPort, tenant14SQLPort = 8014, 20014
	const tenant14ID = 14
	runner.Exec(t, `SELECT crdb_internal.create_tenant($1)`, tenant14ID)

	t.Status("verifying the tenant 14 works and has the proper version")
	tenant14 := createTenantNode(kvAddrs, tenant14ID, tenantNode, tenant14HTTPPort, tenant14SQLPort)
	tenant14.start(ctx, t, c, currentBinary)
	defer tenant14.stop(ctx, t, c)
	verifySQL(t, tenant14.pgURL,
		mkStmt(`CREATE TABLE foo (id INT PRIMARY KEY, v STRING)`),
		mkStmt(`INSERT INTO foo VALUES($1, $2)`, 1, "bar"),
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}),
		mkStmt("SELECT version = crdb_internal.node_executable_version() FROM [SHOW CLUSTER SETTING version]").
			withResults([][]string{{"true"}}))

	t.Status("restarting the tenant 14 server to check it works after a restart")
	tenant13.stop(ctx, t, c)
	tenant13.start(ctx, t, c, currentBinary)

	t.Status("verifying the post-upgrade tenant works and has the proper version")
	verifySQL(t, tenant14.pgURL,
		mkStmt(`SELECT * FROM foo LIMIT 1`).
			withResults([][]string{{"1", "bar"}}),
		mkStmt("SELECT version = crdb_internal.node_executable_version() FROM [SHOW CLUSTER SETTING version]").
			withResults([][]string{{"true"}}))
}

func startTenantServer(
	tenantCtx context.Context,
	c cluster.Cluster,
	node option.NodeListOption,
	binary string,
	kvAddrs []string,
	tenantID int,
	httpPort int,
	sqlPort int,
	extraFlags ...string,
) chan error {
	__antithesis_instrumentation__.Notify(49446)

	args := []string{

		"--insecure",
		"--tenant-id=" + strconv.Itoa(tenantID),
		"--http-addr", ifLocal(c, "127.0.0.1", "0.0.0.0") + ":" + strconv.Itoa(httpPort),
		"--kv-addrs", strings.Join(kvAddrs, ","),

		"--sql-addr", ifLocal(c, "127.0.0.1", "0.0.0.0") + ":" + strconv.Itoa(sqlPort),
	}
	args = append(args, extraFlags...)
	errCh := make(chan error, 1)
	go func() {
		__antithesis_instrumentation__.Notify(49448)
		errCh <- c.RunE(tenantCtx, node,
			append([]string{binary, "mt", "start-sql"}, args...)...,
		)
		close(errCh)
	}()
	__antithesis_instrumentation__.Notify(49447)
	return errCh
}

type sqlVerificationStmt struct {
	stmt            string
	args            []interface{}
	optionalResults [][]string
}

func (s sqlVerificationStmt) withResults(res [][]string) sqlVerificationStmt {
	__antithesis_instrumentation__.Notify(49449)
	s.optionalResults = res
	return s
}

func mkStmt(stmt string, args ...interface{}) sqlVerificationStmt {
	__antithesis_instrumentation__.Notify(49450)
	return sqlVerificationStmt{stmt: stmt, args: args}
}

func verifySQL(t test.Test, url string, stmts ...sqlVerificationStmt) {
	__antithesis_instrumentation__.Notify(49451)
	db, err := gosql.Open("postgres", url)
	if err != nil {
		__antithesis_instrumentation__.Notify(49454)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49455)
	}
	__antithesis_instrumentation__.Notify(49452)
	defer func() { __antithesis_instrumentation__.Notify(49456); _ = db.Close() }()
	__antithesis_instrumentation__.Notify(49453)
	tdb := sqlutils.MakeSQLRunner(db)

	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(49457)
		if stmt.optionalResults == nil {
			__antithesis_instrumentation__.Notify(49458)
			tdb.Exec(t, stmt.stmt, stmt.args...)
		} else {
			__antithesis_instrumentation__.Notify(49459)
			res := tdb.QueryStr(t, stmt.stmt, stmt.args...)
			require.Equal(t, stmt.optionalResults, res)
		}
	}
}
