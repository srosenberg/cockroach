package execinfrapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treewindow"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/errors"
)

func GetAggregateFuncIdx(funcName string) (int32, error) {
	__antithesis_instrumentation__.Notify(478376)
	funcStr := strings.ToUpper(funcName)
	funcIdx, ok := AggregatorSpec_Func_value[funcStr]
	if !ok {
		__antithesis_instrumentation__.Notify(478378)
		return 0, errors.Errorf("unknown aggregate %s", funcStr)
	} else {
		__antithesis_instrumentation__.Notify(478379)
	}
	__antithesis_instrumentation__.Notify(478377)
	return funcIdx, nil
}

func (a AggregatorSpec_Aggregation) Equals(b AggregatorSpec_Aggregation) bool {
	__antithesis_instrumentation__.Notify(478380)
	if a.Func != b.Func || func() bool {
		__antithesis_instrumentation__.Notify(478385)
		return a.Distinct != b.Distinct == true
	}() == true {
		__antithesis_instrumentation__.Notify(478386)
		return false
	} else {
		__antithesis_instrumentation__.Notify(478387)
	}
	__antithesis_instrumentation__.Notify(478381)
	if a.FilterColIdx == nil {
		__antithesis_instrumentation__.Notify(478388)
		if b.FilterColIdx != nil {
			__antithesis_instrumentation__.Notify(478389)
			return false
		} else {
			__antithesis_instrumentation__.Notify(478390)
		}
	} else {
		__antithesis_instrumentation__.Notify(478391)
		if b.FilterColIdx == nil || func() bool {
			__antithesis_instrumentation__.Notify(478392)
			return *a.FilterColIdx != *b.FilterColIdx == true
		}() == true {
			__antithesis_instrumentation__.Notify(478393)
			return false
		} else {
			__antithesis_instrumentation__.Notify(478394)
		}
	}
	__antithesis_instrumentation__.Notify(478382)
	if len(a.ColIdx) != len(b.ColIdx) {
		__antithesis_instrumentation__.Notify(478395)
		return false
	} else {
		__antithesis_instrumentation__.Notify(478396)
	}
	__antithesis_instrumentation__.Notify(478383)
	for i, c := range a.ColIdx {
		__antithesis_instrumentation__.Notify(478397)
		if c != b.ColIdx[i] {
			__antithesis_instrumentation__.Notify(478398)
			return false
		} else {
			__antithesis_instrumentation__.Notify(478399)
		}
	}
	__antithesis_instrumentation__.Notify(478384)
	return true
}

func (spec *AggregatorSpec) IsScalar() bool {
	__antithesis_instrumentation__.Notify(478400)
	switch spec.Type {
	case AggregatorSpec_SCALAR:
		__antithesis_instrumentation__.Notify(478401)
		return true
	case AggregatorSpec_NON_SCALAR:
		__antithesis_instrumentation__.Notify(478402)
		return false
	default:
		__antithesis_instrumentation__.Notify(478403)

		return (len(spec.GroupCols) == 0)
	}
}

func (spec *AggregatorSpec) IsRowCount() bool {
	__antithesis_instrumentation__.Notify(478404)
	return len(spec.Aggregations) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(478405)
		return spec.Aggregations[0].FilterColIdx == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(478406)
		return spec.Aggregations[0].Func == CountRows == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(478407)
		return !spec.Aggregations[0].Distinct == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(478408)
		return spec.IsScalar() == true
	}() == true
}

func GetWindowFuncIdx(funcName string) (int32, error) {
	__antithesis_instrumentation__.Notify(478409)
	funcStr := strings.ToUpper(funcName)
	funcIdx, ok := WindowerSpec_WindowFunc_value[funcStr]
	if !ok {
		__antithesis_instrumentation__.Notify(478411)
		return 0, errors.Errorf("unknown window function %s", funcStr)
	} else {
		__antithesis_instrumentation__.Notify(478412)
	}
	__antithesis_instrumentation__.Notify(478410)
	return funcIdx, nil
}

