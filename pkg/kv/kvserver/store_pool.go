package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/shuffle"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

const (
	TestTimeUntilStoreDead = 5 * time.Millisecond

	TestTimeUntilStoreDeadOff = 24 * time.Hour
)

var FailedReservationsTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"server.failed_reservation_timeout",
	"the amount of time to consider the store throttled for up-replication after a failed reservation call",
	5*time.Second,
	settings.NonNegativeDuration,
)

const timeAfterStoreSuspectSettingName = "server.time_after_store_suspect"

var TimeAfterStoreSuspect = settings.RegisterDurationSetting(
	settings.TenantWritable,
	timeAfterStoreSuspectSettingName,
	"the amount of time we consider a store suspect for after it fails a node liveness heartbeat."+
		" A suspect node would not receive any new replicas or lease transfers, but will keep the replicas it has.",
	30*time.Second,
	settings.NonNegativeDuration,
	func(v time.Duration) error {
		__antithesis_instrumentation__.Notify(125039)

		const maxTimeAfterStoreSuspect = 5 * time.Minute
		if v > maxTimeAfterStoreSuspect {
			__antithesis_instrumentation__.Notify(125041)
			return errors.Errorf("cannot set %s to more than %v: %v",
				timeAfterStoreSuspectSettingName, maxTimeAfterStoreSuspect, v)
		} else {
			__antithesis_instrumentation__.Notify(125042)
		}
		__antithesis_instrumentation__.Notify(125040)
		return nil
	},
)

const timeUntilStoreDeadSettingName = "server.time_until_store_dead"

var TimeUntilStoreDead = func() *settings.DurationSetting {
	__antithesis_instrumentation__.Notify(125043)
	s := settings.RegisterDurationSetting(
		settings.TenantWritable,
		timeUntilStoreDeadSettingName,
		"the time after which if there is no new gossiped information about a store, it is considered dead",
		5*time.Minute,
		func(v time.Duration) error {
			__antithesis_instrumentation__.Notify(125045)

			const minTimeUntilStoreDead = gossip.StoresInterval + 15*time.Second
			if v < minTimeUntilStoreDead {
				__antithesis_instrumentation__.Notify(125047)
				return errors.Errorf("cannot set %s to less than %v: %v",
					timeUntilStoreDeadSettingName, minTimeUntilStoreDead, v)
			} else {
				__antithesis_instrumentation__.Notify(125048)
			}
			__antithesis_instrumentation__.Notify(125046)
			return nil
		},
	)
	__antithesis_instrumentation__.Notify(125044)
	s.SetVisibility(settings.Public)
	return s
}()

type NodeCountFunc func() int

type NodeLivenessFunc func(
	nid roachpb.NodeID, now time.Time, timeUntilStoreDead time.Duration,
) livenesspb.NodeLivenessStatus

func MakeStorePoolNodeLivenessFunc(nodeLiveness *liveness.NodeLiveness) NodeLivenessFunc {
	__antithesis_instrumentation__.Notify(125049)
	return func(
		nodeID roachpb.NodeID, now time.Time, timeUntilStoreDead time.Duration,
	) livenesspb.NodeLivenessStatus {
		__antithesis_instrumentation__.Notify(125050)
		liveness, ok := nodeLiveness.GetLiveness(nodeID)
		if !ok {
			__antithesis_instrumentation__.Notify(125052)
			return livenesspb.NodeLivenessStatus_UNKNOWN
		} else {
			__antithesis_instrumentation__.Notify(125053)
		}
		__antithesis_instrumentation__.Notify(125051)
		return LivenessStatus(liveness.Liveness, now, timeUntilStoreDead)
	}
}

func LivenessStatus(
	l livenesspb.Liveness, now time.Time, deadThreshold time.Duration,
) livenesspb.NodeLivenessStatus {
	__antithesis_instrumentation__.Notify(125054)
	if l.IsDead(now, deadThreshold) {
		__antithesis_instrumentation__.Notify(125057)
		if !l.Membership.Active() {
			__antithesis_instrumentation__.Notify(125059)
			return livenesspb.NodeLivenessStatus_DECOMMISSIONED
		} else {
			__antithesis_instrumentation__.Notify(125060)
		}
		__antithesis_instrumentation__.Notify(125058)
		return livenesspb.NodeLivenessStatus_DEAD
	} else {
		__antithesis_instrumentation__.Notify(125061)
	}
	__antithesis_instrumentation__.Notify(125055)
	if l.IsLive(now) {
		__antithesis_instrumentation__.Notify(125062)
		if !l.Membership.Active() {
			__antithesis_instrumentation__.Notify(125065)
			return livenesspb.NodeLivenessStatus_DECOMMISSIONING
		} else {
			__antithesis_instrumentation__.Notify(125066)
		}
		__antithesis_instrumentation__.Notify(125063)
		if l.Draining {
			__antithesis_instrumentation__.Notify(125067)
			return livenesspb.NodeLivenessStatus_DRAINING
		} else {
			__antithesis_instrumentation__.Notify(125068)
		}
		__antithesis_instrumentation__.Notify(125064)
		return livenesspb.NodeLivenessStatus_LIVE
	} else {
		__antithesis_instrumentation__.Notify(125069)
	}
	__antithesis_instrumentation__.Notify(125056)
	return livenesspb.NodeLivenessStatus_UNAVAILABLE
}

