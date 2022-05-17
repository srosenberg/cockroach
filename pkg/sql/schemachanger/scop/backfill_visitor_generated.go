package scop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "context"

type BackfillOp interface {
	Op
	Visit(context.Context, BackfillVisitor) error
}

type BackfillVisitor interface {
	BackfillIndex(context.Context, BackfillIndex) error
}

func (op BackfillIndex) Visit(ctx context.Context, v BackfillVisitor) error {
	__antithesis_instrumentation__.Notify(582382)
	return v.BackfillIndex(ctx, op)
}
