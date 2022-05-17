package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type NodeID int32

func (n NodeID) String() string {
	__antithesis_instrumentation__.Notify(173832)
	return strconv.FormatInt(int64(n), 10)
}

func (n NodeID) SafeValue() { __antithesis_instrumentation__.Notify(173833) }

type StoreID int32

type StoreIDSlice []StoreID

func (s StoreIDSlice) Len() int { __antithesis_instrumentation__.Notify(173834); return len(s) }
func (s StoreIDSlice) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(173835)
	s[i], s[j] = s[j], s[i]
}
func (s StoreIDSlice) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(173836)
	return s[i] < s[j]
}

func (n StoreID) String() string {
	__antithesis_instrumentation__.Notify(173837)
	return strconv.FormatInt(int64(n), 10)
}

func (n StoreID) SafeValue() { __antithesis_instrumentation__.Notify(173838) }

type RangeID int64

func (r RangeID) String() string {
	__antithesis_instrumentation__.Notify(173839)
	return strconv.FormatInt(int64(r), 10)
}

func (r RangeID) SafeValue() { __antithesis_instrumentation__.Notify(173840) }

type RangeIDSlice []RangeID

func (r RangeIDSlice) Len() int { __antithesis_instrumentation__.Notify(173841); return len(r) }
func (r RangeIDSlice) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(173842)
	r[i], r[j] = r[j], r[i]
}
func (r RangeIDSlice) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(173843)
	return r[i] < r[j]
}

type ReplicaID int32

func (r ReplicaID) String() string {
	__antithesis_instrumentation__.Notify(173844)
	return strconv.FormatInt(int64(r), 10)
}

func (r ReplicaID) SafeValue() { __antithesis_instrumentation__.Notify(173845) }

func (a Attributes) Equals(b Attributes) bool {
	__antithesis_instrumentation__.Notify(173846)

	if len(a.Attrs) != len(b.Attrs) {
		__antithesis_instrumentation__.Notify(173849)
		return false
	} else {
		__antithesis_instrumentation__.Notify(173850)
	}
	__antithesis_instrumentation__.Notify(173847)
	for _, aAttr := range a.Attrs {
		__antithesis_instrumentation__.Notify(173851)
		var found bool
		for _, bAttr := range b.Attrs {
			__antithesis_instrumentation__.Notify(173853)
			if aAttr == bAttr {
				__antithesis_instrumentation__.Notify(173854)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(173855)
			}
		}
		__antithesis_instrumentation__.Notify(173852)
		if !found {
			__antithesis_instrumentation__.Notify(173856)
			return false
		} else {
			__antithesis_instrumentation__.Notify(173857)
		}
	}
	__antithesis_instrumentation__.Notify(173848)
	return true
}

func (a Attributes) String() string {
	__antithesis_instrumentation__.Notify(173858)
	return strings.Join(a.Attrs, ",")
}

type RangeGeneration int64

func (g RangeGeneration) String() string {
	__antithesis_instrumentation__.Notify(173859)
	return strconv.FormatInt(int64(g), 10)
}

func (g RangeGeneration) SafeValue() { __antithesis_instrumentation__.Notify(173860) }

func NewRangeDescriptor(rangeID RangeID, start, end RKey, replicas ReplicaSet) *RangeDescriptor {
	__antithesis_instrumentation__.Notify(173861)
	repls := append([]ReplicaDescriptor(nil), replicas.Descriptors()...)
	for i := range repls {
		__antithesis_instrumentation__.Notify(173863)
		repls[i].ReplicaID = ReplicaID(i + 1)
	}
	__antithesis_instrumentation__.Notify(173862)
	desc := &RangeDescriptor{
		RangeID:       rangeID,
		StartKey:      start,
		EndKey:        end,
		NextReplicaID: ReplicaID(len(repls) + 1),
	}
	desc.SetReplicas(MakeReplicaSet(repls))
	return desc
}