type storeDetail struct {
	desc *roachpb.StoreDescriptor

	throttledUntil time.Time

	throttledBecause string

	lastUpdatedTime time.Time

	lastUnavailable time.Time

	lastAvailable time.Time
}

func (sd storeDetail) isThrottled(now time.Time) bool {
	__antithesis_instrumentation__.Notify(125070)
	return sd.throttledUntil.After(now)
}

func (sd storeDetail) isSuspect(now time.Time, suspectDuration time.Duration) bool {
	__antithesis_instrumentation__.Notify(125071)
	return sd.lastUnavailable.Add(suspectDuration).After(now)
}

type storeStatus int

const (
	_ storeStatus = iota

	storeStatusDead

	storeStatusUnknown

	storeStatusThrottled

	storeStatusAvailable

	storeStatusDecommissioning

	storeStatusSuspect

	storeStatusDraining
)

func (sd *storeDetail) status(
	now time.Time, threshold time.Duration, nl NodeLivenessFunc, suspectDuration time.Duration,
) storeStatus {
	__antithesis_instrumentation__.Notify(125072)

	deadAsOf := sd.lastUpdatedTime.Add(threshold)
	if now.After(deadAsOf) {
		__antithesis_instrumentation__.Notify(125078)

		sd.lastAvailable = time.Time{}
		return storeStatusDead
	} else {
		__antithesis_instrumentation__.Notify(125079)
	}
	__antithesis_instrumentation__.Notify(125073)

	if sd.desc == nil {
		__antithesis_instrumentation__.Notify(125080)
		return storeStatusUnknown
	} else {
		__antithesis_instrumentation__.Notify(125081)
	}
	__antithesis_instrumentation__.Notify(125074)

	switch nl(sd.desc.Node.NodeID, now, threshold) {
	case livenesspb.NodeLivenessStatus_DEAD, livenesspb.NodeLivenessStatus_DECOMMISSIONED:
		__antithesis_instrumentation__.Notify(125082)
		return storeStatusDead
	case livenesspb.NodeLivenessStatus_DECOMMISSIONING:
		__antithesis_instrumentation__.Notify(125083)
		return storeStatusDecommissioning
	case livenesspb.NodeLivenessStatus_UNAVAILABLE:
		__antithesis_instrumentation__.Notify(125084)

		if !sd.lastAvailable.IsZero() {
			__antithesis_instrumentation__.Notify(125089)
			sd.lastUnavailable = now
		} else {
			__antithesis_instrumentation__.Notify(125090)
		}
		__antithesis_instrumentation__.Notify(125085)
		return storeStatusUnknown
	case livenesspb.NodeLivenessStatus_UNKNOWN:
		__antithesis_instrumentation__.Notify(125086)
		return storeStatusUnknown
	case livenesspb.NodeLivenessStatus_DRAINING:
		__antithesis_instrumentation__.Notify(125087)

		sd.lastAvailable = time.Time{}
		return storeStatusDraining
	default:
		__antithesis_instrumentation__.Notify(125088)
	}
	__antithesis_instrumentation__.Notify(125075)

	if sd.isThrottled(now) {
		__antithesis_instrumentation__.Notify(125091)
		return storeStatusThrottled
	} else {
		__antithesis_instrumentation__.Notify(125092)
	}
	__antithesis_instrumentation__.Notify(125076)

	if sd.isSuspect(now, suspectDuration) {
		__antithesis_instrumentation__.Notify(125093)
		return storeStatusSuspect
	} else {
		__antithesis_instrumentation__.Notify(125094)
	}
	__antithesis_instrumentation__.Notify(125077)
	sd.lastAvailable = now
	return storeStatusAvailable
}

type localityWithString struct {
	locality roachpb.Locality
	str      string
}

type StorePool struct {
	log.AmbientContext
	st *cluster.Settings

	clock          *hlc.Clock
	gossip         *gossip.Gossip
	nodeCountFn    NodeCountFunc
	nodeLivenessFn NodeLivenessFunc
	startTime      time.Time
	deterministic  bool

	detailsMu struct {
		syncutil.RWMutex
		storeDetails map[roachpb.StoreID]*storeDetail
	}
	localitiesMu struct {
		syncutil.RWMutex
		nodeLocalities map[roachpb.NodeID]localityWithString
	}

	isStoreReadyForRoutineReplicaTransfer func(context.Context, roachpb.StoreID) bool
}

