package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/constraint"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

const (
	epsilon = 1e-10

	allocatorRandomCount = 2

	maxFractionUsedThreshold = 0.95

	rebalanceToMaxFractionUsedThreshold = 0.925

	minRangeRebalanceThreshold = 2
)

var rangeRebalanceThreshold = func() *settings.FloatSetting {
	__antithesis_instrumentation__.Notify(94645)
	s := settings.RegisterFloatSetting(
		settings.SystemOnly,
		"kv.allocator.range_rebalance_threshold",
		"minimum fraction away from the mean a store's range count can be before it is considered overfull or underfull",
		0.05,
		settings.NonNegativeFloat,
	)
	s.SetVisibility(settings.Public)
	return s
}()

type scorerOptions interface {
	maybeJitterStoreStats(sl StoreList, allocRand allocatorRand) StoreList

	deterministicForTesting() bool

	shouldRebalanceBasedOnThresholds(
		ctx context.Context,
		eqClass equivalenceClass,
		metrics AllocatorMetrics,
	) bool

	balanceScore(sl StoreList, sc roachpb.StoreCapacity) balanceStatus

	rebalanceFromConvergesScore(eqClass equivalenceClass) int

	rebalanceToConvergesScore(eqClass equivalenceClass, candidate roachpb.StoreDescriptor) int

	removalMaximallyConvergesScore(removalCandStoreList StoreList, existing roachpb.StoreDescriptor) int
}

func jittered(val float64, jitter float64, rand allocatorRand) float64 {
	__antithesis_instrumentation__.Notify(94646)
	rand.Lock()
	defer rand.Unlock()
	result := val * jitter * (rand.Float64())
	if rand.Int31()%2 == 0 {
		__antithesis_instrumentation__.Notify(94648)
		result *= -1
	} else {
		__antithesis_instrumentation__.Notify(94649)
	}
	__antithesis_instrumentation__.Notify(94647)
	return result
}

type scatterScorerOptions struct {
	rangeCountScorerOptions

	jitter float64
}

func (o *scatterScorerOptions) maybeJitterStoreStats(
	sl StoreList, allocRand allocatorRand,
) (perturbedSL StoreList) {
	__antithesis_instrumentation__.Notify(94650)
	perturbedStoreDescs := make([]roachpb.StoreDescriptor, 0, len(sl.stores))
	for _, store := range sl.stores {
		__antithesis_instrumentation__.Notify(94652)
		store.Capacity.RangeCount += int32(jittered(
			float64(store.Capacity.RangeCount), o.jitter, allocRand,
		))
		perturbedStoreDescs = append(perturbedStoreDescs, store)
	}
	__antithesis_instrumentation__.Notify(94651)

	return makeStoreList(perturbedStoreDescs)
}

type rangeCountScorerOptions struct {
	deterministic           bool
	rangeRebalanceThreshold float64
}

func (o *rangeCountScorerOptions) maybeJitterStoreStats(
	sl StoreList, _ allocatorRand,
) (perturbedSL StoreList) {
	__antithesis_instrumentation__.Notify(94653)
	return sl
}

func (o *rangeCountScorerOptions) deterministicForTesting() bool {
	__antithesis_instrumentation__.Notify(94654)
	return o.deterministic
}

