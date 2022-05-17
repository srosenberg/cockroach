package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type IndexedVarContainer interface {
	IndexedVarEval(idx int, ctx *EvalContext) (Datum, error)
	IndexedVarResolvedType(idx int) *types.T

	IndexedVarNodeFormatter(idx int) NodeFormatter
}

type IndexedVar struct {
	Idx  int
	Used bool

	col NodeFormatter

	typeAnnotation
}

var _ TypedExpr = &IndexedVar{}

func (*IndexedVar) Variable() { __antithesis_instrumentation__.Notify(610011) }

func (v *IndexedVar) Walk(_ Visitor) Expr {
	__antithesis_instrumentation__.Notify(610012)
	return v
}

func (v *IndexedVar) TypeCheck(
	_ context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(610013)
	if semaCtx.IVarContainer == nil || func() bool {
		__antithesis_instrumentation__.Notify(610015)
		return semaCtx.IVarContainer == unboundContainer == true
	}() == true {
		__antithesis_instrumentation__.Notify(610016)

		return nil, pgerror.Newf(
			pgcode.UndefinedColumn, "column reference @%d not allowed in this context", v.Idx+1)
	} else {
		__antithesis_instrumentation__.Notify(610017)
	}
	__antithesis_instrumentation__.Notify(610014)
	v.typ = semaCtx.IVarContainer.IndexedVarResolvedType(v.Idx)
	return v, nil
}

func (v *IndexedVar) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(610018)
	if ctx.IVarContainer == nil || func() bool {
		__antithesis_instrumentation__.Notify(610020)
		return ctx.IVarContainer == unboundContainer == true
	}() == true {
		__antithesis_instrumentation__.Notify(610021)
		return nil, errors.AssertionFailedf(
			"indexed var must be bound to a container before evaluation")
	} else {
		__antithesis_instrumentation__.Notify(610022)
	}
	__antithesis_instrumentation__.Notify(610019)
	return ctx.IVarContainer.IndexedVarEval(v.Idx, ctx)
}

func (v *IndexedVar) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(610023)
	if v.typ == nil {
		__antithesis_instrumentation__.Notify(610025)
		panic(errors.AssertionFailedf("indexed var must be type checked first"))
	} else {
		__antithesis_instrumentation__.Notify(610026)
	}
	__antithesis_instrumentation__.Notify(610024)
	return v.typ
}

func (v *IndexedVar) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(610027)
	f := ctx.flags
	if ctx.indexedVarFormat != nil {
		__antithesis_instrumentation__.Notify(610028)
		ctx.indexedVarFormat(ctx, v.Idx)
	} else {
		__antithesis_instrumentation__.Notify(610029)
		if f.HasFlags(fmtSymbolicVars) || func() bool {
			__antithesis_instrumentation__.Notify(610030)
			return v.col == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(610031)
			ctx.Printf("@%d", v.Idx+1)
		} else {
			__antithesis_instrumentation__.Notify(610032)
			ctx.FormatNode(v.col)
		}
	}
}

func NewOrdinalReference(r int) *IndexedVar {
	__antithesis_instrumentation__.Notify(610033)
	return &IndexedVar{Idx: r}
}

func NewTypedOrdinalReference(r int, typ *types.T) *IndexedVar {
	__antithesis_instrumentation__.Notify(610034)
	return &IndexedVar{Idx: r, typeAnnotation: typeAnnotation{typ: typ}}
}

type IndexedVarHelper struct {
	vars      []IndexedVar
	container IndexedVarContainer
}

func (h *IndexedVarHelper) Container() IndexedVarContainer {
	__antithesis_instrumentation__.Notify(610035)
	return h.container
}

func (h *IndexedVarHelper) BindIfUnbound(ivar *IndexedVar) (*IndexedVar, error) {
	__antithesis_instrumentation__.Notify(610036)

	if ivar.Idx < 0 || func() bool {
		__antithesis_instrumentation__.Notify(610039)
		return ivar.Idx >= len(h.vars) == true
	}() == true {
		__antithesis_instrumentation__.Notify(610040)
		return ivar, pgerror.Newf(
			pgcode.UndefinedColumn, "invalid column ordinal: @%d", ivar.Idx+1)
	} else {
		__antithesis_instrumentation__.Notify(610041)
	}
	__antithesis_instrumentation__.Notify(610037)

	if !ivar.Used {
		__antithesis_instrumentation__.Notify(610042)
		return h.IndexedVar(ivar.Idx), nil
	} else {
		__antithesis_instrumentation__.Notify(610043)
	}
	__antithesis_instrumentation__.Notify(610038)
	return ivar, nil
}

