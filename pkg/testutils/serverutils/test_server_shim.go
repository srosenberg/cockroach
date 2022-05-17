package serverutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"net/url"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type TestServerInterface interface {
	Start(context.Context) error

	TestTenantInterface

	Node() interface{}

	NodeID() roachpb.NodeID

	StorageClusterID() uuid.UUID

	ServingRPCAddr() string

	ServingSQLAddr() string

	RPCAddr() string

	DB() *kv.DB

	LeaseManager() interface{}

	InternalExecutor() interface{}

	TracerI() interface{}

	GossipI() interface{}

	DistSenderI() interface{}

	MigrationServer() interface{}

	SQLServer() interface{}

	SQLLivenessProvider() interface{}

	StartupMigrationsManager() interface{}

	NodeLiveness() interface{}

	HeartbeatNodeLiveness() error

	NodeDialer() interface{}

	SetDistSQLSpanResolver(spanResolver interface{})

	MustGetSQLCounter(name string) int64

	MustGetSQLNetworkCounter(name string) int64

	WriteSummaries() error

	GetFirstStoreID() roachpb.StoreID

	GetStores() interface{}

	Decommission(ctx context.Context, targetStatus livenesspb.MembershipStatus, nodeIDs []roachpb.NodeID) error

	SplitRange(splitKey roachpb.Key) (left roachpb.RangeDescriptor, right roachpb.RangeDescriptor, err error)

	MergeRanges(leftKey roachpb.Key) (merged roachpb.RangeDescriptor, err error)

	ExpectedInitialRangeCount() (int, error)

	ForceTableGC(ctx context.Context, database, table string, timestamp hlc.Timestamp) error

	UpdateChecker() interface{}

	StartTenant(ctx context.Context, params base.TestTenantArgs) (TestTenantInterface, error)

	ScratchRange() (roachpb.Key, error)

	Engines() []storage.Engine

	MetricsRecorder() *status.MetricsRecorder

	CollectionFactory() interface{}

	SystemTableIDResolver() interface{}

	SpanConfigKVSubscriber() interface{}
}

type TestServerFactory interface {
	New(params base.TestServerArgs) (interface{}, error)
}

var srvFactoryImpl TestServerFactory

func InitTestServerFactory(impl TestServerFactory) {
	__antithesis_instrumentation__.Notify(645928)
	srvFactoryImpl = impl
}

func StartServer(
	t testing.TB, params base.TestServerArgs,
) (TestServerInterface, *gosql.DB, *kv.DB) {
	__antithesis_instrumentation__.Notify(645929)
	server, err := NewServer(params)
	if err != nil {
		__antithesis_instrumentation__.Notify(645932)
		t.Fatalf("%+v", err)
	} else {
		__antithesis_instrumentation__.Notify(645933)
	}
	__antithesis_instrumentation__.Notify(645930)
	if err := server.Start(context.Background()); err != nil {
		__antithesis_instrumentation__.Notify(645934)
		t.Fatalf("%+v", err)
	} else {
		__antithesis_instrumentation__.Notify(645935)
	}
	__antithesis_instrumentation__.Notify(645931)
	goDB := OpenDBConn(
		t, server.ServingSQLAddr(), params.UseDatabase, params.Insecure, server.Stopper())
	return server, goDB, server.DB()
}

func NewServer(params base.TestServerArgs) (TestServerInterface, error) {
	__antithesis_instrumentation__.Notify(645936)
	if srvFactoryImpl == nil {
		__antithesis_instrumentation__.Notify(645939)
		return nil, errors.AssertionFailedf("TestServerFactory not initialized. One needs to be injected " +
			"from the package's TestMain()")
	} else {
		__antithesis_instrumentation__.Notify(645940)
	}
	__antithesis_instrumentation__.Notify(645937)

	srv, err := srvFactoryImpl.New(params)
	if err != nil {
		__antithesis_instrumentation__.Notify(645941)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(645942)
	}
	__antithesis_instrumentation__.Notify(645938)
	return srv.(TestServerInterface), nil
}

