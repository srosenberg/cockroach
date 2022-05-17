package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/storage/enginepb"

type splitStatsHelper struct {
	in splitStatsHelperInput

	absPostSplitLeft  *enginepb.MVCCStats
	absPostSplitRight *enginepb.MVCCStats
}

type splitStatsScanFn func() (enginepb.MVCCStats, error)

type splitStatsHelperInput struct {
	AbsPreSplitBothEstimated enginepb.MVCCStats
	DeltaBatchEstimated      enginepb.MVCCStats

	AbsPostSplitLeftFn splitStatsScanFn

	AbsPostSplitRightFn splitStatsScanFn

	ScanRightFirst bool
}

func makeSplitStatsHelper(input splitStatsHelperInput) (splitStatsHelper, error) {
	__antithesis_instrumentation__.Notify(97853)
	h := splitStatsHelper{
		in: input,
	}

	var absPostSplitFirst enginepb.MVCCStats
	var err error
	if h.in.ScanRightFirst {
		__antithesis_instrumentation__.Notify(97859)
		absPostSplitFirst, err = input.AbsPostSplitRightFn()
		h.absPostSplitRight = &absPostSplitFirst
	} else {
		__antithesis_instrumentation__.Notify(97860)
		absPostSplitFirst, err = input.AbsPostSplitLeftFn()
		h.absPostSplitLeft = &absPostSplitFirst
	}
	__antithesis_instrumentation__.Notify(97854)
	if err != nil {
		__antithesis_instrumentation__.Notify(97861)
		return splitStatsHelper{}, err
	} else {
		__antithesis_instrumentation__.Notify(97862)
	}
	__antithesis_instrumentation__.Notify(97855)

	if h.in.AbsPreSplitBothEstimated.ContainsEstimates == 0 && func() bool {
		__antithesis_instrumentation__.Notify(97863)
		return h.in.DeltaBatchEstimated.ContainsEstimates == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(97864)

		ms := h.in.AbsPreSplitBothEstimated
		ms.Subtract(absPostSplitFirst)
		ms.Add(h.in.DeltaBatchEstimated)
		if h.in.ScanRightFirst {
			__antithesis_instrumentation__.Notify(97866)
			h.absPostSplitLeft = &ms
		} else {
			__antithesis_instrumentation__.Notify(97867)
			h.absPostSplitRight = &ms
		}
		__antithesis_instrumentation__.Notify(97865)
		return h, nil
	} else {
		__antithesis_instrumentation__.Notify(97868)
	}
	__antithesis_instrumentation__.Notify(97856)

	var absPostSplitSecond enginepb.MVCCStats
	if h.in.ScanRightFirst {
		__antithesis_instrumentation__.Notify(97869)
		absPostSplitSecond, err = input.AbsPostSplitLeftFn()
		h.absPostSplitLeft = &absPostSplitSecond
	} else {
		__antithesis_instrumentation__.Notify(97870)
		absPostSplitSecond, err = input.AbsPostSplitRightFn()
		h.absPostSplitRight = &absPostSplitSecond
	}
	__antithesis_instrumentation__.Notify(97857)
	if err != nil {
		__antithesis_instrumentation__.Notify(97871)
		return splitStatsHelper{}, err
	} else {
		__antithesis_instrumentation__.Notify(97872)
	}
	__antithesis_instrumentation__.Notify(97858)
	return h, nil
}

func (h splitStatsHelper) AbsPostSplitRight() *enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(97873)
	return h.absPostSplitRight
}

func (h splitStatsHelper) DeltaPostSplitLeft() enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(97874)

	ms := *h.absPostSplitLeft
	ms.Subtract(h.in.AbsPreSplitBothEstimated)

	return ms
}