func NewStorePool(
	ambient log.AmbientContext,
	st *cluster.Settings,
	g *gossip.Gossip,
	clock *hlc.Clock,
	nodeCountFn NodeCountFunc,
	nodeLivenessFn NodeLivenessFunc,
	deterministic bool,
) *StorePool {
	__antithesis_instrumentation__.Notify(125095)
	sp := &StorePool{
		AmbientContext: ambient,
		st:             st,
		clock:          clock,
		gossip:         g,
		nodeCountFn:    nodeCountFn,
		nodeLivenessFn: nodeLivenessFn,
		startTime:      clock.PhysicalTime(),
		deterministic:  deterministic,
	}
	sp.isStoreReadyForRoutineReplicaTransfer = sp.isStoreReadyForRoutineReplicaTransferInternal
	sp.detailsMu.storeDetails = make(map[roachpb.StoreID]*storeDetail)
	sp.localitiesMu.nodeLocalities = make(map[roachpb.NodeID]localityWithString)

	storeRegex := gossip.MakePrefixPattern(gossip.KeyStorePrefix)
	g.RegisterCallback(storeRegex, sp.storeGossipUpdate, gossip.Redundant)

	return sp
}

func (sp *StorePool) String() string {
	__antithesis_instrumentation__.Notify(125096)
	sp.detailsMu.RLock()
	defer sp.detailsMu.RUnlock()

	ids := make(roachpb.StoreIDSlice, 0, len(sp.detailsMu.storeDetails))
	for id := range sp.detailsMu.storeDetails {
		__antithesis_instrumentation__.Notify(125099)
		ids = append(ids, id)
	}
	__antithesis_instrumentation__.Notify(125097)
	sort.Sort(ids)

	var buf bytes.Buffer
	now := sp.clock.Now().GoTime()
	timeUntilStoreDead := TimeUntilStoreDead.Get(&sp.st.SV)
	timeAfterStoreSuspect := TimeAfterStoreSuspect.Get(&sp.st.SV)

	for _, id := range ids {
		__antithesis_instrumentation__.Notify(125100)
		detail := sp.detailsMu.storeDetails[id]
		fmt.Fprintf(&buf, "%d", id)
		status := detail.status(now, timeUntilStoreDead, sp.nodeLivenessFn, timeAfterStoreSuspect)
		if status != storeStatusAvailable {
			__antithesis_instrumentation__.Notify(125104)
			fmt.Fprintf(&buf, " (status=%d)", status)
		} else {
			__antithesis_instrumentation__.Notify(125105)
		}
		__antithesis_instrumentation__.Notify(125101)
		if detail.desc != nil {
			__antithesis_instrumentation__.Notify(125106)
			fmt.Fprintf(&buf, ": range-count=%d fraction-used=%.2f",
				detail.desc.Capacity.RangeCount, detail.desc.Capacity.FractionUsed())
		} else {
			__antithesis_instrumentation__.Notify(125107)
		}
		__antithesis_instrumentation__.Notify(125102)
		throttled := detail.throttledUntil.Sub(now)
		if throttled > 0 {
			__antithesis_instrumentation__.Notify(125108)
			fmt.Fprintf(&buf, " [throttled=%.1fs]", throttled.Seconds())
		} else {
			__antithesis_instrumentation__.Notify(125109)
		}
		__antithesis_instrumentation__.Notify(125103)
		_, _ = buf.WriteString("\n")
	}
	__antithesis_instrumentation__.Notify(125098)
	return buf.String()
}