func (spec *WindowerSpec_Frame_Mode) initFromAST(w treewindow.WindowFrameMode) error {
	__antithesis_instrumentation__.Notify(478413)
	switch w {
	case treewindow.RANGE:
		__antithesis_instrumentation__.Notify(478415)
		*spec = WindowerSpec_Frame_RANGE
	case treewindow.ROWS:
		__antithesis_instrumentation__.Notify(478416)
		*spec = WindowerSpec_Frame_ROWS
	case treewindow.GROUPS:
		__antithesis_instrumentation__.Notify(478417)
		*spec = WindowerSpec_Frame_GROUPS
	default:
		__antithesis_instrumentation__.Notify(478418)
		return errors.AssertionFailedf("unexpected WindowFrameMode")
	}
	__antithesis_instrumentation__.Notify(478414)
	return nil
}

func (spec *WindowerSpec_Frame_BoundType) initFromAST(bt treewindow.WindowFrameBoundType) error {
	__antithesis_instrumentation__.Notify(478419)
	switch bt {
	case treewindow.UnboundedPreceding:
		__antithesis_instrumentation__.Notify(478421)
		*spec = WindowerSpec_Frame_UNBOUNDED_PRECEDING
	case treewindow.OffsetPreceding:
		__antithesis_instrumentation__.Notify(478422)
		*spec = WindowerSpec_Frame_OFFSET_PRECEDING
	case treewindow.CurrentRow:
		__antithesis_instrumentation__.Notify(478423)
		*spec = WindowerSpec_Frame_CURRENT_ROW
	case treewindow.OffsetFollowing:
		__antithesis_instrumentation__.Notify(478424)
		*spec = WindowerSpec_Frame_OFFSET_FOLLOWING
	case treewindow.UnboundedFollowing:
		__antithesis_instrumentation__.Notify(478425)
		*spec = WindowerSpec_Frame_UNBOUNDED_FOLLOWING
	default:
		__antithesis_instrumentation__.Notify(478426)
		return errors.AssertionFailedf("unexpected WindowFrameBoundType")
	}
	__antithesis_instrumentation__.Notify(478420)
	return nil
}

func (spec *WindowerSpec_Frame_Exclusion) initFromAST(e treewindow.WindowFrameExclusion) error {
	__antithesis_instrumentation__.Notify(478427)
	switch e {
	case treewindow.NoExclusion:
		__antithesis_instrumentation__.Notify(478429)
		*spec = WindowerSpec_Frame_NO_EXCLUSION
	case treewindow.ExcludeCurrentRow:
		__antithesis_instrumentation__.Notify(478430)
		*spec = WindowerSpec_Frame_EXCLUDE_CURRENT_ROW
	case treewindow.ExcludeGroup:
		__antithesis_instrumentation__.Notify(478431)
		*spec = WindowerSpec_Frame_EXCLUDE_GROUP
	case treewindow.ExcludeTies:
		__antithesis_instrumentation__.Notify(478432)
		*spec = WindowerSpec_Frame_EXCLUDE_TIES
	default:
		__antithesis_instrumentation__.Notify(478433)
		return errors.AssertionFailedf("unexpected WindowerFrameExclusion")
	}
	__antithesis_instrumentation__.Notify(478428)
	return nil
}

