package loqrecovery

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/loqrecovery/loqrecoverypb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const nextReplicaIDIncrement = 10

type updatedLocationsMap map[roachpb.NodeID]storeIDSet

func (m updatedLocationsMap) add(node roachpb.NodeID, store roachpb.StoreID) {
	__antithesis_instrumentation__.Notify(109902)
	var set storeIDSet
	var ok bool
	if set, ok = m[node]; !ok {
		__antithesis_instrumentation__.Notify(109904)
		set = make(storeIDSet)
		m[node] = set
	} else {
		__antithesis_instrumentation__.Notify(109905)
	}
	__antithesis_instrumentation__.Notify(109903)
	set[store] = struct{}{}
}

func (m updatedLocationsMap) asMapOfSlices() map[roachpb.NodeID][]roachpb.StoreID {
	__antithesis_instrumentation__.Notify(109906)
	newMap := make(map[roachpb.NodeID][]roachpb.StoreID)
	for k, v := range m {
		__antithesis_instrumentation__.Notify(109908)
		newMap[k] = storeSliceFromSet(v)
	}
	__antithesis_instrumentation__.Notify(109907)
	return newMap
}

type PlanningReport struct {
	TotalReplicas int

	DiscardedNonSurvivors int

	PresentStores []roachpb.StoreID

	MissingStores []roachpb.StoreID

	PlannedUpdates []ReplicaUpdateReport

	UpdatedNodes map[roachpb.NodeID][]roachpb.StoreID

	Problems []Problem
}

func (p PlanningReport) Error() error {
	__antithesis_instrumentation__.Notify(109909)
	if len(p.Problems) > 0 {
		__antithesis_instrumentation__.Notify(109911)
		return &RecoveryError{p.Problems}
	} else {
		__antithesis_instrumentation__.Notify(109912)
	}
	__antithesis_instrumentation__.Notify(109910)
	return nil
}

type ReplicaUpdateReport struct {
	RangeID                    roachpb.RangeID
	StartKey                   roachpb.RKey
	Replica                    roachpb.ReplicaDescriptor
	OldReplica                 roachpb.ReplicaDescriptor
	StoreID                    roachpb.StoreID
	DiscardedAvailableReplicas roachpb.ReplicaSet
	DiscardedDeadReplicas      roachpb.ReplicaSet
}

func PlanReplicas(
	ctx context.Context, nodes []loqrecoverypb.NodeReplicaInfo, deadStores []roachpb.StoreID,
) (loqrecoverypb.ReplicaUpdatePlan, PlanningReport, error) {
	__antithesis_instrumentation__.Notify(109913)
	var report PlanningReport
	updatedLocations := make(updatedLocationsMap)
	var replicas []loqrecoverypb.ReplicaInfo
	for _, node := range nodes {
		__antithesis_instrumentation__.Notify(109920)
		replicas = append(replicas, node.Replicas...)
	}
	__antithesis_instrumentation__.Notify(109914)
	availableStoreIDs, missingStores, err := validateReplicaSets(replicas, deadStores)
	if err != nil {
		__antithesis_instrumentation__.Notify(109921)
		return loqrecoverypb.ReplicaUpdatePlan{}, PlanningReport{}, err
	} else {
		__antithesis_instrumentation__.Notify(109922)
	}
	__antithesis_instrumentation__.Notify(109915)
	report.PresentStores = storeSliceFromSet(availableStoreIDs)
	report.MissingStores = storeSliceFromSet(missingStores)

	replicasByRangeID := groupReplicasByRangeID(replicas)

	var proposedSurvivors []rankedReplicas
	for _, rangeReplicas := range replicasByRangeID {
		__antithesis_instrumentation__.Notify(109923)
		proposedSurvivors = append(proposedSurvivors, rankReplicasBySurvivability(rangeReplicas))
	}
	__antithesis_instrumentation__.Notify(109916)
	problems, err := checkKeyspaceCovering(proposedSurvivors)
	if err != nil {
		__antithesis_instrumentation__.Notify(109924)
		return loqrecoverypb.ReplicaUpdatePlan{}, PlanningReport{}, err
	} else {
		__antithesis_instrumentation__.Notify(109925)
	}
	__antithesis_instrumentation__.Notify(109917)

	var plan []loqrecoverypb.ReplicaUpdate
	for _, p := range proposedSurvivors {
		__antithesis_instrumentation__.Notify(109926)
		report.TotalReplicas += len(p)
		u, ok := makeReplicaUpdateIfNeeded(ctx, p, availableStoreIDs)
		if ok {
			__antithesis_instrumentation__.Notify(109927)
			problems = append(problems, checkDescriptor(p)...)
			plan = append(plan, u)
			report.DiscardedNonSurvivors += len(p) - 1
			report.PlannedUpdates = append(report.PlannedUpdates, makeReplicaUpdateReport(ctx, p, u))
			updatedLocations.add(u.NodeID(), u.StoreID())
			log.Infof(ctx, "replica has lost quorum, recovering: %s -> %s", p.survivor().Desc, u)
		} else {
			__antithesis_instrumentation__.Notify(109928)
			log.Infof(ctx, "range r%d didn't lose quorum", p.rangeID())
		}
	}
	__antithesis_instrumentation__.Notify(109918)

	sort.Slice(problems, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(109929)
		return problems[i].Span().Key.Compare(problems[j].Span().Key) < 0
	})
	__antithesis_instrumentation__.Notify(109919)
	report.Problems = problems
	report.UpdatedNodes = updatedLocations.asMapOfSlices()
	return loqrecoverypb.ReplicaUpdatePlan{Updates: plan}, report, nil
}

