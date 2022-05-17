package ingesting

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func GetIngestingDescriptorPrivileges(
	ctx context.Context,
	txn *kv.Txn,
	descsCol *descs.Collection,
	desc catalog.Descriptor,
	user security.SQLUsername,
	wroteDBs map[descpb.ID]catalog.DatabaseDescriptor,
	wroteSchemas map[descpb.ID]catalog.SchemaDescriptor,
	descCoverage tree.DescriptorCoverage,
) (updatedPrivileges *catpb.PrivilegeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(265551)
	switch desc := desc.(type) {
	case catalog.TableDescriptor:
		__antithesis_instrumentation__.Notify(265553)
		return getIngestingPrivilegesForTableOrSchema(
			ctx,
			txn,
			descsCol,
			desc,
			user,
			wroteDBs,
			wroteSchemas,
			descCoverage,
			privilege.Table,
		)
	case catalog.SchemaDescriptor:
		__antithesis_instrumentation__.Notify(265554)
		return getIngestingPrivilegesForTableOrSchema(
			ctx,
			txn,
			descsCol,
			desc,
			user,
			wroteDBs,
			wroteSchemas,
			descCoverage,
			privilege.Schema,
		)
	case catalog.TypeDescriptor:
		__antithesis_instrumentation__.Notify(265555)

		if descCoverage == tree.RequestedDescriptors {
			__antithesis_instrumentation__.Notify(265557)
			updatedPrivileges = catpb.NewBasePrivilegeDescriptor(user)
		} else {
			__antithesis_instrumentation__.Notify(265558)
		}
	case catalog.DatabaseDescriptor:
		__antithesis_instrumentation__.Notify(265556)

		if descCoverage == tree.RequestedDescriptors {
			__antithesis_instrumentation__.Notify(265559)
			updatedPrivileges = catpb.NewBaseDatabasePrivilegeDescriptor(user)
		} else {
			__antithesis_instrumentation__.Notify(265560)
		}
	}
	__antithesis_instrumentation__.Notify(265552)
	return updatedPrivileges, nil
}

func getIngestingPrivilegesForTableOrSchema(
	ctx context.Context,
	txn *kv.Txn,
	descsCol *descs.Collection,
	desc catalog.Descriptor,
	user security.SQLUsername,
	wroteDBs map[descpb.ID]catalog.DatabaseDescriptor,
	wroteSchemas map[descpb.ID]catalog.SchemaDescriptor,
	descCoverage tree.DescriptorCoverage,
	privilegeType privilege.ObjectType,
) (updatedPrivileges *catpb.PrivilegeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(265561)
	if wrote, ok := wroteDBs[desc.GetParentID()]; ok {
		__antithesis_instrumentation__.Notify(265563)

		updatedPrivileges = wrote.GetPrivileges()
		for i, u := range updatedPrivileges.Users {
			__antithesis_instrumentation__.Notify(265564)
			updatedPrivileges.Users[i].Privileges =
				privilege.ListFromBitField(u.Privileges, privilegeType).ToBitField()
		}
	} else {
		__antithesis_instrumentation__.Notify(265565)
		if descCoverage == tree.RequestedDescriptors {
			__antithesis_instrumentation__.Notify(265566)
			parentDB, err := descsCol.Direct().MustGetDatabaseDescByID(ctx, txn, desc.GetParentID())
			if err != nil {
				__antithesis_instrumentation__.Notify(265569)
				return nil, errors.Wrapf(err, "failed to lookup parent DB %d", errors.Safe(desc.GetParentID()))
			} else {
				__antithesis_instrumentation__.Notify(265570)
			}
			__antithesis_instrumentation__.Notify(265567)
			dbDefaultPrivileges := parentDB.GetDefaultPrivilegeDescriptor()

			var schemaDefaultPrivileges catalog.DefaultPrivilegeDescriptor
			targetObject := tree.Schemas
			switch privilegeType {
			case privilege.Table:
				__antithesis_instrumentation__.Notify(265571)
				targetObject = tree.Tables
				schemaID := desc.GetParentSchemaID()

				if schemaID == keys.PublicSchemaID {
					__antithesis_instrumentation__.Notify(265574)
					schemaDefaultPrivileges = nil
				} else {
					__antithesis_instrumentation__.Notify(265575)
					if schema, ok := wroteSchemas[schemaID]; ok {
						__antithesis_instrumentation__.Notify(265576)

						schemaDefaultPrivileges = schema.GetDefaultPrivilegeDescriptor()
					} else {
						__antithesis_instrumentation__.Notify(265577)

						parentSchema, err := descsCol.Direct().MustGetSchemaDescByID(ctx, txn,
							desc.GetParentSchemaID())
						if err != nil {
							__antithesis_instrumentation__.Notify(265579)
							return nil,
								errors.Wrapf(err, "failed to lookup parent schema %d", errors.Safe(desc.GetParentSchemaID()))
						} else {
							__antithesis_instrumentation__.Notify(265580)
						}
						__antithesis_instrumentation__.Notify(265578)
						schemaDefaultPrivileges = parentSchema.GetDefaultPrivilegeDescriptor()
					}
				}
			case privilege.Schema:
				__antithesis_instrumentation__.Notify(265572)
				schemaDefaultPrivileges = nil
			default:
				__antithesis_instrumentation__.Notify(265573)
				return nil, errors.Newf("unexpected privilege type %T", privilegeType)
			}
			__antithesis_instrumentation__.Notify(265568)

			updatedPrivileges = catprivilege.CreatePrivilegesFromDefaultPrivileges(
				dbDefaultPrivileges, schemaDefaultPrivileges,
				parentDB.GetID(), user, targetObject, parentDB.GetPrivileges())
		} else {
			__antithesis_instrumentation__.Notify(265581)
		}
	}
	__antithesis_instrumentation__.Notify(265562)
	return updatedPrivileges, nil
}