func (r *RangeDescriptor) Equal(other *RangeDescriptor) bool {
	__antithesis_instrumentation__.Notify(173864)
	if other == nil {
		__antithesis_instrumentation__.Notify(173875)
		return r == nil
	} else {
		__antithesis_instrumentation__.Notify(173876)
	}
	__antithesis_instrumentation__.Notify(173865)
	if r == nil {
		__antithesis_instrumentation__.Notify(173877)
		return false
	} else {
		__antithesis_instrumentation__.Notify(173878)
	}
	__antithesis_instrumentation__.Notify(173866)
	if r.RangeID != other.RangeID {
		__antithesis_instrumentation__.Notify(173879)
		return false
	} else {
		__antithesis_instrumentation__.Notify(173880)
	}
	__antithesis_instrumentation__.Notify(173867)
	if r.Generation != other.Generation {
		__antithesis_instrumentation__.Notify(173881)
		return false
	} else {
		__antithesis_instrumentation__.Notify(173882)
	}
	__antithesis_instrumentation__.Notify(173868)
	if !bytes.Equal(r.StartKey, other.StartKey) {
		__antithesis_instrumentation__.Notify(173883)
		return false
	} else {
		__antithesis_instrumentation__.Notify(173884)
	}
	__antithesis_instrumentation__.Notify(173869)
	if !bytes.Equal(r.EndKey, other.EndKey) {
		__antithesis_instrumentation__.Notify(173885)
		return false
	} else {
		__antithesis_instrumentation__.Notify(173886)
	}
	__antithesis_instrumentation__.Notify(173870)
	if len(r.InternalReplicas) != len(other.InternalReplicas) {
		__antithesis_instrumentation__.Notify(173887)
		return false
	} else {
		__antithesis_instrumentation__.Notify(173888)
	}
	__antithesis_instrumentation__.Notify(173871)
	for i := range r.InternalReplicas {
		__antithesis_instrumentation__.Notify(173889)
		if !r.InternalReplicas[i].Equal(&other.InternalReplicas[i]) {
			__antithesis_instrumentation__.Notify(173890)
			return false
		} else {
			__antithesis_instrumentation__.Notify(173891)
		}
	}
	__antithesis_instrumentation__.Notify(173872)
	if r.NextReplicaID != other.NextReplicaID {
		__antithesis_instrumentation__.Notify(173892)
		return false
	} else {
		__antithesis_instrumentation__.Notify(173893)
	}
	__antithesis_instrumentation__.Notify(173873)
	if !r.StickyBit.Equal(other.StickyBit) {
		__antithesis_instrumentation__.Notify(173894)
		return false
	} else {
		__antithesis_instrumentation__.Notify(173895)
	}
	__antithesis_instrumentation__.Notify(173874)
	return true
}

func (r *RangeDescriptor) GetRangeID() RangeID {
	__antithesis_instrumentation__.Notify(173896)
	return r.RangeID
}

func (r *RangeDescriptor) GetStartKey() RKey {
	__antithesis_instrumentation__.Notify(173897)
	return r.StartKey
}

func (r *RangeDescriptor) RSpan() RSpan {
	__antithesis_instrumentation__.Notify(173898)
	return RSpan{Key: r.StartKey, EndKey: r.EndKey}
}

func (r *RangeDescriptor) KeySpan() RSpan {
	__antithesis_instrumentation__.Notify(173899)
	start := r.StartKey
	if r.StartKey.Equal(RKeyMin) {
		__antithesis_instrumentation__.Notify(173901)

		start = RKey(LocalMax)
	} else {
		__antithesis_instrumentation__.Notify(173902)
	}
	__antithesis_instrumentation__.Notify(173900)
	return RSpan{
		Key:    start,
		EndKey: r.EndKey,
	}
}

func (r *RangeDescriptor) ContainsKey(key RKey) bool {
	__antithesis_instrumentation__.Notify(173903)
	return r.RSpan().ContainsKey(key)
}

func (r *RangeDescriptor) ContainsKeyInverted(key RKey) bool {
	__antithesis_instrumentation__.Notify(173904)
	return r.RSpan().ContainsKeyInverted(key)
}

