package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func ReplicaTypeVoterFull() *ReplicaType {
	__antithesis_instrumentation__.Notify(176767)
	t := VOTER_FULL
	return &t
}

func ReplicaTypeVoterIncoming() *ReplicaType {
	__antithesis_instrumentation__.Notify(176768)
	t := VOTER_INCOMING
	return &t
}

func ReplicaTypeVoterOutgoing() *ReplicaType {
	__antithesis_instrumentation__.Notify(176769)
	t := VOTER_OUTGOING
	return &t
}

func ReplicaTypeVoterDemotingLearner() *ReplicaType {
	__antithesis_instrumentation__.Notify(176770)
	t := VOTER_DEMOTING_LEARNER
	return &t
}

func ReplicaTypeLearner() *ReplicaType {
	__antithesis_instrumentation__.Notify(176771)
	t := LEARNER
	return &t
}

func ReplicaTypeNonVoter() *ReplicaType {
	__antithesis_instrumentation__.Notify(176772)
	t := NON_VOTER
	return &t
}

type ReplicaSet struct {
	wrapped []ReplicaDescriptor
}

func MakeReplicaSet(replicas []ReplicaDescriptor) ReplicaSet {
	__antithesis_instrumentation__.Notify(176773)
	return ReplicaSet{wrapped: replicas}
}

func (d ReplicaSet) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(176774)
	for i, desc := range d.wrapped {
		__antithesis_instrumentation__.Notify(176775)
		if i > 0 {
			__antithesis_instrumentation__.Notify(176777)
			w.SafeRune(',')
		} else {
			__antithesis_instrumentation__.Notify(176778)
		}
		__antithesis_instrumentation__.Notify(176776)
		w.Print(desc)
	}
}

func (d ReplicaSet) String() string {
	__antithesis_instrumentation__.Notify(176779)
	return redact.StringWithoutMarkers(d)
}

func (d ReplicaSet) Descriptors() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(176780)
	return d.wrapped
}

func predVoterFull(rDesc ReplicaDescriptor) bool {
	__antithesis_instrumentation__.Notify(176781)
	switch rDesc.GetType() {
	case VOTER_FULL:
		__antithesis_instrumentation__.Notify(176783)
		return true
	default:
		__antithesis_instrumentation__.Notify(176784)
	}
	__antithesis_instrumentation__.Notify(176782)
	return false
}

func predVoterFullOrIncoming(rDesc ReplicaDescriptor) bool {
	__antithesis_instrumentation__.Notify(176785)
	switch rDesc.GetType() {
	case VOTER_FULL, VOTER_INCOMING:
		__antithesis_instrumentation__.Notify(176787)
		return true
	default:
		__antithesis_instrumentation__.Notify(176788)
	}
	__antithesis_instrumentation__.Notify(176786)
	return false
}

func predLearner(rDesc ReplicaDescriptor) bool {
	__antithesis_instrumentation__.Notify(176789)
	return rDesc.GetType() == LEARNER
}

func predNonVoter(rDesc ReplicaDescriptor) bool {
	__antithesis_instrumentation__.Notify(176790)
	return rDesc.GetType() == NON_VOTER
}

func predVoterOrNonVoter(rDesc ReplicaDescriptor) bool {
	__antithesis_instrumentation__.Notify(176791)
	return predVoterFullOrIncoming(rDesc) || func() bool {
		__antithesis_instrumentation__.Notify(176792)
		return predNonVoter(rDesc) == true
	}() == true
}

func predVoterFullOrNonVoter(rDesc ReplicaDescriptor) bool {
	__antithesis_instrumentation__.Notify(176793)
	return predVoterFull(rDesc) || func() bool {
		__antithesis_instrumentation__.Notify(176794)
		return predNonVoter(rDesc) == true
	}() == true
}

func (d ReplicaSet) Voters() ReplicaSet {
	__antithesis_instrumentation__.Notify(176795)
	return d.Filter(predVoterFullOrIncoming)
}

func (d ReplicaSet) VoterDescriptors() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(176796)
	return d.FilterToDescriptors(predVoterFullOrIncoming)
}

