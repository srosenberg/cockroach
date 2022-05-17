package loqrecovery

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type storeIDSet map[roachpb.StoreID]struct{}

func storeSliceFromSet(set storeIDSet) []roachpb.StoreID {
	__antithesis_instrumentation__.Notify(110076)
	storeIDs := make([]roachpb.StoreID, 0, len(set))
	for k := range set {
		__antithesis_instrumentation__.Notify(110079)
		storeIDs = append(storeIDs, k)
	}
	__antithesis_instrumentation__.Notify(110077)
	sort.Slice(storeIDs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(110080)
		return storeIDs[i] < storeIDs[j]
	})
	__antithesis_instrumentation__.Notify(110078)
	return storeIDs
}

func joinStoreIDs(storeIDs storeIDSet) string {
	__antithesis_instrumentation__.Notify(110081)
	storeNames := make([]string, 0, len(storeIDs))
	for _, id := range storeSliceFromSet(storeIDs) {
		__antithesis_instrumentation__.Notify(110083)
		storeNames = append(storeNames, fmt.Sprintf("s%d", id))
	}
	__antithesis_instrumentation__.Notify(110082)
	return strings.Join(storeNames, ", ")
}

func keyMax(key1 roachpb.RKey, key2 roachpb.RKey) roachpb.RKey {
	__antithesis_instrumentation__.Notify(110084)
	if key1.Less(key2) {
		__antithesis_instrumentation__.Notify(110086)
		return key2
	} else {
		__antithesis_instrumentation__.Notify(110087)
	}
	__antithesis_instrumentation__.Notify(110085)
	return key1
}

func keyMin(key1 roachpb.RKey, key2 roachpb.RKey) roachpb.RKey {
	__antithesis_instrumentation__.Notify(110088)
	if key2.Less(key1) {
		__antithesis_instrumentation__.Notify(110090)
		return key2
	} else {
		__antithesis_instrumentation__.Notify(110091)
	}
	__antithesis_instrumentation__.Notify(110089)
	return key1
}

type Problem interface {
	fmt.Stringer

	Span() roachpb.Span
}

type keyspaceGap struct {
	span roachpb.Span

	range1     roachpb.RangeID
	range1Span roachpb.Span

	range2     roachpb.RangeID
	range2Span roachpb.Span
}

func (i keyspaceGap) String() string {
	__antithesis_instrumentation__.Notify(110092)
	return fmt.Sprintf("range gap %v\n  r%d: %v\n  r%d: %v",
		i.span, i.range1, i.range1Span, i.range2, i.range2Span)
}

func (i keyspaceGap) Span() roachpb.Span {
	__antithesis_instrumentation__.Notify(110093)
	return i.span
}

type keyspaceOverlap struct {
	span roachpb.Span

	range1     roachpb.RangeID
	range1Span roachpb.Span

	range2     roachpb.RangeID
	range2Span roachpb.Span
}

func (i keyspaceOverlap) String() string {
	__antithesis_instrumentation__.Notify(110094)
	return fmt.Sprintf("range overlap %v\n  r%d: %v\n  r%d: %v",
		i.span, i.range1, i.range1Span, i.range2, i.range2Span)
}

func (i keyspaceOverlap) Span() roachpb.Span {
	__antithesis_instrumentation__.Notify(110095)
	return i.span
}

type rangeSplit struct {
	rangeID roachpb.RangeID
	span    roachpb.Span

	rHSRangeID   roachpb.RangeID
	rHSRangeSpan roachpb.Span
}

func (i rangeSplit) String() string {
	__antithesis_instrumentation__.Notify(110096)
	return fmt.Sprintf("range has unapplied split operation\n  r%d, %v rhs r%d, %v",
		i.rangeID, i.span, i.rHSRangeID, i.rHSRangeSpan)
}

func (i rangeSplit) Span() roachpb.Span {
	__antithesis_instrumentation__.Notify(110097)
	return i.span
}

type rangeMerge rangeSplit

func (i rangeMerge) String() string {
	__antithesis_instrumentation__.Notify(110098)
	return fmt.Sprintf("range has unapplied merge operation\n  r%d, %v with r%d, %v",
		i.rangeID, i.span, i.rHSRangeID, i.rHSRangeSpan)
}

func (i rangeMerge) Span() roachpb.Span {
	__antithesis_instrumentation__.Notify(110099)
	return i.span
}

type rangeReplicaRemoval struct {
	rangeID roachpb.RangeID
	span    roachpb.Span
}

func (i rangeReplicaRemoval) String() string {
	__antithesis_instrumentation__.Notify(110100)
	return fmt.Sprintf("range has unapplied descriptor change that removes current replica\n  r%d: %v",
		i.rangeID,
		i.span)
}

func (i rangeReplicaRemoval) Span() roachpb.Span {
	__antithesis_instrumentation__.Notify(110101)
	return i.span
}

type RecoveryError struct {
	problems []Problem
}

func (e *RecoveryError) Error() string {
	__antithesis_instrumentation__.Notify(110102)
	return "loss of quorum recovery error"
}

func (e *RecoveryError) ErrorDetail() string {
	__antithesis_instrumentation__.Notify(110103)
	descriptions := make([]string, 0, len(e.problems))
	for _, id := range e.problems {
		__antithesis_instrumentation__.Notify(110105)
		descriptions = append(descriptions, fmt.Sprintf("%v", id))
	}
	__antithesis_instrumentation__.Notify(110104)
	return strings.Join(descriptions, "\n")
}
