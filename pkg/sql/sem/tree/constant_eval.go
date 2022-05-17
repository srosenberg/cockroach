package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/errors"

type ConstantEvalVisitor struct {
	ctx *EvalContext
	err error

	fastIsConstVisitor fastIsConstVisitor
}

var _ Visitor = &ConstantEvalVisitor{}

func MakeConstantEvalVisitor(ctx *EvalContext) ConstantEvalVisitor {
	__antithesis_instrumentation__.Notify(604620)
	return ConstantEvalVisitor{ctx: ctx, fastIsConstVisitor: fastIsConstVisitor{ctx: ctx}}
}

func (v *ConstantEvalVisitor) Err() error {
	__antithesis_instrumentation__.Notify(604621)
	return v.err
}

func (v *ConstantEvalVisitor) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(604622)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(604624)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(604625)
	}
	__antithesis_instrumentation__.Notify(604623)
	return true, expr
}

func (v *ConstantEvalVisitor) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(604626)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(604631)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(604632)
	}
	__antithesis_instrumentation__.Notify(604627)

	typedExpr, ok := expr.(TypedExpr)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(604633)
		return !v.isConst(expr) == true
	}() == true {
		__antithesis_instrumentation__.Notify(604634)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(604635)
	}
	__antithesis_instrumentation__.Notify(604628)

	value, err := typedExpr.Eval(v.ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(604636)

		return expr
	} else {
		__antithesis_instrumentation__.Notify(604637)
	}
	__antithesis_instrumentation__.Notify(604629)
	if value == DNull {
		__antithesis_instrumentation__.Notify(604638)

		retypedNull, ok := ReType(DNull, typedExpr.ResolvedType())
		if !ok {
			__antithesis_instrumentation__.Notify(604640)
			v.err = errors.AssertionFailedf("failed to retype NULL to %s", typedExpr.ResolvedType())
			return expr
		} else {
			__antithesis_instrumentation__.Notify(604641)
		}
		__antithesis_instrumentation__.Notify(604639)
		return retypedNull
	} else {
		__antithesis_instrumentation__.Notify(604642)
	}
	__antithesis_instrumentation__.Notify(604630)
	return value
}

func (v *ConstantEvalVisitor) isConst(expr Expr) bool {
	__antithesis_instrumentation__.Notify(604643)
	return v.fastIsConstVisitor.run(expr)
}