func (sp *StorePool) storeGossipUpdate(_ string, content roachpb.Value) {
	__antithesis_instrumentation__.Notify(125110)
	var storeDesc roachpb.StoreDescriptor
	if err := content.GetProto(&storeDesc); err != nil {
		__antithesis_instrumentation__.Notify(125112)
		ctx := sp.AnnotateCtx(context.TODO())
		log.Errorf(ctx, "%v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(125113)
	}
	__antithesis_instrumentation__.Notify(125111)

	sp.detailsMu.Lock()
	detail := sp.getStoreDetailLocked(storeDesc.StoreID)
	detail.desc = &storeDesc
	detail.lastUpdatedTime = sp.clock.PhysicalTime()
	sp.detailsMu.Unlock()

	sp.localitiesMu.Lock()
	sp.localitiesMu.nodeLocalities[storeDesc.Node.NodeID] =
		localityWithString{storeDesc.Node.Locality, storeDesc.Node.Locality.String()}
	sp.localitiesMu.Unlock()
}

func (sp *StorePool) updateLocalStoreAfterRebalance(
	storeID roachpb.StoreID, rangeUsageInfo RangeUsageInfo, changeType roachpb.ReplicaChangeType,
) {
	__antithesis_instrumentation__.Notify(125114)
	sp.detailsMu.Lock()
	defer sp.detailsMu.Unlock()
	detail := *sp.getStoreDetailLocked(storeID)
	if detail.desc == nil {
		__antithesis_instrumentation__.Notify(125117)

		return
	} else {
		__antithesis_instrumentation__.Notify(125118)
	}
	__antithesis_instrumentation__.Notify(125115)
	switch changeType {
	case roachpb.ADD_VOTER, roachpb.ADD_NON_VOTER:
		__antithesis_instrumentation__.Notify(125119)
		detail.desc.Capacity.RangeCount++
		detail.desc.Capacity.LogicalBytes += rangeUsageInfo.LogicalBytes
		detail.desc.Capacity.WritesPerSecond += rangeUsageInfo.WritesPerSecond
	case roachpb.REMOVE_VOTER, roachpb.REMOVE_NON_VOTER:
		__antithesis_instrumentation__.Notify(125120)
		detail.desc.Capacity.RangeCount--
		if detail.desc.Capacity.LogicalBytes <= rangeUsageInfo.LogicalBytes {
			__antithesis_instrumentation__.Notify(125123)
			detail.desc.Capacity.LogicalBytes = 0
		} else {
			__antithesis_instrumentation__.Notify(125124)
			detail.desc.Capacity.LogicalBytes -= rangeUsageInfo.LogicalBytes
		}
		__antithesis_instrumentation__.Notify(125121)
		if detail.desc.Capacity.WritesPerSecond <= rangeUsageInfo.WritesPerSecond {
			__antithesis_instrumentation__.Notify(125125)
			detail.desc.Capacity.WritesPerSecond = 0
		} else {
			__antithesis_instrumentation__.Notify(125126)
			detail.desc.Capacity.WritesPerSecond -= rangeUsageInfo.WritesPerSecond
		}
	default:
		__antithesis_instrumentation__.Notify(125122)
		return
	}
	__antithesis_instrumentation__.Notify(125116)
	sp.detailsMu.storeDetails[storeID] = &detail
}

func (sp *StorePool) updateLocalStoresAfterLeaseTransfer(
	from roachpb.StoreID, to roachpb.StoreID, rangeQPS float64,
) {
	__antithesis_instrumentation__.Notify(125127)
	sp.detailsMu.Lock()
	defer sp.detailsMu.Unlock()

	fromDetail := *sp.getStoreDetailLocked(from)
	if fromDetail.desc != nil {
		__antithesis_instrumentation__.Notify(125129)
		fromDetail.desc.Capacity.LeaseCount--
		if fromDetail.desc.Capacity.QueriesPerSecond < rangeQPS {
			__antithesis_instrumentation__.Notify(125131)
			fromDetail.desc.Capacity.QueriesPerSecond = 0
		} else {
			__antithesis_instrumentation__.Notify(125132)
			fromDetail.desc.Capacity.QueriesPerSecond -= rangeQPS
		}
		__antithesis_instrumentation__.Notify(125130)
		sp.detailsMu.storeDetails[from] = &fromDetail
	} else {
		__antithesis_instrumentation__.Notify(125133)
	}
	__antithesis_instrumentation__.Notify(125128)

	toDetail := *sp.getStoreDetailLocked(to)
	if toDetail.desc != nil {
		__antithesis_instrumentation__.Notify(125134)
		toDetail.desc.Capacity.LeaseCount++
		toDetail.desc.Capacity.QueriesPerSecond += rangeQPS
		sp.detailsMu.storeDetails[to] = &toDetail
	} else {
		__antithesis_instrumentation__.Notify(125135)
	}
}

func newStoreDetail() *storeDetail {
	__antithesis_instrumentation__.Notify(125136)
	return &storeDetail{}
}

func (sp *StorePool) GetStores() map[roachpb.StoreID]roachpb.StoreDescriptor {
	__antithesis_instrumentation__.Notify(125137)
	sp.detailsMu.RLock()
	defer sp.detailsMu.RUnlock()
	stores := make(map[roachpb.StoreID]roachpb.StoreDescriptor, len(sp.detailsMu.storeDetails))
	for _, s := range sp.detailsMu.storeDetails {
		__antithesis_instrumentation__.Notify(125139)
		if s.desc != nil {
			__antithesis_instrumentation__.Notify(125140)
			stores[s.desc.StoreID] = *s.desc
		} else {
			__antithesis_instrumentation__.Notify(125141)
		}
	}
	__antithesis_instrumentation__.Notify(125138)
	return stores
}

func (sp *StorePool) getStoreDetailLocked(storeID roachpb.StoreID) *storeDetail {
	__antithesis_instrumentation__.Notify(125142)
	detail, ok := sp.detailsMu.storeDetails[storeID]
	if !ok {
		__antithesis_instrumentation__.Notify(125144)

		detail = newStoreDetail()
		detail.lastUpdatedTime = sp.startTime
		sp.detailsMu.storeDetails[storeID] = detail
	} else {
		__antithesis_instrumentation__.Notify(125145)
	}
	__antithesis_instrumentation__.Notify(125143)
	return detail
}

func (sp *StorePool) getStoreDescriptor(storeID roachpb.StoreID) (roachpb.StoreDescriptor, bool) {
	__antithesis_instrumentation__.Notify(125146)
	sp.detailsMu.RLock()
	defer sp.detailsMu.RUnlock()

	if detail, ok := sp.detailsMu.storeDetails[storeID]; ok && func() bool {
		__antithesis_instrumentation__.Notify(125148)
		return detail.desc != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(125149)
		return *detail.desc, true
	} else {
		__antithesis_instrumentation__.Notify(125150)
	}
	__antithesis_instrumentation__.Notify(125147)
	return roachpb.StoreDescriptor{}, false
}

func (sp *StorePool) decommissioningReplicas(
	repls []roachpb.ReplicaDescriptor,
) (decommissioningReplicas []roachpb.ReplicaDescriptor) {
	__antithesis_instrumentation__.Notify(125151)
	sp.detailsMu.Lock()
	defer sp.detailsMu.Unlock()

	now := sp.clock.Now().GoTime()
	timeUntilStoreDead := TimeUntilStoreDead.Get(&sp.st.SV)
	timeAfterStoreSuspect := TimeAfterStoreSuspect.Get(&sp.st.SV)

	for _, repl := range repls {
		__antithesis_instrumentation__.Notify(125153)
		detail := sp.getStoreDetailLocked(repl.StoreID)
		switch detail.status(now, timeUntilStoreDead, sp.nodeLivenessFn, timeAfterStoreSuspect) {
		case storeStatusDecommissioning:
			__antithesis_instrumentation__.Notify(125154)
			decommissioningReplicas = append(decommissioningReplicas, repl)
		default:
			__antithesis_instrumentation__.Notify(125155)
		}
	}
	__antithesis_instrumentation__.Notify(125152)
	return
}

func (sp *StorePool) ClusterNodeCount() int {
	__antithesis_instrumentation__.Notify(125156)
	return sp.nodeCountFn()
}

func (sp *StorePool) IsDead(storeID roachpb.StoreID) (bool, time.Duration, error) {
	__antithesis_instrumentation__.Notify(125157)
	sp.detailsMu.Lock()
	defer sp.detailsMu.Unlock()

	sd, ok := sp.detailsMu.storeDetails[storeID]
	if !ok {
		__antithesis_instrumentation__.Notify(125161)
		return false, 0, errors.Errorf("store %d was not found", storeID)
	} else {
		__antithesis_instrumentation__.Notify(125162)
	}
	__antithesis_instrumentation__.Notify(125158)

	now := sp.clock.Now().GoTime()
	timeUntilStoreDead := TimeUntilStoreDead.Get(&sp.st.SV)

	deadAsOf := sd.lastUpdatedTime.Add(timeUntilStoreDead)
	if now.After(deadAsOf) {
		__antithesis_instrumentation__.Notify(125163)
		return true, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(125164)
	}
	__antithesis_instrumentation__.Notify(125159)

	if sd.desc == nil {
		__antithesis_instrumentation__.Notify(125165)
		return false, 0, errors.Errorf("store %d status unknown, cant tell if it's dead or alive", storeID)
	} else {
		__antithesis_instrumentation__.Notify(125166)
	}
	__antithesis_instrumentation__.Notify(125160)
	return false, deadAsOf.Sub(now), nil
}

func (sp *StorePool) IsUnknown(storeID roachpb.StoreID) (bool, error) {
	__antithesis_instrumentation__.Notify(125167)
	status, err := sp.storeStatus(storeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(125169)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(125170)
	}
	__antithesis_instrumentation__.Notify(125168)
	return status == storeStatusUnknown, nil
}

func (sp *StorePool) IsLive(storeID roachpb.StoreID) (bool, error) {
	__antithesis_instrumentation__.Notify(125171)
	status, err := sp.storeStatus(storeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(125173)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(125174)
	}
	__antithesis_instrumentation__.Notify(125172)
	return status == storeStatusAvailable, nil
}

func (sp *StorePool) storeStatus(storeID roachpb.StoreID) (storeStatus, error) {
	__antithesis_instrumentation__.Notify(125175)
	sp.detailsMu.Lock()
	defer sp.detailsMu.Unlock()

	sd, ok := sp.detailsMu.storeDetails[storeID]
	if !ok {
		__antithesis_instrumentation__.Notify(125177)
		return storeStatusUnknown, errors.Errorf("store %d was not found", storeID)
	} else {
		__antithesis_instrumentation__.Notify(125178)
	}
	__antithesis_instrumentation__.Notify(125176)

	now := sp.clock.Now().GoTime()
	timeUntilStoreDead := TimeUntilStoreDead.Get(&sp.st.SV)
	timeAfterStoreSuspect := TimeAfterStoreSuspect.Get(&sp.st.SV)
	return sd.status(now, timeUntilStoreDead, sp.nodeLivenessFn, timeAfterStoreSuspect), nil
}

func (sp *StorePool) liveAndDeadReplicas(
	repls []roachpb.ReplicaDescriptor, includeSuspectAndDrainingStores bool,
) (liveReplicas, deadReplicas []roachpb.ReplicaDescriptor) {
	__antithesis_instrumentation__.Notify(125179)
	sp.detailsMu.Lock()
	defer sp.detailsMu.Unlock()

	now := sp.clock.Now().GoTime()
	timeUntilStoreDead := TimeUntilStoreDead.Get(&sp.st.SV)
	timeAfterStoreSuspect := TimeAfterStoreSuspect.Get(&sp.st.SV)

	for _, repl := range repls {
		__antithesis_instrumentation__.Notify(125181)
		detail := sp.getStoreDetailLocked(repl.StoreID)

		status := detail.status(now, timeUntilStoreDead, sp.nodeLivenessFn, timeAfterStoreSuspect)
		switch status {
		case storeStatusDead:
			__antithesis_instrumentation__.Notify(125182)
			deadReplicas = append(deadReplicas, repl)
		case storeStatusAvailable, storeStatusThrottled, storeStatusDecommissioning:
			__antithesis_instrumentation__.Notify(125183)

			liveReplicas = append(liveReplicas, repl)
		case storeStatusUnknown:
			__antithesis_instrumentation__.Notify(125184)

		case storeStatusSuspect, storeStatusDraining:
			__antithesis_instrumentation__.Notify(125185)
			if includeSuspectAndDrainingStores {
				__antithesis_instrumentation__.Notify(125187)
				liveReplicas = append(liveReplicas, repl)
			} else {
				__antithesis_instrumentation__.Notify(125188)
			}
		default:
			__antithesis_instrumentation__.Notify(125186)
			log.Fatalf(context.TODO(), "unknown store status %d", status)
		}
	}
	__antithesis_instrumentation__.Notify(125180)
	return
}

type stat struct {
	n, mean float64
}

func (s *stat) update(x float64) {
	__antithesis_instrumentation__.Notify(125189)
	s.n++
	s.mean += (x - s.mean) / s.n
}

type StoreList struct {
	stores []roachpb.StoreDescriptor

	candidateRanges stat

	candidateLeases stat

	candidateLogicalBytes stat

	candidateQueriesPerSecond stat

	candidateWritesPerSecond stat
}

func makeStoreList(descriptors []roachpb.StoreDescriptor) StoreList {
	__antithesis_instrumentation__.Notify(125190)
	sl := StoreList{stores: descriptors}
	for _, desc := range descriptors {
		__antithesis_instrumentation__.Notify(125192)
		if maxCapacityCheck(desc) {
			__antithesis_instrumentation__.Notify(125194)
			sl.candidateRanges.update(float64(desc.Capacity.RangeCount))
		} else {
			__antithesis_instrumentation__.Notify(125195)
		}
		__antithesis_instrumentation__.Notify(125193)
		sl.candidateLeases.update(float64(desc.Capacity.LeaseCount))
		sl.candidateLogicalBytes.update(float64(desc.Capacity.LogicalBytes))
		sl.candidateQueriesPerSecond.update(desc.Capacity.QueriesPerSecond)
		sl.candidateWritesPerSecond.update(desc.Capacity.WritesPerSecond)
	}
	__antithesis_instrumentation__.Notify(125191)
	return sl
}

func (sl StoreList) String() string {
	__antithesis_instrumentation__.Notify(125196)
	var buf bytes.Buffer
	fmt.Fprintf(&buf,
		"  candidate: avg-ranges=%v avg-leases=%v avg-disk-usage=%v avg-queries-per-second=%v",
		sl.candidateRanges.mean,
		sl.candidateLeases.mean,
		humanizeutil.IBytes(int64(sl.candidateLogicalBytes.mean)),
		sl.candidateQueriesPerSecond.mean,
	)
	if len(sl.stores) > 0 {
		__antithesis_instrumentation__.Notify(125199)
		fmt.Fprintf(&buf, "\n")
	} else {
		__antithesis_instrumentation__.Notify(125200)
		fmt.Fprintf(&buf, " <no candidates>")
	}
	__antithesis_instrumentation__.Notify(125197)
	for _, desc := range sl.stores {
		__antithesis_instrumentation__.Notify(125201)
		fmt.Fprintf(&buf, "  %d: ranges=%d leases=%d disk-usage=%s queries-per-second=%.2f l0-sublevels=%d\n",
			desc.StoreID, desc.Capacity.RangeCount,
			desc.Capacity.LeaseCount, humanizeutil.IBytes(desc.Capacity.LogicalBytes),
			desc.Capacity.QueriesPerSecond,
			desc.Capacity.L0Sublevels,
		)
	}
	__antithesis_instrumentation__.Notify(125198)
	return buf.String()
}

func (sl StoreList) excludeInvalid(constraints []roachpb.ConstraintsConjunction) StoreList {
	__antithesis_instrumentation__.Notify(125202)
	if len(constraints) == 0 {
		__antithesis_instrumentation__.Notify(125205)
		return sl
	} else {
		__antithesis_instrumentation__.Notify(125206)
	}
	__antithesis_instrumentation__.Notify(125203)
	var filteredDescs []roachpb.StoreDescriptor
	for _, store := range sl.stores {
		__antithesis_instrumentation__.Notify(125207)
		if ok := isStoreValid(store, constraints); ok {
			__antithesis_instrumentation__.Notify(125208)
			filteredDescs = append(filteredDescs, store)
		} else {
			__antithesis_instrumentation__.Notify(125209)
		}
	}
	__antithesis_instrumentation__.Notify(125204)
	return makeStoreList(filteredDescs)
}

type storeFilter int

const (
	_ storeFilter = iota

	storeFilterNone

	storeFilterThrottled

	storeFilterSuspect
)

type throttledStoreReasons []string

func (sp *StorePool) getStoreList(filter storeFilter) (StoreList, int, throttledStoreReasons) {
	__antithesis_instrumentation__.Notify(125210)
	sp.detailsMu.Lock()
	defer sp.detailsMu.Unlock()

	var storeIDs roachpb.StoreIDSlice
	for storeID := range sp.detailsMu.storeDetails {
		__antithesis_instrumentation__.Notify(125212)
		storeIDs = append(storeIDs, storeID)
	}
	__antithesis_instrumentation__.Notify(125211)
	return sp.getStoreListFromIDsLocked(storeIDs, filter)
}

func (sp *StorePool) getStoreListFromIDs(
	storeIDs roachpb.StoreIDSlice, filter storeFilter,
) (StoreList, int, throttledStoreReasons) {
	__antithesis_instrumentation__.Notify(125213)
	sp.detailsMu.Lock()
	defer sp.detailsMu.Unlock()
	return sp.getStoreListFromIDsLocked(storeIDs, filter)
}

func (sp *StorePool) getStoreListFromIDsLocked(
	storeIDs roachpb.StoreIDSlice, filter storeFilter,
) (StoreList, int, throttledStoreReasons) {
	__antithesis_instrumentation__.Notify(125214)
	if sp.deterministic {
		__antithesis_instrumentation__.Notify(125217)
		sort.Sort(storeIDs)
	} else {
		__antithesis_instrumentation__.Notify(125218)
		shuffle.Shuffle(storeIDs)
	}
	__antithesis_instrumentation__.Notify(125215)

	var aliveStoreCount int
	var throttled throttledStoreReasons
	var storeDescriptors []roachpb.StoreDescriptor

	now := sp.clock.Now().GoTime()
	timeUntilStoreDead := TimeUntilStoreDead.Get(&sp.st.SV)
	timeAfterStoreSuspect := TimeAfterStoreSuspect.Get(&sp.st.SV)

	for _, storeID := range storeIDs {
		__antithesis_instrumentation__.Notify(125219)
		detail, ok := sp.detailsMu.storeDetails[storeID]
		if !ok {
			__antithesis_instrumentation__.Notify(125221)

			continue
		} else {
			__antithesis_instrumentation__.Notify(125222)
		}
		__antithesis_instrumentation__.Notify(125220)
		switch s := detail.status(now, timeUntilStoreDead, sp.nodeLivenessFn, timeAfterStoreSuspect); s {
		case storeStatusThrottled:
			__antithesis_instrumentation__.Notify(125223)
			aliveStoreCount++
			throttled = append(throttled, detail.throttledBecause)
			if filter != storeFilterThrottled {
				__antithesis_instrumentation__.Notify(125229)
				storeDescriptors = append(storeDescriptors, *detail.desc)
			} else {
				__antithesis_instrumentation__.Notify(125230)
			}
		case storeStatusAvailable:
			__antithesis_instrumentation__.Notify(125224)
			aliveStoreCount++
			storeDescriptors = append(storeDescriptors, *detail.desc)
		case storeStatusDraining:
			__antithesis_instrumentation__.Notify(125225)
			throttled = append(throttled, fmt.Sprintf("s%d: draining", storeID))
		case storeStatusSuspect:
			__antithesis_instrumentation__.Notify(125226)
			aliveStoreCount++
			throttled = append(throttled, fmt.Sprintf("s%d: suspect", storeID))
			if filter != storeFilterThrottled && func() bool {
				__antithesis_instrumentation__.Notify(125231)
				return filter != storeFilterSuspect == true
			}() == true {
				__antithesis_instrumentation__.Notify(125232)
				storeDescriptors = append(storeDescriptors, *detail.desc)
			} else {
				__antithesis_instrumentation__.Notify(125233)
			}
		case storeStatusDead, storeStatusUnknown, storeStatusDecommissioning:
			__antithesis_instrumentation__.Notify(125227)

		default:
			__antithesis_instrumentation__.Notify(125228)
			panic(fmt.Sprintf("unknown store status: %d", s))
		}
	}
	__antithesis_instrumentation__.Notify(125216)
	return makeStoreList(storeDescriptors), aliveStoreCount, throttled
}

type throttleReason int

const (
	_ throttleReason = iota
	throttleFailed
)

func (sp *StorePool) throttle(reason throttleReason, why string, storeID roachpb.StoreID) {
	__antithesis_instrumentation__.Notify(125234)
	sp.detailsMu.Lock()
	defer sp.detailsMu.Unlock()
	detail := sp.getStoreDetailLocked(storeID)
	detail.throttledBecause = why

	switch reason {
	case throttleFailed:
		__antithesis_instrumentation__.Notify(125235)
		timeout := FailedReservationsTimeout.Get(&sp.st.SV)
		detail.throttledUntil = sp.clock.PhysicalTime().Add(timeout)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(125237)
			ctx := sp.AnnotateCtx(context.TODO())
			log.Infof(ctx, "snapshot failed (%s), s%d will be throttled for %s until %s",
				why, storeID, timeout, detail.throttledUntil)
		} else {
			__antithesis_instrumentation__.Notify(125238)
		}
	default:
		__antithesis_instrumentation__.Notify(125236)
		log.Warningf(sp.AnnotateCtx(context.TODO()), "unknown throttle reason %v", reason)
	}
}