func (o rangeCountScorerOptions) shouldRebalanceBasedOnThresholds(
	ctx context.Context, eqClass equivalenceClass, metrics AllocatorMetrics,
) bool {
	__antithesis_instrumentation__.Notify(94655)
	store := eqClass.existing
	sl := eqClass.candidateSL
	if len(sl.stores) == 0 {
		__antithesis_instrumentation__.Notify(94659)
		return false
	} else {
		__antithesis_instrumentation__.Notify(94660)
	}
	__antithesis_instrumentation__.Notify(94656)

	overfullThreshold := int32(math.Ceil(overfullRangeThreshold(&o, sl.candidateRanges.mean)))

	if store.Capacity.RangeCount > overfullThreshold {
		__antithesis_instrumentation__.Notify(94661)
		log.VEventf(ctx, 2,
			"s%d: should-rebalance(ranges-overfull): rangeCount=%d, mean=%.2f, overfull-threshold=%d",
			store.StoreID, store.Capacity.RangeCount, sl.candidateRanges.mean, overfullThreshold)
		return true
	} else {
		__antithesis_instrumentation__.Notify(94662)
	}
	__antithesis_instrumentation__.Notify(94657)

	if float64(store.Capacity.RangeCount) > sl.candidateRanges.mean {
		__antithesis_instrumentation__.Notify(94663)
		underfullThreshold := int32(math.Floor(underfullRangeThreshold(&o, sl.candidateRanges.mean)))
		for _, desc := range sl.stores {
			__antithesis_instrumentation__.Notify(94664)
			if desc.Capacity.RangeCount < underfullThreshold {
				__antithesis_instrumentation__.Notify(94665)
				log.VEventf(ctx, 2,
					"s%d: should-rebalance(better-fit-ranges=s%d): rangeCount=%d, otherRangeCount=%d, "+
						"mean=%.2f, underfull-threshold=%d",
					store.StoreID, desc.StoreID, store.Capacity.RangeCount, desc.Capacity.RangeCount,
					sl.candidateRanges.mean, underfullThreshold)
				return true
			} else {
				__antithesis_instrumentation__.Notify(94666)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(94667)
	}
	__antithesis_instrumentation__.Notify(94658)

	return false
}

func (o *rangeCountScorerOptions) balanceScore(
	sl StoreList, sc roachpb.StoreCapacity,
) balanceStatus {
	__antithesis_instrumentation__.Notify(94668)
	maxRangeCount := overfullRangeThreshold(o, sl.candidateRanges.mean)
	minRangeCount := underfullRangeThreshold(o, sl.candidateRanges.mean)
	curRangeCount := float64(sc.RangeCount)
	if curRangeCount < minRangeCount {
		__antithesis_instrumentation__.Notify(94670)
		return underfull
	} else {
		__antithesis_instrumentation__.Notify(94671)
		if curRangeCount >= maxRangeCount {
			__antithesis_instrumentation__.Notify(94672)
			return overfull
		} else {
			__antithesis_instrumentation__.Notify(94673)
		}
	}
	__antithesis_instrumentation__.Notify(94669)
	return aroundTheMean
}

func (o *rangeCountScorerOptions) rebalanceFromConvergesScore(eqClass equivalenceClass) int {
	__antithesis_instrumentation__.Notify(94674)
	if !rebalanceConvergesRangeCountOnMean(
		eqClass.candidateSL, eqClass.existing.Capacity, eqClass.existing.Capacity.RangeCount-1,
	) {
		__antithesis_instrumentation__.Notify(94676)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(94677)
	}
	__antithesis_instrumentation__.Notify(94675)
	return 0
}

func (o *rangeCountScorerOptions) rebalanceToConvergesScore(
	eqClass equivalenceClass, candidate roachpb.StoreDescriptor,
) int {
	__antithesis_instrumentation__.Notify(94678)
	if rebalanceConvergesRangeCountOnMean(eqClass.candidateSL, candidate.Capacity, candidate.Capacity.RangeCount+1) {
		__antithesis_instrumentation__.Notify(94680)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(94681)
	}
	__antithesis_instrumentation__.Notify(94679)
	return 0
}

func (o *rangeCountScorerOptions) removalMaximallyConvergesScore(
	removalCandStoreList StoreList, existing roachpb.StoreDescriptor,
) int {
	__antithesis_instrumentation__.Notify(94682)
	if !rebalanceConvergesRangeCountOnMean(
		removalCandStoreList, existing.Capacity, existing.Capacity.RangeCount-1,
	) {
		__antithesis_instrumentation__.Notify(94684)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(94685)
	}
	__antithesis_instrumentation__.Notify(94683)
	return 0
}

type qpsScorerOptions struct {
	deterministic                             bool
	qpsRebalanceThreshold, minRequiredQPSDiff float64

	qpsPerReplica float64
}

func (o *qpsScorerOptions) maybeJitterStoreStats(sl StoreList, _ allocatorRand) StoreList {
	__antithesis_instrumentation__.Notify(94686)
	return sl
}

func (o *qpsScorerOptions) deterministicForTesting() bool {
	__antithesis_instrumentation__.Notify(94687)
	return o.deterministic
}

func (o qpsScorerOptions) shouldRebalanceBasedOnThresholds(
	ctx context.Context, eqClass equivalenceClass, metrics AllocatorMetrics,
) bool {
	__antithesis_instrumentation__.Notify(94688)
	if len(eqClass.candidateSL.stores) == 0 {
		__antithesis_instrumentation__.Notify(94691)
		return false
	} else {
		__antithesis_instrumentation__.Notify(94692)
	}
	__antithesis_instrumentation__.Notify(94689)

	bestStore, declineReason := o.getRebalanceTargetToMinimizeDelta(eqClass)
	switch declineReason {
	case noBetterCandidate:
		__antithesis_instrumentation__.Notify(94693)
		metrics.loadBasedReplicaRebalanceMetrics.CannotFindBetterCandidate.Inc(1)
		log.VEventf(
			ctx, 4, "could not find a better candidate to replace s%d", eqClass.existing.StoreID,
		)
	case existingNotOverfull:
		__antithesis_instrumentation__.Notify(94694)
		metrics.loadBasedReplicaRebalanceMetrics.ExistingNotOverfull.Inc(1)
		log.VEventf(ctx, 4, "existing store s%d is not overfull", eqClass.existing.StoreID)
	case deltaNotSignificant:
		__antithesis_instrumentation__.Notify(94695)
		metrics.loadBasedReplicaRebalanceMetrics.DeltaNotSignificant.Inc(1)
		log.VEventf(
			ctx, 4,
			"delta between s%d and the next best candidate is not significant enough",
			eqClass.existing.StoreID,
		)
	case significantlySwitchesRelativeDisposition:
		__antithesis_instrumentation__.Notify(94696)
		metrics.loadBasedReplicaRebalanceMetrics.SignificantlySwitchesRelativeDisposition.Inc(1)
		log.VEventf(
			ctx, 4,
			"rebalancing from s%[1]d to the next best candidate could make it significantly hotter than s%[1]d",
			eqClass.existing.StoreID,
		)
	case missingStatsForExistingStore:
		__antithesis_instrumentation__.Notify(94697)
		metrics.loadBasedReplicaRebalanceMetrics.MissingStatsForExistingStore.Inc(1)
		log.VEventf(ctx, 4, "missing QPS stats for s%d", eqClass.existing.StoreID)
	case shouldRebalance:
		__antithesis_instrumentation__.Notify(94698)
		metrics.loadBasedReplicaRebalanceMetrics.ShouldRebalance.Inc(1)
		var bestStoreQPS float64
		for _, store := range eqClass.candidateSL.stores {
			__antithesis_instrumentation__.Notify(94701)
			if bestStore == store.StoreID {
				__antithesis_instrumentation__.Notify(94702)
				bestStoreQPS = store.Capacity.QueriesPerSecond
			} else {
				__antithesis_instrumentation__.Notify(94703)
			}
		}
		__antithesis_instrumentation__.Notify(94699)
		log.VEventf(
			ctx, 4,
			"should rebalance replica with %0.2f qps from s%d (qps=%0.2f) to s%d (qps=%0.2f)",
			o.qpsPerReplica, eqClass.existing.StoreID,
			eqClass.existing.Capacity.QueriesPerSecond,
			bestStore, bestStoreQPS,
		)
	default:
		__antithesis_instrumentation__.Notify(94700)
		log.Fatalf(ctx, "unknown reason to decline rebalance: %v", declineReason)
	}
	__antithesis_instrumentation__.Notify(94690)

	return declineReason == shouldRebalance
}

func (o *qpsScorerOptions) balanceScore(sl StoreList, sc roachpb.StoreCapacity) balanceStatus {
	__antithesis_instrumentation__.Notify(94704)
	maxQPS := overfullQPSThreshold(o, sl.candidateQueriesPerSecond.mean)
	minQPS := underfullQPSThreshold(o, sl.candidateQueriesPerSecond.mean)
	curQPS := sc.QueriesPerSecond
	if curQPS < minQPS {
		__antithesis_instrumentation__.Notify(94706)
		return underfull
	} else {
		__antithesis_instrumentation__.Notify(94707)
		if curQPS >= maxQPS {
			__antithesis_instrumentation__.Notify(94708)
			return overfull
		} else {
			__antithesis_instrumentation__.Notify(94709)
		}
	}
	__antithesis_instrumentation__.Notify(94705)
	return aroundTheMean
}

func (o *qpsScorerOptions) rebalanceFromConvergesScore(eqClass equivalenceClass) int {
	__antithesis_instrumentation__.Notify(94710)
	_, declineReason := o.getRebalanceTargetToMinimizeDelta(eqClass)

	if declineReason == shouldRebalance {
		__antithesis_instrumentation__.Notify(94712)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(94713)
	}
	__antithesis_instrumentation__.Notify(94711)
	return 0
}

func (o *qpsScorerOptions) rebalanceToConvergesScore(
	eqClass equivalenceClass, candidate roachpb.StoreDescriptor,
) int {
	__antithesis_instrumentation__.Notify(94714)
	bestTarget, declineReason := o.getRebalanceTargetToMinimizeDelta(eqClass)
	if declineReason == shouldRebalance && func() bool {
		__antithesis_instrumentation__.Notify(94716)
		return bestTarget == candidate.StoreID == true
	}() == true {
		__antithesis_instrumentation__.Notify(94717)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(94718)
	}
	__antithesis_instrumentation__.Notify(94715)
	return 0
}

func (o *qpsScorerOptions) removalMaximallyConvergesScore(
	removalCandStoreList StoreList, existing roachpb.StoreDescriptor,
) int {
	__antithesis_instrumentation__.Notify(94719)
	maxQPS := float64(-1)
	for _, store := range removalCandStoreList.stores {
		__antithesis_instrumentation__.Notify(94722)
		if store.Capacity.QueriesPerSecond > maxQPS {
			__antithesis_instrumentation__.Notify(94723)
			maxQPS = store.Capacity.QueriesPerSecond
		} else {
			__antithesis_instrumentation__.Notify(94724)
		}
	}
	__antithesis_instrumentation__.Notify(94720)

	if scoresAlmostEqual(maxQPS, existing.Capacity.QueriesPerSecond) {
		__antithesis_instrumentation__.Notify(94725)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(94726)
	}
	__antithesis_instrumentation__.Notify(94721)
	return 0
}

type candidate struct {
	store          roachpb.StoreDescriptor
	valid          bool
	fullDisk       bool
	necessary      bool
	diversityScore float64
	convergesScore int
	balanceScore   balanceStatus
	rangeCount     int
	details        string
}

func (c candidate) String() string {
	__antithesis_instrumentation__.Notify(94727)
	str := fmt.Sprintf("s%d, valid:%t, fulldisk:%t, necessary:%t, diversity:%.2f, converges:%d, "+
		"balance:%d, rangeCount:%d, queriesPerSecond:%.2f",
		c.store.StoreID, c.valid, c.fullDisk, c.necessary, c.diversityScore, c.convergesScore,
		c.balanceScore, c.rangeCount, c.store.Capacity.QueriesPerSecond)
	if c.details != "" {
		__antithesis_instrumentation__.Notify(94729)
		return fmt.Sprintf("%s, details:(%s)", str, c.details)
	} else {
		__antithesis_instrumentation__.Notify(94730)
	}
	__antithesis_instrumentation__.Notify(94728)
	return str
}

func (c candidate) compactString() string {
	__antithesis_instrumentation__.Notify(94731)
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "s%d", c.store.StoreID)
	if !c.valid {
		__antithesis_instrumentation__.Notify(94737)
		fmt.Fprintf(&buf, ", valid:%t", c.valid)
	} else {
		__antithesis_instrumentation__.Notify(94738)
	}
	__antithesis_instrumentation__.Notify(94732)
	if c.fullDisk {
		__antithesis_instrumentation__.Notify(94739)
		fmt.Fprintf(&buf, ", fullDisk:%t", c.fullDisk)
	} else {
		__antithesis_instrumentation__.Notify(94740)
	}
	__antithesis_instrumentation__.Notify(94733)
	if c.necessary {
		__antithesis_instrumentation__.Notify(94741)
		fmt.Fprintf(&buf, ", necessary:%t", c.necessary)
	} else {
		__antithesis_instrumentation__.Notify(94742)
	}
	__antithesis_instrumentation__.Notify(94734)
	if c.diversityScore != 0 {
		__antithesis_instrumentation__.Notify(94743)
		fmt.Fprintf(&buf, ", diversity:%.2f", c.diversityScore)
	} else {
		__antithesis_instrumentation__.Notify(94744)
	}
	__antithesis_instrumentation__.Notify(94735)
	fmt.Fprintf(&buf, ", converges:%d, balance:%d, rangeCount:%d",
		c.convergesScore, c.balanceScore, c.rangeCount)
	if c.details != "" {
		__antithesis_instrumentation__.Notify(94745)
		fmt.Fprintf(&buf, ", details:(%s)", c.details)
	} else {
		__antithesis_instrumentation__.Notify(94746)
	}
	__antithesis_instrumentation__.Notify(94736)
	return buf.String()
}

func (c candidate) less(o candidate) bool {
	__antithesis_instrumentation__.Notify(94747)
	return c.compare(o) < 0
}

func (c candidate) compare(o candidate) float64 {
	__antithesis_instrumentation__.Notify(94748)
	if !o.valid {
		__antithesis_instrumentation__.Notify(94759)
		return 6
	} else {
		__antithesis_instrumentation__.Notify(94760)
	}
	__antithesis_instrumentation__.Notify(94749)
	if !c.valid {
		__antithesis_instrumentation__.Notify(94761)
		return -6
	} else {
		__antithesis_instrumentation__.Notify(94762)
	}
	__antithesis_instrumentation__.Notify(94750)
	if o.fullDisk {
		__antithesis_instrumentation__.Notify(94763)
		return 5
	} else {
		__antithesis_instrumentation__.Notify(94764)
	}
	__antithesis_instrumentation__.Notify(94751)
	if c.fullDisk {
		__antithesis_instrumentation__.Notify(94765)
		return -5
	} else {
		__antithesis_instrumentation__.Notify(94766)
	}
	__antithesis_instrumentation__.Notify(94752)
	if c.necessary != o.necessary {
		__antithesis_instrumentation__.Notify(94767)
		if c.necessary {
			__antithesis_instrumentation__.Notify(94769)
			return 4
		} else {
			__antithesis_instrumentation__.Notify(94770)
		}
		__antithesis_instrumentation__.Notify(94768)
		return -4
	} else {
		__antithesis_instrumentation__.Notify(94771)
	}
	__antithesis_instrumentation__.Notify(94753)
	if !scoresAlmostEqual(c.diversityScore, o.diversityScore) {
		__antithesis_instrumentation__.Notify(94772)
		if c.diversityScore > o.diversityScore {
			__antithesis_instrumentation__.Notify(94774)
			return 3
		} else {
			__antithesis_instrumentation__.Notify(94775)
		}
		__antithesis_instrumentation__.Notify(94773)
		return -3
	} else {
		__antithesis_instrumentation__.Notify(94776)
	}
	__antithesis_instrumentation__.Notify(94754)
	if c.convergesScore != o.convergesScore {
		__antithesis_instrumentation__.Notify(94777)
		if c.convergesScore > o.convergesScore {
			__antithesis_instrumentation__.Notify(94779)
			return 2 + float64(c.convergesScore-o.convergesScore)/10.0
		} else {
			__antithesis_instrumentation__.Notify(94780)
		}
		__antithesis_instrumentation__.Notify(94778)
		return -(2 + float64(o.convergesScore-c.convergesScore)/10.0)
	} else {
		__antithesis_instrumentation__.Notify(94781)
	}
	__antithesis_instrumentation__.Notify(94755)
	if c.balanceScore != o.balanceScore {
		__antithesis_instrumentation__.Notify(94782)
		if c.balanceScore > o.balanceScore {
			__antithesis_instrumentation__.Notify(94784)
			return 1 + (float64(c.balanceScore-o.balanceScore))/10.0
		} else {
			__antithesis_instrumentation__.Notify(94785)
		}
		__antithesis_instrumentation__.Notify(94783)
		return -(1 + (float64(o.balanceScore-c.balanceScore))/10.0)
	} else {
		__antithesis_instrumentation__.Notify(94786)
	}
	__antithesis_instrumentation__.Notify(94756)

	if c.rangeCount == 0 && func() bool {
		__antithesis_instrumentation__.Notify(94787)
		return o.rangeCount == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94788)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(94789)
	}
	__antithesis_instrumentation__.Notify(94757)
	if c.rangeCount < o.rangeCount {
		__antithesis_instrumentation__.Notify(94790)
		return float64(o.rangeCount-c.rangeCount) / float64(o.rangeCount)
	} else {
		__antithesis_instrumentation__.Notify(94791)
	}
	__antithesis_instrumentation__.Notify(94758)
	return -float64(c.rangeCount-o.rangeCount) / float64(c.rangeCount)
}

