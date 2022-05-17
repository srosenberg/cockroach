package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type Expr interface {
	fmt.Stringer
	NodeFormatter

	Walk(Visitor) Expr

	TypeCheck(ctx context.Context, semaCtx *SemaContext, desired *types.T) (TypedExpr, error)
}

type TypedExpr interface {
	Expr

	Eval(*EvalContext) (Datum, error)

	ResolvedType() *types.T
}

type VariableExpr interface {
	Expr
	Variable()
}

var _ VariableExpr = &IndexedVar{}
var _ VariableExpr = &Subquery{}
var _ VariableExpr = UnqualifiedStar{}
var _ VariableExpr = &UnresolvedName{}
var _ VariableExpr = &AllColumnsSelector{}
var _ VariableExpr = &ColumnItem{}

type operatorExpr interface {
	Expr
	operatorExpr()
}

var _ operatorExpr = &AndExpr{}
var _ operatorExpr = &OrExpr{}
var _ operatorExpr = &NotExpr{}
var _ operatorExpr = &IsNullExpr{}
var _ operatorExpr = &IsNotNullExpr{}
var _ operatorExpr = &BinaryExpr{}
var _ operatorExpr = &UnaryExpr{}
var _ operatorExpr = &ComparisonExpr{}
var _ operatorExpr = &RangeCond{}
var _ operatorExpr = &IsOfTypeExpr{}

type Operator interface {
	Operator()
}

var _ Operator = (*UnaryOperator)(nil)
var _ Operator = (*treebin.BinaryOperator)(nil)
var _ Operator = (*treecmp.ComparisonOperator)(nil)

type SubqueryExpr interface {
	Expr
	SubqueryExpr()
}

var _ SubqueryExpr = &Subquery{}

func exprFmtWithParen(ctx *FmtCtx, e Expr) {
	__antithesis_instrumentation__.Notify(609333)
	if _, ok := e.(operatorExpr); ok {
		__antithesis_instrumentation__.Notify(609334)
		ctx.WriteByte('(')
		ctx.FormatNode(e)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(609335)
		ctx.FormatNode(e)
	}
}

type typeAnnotation struct {
	typ *types.T
}

func (ta typeAnnotation) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(609336)
	ta.assertTyped()
	return ta.typ
}

func (ta typeAnnotation) assertTyped() {
	__antithesis_instrumentation__.Notify(609337)
	if ta.typ == nil {
		__antithesis_instrumentation__.Notify(609338)
		panic(errors.AssertionFailedf(
			"ReturnType called on TypedExpr with empty typeAnnotation. " +
				"Was the underlying Expr type-checked before asserting a type of TypedExpr?"))
	} else {
		__antithesis_instrumentation__.Notify(609339)
	}
}

type AndExpr struct {
	Left, Right Expr

	typeAnnotation
}

func (*AndExpr) operatorExpr() { __antithesis_instrumentation__.Notify(609340) }

func binExprFmtWithParen(ctx *FmtCtx, e1 Expr, op string, e2 Expr, pad bool) {
	__antithesis_instrumentation__.Notify(609341)
	exprFmtWithParen(ctx, e1)
	if pad {
		__antithesis_instrumentation__.Notify(609344)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(609345)
	}
	__antithesis_instrumentation__.Notify(609342)
	ctx.WriteString(op)
	if pad {
		__antithesis_instrumentation__.Notify(609346)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(609347)
	}
	__antithesis_instrumentation__.Notify(609343)
	exprFmtWithParen(ctx, e2)
}

func binExprFmtWithParenAndSubOp(ctx *FmtCtx, e1 Expr, subOp, op string, e2 Expr) {
	__antithesis_instrumentation__.Notify(609348)
	exprFmtWithParen(ctx, e1)
	ctx.WriteByte(' ')
	if subOp != "" {
		__antithesis_instrumentation__.Notify(609350)
		ctx.WriteString(subOp)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(609351)
	}
	__antithesis_instrumentation__.Notify(609349)
	ctx.WriteString(op)
	ctx.WriteByte(' ')
	exprFmtWithParen(ctx, e2)
}

func (node *AndExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609352)
	binExprFmtWithParen(ctx, node.Left, "AND", node.Right, true)
}

func NewTypedAndExpr(left, right TypedExpr) *AndExpr {
	__antithesis_instrumentation__.Notify(609353)
	node := &AndExpr{Left: left, Right: right}
	node.typ = types.Bool
	return node
}

func (node *AndExpr) TypedLeft() TypedExpr {
	__antithesis_instrumentation__.Notify(609354)
	return node.Left.(TypedExpr)
}

func (node *AndExpr) TypedRight() TypedExpr {
	__antithesis_instrumentation__.Notify(609355)
	return node.Right.(TypedExpr)
}

type OrExpr struct {
	Left, Right Expr

	typeAnnotation
}

func (*OrExpr) operatorExpr() { __antithesis_instrumentation__.Notify(609356) }

func (node *OrExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609357)
	binExprFmtWithParen(ctx, node.Left, "OR", node.Right, true)
}

func NewTypedOrExpr(left, right TypedExpr) *OrExpr {
	__antithesis_instrumentation__.Notify(609358)
	node := &OrExpr{Left: left, Right: right}
	node.typ = types.Bool
	return node
}

func (node *OrExpr) TypedLeft() TypedExpr {
	__antithesis_instrumentation__.Notify(609359)
	return node.Left.(TypedExpr)
}

func (node *OrExpr) TypedRight() TypedExpr {
	__antithesis_instrumentation__.Notify(609360)
	return node.Right.(TypedExpr)
}

type NotExpr struct {
	Expr Expr

	typeAnnotation
}

func (*NotExpr) operatorExpr() { __antithesis_instrumentation__.Notify(609361) }

func (node *NotExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609362)
	ctx.WriteString("NOT ")
	exprFmtWithParen(ctx, node.Expr)
}

func NewTypedNotExpr(expr TypedExpr) *NotExpr {
	__antithesis_instrumentation__.Notify(609363)
	node := &NotExpr{Expr: expr}
	node.typ = types.Bool
	return node
}

func (node *NotExpr) TypedInnerExpr() TypedExpr {
	__antithesis_instrumentation__.Notify(609364)
	return node.Expr.(TypedExpr)
}

type IsNullExpr struct {
	Expr Expr

	typeAnnotation
}

func (*IsNullExpr) operatorExpr() { __antithesis_instrumentation__.Notify(609365) }

func (node *IsNullExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609366)
	exprFmtWithParen(ctx, node.Expr)
	ctx.WriteString(" IS NULL")
}

func NewTypedIsNullExpr(expr TypedExpr) *IsNullExpr {
	__antithesis_instrumentation__.Notify(609367)
	node := &IsNullExpr{Expr: expr}
	node.typ = types.Bool
	return node
}

func (node *IsNullExpr) TypedInnerExpr() TypedExpr {
	__antithesis_instrumentation__.Notify(609368)
	return node.Expr.(TypedExpr)
}

type IsNotNullExpr struct {
	Expr Expr

	typeAnnotation
}

func (*IsNotNullExpr) operatorExpr() { __antithesis_instrumentation__.Notify(609369) }

func (node *IsNotNullExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609370)
	exprFmtWithParen(ctx, node.Expr)
	ctx.WriteString(" IS NOT NULL")
}

func NewTypedIsNotNullExpr(expr TypedExpr) *IsNotNullExpr {
	__antithesis_instrumentation__.Notify(609371)
	node := &IsNotNullExpr{Expr: expr}
	node.typ = types.Bool
	return node
}

