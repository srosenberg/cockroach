package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/ring"
	"github.com/cockroachdb/errors"
)

type indexedValue struct {
	value tree.Datum
	idx   int
}

type slidingWindow struct {
	values  ring.Buffer
	evalCtx *tree.EvalContext
	cmp     func(*tree.EvalContext, tree.Datum, tree.Datum) int
}

func makeSlidingWindow(
	evalCtx *tree.EvalContext, cmp func(*tree.EvalContext, tree.Datum, tree.Datum) int,
) *slidingWindow {
	__antithesis_instrumentation__.Notify(602721)
	return &slidingWindow{
		evalCtx: evalCtx,
		cmp:     cmp,
	}
}

func (sw *slidingWindow) add(iv *indexedValue) {
	__antithesis_instrumentation__.Notify(602722)
	for i := sw.values.Len() - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(602724)
		if sw.cmp(sw.evalCtx, sw.values.Get(i).(*indexedValue).value, iv.value) > 0 {
			__antithesis_instrumentation__.Notify(602726)
			break
		} else {
			__antithesis_instrumentation__.Notify(602727)
		}
		__antithesis_instrumentation__.Notify(602725)
		sw.values.RemoveLast()
	}
	__antithesis_instrumentation__.Notify(602723)
	sw.values.AddLast(iv)
}

func (sw *slidingWindow) removeAllBefore(idx int) {
	__antithesis_instrumentation__.Notify(602728)
	for sw.values.Len() > 0 && func() bool {
		__antithesis_instrumentation__.Notify(602729)
		return sw.values.Get(0).(*indexedValue).idx < idx == true
	}() == true {
		__antithesis_instrumentation__.Notify(602730)
		sw.values.RemoveFirst()
	}
}

func (sw *slidingWindow) string() string {
	__antithesis_instrumentation__.Notify(602731)
	var builder strings.Builder
	for i := 0; i < sw.values.Len(); i++ {
		__antithesis_instrumentation__.Notify(602733)
		builder.WriteString(fmt.Sprintf("(%v, %v)\t", sw.values.Get(i).(*indexedValue).value, sw.values.Get(i).(*indexedValue).idx))
	}
	__antithesis_instrumentation__.Notify(602732)
	return builder.String()
}

func (sw *slidingWindow) reset() {
	__antithesis_instrumentation__.Notify(602734)
	sw.values.Reset()
}

type slidingWindowFunc struct {
	sw      *slidingWindow
	prevEnd int
}

