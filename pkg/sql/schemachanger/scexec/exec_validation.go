package scexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
)

func executeValidateUniqueIndex(
	ctx context.Context, deps Dependencies, op *scop.ValidateUniqueIndex,
) error {
	__antithesis_instrumentation__.Notify(581592)
	descs, err := deps.Catalog().MustReadImmutableDescriptors(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581597)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581598)
	}
	__antithesis_instrumentation__.Notify(581593)
	desc := descs[0]
	table, ok := desc.(catalog.TableDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(581599)
		return catalog.WrapTableDescRefErr(desc.GetID(), catalog.NewDescriptorTypeError(desc))
	} else {
		__antithesis_instrumentation__.Notify(581600)
	}
	__antithesis_instrumentation__.Notify(581594)
	index, err := table.FindIndexWithID(op.IndexID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581601)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581602)
	}
	__antithesis_instrumentation__.Notify(581595)

	execOverride := sessiondata.InternalExecutorOverride{
		User: security.RootUserName(),
	}
	if index.GetType() == descpb.IndexDescriptor_FORWARD {
		__antithesis_instrumentation__.Notify(581603)
		err = deps.IndexValidator().ValidateForwardIndexes(ctx, table, []catalog.Index{index}, execOverride)
	} else {
		__antithesis_instrumentation__.Notify(581604)
		err = deps.IndexValidator().ValidateInvertedIndexes(ctx, table, []catalog.Index{index}, execOverride)
	}
	__antithesis_instrumentation__.Notify(581596)
	return err
}

func executeValidateCheckConstraint(
	ctx context.Context, deps Dependencies, op *scop.ValidateCheckConstraint,
) error {
	__antithesis_instrumentation__.Notify(581605)
	return errors.Errorf("executeValidateCheckConstraint is not implemented")
}

func executeValidationOps(ctx context.Context, deps Dependencies, execute []scop.Op) error {
	__antithesis_instrumentation__.Notify(581606)
	for _, op := range execute {
		__antithesis_instrumentation__.Notify(581608)
		switch op := op.(type) {
		case *scop.ValidateUniqueIndex:
			__antithesis_instrumentation__.Notify(581609)
			return executeValidateUniqueIndex(ctx, deps, op)
		case *scop.ValidateCheckConstraint:
			__antithesis_instrumentation__.Notify(581610)
			return executeValidateCheckConstraint(ctx, deps, op)
		default:
			__antithesis_instrumentation__.Notify(581611)
			panic("unimplemented")
		}
	}
	__antithesis_instrumentation__.Notify(581607)
	return nil
}