func (node *IsNotNullExpr) TypedInnerExpr() TypedExpr {
	__antithesis_instrumentation__.Notify(609372)
	return node.Expr.(TypedExpr)
}

type ParenExpr struct {
	Expr Expr

	typeAnnotation
}

func (node *ParenExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609373)
	ctx.WriteByte('(')
	ctx.FormatNode(node.Expr)
	ctx.WriteByte(')')
}

func (node *ParenExpr) TypedInnerExpr() TypedExpr {
	__antithesis_instrumentation__.Notify(609374)
	return node.Expr.(TypedExpr)
}

func StripParens(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(609375)
	if p, ok := expr.(*ParenExpr); ok {
		__antithesis_instrumentation__.Notify(609377)
		return StripParens(p.Expr)
	} else {
		__antithesis_instrumentation__.Notify(609378)
	}
	__antithesis_instrumentation__.Notify(609376)
	return expr
}

type ComparisonExpr struct {
	Operator    treecmp.ComparisonOperator
	SubOperator treecmp.ComparisonOperator
	Left, Right Expr

	typeAnnotation
	Fn *CmpOp
}

func (*ComparisonExpr) operatorExpr() { __antithesis_instrumentation__.Notify(609379) }

func (node *ComparisonExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609380)
	opStr := node.Operator.String()

	if !ctx.HasFlags(FmtHideConstants) {
		__antithesis_instrumentation__.Notify(609382)
		if node.Operator.Symbol == treecmp.IsDistinctFrom && func() bool {
			__antithesis_instrumentation__.Notify(609383)
			return (node.Right == DBoolTrue || func() bool {
				__antithesis_instrumentation__.Notify(609384)
				return node.Right == DBoolFalse == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(609385)
			opStr = "IS NOT"
		} else {
			__antithesis_instrumentation__.Notify(609386)
			if node.Operator.Symbol == treecmp.IsNotDistinctFrom && func() bool {
				__antithesis_instrumentation__.Notify(609387)
				return (node.Right == DBoolTrue || func() bool {
					__antithesis_instrumentation__.Notify(609388)
					return node.Right == DBoolFalse == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(609389)
				opStr = "IS"
			} else {
				__antithesis_instrumentation__.Notify(609390)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(609391)
	}
	__antithesis_instrumentation__.Notify(609381)
	if node.Operator.Symbol.HasSubOperator() {
		__antithesis_instrumentation__.Notify(609392)
		binExprFmtWithParenAndSubOp(ctx, node.Left, node.SubOperator.String(), opStr, node.Right)
	} else {
		__antithesis_instrumentation__.Notify(609393)
		binExprFmtWithParen(ctx, node.Left, opStr, node.Right, true)
	}
}

func NewTypedComparisonExpr(op treecmp.ComparisonOperator, left, right TypedExpr) *ComparisonExpr {
	__antithesis_instrumentation__.Notify(609394)
	node := &ComparisonExpr{Operator: op, Left: left, Right: right}
	node.typ = types.Bool
	node.memoizeFn()
	return node
}

func NewTypedComparisonExprWithSubOp(
	op, subOp treecmp.ComparisonOperator, left, right TypedExpr,
) *ComparisonExpr {
	__antithesis_instrumentation__.Notify(609395)
	node := &ComparisonExpr{Operator: op, SubOperator: subOp, Left: left, Right: right}
	node.typ = types.Bool
	node.memoizeFn()
	return node
}

func NewTypedIndirectionExpr(expr, index TypedExpr, typ *types.T) *IndirectionExpr {
	__antithesis_instrumentation__.Notify(609396)
	node := &IndirectionExpr{
		Expr:        expr,
		Indirection: ArraySubscripts{&ArraySubscript{Begin: index}},
	}
	node.typ = typ
	return node
}

func NewTypedCollateExpr(expr TypedExpr, locale string) *CollateExpr {
	__antithesis_instrumentation__.Notify(609397)
	node := &CollateExpr{
		Expr:   expr,
		Locale: locale,
	}
	node.typ = types.MakeCollatedString(types.String, locale)
	return node
}

func NewTypedArrayFlattenExpr(input Expr) *ArrayFlatten {
	__antithesis_instrumentation__.Notify(609398)
	inputTyp := input.(TypedExpr).ResolvedType()
	node := &ArrayFlatten{
		Subquery: input,
	}
	node.typ = types.MakeArray(inputTyp)
	return node
}

func NewTypedIfErrExpr(cond, orElse, errCode TypedExpr) *IfErrExpr {
	__antithesis_instrumentation__.Notify(609399)
	node := &IfErrExpr{
		Cond:    cond,
		Else:    orElse,
		ErrCode: errCode,
	}
	if orElse == nil {
		__antithesis_instrumentation__.Notify(609401)
		node.typ = types.Bool
	} else {
		__antithesis_instrumentation__.Notify(609402)
		node.typ = cond.ResolvedType()
	}
	__antithesis_instrumentation__.Notify(609400)
	return node
}

func (node *ComparisonExpr) memoizeFn() {
	__antithesis_instrumentation__.Notify(609403)
	fOp, fLeft, fRight, _, _ := FoldComparisonExpr(node.Operator, node.Left, node.Right)
	leftRet, rightRet := fLeft.(TypedExpr).ResolvedType(), fRight.(TypedExpr).ResolvedType()
	switch node.Operator.Symbol {
	case treecmp.Any, treecmp.Some, treecmp.All:
		__antithesis_instrumentation__.Notify(609406)

		fOp, _, _, _, _ = FoldComparisonExpr(node.SubOperator, nil, nil)

		switch rightRet.Family() {
		case types.ArrayFamily:
			__antithesis_instrumentation__.Notify(609408)

			rightRet = rightRet.ArrayContents()
		case types.TupleFamily:
			__antithesis_instrumentation__.Notify(609409)

			if len(rightRet.TupleContents()) > 0 {
				__antithesis_instrumentation__.Notify(609411)
				rightRet = rightRet.TupleContents()[0]
			} else {
				__antithesis_instrumentation__.Notify(609412)
				rightRet = leftRet
			}
		default:
			__antithesis_instrumentation__.Notify(609410)
		}
	default:
		__antithesis_instrumentation__.Notify(609407)
	}
	__antithesis_instrumentation__.Notify(609404)

	fn, ok := CmpOps[fOp.Symbol].LookupImpl(leftRet, rightRet)
	if !ok {
		__antithesis_instrumentation__.Notify(609413)
		panic(errors.AssertionFailedf("lookup for ComparisonExpr %s's CmpOp failed",
			AsStringWithFlags(node, FmtShowTypes)))
	} else {
		__antithesis_instrumentation__.Notify(609414)
	}
	__antithesis_instrumentation__.Notify(609405)
	node.Fn = fn
}

func (node *ComparisonExpr) TypedLeft() TypedExpr {
	__antithesis_instrumentation__.Notify(609415)
	return node.Left.(TypedExpr)
}

func (node *ComparisonExpr) TypedRight() TypedExpr {
	__antithesis_instrumentation__.Notify(609416)
	return node.Right.(TypedExpr)
}

type RangeCond struct {
	Not       bool
	Symmetric bool
	Left      Expr
	From, To  Expr

	leftTo TypedExpr

	typeAnnotation
}

func (*RangeCond) operatorExpr() { __antithesis_instrumentation__.Notify(609417) }

func (node *RangeCond) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609418)
	notStr := " BETWEEN "
	if node.Not {
		__antithesis_instrumentation__.Notify(609421)
		notStr = " NOT BETWEEN "
	} else {
		__antithesis_instrumentation__.Notify(609422)
	}
	__antithesis_instrumentation__.Notify(609419)
	exprFmtWithParen(ctx, node.Left)
	ctx.WriteString(notStr)
	if node.Symmetric {
		__antithesis_instrumentation__.Notify(609423)
		ctx.WriteString("SYMMETRIC ")
	} else {
		__antithesis_instrumentation__.Notify(609424)
	}
	__antithesis_instrumentation__.Notify(609420)
	binExprFmtWithParen(ctx, node.From, "AND", node.To, true)
}

func (node *RangeCond) TypedLeftFrom() TypedExpr {
	__antithesis_instrumentation__.Notify(609425)
	return node.Left.(TypedExpr)
}

func (node *RangeCond) TypedFrom() TypedExpr {
	__antithesis_instrumentation__.Notify(609426)
	return node.From.(TypedExpr)
}

func (node *RangeCond) TypedLeftTo() TypedExpr {
	__antithesis_instrumentation__.Notify(609427)
	return node.leftTo
}

func (node *RangeCond) TypedTo() TypedExpr {
	__antithesis_instrumentation__.Notify(609428)
	return node.To.(TypedExpr)
}

type IsOfTypeExpr struct {
	Not   bool
	Expr  Expr
	Types []ResolvableTypeReference

	resolvedTypes []*types.T

	typeAnnotation
}

func (*IsOfTypeExpr) operatorExpr() { __antithesis_instrumentation__.Notify(609429) }

func (node *IsOfTypeExpr) ResolvedTypes() []*types.T {
	__antithesis_instrumentation__.Notify(609430)
	if node.resolvedTypes == nil {
		__antithesis_instrumentation__.Notify(609432)
		panic("ResolvedTypes called on an IsOfTypeExpr before typechecking")
	} else {
		__antithesis_instrumentation__.Notify(609433)
	}
	__antithesis_instrumentation__.Notify(609431)
	return node.resolvedTypes
}

func (node *IsOfTypeExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609434)
	exprFmtWithParen(ctx, node.Expr)
	ctx.WriteString(" IS")
	if node.Not {
		__antithesis_instrumentation__.Notify(609437)
		ctx.WriteString(" NOT")
	} else {
		__antithesis_instrumentation__.Notify(609438)
	}
	__antithesis_instrumentation__.Notify(609435)
	ctx.WriteString(" OF (")
	for i, t := range node.Types {
		__antithesis_instrumentation__.Notify(609439)
		if i > 0 {
			__antithesis_instrumentation__.Notify(609441)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(609442)
		}
		__antithesis_instrumentation__.Notify(609440)
		ctx.FormatTypeReference(t)
	}
	__antithesis_instrumentation__.Notify(609436)
	ctx.WriteByte(')')
}

type IfErrExpr struct {
	Cond    Expr
	Else    Expr
	ErrCode Expr

	typeAnnotation
}

func (node *IfErrExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609443)
	if node.Else != nil {
		__antithesis_instrumentation__.Notify(609447)
		ctx.WriteString("IFERROR(")
	} else {
		__antithesis_instrumentation__.Notify(609448)
		ctx.WriteString("ISERROR(")
	}
	__antithesis_instrumentation__.Notify(609444)
	ctx.FormatNode(node.Cond)
	if node.Else != nil {
		__antithesis_instrumentation__.Notify(609449)
		ctx.WriteString(", ")
		ctx.FormatNode(node.Else)
	} else {
		__antithesis_instrumentation__.Notify(609450)
	}
	__antithesis_instrumentation__.Notify(609445)
	if node.ErrCode != nil {
		__antithesis_instrumentation__.Notify(609451)
		ctx.WriteString(", ")
		ctx.FormatNode(node.ErrCode)
	} else {
		__antithesis_instrumentation__.Notify(609452)
	}
	__antithesis_instrumentation__.Notify(609446)
	ctx.WriteByte(')')
}

type IfExpr struct {
	Cond Expr
	True Expr
	Else Expr

	typeAnnotation
}

func (node *IfExpr) TypedTrueExpr() TypedExpr {
	__antithesis_instrumentation__.Notify(609453)
	return node.True.(TypedExpr)
}

func (node *IfExpr) TypedCondExpr() TypedExpr {
	__antithesis_instrumentation__.Notify(609454)
	return node.Cond.(TypedExpr)
}

func (node *IfExpr) TypedElseExpr() TypedExpr {
	__antithesis_instrumentation__.Notify(609455)
	return node.Else.(TypedExpr)
}

func (node *IfExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609456)
	ctx.WriteString("IF(")
	ctx.FormatNode(node.Cond)
	ctx.WriteString(", ")
	ctx.FormatNode(node.True)
	ctx.WriteString(", ")
	ctx.FormatNode(node.Else)
	ctx.WriteByte(')')
}