func (d ReplicaSet) LearnerDescriptors() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(176797)
	return d.FilterToDescriptors(predLearner)
}

func (d ReplicaSet) NonVoters() ReplicaSet {
	__antithesis_instrumentation__.Notify(176798)
	return d.Filter(predNonVoter)
}

func (d ReplicaSet) NonVoterDescriptors() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(176799)
	return d.FilterToDescriptors(predNonVoter)
}

func (d ReplicaSet) VoterFullAndNonVoterDescriptors() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(176800)
	return d.FilterToDescriptors(predVoterFullOrNonVoter)
}

func (d ReplicaSet) VoterAndNonVoterDescriptors() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(176801)
	return d.FilterToDescriptors(predVoterOrNonVoter)
}

func (d ReplicaSet) Filter(pred func(rDesc ReplicaDescriptor) bool) ReplicaSet {
	__antithesis_instrumentation__.Notify(176802)
	return MakeReplicaSet(d.FilterToDescriptors(pred))
}

func (d ReplicaSet) FilterToDescriptors(
	pred func(rDesc ReplicaDescriptor) bool,
) []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(176803)

	fastpath := true
	out := d.wrapped
	for i := range d.wrapped {
		__antithesis_instrumentation__.Notify(176805)
		if pred(d.wrapped[i]) {
			__antithesis_instrumentation__.Notify(176806)
			if !fastpath {
				__antithesis_instrumentation__.Notify(176807)
				out = append(out, d.wrapped[i])
			} else {
				__antithesis_instrumentation__.Notify(176808)
			}
		} else {
			__antithesis_instrumentation__.Notify(176809)
			if fastpath {
				__antithesis_instrumentation__.Notify(176810)
				out = nil
				out = append(out, d.wrapped[:i]...)
				fastpath = false
			} else {
				__antithesis_instrumentation__.Notify(176811)
			}
		}
	}
	__antithesis_instrumentation__.Notify(176804)
	return out
}

func (d ReplicaSet) AsProto() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(176812)
	return d.wrapped
}

func (d ReplicaSet) DeepCopy() ReplicaSet {
	__antithesis_instrumentation__.Notify(176813)
	return ReplicaSet{
		wrapped: append([]ReplicaDescriptor(nil), d.wrapped...),
	}
}

func (d *ReplicaSet) AddReplica(r ReplicaDescriptor) {
	__antithesis_instrumentation__.Notify(176814)
	d.wrapped = append(d.wrapped, r)
}

func (d *ReplicaSet) RemoveReplica(nodeID NodeID, storeID StoreID) (ReplicaDescriptor, bool) {
	__antithesis_instrumentation__.Notify(176815)
	idx := -1
	for i := range d.wrapped {
		__antithesis_instrumentation__.Notify(176818)
		if d.wrapped[i].NodeID == nodeID && func() bool {
			__antithesis_instrumentation__.Notify(176819)
			return d.wrapped[i].StoreID == storeID == true
		}() == true {
			__antithesis_instrumentation__.Notify(176820)
			idx = i
			break
		} else {
			__antithesis_instrumentation__.Notify(176821)
		}
	}
	__antithesis_instrumentation__.Notify(176816)
	if idx == -1 {
		__antithesis_instrumentation__.Notify(176822)
		return ReplicaDescriptor{}, false
	} else {
		__antithesis_instrumentation__.Notify(176823)
	}
	__antithesis_instrumentation__.Notify(176817)

	d.wrapped[idx], d.wrapped[len(d.wrapped)-1] = d.wrapped[len(d.wrapped)-1], d.wrapped[idx]
	removed := d.wrapped[len(d.wrapped)-1]
	d.wrapped = d.wrapped[:len(d.wrapped)-1]
	return removed, true
}

