package colfetcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(455628)

	var x [1]struct{}
	_ = x[stateInvalid-0]
	_ = x[stateInitFetch-1]
	_ = x[stateResetBatch-2]
	_ = x[stateDecodeFirstKVOfRow-3]
	_ = x[stateFetchNextKVWithUnfinishedRow-4]
	_ = x[stateFinalizeRow-5]
	_ = x[stateEmitLastBatch-6]
	_ = x[stateFinished-7]
}

const _fetcherState_name = "stateInvalidstateInitFetchstateResetBatchstateDecodeFirstKVOfRowstateFetchNextKVWithUnfinishedRowstateFinalizeRowstateEmitLastBatchstateFinished"

var _fetcherState_index = [...]uint8{0, 12, 26, 41, 64, 97, 113, 131, 144}

func (i fetcherState) String() string {
	__antithesis_instrumentation__.Notify(455629)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(455631)
		return i >= fetcherState(len(_fetcherState_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(455632)
		return "fetcherState(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(455633)
	}
	__antithesis_instrumentation__.Notify(455630)
	return _fetcherState_name[_fetcherState_index[i]:_fetcherState_index[i+1]]
}