type NullIfExpr struct {
	Expr1 Expr
	Expr2 Expr

	typeAnnotation
}

func (node *NullIfExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609457)
	ctx.WriteString("NULLIF(")
	ctx.FormatNode(node.Expr1)
	ctx.WriteString(", ")
	ctx.FormatNode(node.Expr2)
	ctx.WriteByte(')')
}

type CoalesceExpr struct {
	Name  string
	Exprs Exprs

	typeAnnotation
}

func NewTypedCoalesceExpr(typedExprs TypedExprs, typ *types.T) *CoalesceExpr {
	__antithesis_instrumentation__.Notify(609458)
	c := &CoalesceExpr{
		Name:  "COALESCE",
		Exprs: make(Exprs, len(typedExprs)),
	}
	for i := range typedExprs {
		__antithesis_instrumentation__.Notify(609460)
		c.Exprs[i] = typedExprs[i]
	}
	__antithesis_instrumentation__.Notify(609459)
	c.typ = typ
	return c
}

func NewTypedArray(typedExprs TypedExprs, typ *types.T) *Array {
	__antithesis_instrumentation__.Notify(609461)
	c := &Array{
		Exprs: make(Exprs, len(typedExprs)),
	}
	for i := range typedExprs {
		__antithesis_instrumentation__.Notify(609463)
		c.Exprs[i] = typedExprs[i]
	}
	__antithesis_instrumentation__.Notify(609462)
	c.typ = typ
	return c
}

func (node *CoalesceExpr) TypedExprAt(idx int) TypedExpr {
	__antithesis_instrumentation__.Notify(609464)
	return node.Exprs[idx].(TypedExpr)
}

func (node *CoalesceExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609465)
	ctx.WriteString(node.Name)
	ctx.WriteByte('(')
	ctx.FormatNode(&node.Exprs)
	ctx.WriteByte(')')
}

type DefaultVal struct{}

func (node DefaultVal) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609466)
	ctx.WriteString("DEFAULT")
}

func (DefaultVal) ResolvedType() *types.T { __antithesis_instrumentation__.Notify(609467); return nil }

type PartitionMaxVal struct{}

func (node PartitionMaxVal) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609468)
	ctx.WriteString("MAXVALUE")
}

type PartitionMinVal struct{}

func (node PartitionMinVal) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609469)
	ctx.WriteString("MINVALUE")
}

type Placeholder struct {
	Idx PlaceholderIdx

	typeAnnotation
}