func (sp *StorePool) getLocalitiesByStore(
	replicas []roachpb.ReplicaDescriptor,
) map[roachpb.StoreID]roachpb.Locality {
	__antithesis_instrumentation__.Notify(125239)
	sp.localitiesMu.RLock()
	defer sp.localitiesMu.RUnlock()
	localities := make(map[roachpb.StoreID]roachpb.Locality)
	for _, replica := range replicas {
		__antithesis_instrumentation__.Notify(125241)
		nodeTier := roachpb.Tier{Key: "node", Value: replica.NodeID.String()}
		if locality, ok := sp.localitiesMu.nodeLocalities[replica.NodeID]; ok {
			__antithesis_instrumentation__.Notify(125242)
			localities[replica.StoreID] = locality.locality.AddTier(nodeTier)
		} else {
			__antithesis_instrumentation__.Notify(125243)
			localities[replica.StoreID] = roachpb.Locality{
				Tiers: []roachpb.Tier{nodeTier},
			}
		}
	}
	__antithesis_instrumentation__.Notify(125240)
	return localities
}

func (sp *StorePool) getLocalitiesByNode(
	replicas []roachpb.ReplicaDescriptor,
) map[roachpb.NodeID]roachpb.Locality {
	__antithesis_instrumentation__.Notify(125244)
	sp.localitiesMu.RLock()
	defer sp.localitiesMu.RUnlock()
	localities := make(map[roachpb.NodeID]roachpb.Locality)
	for _, replica := range replicas {
		__antithesis_instrumentation__.Notify(125246)
		if locality, ok := sp.localitiesMu.nodeLocalities[replica.NodeID]; ok {
			__antithesis_instrumentation__.Notify(125247)
			localities[replica.NodeID] = locality.locality
		} else {
			__antithesis_instrumentation__.Notify(125248)
			localities[replica.NodeID] = roachpb.Locality{}
		}
	}
	__antithesis_instrumentation__.Notify(125245)
	return localities
}

