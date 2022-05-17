package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

func (r RemoteOffset) measuredAt() time.Time {
	__antithesis_instrumentation__.Notify(184720)
	return timeutil.Unix(0, r.MeasuredAt)
}

func (r RemoteOffset) String() string {
	__antithesis_instrumentation__.Notify(184721)
	return fmt.Sprintf("off=%s, err=%s, at=%s", time.Duration(r.Offset), time.Duration(r.Uncertainty), r.measuredAt())
}

type HeartbeatService struct {
	clock *hlc.Clock

	remoteClockMonitor *RemoteClockMonitor

	clusterID *base.ClusterIDContainer
	nodeID    *base.NodeIDContainer
	settings  *cluster.Settings

	clusterName                    string
	disableClusterNameVerification bool

	onHandlePing func(context.Context, *PingRequest) error

	testingAllowNamedRPCToAnonymousServer bool
}

func checkClusterName(clusterName string, peerName string) error {
	__antithesis_instrumentation__.Notify(184722)
	if clusterName != peerName {
		__antithesis_instrumentation__.Notify(184724)
		var err error
		if clusterName == "" {
			__antithesis_instrumentation__.Notify(184726)
			err = errors.Errorf("peer node expects cluster name %q, use --cluster-name to configure", peerName)
		} else {
			__antithesis_instrumentation__.Notify(184727)
			if peerName == "" {
				__antithesis_instrumentation__.Notify(184728)
				err = errors.New("peer node does not have a cluster name configured, cannot use --cluster-name")
			} else {
				__antithesis_instrumentation__.Notify(184729)
				err = errors.Errorf(
					"local cluster name %q does not match peer cluster name %q", clusterName, peerName)
			}
		}
		__antithesis_instrumentation__.Notify(184725)
		log.Ops.Shoutf(context.Background(), severity.ERROR, "%v", err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(184730)
	}
	__antithesis_instrumentation__.Notify(184723)
	return nil
}

func checkVersion(ctx context.Context, st *cluster.Settings, peerVersion roachpb.Version) error {
	__antithesis_instrumentation__.Notify(184731)
	activeVersion := st.Version.ActiveVersionOrEmpty(ctx)
	if activeVersion == (clusterversion.ClusterVersion{}) {
		__antithesis_instrumentation__.Notify(184736)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(184737)
	}
	__antithesis_instrumentation__.Notify(184732)
	if peerVersion == (roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(184738)
		return errors.Errorf(
			"cluster requires at least version %s, but peer did not provide a version", activeVersion)
	} else {
		__antithesis_instrumentation__.Notify(184739)
	}
	__antithesis_instrumentation__.Notify(184733)

	minVersion := activeVersion.Version
	if tenantID, isTenant := roachpb.TenantFromContext(ctx); isTenant && func() bool {
		__antithesis_instrumentation__.Notify(184740)
		return !roachpb.IsSystemTenantID(tenantID.ToUint64()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(184741)
		minVersion = st.Version.BinaryMinSupportedVersion()
	} else {
		__antithesis_instrumentation__.Notify(184742)
	}
	__antithesis_instrumentation__.Notify(184734)
	if peerVersion.Less(minVersion) {
		__antithesis_instrumentation__.Notify(184743)
		return errors.Errorf(
			"cluster requires at least version %s, but peer has version %s",
			minVersion, peerVersion)
	} else {
		__antithesis_instrumentation__.Notify(184744)
	}
	__antithesis_instrumentation__.Notify(184735)
	return nil
}

func (hs *HeartbeatService) Ping(ctx context.Context, args *PingRequest) (*PingResponse, error) {
	__antithesis_instrumentation__.Notify(184745)
	if log.ExpensiveLogEnabled(ctx, 2) {
		__antithesis_instrumentation__.Notify(184753)
		log.Dev.Infof(ctx, "received heartbeat: %+v vs local cluster %+v node %+v", args, hs.clusterID, hs.nodeID)
	} else {
		__antithesis_instrumentation__.Notify(184754)
	}
	__antithesis_instrumentation__.Notify(184746)

	clusterID := hs.clusterID.Get()
	if args.ClusterID != nil && func() bool {
		__antithesis_instrumentation__.Notify(184755)
		return *args.ClusterID != uuid.Nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(184756)
		return clusterID != uuid.Nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(184757)

		if *args.ClusterID != clusterID {
			__antithesis_instrumentation__.Notify(184758)
			return nil, errors.Errorf(
				"client cluster ID %q doesn't match server cluster ID %q", args.ClusterID, clusterID)
		} else {
			__antithesis_instrumentation__.Notify(184759)
		}
	} else {
		__antithesis_instrumentation__.Notify(184760)
	}
	__antithesis_instrumentation__.Notify(184747)

	var nodeID roachpb.NodeID
	if hs.nodeID != nil {
		__antithesis_instrumentation__.Notify(184761)
		nodeID = hs.nodeID.Get()
	} else {
		__antithesis_instrumentation__.Notify(184762)
	}
	__antithesis_instrumentation__.Notify(184748)
	if args.TargetNodeID != 0 && func() bool {
		__antithesis_instrumentation__.Notify(184763)
		return (!hs.testingAllowNamedRPCToAnonymousServer || func() bool {
			__antithesis_instrumentation__.Notify(184764)
			return nodeID != 0 == true
		}() == true) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(184765)
		return args.TargetNodeID != nodeID == true
	}() == true {
		__antithesis_instrumentation__.Notify(184766)

		return nil, errors.Errorf(
			"client requested node ID %d doesn't match server node ID %d", args.TargetNodeID, nodeID)
	} else {
		__antithesis_instrumentation__.Notify(184767)
	}
	__antithesis_instrumentation__.Notify(184749)

	if err := checkVersion(ctx, hs.settings, args.ServerVersion); err != nil {
		__antithesis_instrumentation__.Notify(184768)
		return nil, errors.Wrap(err, "version compatibility check failed on ping request")
	} else {
		__antithesis_instrumentation__.Notify(184769)
	}
	__antithesis_instrumentation__.Notify(184750)

	mo, amo := hs.clock.MaxOffset(), time.Duration(args.OriginMaxOffsetNanos)
	if mo != 0 && func() bool {
		__antithesis_instrumentation__.Notify(184770)
		return amo != 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(184771)
		return mo != amo == true
	}() == true {
		__antithesis_instrumentation__.Notify(184772)
		log.Fatalf(ctx, "locally configured maximum clock offset (%s) "+
			"does not match that of node %s (%s)", mo, args.OriginAddr, amo)
	} else {
		__antithesis_instrumentation__.Notify(184773)
	}
	__antithesis_instrumentation__.Notify(184751)

	if fn := hs.onHandlePing; fn != nil {
		__antithesis_instrumentation__.Notify(184774)
		if err := fn(ctx, args); err != nil {
			__antithesis_instrumentation__.Notify(184775)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(184776)
		}
	} else {
		__antithesis_instrumentation__.Notify(184777)
	}
	__antithesis_instrumentation__.Notify(184752)

	serverOffset := args.Offset

	serverOffset.Offset = -serverOffset.Offset
	hs.remoteClockMonitor.UpdateOffset(ctx, args.OriginAddr, serverOffset, 0)
	return &PingResponse{
		Pong:                           args.Ping,
		ServerTime:                     hs.clock.PhysicalNow(),
		ServerVersion:                  hs.settings.Version.BinaryVersion(),
		ClusterName:                    hs.clusterName,
		DisableClusterNameVerification: hs.disableClusterNameVerification,
	}, nil
}