func (spec *WindowerSpec_Frame_Bounds) initFromAST(
	b tree.WindowFrameBounds, m treewindow.WindowFrameMode, evalCtx *tree.EvalContext,
) error {
	__antithesis_instrumentation__.Notify(478434)
	if b.StartBound == nil {
		__antithesis_instrumentation__.Notify(478439)
		return errors.Errorf("unexpected: Start Bound is nil")
	} else {
		__antithesis_instrumentation__.Notify(478440)
	}
	__antithesis_instrumentation__.Notify(478435)
	spec.Start = WindowerSpec_Frame_Bound{}
	if err := spec.Start.BoundType.initFromAST(b.StartBound.BoundType); err != nil {
		__antithesis_instrumentation__.Notify(478441)
		return err
	} else {
		__antithesis_instrumentation__.Notify(478442)
	}
	__antithesis_instrumentation__.Notify(478436)
	if b.StartBound.HasOffset() {
		__antithesis_instrumentation__.Notify(478443)
		typedStartOffset := b.StartBound.OffsetExpr.(tree.TypedExpr)
		dStartOffset, err := typedStartOffset.Eval(evalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(478446)
			return err
		} else {
			__antithesis_instrumentation__.Notify(478447)
		}
		__antithesis_instrumentation__.Notify(478444)
		if dStartOffset == tree.DNull {
			__antithesis_instrumentation__.Notify(478448)
			return pgerror.Newf(pgcode.NullValueNotAllowed, "frame starting offset must not be null")
		} else {
			__antithesis_instrumentation__.Notify(478449)
		}
		__antithesis_instrumentation__.Notify(478445)
		switch m {
		case treewindow.ROWS:
			__antithesis_instrumentation__.Notify(478450)
			startOffset := int64(tree.MustBeDInt(dStartOffset))
			if startOffset < 0 {
				__antithesis_instrumentation__.Notify(478458)
				return pgerror.Newf(pgcode.InvalidWindowFrameOffset, "frame starting offset must not be negative")
			} else {
				__antithesis_instrumentation__.Notify(478459)
			}
			__antithesis_instrumentation__.Notify(478451)
			spec.Start.IntOffset = uint64(startOffset)
		case treewindow.RANGE:
			__antithesis_instrumentation__.Notify(478452)
			if isNegative(evalCtx, dStartOffset) {
				__antithesis_instrumentation__.Notify(478460)
				return pgerror.Newf(pgcode.InvalidWindowFrameOffset, "invalid preceding or following size in window function")
			} else {
				__antithesis_instrumentation__.Notify(478461)
			}
			__antithesis_instrumentation__.Notify(478453)
			typ := dStartOffset.ResolvedType()
			spec.Start.OffsetType = DatumInfo{Encoding: descpb.DatumEncoding_VALUE, Type: typ}
			var buf []byte
			var a tree.DatumAlloc
			datum := rowenc.DatumToEncDatum(typ, dStartOffset)
			buf, err = datum.Encode(typ, &a, descpb.DatumEncoding_VALUE, buf)
			if err != nil {
				__antithesis_instrumentation__.Notify(478462)
				return err
			} else {
				__antithesis_instrumentation__.Notify(478463)
			}
			__antithesis_instrumentation__.Notify(478454)
			spec.Start.TypedOffset = buf
		case treewindow.GROUPS:
			__antithesis_instrumentation__.Notify(478455)
			startOffset := int64(tree.MustBeDInt(dStartOffset))
			if startOffset < 0 {
				__antithesis_instrumentation__.Notify(478464)
				return pgerror.Newf(pgcode.InvalidWindowFrameOffset, "frame starting offset must not be negative")
			} else {
				__antithesis_instrumentation__.Notify(478465)
			}
			__antithesis_instrumentation__.Notify(478456)
			spec.Start.IntOffset = uint64(startOffset)
		default:
			__antithesis_instrumentation__.Notify(478457)
		}
	} else {
		__antithesis_instrumentation__.Notify(478466)
	}
	__antithesis_instrumentation__.Notify(478437)

	if b.EndBound != nil {
		__antithesis_instrumentation__.Notify(478467)
		spec.End = &WindowerSpec_Frame_Bound{}
		if err := spec.End.BoundType.initFromAST(b.EndBound.BoundType); err != nil {
			__antithesis_instrumentation__.Notify(478469)
			return err
		} else {
			__antithesis_instrumentation__.Notify(478470)
		}
		__antithesis_instrumentation__.Notify(478468)
		if b.EndBound.HasOffset() {
			__antithesis_instrumentation__.Notify(478471)
			typedEndOffset := b.EndBound.OffsetExpr.(tree.TypedExpr)
			dEndOffset, err := typedEndOffset.Eval(evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(478474)
				return err
			} else {
				__antithesis_instrumentation__.Notify(478475)
			}
			__antithesis_instrumentation__.Notify(478472)
			if dEndOffset == tree.DNull {
				__antithesis_instrumentation__.Notify(478476)
				return pgerror.Newf(pgcode.NullValueNotAllowed, "frame ending offset must not be null")
			} else {
				__antithesis_instrumentation__.Notify(478477)
			}
			__antithesis_instrumentation__.Notify(478473)
			switch m {
			case treewindow.ROWS:
				__antithesis_instrumentation__.Notify(478478)
				endOffset := int64(tree.MustBeDInt(dEndOffset))
				if endOffset < 0 {
					__antithesis_instrumentation__.Notify(478486)
					return pgerror.Newf(pgcode.InvalidWindowFrameOffset, "frame ending offset must not be negative")
				} else {
					__antithesis_instrumentation__.Notify(478487)
				}
				__antithesis_instrumentation__.Notify(478479)
				spec.End.IntOffset = uint64(endOffset)
			case treewindow.RANGE:
				__antithesis_instrumentation__.Notify(478480)
				if isNegative(evalCtx, dEndOffset) {
					__antithesis_instrumentation__.Notify(478488)
					return pgerror.Newf(pgcode.InvalidWindowFrameOffset, "invalid preceding or following size in window function")
				} else {
					__antithesis_instrumentation__.Notify(478489)
				}
				__antithesis_instrumentation__.Notify(478481)
				typ := dEndOffset.ResolvedType()
				spec.End.OffsetType = DatumInfo{Encoding: descpb.DatumEncoding_VALUE, Type: typ}
				var buf []byte
				var a tree.DatumAlloc
				datum := rowenc.DatumToEncDatum(typ, dEndOffset)
				buf, err = datum.Encode(typ, &a, descpb.DatumEncoding_VALUE, buf)
				if err != nil {
					__antithesis_instrumentation__.Notify(478490)
					return err
				} else {
					__antithesis_instrumentation__.Notify(478491)
				}
				__antithesis_instrumentation__.Notify(478482)
				spec.End.TypedOffset = buf
			case treewindow.GROUPS:
				__antithesis_instrumentation__.Notify(478483)
				endOffset := int64(tree.MustBeDInt(dEndOffset))
				if endOffset < 0 {
					__antithesis_instrumentation__.Notify(478492)
					return pgerror.Newf(pgcode.InvalidWindowFrameOffset, "frame ending offset must not be negative")
				} else {
					__antithesis_instrumentation__.Notify(478493)
				}
				__antithesis_instrumentation__.Notify(478484)
				spec.End.IntOffset = uint64(endOffset)
			default:
				__antithesis_instrumentation__.Notify(478485)
			}
		} else {
			__antithesis_instrumentation__.Notify(478494)
		}
	} else {
		__antithesis_instrumentation__.Notify(478495)
	}
	__antithesis_instrumentation__.Notify(478438)

	return nil
}

