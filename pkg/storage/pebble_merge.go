package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

func sortAndDeduplicateRows(ts *roachpb.InternalTimeSeriesData) {
	__antithesis_instrumentation__.Notify(643089)

	isSortedUniq := true
	for i := 1; i < len(ts.Samples); i++ {
		__antithesis_instrumentation__.Notify(643095)
		if ts.Samples[i-1].Offset >= ts.Samples[i].Offset {
			__antithesis_instrumentation__.Notify(643096)
			isSortedUniq = false
			break
		} else {
			__antithesis_instrumentation__.Notify(643097)
		}
	}
	__antithesis_instrumentation__.Notify(643090)
	if isSortedUniq {
		__antithesis_instrumentation__.Notify(643098)
		return
	} else {
		__antithesis_instrumentation__.Notify(643099)
	}
	__antithesis_instrumentation__.Notify(643091)

	sortedSrcIdxs := make([]int, len(ts.Samples))
	for i := range sortedSrcIdxs {
		__antithesis_instrumentation__.Notify(643100)
		sortedSrcIdxs[i] = i
	}
	__antithesis_instrumentation__.Notify(643092)
	sort.SliceStable(sortedSrcIdxs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(643101)
		return ts.Samples[sortedSrcIdxs[i]].Offset < ts.Samples[sortedSrcIdxs[j]].Offset
	})
	__antithesis_instrumentation__.Notify(643093)

	uniqSortedSrcIdxs := make([]int, 0, len(ts.Samples))
	for destIdx := range sortedSrcIdxs {
		__antithesis_instrumentation__.Notify(643102)
		if destIdx == len(sortedSrcIdxs)-1 || func() bool {
			__antithesis_instrumentation__.Notify(643103)
			return ts.Samples[sortedSrcIdxs[destIdx]].Offset != ts.Samples[sortedSrcIdxs[destIdx+1]].Offset == true
		}() == true {
			__antithesis_instrumentation__.Notify(643104)
			uniqSortedSrcIdxs = append(uniqSortedSrcIdxs, sortedSrcIdxs[destIdx])
		} else {
			__antithesis_instrumentation__.Notify(643105)
		}
	}
	__antithesis_instrumentation__.Notify(643094)

	origSamples := ts.Samples
	ts.Samples = make([]roachpb.InternalTimeSeriesSample, len(uniqSortedSrcIdxs))

	for destIdx, srcIdx := range uniqSortedSrcIdxs {
		__antithesis_instrumentation__.Notify(643106)
		ts.Samples[destIdx] = origSamples[srcIdx]
	}
}

