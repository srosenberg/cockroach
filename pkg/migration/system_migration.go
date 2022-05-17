package migration

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/logtags"
)

type Cluster interface {
	NumNodes(ctx context.Context) (int, error)

	ForEveryNode(
		ctx context.Context,
		op string,
		fn func(context.Context, serverpb.MigrationClient) error,
	) error

	UntilClusterStable(ctx context.Context, fn func() error) error

	IterateRangeDescriptors(
		ctx context.Context,
		size int,
		init func(),
		f func(descriptors ...roachpb.RangeDescriptor) error,
	) error
}

type SystemDeps struct {
	Cluster    Cluster
	DB         *kv.DB
	DistSender *kvcoord.DistSender
	Stopper    *stop.Stopper
}

type SystemMigration struct {
	migration
	fn SystemMigrationFunc
}

type SystemMigrationFunc func(context.Context, clusterversion.ClusterVersion, SystemDeps, *jobs.Job) error

func NewSystemMigration(
	description string, cv clusterversion.ClusterVersion, fn SystemMigrationFunc,
) *SystemMigration {
	__antithesis_instrumentation__.Notify(128751)
	return &SystemMigration{
		migration: migration{
			description: description,
			cv:          cv,
		},
		fn: fn,
	}
}

func (m *SystemMigration) Run(
	ctx context.Context, cv clusterversion.ClusterVersion, d SystemDeps, job *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128752)
	ctx = logtags.AddTag(ctx, fmt.Sprintf("migration=%s", cv), nil)
	return m.fn(ctx, cv, d, job)
}
