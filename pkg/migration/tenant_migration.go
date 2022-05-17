package migration

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/logtags"
)

type TenantDeps struct {
	DB                *kv.DB
	Codec             keys.SQLCodec
	Settings          *cluster.Settings
	CollectionFactory *descs.CollectionFactory
	LeaseManager      *lease.Manager
	InternalExecutor  sqlutil.InternalExecutor

	SpanConfig struct {
		spanconfig.KVAccessor
		spanconfig.Splitter
		Default roachpb.SpanConfig
	}

	TestingKnobs *TestingKnobs
}

type TenantMigrationFunc func(context.Context, clusterversion.ClusterVersion, TenantDeps, *jobs.Job) error

type PreconditionFunc func(context.Context, clusterversion.ClusterVersion, TenantDeps) error

type TenantMigration struct {
	migration
	fn           TenantMigrationFunc
	precondition PreconditionFunc
}

var _ Migration = (*TenantMigration)(nil)

func NewTenantMigration(
	description string,
	cv clusterversion.ClusterVersion,
	precondition PreconditionFunc,
	fn TenantMigrationFunc,
) *TenantMigration {
	__antithesis_instrumentation__.Notify(128753)
	m := &TenantMigration{
		migration: migration{
			description: description,
			cv:          cv,
		},
		fn:           fn,
		precondition: precondition,
	}
	return m
}

func (m *TenantMigration) Run(
	ctx context.Context, cv clusterversion.ClusterVersion, d TenantDeps, job *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128754)
	ctx = logtags.AddTag(ctx, fmt.Sprintf("migration=%s", cv), nil)
	return m.fn(ctx, cv, d, job)
}

func (m *TenantMigration) Precondition(
	ctx context.Context, cv clusterversion.ClusterVersion, d TenantDeps,
) error {
	__antithesis_instrumentation__.Notify(128755)
	ctx = logtags.AddTag(ctx, fmt.Sprintf("migration=%s,precondition", cv), nil)
	return m.precondition(ctx, cv, d)
}
