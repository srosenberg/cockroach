package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/errors"
)

func (tc *Collection) GetMutableDatabaseByName(
	ctx context.Context, txn *kv.Txn, name string, flags tree.DatabaseLookupFlags,
) (*dbdesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(264152)
	flags.RequireMutable = true
	desc, err := tc.getDatabaseByName(ctx, txn, name, flags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264154)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(264155)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264156)
	}
	__antithesis_instrumentation__.Notify(264153)
	return desc.(*dbdesc.Mutable), nil
}

func (tc *Collection) GetImmutableDatabaseByName(
	ctx context.Context, txn *kv.Txn, name string, flags tree.DatabaseLookupFlags,
) (catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(264157)
	flags.RequireMutable = false
	return tc.getDatabaseByName(ctx, txn, name, flags)
}

func (tc *Collection) GetDatabaseDesc(
	ctx context.Context, txn *kv.Txn, name string, flags tree.DatabaseLookupFlags,
) (desc catalog.DatabaseDescriptor, err error) {
	__antithesis_instrumentation__.Notify(264158)
	return tc.getDatabaseByName(ctx, txn, name, flags)
}

func (tc *Collection) getDatabaseByName(
	ctx context.Context, txn *kv.Txn, name string, flags tree.DatabaseLookupFlags,
) (catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(264159)
	const alwaysLookupLeasedPublicSchema = false
	found, desc, err := tc.getByName(
		ctx, txn, nil, nil, name, flags.AvoidLeased, flags.RequireMutable, flags.AvoidSynthetic,
		alwaysLookupLeasedPublicSchema,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(264163)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264164)
		if !found {
			__antithesis_instrumentation__.Notify(264165)
			if flags.Required {
				__antithesis_instrumentation__.Notify(264167)
				return nil, sqlerrors.NewUndefinedDatabaseError(name)
			} else {
				__antithesis_instrumentation__.Notify(264168)
			}
			__antithesis_instrumentation__.Notify(264166)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(264169)
		}
	}
	__antithesis_instrumentation__.Notify(264160)
	db, ok := desc.(catalog.DatabaseDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(264170)
		if flags.Required {
			__antithesis_instrumentation__.Notify(264172)
			return nil, sqlerrors.NewUndefinedDatabaseError(name)
		} else {
			__antithesis_instrumentation__.Notify(264173)
		}
		__antithesis_instrumentation__.Notify(264171)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264174)
	}
	__antithesis_instrumentation__.Notify(264161)
	if dropped, err := filterDescriptorState(db, flags.Required, flags); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264175)
		return dropped == true
	}() == true {
		__antithesis_instrumentation__.Notify(264176)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264177)
	}
	__antithesis_instrumentation__.Notify(264162)
	return db, nil
}

func (tc *Collection) GetImmutableDatabaseByID(
	ctx context.Context, txn *kv.Txn, dbID descpb.ID, flags tree.DatabaseLookupFlags,
) (bool, catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(264178)
	flags.RequireMutable = false
	return tc.getDatabaseByID(ctx, txn, dbID, flags)
}

func (tc *Collection) getDatabaseByID(
	ctx context.Context, txn *kv.Txn, dbID descpb.ID, flags tree.DatabaseLookupFlags,
) (bool, catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(264179)
	descs, err := tc.getDescriptorsByID(ctx, txn, flags, dbID)
	if err != nil {
		__antithesis_instrumentation__.Notify(264182)
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(264184)
			if flags.Required {
				__antithesis_instrumentation__.Notify(264186)
				return false, nil, sqlerrors.NewUndefinedDatabaseError(fmt.Sprintf("[%d]", dbID))
			} else {
				__antithesis_instrumentation__.Notify(264187)
			}
			__antithesis_instrumentation__.Notify(264185)
			return false, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(264188)
		}
		__antithesis_instrumentation__.Notify(264183)
		return false, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264189)
	}
	__antithesis_instrumentation__.Notify(264180)
	db, ok := descs[0].(catalog.DatabaseDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(264190)
		return false, nil, sqlerrors.NewUndefinedDatabaseError(fmt.Sprintf("[%d]", dbID))
	} else {
		__antithesis_instrumentation__.Notify(264191)
	}
	__antithesis_instrumentation__.Notify(264181)
	return true, db, nil
}
