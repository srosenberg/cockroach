package scop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "context"

type ValidationOp interface {
	Op
	Visit(context.Context, ValidationVisitor) error
}

type ValidationVisitor interface {
	ValidateUniqueIndex(context.Context, ValidateUniqueIndex) error
	ValidateCheckConstraint(context.Context, ValidateCheckConstraint) error
}

func (op ValidateUniqueIndex) Visit(ctx context.Context, v ValidationVisitor) error {
	__antithesis_instrumentation__.Notify(582486)
	return v.ValidateUniqueIndex(ctx, op)
}

func (op ValidateCheckConstraint) Visit(ctx context.Context, v ValidationVisitor) error {
	__antithesis_instrumentation__.Notify(582487)
	return v.ValidateCheckConstraint(ctx, op)
}