func validateReplicaSets(
	replicas []loqrecoverypb.ReplicaInfo, deadStores []roachpb.StoreID,
) (availableStoreIDs, missingStoreIDs storeIDSet, _ error) {
	__antithesis_instrumentation__.Notify(109930)

	availableStoreIDs = make(storeIDSet)
	missingStoreIDs = make(storeIDSet)
	for _, replicaDescriptor := range replicas {
		__antithesis_instrumentation__.Notify(109937)
		availableStoreIDs[replicaDescriptor.StoreID] = struct{}{}
		for _, replicaDesc := range replicaDescriptor.Desc.InternalReplicas {
			__antithesis_instrumentation__.Notify(109938)
			missingStoreIDs[replicaDesc.StoreID] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(109931)

	for id := range availableStoreIDs {
		__antithesis_instrumentation__.Notify(109939)
		delete(missingStoreIDs, id)
	}
	__antithesis_instrumentation__.Notify(109932)

	missingButNotDeadStoreIDs := make(storeIDSet)
	for id := range missingStoreIDs {
		__antithesis_instrumentation__.Notify(109940)
		missingButNotDeadStoreIDs[id] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(109933)

	suspiciousStoreIDs := make(storeIDSet)
	for _, id := range deadStores {
		__antithesis_instrumentation__.Notify(109941)
		delete(missingButNotDeadStoreIDs, id)
		if _, ok := availableStoreIDs[id]; ok {
			__antithesis_instrumentation__.Notify(109942)
			suspiciousStoreIDs[id] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(109943)
		}
	}
	__antithesis_instrumentation__.Notify(109934)
	if len(suspiciousStoreIDs) > 0 {
		__antithesis_instrumentation__.Notify(109944)
		return nil, nil, errors.Errorf(
			"stores %s are listed as dead, but replica info is provided for them",
			joinStoreIDs(suspiciousStoreIDs))
	} else {
		__antithesis_instrumentation__.Notify(109945)
	}
	__antithesis_instrumentation__.Notify(109935)
	if len(deadStores) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(109946)
		return len(missingButNotDeadStoreIDs) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(109947)

		return nil, nil, errors.Errorf(
			"information about stores %s were not provided, nor they are listed as dead",
			joinStoreIDs(missingButNotDeadStoreIDs))
	} else {
		__antithesis_instrumentation__.Notify(109948)
	}
	__antithesis_instrumentation__.Notify(109936)
	return availableStoreIDs, missingStoreIDs, nil
}

func groupReplicasByRangeID(
	descriptors []loqrecoverypb.ReplicaInfo,
) map[roachpb.RangeID][]loqrecoverypb.ReplicaInfo {
	__antithesis_instrumentation__.Notify(109949)
	groupedRanges := make(map[roachpb.RangeID][]loqrecoverypb.ReplicaInfo)
	for _, descriptor := range descriptors {
		__antithesis_instrumentation__.Notify(109951)
		groupedRanges[descriptor.Desc.RangeID] = append(
			groupedRanges[descriptor.Desc.RangeID], descriptor)
	}
	__antithesis_instrumentation__.Notify(109950)
	return groupedRanges
}

type rankedReplicas []loqrecoverypb.ReplicaInfo

func (p rankedReplicas) startKey() roachpb.RKey {
	__antithesis_instrumentation__.Notify(109952)
	return p[0].Desc.StartKey
}

func (p rankedReplicas) endKey() roachpb.RKey {
	__antithesis_instrumentation__.Notify(109953)
	return p[0].Desc.EndKey
}

func (p rankedReplicas) span() roachpb.Span {
	__antithesis_instrumentation__.Notify(109954)
	return roachpb.Span{Key: roachpb.Key(p[0].Desc.StartKey), EndKey: roachpb.Key(p[0].Desc.EndKey)}
}

func (p rankedReplicas) rangeID() roachpb.RangeID {
	__antithesis_instrumentation__.Notify(109955)
	return p[0].Desc.RangeID
}

func (p rankedReplicas) nodeID() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(109956)
	return p[0].NodeID
}

func (p rankedReplicas) storeID() roachpb.StoreID {
	__antithesis_instrumentation__.Notify(109957)
	return p[0].StoreID
}

func (p rankedReplicas) survivor() *loqrecoverypb.ReplicaInfo {
	__antithesis_instrumentation__.Notify(109958)
	return &p[0]
}

func rankReplicasBySurvivability(replicas []loqrecoverypb.ReplicaInfo) rankedReplicas {
	__antithesis_instrumentation__.Notify(109959)
	isVoter := func(desc loqrecoverypb.ReplicaInfo) int {
		__antithesis_instrumentation__.Notify(109962)
		for _, replica := range desc.Desc.InternalReplicas {
			__antithesis_instrumentation__.Notify(109964)
			if replica.StoreID == desc.StoreID {
				__antithesis_instrumentation__.Notify(109965)
				if replica.IsVoterNewConfig() {
					__antithesis_instrumentation__.Notify(109967)
					return 1
				} else {
					__antithesis_instrumentation__.Notify(109968)
				}
				__antithesis_instrumentation__.Notify(109966)
				return 0
			} else {
				__antithesis_instrumentation__.Notify(109969)
			}
		}
		__antithesis_instrumentation__.Notify(109963)

		return 0
	}
	__antithesis_instrumentation__.Notify(109960)
	sort.Slice(replicas, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(109970)

		voterI := isVoter(replicas[i])
		voterJ := isVoter(replicas[j])
		if voterI > voterJ {
			__antithesis_instrumentation__.Notify(109975)
			return true
		} else {
			__antithesis_instrumentation__.Notify(109976)
		}
		__antithesis_instrumentation__.Notify(109971)
		if voterI < voterJ {
			__antithesis_instrumentation__.Notify(109977)
			return false
		} else {
			__antithesis_instrumentation__.Notify(109978)
		}
		__antithesis_instrumentation__.Notify(109972)
		if replicas[i].RaftAppliedIndex > replicas[j].RaftAppliedIndex {
			__antithesis_instrumentation__.Notify(109979)
			return true
		} else {
			__antithesis_instrumentation__.Notify(109980)
		}
		__antithesis_instrumentation__.Notify(109973)
		if replicas[i].RaftAppliedIndex < replicas[j].RaftAppliedIndex {
			__antithesis_instrumentation__.Notify(109981)
			return false
		} else {
			__antithesis_instrumentation__.Notify(109982)
		}
		__antithesis_instrumentation__.Notify(109974)
		return replicas[i].StoreID > replicas[j].StoreID
	})
	__antithesis_instrumentation__.Notify(109961)
	return replicas
}

