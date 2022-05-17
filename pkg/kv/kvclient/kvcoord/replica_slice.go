package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/shuffle"
)

type ReplicaInfo struct {
	roachpb.ReplicaDescriptor
	NodeDesc *roachpb.NodeDescriptor
}

func (i ReplicaInfo) locality() []roachpb.Tier {
	__antithesis_instrumentation__.Notify(87894)
	return i.NodeDesc.Locality.Tiers
}

func (i ReplicaInfo) addr() string {
	__antithesis_instrumentation__.Notify(87895)
	return i.NodeDesc.Address.String()
}

type ReplicaSlice []ReplicaInfo

type ReplicaSliceFilter int

const (
	OnlyPotentialLeaseholders ReplicaSliceFilter = iota

	AllExtantReplicas
)

func NewReplicaSlice(
	ctx context.Context,
	nodeDescs NodeDescStore,
	desc *roachpb.RangeDescriptor,
	leaseholder *roachpb.ReplicaDescriptor,
	filter ReplicaSliceFilter,
) (ReplicaSlice, error) {
	__antithesis_instrumentation__.Notify(87896)
	if leaseholder != nil {
		__antithesis_instrumentation__.Notify(87903)
		if _, ok := desc.GetReplicaDescriptorByID(leaseholder.ReplicaID); !ok {
			__antithesis_instrumentation__.Notify(87904)
			log.Fatalf(ctx, "leaseholder not in descriptor; leaseholder: %s, desc: %s", leaseholder, desc)
		} else {
			__antithesis_instrumentation__.Notify(87905)
		}
	} else {
		__antithesis_instrumentation__.Notify(87906)
	}
	__antithesis_instrumentation__.Notify(87897)
	canReceiveLease := func(rDesc roachpb.ReplicaDescriptor) bool {
		__antithesis_instrumentation__.Notify(87907)
		if err := roachpb.CheckCanReceiveLease(rDesc, desc); err != nil {
			__antithesis_instrumentation__.Notify(87909)
			return false
		} else {
			__antithesis_instrumentation__.Notify(87910)
		}
		__antithesis_instrumentation__.Notify(87908)
		return true
	}
	__antithesis_instrumentation__.Notify(87898)

	var replicas []roachpb.ReplicaDescriptor
	switch filter {
	case OnlyPotentialLeaseholders:
		__antithesis_instrumentation__.Notify(87911)
		replicas = desc.Replicas().Filter(canReceiveLease).Descriptors()
	case AllExtantReplicas:
		__antithesis_instrumentation__.Notify(87912)
		replicas = desc.Replicas().VoterAndNonVoterDescriptors()
	default:
		__antithesis_instrumentation__.Notify(87913)
		log.Fatalf(ctx, "unknown ReplicaSliceFilter %v", filter)
	}
	__antithesis_instrumentation__.Notify(87899)

	if leaseholder != nil && func() bool {
		__antithesis_instrumentation__.Notify(87914)
		return len(replicas) < len(desc.Replicas().Descriptors()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(87915)
		found := false
		for _, v := range replicas {
			__antithesis_instrumentation__.Notify(87917)
			if v == *leaseholder {
				__antithesis_instrumentation__.Notify(87918)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(87919)
			}
		}
		__antithesis_instrumentation__.Notify(87916)
		if !found {
			__antithesis_instrumentation__.Notify(87920)
			log.Eventf(ctx, "the descriptor has the leaseholder as a learner; including it anyway")
			replicas = append(replicas, *leaseholder)
		} else {
			__antithesis_instrumentation__.Notify(87921)
		}
	} else {
		__antithesis_instrumentation__.Notify(87922)
	}
	__antithesis_instrumentation__.Notify(87900)
	rs := make(ReplicaSlice, 0, len(replicas))
	for _, r := range replicas {
		__antithesis_instrumentation__.Notify(87923)
		nd, err := nodeDescs.GetNodeDescriptor(r.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(87925)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(87927)
				log.Infof(ctx, "node %d is not gossiped: %v", r.NodeID, err)
			} else {
				__antithesis_instrumentation__.Notify(87928)
			}
			__antithesis_instrumentation__.Notify(87926)
			continue
		} else {
			__antithesis_instrumentation__.Notify(87929)
		}
		__antithesis_instrumentation__.Notify(87924)
		rs = append(rs, ReplicaInfo{
			ReplicaDescriptor: r,
			NodeDesc:          nd,
		})
	}
	__antithesis_instrumentation__.Notify(87901)
	if len(rs) == 0 {
		__antithesis_instrumentation__.Notify(87930)
		return nil, newSendError(
			fmt.Sprintf("no replica node addresses available via gossip for r%d", desc.RangeID))
	} else {
		__antithesis_instrumentation__.Notify(87931)
	}
	__antithesis_instrumentation__.Notify(87902)
	return rs, nil
}

