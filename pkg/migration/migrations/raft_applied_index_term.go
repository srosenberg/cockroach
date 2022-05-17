package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

const defaultPageSize = 200

func raftAppliedIndexTermMigration(
	ctx context.Context, cv clusterversion.ClusterVersion, deps migration.SystemDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128563)
	var batchIdx, numMigratedRanges int
	init := func() { __antithesis_instrumentation__.Notify(128566); batchIdx, numMigratedRanges = 1, 0 }
	__antithesis_instrumentation__.Notify(128564)
	if err := deps.Cluster.IterateRangeDescriptors(ctx, defaultPageSize, init, func(descriptors ...roachpb.RangeDescriptor) error {
		__antithesis_instrumentation__.Notify(128567)
		for _, desc := range descriptors {
			__antithesis_instrumentation__.Notify(128569)

			start, end := desc.StartKey, desc.EndKey
			if bytes.Compare(desc.StartKey, keys.LocalMax) < 0 {
				__antithesis_instrumentation__.Notify(128571)
				start, _ = keys.Addr(keys.LocalMax)
			} else {
				__antithesis_instrumentation__.Notify(128572)
			}
			__antithesis_instrumentation__.Notify(128570)
			if err := deps.DB.Migrate(ctx, start, end, cv.Version); err != nil {
				__antithesis_instrumentation__.Notify(128573)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128574)
			}
		}
		__antithesis_instrumentation__.Notify(128568)

		numMigratedRanges += len(descriptors)
		log.Infof(ctx, "[batch %d/??] migrated %d ranges", batchIdx, numMigratedRanges)
		batchIdx++

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(128575)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128576)
	}
	__antithesis_instrumentation__.Notify(128565)

	log.Infof(ctx, "[batch %d/%d] migrated %d ranges", batchIdx, batchIdx, numMigratedRanges)

	req := &serverpb.SyncAllEnginesRequest{}
	op := "flush-stores"
	return deps.Cluster.ForEveryNode(ctx, op, func(ctx context.Context, client serverpb.MigrationClient) error {
		__antithesis_instrumentation__.Notify(128577)
		_, err := client.SyncAllEngines(ctx, req)
		return err
	})
}

func postRaftAppliedIndexTermMigration(
	ctx context.Context, cv clusterversion.ClusterVersion, deps migration.SystemDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128578)

	truncStateVersion := clusterversion.ByKey(clusterversion.AddRaftAppliedIndexTermMigration)
	req := &serverpb.PurgeOutdatedReplicasRequest{Version: &truncStateVersion}
	op := fmt.Sprintf("purge-outdated-replicas=%s", req.Version)
	return deps.Cluster.ForEveryNode(ctx, op, func(ctx context.Context, client serverpb.MigrationClient) error {
		__antithesis_instrumentation__.Notify(128579)
		_, err := client.PurgeOutdatedReplicas(ctx, req)
		return err
	})
}
