package execinfrapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

func ConvertToColumnOrdering(specOrdering Ordering) colinfo.ColumnOrdering {
	__antithesis_instrumentation__.Notify(474566)
	ordering := make(colinfo.ColumnOrdering, len(specOrdering.Columns))
	for i, c := range specOrdering.Columns {
		__antithesis_instrumentation__.Notify(474568)
		ordering[i].ColIdx = int(c.ColIdx)
		if c.Direction == Ordering_Column_ASC {
			__antithesis_instrumentation__.Notify(474569)
			ordering[i].Direction = encoding.Ascending
		} else {
			__antithesis_instrumentation__.Notify(474570)
			ordering[i].Direction = encoding.Descending
		}
	}
	__antithesis_instrumentation__.Notify(474567)
	return ordering
}

func ConvertToSpecOrdering(columnOrdering colinfo.ColumnOrdering) Ordering {
	__antithesis_instrumentation__.Notify(474571)
	return ConvertToMappedSpecOrdering(columnOrdering, nil)
}

func ConvertToMappedSpecOrdering(
	columnOrdering colinfo.ColumnOrdering, planToStreamColMap []int,
) Ordering {
	__antithesis_instrumentation__.Notify(474572)
	specOrdering := Ordering{}
	specOrdering.Columns = make([]Ordering_Column, len(columnOrdering))
	for i, c := range columnOrdering {
		__antithesis_instrumentation__.Notify(474574)
		colIdx := c.ColIdx
		if planToStreamColMap != nil {
			__antithesis_instrumentation__.Notify(474576)
			colIdx = planToStreamColMap[c.ColIdx]
			if colIdx == -1 {
				__antithesis_instrumentation__.Notify(474577)
				panic(errors.AssertionFailedf("column %d in sort ordering not available", c.ColIdx))
			} else {
				__antithesis_instrumentation__.Notify(474578)
			}
		} else {
			__antithesis_instrumentation__.Notify(474579)
		}
		__antithesis_instrumentation__.Notify(474575)
		specOrdering.Columns[i].ColIdx = uint32(colIdx)
		if c.Direction == encoding.Ascending {
			__antithesis_instrumentation__.Notify(474580)
			specOrdering.Columns[i].Direction = Ordering_Column_ASC
		} else {
			__antithesis_instrumentation__.Notify(474581)
			specOrdering.Columns[i].Direction = Ordering_Column_DESC
		}
	}
	__antithesis_instrumentation__.Notify(474573)
	return specOrdering
}

func ExprFmtCtxBase(evalCtx *tree.EvalContext) *tree.FmtCtx {
	__antithesis_instrumentation__.Notify(474582)
	fmtCtx := evalCtx.FmtCtx(
		tree.FmtCheckEquivalence,
		tree.FmtPlaceholderFormat(
			func(fmtCtx *tree.FmtCtx, p *tree.Placeholder) {
				__antithesis_instrumentation__.Notify(474584)
				d, err := p.Eval(evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(474586)
					panic(errors.NewAssertionErrorWithWrappedErrf(err, "failed to serialize placeholder"))
				} else {
					__antithesis_instrumentation__.Notify(474587)
				}
				__antithesis_instrumentation__.Notify(474585)
				d.Format(fmtCtx)
			},
		),
	)
	__antithesis_instrumentation__.Notify(474583)
	return fmtCtx
}

type Expression struct {
	Version string

	Expr string

	LocalExpr tree.TypedExpr
}

func (e *Expression) Empty() bool {
	__antithesis_instrumentation__.Notify(474588)
	return e.Expr == "" && func() bool {
		__antithesis_instrumentation__.Notify(474589)
		return e.LocalExpr == nil == true
	}() == true
}

func (e Expression) String() string {
	__antithesis_instrumentation__.Notify(474590)
	if e.Expr != "" {
		__antithesis_instrumentation__.Notify(474593)
		return e.Expr
	} else {
		__antithesis_instrumentation__.Notify(474594)
	}
	__antithesis_instrumentation__.Notify(474591)
	if e.LocalExpr != nil {
		__antithesis_instrumentation__.Notify(474595)
		ctx := tree.NewFmtCtx(tree.FmtCheckEquivalence)
		ctx.FormatNode(e.LocalExpr)
		return ctx.CloseAndGetString()
	} else {
		__antithesis_instrumentation__.Notify(474596)
	}
	__antithesis_instrumentation__.Notify(474592)
	return "none"
}