func isNegative(evalCtx *tree.EvalContext, offset tree.Datum) bool {
	__antithesis_instrumentation__.Notify(478496)
	switch o := offset.(type) {
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(478497)
		return *o < 0
	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(478498)
		return o.Negative
	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(478499)
		return *o < 0
	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(478500)
		return o.Compare(evalCtx, &tree.DInterval{Duration: duration.Duration{}}) < 0
	default:
		__antithesis_instrumentation__.Notify(478501)
		panic("unexpected offset type")
	}
}

func (spec *WindowerSpec_Frame) InitFromAST(f *tree.WindowFrame, evalCtx *tree.EvalContext) error {
	__antithesis_instrumentation__.Notify(478502)
	if err := spec.Mode.initFromAST(f.Mode); err != nil {
		__antithesis_instrumentation__.Notify(478505)
		return err
	} else {
		__antithesis_instrumentation__.Notify(478506)
	}
	__antithesis_instrumentation__.Notify(478503)
	if err := spec.Exclusion.initFromAST(f.Exclusion); err != nil {
		__antithesis_instrumentation__.Notify(478507)
		return err
	} else {
		__antithesis_instrumentation__.Notify(478508)
	}
	__antithesis_instrumentation__.Notify(478504)
	return spec.Bounds.initFromAST(f.Bounds, f.Mode, evalCtx)
}

func (spec WindowerSpec_Frame_Mode) convertToAST() (treewindow.WindowFrameMode, error) {
	__antithesis_instrumentation__.Notify(478509)
	switch spec {
	case WindowerSpec_Frame_RANGE:
		__antithesis_instrumentation__.Notify(478510)
		return treewindow.RANGE, nil
	case WindowerSpec_Frame_ROWS:
		__antithesis_instrumentation__.Notify(478511)
		return treewindow.ROWS, nil
	case WindowerSpec_Frame_GROUPS:
		__antithesis_instrumentation__.Notify(478512)
		return treewindow.GROUPS, nil
	default:
		__antithesis_instrumentation__.Notify(478513)
		return treewindow.WindowFrameMode(0), errors.AssertionFailedf("unexpected WindowerSpec_Frame_Mode")
	}
}

