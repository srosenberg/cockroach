package migrationcluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/errors"
)

type TenantCluster struct {
	db *kv.DB
}

func NewTenantCluster(db *kv.DB) *TenantCluster {
	__antithesis_instrumentation__.Notify(128235)
	return &TenantCluster{db: db}
}

func (t *TenantCluster) NumNodes(ctx context.Context) (int, error) {
	__antithesis_instrumentation__.Notify(128236)
	return 0, errors.AssertionFailedf("non-system tenants cannot iterate nodes")
}

func (t *TenantCluster) ForEveryNode(
	ctx context.Context, op string, fn func(context.Context, serverpb.MigrationClient) error,
) error {
	__antithesis_instrumentation__.Notify(128237)
	return errors.AssertionFailedf("non-system tenants cannot iterate nodes")
}

func (t TenantCluster) UntilClusterStable(ctx context.Context, fn func() error) error {
	__antithesis_instrumentation__.Notify(128238)
	return nil
}

func (t TenantCluster) IterateRangeDescriptors(
	ctx context.Context, size int, init func(), f func(descriptors ...roachpb.RangeDescriptor) error,
) error {
	__antithesis_instrumentation__.Notify(128239)
	return errors.AssertionFailedf("non-system tenants cannot iterate ranges")
}