func (r *RangeDescriptor) ContainsKeyRange(start, end RKey) bool {
	__antithesis_instrumentation__.Notify(173905)
	return r.RSpan().ContainsKeyRange(start, end)
}

func (r *RangeDescriptor) Replicas() ReplicaSet {
	__antithesis_instrumentation__.Notify(173906)
	return MakeReplicaSet(r.InternalReplicas)
}

func (r *RangeDescriptor) SetReplicas(replicas ReplicaSet) {
	__antithesis_instrumentation__.Notify(173907)
	r.InternalReplicas = replicas.AsProto()
}

func (r *RangeDescriptor) SetReplicaType(
	nodeID NodeID, storeID StoreID, typ ReplicaType,
) (ReplicaDescriptor, ReplicaType, bool) {
	__antithesis_instrumentation__.Notify(173908)
	for i := range r.InternalReplicas {
		__antithesis_instrumentation__.Notify(173910)
		desc := &r.InternalReplicas[i]
		if desc.StoreID == storeID && func() bool {
			__antithesis_instrumentation__.Notify(173911)
			return desc.NodeID == nodeID == true
		}() == true {
			__antithesis_instrumentation__.Notify(173912)
			prevTyp := desc.GetType()
			if typ != VOTER_FULL {
				__antithesis_instrumentation__.Notify(173914)
				desc.Type = &typ
			} else {
				__antithesis_instrumentation__.Notify(173915)

				desc.Type = nil
			}
			__antithesis_instrumentation__.Notify(173913)
			return *desc, prevTyp, true
		} else {
			__antithesis_instrumentation__.Notify(173916)
		}
	}
	__antithesis_instrumentation__.Notify(173909)
	return ReplicaDescriptor{}, 0, false
}

func (r *RangeDescriptor) AddReplica(
	nodeID NodeID, storeID StoreID, typ ReplicaType,
) ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(173917)
	var typPtr *ReplicaType

	if typ != VOTER_FULL {
		__antithesis_instrumentation__.Notify(173919)
		typPtr = &typ
	} else {
		__antithesis_instrumentation__.Notify(173920)
	}
	__antithesis_instrumentation__.Notify(173918)
	toAdd := ReplicaDescriptor{
		NodeID:    nodeID,
		StoreID:   storeID,
		ReplicaID: r.NextReplicaID,
		Type:      typPtr,
	}
	rs := r.Replicas()
	rs.AddReplica(toAdd)
	r.SetReplicas(rs)
	r.NextReplicaID++
	return toAdd
}

func (r *RangeDescriptor) RemoveReplica(nodeID NodeID, storeID StoreID) (ReplicaDescriptor, bool) {
	__antithesis_instrumentation__.Notify(173921)
	rs := r.Replicas()
	removedRepl, ok := rs.RemoveReplica(nodeID, storeID)
	if ok {
		__antithesis_instrumentation__.Notify(173923)
		r.SetReplicas(rs)
	} else {
		__antithesis_instrumentation__.Notify(173924)
	}
	__antithesis_instrumentation__.Notify(173922)
	return removedRepl, ok
}

func (r *RangeDescriptor) GetReplicaDescriptor(storeID StoreID) (ReplicaDescriptor, bool) {
	__antithesis_instrumentation__.Notify(173925)
	for _, repDesc := range r.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(173927)
		if repDesc.StoreID == storeID {
			__antithesis_instrumentation__.Notify(173928)
			return repDesc, true
		} else {
			__antithesis_instrumentation__.Notify(173929)
		}
	}
	__antithesis_instrumentation__.Notify(173926)
	return ReplicaDescriptor{}, false
}

func (r *RangeDescriptor) GetReplicaDescriptorByID(replicaID ReplicaID) (ReplicaDescriptor, bool) {
	__antithesis_instrumentation__.Notify(173930)
	for _, repDesc := range r.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(173932)
		if repDesc.ReplicaID == replicaID {
			__antithesis_instrumentation__.Notify(173933)
			return repDesc, true
		} else {
			__antithesis_instrumentation__.Notify(173934)
		}
	}
	__antithesis_instrumentation__.Notify(173931)
	return ReplicaDescriptor{}, false
}

