package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (p *planner) prepareSetSchema(
	ctx context.Context, db catalog.DatabaseDescriptor, desc catalog.MutableDescriptor, schema string,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(622022)

	var objectName tree.ObjectName
	switch t := desc.(type) {
	case *tabledesc.Mutable:
		__antithesis_instrumentation__.Notify(622028)
		objectName = tree.NewUnqualifiedTableName(tree.Name(desc.GetName()))
	case *typedesc.Mutable:
		__antithesis_instrumentation__.Notify(622029)
		objectName = tree.NewUnqualifiedTypeName(desc.GetName())
	default:
		__antithesis_instrumentation__.Notify(622030)
		return 0, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"no table or type was found for SET SCHEMA command, found %T", t)
	}
	__antithesis_instrumentation__.Notify(622023)

	res, err := p.Descriptors().GetMutableSchemaByName(
		ctx, p.txn, db, schema, tree.SchemaLookupFlags{
			Required:       true,
			RequireMutable: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(622031)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(622032)
	}
	__antithesis_instrumentation__.Notify(622024)

	switch res.SchemaKind() {
	case catalog.SchemaTemporary:
		__antithesis_instrumentation__.Notify(622033)
		return 0, pgerror.Newf(pgcode.FeatureNotSupported,
			"cannot move objects into or out of temporary schemas")
	case catalog.SchemaVirtual:
		__antithesis_instrumentation__.Notify(622034)
		return 0, pgerror.Newf(pgcode.FeatureNotSupported,
			"cannot move objects into or out of virtual schemas")
	case catalog.SchemaPublic:
		__antithesis_instrumentation__.Notify(622035)

	default:
		__antithesis_instrumentation__.Notify(622036)

		err = p.CheckPrivilege(ctx, res, privilege.CREATE)
		if err != nil {
			__antithesis_instrumentation__.Notify(622037)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(622038)
		}
	}
	__antithesis_instrumentation__.Notify(622025)

	desiredSchemaID := res.GetID()

	if desiredSchemaID == desc.GetParentSchemaID() {
		__antithesis_instrumentation__.Notify(622039)
		return desiredSchemaID, nil
	} else {
		__antithesis_instrumentation__.Notify(622040)
	}
	__antithesis_instrumentation__.Notify(622026)

	err = p.Descriptors().Direct().CheckObjectCollision(ctx, p.txn, db.GetID(), desiredSchemaID, objectName)
	if err != nil {
		__antithesis_instrumentation__.Notify(622041)
		return descpb.InvalidID, err
	} else {
		__antithesis_instrumentation__.Notify(622042)
	}
	__antithesis_instrumentation__.Notify(622027)

	return desiredSchemaID, nil
}