func (w *slidingWindowFunc) Compute(
	ctx context.Context, evalCtx *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602735)
	frameStartIdx, err := wfr.FrameStartIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602741)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602742)
	}
	__antithesis_instrumentation__.Notify(602736)
	frameEndIdx, err := wfr.FrameEndIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602743)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602744)
	}
	__antithesis_instrumentation__.Notify(602737)

	if !wfr.Frame.DefaultFrameExclusion() {
		__antithesis_instrumentation__.Notify(602745)

		var res tree.Datum
		for idx := frameStartIdx; idx < frameEndIdx; idx++ {
			__antithesis_instrumentation__.Notify(602748)
			if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
				__antithesis_instrumentation__.Notify(602752)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(602753)
				if skipped {
					__antithesis_instrumentation__.Notify(602754)
					continue
				} else {
					__antithesis_instrumentation__.Notify(602755)
				}
			}
			__antithesis_instrumentation__.Notify(602749)
			args, err := wfr.ArgsByRowIdx(ctx, idx)
			if err != nil {
				__antithesis_instrumentation__.Notify(602756)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(602757)
			}
			__antithesis_instrumentation__.Notify(602750)
			if args[0] == tree.DNull {
				__antithesis_instrumentation__.Notify(602758)

				continue
			} else {
				__antithesis_instrumentation__.Notify(602759)
			}
			__antithesis_instrumentation__.Notify(602751)
			if res == nil {
				__antithesis_instrumentation__.Notify(602760)
				res = args[0]
			} else {
				__antithesis_instrumentation__.Notify(602761)
				if w.sw.cmp(evalCtx, args[0], res) > 0 {
					__antithesis_instrumentation__.Notify(602762)
					res = args[0]
				} else {
					__antithesis_instrumentation__.Notify(602763)
				}
			}
		}
		__antithesis_instrumentation__.Notify(602746)
		if res == nil {
			__antithesis_instrumentation__.Notify(602764)

			return tree.DNull, nil
		} else {
			__antithesis_instrumentation__.Notify(602765)
		}
		__antithesis_instrumentation__.Notify(602747)
		return res, nil
	} else {
		__antithesis_instrumentation__.Notify(602766)
	}
	__antithesis_instrumentation__.Notify(602738)

	w.sw.removeAllBefore(frameStartIdx)

	for idx := max(w.prevEnd, frameStartIdx); idx < frameEndIdx; idx++ {
		__antithesis_instrumentation__.Notify(602767)
		if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
			__antithesis_instrumentation__.Notify(602771)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602772)
			if skipped {
				__antithesis_instrumentation__.Notify(602773)
				continue
			} else {
				__antithesis_instrumentation__.Notify(602774)
			}
		}
		__antithesis_instrumentation__.Notify(602768)
		args, err := wfr.ArgsByRowIdx(ctx, idx)
		if err != nil {
			__antithesis_instrumentation__.Notify(602775)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602776)
		}
		__antithesis_instrumentation__.Notify(602769)
		value := args[0]
		if value == tree.DNull {
			__antithesis_instrumentation__.Notify(602777)

			continue
		} else {
			__antithesis_instrumentation__.Notify(602778)
		}
		__antithesis_instrumentation__.Notify(602770)
		w.sw.add(&indexedValue{value: value, idx: idx})
	}
	__antithesis_instrumentation__.Notify(602739)
	w.prevEnd = frameEndIdx

	if w.sw.values.Len() == 0 {
		__antithesis_instrumentation__.Notify(602779)

		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(602780)
	}
	__antithesis_instrumentation__.Notify(602740)

	return w.sw.values.GetFirst().(*indexedValue).value, nil
}

func max(a, b int) int {
	__antithesis_instrumentation__.Notify(602781)
	if a > b {
		__antithesis_instrumentation__.Notify(602783)
		return a
	} else {
		__antithesis_instrumentation__.Notify(602784)
	}
	__antithesis_instrumentation__.Notify(602782)
	return b
}

func (w *slidingWindowFunc) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(602785)
	w.prevEnd = 0
	w.sw.reset()
}

func (w *slidingWindowFunc) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602786)
	w.sw = nil
}

type slidingWindowSumFunc struct {
	agg                tree.AggregateFunc
	prevStart, prevEnd int

	lastNonNullIdx int
}

const noNonNullSeen = -1

func newSlidingWindowSumFunc(agg tree.AggregateFunc) *slidingWindowSumFunc {
	__antithesis_instrumentation__.Notify(602787)
	return &slidingWindowSumFunc{
		agg:            agg,
		lastNonNullIdx: noNonNullSeen,
	}
}