func NewPlaceholder(name string) (*Placeholder, error) {
	__antithesis_instrumentation__.Notify(609470)
	uval, err := strconv.ParseUint(name, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(609473)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(609474)
	}
	__antithesis_instrumentation__.Notify(609471)

	if uval == 0 || func() bool {
		__antithesis_instrumentation__.Notify(609475)
		return uval > MaxPlaceholderIdx+1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(609476)
		return nil, pgerror.Newf(
			pgcode.NumericValueOutOfRange,
			"placeholder index must be between 1 and %d", MaxPlaceholderIdx+1,
		)
	} else {
		__antithesis_instrumentation__.Notify(609477)
	}
	__antithesis_instrumentation__.Notify(609472)
	return &Placeholder{Idx: PlaceholderIdx(uval - 1)}, nil
}

func (node *Placeholder) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609478)
	if ctx.placeholderFormat != nil {
		__antithesis_instrumentation__.Notify(609480)
		ctx.placeholderFormat(ctx, node)
		return
	} else {
		__antithesis_instrumentation__.Notify(609481)
	}
	__antithesis_instrumentation__.Notify(609479)
	ctx.Printf("$%d", node.Idx+1)
}

func (node *Placeholder) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(609482)
	if node.typ == nil {
		__antithesis_instrumentation__.Notify(609484)
		return types.Any
	} else {
		__antithesis_instrumentation__.Notify(609485)
	}
	__antithesis_instrumentation__.Notify(609483)
	return node.typ
}

type Tuple struct {
	Exprs  Exprs
	Labels []string

	Row bool

	typ *types.T
}

func NewTypedTuple(typ *types.T, typedExprs Exprs) *Tuple {
	__antithesis_instrumentation__.Notify(609486)
	return &Tuple{
		Exprs:  typedExprs,
		Labels: typ.TupleLabels(),
		typ:    typ,
	}
}

func (node *Tuple) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609487)

	if len(node.Labels) > 0 {
		__antithesis_instrumentation__.Notify(609490)
		ctx.WriteByte('(')
	} else {
		__antithesis_instrumentation__.Notify(609491)
	}
	__antithesis_instrumentation__.Notify(609488)
	ctx.WriteByte('(')
	ctx.FormatNode(&node.Exprs)
	if len(node.Exprs) == 1 {
		__antithesis_instrumentation__.Notify(609492)

		ctx.WriteByte(',')
	} else {
		__antithesis_instrumentation__.Notify(609493)
	}
	__antithesis_instrumentation__.Notify(609489)
	ctx.WriteByte(')')
	if len(node.Labels) > 0 {
		__antithesis_instrumentation__.Notify(609494)
		ctx.WriteString(" AS ")
		comma := ""
		for i := range node.Labels {
			__antithesis_instrumentation__.Notify(609496)
			ctx.WriteString(comma)
			ctx.FormatNode((*Name)(&node.Labels[i]))
			comma = ", "
		}
		__antithesis_instrumentation__.Notify(609495)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(609497)
	}
}

func (node *Tuple) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(609498)
	return node.typ
}

type Array struct {
	Exprs Exprs

	typeAnnotation
}

func (node *Array) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609499)
	ctx.WriteString("ARRAY[")
	ctx.FormatNode(&node.Exprs)
	ctx.WriteByte(']')

	if ctx.HasFlags(FmtParsable) && func() bool {
		__antithesis_instrumentation__.Notify(609500)
		return node.typ != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(609501)
		if node.typ.ArrayContents().Family() != types.UnknownFamily {
			__antithesis_instrumentation__.Notify(609502)
			ctx.WriteString(":::")
			ctx.FormatTypeReference(node.typ)
		} else {
			__antithesis_instrumentation__.Notify(609503)
		}
	} else {
		__antithesis_instrumentation__.Notify(609504)
	}
}

type ArrayFlatten struct {
	Subquery Expr

	typeAnnotation
}

func (node *ArrayFlatten) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609505)
	ctx.WriteString("ARRAY ")
	exprFmtWithParen(ctx, node.Subquery)
	if ctx.HasFlags(FmtParsable) {
		__antithesis_instrumentation__.Notify(609506)
		if t, ok := node.Subquery.(*DTuple); ok {
			__antithesis_instrumentation__.Notify(609507)
			if len(t.D) == 0 {
				__antithesis_instrumentation__.Notify(609508)
				ctx.WriteString(":::")
				ctx.Buffer.WriteString(node.typ.SQLString())
			} else {
				__antithesis_instrumentation__.Notify(609509)
			}
		} else {
			__antithesis_instrumentation__.Notify(609510)
		}
	} else {
		__antithesis_instrumentation__.Notify(609511)
	}
}

type Exprs []Expr

func (node *Exprs) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609512)
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(609513)
		if i > 0 {
			__antithesis_instrumentation__.Notify(609515)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(609516)
		}
		__antithesis_instrumentation__.Notify(609514)
		ctx.FormatNode(n)
	}
}

type TypedExprs []TypedExpr

func (node *TypedExprs) String() string {
	__antithesis_instrumentation__.Notify(609517)
	var prefix string
	var buf bytes.Buffer
	for _, n := range *node {
		__antithesis_instrumentation__.Notify(609519)
		fmt.Fprintf(&buf, "%s%s", prefix, n)
		prefix = ", "
	}
	__antithesis_instrumentation__.Notify(609518)
	return buf.String()
}

type Subquery struct {
	Select SelectStatement
	Exists bool

	Idx int

	typeAnnotation
}

func (node *Subquery) SetType(t *types.T) {
	__antithesis_instrumentation__.Notify(609520)
	node.typ = t
}

func (*Subquery) Variable() { __antithesis_instrumentation__.Notify(609521) }

func (*Subquery) SubqueryExpr() { __antithesis_instrumentation__.Notify(609522) }

func (node *Subquery) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609523)
	if ctx.HasFlags(FmtSymbolicSubqueries) {
		__antithesis_instrumentation__.Notify(609524)
		ctx.Printf("@S%d", node.Idx)
	} else {
		__antithesis_instrumentation__.Notify(609525)

		ctx.WithFlags(ctx.flags & ^FmtShowTypes, func() {
			__antithesis_instrumentation__.Notify(609526)
			if node.Exists {
				__antithesis_instrumentation__.Notify(609528)
				ctx.WriteString("EXISTS ")
			} else {
				__antithesis_instrumentation__.Notify(609529)
			}
			__antithesis_instrumentation__.Notify(609527)
			if node.Select == nil {
				__antithesis_instrumentation__.Notify(609530)

				ctx.WriteString("<unknown>")
			} else {
				__antithesis_instrumentation__.Notify(609531)
				ctx.FormatNode(node.Select)
			}
		})
	}
}

type TypedDummy struct {
	Typ *types.T
}

func (node *TypedDummy) String() string {
	__antithesis_instrumentation__.Notify(609532)
	return AsString(node)
}

func (node *TypedDummy) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609533)
	ctx.WriteString("dummyvalof(")
	ctx.FormatTypeReference(node.Typ)
	ctx.WriteString(")")
}

func (node *TypedDummy) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(609534)
	return node.Typ
}

func (node *TypedDummy) TypeCheck(context.Context, *SemaContext, *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(609535)
	return node, nil
}

func (node *TypedDummy) Walk(Visitor) Expr {
	__antithesis_instrumentation__.Notify(609536)
	return node
}

func (node *TypedDummy) Eval(*EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(609537)
	return nil, errors.AssertionFailedf("should not eval typed dummy")
}