func (e *Error) String() string {
	__antithesis_instrumentation__.Notify(474597)
	if err := e.ErrorDetail(context.TODO()); err != nil {
		__antithesis_instrumentation__.Notify(474599)
		return err.Error()
	} else {
		__antithesis_instrumentation__.Notify(474600)
	}
	__antithesis_instrumentation__.Notify(474598)
	return "<nil>"
}

func NewError(ctx context.Context, err error) *Error {
	__antithesis_instrumentation__.Notify(474601)
	resErr := &Error{}

	ctx = logtags.AddTag(ctx, "sent-error", nil)
	fullError := errors.EncodeError(ctx, errors.WithContextTags(err, ctx))
	resErr.FullError = &fullError

	return resErr
}

func (e *Error) ErrorDetail(ctx context.Context) (err error) {
	__antithesis_instrumentation__.Notify(474602)
	if e == nil {
		__antithesis_instrumentation__.Notify(474606)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(474607)
	}
	__antithesis_instrumentation__.Notify(474603)
	defer func() {
		__antithesis_instrumentation__.Notify(474608)
		ctx = logtags.AddTag(ctx, "received-error", nil)
		err = errors.WithContextTags(err, ctx)
	}()
	__antithesis_instrumentation__.Notify(474604)

	if e.FullError != nil {
		__antithesis_instrumentation__.Notify(474609)

		return errors.DecodeError(ctx, *e.FullError)
	} else {
		__antithesis_instrumentation__.Notify(474610)
	}
	__antithesis_instrumentation__.Notify(474605)

	return errors.AssertionFailedf("unknown error from remote node")
}

type ProducerMetadata struct {
	Ranges []roachpb.RangeInfo

	Err error

	TraceData []tracingpb.RecordedSpan

	LeafTxnFinalState *roachpb.LeafTxnFinalState

	RowNum *RemoteProducerMetadata_RowNum

	SamplerProgress *RemoteProducerMetadata_SamplerProgress

	BulkProcessorProgress *RemoteProducerMetadata_BulkProcessorProgress

	Metrics *RemoteProducerMetadata_Metrics
}

var (
	producerMetadataPool = sync.Pool{
		New: func() interface{} {
			__antithesis_instrumentation__.Notify(474611)
			return &ProducerMetadata{}
		},
	}

	rpmMetricsPool = sync.Pool{
		New: func() interface{} {
			__antithesis_instrumentation__.Notify(474612)
			return &RemoteProducerMetadata_Metrics{}
		},
	}
)

func (meta *ProducerMetadata) Release() {
	__antithesis_instrumentation__.Notify(474613)
	*meta = ProducerMetadata{}
	producerMetadataPool.Put(meta)
}

func (meta *RemoteProducerMetadata_Metrics) Release() {
	__antithesis_instrumentation__.Notify(474614)
	*meta = RemoteProducerMetadata_Metrics{}
	rpmMetricsPool.Put(meta)
}

func GetProducerMeta() *ProducerMetadata {
	__antithesis_instrumentation__.Notify(474615)
	return producerMetadataPool.Get().(*ProducerMetadata)
}

func GetMetricsMeta() *RemoteProducerMetadata_Metrics {
	__antithesis_instrumentation__.Notify(474616)
	return rpmMetricsPool.Get().(*RemoteProducerMetadata_Metrics)
}