type candidateList []candidate

func (cl candidateList) String() string {
	__antithesis_instrumentation__.Notify(94792)
	if len(cl) == 0 {
		__antithesis_instrumentation__.Notify(94795)
		return "[]"
	} else {
		__antithesis_instrumentation__.Notify(94796)
	}
	__antithesis_instrumentation__.Notify(94793)
	var buffer bytes.Buffer
	buffer.WriteRune('[')
	for _, c := range cl {
		__antithesis_instrumentation__.Notify(94797)
		buffer.WriteRune('\n')
		buffer.WriteString(c.String())
	}
	__antithesis_instrumentation__.Notify(94794)
	buffer.WriteRune(']')
	return buffer.String()
}

type byScore candidateList

var _ sort.Interface = byScore(nil)

func (c byScore) Len() int { __antithesis_instrumentation__.Notify(94798); return len(c) }
func (c byScore) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(94799)
	return c[i].less(c[j])
}
func (c byScore) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(94800)
	c[i], c[j] = c[j], c[i]
}

type byScoreAndID candidateList

var _ sort.Interface = byScoreAndID(nil)

func (c byScoreAndID) Len() int { __antithesis_instrumentation__.Notify(94801); return len(c) }
func (c byScoreAndID) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(94802)
	if scoresAlmostEqual(c[i].diversityScore, c[j].diversityScore) && func() bool {
		__antithesis_instrumentation__.Notify(94804)
		return c[i].convergesScore == c[j].convergesScore == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(94805)
		return c[i].balanceScore == c[j].balanceScore == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(94806)
		return c[i].rangeCount == c[j].rangeCount == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(94807)
		return c[i].necessary == c[j].necessary == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(94808)
		return c[i].fullDisk == c[j].fullDisk == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(94809)
		return c[i].valid == c[j].valid == true
	}() == true {
		__antithesis_instrumentation__.Notify(94810)
		return c[i].store.StoreID < c[j].store.StoreID
	} else {
		__antithesis_instrumentation__.Notify(94811)
	}
	__antithesis_instrumentation__.Notify(94803)
	return c[i].less(c[j])
}
func (c byScoreAndID) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(94812)
	c[i], c[j] = c[j], c[i]
}

func (cl candidateList) onlyValidAndNotFull() candidateList {
	__antithesis_instrumentation__.Notify(94813)
	for i := len(cl) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(94815)
		if cl[i].valid && func() bool {
			__antithesis_instrumentation__.Notify(94816)
			return !cl[i].fullDisk == true
		}() == true {
			__antithesis_instrumentation__.Notify(94817)
			return cl[:i+1]
		} else {
			__antithesis_instrumentation__.Notify(94818)
		}
	}
	__antithesis_instrumentation__.Notify(94814)
	return candidateList{}
}

func (cl candidateList) best() candidateList {
	__antithesis_instrumentation__.Notify(94819)
	cl = cl.onlyValidAndNotFull()
	if len(cl) <= 1 {
		__antithesis_instrumentation__.Notify(94822)
		return cl
	} else {
		__antithesis_instrumentation__.Notify(94823)
	}
	__antithesis_instrumentation__.Notify(94820)
	for i := 1; i < len(cl); i++ {
		__antithesis_instrumentation__.Notify(94824)
		if cl[i].necessary == cl[0].necessary && func() bool {
			__antithesis_instrumentation__.Notify(94826)
			return scoresAlmostEqual(cl[i].diversityScore, cl[0].diversityScore) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(94827)
			return cl[i].convergesScore == cl[0].convergesScore == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(94828)
			return cl[i].balanceScore == cl[0].balanceScore == true
		}() == true {
			__antithesis_instrumentation__.Notify(94829)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94830)
		}
		__antithesis_instrumentation__.Notify(94825)
		return cl[:i]
	}
	__antithesis_instrumentation__.Notify(94821)
	return cl
}