func (w *slidingWindowSumFunc) removeAllBefore(
	ctx context.Context, evalCtx *tree.EvalContext, wfr *tree.WindowFrameRun,
) error {
	__antithesis_instrumentation__.Notify(602788)
	frameStartIdx, err := wfr.FrameStartIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602791)
		return err
	} else {
		__antithesis_instrumentation__.Notify(602792)
	}
	__antithesis_instrumentation__.Notify(602789)
	for idx := w.prevStart; idx < frameStartIdx && func() bool {
		__antithesis_instrumentation__.Notify(602793)
		return idx < w.prevEnd == true
	}() == true; idx++ {
		__antithesis_instrumentation__.Notify(602794)
		if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
			__antithesis_instrumentation__.Notify(602799)
			return err
		} else {
			__antithesis_instrumentation__.Notify(602800)
			if skipped {
				__antithesis_instrumentation__.Notify(602801)
				continue
			} else {
				__antithesis_instrumentation__.Notify(602802)
			}
		}
		__antithesis_instrumentation__.Notify(602795)
		args, err := wfr.ArgsByRowIdx(ctx, idx)
		if err != nil {
			__antithesis_instrumentation__.Notify(602803)
			return err
		} else {
			__antithesis_instrumentation__.Notify(602804)
		}
		__antithesis_instrumentation__.Notify(602796)
		value := args[0]
		if value == tree.DNull {
			__antithesis_instrumentation__.Notify(602805)

			continue
		} else {
			__antithesis_instrumentation__.Notify(602806)
		}
		__antithesis_instrumentation__.Notify(602797)
		switch v := value.(type) {
		case *tree.DInt:
			__antithesis_instrumentation__.Notify(602807)
			err = w.agg.Add(ctx, tree.NewDInt(-*v))
		case *tree.DDecimal:
			__antithesis_instrumentation__.Notify(602808)
			d := tree.DDecimal{}
			d.Neg(&v.Decimal)
			err = w.agg.Add(ctx, &d)
		case *tree.DFloat:
			__antithesis_instrumentation__.Notify(602809)
			err = w.agg.Add(ctx, tree.NewDFloat(-*v))
		case *tree.DInterval:
			__antithesis_instrumentation__.Notify(602810)
			err = w.agg.Add(ctx, &tree.DInterval{Duration: duration.Duration{}.Sub(v.Duration)})
		default:
			__antithesis_instrumentation__.Notify(602811)
			err = errors.AssertionFailedf("unexpected value %v", v)
		}
		__antithesis_instrumentation__.Notify(602798)
		if err != nil {
			__antithesis_instrumentation__.Notify(602812)
			return err
		} else {
			__antithesis_instrumentation__.Notify(602813)
		}
	}
	__antithesis_instrumentation__.Notify(602790)
	return nil
}

func (w *slidingWindowSumFunc) Compute(
	ctx context.Context, evalCtx *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602814)
	frameStartIdx, err := wfr.FrameStartIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602821)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602822)
	}
	__antithesis_instrumentation__.Notify(602815)
	frameEndIdx, err := wfr.FrameEndIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602823)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602824)
	}
	__antithesis_instrumentation__.Notify(602816)
	if !wfr.Frame.DefaultFrameExclusion() {
		__antithesis_instrumentation__.Notify(602825)

		w.agg.Reset(ctx)
		for idx := frameStartIdx; idx < frameEndIdx; idx++ {
			__antithesis_instrumentation__.Notify(602827)
			if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
				__antithesis_instrumentation__.Notify(602830)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(602831)
				if skipped {
					__antithesis_instrumentation__.Notify(602832)
					continue
				} else {
					__antithesis_instrumentation__.Notify(602833)
				}
			}
			__antithesis_instrumentation__.Notify(602828)
			args, err := wfr.ArgsByRowIdx(ctx, idx)
			if err != nil {
				__antithesis_instrumentation__.Notify(602834)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(602835)
			}
			__antithesis_instrumentation__.Notify(602829)
			if err = w.agg.Add(ctx, args[0]); err != nil {
				__antithesis_instrumentation__.Notify(602836)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(602837)
			}
		}
		__antithesis_instrumentation__.Notify(602826)
		return w.agg.Result()
	} else {
		__antithesis_instrumentation__.Notify(602838)
	}
	__antithesis_instrumentation__.Notify(602817)

	if err = w.removeAllBefore(ctx, evalCtx, wfr); err != nil {
		__antithesis_instrumentation__.Notify(602839)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602840)
	}
	__antithesis_instrumentation__.Notify(602818)

	for idx := max(w.prevEnd, frameStartIdx); idx < frameEndIdx; idx++ {
		__antithesis_instrumentation__.Notify(602841)
		if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
			__antithesis_instrumentation__.Notify(602844)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602845)
			if skipped {
				__antithesis_instrumentation__.Notify(602846)
				continue
			} else {
				__antithesis_instrumentation__.Notify(602847)
			}
		}
		__antithesis_instrumentation__.Notify(602842)
		args, err := wfr.ArgsByRowIdx(ctx, idx)
		if err != nil {
			__antithesis_instrumentation__.Notify(602848)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602849)
		}
		__antithesis_instrumentation__.Notify(602843)
		if args[0] != tree.DNull {
			__antithesis_instrumentation__.Notify(602850)
			w.lastNonNullIdx = idx
			err = w.agg.Add(ctx, args[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(602851)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(602852)
			}
		} else {
			__antithesis_instrumentation__.Notify(602853)
		}
	}
	__antithesis_instrumentation__.Notify(602819)

	w.prevStart = frameStartIdx
	w.prevEnd = frameEndIdx

	onlyNulls := w.lastNonNullIdx < frameStartIdx
	if frameStartIdx == frameEndIdx || func() bool {
		__antithesis_instrumentation__.Notify(602854)
		return onlyNulls == true
	}() == true {
		__antithesis_instrumentation__.Notify(602855)

		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(602856)
	}
	__antithesis_instrumentation__.Notify(602820)
	return w.agg.Result()
}