func sortAndDeduplicateColumns(ts *roachpb.InternalTimeSeriesData) {
	__antithesis_instrumentation__.Notify(643107)

	isSortedUniq := true
	for i := 1; i < len(ts.Offset); i++ {
		__antithesis_instrumentation__.Notify(643114)
		if ts.Offset[i-1] >= ts.Offset[i] {
			__antithesis_instrumentation__.Notify(643115)
			isSortedUniq = false
			break
		} else {
			__antithesis_instrumentation__.Notify(643116)
		}
	}
	__antithesis_instrumentation__.Notify(643108)
	if isSortedUniq {
		__antithesis_instrumentation__.Notify(643117)
		return
	} else {
		__antithesis_instrumentation__.Notify(643118)
	}
	__antithesis_instrumentation__.Notify(643109)

	sortedSrcIdxs := make([]int, len(ts.Offset))
	for i := range sortedSrcIdxs {
		__antithesis_instrumentation__.Notify(643119)
		sortedSrcIdxs[i] = i
	}
	__antithesis_instrumentation__.Notify(643110)
	sort.SliceStable(sortedSrcIdxs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(643120)
		return ts.Offset[sortedSrcIdxs[i]] < ts.Offset[sortedSrcIdxs[j]]
	})
	__antithesis_instrumentation__.Notify(643111)

	uniqSortedSrcIdxs := make([]int, 0, len(ts.Offset))
	for destIdx := range sortedSrcIdxs {
		__antithesis_instrumentation__.Notify(643121)
		if destIdx == len(sortedSrcIdxs)-1 || func() bool {
			__antithesis_instrumentation__.Notify(643122)
			return ts.Offset[sortedSrcIdxs[destIdx]] != ts.Offset[sortedSrcIdxs[destIdx+1]] == true
		}() == true {
			__antithesis_instrumentation__.Notify(643123)
			uniqSortedSrcIdxs = append(uniqSortedSrcIdxs, sortedSrcIdxs[destIdx])
		} else {
			__antithesis_instrumentation__.Notify(643124)
		}
	}
	__antithesis_instrumentation__.Notify(643112)

	origOffset, origLast, origCount, origSum, origMin, origMax, origFirst, origVariance :=
		ts.Offset, ts.Last, ts.Count, ts.Sum, ts.Min, ts.Max, ts.First, ts.Variance
	ts.Offset = make([]int32, len(uniqSortedSrcIdxs))
	ts.Last = make([]float64, len(uniqSortedSrcIdxs))

	if len(origCount) > 0 {
		__antithesis_instrumentation__.Notify(643125)
		ts.Count = make([]uint32, len(uniqSortedSrcIdxs))
		ts.Sum = make([]float64, len(uniqSortedSrcIdxs))
		ts.Min = make([]float64, len(uniqSortedSrcIdxs))
		ts.Max = make([]float64, len(uniqSortedSrcIdxs))
		ts.First = make([]float64, len(uniqSortedSrcIdxs))
		ts.Variance = make([]float64, len(uniqSortedSrcIdxs))
	} else {
		__antithesis_instrumentation__.Notify(643126)
	}
	__antithesis_instrumentation__.Notify(643113)

	for destIdx, srcIdx := range uniqSortedSrcIdxs {
		__antithesis_instrumentation__.Notify(643127)
		ts.Offset[destIdx] = origOffset[srcIdx]
		ts.Last[destIdx] = origLast[srcIdx]

		if len(origCount) > 0 {
			__antithesis_instrumentation__.Notify(643128)
			ts.Count[destIdx] = origCount[srcIdx]
			ts.Sum[destIdx] = origSum[srcIdx]
			ts.Min[destIdx] = origMin[srcIdx]
			ts.Max[destIdx] = origMax[srcIdx]
			ts.First[destIdx] = origFirst[srcIdx]
			ts.Variance[destIdx] = origVariance[srcIdx]
		} else {
			__antithesis_instrumentation__.Notify(643129)
		}
	}
}

func ensureColumnar(ts *roachpb.InternalTimeSeriesData) {
	__antithesis_instrumentation__.Notify(643130)
	for _, sample := range ts.Samples {
		__antithesis_instrumentation__.Notify(643132)
		ts.Offset = append(ts.Offset, sample.Offset)
		ts.Last = append(ts.Last, sample.Sum)
	}
	__antithesis_instrumentation__.Notify(643131)
	ts.Samples = ts.Samples[:0]
}

type MVCCValueMerger struct {
	timeSeriesOps []roachpb.InternalTimeSeriesData
	rawByteOps    [][]byte
	oldestMergeTS hlc.LegacyTimestamp
	oldToNew      bool

	meta enginepb.MVCCMetadata
}

const (
	mvccChecksumSize = 4
	mvccTagPos       = mvccChecksumSize
	mvccHeaderSize   = mvccChecksumSize + 1
)

func (t *MVCCValueMerger) ensureOrder(oldToNew bool) {
	__antithesis_instrumentation__.Notify(643133)
	if oldToNew == t.oldToNew {
		__antithesis_instrumentation__.Notify(643137)
		return
	} else {
		__antithesis_instrumentation__.Notify(643138)
	}
	__antithesis_instrumentation__.Notify(643134)

	for i := 0; i < len(t.timeSeriesOps)/2; i++ {
		__antithesis_instrumentation__.Notify(643139)
		t.timeSeriesOps[i], t.timeSeriesOps[len(t.timeSeriesOps)-1-i] = t.timeSeriesOps[len(t.timeSeriesOps)-1-i], t.timeSeriesOps[i]
	}
	__antithesis_instrumentation__.Notify(643135)
	for i := 0; i < len(t.rawByteOps)/2; i++ {
		__antithesis_instrumentation__.Notify(643140)
		t.rawByteOps[i], t.rawByteOps[len(t.rawByteOps)-1-i] = t.rawByteOps[len(t.rawByteOps)-1-i], t.rawByteOps[i]
	}
	__antithesis_instrumentation__.Notify(643136)
	t.oldToNew = oldToNew
}