func (d ReplicaSet) InAtomicReplicationChange() bool {
	__antithesis_instrumentation__.Notify(176824)
	for _, rDesc := range d.wrapped {
		__antithesis_instrumentation__.Notify(176826)
		switch rDesc.GetType() {
		case VOTER_INCOMING, VOTER_OUTGOING, VOTER_DEMOTING_LEARNER,
			VOTER_DEMOTING_NON_VOTER:
			__antithesis_instrumentation__.Notify(176827)
			return true
		case VOTER_FULL, LEARNER, NON_VOTER:
			__antithesis_instrumentation__.Notify(176828)
		default:
			__antithesis_instrumentation__.Notify(176829)
			panic(fmt.Sprintf("unknown replica type %d", rDesc.GetType()))
		}
	}
	__antithesis_instrumentation__.Notify(176825)
	return false
}

func (d ReplicaSet) ConfState() raftpb.ConfState {
	__antithesis_instrumentation__.Notify(176830)
	var cs raftpb.ConfState
	joint := d.InAtomicReplicationChange()

	for _, rep := range d.wrapped {
		__antithesis_instrumentation__.Notify(176832)
		id := uint64(rep.ReplicaID)
		typ := rep.GetType()
		switch typ {
		case VOTER_FULL:
			__antithesis_instrumentation__.Notify(176833)
			cs.Voters = append(cs.Voters, id)
			if joint {
				__antithesis_instrumentation__.Notify(176840)
				cs.VotersOutgoing = append(cs.VotersOutgoing, id)
			} else {
				__antithesis_instrumentation__.Notify(176841)
			}
		case VOTER_INCOMING:
			__antithesis_instrumentation__.Notify(176834)
			cs.Voters = append(cs.Voters, id)
		case VOTER_OUTGOING:
			__antithesis_instrumentation__.Notify(176835)
			cs.VotersOutgoing = append(cs.VotersOutgoing, id)
		case VOTER_DEMOTING_LEARNER, VOTER_DEMOTING_NON_VOTER:
			__antithesis_instrumentation__.Notify(176836)
			cs.VotersOutgoing = append(cs.VotersOutgoing, id)
			cs.LearnersNext = append(cs.LearnersNext, id)
		case LEARNER:
			__antithesis_instrumentation__.Notify(176837)
			cs.Learners = append(cs.Learners, id)
		case NON_VOTER:
			__antithesis_instrumentation__.Notify(176838)
			cs.Learners = append(cs.Learners, id)
		default:
			__antithesis_instrumentation__.Notify(176839)
			panic(fmt.Sprintf("unknown ReplicaType %d", typ))
		}
	}
	__antithesis_instrumentation__.Notify(176831)
	return cs
}

func (d ReplicaSet) CanMakeProgress(liveFunc func(descriptor ReplicaDescriptor) bool) bool {
	__antithesis_instrumentation__.Notify(176842)
	return d.ReplicationStatus(liveFunc, 0).Available
}

type RangeStatusReport struct {
	Available bool

	UnderReplicated bool

	OverReplicated bool
}