func (sp *StorePool) getNodeLocalityString(nodeID roachpb.NodeID) string {
	__antithesis_instrumentation__.Notify(125249)
	sp.localitiesMu.RLock()
	defer sp.localitiesMu.RUnlock()
	locality, ok := sp.localitiesMu.nodeLocalities[nodeID]
	if !ok {
		__antithesis_instrumentation__.Notify(125251)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(125252)
	}
	__antithesis_instrumentation__.Notify(125250)
	return locality.str
}

func (sp *StorePool) isStoreReadyForRoutineReplicaTransferInternal(
	ctx context.Context, targetStoreID roachpb.StoreID,
) bool {
	__antithesis_instrumentation__.Notify(125253)
	status, err := sp.storeStatus(targetStoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(125255)
		return false
	} else {
		__antithesis_instrumentation__.Notify(125256)
	}
	__antithesis_instrumentation__.Notify(125254)
	switch status {
	case storeStatusThrottled, storeStatusAvailable:
		__antithesis_instrumentation__.Notify(125257)
		log.VEventf(ctx, 3,
			"s%d is a live target, candidate for rebalancing", targetStoreID)
		return true
	case storeStatusDead, storeStatusUnknown, storeStatusDecommissioning, storeStatusSuspect, storeStatusDraining:
		__antithesis_instrumentation__.Notify(125258)
		log.VEventf(ctx, 3,
			"not considering non-live store s%d (%v)", targetStoreID, status)
		return false
	default:
		__antithesis_instrumentation__.Notify(125259)
		panic(fmt.Sprintf("unknown store status: %d", status))
	}
}
