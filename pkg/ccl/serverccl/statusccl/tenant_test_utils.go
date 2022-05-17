package statusccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"math/rand"
	"net/http"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/contention"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/tests"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/stretchr/testify/require"
)

type serverIdx int

const randomServer serverIdx = -1

type testTenant struct {
	tenant                   serverutils.TestTenantInterface
	tenantConn               *gosql.DB
	tenantDB                 *sqlutils.SQLRunner
	tenantStatus             serverpb.SQLStatusServer
	tenantSQLStats           *persistedsqlstats.PersistedSQLStats
	tenantContentionRegistry *contention.Registry
}

func newTestTenant(
	t *testing.T,
	server serverutils.TestServerInterface,
	existing bool,
	tenantID roachpb.TenantID,
	knobs base.TestingKnobs,
) *testTenant {
	__antithesis_instrumentation__.Notify(20857)
	t.Helper()

	tenantParams := tests.CreateTestTenantParams(tenantID)
	tenantParams.Existing = existing
	tenantParams.TestingKnobs = knobs

	tenant, tenantConn := serverutils.StartTenant(t, server, tenantParams)
	sqlDB := sqlutils.MakeSQLRunner(tenantConn)
	status := tenant.StatusServer().(serverpb.SQLStatusServer)
	sqlStats := tenant.PGServer().(*pgwire.Server).SQLServer.
		GetSQLStatsProvider().(*persistedsqlstats.PersistedSQLStats)
	contentionRegistry := tenant.ExecutorConfig().(sql.ExecutorConfig).ContentionRegistry

	return &testTenant{
		tenant:                   tenant,
		tenantConn:               tenantConn,
		tenantDB:                 sqlDB,
		tenantStatus:             status,
		tenantSQLStats:           sqlStats,
		tenantContentionRegistry: contentionRegistry,
	}
}

func (h *testTenant) cleanup(t *testing.T) {
	__antithesis_instrumentation__.Notify(20858)
	require.NoError(t, h.tenantConn.Close())
}

type tenantTestHelper struct {
	hostCluster serverutils.TestClusterInterface

	tenantTestCluster    tenantCluster
	tenantControlCluster tenantCluster
}

func newTestTenantHelper(
	t *testing.T, tenantClusterSize int, knobs base.TestingKnobs,
) *tenantTestHelper {
	__antithesis_instrumentation__.Notify(20859)
	t.Helper()

	params, _ := tests.CreateTestServerParams()
	testCluster := serverutils.StartNewTestCluster(t, 1, base.TestClusterArgs{
		ServerArgs: params,
	})
	server := testCluster.Server(0)

	return &tenantTestHelper{
		hostCluster: testCluster,
		tenantTestCluster: newTenantCluster(
			t,
			server,
			tenantClusterSize,
			security.EmbeddedTenantIDs()[0],
			knobs,
		),

		tenantControlCluster: newTenantCluster(
			t,
			server,
			1,
			security.EmbeddedTenantIDs()[1],
			knobs,
		),
	}
}

func (h *tenantTestHelper) testCluster() tenantCluster {
	__antithesis_instrumentation__.Notify(20860)
	return h.tenantTestCluster
}

func (h *tenantTestHelper) controlCluster() tenantCluster {
	__antithesis_instrumentation__.Notify(20861)
	return h.tenantControlCluster
}

func (h *tenantTestHelper) cleanup(ctx context.Context, t *testing.T) {
	__antithesis_instrumentation__.Notify(20862)
	t.Helper()
	h.hostCluster.Stopper().Stop(ctx)
	h.tenantTestCluster.cleanup(t)
	h.tenantControlCluster.cleanup(t)
}

type tenantCluster []*testTenant

func newTenantCluster(
	t *testing.T,
	server serverutils.TestServerInterface,
	tenantClusterSize int,
	tenantID uint64,
	knobs base.TestingKnobs,
) tenantCluster {
	__antithesis_instrumentation__.Notify(20863)
	t.Helper()

	cluster := make([]*testTenant, tenantClusterSize)
	existing := false
	for i := 0; i < tenantClusterSize; i++ {
		__antithesis_instrumentation__.Notify(20865)
		cluster[i] =
			newTestTenant(t, server, existing, roachpb.MakeTenantID(tenantID), knobs)
		existing = true
	}
	__antithesis_instrumentation__.Notify(20864)

	return cluster
}

func (c tenantCluster) tenantConn(idx serverIdx) *sqlutils.SQLRunner {
	__antithesis_instrumentation__.Notify(20866)
	return c.tenant(idx).tenantDB
}

func (c tenantCluster) tenantHTTPClient(t *testing.T, idx serverIdx, isAdmin bool) *httpClient {
	__antithesis_instrumentation__.Notify(20867)
	client, err := c.tenant(idx).tenant.GetAuthenticatedHTTPClient(isAdmin)
	require.NoError(t, err)
	return &httpClient{t: t, client: client, baseURL: c[idx].tenant.AdminURL()}
}

func (c tenantCluster) tenantAdminHTTPClient(t *testing.T, idx serverIdx) *httpClient {
	__antithesis_instrumentation__.Notify(20868)
	return c.tenantHTTPClient(t, idx, true)
}

func (c tenantCluster) tenantSQLStats(idx serverIdx) *persistedsqlstats.PersistedSQLStats {
	__antithesis_instrumentation__.Notify(20869)
	return c.tenant(idx).tenantSQLStats
}

func (c tenantCluster) tenantStatusSrv(idx serverIdx) serverpb.SQLStatusServer {
	__antithesis_instrumentation__.Notify(20870)
	return c.tenant(idx).tenantStatus
}

func (c tenantCluster) tenantContentionRegistry(idx serverIdx) *contention.Registry {
	__antithesis_instrumentation__.Notify(20871)
	return c.tenant(idx).tenantContentionRegistry
}

func (c tenantCluster) cleanup(t *testing.T) {
	__antithesis_instrumentation__.Notify(20872)
	for _, tenant := range c {
		__antithesis_instrumentation__.Notify(20873)
		tenant.cleanup(t)
	}
}

func (c tenantCluster) tenant(idx serverIdx) *testTenant {
	__antithesis_instrumentation__.Notify(20874)
	if idx == randomServer {
		__antithesis_instrumentation__.Notify(20876)
		return c[rand.Intn(len(c))]
	} else {
		__antithesis_instrumentation__.Notify(20877)
	}
	__antithesis_instrumentation__.Notify(20875)

	return c[idx]
}

type httpClient struct {
	t       *testing.T
	client  http.Client
	baseURL string
}

func (c *httpClient) GetJSON(path string, response protoutil.Message) {
	__antithesis_instrumentation__.Notify(20878)
	err := httputil.GetJSON(c.client, c.baseURL+path, response)
	require.NoError(c.t, err)
}

func (c *httpClient) PostJSON(path string, request protoutil.Message, response protoutil.Message) {
	__antithesis_instrumentation__.Notify(20879)
	err := c.PostJSONChecked(path, request, response)
	require.NoError(c.t, err)
}

func (c *httpClient) PostJSONChecked(
	path string, request protoutil.Message, response protoutil.Message,
) error {
	__antithesis_instrumentation__.Notify(20880)
	return httputil.PostJSON(c.client, c.baseURL+path, request, response)
}

func (c *httpClient) Close() {
	__antithesis_instrumentation__.Notify(20881)
	c.client.CloseIdleConnections()
}
