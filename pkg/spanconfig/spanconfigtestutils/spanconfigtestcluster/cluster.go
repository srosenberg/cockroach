package spanconfigtestcluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"net/url"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigreconciler"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigsqltranslator"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigtestutils"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/testcluster"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

type Handle struct {
	t       *testing.T
	tc      *testcluster.TestCluster
	ts      map[roachpb.TenantID]*Tenant
	scKnobs *spanconfig.TestingKnobs
}

func NewHandle(
	t *testing.T, tc *testcluster.TestCluster, scKnobs *spanconfig.TestingKnobs,
) *Handle {
	__antithesis_instrumentation__.Notify(241739)
	return &Handle{
		t:       t,
		tc:      tc,
		ts:      make(map[roachpb.TenantID]*Tenant),
		scKnobs: scKnobs,
	}
}

func (h *Handle) InitializeTenant(ctx context.Context, tenID roachpb.TenantID) *Tenant {
	__antithesis_instrumentation__.Notify(241740)
	testServer := h.tc.Server(0)
	tenantState := &Tenant{t: h.t}
	if tenID == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(241743)
		tenantState.TestTenantInterface = testServer
		tenantState.db = sqlutils.MakeSQLRunner(h.tc.ServerConn(0))
		tenantState.cleanup = func() { __antithesis_instrumentation__.Notify(241744) }
	} else {
		__antithesis_instrumentation__.Notify(241745)
		tenantArgs := base.TestTenantArgs{
			TenantID: tenID,
			TestingKnobs: base.TestingKnobs{
				SpanConfig: h.scKnobs,
			},
		}
		var err error
		tenantState.TestTenantInterface, err = testServer.StartTenant(ctx, tenantArgs)
		require.NoError(h.t, err)

		pgURL, cleanupPGUrl := sqlutils.PGUrl(h.t, tenantState.SQLAddr(), "Tenant", url.User(security.RootUser))
		tenantSQLDB, err := gosql.Open("postgres", pgURL.String())
		require.NoError(h.t, err)

		tenantState.db = sqlutils.MakeSQLRunner(tenantSQLDB)
		tenantState.cleanup = func() {
			__antithesis_instrumentation__.Notify(241746)
			require.NoError(h.t, tenantSQLDB.Close())
			cleanupPGUrl()
		}
	}
	__antithesis_instrumentation__.Notify(241741)

	var tenKnobs *spanconfig.TestingKnobs
	if scKnobs := tenantState.TestingKnobs().SpanConfig; scKnobs != nil {
		__antithesis_instrumentation__.Notify(241747)
		tenKnobs = scKnobs.(*spanconfig.TestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(241748)
	}
	__antithesis_instrumentation__.Notify(241742)
	tenExecCfg := tenantState.ExecutorConfig().(sql.ExecutorConfig)
	tenKVAccessor := tenantState.SpanConfigKVAccessor().(spanconfig.KVAccessor)
	tenSQLTranslatorFactory := tenantState.SpanConfigSQLTranslatorFactory().(*spanconfigsqltranslator.Factory)
	tenSQLWatcher := tenantState.SpanConfigSQLWatcher().(spanconfig.SQLWatcher)

	tenantState.recorder = spanconfigtestutils.NewKVAccessorRecorder(tenKVAccessor)
	tenantState.reconciler = spanconfigreconciler.New(
		tenSQLWatcher,
		tenSQLTranslatorFactory,
		tenantState.recorder,
		&tenExecCfg,
		tenExecCfg.Codec,
		tenID,
		tenKnobs,
	)

	h.ts[tenID] = tenantState
	return tenantState
}

func (h *Handle) AllowSecondaryTenantToSetZoneConfigurations(t *testing.T, tenID roachpb.TenantID) {
	__antithesis_instrumentation__.Notify(241749)
	_, found := h.LookupTenant(tenID)
	require.True(t, found)
	sqlDB := sqlutils.MakeSQLRunner(h.tc.ServerConn(0))
	sqlDB.Exec(
		t,
		"ALTER TENANT $1 SET CLUSTER SETTING sql.zone_configs.allow_for_secondary_tenant.enabled = true",
		tenID.ToUint64(),
	)
}

func (h *Handle) EnsureTenantCanSetZoneConfigurationsOrFatal(t *testing.T, tenant *Tenant) {
	__antithesis_instrumentation__.Notify(241750)
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(241751)
		var val string
		tenant.QueryRow(
			"SHOW CLUSTER SETTING sql.zone_configs.allow_for_secondary_tenant.enabled",
		).Scan(&val)

		if val == "false" {
			__antithesis_instrumentation__.Notify(241753)
			return errors.New(
				"waiting for sql.zone_configs.allow_for_secondary_tenant.enabled to be updated",
			)
		} else {
			__antithesis_instrumentation__.Notify(241754)
		}
		__antithesis_instrumentation__.Notify(241752)
		return nil
	})
}

func (h *Handle) LookupTenant(tenantID roachpb.TenantID) (_ *Tenant, found bool) {
	__antithesis_instrumentation__.Notify(241755)
	s, ok := h.ts[tenantID]
	return s, ok
}

func (h *Handle) Tenants() []*Tenant {
	__antithesis_instrumentation__.Notify(241756)
	ts := make([]*Tenant, 0, len(h.ts))
	for _, tenantState := range h.ts {
		__antithesis_instrumentation__.Notify(241758)
		ts = append(ts, tenantState)
	}
	__antithesis_instrumentation__.Notify(241757)
	return ts
}

func (h *Handle) Cleanup() {
	__antithesis_instrumentation__.Notify(241759)
	for _, tenantState := range h.ts {
		__antithesis_instrumentation__.Notify(241761)
		tenantState.cleanup()
	}
	__antithesis_instrumentation__.Notify(241760)
	h.ts = nil
}