func (r *RangeDescriptor) ContainsVoterIncoming() bool {
	__antithesis_instrumentation__.Notify(173935)
	for _, repDesc := range r.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(173937)
		if repDesc.GetType() == VOTER_INCOMING {
			__antithesis_instrumentation__.Notify(173938)
			return true
		} else {
			__antithesis_instrumentation__.Notify(173939)
		}
	}
	__antithesis_instrumentation__.Notify(173936)
	return false
}

func (r *RangeDescriptor) IsInitialized() bool {
	__antithesis_instrumentation__.Notify(173940)
	return len(r.EndKey) != 0
}

func (r *RangeDescriptor) IncrementGeneration() {
	__antithesis_instrumentation__.Notify(173941)
	r.Generation++
}

func (r *RangeDescriptor) GetStickyBit() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(173942)
	if r.StickyBit == nil {
		__antithesis_instrumentation__.Notify(173944)
		return hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(173945)
	}
	__antithesis_instrumentation__.Notify(173943)
	return *r.StickyBit
}

func (r *RangeDescriptor) Validate() error {
	__antithesis_instrumentation__.Notify(173946)
	if r.NextReplicaID == 0 {
		__antithesis_instrumentation__.Notify(173949)
		return errors.Errorf("NextReplicaID must be non-zero")
	} else {
		__antithesis_instrumentation__.Notify(173950)
	}
	__antithesis_instrumentation__.Notify(173947)
	seen := map[ReplicaID]struct{}{}
	stores := map[StoreID]struct{}{}
	for i, rep := range r.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(173951)
		if err := rep.Validate(); err != nil {
			__antithesis_instrumentation__.Notify(173956)
			return errors.Wrapf(err, "replica %d is invalid", i)
		} else {
			__antithesis_instrumentation__.Notify(173957)
		}
		__antithesis_instrumentation__.Notify(173952)
		if rep.ReplicaID >= r.NextReplicaID {
			__antithesis_instrumentation__.Notify(173958)
			return errors.Errorf("ReplicaID %d must be less than NextReplicaID %d",
				rep.ReplicaID, r.NextReplicaID)
		} else {
			__antithesis_instrumentation__.Notify(173959)
		}
		__antithesis_instrumentation__.Notify(173953)

		if _, ok := seen[rep.ReplicaID]; ok {
			__antithesis_instrumentation__.Notify(173960)
			return errors.Errorf("ReplicaID %d was reused", rep.ReplicaID)
		} else {
			__antithesis_instrumentation__.Notify(173961)
		}
		__antithesis_instrumentation__.Notify(173954)
		seen[rep.ReplicaID] = struct{}{}

		if _, ok := stores[rep.StoreID]; ok {
			__antithesis_instrumentation__.Notify(173962)
			return errors.Errorf("StoreID %d was reused", rep.StoreID)
		} else {
			__antithesis_instrumentation__.Notify(173963)
		}
		__antithesis_instrumentation__.Notify(173955)
		stores[rep.StoreID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(173948)
	return nil
}

func (r RangeDescriptor) String() string {
	__antithesis_instrumentation__.Notify(173964)
	return redact.StringWithoutMarkers(r)
}

func (r RangeDescriptor) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(173965)
	w.Printf("r%d:", r.RangeID)
	if !r.IsInitialized() {
		__antithesis_instrumentation__.Notify(173969)
		w.SafeString("{-}")
	} else {
		__antithesis_instrumentation__.Notify(173970)
		w.Print(r.RSpan())
	}
	__antithesis_instrumentation__.Notify(173966)
	w.SafeString(" [")

	if allReplicas := r.Replicas().Descriptors(); len(allReplicas) > 0 {
		__antithesis_instrumentation__.Notify(173971)
		for i, rep := range allReplicas {
			__antithesis_instrumentation__.Notify(173972)
			if i > 0 {
				__antithesis_instrumentation__.Notify(173974)
				w.SafeString(", ")
			} else {
				__antithesis_instrumentation__.Notify(173975)
			}
			__antithesis_instrumentation__.Notify(173973)
			w.Print(rep)
		}
	} else {
		__antithesis_instrumentation__.Notify(173976)
		w.SafeString("<no replicas>")
	}
	__antithesis_instrumentation__.Notify(173967)
	w.Printf(", next=%d, gen=%d", r.NextReplicaID, r.Generation)
	if s := r.GetStickyBit(); !s.IsEmpty() {
		__antithesis_instrumentation__.Notify(173977)
		w.Printf(", sticky=%s", s)
	} else {
		__antithesis_instrumentation__.Notify(173978)
	}
	__antithesis_instrumentation__.Notify(173968)
	w.SafeString("]")
}

