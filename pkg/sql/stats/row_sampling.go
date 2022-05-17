package stats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type SampledRow struct {
	Row  rowenc.EncDatumRow
	Rank uint64
}

type SampleReservoir struct {
	samples  []SampledRow
	colTypes []*types.T
	da       tree.DatumAlloc
	ra       rowenc.EncDatumRowAlloc
	memAcc   *mon.BoundAccount

	minNumSamples int

	sampleCols util.FastIntSet
}

var _ heap.Interface = &SampleReservoir{}

func (sr *SampleReservoir) Init(
	numSamples, minNumSamples int,
	colTypes []*types.T,
	memAcc *mon.BoundAccount,
	sampleCols util.FastIntSet,
) {
	__antithesis_instrumentation__.Notify(626734)
	if minNumSamples < 1 || func() bool {
		__antithesis_instrumentation__.Notify(626736)
		return minNumSamples > numSamples == true
	}() == true {
		__antithesis_instrumentation__.Notify(626737)
		minNumSamples = numSamples
	} else {
		__antithesis_instrumentation__.Notify(626738)
	}
	__antithesis_instrumentation__.Notify(626735)
	sr.samples = make([]SampledRow, 0, numSamples)
	sr.minNumSamples = minNumSamples
	sr.colTypes = colTypes
	sr.memAcc = memAcc
	sr.sampleCols = sampleCols
}

func (sr *SampleReservoir) Disable() {
	__antithesis_instrumentation__.Notify(626739)
	sr.samples = nil
}

func (sr *SampleReservoir) Len() int {
	__antithesis_instrumentation__.Notify(626740)
	return len(sr.samples)
}

func (sr *SampleReservoir) Cap() int {
	__antithesis_instrumentation__.Notify(626741)
	return cap(sr.samples)
}

func (sr *SampleReservoir) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(626742)

	return sr.samples[i].Rank > sr.samples[j].Rank
}

func (sr *SampleReservoir) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(626743)
	sr.samples[i], sr.samples[j] = sr.samples[j], sr.samples[i]
}

func (sr *SampleReservoir) Push(x interface{}) {
	__antithesis_instrumentation__.Notify(626744)
	panic("unimplemented")
}

func (sr *SampleReservoir) Pop() interface{} {
	__antithesis_instrumentation__.Notify(626745)
	n := len(sr.samples)
	samp := sr.samples[n-1]
	sr.samples[n-1] = SampledRow{}
	sr.samples = sr.samples[:n-1]
	return samp
}

func (sr *SampleReservoir) MaybeResize(ctx context.Context, k int) bool {
	__antithesis_instrumentation__.Notify(626746)
	if k >= cap(sr.samples) {
		__antithesis_instrumentation__.Notify(626749)
		return false
	} else {
		__antithesis_instrumentation__.Notify(626750)
	}
	__antithesis_instrumentation__.Notify(626747)

	heap.Init(sr)
	for len(sr.samples) > k {
		__antithesis_instrumentation__.Notify(626751)
		samp := heap.Pop(sr).(SampledRow)
		if sr.memAcc != nil {
			__antithesis_instrumentation__.Notify(626752)
			sr.memAcc.Shrink(ctx, int64(samp.Row.Size()))
		} else {
			__antithesis_instrumentation__.Notify(626753)
		}
	}
	__antithesis_instrumentation__.Notify(626748)

	samples := make([]SampledRow, len(sr.samples), k)
	copy(samples, sr.samples)
	sr.samples = samples
	return true
}

func (sr *SampleReservoir) retryMaybeResize(ctx context.Context, op func() error) error {
	__antithesis_instrumentation__.Notify(626754)
	for {
		__antithesis_instrumentation__.Notify(626755)
		if err := op(); err == nil || func() bool {
			__antithesis_instrumentation__.Notify(626757)
			return !sqlerrors.IsOutOfMemoryError(err) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(626758)
			return len(sr.samples) == 0 == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(626759)
			return len(sr.samples)/2 < sr.minNumSamples == true
		}() == true {
			__antithesis_instrumentation__.Notify(626760)
			return err
		} else {
			__antithesis_instrumentation__.Notify(626761)
		}
		__antithesis_instrumentation__.Notify(626756)

		sr.MaybeResize(ctx, len(sr.samples)/2)
	}
}

