// Package migrationcluster provides implementations of migration.Cluster.
package migrationcluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"google.golang.org/grpc"
)

type Cluster struct {
	c ClusterConfig
}

type ClusterConfig struct {
	NodeLiveness NodeLiveness

	Dialer NodeDialer

	DB *kv.DB
}

type NodeDialer interface {
	Dial(context.Context, roachpb.NodeID, rpc.ConnectionClass) (*grpc.ClientConn, error)
}

type NodeLiveness interface {
	GetLivenessesFromKV(context.Context) ([]livenesspb.Liveness, error)
	IsLive(roachpb.NodeID) (bool, error)
}

func New(cfg ClusterConfig) *Cluster {
	__antithesis_instrumentation__.Notify(128149)
	return &Cluster{c: cfg}
}

func (c *Cluster) UntilClusterStable(ctx context.Context, fn func() error) error {
	__antithesis_instrumentation__.Notify(128150)
	ns, err := NodesFromNodeLiveness(ctx, c.c.NodeLiveness)
	if err != nil {
		__antithesis_instrumentation__.Notify(128153)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128154)
	}
	__antithesis_instrumentation__.Notify(128151)

	for {
		__antithesis_instrumentation__.Notify(128155)
		if err := fn(); err != nil {
			__antithesis_instrumentation__.Notify(128159)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128160)
		}
		__antithesis_instrumentation__.Notify(128156)
		curNodes, err := NodesFromNodeLiveness(ctx, c.c.NodeLiveness)
		if err != nil {
			__antithesis_instrumentation__.Notify(128161)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128162)
		}
		__antithesis_instrumentation__.Notify(128157)

		if ok, diffs := ns.Identical(curNodes); !ok {
			__antithesis_instrumentation__.Notify(128163)
			log.Infof(ctx, "%s, retrying", diffs)
			ns = curNodes
			continue
		} else {
			__antithesis_instrumentation__.Notify(128164)
		}
		__antithesis_instrumentation__.Notify(128158)
		break
	}
	__antithesis_instrumentation__.Notify(128152)
	return nil
}

func (c *Cluster) NumNodes(ctx context.Context) (int, error) {
	__antithesis_instrumentation__.Notify(128165)
	ns, err := NodesFromNodeLiveness(ctx, c.c.NodeLiveness)
	if err != nil {
		__antithesis_instrumentation__.Notify(128167)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(128168)
	}
	__antithesis_instrumentation__.Notify(128166)
	return len(ns), nil
}

func (c *Cluster) ForEveryNode(
	ctx context.Context, op string, fn func(context.Context, serverpb.MigrationClient) error,
) error {
	__antithesis_instrumentation__.Notify(128169)

	ns, err := NodesFromNodeLiveness(ctx, c.c.NodeLiveness)
	if err != nil {
		__antithesis_instrumentation__.Notify(128172)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128173)
	}
	__antithesis_instrumentation__.Notify(128170)

	qp := quotapool.NewIntPool("every-node", 25)
	log.Infof(ctx, "executing %s on nodes %s", redact.Safe(op), ns)
	grp := ctxgroup.WithContext(ctx)

	for _, node := range ns {
		__antithesis_instrumentation__.Notify(128174)
		id := node.ID
		alloc, err := qp.Acquire(ctx, 1)
		if err != nil {
			__antithesis_instrumentation__.Notify(128176)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128177)
		}
		__antithesis_instrumentation__.Notify(128175)

		grp.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(128178)
			defer alloc.Release()

			conn, err := c.c.Dialer.Dial(ctx, id, rpc.DefaultClass)
			if err != nil {
				__antithesis_instrumentation__.Notify(128180)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128181)
			}
			__antithesis_instrumentation__.Notify(128179)
			client := serverpb.NewMigrationClient(conn)
			return fn(ctx, client)
		})
	}
	__antithesis_instrumentation__.Notify(128171)
	return grp.Wait()
}

func (c *Cluster) IterateRangeDescriptors(
	ctx context.Context, blockSize int, init func(), fn func(...roachpb.RangeDescriptor) error,
) error {
	__antithesis_instrumentation__.Notify(128182)
	return c.c.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(128183)

		init()

		var lastRangeIDInMeta1 roachpb.RangeID
		return txn.Iterate(ctx, keys.MetaMin, keys.MetaMax, blockSize,
			func(rows []kv.KeyValue) error {
				__antithesis_instrumentation__.Notify(128184)
				descriptors := make([]roachpb.RangeDescriptor, 0, len(rows))
				var desc roachpb.RangeDescriptor
				for _, row := range rows {
					__antithesis_instrumentation__.Notify(128186)
					err := row.ValueProto(&desc)
					if err != nil {
						__antithesis_instrumentation__.Notify(128189)
						return errors.Wrapf(err, "unable to unmarshal range descriptor from %s", row.Key)
					} else {
						__antithesis_instrumentation__.Notify(128190)
					}
					__antithesis_instrumentation__.Notify(128187)

					if desc.RangeID == lastRangeIDInMeta1 {
						__antithesis_instrumentation__.Notify(128191)
						continue
					} else {
						__antithesis_instrumentation__.Notify(128192)
					}
					__antithesis_instrumentation__.Notify(128188)

					descriptors = append(descriptors, desc)
					if keys.InMeta1(keys.RangeMetaKey(desc.StartKey)) {
						__antithesis_instrumentation__.Notify(128193)
						lastRangeIDInMeta1 = desc.RangeID
					} else {
						__antithesis_instrumentation__.Notify(128194)
					}
				}
				__antithesis_instrumentation__.Notify(128185)

				return fn(descriptors...)
			})
	})
}