func checkKeyspaceCovering(replicas []rankedReplicas) ([]Problem, error) {
	__antithesis_instrumentation__.Notify(109983)
	sort.Slice(replicas, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(109987)

		if replicas[i].startKey().Less(replicas[j].startKey()) {
			__antithesis_instrumentation__.Notify(109990)
			return true
		} else {
			__antithesis_instrumentation__.Notify(109991)
		}
		__antithesis_instrumentation__.Notify(109988)
		if replicas[i].startKey().Equal(replicas[j].startKey()) {
			__antithesis_instrumentation__.Notify(109992)
			return replicas[i].rangeID() < replicas[j].rangeID()
		} else {
			__antithesis_instrumentation__.Notify(109993)
		}
		__antithesis_instrumentation__.Notify(109989)
		return false
	})
	__antithesis_instrumentation__.Notify(109984)
	var problems []Problem
	prevDesc := rankedReplicas{{Desc: roachpb.RangeDescriptor{}}}

	for _, rankedDescriptors := range replicas {
		__antithesis_instrumentation__.Notify(109994)

		r, err := rankedDescriptors.survivor().Replica()
		if err != nil {
			__antithesis_instrumentation__.Notify(109998)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(109999)
		}
		__antithesis_instrumentation__.Notify(109995)
		if !r.IsVoterNewConfig() {
			__antithesis_instrumentation__.Notify(110000)
			continue
		} else {
			__antithesis_instrumentation__.Notify(110001)
		}
		__antithesis_instrumentation__.Notify(109996)
		switch {
		case rankedDescriptors.startKey().Less(prevDesc.endKey()):
			__antithesis_instrumentation__.Notify(110002)
			start := keyMax(rankedDescriptors.startKey(), prevDesc.startKey())
			end := keyMin(rankedDescriptors.endKey(), prevDesc.endKey())
			problems = append(problems, keyspaceOverlap{
				span:       roachpb.Span{Key: roachpb.Key(start), EndKey: roachpb.Key(end)},
				range1:     prevDesc.rangeID(),
				range1Span: prevDesc.span(),
				range2:     rankedDescriptors.rangeID(),
				range2Span: rankedDescriptors.span(),
			})
		case prevDesc.endKey().Less(rankedDescriptors.startKey()):
			__antithesis_instrumentation__.Notify(110003)
			problems = append(problems, keyspaceGap{
				span: roachpb.Span{
					Key:    roachpb.Key(prevDesc.endKey()),
					EndKey: roachpb.Key(rankedDescriptors.startKey()),
				},
				range1:     prevDesc.rangeID(),
				range1Span: prevDesc.span(),
				range2:     rankedDescriptors.rangeID(),
				range2Span: rankedDescriptors.span(),
			})
		default:
			__antithesis_instrumentation__.Notify(110004)
		}
		__antithesis_instrumentation__.Notify(109997)

		if prevDesc.endKey().Less(rankedDescriptors.endKey()) {
			__antithesis_instrumentation__.Notify(110005)
			prevDesc = rankedDescriptors
		} else {
			__antithesis_instrumentation__.Notify(110006)
		}
	}
	__antithesis_instrumentation__.Notify(109985)
	if !prevDesc.endKey().Equal(roachpb.RKeyMax) {
		__antithesis_instrumentation__.Notify(110007)
		problems = append(problems, keyspaceGap{
			span:       roachpb.Span{Key: roachpb.Key(prevDesc.endKey()), EndKey: roachpb.KeyMax},
			range1:     prevDesc.rangeID(),
			range1Span: prevDesc.span(),
			range2:     roachpb.RangeID(0),
			range2Span: roachpb.Span{Key: roachpb.KeyMax, EndKey: roachpb.KeyMax},
		})
	} else {
		__antithesis_instrumentation__.Notify(110008)
	}
	__antithesis_instrumentation__.Notify(109986)

	return problems, nil
}

