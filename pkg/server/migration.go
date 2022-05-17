package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
)

type migrationServer struct {
	server *Server

	syncutil.Mutex
}

var _ serverpb.MigrationServer = &migrationServer{}

func (m *migrationServer) ValidateTargetClusterVersion(
	ctx context.Context, req *serverpb.ValidateTargetClusterVersionRequest,
) (*serverpb.ValidateTargetClusterVersionResponse, error) {
	__antithesis_instrumentation__.Notify(194295)
	ctx, span := m.server.AnnotateCtxWithSpan(ctx, "validate-cluster-version")
	defer span.Finish()
	ctx = logtags.AddTag(ctx, "validate-cluster-version", nil)

	targetCV := req.ClusterVersion
	versionSetting := m.server.ClusterSettings().Version

	if targetCV.Less(versionSetting.BinaryMinSupportedVersion()) {
		__antithesis_instrumentation__.Notify(194298)
		msg := fmt.Sprintf("target cluster version %s less than binary's min supported version %s",
			targetCV, versionSetting.BinaryMinSupportedVersion())
		log.Warningf(ctx, "%s", msg)
		return nil, errors.Newf("%s", redact.Safe(msg))
	} else {
		__antithesis_instrumentation__.Notify(194299)
	}
	__antithesis_instrumentation__.Notify(194296)

	if versionSetting.BinaryVersion().Less(targetCV.Version) {
		__antithesis_instrumentation__.Notify(194300)
		msg := fmt.Sprintf("binary version %s less than target cluster version %s",
			versionSetting.BinaryVersion(), targetCV)
		log.Warningf(ctx, "%s", msg)
		return nil, errors.Newf("%s", redact.Safe(msg))
	} else {
		__antithesis_instrumentation__.Notify(194301)
	}
	__antithesis_instrumentation__.Notify(194297)

	resp := &serverpb.ValidateTargetClusterVersionResponse{}
	return resp, nil
}

func (m *migrationServer) BumpClusterVersion(
	ctx context.Context, req *serverpb.BumpClusterVersionRequest,
) (*serverpb.BumpClusterVersionResponse, error) {
	__antithesis_instrumentation__.Notify(194302)
	const opName = "bump-cluster-version"
	ctx, span := m.server.AnnotateCtxWithSpan(ctx, opName)
	defer span.Finish()
	ctx = logtags.AddTag(ctx, opName, nil)

	if err := m.server.stopper.RunTaskWithErr(ctx, opName, func(
		ctx context.Context,
	) error {
		__antithesis_instrumentation__.Notify(194304)
		m.Lock()
		defer m.Unlock()
		return bumpClusterVersion(ctx, m.server.st, *req.ClusterVersion, m.server.engines)
	}); err != nil {
		__antithesis_instrumentation__.Notify(194305)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194306)
	}
	__antithesis_instrumentation__.Notify(194303)
	return &serverpb.BumpClusterVersionResponse{}, nil
}