var binaryOpPrio = [...]int{
	treebin.Pow:  1,
	treebin.Mult: 2, treebin.Div: 2, treebin.FloorDiv: 2, treebin.Mod: 2,
	treebin.Plus: 3, treebin.Minus: 3,
	treebin.LShift: 4, treebin.RShift: 4,
	treebin.Bitand: 5,
	treebin.Bitxor: 6,
	treebin.Bitor:  7,
	treebin.Concat: 8, treebin.JSONFetchVal: 8, treebin.JSONFetchText: 8, treebin.JSONFetchValPath: 8, treebin.JSONFetchTextPath: 8,
}

var binaryOpFullyAssoc = [...]bool{
	treebin.Pow:  false,
	treebin.Mult: true, treebin.Div: false, treebin.FloorDiv: false, treebin.Mod: false,
	treebin.Plus: true, treebin.Minus: false,
	treebin.LShift: false, treebin.RShift: false,
	treebin.Bitand: true,
	treebin.Bitxor: true,
	treebin.Bitor:  true,
	treebin.Concat: true, treebin.JSONFetchVal: false, treebin.JSONFetchText: false, treebin.JSONFetchValPath: false, treebin.JSONFetchTextPath: false,
}

type BinaryExpr struct {
	Operator    treebin.BinaryOperator
	Left, Right Expr

	typeAnnotation
	Fn *BinOp
}

func (node *BinaryExpr) TypedLeft() TypedExpr {
	__antithesis_instrumentation__.Notify(609538)
	return node.Left.(TypedExpr)
}

func (node *BinaryExpr) TypedRight() TypedExpr {
	__antithesis_instrumentation__.Notify(609539)
	return node.Right.(TypedExpr)
}

func (node *BinaryExpr) ResolvedBinOp() *BinOp {
	__antithesis_instrumentation__.Notify(609540)
	return node.Fn
}

func NewTypedBinaryExpr(
	op treebin.BinaryOperator, left, right TypedExpr, typ *types.T,
) *BinaryExpr {
	__antithesis_instrumentation__.Notify(609541)
	node := &BinaryExpr{Operator: op, Left: left, Right: right}
	node.typ = typ
	node.memoizeFn()
	return node
}

func (*BinaryExpr) operatorExpr() { __antithesis_instrumentation__.Notify(609542) }

func (node *BinaryExpr) memoizeFn() {
	__antithesis_instrumentation__.Notify(609543)
	leftRet, rightRet := node.Left.(TypedExpr).ResolvedType(), node.Right.(TypedExpr).ResolvedType()
	fn, ok := BinOps[node.Operator.Symbol].lookupImpl(leftRet, rightRet)
	if !ok {
		__antithesis_instrumentation__.Notify(609545)
		panic(errors.AssertionFailedf("lookup for BinaryExpr %s's BinOp failed",
			AsStringWithFlags(node, FmtShowTypes)))
	} else {
		__antithesis_instrumentation__.Notify(609546)
	}
	__antithesis_instrumentation__.Notify(609544)
	node.Fn = fn
}

func newBinExprIfValidOverload(
	op treebin.BinaryOperator, left TypedExpr, right TypedExpr,
) *BinaryExpr {
	__antithesis_instrumentation__.Notify(609547)
	leftRet, rightRet := left.ResolvedType(), right.ResolvedType()
	fn, ok := BinOps[op.Symbol].lookupImpl(leftRet, rightRet)
	if ok {
		__antithesis_instrumentation__.Notify(609549)
		expr := &BinaryExpr{
			Operator: op,
			Left:     left,
			Right:    right,
			Fn:       fn,
		}
		expr.typ = returnTypeToFixedType(fn.returnType(), []TypedExpr{left, right})
		return expr
	} else {
		__antithesis_instrumentation__.Notify(609550)
	}
	__antithesis_instrumentation__.Notify(609548)
	return nil
}

func (node *BinaryExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609551)
	binExprFmtWithParen(ctx, node.Left, node.Operator.String(), node.Right, node.Operator.Symbol.IsPadded())
}

type UnaryOperator struct {
	Symbol UnaryOperatorSymbol

	IsExplicitOperator bool
}

func MakeUnaryOperator(symbol UnaryOperatorSymbol) UnaryOperator {
	__antithesis_instrumentation__.Notify(609552)
	return UnaryOperator{Symbol: symbol}
}

func (o UnaryOperator) String() string {
	__antithesis_instrumentation__.Notify(609553)
	if o.IsExplicitOperator {
		__antithesis_instrumentation__.Notify(609555)
		return fmt.Sprintf("OPERATOR(%s)", o.Symbol.String())
	} else {
		__antithesis_instrumentation__.Notify(609556)
	}
	__antithesis_instrumentation__.Notify(609554)
	return o.Symbol.String()
}

func (UnaryOperator) Operator() { __antithesis_instrumentation__.Notify(609557) }

type UnaryOperatorSymbol uint8

const (
	UnaryMinus UnaryOperatorSymbol = iota
	UnaryComplement
	UnarySqrt
	UnaryCbrt
	UnaryPlus

	NumUnaryOperatorSymbols
)

var _ = NumUnaryOperatorSymbols

var unaryOpName = [...]string{
	UnaryMinus:      "-",
	UnaryPlus:       "+",
	UnaryComplement: "~",
	UnarySqrt:       "|/",
	UnaryCbrt:       "||/",
}

func (i UnaryOperatorSymbol) String() string {
	__antithesis_instrumentation__.Notify(609558)
	if i > UnaryOperatorSymbol(len(unaryOpName)-1) {
		__antithesis_instrumentation__.Notify(609560)
		return fmt.Sprintf("UnaryOp(%d)", i)
	} else {
		__antithesis_instrumentation__.Notify(609561)
	}
	__antithesis_instrumentation__.Notify(609559)
	return unaryOpName[i]
}

type UnaryExpr struct {
	Operator UnaryOperator
	Expr     Expr

	typeAnnotation
	fn *UnaryOp
}

func (*UnaryExpr) operatorExpr() { __antithesis_instrumentation__.Notify(609562) }

