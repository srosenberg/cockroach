// Package systemconfigwatchertest exists to exercise systemconfigwatcher
// in both ccl and non-ccl configurations.
package systemconfigwatchertest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"sort"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvtenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemConfigWatcher(t *testing.T, skipSecondary bool) {
	__antithesis_instrumentation__.Notify(238154)
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	ctx := context.Background()
	s, sqlDB, kvDB := serverutils.StartServer(t, base.TestServerArgs{})
	defer s.Stopper().Stop(ctx)
	tdb := sqlutils.MakeSQLRunner(sqlDB)

	tdb.Exec(t, "SET CLUSTER SETTING kv.closed_timestamp.target_duration = '10 ms'")
	tdb.Exec(t, "SET CLUSTER SETTING kv.closed_timestamp.side_transport_interval = '10 ms'")

	t.Run("system", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(238156)
		runTest(t, s, sqlDB, nil)
	})
	__antithesis_instrumentation__.Notify(238155)
	if !skipSecondary {
		__antithesis_instrumentation__.Notify(238157)
		t.Run("secondary", func(t *testing.T) {
			__antithesis_instrumentation__.Notify(238158)
			tenant, tenantDB := serverutils.StartTenant(t, s, base.TestTenantArgs{
				TenantID: serverutils.TestTenantID(),
			})

			runTest(t, tenant, tenantDB, func(t *testing.T) []roachpb.KeyValue {
				__antithesis_instrumentation__.Notify(238159)
				return kvtenant.GossipSubscriptionSystemConfigMask.Apply(
					config.SystemConfigEntries{
						Values: getSystemConfig(ctx, t, keys.SystemSQLCodec, kvDB),
					},
				).Values
			})
		})
	} else {
		__antithesis_instrumentation__.Notify(238160)
	}
}

func runTest(
	t *testing.T,
	s serverutils.TestTenantInterface,
	sqlDB *gosql.DB,
	extraRows func(t *testing.T) []roachpb.KeyValue,
) {
	__antithesis_instrumentation__.Notify(238161)
	ctx := context.Background()
	tdb := sqlutils.MakeSQLRunner(sqlDB)
	execCfg := s.ExecutorConfig().(sql.ExecutorConfig)
	kvDB := execCfg.DB
	r := execCfg.SystemConfig
	rc, _ := r.RegisterSystemConfigChannel()
	clearChan := func() {
		__antithesis_instrumentation__.Notify(238165)
		select {
		case <-rc:
			__antithesis_instrumentation__.Notify(238166)
		default:
			__antithesis_instrumentation__.Notify(238167)
		}
	}
	__antithesis_instrumentation__.Notify(238162)
	checkEqual := func(t *testing.T) error {
		__antithesis_instrumentation__.Notify(238168)
		rs := r.GetSystemConfig()
		if rs == nil {
			__antithesis_instrumentation__.Notify(238172)
			return errors.New("nil config")
		} else {
			__antithesis_instrumentation__.Notify(238173)
		}
		__antithesis_instrumentation__.Notify(238169)
		sc := getSystemConfig(ctx, t, execCfg.Codec, kvDB)
		if extraRows != nil {
			__antithesis_instrumentation__.Notify(238174)
			sc = append(sc, extraRows(t)...)
			sort.Sort(roachpb.KeyValueByKey(sc))
		} else {
			__antithesis_instrumentation__.Notify(238175)
		}
		__antithesis_instrumentation__.Notify(238170)
		if !assert.Equal(noopT{}, sc, rs.Values) {
			__antithesis_instrumentation__.Notify(238176)
			return errors.Errorf("mismatch: %v", pretty.Diff(sc, rs.Values))
		} else {
			__antithesis_instrumentation__.Notify(238177)
		}
		__antithesis_instrumentation__.Notify(238171)
		return nil
	}
	__antithesis_instrumentation__.Notify(238163)
	waitForEqual := func(t *testing.T) {
		__antithesis_instrumentation__.Notify(238178)
		testutils.SucceedsSoon(t, func() error {
			__antithesis_instrumentation__.Notify(238179)
			return checkEqual(t)
		})
	}
	__antithesis_instrumentation__.Notify(238164)
	waitForEqual(t)
	clearChan()
	tdb.Exec(t, "CREATE TABLE foo (i INT PRIMARY KEY)")
	<-rc
	waitForEqual(t)
}

func getSystemConfig(
	ctx context.Context, t *testing.T, codec keys.SQLCodec, kvDB *kv.DB,
) []roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(238180)
	var ba roachpb.BatchRequest
	ba.Add(roachpb.NewScan(
		append(codec.TenantPrefix(), keys.SystemConfigSpan.Key...),
		append(codec.TenantPrefix(), keys.SystemConfigSpan.EndKey...),
		false,
	))
	br, pErr := kvDB.NonTransactionalSender().Send(ctx, ba)
	require.NoError(t, pErr.GoError())
	return br.Responses[0].GetScan().Rows
}

type noopT struct{}

func (noopT) Errorf(string, ...interface{}) { __antithesis_instrumentation__.Notify(238181) }

var _ assert.TestingT = (*noopT)(nil)