func (cl candidateList) worst() candidateList {
	__antithesis_instrumentation__.Notify(94831)
	if len(cl) <= 1 {
		__antithesis_instrumentation__.Notify(94836)
		return cl
	} else {
		__antithesis_instrumentation__.Notify(94837)
	}
	__antithesis_instrumentation__.Notify(94832)

	if !cl[len(cl)-1].valid {
		__antithesis_instrumentation__.Notify(94838)
		for i := len(cl) - 2; i >= 0; i-- {
			__antithesis_instrumentation__.Notify(94839)
			if cl[i].valid {
				__antithesis_instrumentation__.Notify(94840)
				return cl[i+1:]
			} else {
				__antithesis_instrumentation__.Notify(94841)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(94842)
	}
	__antithesis_instrumentation__.Notify(94833)

	if cl[len(cl)-1].fullDisk {
		__antithesis_instrumentation__.Notify(94843)
		for i := len(cl) - 2; i >= 0; i-- {
			__antithesis_instrumentation__.Notify(94844)
			if !cl[i].fullDisk {
				__antithesis_instrumentation__.Notify(94845)
				return cl[i+1:]
			} else {
				__antithesis_instrumentation__.Notify(94846)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(94847)
	}
	__antithesis_instrumentation__.Notify(94834)

	for i := len(cl) - 2; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(94848)
		if cl[i].necessary == cl[len(cl)-1].necessary && func() bool {
			__antithesis_instrumentation__.Notify(94850)
			return scoresAlmostEqual(cl[i].diversityScore, cl[len(cl)-1].diversityScore) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(94851)
			return cl[i].convergesScore == cl[len(cl)-1].convergesScore == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(94852)
			return cl[i].balanceScore == cl[len(cl)-1].balanceScore == true
		}() == true {
			__antithesis_instrumentation__.Notify(94853)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94854)
		}
		__antithesis_instrumentation__.Notify(94849)
		return cl[i+1:]
	}
	__antithesis_instrumentation__.Notify(94835)
	return cl
}

func (cl candidateList) betterThan(c candidate) candidateList {
	__antithesis_instrumentation__.Notify(94855)
	for i := 0; i < len(cl); i++ {
		__antithesis_instrumentation__.Notify(94857)
		if !c.less(cl[i]) {
			__antithesis_instrumentation__.Notify(94858)
			return cl[:i]
		} else {
			__antithesis_instrumentation__.Notify(94859)
		}
	}
	__antithesis_instrumentation__.Notify(94856)
	return cl
}

func (cl candidateList) selectGood(randGen allocatorRand) *candidate {
	__antithesis_instrumentation__.Notify(94860)
	cl = cl.best()
	if len(cl) == 0 {
		__antithesis_instrumentation__.Notify(94864)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(94865)
	}
	__antithesis_instrumentation__.Notify(94861)
	if len(cl) == 1 {
		__antithesis_instrumentation__.Notify(94866)
		return &cl[0]
	} else {
		__antithesis_instrumentation__.Notify(94867)
	}
	__antithesis_instrumentation__.Notify(94862)
	randGen.Lock()
	order := randGen.Perm(len(cl))
	randGen.Unlock()
	best := &cl[order[0]]
	for i := 1; i < allocatorRandomCount; i++ {
		__antithesis_instrumentation__.Notify(94868)
		if best.less(cl[order[i]]) {
			__antithesis_instrumentation__.Notify(94869)
			best = &cl[order[i]]
		} else {
			__antithesis_instrumentation__.Notify(94870)
		}
	}
	__antithesis_instrumentation__.Notify(94863)
	return best
}

func (cl candidateList) selectBad(randGen allocatorRand) *candidate {
	__antithesis_instrumentation__.Notify(94871)
	cl = cl.worst()
	if len(cl) == 0 {
		__antithesis_instrumentation__.Notify(94875)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(94876)
	}
	__antithesis_instrumentation__.Notify(94872)
	if len(cl) == 1 {
		__antithesis_instrumentation__.Notify(94877)
		return &cl[0]
	} else {
		__antithesis_instrumentation__.Notify(94878)
	}
	__antithesis_instrumentation__.Notify(94873)
	randGen.Lock()
	order := randGen.Perm(len(cl))
	randGen.Unlock()
	worst := &cl[order[0]]
	for i := 1; i < allocatorRandomCount; i++ {
		__antithesis_instrumentation__.Notify(94879)
		if cl[order[i]].less(*worst) {
			__antithesis_instrumentation__.Notify(94880)
			worst = &cl[order[i]]
		} else {
			__antithesis_instrumentation__.Notify(94881)
		}
	}
	__antithesis_instrumentation__.Notify(94874)
	return worst
}

func (cl candidateList) removeCandidate(c candidate) candidateList {
	__antithesis_instrumentation__.Notify(94882)
	for i := 0; i < len(cl); i++ {
		__antithesis_instrumentation__.Notify(94884)
		if cl[i].store.StoreID == c.store.StoreID {
			__antithesis_instrumentation__.Notify(94885)
			cl = append(cl[:i], cl[i+1:]...)
			break
		} else {
			__antithesis_instrumentation__.Notify(94886)
		}
	}
	__antithesis_instrumentation__.Notify(94883)
	return cl
}

func rankedCandidateListForAllocation(
	ctx context.Context,
	candidateStores StoreList,
	constraintsCheck constraintsCheckFn,
	existingReplicas []roachpb.ReplicaDescriptor,
	existingStoreLocalities map[roachpb.StoreID]roachpb.Locality,
	isStoreValidForRoutineReplicaTransfer func(context.Context, roachpb.StoreID) bool,
	allowMultipleReplsPerNode bool,
	options *rangeCountScorerOptions,
) candidateList {
	__antithesis_instrumentation__.Notify(94887)
	var candidates candidateList
	existingReplTargets := roachpb.MakeReplicaSet(existingReplicas).ReplicationTargets()
	for _, s := range candidateStores.stores {
		__antithesis_instrumentation__.Notify(94890)

		if storeHasReplica(s.StoreID, existingReplTargets) {
			__antithesis_instrumentation__.Notify(94896)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94897)
		}
		__antithesis_instrumentation__.Notify(94891)

		if !allowMultipleReplsPerNode && func() bool {
			__antithesis_instrumentation__.Notify(94898)
			return nodeHasReplica(s.Node.NodeID, existingReplTargets) == true
		}() == true {
			__antithesis_instrumentation__.Notify(94899)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94900)
		}
		__antithesis_instrumentation__.Notify(94892)
		if !isStoreValidForRoutineReplicaTransfer(ctx, s.StoreID) {
			__antithesis_instrumentation__.Notify(94901)
			log.VEventf(
				ctx,
				3,
				"not considering store s%d as a potential rebalance candidate because it is on a non-live node n%d",
				s.StoreID,
				s.Node.NodeID,
			)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94902)
		}
		__antithesis_instrumentation__.Notify(94893)
		constraintsOK, necessary := constraintsCheck(s)
		if !constraintsOK {
			__antithesis_instrumentation__.Notify(94903)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94904)
		}
		__antithesis_instrumentation__.Notify(94894)

		if !maxCapacityCheck(s) {
			__antithesis_instrumentation__.Notify(94905)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94906)
		}
		__antithesis_instrumentation__.Notify(94895)
		diversityScore := diversityAllocateScore(s, existingStoreLocalities)
		balanceScore := options.balanceScore(candidateStores, s.Capacity)
		candidates = append(candidates, candidate{
			store:          s,
			valid:          constraintsOK,
			necessary:      necessary,
			diversityScore: diversityScore,
			balanceScore:   balanceScore,
			rangeCount:     int(s.Capacity.RangeCount),
		})
	}
	__antithesis_instrumentation__.Notify(94888)
	if options.deterministicForTesting() {
		__antithesis_instrumentation__.Notify(94907)
		sort.Sort(sort.Reverse(byScoreAndID(candidates)))
	} else {
		__antithesis_instrumentation__.Notify(94908)
		sort.Sort(sort.Reverse(byScore(candidates)))
	}
	__antithesis_instrumentation__.Notify(94889)
	return candidates
}

func candidateListForRemoval(
	existingReplsStoreList StoreList,
	constraintsCheck constraintsCheckFn,
	existingStoreLocalities map[roachpb.StoreID]roachpb.Locality,
	options scorerOptions,
) candidateList {
	__antithesis_instrumentation__.Notify(94909)
	var candidates candidateList
	for _, s := range existingReplsStoreList.stores {
		__antithesis_instrumentation__.Notify(94915)
		constraintsOK, necessary := constraintsCheck(s)
		if !constraintsOK {
			__antithesis_instrumentation__.Notify(94917)
			candidates = append(candidates, candidate{
				store:     s,
				valid:     false,
				necessary: necessary,
				details:   "constraint check fail",
			})
			continue
		} else {
			__antithesis_instrumentation__.Notify(94918)
		}
		__antithesis_instrumentation__.Notify(94916)
		diversityScore := diversityRemovalScore(s.StoreID, existingStoreLocalities)
		candidates = append(candidates, candidate{
			store:          s,
			valid:          constraintsOK,
			necessary:      necessary,
			fullDisk:       !maxCapacityCheck(s),
			diversityScore: diversityScore,
		})
	}
	__antithesis_instrumentation__.Notify(94910)
	if options.deterministicForTesting() {
		__antithesis_instrumentation__.Notify(94919)
		sort.Sort(sort.Reverse(byScoreAndID(candidates)))
	} else {
		__antithesis_instrumentation__.Notify(94920)
		sort.Sort(sort.Reverse(byScore(candidates)))
	}
	__antithesis_instrumentation__.Notify(94911)

	candidates = candidates.worst()
	removalCandidateStores := make([]roachpb.StoreDescriptor, 0, len(candidates))
	for _, cand := range candidates {
		__antithesis_instrumentation__.Notify(94921)
		removalCandidateStores = append(removalCandidateStores, cand.store)
	}
	__antithesis_instrumentation__.Notify(94912)
	removalCandidateStoreList := makeStoreList(removalCandidateStores)
	for i := range candidates {
		__antithesis_instrumentation__.Notify(94922)

		candidates[i].convergesScore = options.removalMaximallyConvergesScore(
			removalCandidateStoreList, candidates[i].store,
		)
		candidates[i].balanceScore = options.balanceScore(
			removalCandidateStoreList, candidates[i].store.Capacity,
		)
		candidates[i].rangeCount = int(candidates[i].store.Capacity.RangeCount)
	}
	__antithesis_instrumentation__.Notify(94913)

	if options.deterministicForTesting() {
		__antithesis_instrumentation__.Notify(94923)
		sort.Sort(sort.Reverse(byScoreAndID(candidates)))
	} else {
		__antithesis_instrumentation__.Notify(94924)
		sort.Sort(sort.Reverse(byScore(candidates)))
	}
	__antithesis_instrumentation__.Notify(94914)
	return candidates
}

type rebalanceOptions struct {
	existing   candidate
	candidates candidateList
}

type equivalenceClass struct {
	existing roachpb.StoreDescriptor

	candidateSL StoreList
	candidates  candidateList
}

const (
	maxQPSTransferOvershoot = 500
)

type declineReason int

const (
	shouldRebalance declineReason = iota

	noBetterCandidate

	existingNotOverfull

	deltaNotSignificant

	significantlySwitchesRelativeDisposition

	missingStatsForExistingStore
)