func (node *UnaryExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609563)
	ctx.WriteString(node.Operator.String())
	e := node.Expr
	_, isOp := e.(operatorExpr)
	_, isDatum := e.(Datum)
	_, isConstant := e.(Constant)
	if isOp || func() bool {
		__antithesis_instrumentation__.Notify(609564)
		return (node.Operator.Symbol == UnaryMinus && func() bool {
			__antithesis_instrumentation__.Notify(609565)
			return (isDatum || func() bool {
				__antithesis_instrumentation__.Notify(609566)
				return isConstant == true
			}() == true) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(609567)
		ctx.WriteByte('(')
		ctx.FormatNode(e)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(609568)
		ctx.FormatNode(e)
	}
}

func (node *UnaryExpr) TypedInnerExpr() TypedExpr {
	__antithesis_instrumentation__.Notify(609569)
	return node.Expr.(TypedExpr)
}

func NewTypedUnaryExpr(op UnaryOperator, expr TypedExpr, typ *types.T) *UnaryExpr {
	__antithesis_instrumentation__.Notify(609570)
	node := &UnaryExpr{Operator: op, Expr: expr}
	node.typ = typ
	innerType := expr.ResolvedType()
	for _, o := range UnaryOps[op.Symbol] {
		__antithesis_instrumentation__.Notify(609572)
		o := o.(*UnaryOp)
		if innerType.Equivalent(o.Typ) && func() bool {
			__antithesis_instrumentation__.Notify(609573)
			return node.typ.Equivalent(o.ReturnType) == true
		}() == true {
			__antithesis_instrumentation__.Notify(609574)
			node.fn = o
			return node
		} else {
			__antithesis_instrumentation__.Notify(609575)
		}
	}
	__antithesis_instrumentation__.Notify(609571)
	panic(errors.AssertionFailedf("invalid TypedExpr with unary op %d: %s", op.Symbol, expr))
}

type FuncExpr struct {
	Func  ResolvableFunctionReference
	Type  funcType
	Exprs Exprs

	Filter    Expr
	WindowDef *WindowDef

	AggType AggType

	OrderBy OrderBy

	typeAnnotation
	fnProps *FunctionProperties
	fn      *Overload
}

func NewTypedFuncExpr(
	ref ResolvableFunctionReference,
	aggQualifier funcType,
	exprs TypedExprs,
	filter TypedExpr,
	windowDef *WindowDef,
	typ *types.T,
	props *FunctionProperties,
	overload *Overload,
) *FuncExpr {
	__antithesis_instrumentation__.Notify(609576)
	f := &FuncExpr{
		Func:           ref,
		Type:           aggQualifier,
		Exprs:          make(Exprs, len(exprs)),
		Filter:         filter,
		WindowDef:      windowDef,
		typeAnnotation: typeAnnotation{typ: typ},
		fn:             overload,
		fnProps:        props,
	}
	for i, e := range exprs {
		__antithesis_instrumentation__.Notify(609578)
		f.Exprs[i] = e
	}
	__antithesis_instrumentation__.Notify(609577)
	return f
}

func (node *FuncExpr) ResolvedOverload() *Overload {
	__antithesis_instrumentation__.Notify(609579)
	return node.fn
}

func (node *FuncExpr) IsGeneratorApplication() bool {
	__antithesis_instrumentation__.Notify(609580)
	return node.fn != nil && func() bool {
		__antithesis_instrumentation__.Notify(609581)
		return (node.fn.Generator != nil || func() bool {
			__antithesis_instrumentation__.Notify(609582)
			return node.fn.GeneratorWithExprs != nil == true
		}() == true) == true
	}() == true
}

func (node *FuncExpr) IsWindowFunctionApplication() bool {
	__antithesis_instrumentation__.Notify(609583)
	return node.WindowDef != nil
}

func (node *FuncExpr) IsDistSQLBlocklist() bool {
	__antithesis_instrumentation__.Notify(609584)
	return (node.fn != nil && func() bool {
		__antithesis_instrumentation__.Notify(609585)
		return node.fn.DistsqlBlocklist == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(609586)
		return (node.fnProps != nil && func() bool {
			__antithesis_instrumentation__.Notify(609587)
			return node.fnProps.DistsqlBlocklist == true
		}() == true) == true
	}() == true
}

func (node *FuncExpr) CanHandleNulls() bool {
	__antithesis_instrumentation__.Notify(609588)
	return node.fnProps != nil && func() bool {
		__antithesis_instrumentation__.Notify(609589)
		return node.fnProps.NullableArgs == true
	}() == true
}

type funcType int

const (
	_ funcType = iota
	DistinctFuncType
	AllFuncType
)

var funcTypeName = [...]string{
	DistinctFuncType: "DISTINCT",
	AllFuncType:      "ALL",
}

type AggType int

const (
	_ AggType = iota

	GeneralAgg

	OrderedSetAgg
)

func (node *FuncExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609590)
	var typ string
	if node.Type != 0 {
		__antithesis_instrumentation__.Notify(609597)
		typ = funcTypeName[node.Type] + " "
	} else {
		__antithesis_instrumentation__.Notify(609598)
	}
	__antithesis_instrumentation__.Notify(609591)

	ctx.WithFlags(ctx.flags&^FmtAnonymize&^FmtMarkRedactionNode, func() {
		__antithesis_instrumentation__.Notify(609599)
		ctx.FormatNode(&node.Func)
	})
	__antithesis_instrumentation__.Notify(609592)

	ctx.WriteByte('(')
	ctx.WriteString(typ)
	ctx.FormatNode(&node.Exprs)
	if node.AggType == GeneralAgg && func() bool {
		__antithesis_instrumentation__.Notify(609600)
		return len(node.OrderBy) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(609601)
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.OrderBy)
	} else {
		__antithesis_instrumentation__.Notify(609602)
	}
	__antithesis_instrumentation__.Notify(609593)
	ctx.WriteByte(')')
	if ctx.HasFlags(FmtParsable) && func() bool {
		__antithesis_instrumentation__.Notify(609603)
		return node.typ != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(609604)
		if node.fnProps.AmbiguousReturnType {
			__antithesis_instrumentation__.Notify(609605)

			if node.typ.Family() != types.TupleFamily {
				__antithesis_instrumentation__.Notify(609606)
				ctx.WriteString(":::")
				ctx.Buffer.WriteString(node.typ.SQLString())
			} else {
				__antithesis_instrumentation__.Notify(609607)
			}
		} else {
			__antithesis_instrumentation__.Notify(609608)
		}
	} else {
		__antithesis_instrumentation__.Notify(609609)
	}
	__antithesis_instrumentation__.Notify(609594)
	if node.AggType == OrderedSetAgg && func() bool {
		__antithesis_instrumentation__.Notify(609610)
		return len(node.OrderBy) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(609611)
		ctx.WriteString(" WITHIN GROUP (")
		ctx.FormatNode(&node.OrderBy)
		ctx.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(609612)
	}
	__antithesis_instrumentation__.Notify(609595)
	if node.Filter != nil {
		__antithesis_instrumentation__.Notify(609613)
		ctx.WriteString(" FILTER (WHERE ")
		ctx.FormatNode(node.Filter)
		ctx.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(609614)
	}
	__antithesis_instrumentation__.Notify(609596)
	if window := node.WindowDef; window != nil {
		__antithesis_instrumentation__.Notify(609615)
		ctx.WriteString(" OVER ")
		if window.Name != "" {
			__antithesis_instrumentation__.Notify(609616)
			ctx.FormatNode(&window.Name)
		} else {
			__antithesis_instrumentation__.Notify(609617)
			ctx.FormatNode(window)
		}
	} else {
		__antithesis_instrumentation__.Notify(609618)
	}
}

type CaseExpr struct {
	Expr  Expr
	Whens []*When
	Else  Expr

	typeAnnotation
}

func (node *CaseExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609619)
	ctx.WriteString("CASE ")
	if node.Expr != nil {
		__antithesis_instrumentation__.Notify(609623)
		ctx.FormatNode(node.Expr)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(609624)
	}
	__antithesis_instrumentation__.Notify(609620)
	for _, when := range node.Whens {
		__antithesis_instrumentation__.Notify(609625)
		ctx.FormatNode(when)
		ctx.WriteByte(' ')
	}
	__antithesis_instrumentation__.Notify(609621)
	if node.Else != nil {
		__antithesis_instrumentation__.Notify(609626)
		ctx.WriteString("ELSE ")
		ctx.FormatNode(node.Else)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(609627)
	}
	__antithesis_instrumentation__.Notify(609622)
	ctx.WriteString("END")
}

func NewTypedCaseExpr(
	expr TypedExpr, whens []*When, elseStmt TypedExpr, typ *types.T,
) (*CaseExpr, error) {
	__antithesis_instrumentation__.Notify(609628)
	node := &CaseExpr{Expr: expr, Whens: whens, Else: elseStmt}
	node.typ = typ
	return node, nil
}

type When struct {
	Cond Expr
	Val  Expr
}

func (node *When) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609629)
	ctx.WriteString("WHEN ")
	ctx.FormatNode(node.Cond)
	ctx.WriteString(" THEN ")
	ctx.FormatNode(node.Val)
}