func MakeIndexedVarHelper(container IndexedVarContainer, numVars int) IndexedVarHelper {
	__antithesis_instrumentation__.Notify(610044)
	return IndexedVarHelper{
		vars:      make([]IndexedVar, numVars),
		container: container,
	}
}

func (h *IndexedVarHelper) AppendSlot() int {
	__antithesis_instrumentation__.Notify(610045)
	h.vars = append(h.vars, IndexedVar{})
	return len(h.vars) - 1
}

func (h *IndexedVarHelper) checkIndex(idx int) {
	__antithesis_instrumentation__.Notify(610046)
	if idx < 0 || func() bool {
		__antithesis_instrumentation__.Notify(610047)
		return idx >= len(h.vars) == true
	}() == true {
		__antithesis_instrumentation__.Notify(610048)
		panic(errors.AssertionFailedf(
			"invalid var index %d (columns: %d)", redact.Safe(idx), redact.Safe(len(h.vars))))
	} else {
		__antithesis_instrumentation__.Notify(610049)
	}
}

func (h *IndexedVarHelper) IndexedVar(idx int) *IndexedVar {
	__antithesis_instrumentation__.Notify(610050)
	h.checkIndex(idx)
	v := &h.vars[idx]
	v.Idx = idx
	v.Used = true
	v.typ = h.container.IndexedVarResolvedType(idx)
	v.col = h.container.IndexedVarNodeFormatter(idx)
	return v
}

func (h *IndexedVarHelper) IndexedVarWithType(idx int, typ *types.T) *IndexedVar {
	__antithesis_instrumentation__.Notify(610051)
	h.checkIndex(idx)
	v := &h.vars[idx]
	v.Idx = idx
	v.Used = true
	v.typ = typ
	return v
}

func (h *IndexedVarHelper) IndexedVarUsed(idx int) bool {
	__antithesis_instrumentation__.Notify(610052)
	h.checkIndex(idx)
	return h.vars[idx].Used
}

func (h *IndexedVarHelper) GetIndexedVars() []IndexedVar {
	__antithesis_instrumentation__.Notify(610053)
	return h.vars
}

func (h *IndexedVarHelper) Rebind(expr TypedExpr) TypedExpr {
	__antithesis_instrumentation__.Notify(610054)
	if expr == nil {
		__antithesis_instrumentation__.Notify(610056)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(610057)
	}
	__antithesis_instrumentation__.Notify(610055)
	ret, _ := WalkExpr(h, expr)
	return ret.(TypedExpr)
}

var _ Visitor = &IndexedVarHelper{}

func (h *IndexedVarHelper) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(610058)
	if iv, ok := expr.(*IndexedVar); ok {
		__antithesis_instrumentation__.Notify(610060)
		return false, h.IndexedVar(iv.Idx)
	} else {
		__antithesis_instrumentation__.Notify(610061)
	}
	__antithesis_instrumentation__.Notify(610059)
	return true, expr
}

func (*IndexedVarHelper) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(610062)
	return expr
}

type unboundContainerType struct{}

var unboundContainer = &unboundContainerType{}

func (*unboundContainerType) IndexedVarEval(idx int, _ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(610063)
	return nil, errors.AssertionFailedf("unbound ordinal reference @%d", redact.Safe(idx+1))
}

func (*unboundContainerType) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(610064)
	panic(errors.AssertionFailedf("unbound ordinal reference @%d", redact.Safe(idx+1)))
}

func (*unboundContainerType) IndexedVarNodeFormatter(idx int) NodeFormatter {
	__antithesis_instrumentation__.Notify(610065)
	panic(errors.AssertionFailedf("unbound ordinal reference @%d", redact.Safe(idx+1)))
}

type typeContainer struct {
	types []*types.T
}

var _ IndexedVarContainer = &typeContainer{}

func (tc *typeContainer) IndexedVarEval(idx int, ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(610066)
	return nil, errors.AssertionFailedf("no eval allowed in typeContainer")
}

func (tc *typeContainer) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(610067)
	return tc.types[idx]
}

func (tc *typeContainer) IndexedVarNodeFormatter(idx int) NodeFormatter {
	__antithesis_instrumentation__.Notify(610068)
	return nil
}

func MakeTypesOnlyIndexedVarHelper(types []*types.T) IndexedVarHelper {
	__antithesis_instrumentation__.Notify(610069)
	c := &typeContainer{types: types}
	return MakeIndexedVarHelper(c, len(types))
}