func bestStoreToMinimizeQPSDelta(
	replQPS, rebalanceThreshold, minRequiredQPSDiff float64,
	existing roachpb.StoreID,
	candidates []roachpb.StoreID,
	storeDescMap map[roachpb.StoreID]*roachpb.StoreDescriptor,
) (bestCandidate roachpb.StoreID, reason declineReason) {
	__antithesis_instrumentation__.Notify(94925)
	storeQPSMap := make(map[roachpb.StoreID]float64, len(candidates)+1)
	for _, store := range candidates {
		__antithesis_instrumentation__.Notify(94935)
		if desc, ok := storeDescMap[store]; ok {
			__antithesis_instrumentation__.Notify(94936)
			storeQPSMap[store] = desc.Capacity.QueriesPerSecond
		} else {
			__antithesis_instrumentation__.Notify(94937)
		}
	}
	__antithesis_instrumentation__.Notify(94926)
	desc, ok := storeDescMap[existing]
	if !ok {
		__antithesis_instrumentation__.Notify(94938)
		return 0, missingStatsForExistingStore
	} else {
		__antithesis_instrumentation__.Notify(94939)
	}
	__antithesis_instrumentation__.Notify(94927)
	storeQPSMap[existing] = desc.Capacity.QueriesPerSecond

	domain := append(candidates, existing)
	storeDescs := make([]roachpb.StoreDescriptor, 0, len(domain))
	for _, desc := range storeDescMap {
		__antithesis_instrumentation__.Notify(94940)
		storeDescs = append(storeDescs, *desc)
	}
	__antithesis_instrumentation__.Notify(94928)
	domainStoreList := makeStoreList(storeDescs)

	bestCandidate = getCandidateWithMinQPS(storeQPSMap, candidates)
	if bestCandidate == 0 {
		__antithesis_instrumentation__.Notify(94941)
		return 0, noBetterCandidate
	} else {
		__antithesis_instrumentation__.Notify(94942)
	}
	__antithesis_instrumentation__.Notify(94929)

	bestCandQPS := storeQPSMap[bestCandidate]
	existingQPS := storeQPSMap[existing]
	if bestCandQPS > existingQPS {
		__antithesis_instrumentation__.Notify(94943)
		return 0, noBetterCandidate
	} else {
		__antithesis_instrumentation__.Notify(94944)
	}
	__antithesis_instrumentation__.Notify(94930)

	existingQPSIgnoringRepl := math.Max(existingQPS-replQPS, 0)

	diffIgnoringRepl := existingQPSIgnoringRepl - bestCandQPS
	if diffIgnoringRepl < minRequiredQPSDiff {
		__antithesis_instrumentation__.Notify(94945)
		return 0, deltaNotSignificant
	} else {
		__antithesis_instrumentation__.Notify(94946)
	}
	__antithesis_instrumentation__.Notify(94931)

	mean := domainStoreList.candidateQueriesPerSecond.mean
	overfullThreshold := overfullQPSThreshold(
		&qpsScorerOptions{qpsRebalanceThreshold: rebalanceThreshold},
		mean,
	)
	if existingQPS < overfullThreshold {
		__antithesis_instrumentation__.Notify(94947)
		return 0, existingNotOverfull
	} else {
		__antithesis_instrumentation__.Notify(94948)
	}
	__antithesis_instrumentation__.Notify(94932)

	currentQPSDelta := getQPSDelta(storeQPSMap, domain)

	storeQPSMap[bestCandidate] += replQPS

	storeQPSMap[existing] = existingQPSIgnoringRepl
	bestCandQPSWithRepl := storeQPSMap[bestCandidate]

	if existingQPSIgnoringRepl+maxQPSTransferOvershoot < bestCandQPSWithRepl {
		__antithesis_instrumentation__.Notify(94949)
		return 0, significantlySwitchesRelativeDisposition
	} else {
		__antithesis_instrumentation__.Notify(94950)
	}
	__antithesis_instrumentation__.Notify(94933)

	newQPSDelta := getQPSDelta(storeQPSMap, domain)
	if currentQPSDelta < newQPSDelta {
		__antithesis_instrumentation__.Notify(94951)
		panic(
			fmt.Sprintf(
				"programming error: projected QPS delta higher than current delta;"+
					" existing: %0.2f qps, coldest candidate: %0.2f qps, replica/lease: %0.2f qps",
				existingQPS, bestCandQPS, replQPS,
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(94952)
	}
	__antithesis_instrumentation__.Notify(94934)

	return bestCandidate, shouldRebalance
}

func (o *qpsScorerOptions) getRebalanceTargetToMinimizeDelta(
	eqClass equivalenceClass,
) (bestStore roachpb.StoreID, declineReason declineReason) {
	__antithesis_instrumentation__.Notify(94953)
	domainStoreList := makeStoreList(append(eqClass.candidateSL.stores, eqClass.existing))
	candidates := make([]roachpb.StoreID, 0, len(eqClass.candidateSL.stores))
	for _, store := range eqClass.candidateSL.stores {
		__antithesis_instrumentation__.Notify(94955)
		candidates = append(candidates, store.StoreID)
	}
	__antithesis_instrumentation__.Notify(94954)
	return bestStoreToMinimizeQPSDelta(
		o.qpsPerReplica,
		o.qpsRebalanceThreshold,
		o.minRequiredQPSDiff,
		eqClass.existing.StoreID,
		candidates,
		storeListToMap(domainStoreList),
	)
}

func rankedCandidateListForRebalancing(
	ctx context.Context,
	allStores StoreList,
	removalConstraintsChecker constraintsCheckFn,
	rebalanceConstraintsChecker rebalanceConstraintsCheckFn,
	existingReplicasForType, replicasOnExemptedStores []roachpb.ReplicaDescriptor,
	existingStoreLocalities map[roachpb.StoreID]roachpb.Locality,
	isStoreValidForRoutineReplicaTransfer func(context.Context, roachpb.StoreID) bool,
	options scorerOptions,
	metrics AllocatorMetrics,
) []rebalanceOptions {
	__antithesis_instrumentation__.Notify(94956)

	existingStores := make(map[roachpb.StoreID]candidate)
	var needRebalanceFrom bool
	curDiversityScore := rangeDiversityScore(existingStoreLocalities)
	for _, store := range allStores.stores {
		__antithesis_instrumentation__.Notify(94962)
		for _, repl := range existingReplicasForType {
			__antithesis_instrumentation__.Notify(94963)
			if store.StoreID != repl.StoreID {
				__antithesis_instrumentation__.Notify(94967)
				continue
			} else {
				__antithesis_instrumentation__.Notify(94968)
			}
			__antithesis_instrumentation__.Notify(94964)
			valid, necessary := removalConstraintsChecker(store)
			fullDisk := !maxCapacityCheck(store)
			if !valid {
				__antithesis_instrumentation__.Notify(94969)
				if !needRebalanceFrom {
					__antithesis_instrumentation__.Notify(94971)
					log.VEventf(ctx, 2, "s%d: should-rebalance(invalid): locality:%q",
						store.StoreID, store.Locality())
				} else {
					__antithesis_instrumentation__.Notify(94972)
				}
				__antithesis_instrumentation__.Notify(94970)
				needRebalanceFrom = true
			} else {
				__antithesis_instrumentation__.Notify(94973)
			}
			__antithesis_instrumentation__.Notify(94965)
			if fullDisk {
				__antithesis_instrumentation__.Notify(94974)
				if !needRebalanceFrom {
					__antithesis_instrumentation__.Notify(94976)
					log.VEventf(ctx, 2, "s%d: should-rebalance(full-disk): capacity:%q",
						store.StoreID, store.Capacity)
				} else {
					__antithesis_instrumentation__.Notify(94977)
				}
				__antithesis_instrumentation__.Notify(94975)
				needRebalanceFrom = true
			} else {
				__antithesis_instrumentation__.Notify(94978)
			}
			__antithesis_instrumentation__.Notify(94966)
			existingStores[store.StoreID] = candidate{
				store:          store,
				valid:          valid,
				necessary:      necessary,
				fullDisk:       fullDisk,
				diversityScore: curDiversityScore,
			}
		}
	}
	__antithesis_instrumentation__.Notify(94957)

	var equivalenceClasses []equivalenceClass
	var needRebalanceTo bool
	for _, existing := range existingStores {
		__antithesis_instrumentation__.Notify(94979)
		var comparableCands candidateList
		for _, store := range allStores.stores {
			__antithesis_instrumentation__.Notify(94983)

			if store.StoreID == existing.store.StoreID {
				__antithesis_instrumentation__.Notify(94988)
				continue
			} else {
				__antithesis_instrumentation__.Notify(94989)
			}
			__antithesis_instrumentation__.Notify(94984)

			if !isStoreValidForRoutineReplicaTransfer(ctx, store.StoreID) {
				__antithesis_instrumentation__.Notify(94990)
				log.VEventf(
					ctx,
					3,
					"not considering store s%d as a potential rebalance candidate because it is on a non-live node n%d",
					store.StoreID,
					store.Node.NodeID,
				)
				continue
			} else {
				__antithesis_instrumentation__.Notify(94991)
			}
			__antithesis_instrumentation__.Notify(94985)

			var exempted bool
			for _, replOnExemptedStore := range replicasOnExemptedStores {
				__antithesis_instrumentation__.Notify(94992)
				if store.StoreID == replOnExemptedStore.StoreID {
					__antithesis_instrumentation__.Notify(94993)
					log.VEventf(
						ctx,
						6,
						"s%d is not a possible rebalance candidate for non-voters because it already has a voter of the range; ignoring",
						store.StoreID,
					)
					exempted = true
					break
				} else {
					__antithesis_instrumentation__.Notify(94994)
				}
			}
			__antithesis_instrumentation__.Notify(94986)
			if exempted {
				__antithesis_instrumentation__.Notify(94995)
				continue
			} else {
				__antithesis_instrumentation__.Notify(94996)
			}
			__antithesis_instrumentation__.Notify(94987)

			constraintsOK, necessary := rebalanceConstraintsChecker(store, existing.store)
			maxCapacityOK := maxCapacityCheck(store)
			diversityScore := diversityRebalanceFromScore(
				store, existing.store.StoreID, existingStoreLocalities)
			cand := candidate{
				store:          store,
				valid:          constraintsOK,
				necessary:      necessary,
				fullDisk:       !maxCapacityOK,
				diversityScore: diversityScore,
			}
			if !cand.less(existing) {
				__antithesis_instrumentation__.Notify(94997)

				comparableCands = append(comparableCands, cand)

				if !needRebalanceFrom && func() bool {
					__antithesis_instrumentation__.Notify(94998)
					return !needRebalanceTo == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(94999)
					return existing.less(cand) == true
				}() == true {
					__antithesis_instrumentation__.Notify(95000)
					needRebalanceTo = true
					log.VEventf(ctx, 2,
						"s%d: should-rebalance(necessary/diversity=s%d): oldNecessary:%t, newNecessary:%t, "+
							"oldDiversity:%f, newDiversity:%f, locality:%q",
						existing.store.StoreID, store.StoreID, existing.necessary, cand.necessary,
						existing.diversityScore, cand.diversityScore, store.Locality())
				} else {
					__antithesis_instrumentation__.Notify(95001)
				}
			} else {
				__antithesis_instrumentation__.Notify(95002)
			}
		}
		__antithesis_instrumentation__.Notify(94980)
		if options.deterministicForTesting() {
			__antithesis_instrumentation__.Notify(95003)
			sort.Sort(sort.Reverse(byScoreAndID(comparableCands)))
		} else {
			__antithesis_instrumentation__.Notify(95004)
			sort.Sort(sort.Reverse(byScore(comparableCands)))
		}
		__antithesis_instrumentation__.Notify(94981)

		bestCands := comparableCands.best()

		bestStores := make([]roachpb.StoreDescriptor, len(bestCands))
		for i := range bestCands {
			__antithesis_instrumentation__.Notify(95005)
			bestStores[i] = bestCands[i].store
		}
		__antithesis_instrumentation__.Notify(94982)
		eqClass := equivalenceClass{
			existing:    existing.store,
			candidateSL: makeStoreList(bestStores),
			candidates:  bestCands,
		}
		equivalenceClasses = append(equivalenceClasses, eqClass)
	}
	__antithesis_instrumentation__.Notify(94958)

	needRebalance := needRebalanceFrom || func() bool {
		__antithesis_instrumentation__.Notify(95006)
		return needRebalanceTo == true
	}() == true
	var shouldRebalanceCheck bool
	if !needRebalance {
		__antithesis_instrumentation__.Notify(95007)
		for _, eqClass := range equivalenceClasses {
			__antithesis_instrumentation__.Notify(95008)
			if options.shouldRebalanceBasedOnThresholds(
				ctx,
				eqClass,
				metrics,
			) {
				__antithesis_instrumentation__.Notify(95009)
				shouldRebalanceCheck = true
				break
			} else {
				__antithesis_instrumentation__.Notify(95010)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(95011)
	}
	__antithesis_instrumentation__.Notify(94959)

	if !needRebalance && func() bool {
		__antithesis_instrumentation__.Notify(95012)
		return !shouldRebalanceCheck == true
	}() == true {
		__antithesis_instrumentation__.Notify(95013)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(95014)
	}
	__antithesis_instrumentation__.Notify(94960)

	results := make([]rebalanceOptions, 0, len(equivalenceClasses))
	for _, comparable := range equivalenceClasses {
		__antithesis_instrumentation__.Notify(95015)
		existing, ok := existingStores[comparable.existing.StoreID]
		if !ok {
			__antithesis_instrumentation__.Notify(95022)
			log.Errorf(ctx, "BUG: missing candidate struct for existing store %+v; stores: %+v",
				comparable.existing, existingStores)
			continue
		} else {
			__antithesis_instrumentation__.Notify(95023)
		}
		__antithesis_instrumentation__.Notify(95016)
		if !existing.valid {
			__antithesis_instrumentation__.Notify(95024)
			existing.details = "constraint check fail"
		} else {
			__antithesis_instrumentation__.Notify(95025)

			convergesScore := options.rebalanceFromConvergesScore(comparable)
			balanceScore := options.balanceScore(comparable.candidateSL, existing.store.Capacity)
			existing.convergesScore = convergesScore
			existing.balanceScore = balanceScore
			existing.rangeCount = int(existing.store.Capacity.RangeCount)
		}
		__antithesis_instrumentation__.Notify(95017)

		var candidates candidateList
		for _, cand := range comparable.candidates {
			__antithesis_instrumentation__.Notify(95026)

			if _, ok := existingStores[cand.store.StoreID]; ok {
				__antithesis_instrumentation__.Notify(95028)
				continue
			} else {
				__antithesis_instrumentation__.Notify(95029)
			}
			__antithesis_instrumentation__.Notify(95027)

			s := cand.store
			cand.fullDisk = !rebalanceToMaxCapacityCheck(s)
			cand.balanceScore = options.balanceScore(comparable.candidateSL, s.Capacity)
			cand.convergesScore = options.rebalanceToConvergesScore(comparable, s)
			cand.rangeCount = int(s.Capacity.RangeCount)
			candidates = append(candidates, cand)
		}
		__antithesis_instrumentation__.Notify(95018)

		if len(candidates) == 0 {
			__antithesis_instrumentation__.Notify(95030)
			continue
		} else {
			__antithesis_instrumentation__.Notify(95031)
		}
		__antithesis_instrumentation__.Notify(95019)

		if options.deterministicForTesting() {
			__antithesis_instrumentation__.Notify(95032)
			sort.Sort(sort.Reverse(byScoreAndID(candidates)))
		} else {
			__antithesis_instrumentation__.Notify(95033)
			sort.Sort(sort.Reverse(byScore(candidates)))
		}
		__antithesis_instrumentation__.Notify(95020)

		improvementCandidates := candidates.betterThan(existing)
		if len(improvementCandidates) == 0 {
			__antithesis_instrumentation__.Notify(95034)
			continue
		} else {
			__antithesis_instrumentation__.Notify(95035)
		}
		__antithesis_instrumentation__.Notify(95021)
		results = append(results, rebalanceOptions{
			existing:   existing,
			candidates: improvementCandidates,
		})
		log.VEventf(ctx, 5, "rebalance candidates #%d: %s\nexisting replicas: %s",
			len(results), results[len(results)-1].candidates, results[len(results)-1].existing)
	}
	__antithesis_instrumentation__.Notify(94961)

	return results
}

func bestRebalanceTarget(
	randGen allocatorRand, options []rebalanceOptions,
) (target, existingCandidate *candidate) {
	__antithesis_instrumentation__.Notify(95036)
	bestIdx := -1
	var bestTarget *candidate
	var replaces candidate
	for i, option := range options {
		__antithesis_instrumentation__.Notify(95039)
		if len(option.candidates) == 0 {
			__antithesis_instrumentation__.Notify(95042)
			continue
		} else {
			__antithesis_instrumentation__.Notify(95043)
		}
		__antithesis_instrumentation__.Notify(95040)
		target := option.candidates.selectGood(randGen)
		if target == nil {
			__antithesis_instrumentation__.Notify(95044)
			continue
		} else {
			__antithesis_instrumentation__.Notify(95045)
		}
		__antithesis_instrumentation__.Notify(95041)
		existing := option.existing
		if betterRebalanceTarget(target, &existing, bestTarget, &replaces) == target {
			__antithesis_instrumentation__.Notify(95046)
			bestIdx = i
			bestTarget = target
			replaces = existing
		} else {
			__antithesis_instrumentation__.Notify(95047)
		}
	}
	__antithesis_instrumentation__.Notify(95037)
	if bestIdx == -1 {
		__antithesis_instrumentation__.Notify(95048)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(95049)
	}
	__antithesis_instrumentation__.Notify(95038)

	copiedTarget := *bestTarget
	options[bestIdx].candidates = options[bestIdx].candidates.removeCandidate(copiedTarget)
	return &copiedTarget, &options[bestIdx].existing
}

func betterRebalanceTarget(target1, existing1, target2, existing2 *candidate) *candidate {
	__antithesis_instrumentation__.Notify(95050)
	if target2 == nil {
		__antithesis_instrumentation__.Notify(95054)
		return target1
	} else {
		__antithesis_instrumentation__.Notify(95055)
	}
	__antithesis_instrumentation__.Notify(95051)

	comp1 := target1.compare(*existing1)
	comp2 := target2.compare(*existing2)
	if !scoresAlmostEqual(comp1, comp2) {
		__antithesis_instrumentation__.Notify(95056)
		if comp1 > comp2 {
			__antithesis_instrumentation__.Notify(95058)
			return target1
		} else {
			__antithesis_instrumentation__.Notify(95059)
		}
		__antithesis_instrumentation__.Notify(95057)
		if comp1 < comp2 {
			__antithesis_instrumentation__.Notify(95060)
			return target2
		} else {
			__antithesis_instrumentation__.Notify(95061)
		}
	} else {
		__antithesis_instrumentation__.Notify(95062)
	}
	__antithesis_instrumentation__.Notify(95052)

	if target1.less(*target2) {
		__antithesis_instrumentation__.Notify(95063)
		return target2
	} else {
		__antithesis_instrumentation__.Notify(95064)
	}
	__antithesis_instrumentation__.Notify(95053)
	return target1
}

func nodeHasReplica(nodeID roachpb.NodeID, existing []roachpb.ReplicationTarget) bool {
	__antithesis_instrumentation__.Notify(95065)
	for _, r := range existing {
		__antithesis_instrumentation__.Notify(95067)
		if r.NodeID == nodeID {
			__antithesis_instrumentation__.Notify(95068)
			return true
		} else {
			__antithesis_instrumentation__.Notify(95069)
		}
	}
	__antithesis_instrumentation__.Notify(95066)
	return false
}

func storeHasReplica(storeID roachpb.StoreID, existing []roachpb.ReplicationTarget) bool {
	__antithesis_instrumentation__.Notify(95070)
	for _, r := range existing {
		__antithesis_instrumentation__.Notify(95072)
		if r.StoreID == storeID {
			__antithesis_instrumentation__.Notify(95073)
			return true
		} else {
			__antithesis_instrumentation__.Notify(95074)
		}
	}
	__antithesis_instrumentation__.Notify(95071)
	return false
}

type constraintsCheckFn func(roachpb.StoreDescriptor) (valid, necessary bool)

type rebalanceConstraintsCheckFn func(toStore, fromStore roachpb.StoreDescriptor) (valid, necessary bool)

func voterConstraintsCheckerForAllocation(
	overallConstraints, voterConstraints constraint.AnalyzedConstraints,
) constraintsCheckFn {
	__antithesis_instrumentation__.Notify(95075)
	return func(s roachpb.StoreDescriptor) (valid, necessary bool) {
		__antithesis_instrumentation__.Notify(95076)
		overallConstraintsOK, necessaryOverall := allocateConstraintsCheck(s, overallConstraints)
		voterConstraintsOK, necessaryForVoters := allocateConstraintsCheck(s, voterConstraints)

		return overallConstraintsOK && func() bool {
				__antithesis_instrumentation__.Notify(95077)
				return voterConstraintsOK == true
			}() == true, necessaryOverall || func() bool {
				__antithesis_instrumentation__.Notify(95078)
				return necessaryForVoters == true
			}() == true
	}
}

func nonVoterConstraintsCheckerForAllocation(
	overallConstraints constraint.AnalyzedConstraints,
) constraintsCheckFn {
	__antithesis_instrumentation__.Notify(95079)
	return func(s roachpb.StoreDescriptor) (valid, necessary bool) {
		__antithesis_instrumentation__.Notify(95080)
		return allocateConstraintsCheck(s, overallConstraints)
	}
}

func voterConstraintsCheckerForRemoval(
	overallConstraints, voterConstraints constraint.AnalyzedConstraints,
) constraintsCheckFn {
	__antithesis_instrumentation__.Notify(95081)
	return func(s roachpb.StoreDescriptor) (valid, necessary bool) {
		__antithesis_instrumentation__.Notify(95082)
		overallConstraintsOK, necessaryOverall := removeConstraintsCheck(s, overallConstraints)
		voterConstraintsOK, necessaryForVoters := removeConstraintsCheck(s, voterConstraints)

		return overallConstraintsOK && func() bool {
				__antithesis_instrumentation__.Notify(95083)
				return voterConstraintsOK == true
			}() == true, necessaryOverall || func() bool {
				__antithesis_instrumentation__.Notify(95084)
				return necessaryForVoters == true
			}() == true
	}
}

func nonVoterConstraintsCheckerForRemoval(
	overallConstraints constraint.AnalyzedConstraints,
) constraintsCheckFn {
	__antithesis_instrumentation__.Notify(95085)
	return func(s roachpb.StoreDescriptor) (valid, necessary bool) {
		__antithesis_instrumentation__.Notify(95086)
		return removeConstraintsCheck(s, overallConstraints)
	}
}

func voterConstraintsCheckerForRebalance(
	overallConstraints, voterConstraints constraint.AnalyzedConstraints,
) rebalanceConstraintsCheckFn {
	__antithesis_instrumentation__.Notify(95087)
	return func(toStore, fromStore roachpb.StoreDescriptor) (valid, necessary bool) {
		__antithesis_instrumentation__.Notify(95088)
		overallConstraintsOK, necessaryOverall := rebalanceFromConstraintsCheck(toStore, fromStore, overallConstraints)
		voterConstraintsOK, necessaryForVoters := rebalanceFromConstraintsCheck(toStore, fromStore, voterConstraints)

		return overallConstraintsOK && func() bool {
				__antithesis_instrumentation__.Notify(95089)
				return voterConstraintsOK == true
			}() == true, necessaryOverall || func() bool {
				__antithesis_instrumentation__.Notify(95090)
				return necessaryForVoters == true
			}() == true
	}
}

func nonVoterConstraintsCheckerForRebalance(
	overallConstraints constraint.AnalyzedConstraints,
) rebalanceConstraintsCheckFn {
	__antithesis_instrumentation__.Notify(95091)
	return func(toStore, fromStore roachpb.StoreDescriptor) (valid, necessary bool) {
		__antithesis_instrumentation__.Notify(95092)
		return rebalanceFromConstraintsCheck(toStore, fromStore, overallConstraints)
	}
}

func allocateConstraintsCheck(
	store roachpb.StoreDescriptor, analyzed constraint.AnalyzedConstraints,
) (valid bool, necessary bool) {
	__antithesis_instrumentation__.Notify(95093)

	if len(analyzed.Constraints) == 0 {
		__antithesis_instrumentation__.Notify(95097)
		return true, false
	} else {
		__antithesis_instrumentation__.Notify(95098)
	}
	__antithesis_instrumentation__.Notify(95094)

	for i, constraints := range analyzed.Constraints {
		__antithesis_instrumentation__.Notify(95099)
		if constraintsOK := constraint.ConjunctionsCheck(
			store, constraints.Constraints,
		); constraintsOK {
			__antithesis_instrumentation__.Notify(95100)
			valid = true
			matchingStores := analyzed.SatisfiedBy[i]

			if len(matchingStores) < int(constraints.NumReplicas) {
				__antithesis_instrumentation__.Notify(95101)
				return true, true
			} else {
				__antithesis_instrumentation__.Notify(95102)
			}
		} else {
			__antithesis_instrumentation__.Notify(95103)
		}
	}
	__antithesis_instrumentation__.Notify(95095)

	if analyzed.UnconstrainedReplicas {
		__antithesis_instrumentation__.Notify(95104)
		valid = true
	} else {
		__antithesis_instrumentation__.Notify(95105)
	}
	__antithesis_instrumentation__.Notify(95096)

	return valid, false
}

func removeConstraintsCheck(
	store roachpb.StoreDescriptor, analyzed constraint.AnalyzedConstraints,
) (valid bool, necessary bool) {
	__antithesis_instrumentation__.Notify(95106)

	if len(analyzed.Constraints) == 0 {
		__antithesis_instrumentation__.Notify(95110)
		return true, false
	} else {
		__antithesis_instrumentation__.Notify(95111)
	}
	__antithesis_instrumentation__.Notify(95107)

	if len(analyzed.Satisfies[store.StoreID]) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(95112)
		return !analyzed.UnconstrainedReplicas == true
	}() == true {
		__antithesis_instrumentation__.Notify(95113)
		return false, false
	} else {
		__antithesis_instrumentation__.Notify(95114)
	}
	__antithesis_instrumentation__.Notify(95108)

	for _, constraintIdx := range analyzed.Satisfies[store.StoreID] {
		__antithesis_instrumentation__.Notify(95115)
		if len(analyzed.SatisfiedBy[constraintIdx]) <= int(analyzed.Constraints[constraintIdx].NumReplicas) {
			__antithesis_instrumentation__.Notify(95116)
			return true, true
		} else {
			__antithesis_instrumentation__.Notify(95117)
		}
	}
	__antithesis_instrumentation__.Notify(95109)

	return true, false
}

func rebalanceFromConstraintsCheck(
	store, fromStoreID roachpb.StoreDescriptor, analyzed constraint.AnalyzedConstraints,
) (valid bool, necessary bool) {
	__antithesis_instrumentation__.Notify(95118)

	if len(analyzed.Constraints) == 0 {
		__antithesis_instrumentation__.Notify(95122)
		return true, false
	} else {
		__antithesis_instrumentation__.Notify(95123)
	}
	__antithesis_instrumentation__.Notify(95119)

	for i, constraints := range analyzed.Constraints {
		__antithesis_instrumentation__.Notify(95124)
		if constraintsOK := constraint.ConjunctionsCheck(
			store, constraints.Constraints,
		); constraintsOK {
			__antithesis_instrumentation__.Notify(95125)
			valid = true
			matchingStores := analyzed.SatisfiedBy[i]
			if len(matchingStores) < int(constraints.NumReplicas) || func() bool {
				__antithesis_instrumentation__.Notify(95126)
				return (len(matchingStores) == int(constraints.NumReplicas) && func() bool {
					__antithesis_instrumentation__.Notify(95127)
					return containsStore(analyzed.SatisfiedBy[i], fromStoreID.StoreID) == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(95128)
				return true, true
			} else {
				__antithesis_instrumentation__.Notify(95129)
			}
		} else {
			__antithesis_instrumentation__.Notify(95130)
		}
	}
	__antithesis_instrumentation__.Notify(95120)

	if analyzed.UnconstrainedReplicas {
		__antithesis_instrumentation__.Notify(95131)
		valid = true
	} else {
		__antithesis_instrumentation__.Notify(95132)
	}
	__antithesis_instrumentation__.Notify(95121)

	return valid, false
}

func containsStore(stores []roachpb.StoreID, target roachpb.StoreID) bool {
	__antithesis_instrumentation__.Notify(95133)
	for _, storeID := range stores {
		__antithesis_instrumentation__.Notify(95135)
		if storeID == target {
			__antithesis_instrumentation__.Notify(95136)
			return true
		} else {
			__antithesis_instrumentation__.Notify(95137)
		}
	}
	__antithesis_instrumentation__.Notify(95134)
	return false
}

func isStoreValid(
	store roachpb.StoreDescriptor, constraints []roachpb.ConstraintsConjunction,
) bool {
	__antithesis_instrumentation__.Notify(95138)
	if len(constraints) == 0 {
		__antithesis_instrumentation__.Notify(95141)
		return true
	} else {
		__antithesis_instrumentation__.Notify(95142)
	}
	__antithesis_instrumentation__.Notify(95139)

	for _, subConstraints := range constraints {
		__antithesis_instrumentation__.Notify(95143)
		if constraintsOK := constraint.ConjunctionsCheck(
			store, subConstraints.Constraints,
		); constraintsOK {
			__antithesis_instrumentation__.Notify(95144)
			return true
		} else {
			__antithesis_instrumentation__.Notify(95145)
		}
	}
	__antithesis_instrumentation__.Notify(95140)
	return false
}

func rangeDiversityScore(existingStoreLocalities map[roachpb.StoreID]roachpb.Locality) float64 {
	__antithesis_instrumentation__.Notify(95146)
	var sumScore float64
	var numSamples int
	for s1, l1 := range existingStoreLocalities {
		__antithesis_instrumentation__.Notify(95149)
		for s2, l2 := range existingStoreLocalities {
			__antithesis_instrumentation__.Notify(95150)

			if s2 <= s1 {
				__antithesis_instrumentation__.Notify(95152)
				continue
			} else {
				__antithesis_instrumentation__.Notify(95153)
			}
			__antithesis_instrumentation__.Notify(95151)
			sumScore += l1.DiversityScore(l2)
			numSamples++
		}
	}
	__antithesis_instrumentation__.Notify(95147)
	if numSamples == 0 {
		__antithesis_instrumentation__.Notify(95154)
		return roachpb.MaxDiversityScore
	} else {
		__antithesis_instrumentation__.Notify(95155)
	}
	__antithesis_instrumentation__.Notify(95148)
	return sumScore / float64(numSamples)
}

func diversityAllocateScore(
	store roachpb.StoreDescriptor, existingStoreLocalities map[roachpb.StoreID]roachpb.Locality,
) float64 {
	__antithesis_instrumentation__.Notify(95156)
	var sumScore float64
	var numSamples int

	for _, locality := range existingStoreLocalities {
		__antithesis_instrumentation__.Notify(95159)
		newScore := store.Locality().DiversityScore(locality)
		sumScore += newScore
		numSamples++
	}
	__antithesis_instrumentation__.Notify(95157)

	if numSamples == 0 {
		__antithesis_instrumentation__.Notify(95160)
		return roachpb.MaxDiversityScore
	} else {
		__antithesis_instrumentation__.Notify(95161)
	}
	__antithesis_instrumentation__.Notify(95158)
	return sumScore / float64(numSamples)
}

func diversityRemovalScore(
	storeID roachpb.StoreID, existingStoreLocalities map[roachpb.StoreID]roachpb.Locality,
) float64 {
	__antithesis_instrumentation__.Notify(95162)
	var sumScore float64
	var numSamples int
	locality := existingStoreLocalities[storeID]

	for otherStoreID, otherLocality := range existingStoreLocalities {
		__antithesis_instrumentation__.Notify(95165)
		if otherStoreID == storeID {
			__antithesis_instrumentation__.Notify(95167)
			continue
		} else {
			__antithesis_instrumentation__.Notify(95168)
		}
		__antithesis_instrumentation__.Notify(95166)
		newScore := locality.DiversityScore(otherLocality)
		sumScore += newScore
		numSamples++
	}
	__antithesis_instrumentation__.Notify(95163)
	if numSamples == 0 {
		__antithesis_instrumentation__.Notify(95169)
		return roachpb.MaxDiversityScore
	} else {
		__antithesis_instrumentation__.Notify(95170)
	}
	__antithesis_instrumentation__.Notify(95164)
	return sumScore / float64(numSamples)
}

func diversityRebalanceScore(
	store roachpb.StoreDescriptor, existingStoreLocalities map[roachpb.StoreID]roachpb.Locality,
) float64 {
	__antithesis_instrumentation__.Notify(95171)
	if len(existingStoreLocalities) == 0 {
		__antithesis_instrumentation__.Notify(95174)
		return roachpb.MaxDiversityScore
	} else {
		__antithesis_instrumentation__.Notify(95175)
	}
	__antithesis_instrumentation__.Notify(95172)
	var maxScore float64

	for removedStoreID := range existingStoreLocalities {
		__antithesis_instrumentation__.Notify(95176)
		score := diversityRebalanceFromScore(store, removedStoreID, existingStoreLocalities)
		if score > maxScore {
			__antithesis_instrumentation__.Notify(95177)
			maxScore = score
		} else {
			__antithesis_instrumentation__.Notify(95178)
		}
	}
	__antithesis_instrumentation__.Notify(95173)
	return maxScore
}

func diversityRebalanceFromScore(
	store roachpb.StoreDescriptor,
	fromStoreID roachpb.StoreID,
	existingStoreLocalities map[roachpb.StoreID]roachpb.Locality,
) float64 {
	__antithesis_instrumentation__.Notify(95179)

	var sumScore float64
	var numSamples int
	for storeID, locality := range existingStoreLocalities {
		__antithesis_instrumentation__.Notify(95182)
		if storeID == fromStoreID {
			__antithesis_instrumentation__.Notify(95184)
			continue
		} else {
			__antithesis_instrumentation__.Notify(95185)
		}
		__antithesis_instrumentation__.Notify(95183)
		newScore := store.Locality().DiversityScore(locality)
		sumScore += newScore
		numSamples++
		for otherStoreID, otherLocality := range existingStoreLocalities {
			__antithesis_instrumentation__.Notify(95186)

			if otherStoreID <= storeID || func() bool {
				__antithesis_instrumentation__.Notify(95188)
				return otherStoreID == fromStoreID == true
			}() == true {
				__antithesis_instrumentation__.Notify(95189)
				continue
			} else {
				__antithesis_instrumentation__.Notify(95190)
			}
			__antithesis_instrumentation__.Notify(95187)
			newScore := locality.DiversityScore(otherLocality)
			sumScore += newScore
			numSamples++
		}
	}
	__antithesis_instrumentation__.Notify(95180)
	if numSamples == 0 {
		__antithesis_instrumentation__.Notify(95191)
		return roachpb.MaxDiversityScore
	} else {
		__antithesis_instrumentation__.Notify(95192)
	}
	__antithesis_instrumentation__.Notify(95181)
	return sumScore / float64(numSamples)
}

type balanceStatus int

const (
	overfull      balanceStatus = -1
	aroundTheMean balanceStatus = 0
	underfull     balanceStatus = 1
)

func overfullRangeThreshold(options *rangeCountScorerOptions, mean float64) float64 {
	__antithesis_instrumentation__.Notify(95193)
	return mean + math.Max(mean*options.rangeRebalanceThreshold, minRangeRebalanceThreshold)
}

func underfullRangeThreshold(options *rangeCountScorerOptions, mean float64) float64 {
	__antithesis_instrumentation__.Notify(95194)
	return mean - math.Max(mean*options.rangeRebalanceThreshold, minRangeRebalanceThreshold)
}

func overfullQPSThreshold(options *qpsScorerOptions, mean float64) float64 {
	__antithesis_instrumentation__.Notify(95195)
	return mean + math.Max(mean*options.qpsRebalanceThreshold, minQPSThresholdDifference)
}

func underfullQPSThreshold(options *qpsScorerOptions, mean float64) float64 {
	__antithesis_instrumentation__.Notify(95196)
	return mean - math.Max(mean*options.qpsRebalanceThreshold, minQPSThresholdDifference)
}

func rebalanceConvergesRangeCountOnMean(
	sl StoreList, sc roachpb.StoreCapacity, newRangeCount int32,
) bool {
	__antithesis_instrumentation__.Notify(95197)
	return convergesOnMean(float64(sc.RangeCount), float64(newRangeCount), sl.candidateRanges.mean)
}

func convergesOnMean(oldVal, newVal, mean float64) bool {
	__antithesis_instrumentation__.Notify(95198)
	return math.Abs(newVal-mean) < math.Abs(oldVal-mean)
}

func maxCapacityCheck(store roachpb.StoreDescriptor) bool {
	__antithesis_instrumentation__.Notify(95199)
	return store.Capacity.FractionUsed() < maxFractionUsedThreshold
}

func rebalanceToMaxCapacityCheck(store roachpb.StoreDescriptor) bool {
	__antithesis_instrumentation__.Notify(95200)
	return store.Capacity.FractionUsed() < rebalanceToMaxFractionUsedThreshold
}

func scoresAlmostEqual(score1, score2 float64) bool {
	__antithesis_instrumentation__.Notify(95201)
	return math.Abs(score1-score2) < epsilon
}