func (sr *SampleReservoir) SampleRow(
	ctx context.Context, evalCtx *tree.EvalContext, row rowenc.EncDatumRow, rank uint64,
) error {
	__antithesis_instrumentation__.Notify(626762)
	return sr.retryMaybeResize(ctx, func() error {
		__antithesis_instrumentation__.Notify(626763)
		if len(sr.samples) < cap(sr.samples) {
			__antithesis_instrumentation__.Notify(626766)

			rowCopy := sr.ra.AllocRow(len(row))

			if sr.memAcc != nil {
				__antithesis_instrumentation__.Notify(626770)
				if err := sr.memAcc.Grow(ctx, int64(rowCopy.Size())); err != nil {
					__antithesis_instrumentation__.Notify(626771)
					return err
				} else {
					__antithesis_instrumentation__.Notify(626772)
				}
			} else {
				__antithesis_instrumentation__.Notify(626773)
			}
			__antithesis_instrumentation__.Notify(626767)
			if err := sr.copyRow(ctx, evalCtx, rowCopy, row); err != nil {
				__antithesis_instrumentation__.Notify(626774)
				return err
			} else {
				__antithesis_instrumentation__.Notify(626775)
			}
			__antithesis_instrumentation__.Notify(626768)
			sr.samples = append(sr.samples, SampledRow{Row: rowCopy, Rank: rank})
			if len(sr.samples) == cap(sr.samples) {
				__antithesis_instrumentation__.Notify(626776)

				heap.Init(sr)
			} else {
				__antithesis_instrumentation__.Notify(626777)
			}
			__antithesis_instrumentation__.Notify(626769)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(626778)
		}
		__antithesis_instrumentation__.Notify(626764)

		if len(sr.samples) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(626779)
			return rank < sr.samples[0].Rank == true
		}() == true {
			__antithesis_instrumentation__.Notify(626780)
			if err := sr.copyRow(ctx, evalCtx, sr.samples[0].Row, row); err != nil {
				__antithesis_instrumentation__.Notify(626782)

				return err
			} else {
				__antithesis_instrumentation__.Notify(626783)
			}
			__antithesis_instrumentation__.Notify(626781)
			sr.samples[0].Rank = rank
			heap.Fix(sr, 0)
		} else {
			__antithesis_instrumentation__.Notify(626784)
		}
		__antithesis_instrumentation__.Notify(626765)
		return nil
	})
}

func (sr *SampleReservoir) Get() []SampledRow {
	__antithesis_instrumentation__.Notify(626785)
	return sr.samples
}

func (sr *SampleReservoir) GetNonNullDatums(
	ctx context.Context, memAcc *mon.BoundAccount, colIdx int,
) (values tree.Datums, err error) {
	__antithesis_instrumentation__.Notify(626786)
	err = sr.retryMaybeResize(ctx, func() error {
		__antithesis_instrumentation__.Notify(626788)

		if memAcc != nil {
			__antithesis_instrumentation__.Notify(626791)
			if err := memAcc.Grow(ctx, memsize.DatumOverhead*int64(len(sr.samples))); err != nil {
				__antithesis_instrumentation__.Notify(626792)
				return err
			} else {
				__antithesis_instrumentation__.Notify(626793)
			}
		} else {
			__antithesis_instrumentation__.Notify(626794)
		}
		__antithesis_instrumentation__.Notify(626789)
		values = make(tree.Datums, 0, len(sr.samples))
		for _, sample := range sr.samples {
			__antithesis_instrumentation__.Notify(626795)
			ed := &sample.Row[colIdx]
			if ed.Datum == nil {
				__antithesis_instrumentation__.Notify(626797)
				values = nil
				return errors.AssertionFailedf("value in column %d not decoded", colIdx)
			} else {
				__antithesis_instrumentation__.Notify(626798)
			}
			__antithesis_instrumentation__.Notify(626796)
			if !ed.IsNull() {
				__antithesis_instrumentation__.Notify(626799)
				values = append(values, ed.Datum)
			} else {
				__antithesis_instrumentation__.Notify(626800)
			}
		}
		__antithesis_instrumentation__.Notify(626790)
		return nil
	})
	__antithesis_instrumentation__.Notify(626787)
	return
}