func (r ReplicationTarget) String() string {
	__antithesis_instrumentation__.Notify(173979)
	return redact.StringWithoutMarkers(r)
}

func (r ReplicationTarget) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(173980)
	w.Printf("n%d,s%d", r.NodeID, r.StoreID)
}

func (r ReplicaDescriptor) String() string {
	__antithesis_instrumentation__.Notify(173981)
	return redact.StringWithoutMarkers(r)
}

func (r ReplicaDescriptor) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(173982)
	w.Printf("(n%d,s%d):", r.NodeID, r.StoreID)
	if r.ReplicaID == 0 {
		__antithesis_instrumentation__.Notify(173984)
		w.SafeRune('?')
	} else {
		__antithesis_instrumentation__.Notify(173985)
		w.Print(r.ReplicaID)
	}
	__antithesis_instrumentation__.Notify(173983)
	if typ := r.GetType(); typ != VOTER_FULL {
		__antithesis_instrumentation__.Notify(173986)
		w.Print(typ)
	} else {
		__antithesis_instrumentation__.Notify(173987)
	}
}

func (r ReplicaDescriptor) Validate() error {
	__antithesis_instrumentation__.Notify(173988)
	if r.NodeID == 0 {
		__antithesis_instrumentation__.Notify(173992)
		return errors.Errorf("NodeID must not be zero")
	} else {
		__antithesis_instrumentation__.Notify(173993)
	}
	__antithesis_instrumentation__.Notify(173989)
	if r.StoreID == 0 {
		__antithesis_instrumentation__.Notify(173994)
		return errors.Errorf("StoreID must not be zero")
	} else {
		__antithesis_instrumentation__.Notify(173995)
	}
	__antithesis_instrumentation__.Notify(173990)
	if r.ReplicaID == 0 {
		__antithesis_instrumentation__.Notify(173996)
		return errors.Errorf("ReplicaID must not be zero")
	} else {
		__antithesis_instrumentation__.Notify(173997)
	}
	__antithesis_instrumentation__.Notify(173991)
	return nil
}

func (r ReplicaDescriptor) GetType() ReplicaType {
	__antithesis_instrumentation__.Notify(173998)
	if r.Type == nil {
		__antithesis_instrumentation__.Notify(174000)
		return VOTER_FULL
	} else {
		__antithesis_instrumentation__.Notify(174001)
	}
	__antithesis_instrumentation__.Notify(173999)
	return *r.Type
}

func (r ReplicaType) SafeValue() { __antithesis_instrumentation__.Notify(174002) }

func (r ReplicaDescriptor) IsVoterOldConfig() bool {
	__antithesis_instrumentation__.Notify(174003)
	switch r.GetType() {
	case VOTER_FULL, VOTER_OUTGOING, VOTER_DEMOTING_NON_VOTER, VOTER_DEMOTING_LEARNER:
		__antithesis_instrumentation__.Notify(174004)
		return true
	default:
		__antithesis_instrumentation__.Notify(174005)
		return false
	}
}

func (r ReplicaDescriptor) IsVoterNewConfig() bool {
	__antithesis_instrumentation__.Notify(174006)
	switch r.GetType() {
	case VOTER_FULL, VOTER_INCOMING:
		__antithesis_instrumentation__.Notify(174007)
		return true
	default:
		__antithesis_instrumentation__.Notify(174008)
		return false
	}
}