func makeReplicaUpdateIfNeeded(
	ctx context.Context, p rankedReplicas, liveStoreIDs storeIDSet,
) (loqrecoverypb.ReplicaUpdate, bool) {
	__antithesis_instrumentation__.Notify(110009)
	if p.survivor().Desc.Replicas().CanMakeProgress(func(rep roachpb.ReplicaDescriptor) bool {
		__antithesis_instrumentation__.Notify(110013)
		_, ok := liveStoreIDs[rep.StoreID]
		return ok
	}) {
		__antithesis_instrumentation__.Notify(110014)
		return loqrecoverypb.ReplicaUpdate{}, false
	} else {
		__antithesis_instrumentation__.Notify(110015)
	}
	__antithesis_instrumentation__.Notify(110010)

	nextReplicaID := p.survivor().Desc.NextReplicaID
	for _, r := range p[1:] {
		__antithesis_instrumentation__.Notify(110016)
		if r.Desc.NextReplicaID > nextReplicaID {
			__antithesis_instrumentation__.Notify(110017)
			nextReplicaID = r.Desc.NextReplicaID
		} else {
			__antithesis_instrumentation__.Notify(110018)
		}
	}
	__antithesis_instrumentation__.Notify(110011)

	replica, err := p.survivor().Replica()
	if err != nil {
		__antithesis_instrumentation__.Notify(110019)

		log.Fatalf(ctx, "unexpected invalid replica info while making recovery plan, "+
			"we should never have unvalidated descriptors at planning stage, they must be detected "+
			"while performing keyspace coverage check: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(110020)
	}
	__antithesis_instrumentation__.Notify(110012)

	return loqrecoverypb.ReplicaUpdate{
		RangeID:      p.rangeID(),
		StartKey:     loqrecoverypb.RecoveryKey(p.startKey()),
		OldReplicaID: replica.ReplicaID,
		NewReplica: roachpb.ReplicaDescriptor{
			NodeID:    p.nodeID(),
			StoreID:   p.storeID(),
			ReplicaID: nextReplicaID + nextReplicaIDIncrement,
		},
		NextReplicaID: nextReplicaID + nextReplicaIDIncrement + 1,
	}, true
}

func checkDescriptor(rankedDescriptors rankedReplicas) (problems []Problem) {
	__antithesis_instrumentation__.Notify(110021)

	for _, change := range rankedDescriptors.survivor().RaftLogDescriptorChanges {
		__antithesis_instrumentation__.Notify(110023)
		switch change.ChangeType {
		case loqrecoverypb.DescriptorChangeType_Split:
			__antithesis_instrumentation__.Notify(110024)
			problems = append(problems, rangeSplit{
				rangeID:    rankedDescriptors.rangeID(),
				span:       rankedDescriptors.span(),
				rHSRangeID: change.OtherDesc.RangeID,
				rHSRangeSpan: roachpb.Span{
					Key:    roachpb.Key(change.OtherDesc.StartKey),
					EndKey: roachpb.Key(change.OtherDesc.EndKey),
				},
			})
		case loqrecoverypb.DescriptorChangeType_Merge:
			__antithesis_instrumentation__.Notify(110025)
			problems = append(problems, rangeMerge{
				rangeID:    rankedDescriptors.rangeID(),
				span:       rankedDescriptors.span(),
				rHSRangeID: change.OtherDesc.RangeID,
				rHSRangeSpan: roachpb.Span{
					Key:    roachpb.Key(change.OtherDesc.StartKey),
					EndKey: roachpb.Key(change.OtherDesc.EndKey),
				},
			})
		case loqrecoverypb.DescriptorChangeType_ReplicaChange:
			__antithesis_instrumentation__.Notify(110026)

			_, ok := change.Desc.GetReplicaDescriptor(rankedDescriptors.storeID())
			if !ok {
				__antithesis_instrumentation__.Notify(110028)
				problems = append(problems, rangeReplicaRemoval{
					rangeID: rankedDescriptors.rangeID(),
					span:    rankedDescriptors.span(),
				})
			} else {
				__antithesis_instrumentation__.Notify(110029)
			}
		default:
			__antithesis_instrumentation__.Notify(110027)
		}
	}
	__antithesis_instrumentation__.Notify(110022)
	return
}

func makeReplicaUpdateReport(
	ctx context.Context, p rankedReplicas, update loqrecoverypb.ReplicaUpdate,
) ReplicaUpdateReport {
	__antithesis_instrumentation__.Notify(110030)
	oldReplica, err := p.survivor().Replica()
	if err != nil {
		__antithesis_instrumentation__.Notify(110033)

		log.Fatalf(ctx, "unexpected invalid replica info while making recovery plan: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(110034)
	}
	__antithesis_instrumentation__.Notify(110031)

	discardedDead := p.survivor().Desc.Replicas()
	discardedDead.RemoveReplica(update.NodeID(), update.StoreID())

	discardedAvailable := roachpb.ReplicaSet{}
	for _, replica := range p[1:] {
		__antithesis_instrumentation__.Notify(110035)
		discardedDead.RemoveReplica(replica.NodeID, replica.StoreID)
		r, _ := replica.Desc.GetReplicaDescriptor(replica.StoreID)
		discardedAvailable.AddReplica(r)
	}
	__antithesis_instrumentation__.Notify(110032)

	return ReplicaUpdateReport{
		RangeID:                    p.rangeID(),
		StartKey:                   p.startKey(),
		OldReplica:                 oldReplica,
		Replica:                    update.NewReplica,
		StoreID:                    p.storeID(),
		DiscardedDeadReplicas:      discardedDead,
		DiscardedAvailableReplicas: discardedAvailable,
	}
}