func (spec WindowerSpec_Frame_BoundType) convertToAST() (treewindow.WindowFrameBoundType, error) {
	__antithesis_instrumentation__.Notify(478514)
	switch spec {
	case WindowerSpec_Frame_UNBOUNDED_PRECEDING:
		__antithesis_instrumentation__.Notify(478515)
		return treewindow.UnboundedPreceding, nil
	case WindowerSpec_Frame_OFFSET_PRECEDING:
		__antithesis_instrumentation__.Notify(478516)
		return treewindow.OffsetPreceding, nil
	case WindowerSpec_Frame_CURRENT_ROW:
		__antithesis_instrumentation__.Notify(478517)
		return treewindow.CurrentRow, nil
	case WindowerSpec_Frame_OFFSET_FOLLOWING:
		__antithesis_instrumentation__.Notify(478518)
		return treewindow.OffsetFollowing, nil
	case WindowerSpec_Frame_UNBOUNDED_FOLLOWING:
		__antithesis_instrumentation__.Notify(478519)
		return treewindow.UnboundedFollowing, nil
	default:
		__antithesis_instrumentation__.Notify(478520)
		return treewindow.WindowFrameBoundType(0), errors.AssertionFailedf("unexpected WindowerSpec_Frame_BoundType")
	}
}

func (spec WindowerSpec_Frame_Exclusion) convertToAST() (treewindow.WindowFrameExclusion, error) {
	__antithesis_instrumentation__.Notify(478521)
	switch spec {
	case WindowerSpec_Frame_NO_EXCLUSION:
		__antithesis_instrumentation__.Notify(478522)
		return treewindow.NoExclusion, nil
	case WindowerSpec_Frame_EXCLUDE_CURRENT_ROW:
		__antithesis_instrumentation__.Notify(478523)
		return treewindow.ExcludeCurrentRow, nil
	case WindowerSpec_Frame_EXCLUDE_GROUP:
		__antithesis_instrumentation__.Notify(478524)
		return treewindow.ExcludeGroup, nil
	case WindowerSpec_Frame_EXCLUDE_TIES:
		__antithesis_instrumentation__.Notify(478525)
		return treewindow.ExcludeTies, nil
	default:
		__antithesis_instrumentation__.Notify(478526)
		return treewindow.WindowFrameExclusion(0), errors.AssertionFailedf("unexpected WindowerSpec_Frame_Exclusion")
	}
}

func (spec WindowerSpec_Frame_Bounds) convertToAST() (tree.WindowFrameBounds, error) {
	__antithesis_instrumentation__.Notify(478527)
	bounds := tree.WindowFrameBounds{}
	startBoundType, err := spec.Start.BoundType.convertToAST()
	if err != nil {
		__antithesis_instrumentation__.Notify(478530)
		return bounds, err
	} else {
		__antithesis_instrumentation__.Notify(478531)
	}
	__antithesis_instrumentation__.Notify(478528)
	bounds.StartBound = &tree.WindowFrameBound{
		BoundType: startBoundType,
	}

	if spec.End != nil {
		__antithesis_instrumentation__.Notify(478532)
		endBoundType, err := spec.End.BoundType.convertToAST()
		if err != nil {
			__antithesis_instrumentation__.Notify(478534)
			return bounds, err
		} else {
			__antithesis_instrumentation__.Notify(478535)
		}
		__antithesis_instrumentation__.Notify(478533)
		bounds.EndBound = &tree.WindowFrameBound{BoundType: endBoundType}
	} else {
		__antithesis_instrumentation__.Notify(478536)
	}
	__antithesis_instrumentation__.Notify(478529)
	return bounds, nil
}

func (spec *WindowerSpec_Frame) ConvertToAST() (*tree.WindowFrame, error) {
	__antithesis_instrumentation__.Notify(478537)
	mode, err := spec.Mode.convertToAST()
	if err != nil {
		__antithesis_instrumentation__.Notify(478541)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(478542)
	}
	__antithesis_instrumentation__.Notify(478538)
	bounds, err := spec.Bounds.convertToAST()
	if err != nil {
		__antithesis_instrumentation__.Notify(478543)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(478544)
	}
	__antithesis_instrumentation__.Notify(478539)
	exclusion, err := spec.Exclusion.convertToAST()
	if err != nil {
		__antithesis_instrumentation__.Notify(478545)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(478546)
	}
	__antithesis_instrumentation__.Notify(478540)
	return &tree.WindowFrame{
		Mode:      mode,
		Bounds:    bounds,
		Exclusion: exclusion,
	}, nil
}

func (spec *JoinReaderSpec) IsIndexJoin() bool {
	__antithesis_instrumentation__.Notify(478547)
	return len(spec.LookupColumns) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(478548)
		return spec.LookupExpr.Empty() == true
	}() == true
}