func (r ReplicaDescriptor) IsAnyVoter() bool {
	__antithesis_instrumentation__.Notify(174009)
	switch r.GetType() {
	case VOTER_FULL, VOTER_INCOMING, VOTER_OUTGOING, VOTER_DEMOTING_NON_VOTER, VOTER_DEMOTING_LEARNER:
		__antithesis_instrumentation__.Notify(174010)
		return true
	default:
		__antithesis_instrumentation__.Notify(174011)
		return false
	}
}

func PercentilesFromData(data []float64) Percentiles {
	__antithesis_instrumentation__.Notify(174012)
	sort.Float64s(data)

	return Percentiles{
		P10:  percentileFromSortedData(data, 10),
		P25:  percentileFromSortedData(data, 25),
		P50:  percentileFromSortedData(data, 50),
		P75:  percentileFromSortedData(data, 75),
		P90:  percentileFromSortedData(data, 90),
		PMax: percentileFromSortedData(data, 100),
	}
}

func percentileFromSortedData(data []float64, percent float64) float64 {
	__antithesis_instrumentation__.Notify(174013)
	if len(data) == 0 {
		__antithesis_instrumentation__.Notify(174017)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(174018)
	}
	__antithesis_instrumentation__.Notify(174014)
	if percent < 0 {
		__antithesis_instrumentation__.Notify(174019)
		percent = 0
	} else {
		__antithesis_instrumentation__.Notify(174020)
	}
	__antithesis_instrumentation__.Notify(174015)
	if percent >= 100 {
		__antithesis_instrumentation__.Notify(174021)
		return data[len(data)-1]
	} else {
		__antithesis_instrumentation__.Notify(174022)
	}
	__antithesis_instrumentation__.Notify(174016)

	idx := int(float64(len(data)) * percent / 100.0)
	return data[idx]
}

func (p Percentiles) String() string {
	__antithesis_instrumentation__.Notify(174023)
	return redact.StringWithoutMarkers(p)
}

func (p Percentiles) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(174024)
	w.Printf("p10=%.2f p25=%.2f p50=%.2f p75=%.2f p90=%.2f pMax=%.2f",
		p.P10, p.P25, p.P50, p.P75, p.P90, p.PMax)
}

func (sc FileStoreProperties) String() string {
	__antithesis_instrumentation__.Notify(174025)
	return redact.StringWithoutMarkers(sc)
}

func (sc FileStoreProperties) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(174026)
	w.Printf("{path=%s, fs=%s, blkdev=%s, mnt=%s opts=%s}",
		sc.Path,
		redact.SafeString(sc.FsType),
		sc.BlockDevice,
		sc.MountPoint,
		sc.MountOptions)
}

func (sc StoreCapacity) String() string {
	__antithesis_instrumentation__.Notify(174027)
	return redact.StringWithoutMarkers(sc)
}

func (sc StoreCapacity) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(174028)
	w.Printf("disk (capacity=%s, available=%s, used=%s, logicalBytes=%s), "+
		"ranges=%d, leases=%d, queries=%.2f, writes=%.2f, "+
		"l0Sublevels=%d, bytesPerReplica={%s}, writesPerReplica={%s}",
		humanizeutil.IBytes(sc.Capacity), humanizeutil.IBytes(sc.Available),
		humanizeutil.IBytes(sc.Used), humanizeutil.IBytes(sc.LogicalBytes),
		sc.RangeCount, sc.LeaseCount, sc.QueriesPerSecond, sc.WritesPerSecond,
		sc.L0Sublevels, sc.BytesPerReplica, sc.WritesPerReplica)
}

func (sc StoreCapacity) FractionUsed() float64 {
	__antithesis_instrumentation__.Notify(174029)
	if sc.Capacity == 0 {
		__antithesis_instrumentation__.Notify(174032)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(174033)
	}
	__antithesis_instrumentation__.Notify(174030)

	if sc.Used == 0 {
		__antithesis_instrumentation__.Notify(174034)
		return float64(sc.Capacity-sc.Available) / float64(sc.Capacity)
	} else {
		__antithesis_instrumentation__.Notify(174035)
	}
	__antithesis_instrumentation__.Notify(174031)
	return float64(sc.Used) / float64(sc.Available+sc.Used)
}