func OpenDBConnE(
	sqlAddr string, useDatabase string, insecure bool, stopper *stop.Stopper,
) (*gosql.DB, error) {
	__antithesis_instrumentation__.Notify(645943)
	pgURL, cleanupGoDB, err := sqlutils.PGUrlE(
		sqlAddr, "StartServer", url.User(security.RootUser))
	if err != nil {
		__antithesis_instrumentation__.Notify(645948)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(645949)
	}
	__antithesis_instrumentation__.Notify(645944)

	pgURL.Path = useDatabase
	if insecure {
		__antithesis_instrumentation__.Notify(645950)
		pgURL.RawQuery = "sslmode=disable"
	} else {
		__antithesis_instrumentation__.Notify(645951)
	}
	__antithesis_instrumentation__.Notify(645945)
	goDB, err := gosql.Open("postgres", pgURL.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(645952)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(645953)
	}
	__antithesis_instrumentation__.Notify(645946)

	stopper.AddCloser(
		stop.CloserFn(func() {
			__antithesis_instrumentation__.Notify(645954)
			_ = goDB.Close()
			cleanupGoDB()
		}))
	__antithesis_instrumentation__.Notify(645947)
	return goDB, nil
}

func OpenDBConn(
	t testing.TB, sqlAddr string, useDatabase string, insecure bool, stopper *stop.Stopper,
) *gosql.DB {
	__antithesis_instrumentation__.Notify(645955)
	conn, err := OpenDBConnE(sqlAddr, useDatabase, insecure, stopper)
	if err != nil {
		__antithesis_instrumentation__.Notify(645957)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(645958)
	}
	__antithesis_instrumentation__.Notify(645956)
	return conn
}

func StartServerRaw(args base.TestServerArgs) (TestServerInterface, error) {
	__antithesis_instrumentation__.Notify(645959)
	server, err := NewServer(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(645962)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(645963)
	}
	__antithesis_instrumentation__.Notify(645960)
	if err := server.Start(context.Background()); err != nil {
		__antithesis_instrumentation__.Notify(645964)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(645965)
	}
	__antithesis_instrumentation__.Notify(645961)
	return server, nil
}

func StartTenant(
	t testing.TB, ts TestServerInterface, params base.TestTenantArgs,
) (TestTenantInterface, *gosql.DB) {
	__antithesis_instrumentation__.Notify(645966)
	tenant, err := ts.StartTenant(context.Background(), params)
	if err != nil {
		__antithesis_instrumentation__.Notify(645969)
		t.Fatalf("%+v", err)
	} else {
		__antithesis_instrumentation__.Notify(645970)
	}
	__antithesis_instrumentation__.Notify(645967)

	stopper := params.Stopper
	if stopper == nil {
		__antithesis_instrumentation__.Notify(645971)
		stopper = ts.Stopper()
	} else {
		__antithesis_instrumentation__.Notify(645972)
	}
	__antithesis_instrumentation__.Notify(645968)

	goDB := OpenDBConn(
		t, tenant.SQLAddr(), params.UseDatabase, false, stopper)
	return tenant, goDB
}

func TestTenantID() roachpb.TenantID {
	__antithesis_instrumentation__.Notify(645973)
	return roachpb.MakeTenantID(security.EmbeddedTenantIDs()[0])
}

func GetJSONProto(ts TestServerInterface, path string, response protoutil.Message) error {
	__antithesis_instrumentation__.Notify(645974)
	return GetJSONProtoWithAdminOption(ts, path, response, true)
}

func GetJSONProtoWithAdminOption(
	ts TestServerInterface, path string, response protoutil.Message, isAdmin bool,
) error {
	__antithesis_instrumentation__.Notify(645975)
	httpClient, err := ts.GetAuthenticatedHTTPClient(isAdmin)
	if err != nil {
		__antithesis_instrumentation__.Notify(645977)
		return err
	} else {
		__antithesis_instrumentation__.Notify(645978)
	}
	__antithesis_instrumentation__.Notify(645976)
	return httputil.GetJSON(httpClient, ts.AdminURL()+path, response)
}

func PostJSONProto(ts TestServerInterface, path string, request, response protoutil.Message) error {
	__antithesis_instrumentation__.Notify(645979)
	return PostJSONProtoWithAdminOption(ts, path, request, response, true)
}

func PostJSONProtoWithAdminOption(
	ts TestServerInterface, path string, request, response protoutil.Message, isAdmin bool,
) error {
	__antithesis_instrumentation__.Notify(645980)
	httpClient, err := ts.GetAuthenticatedHTTPClient(isAdmin)
	if err != nil {
		__antithesis_instrumentation__.Notify(645982)
		return err
	} else {
		__antithesis_instrumentation__.Notify(645983)
	}
	__antithesis_instrumentation__.Notify(645981)
	return httputil.PostJSON(httpClient, ts.AdminURL()+path, request, response)
}
