package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/errors"
)

func (tc *Collection) GetMutableSchemaByName(
	ctx context.Context,
	txn *kv.Txn,
	db catalog.DatabaseDescriptor,
	schemaName string,
	flags tree.SchemaLookupFlags,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(264824)
	flags.RequireMutable = true
	return tc.getSchemaByName(ctx, txn, db, schemaName, flags)
}

func (tc *Collection) GetSchemaByName(
	ctx context.Context,
	txn *kv.Txn,
	db catalog.DatabaseDescriptor,
	scName string,
	flags tree.SchemaLookupFlags,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(264825)
	return tc.getSchemaByName(ctx, txn, db, scName, flags)
}

func (tc *Collection) getSchemaByName(
	ctx context.Context,
	txn *kv.Txn,
	db catalog.DatabaseDescriptor,
	schemaName string,
	flags tree.SchemaLookupFlags,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(264826)
	const alwaysLookupLeasedPublicSchema = false
	return tc.getSchemaByNameMaybeLookingUpPublicSchema(
		ctx, txn, db, schemaName, flags, alwaysLookupLeasedPublicSchema,
	)
}

func (tc *Collection) getSchemaByNameMaybeLookingUpPublicSchema(
	ctx context.Context,
	txn *kv.Txn,
	db catalog.DatabaseDescriptor,
	schemaName string,
	flags tree.SchemaLookupFlags,
	alwaysLookupLeasedPublicSchema bool,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(264827)
	found, desc, err := tc.getByName(
		ctx, txn, db, nil, schemaName, flags.AvoidLeased, flags.RequireMutable,
		flags.AvoidSynthetic, alwaysLookupLeasedPublicSchema,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(264831)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264832)
		if !found {
			__antithesis_instrumentation__.Notify(264833)
			if flags.Required {
				__antithesis_instrumentation__.Notify(264835)
				return nil, sqlerrors.NewUndefinedSchemaError(schemaName)
			} else {
				__antithesis_instrumentation__.Notify(264836)
			}
			__antithesis_instrumentation__.Notify(264834)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(264837)
		}
	}
	__antithesis_instrumentation__.Notify(264828)
	schema, ok := desc.(catalog.SchemaDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(264838)
		if flags.Required {
			__antithesis_instrumentation__.Notify(264840)
			return nil, sqlerrors.NewUndefinedSchemaError(schemaName)
		} else {
			__antithesis_instrumentation__.Notify(264841)
		}
		__antithesis_instrumentation__.Notify(264839)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264842)
	}
	__antithesis_instrumentation__.Notify(264829)
	if dropped, err := filterDescriptorState(schema, flags.Required, flags); dropped || func() bool {
		__antithesis_instrumentation__.Notify(264843)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(264844)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264845)
	}
	__antithesis_instrumentation__.Notify(264830)
	return schema, nil
}

func (tc *Collection) GetImmutableSchemaByID(
	ctx context.Context, txn *kv.Txn, schemaID descpb.ID, flags tree.SchemaLookupFlags,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(264846)
	flags.RequireMutable = false
	return tc.getSchemaByID(ctx, txn, schemaID, flags)
}

func (tc *Collection) GetImmutableSchemaByName(
	ctx context.Context,
	txn *kv.Txn,
	db catalog.DatabaseDescriptor,
	schemaName string,
	flags tree.SchemaLookupFlags,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(264847)
	flags.RequireMutable = false
	return tc.getSchemaByName(ctx, txn, db, schemaName, flags)
}

func (tc *Collection) getSchemaByID(
	ctx context.Context, txn *kv.Txn, schemaID descpb.ID, flags tree.SchemaLookupFlags,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(264848)

	if schemaID == keys.PublicSchemaID {
		__antithesis_instrumentation__.Notify(264854)
		return schemadesc.GetPublicSchema(), nil
	} else {
		__antithesis_instrumentation__.Notify(264855)
	}
	__antithesis_instrumentation__.Notify(264849)
	if sc, err := tc.virtual.getSchemaByID(
		ctx, schemaID, flags.RequireMutable,
	); err != nil {
		__antithesis_instrumentation__.Notify(264856)
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(264858)
			if flags.Required {
				__antithesis_instrumentation__.Notify(264860)
				return nil, sqlerrors.NewUndefinedSchemaError(fmt.Sprintf("[%d]", schemaID))
			} else {
				__antithesis_instrumentation__.Notify(264861)
			}
			__antithesis_instrumentation__.Notify(264859)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(264862)
		}
		__antithesis_instrumentation__.Notify(264857)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264863)
		if sc != nil {
			__antithesis_instrumentation__.Notify(264864)
			return sc, err
		} else {
			__antithesis_instrumentation__.Notify(264865)
		}
	}
	__antithesis_instrumentation__.Notify(264850)

	if sc := tc.temporary.getSchemaByID(ctx, schemaID); sc != nil {
		__antithesis_instrumentation__.Notify(264866)
		return sc, nil
	} else {
		__antithesis_instrumentation__.Notify(264867)
	}
	__antithesis_instrumentation__.Notify(264851)

	descs, err := tc.getDescriptorsByID(ctx, txn, flags, schemaID)
	if err != nil {
		__antithesis_instrumentation__.Notify(264868)
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(264870)
			if flags.Required {
				__antithesis_instrumentation__.Notify(264872)
				return nil, sqlerrors.NewUndefinedSchemaError(fmt.Sprintf("[%d]", schemaID))
			} else {
				__antithesis_instrumentation__.Notify(264873)
			}
			__antithesis_instrumentation__.Notify(264871)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(264874)
		}
		__antithesis_instrumentation__.Notify(264869)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264875)
	}
	__antithesis_instrumentation__.Notify(264852)
	schemaDesc, ok := descs[0].(catalog.SchemaDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(264876)
		return nil, sqlerrors.NewUndefinedSchemaError(fmt.Sprintf("[%d]", schemaID))
	} else {
		__antithesis_instrumentation__.Notify(264877)
	}
	__antithesis_instrumentation__.Notify(264853)

	return schemaDesc, nil
}