func (n *NodeDescriptor) AddressForLocality(loc Locality) *util.UnresolvedAddr {
	__antithesis_instrumentation__.Notify(174036)

	for i := range n.LocalityAddress {
		__antithesis_instrumentation__.Notify(174038)
		nLoc := &n.LocalityAddress[i]
		for _, loc := range loc.Tiers {
			__antithesis_instrumentation__.Notify(174039)
			if loc == nLoc.LocalityTier {
				__antithesis_instrumentation__.Notify(174040)
				return &nLoc.Address
			} else {
				__antithesis_instrumentation__.Notify(174041)
			}
		}
	}
	__antithesis_instrumentation__.Notify(174037)
	return &n.Address
}

func (n *NodeDescriptor) CheckedSQLAddress() *util.UnresolvedAddr {
	__antithesis_instrumentation__.Notify(174042)
	if n.SQLAddress.IsEmpty() {
		__antithesis_instrumentation__.Notify(174044)
		return &n.Address
	} else {
		__antithesis_instrumentation__.Notify(174045)
	}
	__antithesis_instrumentation__.Notify(174043)
	return &n.SQLAddress
}

func (t Tier) String() string {
	__antithesis_instrumentation__.Notify(174046)
	return fmt.Sprintf("%s=%s", t.Key, t.Value)
}

func (t *Tier) FromString(tier string) error {
	__antithesis_instrumentation__.Notify(174047)
	parts := strings.Split(tier, "=")
	if len(parts) != 2 || func() bool {
		__antithesis_instrumentation__.Notify(174049)
		return len(parts[0]) == 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(174050)
		return len(parts[1]) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(174051)
		return errors.Errorf("tier must be in the form \"key=value\" not %q", tier)
	} else {
		__antithesis_instrumentation__.Notify(174052)
	}
	__antithesis_instrumentation__.Notify(174048)
	t.Key = parts[0]
	t.Value = parts[1]
	return nil
}

func (l Locality) String() string {
	__antithesis_instrumentation__.Notify(174053)
	tiers := make([]string, len(l.Tiers))
	for i, tier := range l.Tiers {
		__antithesis_instrumentation__.Notify(174055)
		tiers[i] = tier.String()
	}
	__antithesis_instrumentation__.Notify(174054)
	return strings.Join(tiers, ",")
}

func (Locality) Type() string {
	__antithesis_instrumentation__.Notify(174056)
	return "Locality"
}

func (l Locality) Equals(r Locality) bool {
	__antithesis_instrumentation__.Notify(174057)
	if len(l.Tiers) != len(r.Tiers) {
		__antithesis_instrumentation__.Notify(174060)
		return false
	} else {
		__antithesis_instrumentation__.Notify(174061)
	}
	__antithesis_instrumentation__.Notify(174058)
	for i := range l.Tiers {
		__antithesis_instrumentation__.Notify(174062)
		if l.Tiers[i] != r.Tiers[i] {
			__antithesis_instrumentation__.Notify(174063)
			return false
		} else {
			__antithesis_instrumentation__.Notify(174064)
		}
	}
	__antithesis_instrumentation__.Notify(174059)
	return true
}

const MaxDiversityScore = 1.0

func (l Locality) DiversityScore(other Locality) float64 {
	__antithesis_instrumentation__.Notify(174065)
	length := len(l.Tiers)
	if len(other.Tiers) < length {
		__antithesis_instrumentation__.Notify(174069)
		length = len(other.Tiers)
	} else {
		__antithesis_instrumentation__.Notify(174070)
	}
	__antithesis_instrumentation__.Notify(174066)
	for i := 0; i < length; i++ {
		__antithesis_instrumentation__.Notify(174071)
		if l.Tiers[i].Value != other.Tiers[i].Value {
			__antithesis_instrumentation__.Notify(174072)
			return float64(length-i) / float64(length)
		} else {
			__antithesis_instrumentation__.Notify(174073)
		}
	}
	__antithesis_instrumentation__.Notify(174067)
	if len(l.Tiers) != len(other.Tiers) {
		__antithesis_instrumentation__.Notify(174074)
		return MaxDiversityScore / float64(length+1)
	} else {
		__antithesis_instrumentation__.Notify(174075)
	}
	__antithesis_instrumentation__.Notify(174068)
	return 0
}