func (sr *SampleReservoir) copyRow(
	ctx context.Context, evalCtx *tree.EvalContext, dst, src rowenc.EncDatumRow,
) error {
	__antithesis_instrumentation__.Notify(626801)
	for i := range src {
		__antithesis_instrumentation__.Notify(626803)
		if !sr.sampleCols.Contains(i) {
			__antithesis_instrumentation__.Notify(626807)
			dst[i].Datum = tree.DNull
			continue
		} else {
			__antithesis_instrumentation__.Notify(626808)
		}
		__antithesis_instrumentation__.Notify(626804)

		if err := src[i].EnsureDecoded(sr.colTypes[i], &sr.da); err != nil {
			__antithesis_instrumentation__.Notify(626809)
			return err
		} else {
			__antithesis_instrumentation__.Notify(626810)
		}
		__antithesis_instrumentation__.Notify(626805)
		beforeSize := dst[i].Size()
		dst[i] = rowenc.DatumToEncDatum(sr.colTypes[i], src[i].Datum)
		afterSize := dst[i].Size()

		if afterSize > uintptr(maxBytesPerSample) {
			__antithesis_instrumentation__.Notify(626811)
			dst[i].Datum = truncateDatum(evalCtx, dst[i].Datum, maxBytesPerSample)
			afterSize = dst[i].Size()
		} else {
			__antithesis_instrumentation__.Notify(626812)
			dst[i].Datum = deepCopyDatum(evalCtx, dst[i].Datum)
		}
		__antithesis_instrumentation__.Notify(626806)

		if sr.memAcc != nil {
			__antithesis_instrumentation__.Notify(626813)
			if err := sr.memAcc.Resize(ctx, int64(beforeSize), int64(afterSize)); err != nil {
				__antithesis_instrumentation__.Notify(626814)
				return err
			} else {
				__antithesis_instrumentation__.Notify(626815)
			}
		} else {
			__antithesis_instrumentation__.Notify(626816)
		}
	}
	__antithesis_instrumentation__.Notify(626802)
	return nil
}

const maxBytesPerSample = 400

func truncateDatum(evalCtx *tree.EvalContext, d tree.Datum, maxBytes int) tree.Datum {
	__antithesis_instrumentation__.Notify(626817)
	switch t := d.(type) {
	case *tree.DBitArray:
		__antithesis_instrumentation__.Notify(626818)
		b := tree.DBitArray{BitArray: t.ToWidth(uint(maxBytes * 8))}
		return &b

	case *tree.DBytes:
		__antithesis_instrumentation__.Notify(626819)

		b := make([]byte, maxBytes)
		copy(b, *t)
		return tree.NewDBytes(tree.DBytes(b))

	case *tree.DString:
		__antithesis_instrumentation__.Notify(626820)
		return tree.NewDString(truncateString(string(*t), maxBytes))

	case *tree.DCollatedString:
		__antithesis_instrumentation__.Notify(626821)
		contents := truncateString(t.Contents, maxBytes)

		res, err := tree.NewDCollatedString(contents, t.Locale, &evalCtx.CollationEnv)
		if err != nil {
			__antithesis_instrumentation__.Notify(626825)
			return d
		} else {
			__antithesis_instrumentation__.Notify(626826)
		}
		__antithesis_instrumentation__.Notify(626822)
		return res

	case *tree.DOidWrapper:
		__antithesis_instrumentation__.Notify(626823)
		return &tree.DOidWrapper{
			Wrapped: truncateDatum(evalCtx, t.Wrapped, maxBytes),
			Oid:     t.Oid,
		}

	default:
		__antithesis_instrumentation__.Notify(626824)

		return d
	}
}

func truncateString(s string, maxBytes int) string {
	__antithesis_instrumentation__.Notify(626827)
	last := 0

	for i := range s {
		__antithesis_instrumentation__.Notify(626829)
		if i > maxBytes {
			__antithesis_instrumentation__.Notify(626831)
			break
		} else {
			__antithesis_instrumentation__.Notify(626832)
		}
		__antithesis_instrumentation__.Notify(626830)
		last = i
	}
	__antithesis_instrumentation__.Notify(626828)

	b := make([]byte, last)
	copy(b, s)
	return string(b)
}

func deepCopyDatum(evalCtx *tree.EvalContext, d tree.Datum) tree.Datum {
	__antithesis_instrumentation__.Notify(626833)
	switch t := d.(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(626834)
		return tree.NewDString(deepCopyString(string(*t)))

	case *tree.DCollatedString:
		__antithesis_instrumentation__.Notify(626835)
		return &tree.DCollatedString{
			Contents: deepCopyString(t.Contents),
			Locale:   t.Locale,
			Key:      t.Key,
		}

	case *tree.DOidWrapper:
		__antithesis_instrumentation__.Notify(626836)
		return &tree.DOidWrapper{
			Wrapped: deepCopyDatum(evalCtx, t.Wrapped),
			Oid:     t.Oid,
		}

	default:
		__antithesis_instrumentation__.Notify(626837)

		return d
	}
}

func deepCopyString(s string) string {
	__antithesis_instrumentation__.Notify(626838)
	b := make([]byte, len(s))
	copy(b, s)
	return string(b)
}