func (t *MVCCValueMerger) deserializeMVCCValueAndAppend(value []byte) error {
	__antithesis_instrumentation__.Notify(643141)
	if err := protoutil.Unmarshal(value, &t.meta); err != nil {
		__antithesis_instrumentation__.Notify(643146)
		return errors.Wrap(err, "corrupted operand value")
	} else {
		__antithesis_instrumentation__.Notify(643147)
	}
	__antithesis_instrumentation__.Notify(643142)
	if len(t.meta.RawBytes) < mvccHeaderSize {
		__antithesis_instrumentation__.Notify(643148)
		return errors.Errorf("operand value too short")
	} else {
		__antithesis_instrumentation__.Notify(643149)
	}
	__antithesis_instrumentation__.Notify(643143)
	if t.meta.RawBytes[mvccTagPos] == byte(roachpb.ValueType_TIMESERIES) {
		__antithesis_instrumentation__.Notify(643150)
		if t.rawByteOps != nil {
			__antithesis_instrumentation__.Notify(643152)
			return errors.Errorf("inconsistent value types for timeseries merge")
		} else {
			__antithesis_instrumentation__.Notify(643153)
		}
		__antithesis_instrumentation__.Notify(643151)
		t.timeSeriesOps = append(t.timeSeriesOps, roachpb.InternalTimeSeriesData{})
		ts := &t.timeSeriesOps[len(t.timeSeriesOps)-1]
		if err := protoutil.Unmarshal(t.meta.RawBytes[mvccHeaderSize:], ts); err != nil {
			__antithesis_instrumentation__.Notify(643154)
			return errors.Wrap(err, "corrupted timeseries")
		} else {
			__antithesis_instrumentation__.Notify(643155)
		}
	} else {
		__antithesis_instrumentation__.Notify(643156)
		if t.timeSeriesOps != nil {
			__antithesis_instrumentation__.Notify(643158)
			return errors.Errorf("inconsistent value types for non-timeseries merge")
		} else {
			__antithesis_instrumentation__.Notify(643159)
		}
		__antithesis_instrumentation__.Notify(643157)
		t.rawByteOps = append(t.rawByteOps, t.meta.RawBytes[mvccHeaderSize:])
	}
	__antithesis_instrumentation__.Notify(643144)

	if t.meta.MergeTimestamp != nil && func() bool {
		__antithesis_instrumentation__.Notify(643160)
		return (t.oldestMergeTS == hlc.LegacyTimestamp{} || func() bool {
			__antithesis_instrumentation__.Notify(643161)
			return !t.oldToNew == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(643162)
		t.oldestMergeTS = *t.meta.MergeTimestamp
	} else {
		__antithesis_instrumentation__.Notify(643163)
	}
	__antithesis_instrumentation__.Notify(643145)
	return nil
}

func (t *MVCCValueMerger) MergeNewer(value []byte) error {
	__antithesis_instrumentation__.Notify(643164)
	t.ensureOrder(true)
	if err := t.deserializeMVCCValueAndAppend(value); err != nil {
		__antithesis_instrumentation__.Notify(643166)
		return err
	} else {
		__antithesis_instrumentation__.Notify(643167)
	}
	__antithesis_instrumentation__.Notify(643165)
	return nil
}

func (t *MVCCValueMerger) MergeOlder(value []byte) error {
	__antithesis_instrumentation__.Notify(643168)
	t.ensureOrder(false)
	if err := t.deserializeMVCCValueAndAppend(value); err != nil {
		__antithesis_instrumentation__.Notify(643170)
		return err
	} else {
		__antithesis_instrumentation__.Notify(643171)
	}
	__antithesis_instrumentation__.Notify(643169)
	return nil
}

func (t *MVCCValueMerger) Finish(includesBase bool) ([]byte, io.Closer, error) {
	__antithesis_instrumentation__.Notify(643172)
	isColumnar := false
	if t.timeSeriesOps == nil && func() bool {
		__antithesis_instrumentation__.Notify(643180)
		return t.rawByteOps == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(643181)
		return nil, nil, errors.Errorf("empty merge unsupported")
	} else {
		__antithesis_instrumentation__.Notify(643182)
	}
	__antithesis_instrumentation__.Notify(643173)
	t.ensureOrder(true)
	if t.timeSeriesOps == nil {
		__antithesis_instrumentation__.Notify(643183)

		totalLen := 0
		for _, rawByteOp := range t.rawByteOps {
			__antithesis_instrumentation__.Notify(643187)
			totalLen += len(rawByteOp)
		}
		__antithesis_instrumentation__.Notify(643184)

		var meta enginepb.MVCCMetadataSubsetForMergeSerialization
		meta.RawBytes = make([]byte, mvccHeaderSize, mvccHeaderSize+totalLen)
		meta.RawBytes[mvccTagPos] = byte(roachpb.ValueType_BYTES)
		for _, rawByteOp := range t.rawByteOps {
			__antithesis_instrumentation__.Notify(643188)
			meta.RawBytes = append(meta.RawBytes, rawByteOp...)
		}
		__antithesis_instrumentation__.Notify(643185)
		res, err := protoutil.Marshal(&meta)
		if err != nil {
			__antithesis_instrumentation__.Notify(643189)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(643190)
		}
		__antithesis_instrumentation__.Notify(643186)
		return res, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(643191)
	}
	__antithesis_instrumentation__.Notify(643174)

	var merged roachpb.InternalTimeSeriesData
	merged.StartTimestampNanos = t.timeSeriesOps[0].StartTimestampNanos
	merged.SampleDurationNanos = t.timeSeriesOps[0].SampleDurationNanos
	for _, timeSeriesOp := range t.timeSeriesOps {
		__antithesis_instrumentation__.Notify(643192)
		if timeSeriesOp.StartTimestampNanos != merged.StartTimestampNanos {
			__antithesis_instrumentation__.Notify(643196)
			return nil, nil, errors.Errorf("start timestamp mismatch")
		} else {
			__antithesis_instrumentation__.Notify(643197)
		}
		__antithesis_instrumentation__.Notify(643193)
		if timeSeriesOp.SampleDurationNanos != merged.SampleDurationNanos {
			__antithesis_instrumentation__.Notify(643198)
			return nil, nil, errors.Errorf("sample duration mismatch")
		} else {
			__antithesis_instrumentation__.Notify(643199)
		}
		__antithesis_instrumentation__.Notify(643194)
		if !isColumnar && func() bool {
			__antithesis_instrumentation__.Notify(643200)
			return len(timeSeriesOp.Offset) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(643201)
			ensureColumnar(&merged)
			ensureColumnar(&timeSeriesOp)
			isColumnar = true
		} else {
			__antithesis_instrumentation__.Notify(643202)
			if isColumnar {
				__antithesis_instrumentation__.Notify(643203)
				ensureColumnar(&timeSeriesOp)
			} else {
				__antithesis_instrumentation__.Notify(643204)
			}
		}
		__antithesis_instrumentation__.Notify(643195)
		proto.Merge(&merged, &timeSeriesOp)
	}
	__antithesis_instrumentation__.Notify(643175)
	if isColumnar {
		__antithesis_instrumentation__.Notify(643205)
		sortAndDeduplicateColumns(&merged)
	} else {
		__antithesis_instrumentation__.Notify(643206)
		sortAndDeduplicateRows(&merged)
	}
	__antithesis_instrumentation__.Notify(643176)
	tsBytes, err := protoutil.Marshal(&merged)
	if err != nil {
		__antithesis_instrumentation__.Notify(643207)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(643208)
	}
	__antithesis_instrumentation__.Notify(643177)

	var meta enginepb.MVCCMetadataSubsetForMergeSerialization
	if !(t.oldestMergeTS == hlc.LegacyTimestamp{}) {
		__antithesis_instrumentation__.Notify(643209)
		meta.MergeTimestamp = &t.oldestMergeTS
	} else {
		__antithesis_instrumentation__.Notify(643210)
	}
	__antithesis_instrumentation__.Notify(643178)
	tsTag := byte(roachpb.ValueType_TIMESERIES)
	header := make([]byte, mvccHeaderSize)
	header[mvccTagPos] = tsTag
	meta.RawBytes = append(header, tsBytes...)
	res, err := protoutil.Marshal(&meta)
	if err != nil {
		__antithesis_instrumentation__.Notify(643211)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(643212)
	}
	__antithesis_instrumentation__.Notify(643179)
	return res, nil, nil
}

func serializeMergeInputs(sources ...roachpb.InternalTimeSeriesData) ([][]byte, error) {
	__antithesis_instrumentation__.Notify(643213)

	srcBytes := make([][]byte, 0, len(sources))
	var val roachpb.Value
	for _, src := range sources {
		__antithesis_instrumentation__.Notify(643215)
		if err := val.SetProto(&src); err != nil {
			__antithesis_instrumentation__.Notify(643218)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(643219)
		}
		__antithesis_instrumentation__.Notify(643216)
		bytes, err := protoutil.Marshal(&enginepb.MVCCMetadata{
			RawBytes: val.RawBytes,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(643220)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(643221)
		}
		__antithesis_instrumentation__.Notify(643217)
		srcBytes = append(srcBytes, bytes)
	}
	__antithesis_instrumentation__.Notify(643214)
	return srcBytes, nil
}

func deserializeMergeOutput(mergedBytes []byte) (roachpb.InternalTimeSeriesData, error) {
	__antithesis_instrumentation__.Notify(643222)

	var meta enginepb.MVCCMetadata
	if err := protoutil.Unmarshal(mergedBytes, &meta); err != nil {
		__antithesis_instrumentation__.Notify(643225)
		return roachpb.InternalTimeSeriesData{}, err
	} else {
		__antithesis_instrumentation__.Notify(643226)
	}
	__antithesis_instrumentation__.Notify(643223)
	mergedTS, err := MakeValue(meta).GetTimeseries()
	if err != nil {
		__antithesis_instrumentation__.Notify(643227)
		return roachpb.InternalTimeSeriesData{}, err
	} else {
		__antithesis_instrumentation__.Notify(643228)
	}
	__antithesis_instrumentation__.Notify(643224)
	return mergedTS, nil
}

func MergeInternalTimeSeriesData(
	usePartialMerge bool, sources ...roachpb.InternalTimeSeriesData,
) (roachpb.InternalTimeSeriesData, error) {
	__antithesis_instrumentation__.Notify(643229)

	var mvccMerger MVCCValueMerger
	srcBytes, err := serializeMergeInputs(sources...)
	if err != nil {
		__antithesis_instrumentation__.Notify(643234)
		return roachpb.InternalTimeSeriesData{}, err
	} else {
		__antithesis_instrumentation__.Notify(643235)
	}
	__antithesis_instrumentation__.Notify(643230)
	for _, bytes := range srcBytes {
		__antithesis_instrumentation__.Notify(643236)
		if err := mvccMerger.MergeNewer(bytes); err != nil {
			__antithesis_instrumentation__.Notify(643237)
			return roachpb.InternalTimeSeriesData{}, err
		} else {
			__antithesis_instrumentation__.Notify(643238)
		}
	}
	__antithesis_instrumentation__.Notify(643231)
	resBytes, closer, err := mvccMerger.Finish(!usePartialMerge)
	if err != nil {
		__antithesis_instrumentation__.Notify(643239)
		return roachpb.InternalTimeSeriesData{}, err
	} else {
		__antithesis_instrumentation__.Notify(643240)
	}
	__antithesis_instrumentation__.Notify(643232)
	res, err := deserializeMergeOutput(resBytes)
	if closer != nil {
		__antithesis_instrumentation__.Notify(643241)
		_ = closer.Close()
	} else {
		__antithesis_instrumentation__.Notify(643242)
	}
	__antithesis_instrumentation__.Notify(643233)
	return res, err
}