func bumpClusterVersion(
	ctx context.Context, st *cluster.Settings, newCV clusterversion.ClusterVersion, engines Engines,
) error {
	__antithesis_instrumentation__.Notify(194307)

	versionSetting := st.Version
	prevCV, err := kvserver.SynthesizeClusterVersionFromEngines(
		ctx, engines, versionSetting.BinaryVersion(),
		versionSetting.BinaryMinSupportedVersion(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(194312)
		return err
	} else {
		__antithesis_instrumentation__.Notify(194313)
	}
	__antithesis_instrumentation__.Notify(194308)

	if !prevCV.Less(newCV.Version) {
		__antithesis_instrumentation__.Notify(194314)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(194315)
	}
	__antithesis_instrumentation__.Notify(194309)

	if err := kvserver.WriteClusterVersionToEngines(ctx, engines, newCV); err != nil {
		__antithesis_instrumentation__.Notify(194316)
		return err
	} else {
		__antithesis_instrumentation__.Notify(194317)
	}
	__antithesis_instrumentation__.Notify(194310)

	if err := st.Version.SetActiveVersion(ctx, newCV); err != nil {
		__antithesis_instrumentation__.Notify(194318)
		return err
	} else {
		__antithesis_instrumentation__.Notify(194319)
	}
	__antithesis_instrumentation__.Notify(194311)
	log.Infof(ctx, "active cluster version setting is now %s (up from %s)",
		newCV.PrettyPrint(), prevCV.PrettyPrint())
	return nil
}

func (m *migrationServer) SyncAllEngines(
	ctx context.Context, _ *serverpb.SyncAllEnginesRequest,
) (*serverpb.SyncAllEnginesResponse, error) {
	__antithesis_instrumentation__.Notify(194320)
	const opName = "sync-all-engines"
	ctx, span := m.server.AnnotateCtxWithSpan(ctx, opName)
	defer span.Finish()
	ctx = logtags.AddTag(ctx, opName, nil)

	if err := m.server.stopper.RunTaskWithErr(ctx, opName, func(
		ctx context.Context,
	) error {
		__antithesis_instrumentation__.Notify(194322)

		m.server.node.waitForAdditionalStoreInit()

		for _, eng := range m.server.engines {
			__antithesis_instrumentation__.Notify(194324)
			batch := eng.NewBatch()
			if err := batch.LogData(nil); err != nil {
				__antithesis_instrumentation__.Notify(194325)
				return err
			} else {
				__antithesis_instrumentation__.Notify(194326)
			}
		}
		__antithesis_instrumentation__.Notify(194323)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(194327)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194328)
	}
	__antithesis_instrumentation__.Notify(194321)

	log.Infof(ctx, "synced %d engines", len(m.server.engines))
	resp := &serverpb.SyncAllEnginesResponse{}
	return resp, nil
}

func (m *migrationServer) PurgeOutdatedReplicas(
	ctx context.Context, req *serverpb.PurgeOutdatedReplicasRequest,
) (*serverpb.PurgeOutdatedReplicasResponse, error) {
	__antithesis_instrumentation__.Notify(194329)
	const opName = "purged-outdated-replicas"
	ctx, span := m.server.AnnotateCtxWithSpan(ctx, opName)
	defer span.Finish()
	ctx = logtags.AddTag(ctx, opName, nil)

	if err := m.server.stopper.RunTaskWithErr(ctx, opName, func(
		ctx context.Context,
	) error {
		__antithesis_instrumentation__.Notify(194331)

		m.server.node.waitForAdditionalStoreInit()

		return m.server.node.stores.VisitStores(func(s *kvserver.Store) error {
			__antithesis_instrumentation__.Notify(194332)
			return s.PurgeOutdatedReplicas(ctx, *req.Version)
		})
	}); err != nil {
		__antithesis_instrumentation__.Notify(194333)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194334)
	}
	__antithesis_instrumentation__.Notify(194330)

	resp := &serverpb.PurgeOutdatedReplicasResponse{}
	return resp, nil
}

func (m *migrationServer) WaitForSpanConfigSubscription(
	ctx context.Context, _ *serverpb.WaitForSpanConfigSubscriptionRequest,
) (*serverpb.WaitForSpanConfigSubscriptionResponse, error) {
	__antithesis_instrumentation__.Notify(194335)
	const opName = "wait-for-spanconfig-subscription"
	ctx, span := m.server.AnnotateCtxWithSpan(ctx, opName)
	defer span.Finish()
	ctx = logtags.AddTag(ctx, opName, nil)

	if err := m.server.stopper.RunTaskWithErr(ctx, opName, func(
		ctx context.Context,
	) error {
		__antithesis_instrumentation__.Notify(194337)

		m.server.node.waitForAdditionalStoreInit()

		return m.server.node.stores.VisitStores(func(s *kvserver.Store) error {
			__antithesis_instrumentation__.Notify(194338)
			return s.WaitForSpanConfigSubscription(ctx)
		})
	}); err != nil {
		__antithesis_instrumentation__.Notify(194339)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194340)
	}
	__antithesis_instrumentation__.Notify(194336)

	resp := &serverpb.WaitForSpanConfigSubscriptionResponse{}
	return resp, nil
}