func (l *Locality) Set(value string) error {
	__antithesis_instrumentation__.Notify(174076)
	if len(l.Tiers) > 0 {
		__antithesis_instrumentation__.Notify(174080)
		return errors.New("can't set locality more than once")
	} else {
		__antithesis_instrumentation__.Notify(174081)
	}
	__antithesis_instrumentation__.Notify(174077)
	if len(value) == 0 {
		__antithesis_instrumentation__.Notify(174082)
		return errors.New("can't have empty locality")
	} else {
		__antithesis_instrumentation__.Notify(174083)
	}
	__antithesis_instrumentation__.Notify(174078)

	tiersStr := strings.Split(value, ",")
	tiers := make([]Tier, len(tiersStr))
	for i, tier := range tiersStr {
		__antithesis_instrumentation__.Notify(174084)
		if err := tiers[i].FromString(tier); err != nil {
			__antithesis_instrumentation__.Notify(174085)
			return err
		} else {
			__antithesis_instrumentation__.Notify(174086)
		}
	}
	__antithesis_instrumentation__.Notify(174079)
	l.Tiers = tiers
	return nil
}

func (l *Locality) Find(key string) (value string, ok bool) {
	__antithesis_instrumentation__.Notify(174087)
	for i := range l.Tiers {
		__antithesis_instrumentation__.Notify(174089)
		if l.Tiers[i].Key == key {
			__antithesis_instrumentation__.Notify(174090)
			return l.Tiers[i].Value, true
		} else {
			__antithesis_instrumentation__.Notify(174091)
		}
	}
	__antithesis_instrumentation__.Notify(174088)
	return "", false
}

var DefaultLocationInformation = []struct {
	Locality  Locality
	Latitude  string
	Longitude string
}{
	{
		Locality:  Locality{Tiers: []Tier{{Key: "region", Value: "us-east1"}}},
		Latitude:  "33.836082",
		Longitude: "-81.163727",
	},
	{
		Locality:  Locality{Tiers: []Tier{{Key: "region", Value: "us-east4"}}},
		Latitude:  "37.478397",
		Longitude: "-76.453077",
	},
	{
		Locality:  Locality{Tiers: []Tier{{Key: "region", Value: "us-central1"}}},
		Latitude:  "42.032974",
		Longitude: "-93.581543",
	},
	{
		Locality:  Locality{Tiers: []Tier{{Key: "region", Value: "us-west1"}}},
		Latitude:  "43.804133",
		Longitude: "-120.554201",
	},
	{
		Locality:  Locality{Tiers: []Tier{{Key: "region", Value: "europe-west1"}}},
		Latitude:  "50.44816",
		Longitude: "3.81886",
	},
}

func (s StoreDescriptor) Locality() Locality {
	__antithesis_instrumentation__.Notify(174092)
	return s.Node.Locality.AddTier(
		Tier{Key: "node", Value: s.Node.NodeID.String()})
}

func (l Locality) AddTier(tier Tier) Locality {
	__antithesis_instrumentation__.Notify(174093)
	if len(l.Tiers) > 0 {
		__antithesis_instrumentation__.Notify(174095)
		tiers := make([]Tier, len(l.Tiers), len(l.Tiers)+1)
		copy(tiers, l.Tiers)
		tiers = append(tiers, tier)
		return Locality{Tiers: tiers}
	} else {
		__antithesis_instrumentation__.Notify(174096)
	}
	__antithesis_instrumentation__.Notify(174094)
	return Locality{Tiers: []Tier{tier}}
}