func (d ReplicaSet) ReplicationStatus(
	liveFunc func(descriptor ReplicaDescriptor) bool, neededVoters int,
) RangeStatusReport {
	__antithesis_instrumentation__.Notify(176843)
	var res RangeStatusReport

	isBoth := func(
		pred1 func(rDesc ReplicaDescriptor) bool,
		pred2 func(rDesc ReplicaDescriptor) bool) func(ReplicaDescriptor) bool {
		__antithesis_instrumentation__.Notify(176845)
		return func(rDesc ReplicaDescriptor) bool {
			__antithesis_instrumentation__.Notify(176846)
			return pred1(rDesc) && func() bool {
				__antithesis_instrumentation__.Notify(176847)
				return pred2(rDesc) == true
			}() == true
		}
	}
	__antithesis_instrumentation__.Notify(176844)

	votersOldGroup := d.FilterToDescriptors(ReplicaDescriptor.IsVoterOldConfig)
	liveVotersOldGroup := d.FilterToDescriptors(isBoth(ReplicaDescriptor.IsVoterOldConfig, liveFunc))

	n := len(votersOldGroup)

	availableOutgoingGroup := (n == 0) || func() bool {
		__antithesis_instrumentation__.Notify(176848)
		return (len(liveVotersOldGroup) >= n/2+1) == true
	}() == true

	votersNewGroup := d.FilterToDescriptors(ReplicaDescriptor.IsVoterNewConfig)
	liveVotersNewGroup := d.FilterToDescriptors(isBoth(ReplicaDescriptor.IsVoterNewConfig, liveFunc))

	n = len(votersNewGroup)
	availableIncomingGroup := len(liveVotersNewGroup) >= n/2+1

	res.Available = availableIncomingGroup && func() bool {
		__antithesis_instrumentation__.Notify(176849)
		return availableOutgoingGroup == true
	}() == true

	underReplicatedOldGroup := len(liveVotersOldGroup) < neededVoters
	underReplicatedNewGroup := len(liveVotersNewGroup) < neededVoters
	overReplicatedOldGroup := len(votersOldGroup) > neededVoters
	overReplicatedNewGroup := len(votersNewGroup) > neededVoters
	res.UnderReplicated = underReplicatedOldGroup || func() bool {
		__antithesis_instrumentation__.Notify(176850)
		return underReplicatedNewGroup == true
	}() == true
	res.OverReplicated = overReplicatedOldGroup || func() bool {
		__antithesis_instrumentation__.Notify(176851)
		return overReplicatedNewGroup == true
	}() == true
	return res
}

func Empty(target ReplicationTarget) bool {
	__antithesis_instrumentation__.Notify(176852)
	return target == ReplicationTarget{}
}

func (d ReplicaSet) ReplicationTargets() (out []ReplicationTarget) {
	__antithesis_instrumentation__.Notify(176853)
	descs := d.Descriptors()
	out = make([]ReplicationTarget, len(descs))
	for i := range descs {
		__antithesis_instrumentation__.Notify(176855)
		repl := &descs[i]
		out[i].NodeID, out[i].StoreID = repl.NodeID, repl.StoreID
	}
	__antithesis_instrumentation__.Notify(176854)
	return out
}

func (c ReplicaChangeType) IsAddition() bool {
	__antithesis_instrumentation__.Notify(176856)
	switch c {
	case ADD_NON_VOTER, ADD_VOTER:
		__antithesis_instrumentation__.Notify(176857)
		return true
	case REMOVE_NON_VOTER, REMOVE_VOTER:
		__antithesis_instrumentation__.Notify(176858)
		return false
	default:
		__antithesis_instrumentation__.Notify(176859)
		panic(fmt.Sprintf("unexpected ReplicaChangeType %s", c))
	}
}

func (c ReplicaChangeType) IsRemoval() bool {
	__antithesis_instrumentation__.Notify(176860)
	switch c {
	case ADD_NON_VOTER, ADD_VOTER:
		__antithesis_instrumentation__.Notify(176861)
		return false
	case REMOVE_NON_VOTER, REMOVE_VOTER:
		__antithesis_instrumentation__.Notify(176862)
		return true
	default:
		__antithesis_instrumentation__.Notify(176863)
		panic(fmt.Sprintf("unexpected ReplicaChangeType %s", c))
	}
}

var errReplicaNotFound = errors.Errorf(`replica not found in RangeDescriptor`)
var errReplicaCannotHoldLease = errors.Errorf("replica cannot hold lease")

func CheckCanReceiveLease(wouldbeLeaseholder ReplicaDescriptor, rngDesc *RangeDescriptor) error {
	__antithesis_instrumentation__.Notify(176864)
	repDesc, ok := rngDesc.GetReplicaDescriptorByID(wouldbeLeaseholder.ReplicaID)
	if !ok {
		__antithesis_instrumentation__.Notify(176866)
		return errReplicaNotFound
	} else {
		__antithesis_instrumentation__.Notify(176867)
		if !repDesc.IsVoterNewConfig() {
			__antithesis_instrumentation__.Notify(176868)
			return errReplicaCannotHoldLease
		} else {
			__antithesis_instrumentation__.Notify(176869)
		}
	}
	__antithesis_instrumentation__.Notify(176865)
	return nil
}
