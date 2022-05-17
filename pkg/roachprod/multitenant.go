package roachprod

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/errors"
)

func StartTenant(
	ctx context.Context,
	l *logger.Logger,
	tenantCluster string,
	hostCluster string,
	startOpts install.StartOpts,
	clusterSettingsOpts ...install.ClusterSettingOption,
) error {
	__antithesis_instrumentation__.Notify(181881)
	tc, err := newCluster(l, tenantCluster, clusterSettingsOpts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(181889)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181890)
	}
	__antithesis_instrumentation__.Notify(181882)

	hc, err := newCluster(l, hostCluster, clusterSettingsOpts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(181891)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181892)
	}
	__antithesis_instrumentation__.Notify(181883)

	if tc.Name == hc.Name {
		__antithesis_instrumentation__.Notify(181893)

		for _, n1 := range tc.Nodes {
			__antithesis_instrumentation__.Notify(181894)
			for _, n2 := range hc.Nodes {
				__antithesis_instrumentation__.Notify(181895)
				if n1 == n2 {
					__antithesis_instrumentation__.Notify(181896)
					return errors.Errorf("host and tenant nodes must be disjoint")
				} else {
					__antithesis_instrumentation__.Notify(181897)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(181898)
	}
	__antithesis_instrumentation__.Notify(181884)

	if tc.Secure {
		__antithesis_instrumentation__.Notify(181899)

		return errors.Errorf("secure mode not implemented for tenants yet")
	} else {
		__antithesis_instrumentation__.Notify(181900)
	}
	__antithesis_instrumentation__.Notify(181885)

	startOpts.Target = install.StartTenantSQL
	if startOpts.TenantID < 2 {
		__antithesis_instrumentation__.Notify(181901)
		return errors.Errorf("invalid tenant ID %d (must be 2 or higher)", startOpts.TenantID)
	} else {
		__antithesis_instrumentation__.Notify(181902)
	}
	__antithesis_instrumentation__.Notify(181886)

	saveNodes := hc.Nodes
	hc.Nodes = hc.Nodes[:1]
	l.Printf("Creating tenant metadata")
	if err := hc.RunSQL(ctx, l, []string{
		`-e`,
		fmt.Sprintf(createTenantIfNotExistsQuery, startOpts.TenantID),
	}); err != nil {
		__antithesis_instrumentation__.Notify(181903)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181904)
	}
	__antithesis_instrumentation__.Notify(181887)
	hc.Nodes = saveNodes

	var kvAddrs []string
	for _, node := range hc.Nodes {
		__antithesis_instrumentation__.Notify(181905)
		kvAddrs = append(kvAddrs, fmt.Sprintf("%s:%d", hc.Host(node), hc.NodePort(node)))
	}
	__antithesis_instrumentation__.Notify(181888)
	startOpts.KVAddrs = strings.Join(kvAddrs, ",")
	return tc.Start(ctx, l, startOpts)
}

const createTenantIfNotExistsQuery = `
SELECT
  CASE (SELECT 1 FROM system.tenants WHERE id = %[1]d) IS NULL
  WHEN true
  THEN (
    crdb_internal.create_tenant(%[1]d),
    crdb_internal.update_tenant_resource_limits(%[1]d, 1000000000, 10000, 0, now(), 0)
  )::STRING
  ELSE 'already exists'
  END;`