type castSyntaxMode int

const (
	CastExplicit castSyntaxMode = iota
	CastShort
	CastPrepend
)

type CastExpr struct {
	Expr Expr
	Type ResolvableTypeReference

	typeAnnotation
	SyntaxMode castSyntaxMode
}

func (node *CastExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609630)
	switch node.SyntaxMode {
	case CastPrepend:
		__antithesis_instrumentation__.Notify(609631)

		if _, ok := node.Expr.(*StrVal); ok {
			__antithesis_instrumentation__.Notify(609635)
			ctx.FormatTypeReference(node.Type)
			ctx.WriteByte(' ')

			if ctx.HasFlags(FmtHideConstants) {
				__antithesis_instrumentation__.Notify(609637)
				ctx.WriteString("'_'")
			} else {
				__antithesis_instrumentation__.Notify(609638)
				ctx.FormatNode(node.Expr)
			}
			__antithesis_instrumentation__.Notify(609636)
			break
		} else {
			__antithesis_instrumentation__.Notify(609639)
		}
		__antithesis_instrumentation__.Notify(609632)
		fallthrough
	case CastShort:
		__antithesis_instrumentation__.Notify(609633)
		exprFmtWithParen(ctx, node.Expr)
		ctx.WriteString("::")
		ctx.FormatTypeReference(node.Type)
	default:
		__antithesis_instrumentation__.Notify(609634)
		ctx.WriteString("CAST(")
		ctx.FormatNode(node.Expr)
		ctx.WriteString(" AS ")
		if typ, ok := GetStaticallyKnownType(node.Type); ok && func() bool {
			__antithesis_instrumentation__.Notify(609640)
			return typ.Family() == types.CollatedStringFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(609641)

			strTyp := types.MakeScalar(
				types.StringFamily,
				typ.Oid(),
				typ.Precision(),
				typ.Width(),
				"",
			)
			ctx.WriteString(strTyp.SQLString())
			ctx.WriteString(") COLLATE ")
			lex.EncodeLocaleName(&ctx.Buffer, typ.Locale())
		} else {
			__antithesis_instrumentation__.Notify(609642)
			ctx.FormatTypeReference(node.Type)
			ctx.WriteByte(')')
		}
	}
}

func NewTypedCastExpr(expr TypedExpr, typ *types.T) *CastExpr {
	__antithesis_instrumentation__.Notify(609643)
	node := &CastExpr{Expr: expr, Type: typ, SyntaxMode: CastShort}
	node.typ = typ
	return node
}

type ArraySubscripts []*ArraySubscript

func (a *ArraySubscripts) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609644)
	for _, s := range *a {
		__antithesis_instrumentation__.Notify(609645)
		ctx.FormatNode(s)
	}
}

type IndirectionExpr struct {
	Expr        Expr
	Indirection ArraySubscripts

	typeAnnotation
}

func (node *IndirectionExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609646)

	var annotateArray bool
	if arr, ok := node.Expr.(*Array); ctx.HasFlags(FmtParsable) && func() bool {
		__antithesis_instrumentation__.Notify(609649)
		return ok == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(609650)
		return arr.typ != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(609651)
		if arr.typ.ArrayContents().Family() != types.UnknownFamily {
			__antithesis_instrumentation__.Notify(609652)
			annotateArray = true
		} else {
			__antithesis_instrumentation__.Notify(609653)
		}
	} else {
		__antithesis_instrumentation__.Notify(609654)
	}
	__antithesis_instrumentation__.Notify(609647)
	if _, isCast := node.Expr.(*CastExpr); isCast || func() bool {
		__antithesis_instrumentation__.Notify(609655)
		return annotateArray == true
	}() == true {
		__antithesis_instrumentation__.Notify(609656)
		withParens := ParenExpr{Expr: node.Expr}
		exprFmtWithParen(ctx, &withParens)
	} else {
		__antithesis_instrumentation__.Notify(609657)
		exprFmtWithParen(ctx, node.Expr)
	}
	__antithesis_instrumentation__.Notify(609648)
	ctx.FormatNode(&node.Indirection)
}

type annotateSyntaxMode int

const (
	AnnotateExplicit annotateSyntaxMode = iota
	AnnotateShort
)

type AnnotateTypeExpr struct {
	Expr Expr
	Type ResolvableTypeReference

	SyntaxMode annotateSyntaxMode
}

func (node *AnnotateTypeExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609658)
	switch node.SyntaxMode {
	case AnnotateShort:
		__antithesis_instrumentation__.Notify(609659)
		exprFmtWithParen(ctx, node.Expr)
		ctx.WriteString(":::")
		ctx.FormatTypeReference(node.Type)

	default:
		__antithesis_instrumentation__.Notify(609660)
		ctx.WriteString("ANNOTATE_TYPE(")
		ctx.FormatNode(node.Expr)
		ctx.WriteString(", ")
		ctx.FormatTypeReference(node.Type)
		ctx.WriteByte(')')
	}
}

func (node *AnnotateTypeExpr) TypedInnerExpr() TypedExpr {
	__antithesis_instrumentation__.Notify(609661)
	return node.Expr.(TypedExpr)
}

type CollateExpr struct {
	Expr   Expr
	Locale string

	typeAnnotation
}

func (node *CollateExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609662)
	exprFmtWithParen(ctx, node.Expr)
	ctx.WriteString(" COLLATE ")
	lex.EncodeLocaleName(&ctx.Buffer, node.Locale)
}

type TupleStar struct {
	Expr Expr
}

func (node *TupleStar) NormalizeVarName() (VarName, error) {
	__antithesis_instrumentation__.Notify(609663)
	return node, nil
}

func (node *TupleStar) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609664)
	ctx.WriteByte('(')
	ctx.FormatNode(node.Expr)
	ctx.WriteString(").*")
}

type ColumnAccessExpr struct {
	Expr Expr

	ByIndex bool

	ColName Name

	ColIndex int

	typeAnnotation
}

func NewTypedColumnAccessExpr(expr TypedExpr, colName Name, colIdx int) *ColumnAccessExpr {
	__antithesis_instrumentation__.Notify(609665)
	return &ColumnAccessExpr{
		Expr:           expr,
		ColName:        colName,
		ByIndex:        colName == "",
		ColIndex:       colIdx,
		typeAnnotation: typeAnnotation{typ: expr.ResolvedType().TupleContents()[colIdx]},
	}
}

func (node *ColumnAccessExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609666)
	ctx.WriteByte('(')
	ctx.FormatNode(node.Expr)
	ctx.WriteString(").")
	if node.ByIndex {
		__antithesis_instrumentation__.Notify(609667)
		fmt.Fprintf(ctx, "@%d", node.ColIndex+1)
	} else {
		__antithesis_instrumentation__.Notify(609668)
		ctx.FormatNode(&node.ColName)
	}
}