func (w *slidingWindowSumFunc) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(602857)
	w.prevStart = 0
	w.prevEnd = 0
	w.lastNonNullIdx = noNonNullSeen
	w.agg.Reset(ctx)
}

func (w *slidingWindowSumFunc) Close(ctx context.Context, _ *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602858)
	w.agg.Close(ctx)
}

type avgWindowFunc struct {
	sum *slidingWindowSumFunc
}

func (w *avgWindowFunc) Compute(
	ctx context.Context, evalCtx *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602859)
	sum, err := w.sum.Compute(ctx, evalCtx, wfr)
	if err != nil {
		__antithesis_instrumentation__.Notify(602865)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602866)
	}
	__antithesis_instrumentation__.Notify(602860)
	if sum == tree.DNull {
		__antithesis_instrumentation__.Notify(602867)

		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(602868)
	}
	__antithesis_instrumentation__.Notify(602861)

	frameSize := 0
	frameStartIdx, err := wfr.FrameStartIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602869)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602870)
	}
	__antithesis_instrumentation__.Notify(602862)
	frameEndIdx, err := wfr.FrameEndIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602871)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602872)
	}
	__antithesis_instrumentation__.Notify(602863)
	for idx := frameStartIdx; idx < frameEndIdx; idx++ {
		__antithesis_instrumentation__.Notify(602873)
		if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
			__antithesis_instrumentation__.Notify(602877)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602878)
			if skipped {
				__antithesis_instrumentation__.Notify(602879)
				continue
			} else {
				__antithesis_instrumentation__.Notify(602880)
			}
		}
		__antithesis_instrumentation__.Notify(602874)
		args, err := wfr.ArgsByRowIdx(ctx, idx)
		if err != nil {
			__antithesis_instrumentation__.Notify(602881)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602882)
		}
		__antithesis_instrumentation__.Notify(602875)
		if args[0] == tree.DNull {
			__antithesis_instrumentation__.Notify(602883)

			continue
		} else {
			__antithesis_instrumentation__.Notify(602884)
		}
		__antithesis_instrumentation__.Notify(602876)
		frameSize++
	}
	__antithesis_instrumentation__.Notify(602864)

	switch t := sum.(type) {
	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(602885)
		return tree.NewDFloat(*t / tree.DFloat(frameSize)), nil
	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(602886)
		var avg tree.DDecimal
		count := apd.New(int64(frameSize), 0)
		_, err := tree.DecimalCtx.Quo(&avg.Decimal, &t.Decimal, count)
		return &avg, err
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(602887)
		dd := tree.DDecimal{}
		dd.SetInt64(int64(*t))
		var avg tree.DDecimal
		count := apd.New(int64(frameSize), 0)
		_, err := tree.DecimalCtx.Quo(&avg.Decimal, &dd.Decimal, count)
		return &avg, err
	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(602888)
		return &tree.DInterval{Duration: t.Duration.Div(int64(frameSize))}, nil
	default:
		__antithesis_instrumentation__.Notify(602889)
		return nil, errors.AssertionFailedf("unexpected SUM result type: %s", t)
	}
}

func (w *avgWindowFunc) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(602890)
	w.sum.Reset(ctx)
}

func (w *avgWindowFunc) Close(ctx context.Context, evalCtx *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602891)
	w.sum.Close(ctx, evalCtx)
}