func RemoteProducerMetaToLocalMeta(
	ctx context.Context, rpm RemoteProducerMetadata,
) (ProducerMetadata, bool) {
	__antithesis_instrumentation__.Notify(474617)
	meta := GetProducerMeta()
	switch v := rpm.Value.(type) {
	case *RemoteProducerMetadata_RangeInfo:
		__antithesis_instrumentation__.Notify(474619)
		meta.Ranges = v.RangeInfo.RangeInfo
	case *RemoteProducerMetadata_TraceData_:
		__antithesis_instrumentation__.Notify(474620)
		meta.TraceData = v.TraceData.CollectedSpans
	case *RemoteProducerMetadata_LeafTxnFinalState:
		__antithesis_instrumentation__.Notify(474621)
		meta.LeafTxnFinalState = v.LeafTxnFinalState
	case *RemoteProducerMetadata_RowNum_:
		__antithesis_instrumentation__.Notify(474622)
		meta.RowNum = v.RowNum
	case *RemoteProducerMetadata_SamplerProgress_:
		__antithesis_instrumentation__.Notify(474623)
		meta.SamplerProgress = v.SamplerProgress
	case *RemoteProducerMetadata_BulkProcessorProgress_:
		__antithesis_instrumentation__.Notify(474624)
		meta.BulkProcessorProgress = v.BulkProcessorProgress
	case *RemoteProducerMetadata_Error:
		__antithesis_instrumentation__.Notify(474625)
		meta.Err = v.Error.ErrorDetail(ctx)
	case *RemoteProducerMetadata_Metrics_:
		__antithesis_instrumentation__.Notify(474626)
		meta.Metrics = v.Metrics
	default:
		__antithesis_instrumentation__.Notify(474627)
		return *meta, false
	}
	__antithesis_instrumentation__.Notify(474618)
	return *meta, true
}

func LocalMetaToRemoteProducerMeta(
	ctx context.Context, meta ProducerMetadata,
) RemoteProducerMetadata {
	__antithesis_instrumentation__.Notify(474628)
	var rpm RemoteProducerMetadata
	if meta.Ranges != nil {
		__antithesis_instrumentation__.Notify(474630)
		rpm.Value = &RemoteProducerMetadata_RangeInfo{
			RangeInfo: &RemoteProducerMetadata_RangeInfos{
				RangeInfo: meta.Ranges,
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(474631)
		if meta.TraceData != nil {
			__antithesis_instrumentation__.Notify(474632)
			rpm.Value = &RemoteProducerMetadata_TraceData_{
				TraceData: &RemoteProducerMetadata_TraceData{
					CollectedSpans: meta.TraceData,
				},
			}
		} else {
			__antithesis_instrumentation__.Notify(474633)
			if meta.LeafTxnFinalState != nil {
				__antithesis_instrumentation__.Notify(474634)
				rpm.Value = &RemoteProducerMetadata_LeafTxnFinalState{
					LeafTxnFinalState: meta.LeafTxnFinalState,
				}
			} else {
				__antithesis_instrumentation__.Notify(474635)
				if meta.RowNum != nil {
					__antithesis_instrumentation__.Notify(474636)
					rpm.Value = &RemoteProducerMetadata_RowNum_{
						RowNum: meta.RowNum,
					}
				} else {
					__antithesis_instrumentation__.Notify(474637)
					if meta.SamplerProgress != nil {
						__antithesis_instrumentation__.Notify(474638)
						rpm.Value = &RemoteProducerMetadata_SamplerProgress_{
							SamplerProgress: meta.SamplerProgress,
						}
					} else {
						__antithesis_instrumentation__.Notify(474639)
						if meta.BulkProcessorProgress != nil {
							__antithesis_instrumentation__.Notify(474640)
							rpm.Value = &RemoteProducerMetadata_BulkProcessorProgress_{
								BulkProcessorProgress: meta.BulkProcessorProgress,
							}
						} else {
							__antithesis_instrumentation__.Notify(474641)
							if meta.Metrics != nil {
								__antithesis_instrumentation__.Notify(474642)
								rpm.Value = &RemoteProducerMetadata_Metrics_{
									Metrics: meta.Metrics,
								}
							} else {
								__antithesis_instrumentation__.Notify(474643)
								if meta.Err != nil {
									__antithesis_instrumentation__.Notify(474644)
									rpm.Value = &RemoteProducerMetadata_Error{
										Error: NewError(ctx, meta.Err),
									}
								} else {
									__antithesis_instrumentation__.Notify(474645)
									if buildutil.CrdbTestBuild {
										__antithesis_instrumentation__.Notify(474646)
										panic("unhandled field in local meta or all fields are nil")
									} else {
										__antithesis_instrumentation__.Notify(474647)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(474629)
	return rpm
}

type DistSQLRemoteFlowInfo struct {
	FlowID FlowID

	Timestamp time.Time

	StatementSQL string
}
