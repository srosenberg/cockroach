package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/loqrecovery"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/loqrecovery/loqrecoverypb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

func logPendingLossOfQuorumRecoveryEvents(ctx context.Context, stores *kvserver.Stores) {
	__antithesis_instrumentation__.Notify(194272)
	if err := stores.VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194273)

		eventCount, err := loqrecovery.RegisterOfflineRecoveryEvents(
			ctx,
			s.Engine(),
			func(ctx context.Context, record loqrecoverypb.ReplicaRecoveryRecord) (bool, error) {
				__antithesis_instrumentation__.Notify(194276)
				event := record.AsStructuredLog()
				log.StructuredEvent(ctx, &event)
				return false, nil
			})
		__antithesis_instrumentation__.Notify(194274)
		if eventCount > 0 {
			__antithesis_instrumentation__.Notify(194277)
			log.Infof(
				ctx, "registered %d loss of quorum replica recovery events for s%d",
				eventCount, s.Ident.StoreID)
		} else {
			__antithesis_instrumentation__.Notify(194278)
		}
		__antithesis_instrumentation__.Notify(194275)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(194279)

		log.Errorf(ctx, "failed to record loss of quorum recovery events: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(194280)
	}
}

func publishPendingLossOfQuorumRecoveryEvents(
	ctx context.Context, stores *kvserver.Stores, stopper *stop.Stopper,
) {
	__antithesis_instrumentation__.Notify(194281)
	_ = stopper.RunAsyncTask(ctx, "publish-loss-of-quorum-events", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(194282)
		if err := stores.VisitStores(func(s *kvserver.Store) error {
			__antithesis_instrumentation__.Notify(194283)
			recoveryEventsSupported := s.ClusterSettings().Version.IsActive(ctx,
				clusterversion.UnsafeLossOfQuorumRecoveryRangeLog)
			_, err := loqrecovery.RegisterOfflineRecoveryEvents(
				ctx,
				s.Engine(),
				func(ctx context.Context, record loqrecoverypb.ReplicaRecoveryRecord) (bool, error) {
					__antithesis_instrumentation__.Notify(194285)
					sqlExec := func(ctx context.Context, stmt string, args ...interface{}) (int, error) {
						__antithesis_instrumentation__.Notify(194288)
						return s.GetStoreConfig().SQLExecutor.ExecEx(ctx, "", nil,
							sessiondata.InternalExecutorOverride{User: security.RootUserName()}, stmt, args...)
					}
					__antithesis_instrumentation__.Notify(194286)
					if recoveryEventsSupported {
						__antithesis_instrumentation__.Notify(194289)
						if err := loqrecovery.UpdateRangeLogWithRecovery(ctx, sqlExec, record); err != nil {
							__antithesis_instrumentation__.Notify(194290)
							return false, errors.Wrap(err,
								"loss of quorum recovery failed to write RangeLog entry")
						} else {
							__antithesis_instrumentation__.Notify(194291)
						}
					} else {
						__antithesis_instrumentation__.Notify(194292)
					}
					__antithesis_instrumentation__.Notify(194287)

					s.Metrics().RangeLossOfQuorumRecoveries.Inc(1)
					return true, nil
				})
			__antithesis_instrumentation__.Notify(194284)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(194293)

			log.Errorf(ctx, "failed to update range log with loss of quorum recovery events: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(194294)
		}
	})
}