var _ shuffle.Interface = ReplicaSlice{}

func (rs ReplicaSlice) Len() int { __antithesis_instrumentation__.Notify(87932); return len(rs) }

func (rs ReplicaSlice) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(87933)
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs ReplicaSlice) Find(id roachpb.ReplicaID) int {
	__antithesis_instrumentation__.Notify(87934)
	for i := range rs {
		__antithesis_instrumentation__.Notify(87936)
		if rs[i].ReplicaID == id {
			__antithesis_instrumentation__.Notify(87937)
			return i
		} else {
			__antithesis_instrumentation__.Notify(87938)
		}
	}
	__antithesis_instrumentation__.Notify(87935)
	return -1
}

func (rs ReplicaSlice) MoveToFront(i int) {
	__antithesis_instrumentation__.Notify(87939)
	if i >= len(rs) {
		__antithesis_instrumentation__.Notify(87941)
		panic("out of bound index")
	} else {
		__antithesis_instrumentation__.Notify(87942)
	}
	__antithesis_instrumentation__.Notify(87940)
	front := rs[i]

	copy(rs[1:], rs[:i])
	rs[0] = front
}

func localityMatch(a, b []roachpb.Tier) int {
	__antithesis_instrumentation__.Notify(87943)
	if len(a) == 0 {
		__antithesis_instrumentation__.Notify(87946)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(87947)
	}
	__antithesis_instrumentation__.Notify(87944)
	for i := range a {
		__antithesis_instrumentation__.Notify(87948)
		if i >= len(b) || func() bool {
			__antithesis_instrumentation__.Notify(87949)
			return a[i] != b[i] == true
		}() == true {
			__antithesis_instrumentation__.Notify(87950)
			return i
		} else {
			__antithesis_instrumentation__.Notify(87951)
		}
	}
	__antithesis_instrumentation__.Notify(87945)
	return len(a)
}

type LatencyFunc func(string) (time.Duration, bool)

func (rs ReplicaSlice) OptimizeReplicaOrder(
	nodeDesc *roachpb.NodeDescriptor, latencyFn LatencyFunc,
) {
	__antithesis_instrumentation__.Notify(87952)

	if nodeDesc == nil {
		__antithesis_instrumentation__.Notify(87954)
		shuffle.Shuffle(rs)
		return
	} else {
		__antithesis_instrumentation__.Notify(87955)
	}
	__antithesis_instrumentation__.Notify(87953)

	sort.Slice(rs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(87956)

		if rs[i].NodeID == rs[j].NodeID {
			__antithesis_instrumentation__.Notify(87961)
			return false
		} else {
			__antithesis_instrumentation__.Notify(87962)
		}
		__antithesis_instrumentation__.Notify(87957)

		if rs[i].NodeID == nodeDesc.NodeID {
			__antithesis_instrumentation__.Notify(87963)
			return true
		} else {
			__antithesis_instrumentation__.Notify(87964)
		}
		__antithesis_instrumentation__.Notify(87958)
		if rs[j].NodeID == nodeDesc.NodeID {
			__antithesis_instrumentation__.Notify(87965)
			return false
		} else {
			__antithesis_instrumentation__.Notify(87966)
		}
		__antithesis_instrumentation__.Notify(87959)

		if latencyFn != nil {
			__antithesis_instrumentation__.Notify(87967)
			latencyI, okI := latencyFn(rs[i].addr())
			latencyJ, okJ := latencyFn(rs[j].addr())
			if okI && func() bool {
				__antithesis_instrumentation__.Notify(87968)
				return okJ == true
			}() == true {
				__antithesis_instrumentation__.Notify(87969)
				return latencyI < latencyJ
			} else {
				__antithesis_instrumentation__.Notify(87970)
			}
		} else {
			__antithesis_instrumentation__.Notify(87971)
		}
		__antithesis_instrumentation__.Notify(87960)
		attrMatchI := localityMatch(nodeDesc.Locality.Tiers, rs[i].locality())
		attrMatchJ := localityMatch(nodeDesc.Locality.Tiers, rs[j].locality())

		return attrMatchI > attrMatchJ
	})
}

func (rs ReplicaSlice) Descriptors() []roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(87972)
	reps := make([]roachpb.ReplicaDescriptor, len(rs))
	for i := range rs {
		__antithesis_instrumentation__.Notify(87974)
		reps[i] = rs[i].ReplicaDescriptor
	}
	__antithesis_instrumentation__.Notify(87973)
	return reps
}