func (node *AliasedTableExpr) String() string {
	__antithesis_instrumentation__.Notify(609669)
	return AsString(node)
}
func (node *ParenTableExpr) String() string {
	__antithesis_instrumentation__.Notify(609670)
	return AsString(node)
}
func (node *JoinTableExpr) String() string {
	__antithesis_instrumentation__.Notify(609671)
	return AsString(node)
}
func (node *AndExpr) String() string {
	__antithesis_instrumentation__.Notify(609672)
	return AsString(node)
}
func (node *Array) String() string {
	__antithesis_instrumentation__.Notify(609673)
	return AsString(node)
}
func (node *BinaryExpr) String() string {
	__antithesis_instrumentation__.Notify(609674)
	return AsString(node)
}
func (node *CaseExpr) String() string {
	__antithesis_instrumentation__.Notify(609675)
	return AsString(node)
}
func (node *CastExpr) String() string {
	__antithesis_instrumentation__.Notify(609676)
	return AsString(node)
}
func (node *CoalesceExpr) String() string {
	__antithesis_instrumentation__.Notify(609677)
	return AsString(node)
}
func (node *ColumnAccessExpr) String() string {
	__antithesis_instrumentation__.Notify(609678)
	return AsString(node)
}
func (node *CollateExpr) String() string {
	__antithesis_instrumentation__.Notify(609679)
	return AsString(node)
}
func (node *ComparisonExpr) String() string {
	__antithesis_instrumentation__.Notify(609680)
	return AsString(node)
}
func (node *Datums) String() string {
	__antithesis_instrumentation__.Notify(609681)
	return AsString(node)
}
func (node *DBitArray) String() string {
	__antithesis_instrumentation__.Notify(609682)
	return AsString(node)
}
func (node *DBool) String() string {
	__antithesis_instrumentation__.Notify(609683)
	return AsString(node)
}
func (node *DBytes) String() string {
	__antithesis_instrumentation__.Notify(609684)
	return AsString(node)
}
func (node *DDate) String() string {
	__antithesis_instrumentation__.Notify(609685)
	return AsString(node)
}
func (node *DTime) String() string {
	__antithesis_instrumentation__.Notify(609686)
	return AsString(node)
}
func (node *DTimeTZ) String() string {
	__antithesis_instrumentation__.Notify(609687)
	return AsString(node)
}
func (node *DDecimal) String() string {
	__antithesis_instrumentation__.Notify(609688)
	return AsString(node)
}
func (node *DFloat) String() string {
	__antithesis_instrumentation__.Notify(609689)
	return AsString(node)
}
func (node *DBox2D) String() string {
	__antithesis_instrumentation__.Notify(609690)
	return AsString(node)
}
func (node *DGeography) String() string {
	__antithesis_instrumentation__.Notify(609691)
	return AsString(node)
}
func (node *DGeometry) String() string {
	__antithesis_instrumentation__.Notify(609692)
	return AsString(node)
}
func (node *DInt) String() string {
	__antithesis_instrumentation__.Notify(609693)
	return AsString(node)
}
func (node *DInterval) String() string {
	__antithesis_instrumentation__.Notify(609694)
	return AsString(node)
}
func (node *DJSON) String() string {
	__antithesis_instrumentation__.Notify(609695)
	return AsString(node)
}
func (node *DUuid) String() string {
	__antithesis_instrumentation__.Notify(609696)
	return AsString(node)
}
func (node *DIPAddr) String() string {
	__antithesis_instrumentation__.Notify(609697)
	return AsString(node)
}
func (node *DString) String() string {
	__antithesis_instrumentation__.Notify(609698)
	return AsString(node)
}
func (node *DCollatedString) String() string {
	__antithesis_instrumentation__.Notify(609699)
	return AsString(node)
}
func (node *DTimestamp) String() string {
	__antithesis_instrumentation__.Notify(609700)
	return AsString(node)
}
func (node *DTimestampTZ) String() string {
	__antithesis_instrumentation__.Notify(609701)
	return AsString(node)
}
func (node *DTuple) String() string {
	__antithesis_instrumentation__.Notify(609702)
	return AsString(node)
}
func (node *DArray) String() string {
	__antithesis_instrumentation__.Notify(609703)
	return AsString(node)
}
func (node *DOid) String() string {
	__antithesis_instrumentation__.Notify(609704)
	return AsString(node)
}
func (node *DOidWrapper) String() string {
	__antithesis_instrumentation__.Notify(609705)
	return AsString(node)
}
func (node *DVoid) String() string {
	__antithesis_instrumentation__.Notify(609706)
	return AsString(node)
}
func (node *Exprs) String() string {
	__antithesis_instrumentation__.Notify(609707)
	return AsString(node)
}
func (node *ArrayFlatten) String() string {
	__antithesis_instrumentation__.Notify(609708)
	return AsString(node)
}
func (node *FuncExpr) String() string {
	__antithesis_instrumentation__.Notify(609709)
	return AsString(node)
}
func (node *IfExpr) String() string {
	__antithesis_instrumentation__.Notify(609710)
	return AsString(node)
}
func (node *IfErrExpr) String() string {
	__antithesis_instrumentation__.Notify(609711)
	return AsString(node)
}
func (node *IndexedVar) String() string {
	__antithesis_instrumentation__.Notify(609712)
	return AsString(node)
}
func (node *IndirectionExpr) String() string {
	__antithesis_instrumentation__.Notify(609713)
	return AsString(node)
}
func (node *IsOfTypeExpr) String() string {
	__antithesis_instrumentation__.Notify(609714)
	return AsString(node)
}
func (node *Name) String() string {
	__antithesis_instrumentation__.Notify(609715)
	return AsString(node)
}
func (node *UnrestrictedName) String() string {
	__antithesis_instrumentation__.Notify(609716)
	return AsString(node)
}
func (node *NotExpr) String() string {
	__antithesis_instrumentation__.Notify(609717)
	return AsString(node)
}
func (node *IsNullExpr) String() string {
	__antithesis_instrumentation__.Notify(609718)
	return AsString(node)
}
func (node *IsNotNullExpr) String() string {
	__antithesis_instrumentation__.Notify(609719)
	return AsString(node)
}
func (node *NullIfExpr) String() string {
	__antithesis_instrumentation__.Notify(609720)
	return AsString(node)
}
func (node *NumVal) String() string {
	__antithesis_instrumentation__.Notify(609721)
	return AsString(node)
}
func (node *OrExpr) String() string {
	__antithesis_instrumentation__.Notify(609722)
	return AsString(node)
}
func (node *ParenExpr) String() string {
	__antithesis_instrumentation__.Notify(609723)
	return AsString(node)
}
func (node *RangeCond) String() string {
	__antithesis_instrumentation__.Notify(609724)
	return AsString(node)
}
func (node *StrVal) String() string {
	__antithesis_instrumentation__.Notify(609725)
	return AsString(node)
}
func (node *Subquery) String() string {
	__antithesis_instrumentation__.Notify(609726)
	return AsString(node)
}
func (node *Tuple) String() string {
	__antithesis_instrumentation__.Notify(609727)
	return AsString(node)
}
func (node *TupleStar) String() string {
	__antithesis_instrumentation__.Notify(609728)
	return AsString(node)
}
func (node *AnnotateTypeExpr) String() string {
	__antithesis_instrumentation__.Notify(609729)
	return AsString(node)
}
func (node *UnaryExpr) String() string {
	__antithesis_instrumentation__.Notify(609730)
	return AsString(node)
}
func (node DefaultVal) String() string {
	__antithesis_instrumentation__.Notify(609731)
	return AsString(node)
}
func (node PartitionMaxVal) String() string {
	__antithesis_instrumentation__.Notify(609732)
	return AsString(node)
}
func (node PartitionMinVal) String() string {
	__antithesis_instrumentation__.Notify(609733)
	return AsString(node)
}
func (node *Placeholder) String() string {
	__antithesis_instrumentation__.Notify(609734)
	return AsString(node)
}
func (node dNull) String() string {
	__antithesis_instrumentation__.Notify(609735)
	return AsString(node)
}
func (list *NameList) String() string {
	__antithesis_instrumentation__.Notify(609736)
	return AsString(list)
}
