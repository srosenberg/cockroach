package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type alterSchemaNode struct {
	n    *tree.AlterSchema
	db   *dbdesc.Mutable
	desc *schemadesc.Mutable
}

var _ planNode = &alterSchemaNode{n: nil}

func (p *planner) AlterSchema(ctx context.Context, n *tree.AlterSchema) (planNode, error) {
	__antithesis_instrumentation__.Notify(243769)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER SCHEMA",
	); err != nil {
		__antithesis_instrumentation__.Notify(243775)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243776)
	}
	__antithesis_instrumentation__.Notify(243770)

	dbName := p.CurrentDatabase()
	if n.Schema.ExplicitCatalog {
		__antithesis_instrumentation__.Notify(243777)
		dbName = n.Schema.Catalog()
	} else {
		__antithesis_instrumentation__.Notify(243778)
	}
	__antithesis_instrumentation__.Notify(243771)
	db, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, dbName,
		tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(243779)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243780)
	}
	__antithesis_instrumentation__.Notify(243772)
	schema, err := p.Descriptors().GetSchemaByName(ctx, p.txn, db,
		string(n.Schema.SchemaName), tree.SchemaLookupFlags{
			Required:       true,
			RequireMutable: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(243781)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243782)
	}
	__antithesis_instrumentation__.Notify(243773)

	if schema.GetName() == tree.PublicSchema {
		__antithesis_instrumentation__.Notify(243783)
		return nil, pgerror.Newf(pgcode.InvalidSchemaName, "cannot modify schema %q", n.Schema.String())
	} else {
		__antithesis_instrumentation__.Notify(243784)
	}
	__antithesis_instrumentation__.Notify(243774)
	switch schema.SchemaKind() {
	case catalog.SchemaPublic, catalog.SchemaVirtual, catalog.SchemaTemporary:
		__antithesis_instrumentation__.Notify(243785)
		return nil, pgerror.Newf(pgcode.InvalidSchemaName, "cannot modify schema %q", n.Schema.String())
	case catalog.SchemaUserDefined:
		__antithesis_instrumentation__.Notify(243786)
		desc := schema.(*schemadesc.Mutable)

		hasAdmin, err := p.HasAdminRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(243790)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(243791)
		}
		__antithesis_instrumentation__.Notify(243787)
		if !hasAdmin {
			__antithesis_instrumentation__.Notify(243792)
			hasOwnership, err := p.HasOwnership(ctx, desc)
			if err != nil {
				__antithesis_instrumentation__.Notify(243794)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(243795)
			}
			__antithesis_instrumentation__.Notify(243793)
			if !hasOwnership {
				__antithesis_instrumentation__.Notify(243796)
				return nil, pgerror.Newf(pgcode.InsufficientPrivilege, "must be owner of schema %q", desc.Name)
			} else {
				__antithesis_instrumentation__.Notify(243797)
			}
		} else {
			__antithesis_instrumentation__.Notify(243798)
		}
		__antithesis_instrumentation__.Notify(243788)
		sqltelemetry.IncrementUserDefinedSchemaCounter(sqltelemetry.UserDefinedSchemaAlter)
		return &alterSchemaNode{n: n, db: db, desc: desc}, nil
	default:
		__antithesis_instrumentation__.Notify(243789)
		return nil, errors.AssertionFailedf("unknown schema kind")
	}
}

func (n *alterSchemaNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(243799)
	switch t := n.n.Cmd.(type) {
	case *tree.AlterSchemaRename:
		__antithesis_instrumentation__.Notify(243800)
		newName := string(t.NewName)

		oldQualifiedSchemaName, err := params.p.getQualifiedSchemaName(params.ctx, n.desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(243807)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243808)
		}
		__antithesis_instrumentation__.Notify(243801)

		if err := params.p.renameSchema(
			params.ctx, n.db, n.desc, newName, tree.AsStringWithFQNames(n.n, params.Ann()),
		); err != nil {
			__antithesis_instrumentation__.Notify(243809)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243810)
		}
		__antithesis_instrumentation__.Notify(243802)

		newQualifiedSchemaName, err := params.p.getQualifiedSchemaName(params.ctx, n.desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(243811)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243812)
		}
		__antithesis_instrumentation__.Notify(243803)

		return params.p.logEvent(params.ctx, n.desc.ID, &eventpb.RenameSchema{
			SchemaName:    oldQualifiedSchemaName.String(),
			NewSchemaName: newQualifiedSchemaName.String(),
		})
	case *tree.AlterSchemaOwner:
		__antithesis_instrumentation__.Notify(243804)
		newOwner, err := t.Owner.ToSQLUsername(params.p.SessionData(), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(243813)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243814)
		}
		__antithesis_instrumentation__.Notify(243805)
		return params.p.alterSchemaOwner(
			params.ctx, n.desc, newOwner, tree.AsStringWithFQNames(n.n, params.Ann()),
		)
	default:
		__antithesis_instrumentation__.Notify(243806)
		return errors.AssertionFailedf("unknown schema cmd %T", t)
	}
}

func (p *planner) alterSchemaOwner(
	ctx context.Context,
	scDesc *schemadesc.Mutable,
	newOwner security.SQLUsername,
	jobDescription string,
) error {
	__antithesis_instrumentation__.Notify(243815)
	oldOwner := scDesc.GetPrivileges().Owner()

	if err := p.checkCanAlterToNewOwner(ctx, scDesc, newOwner); err != nil {
		__antithesis_instrumentation__.Notify(243821)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243822)
	}
	__antithesis_instrumentation__.Notify(243816)

	parentDBDesc, err := p.Descriptors().GetMutableDescriptorByID(ctx, p.txn, scDesc.GetParentID())
	if err != nil {
		__antithesis_instrumentation__.Notify(243823)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243824)
	}
	__antithesis_instrumentation__.Notify(243817)
	if err := p.CheckPrivilege(ctx, parentDBDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(243825)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243826)
	}
	__antithesis_instrumentation__.Notify(243818)

	if err := p.setNewSchemaOwner(ctx, parentDBDesc.(*dbdesc.Mutable), scDesc, newOwner); err != nil {
		__antithesis_instrumentation__.Notify(243827)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243828)
	}
	__antithesis_instrumentation__.Notify(243819)

	if newOwner == oldOwner {
		__antithesis_instrumentation__.Notify(243829)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(243830)
	}
	__antithesis_instrumentation__.Notify(243820)

	return p.writeSchemaDescChange(ctx, scDesc, jobDescription)
}

func (p *planner) setNewSchemaOwner(
	ctx context.Context,
	dbDesc *dbdesc.Mutable,
	scDesc *schemadesc.Mutable,
	newOwner security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(243831)

	privs := scDesc.GetPrivileges()
	privs.SetOwner(newOwner)

	qualifiedSchemaName := &tree.ObjectNamePrefix{
		CatalogName:     tree.Name(dbDesc.GetName()),
		SchemaName:      tree.Name(scDesc.GetName()),
		ExplicitCatalog: true,
		ExplicitSchema:  true,
	}

	return p.logEvent(ctx,
		scDesc.GetID(),
		&eventpb.AlterSchemaOwner{
			SchemaName: qualifiedSchemaName.String(),
			Owner:      newOwner.Normalized(),
		})
}

func (p *planner) renameSchema(
	ctx context.Context, db *dbdesc.Mutable, desc *schemadesc.Mutable, newName string, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(243832)
	oldNameKey := descpb.NameInfo{
		ParentID:       desc.GetParentID(),
		ParentSchemaID: desc.GetParentSchemaID(),
		Name:           desc.GetName(),
	}

	found, _, err := schemaExists(ctx, p.txn, p.Descriptors(), db.ID, newName)
	if err != nil {
		__antithesis_instrumentation__.Notify(243841)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243842)
	}
	__antithesis_instrumentation__.Notify(243833)
	if found {
		__antithesis_instrumentation__.Notify(243843)
		return sqlerrors.NewSchemaAlreadyExistsError(newName)
	} else {
		__antithesis_instrumentation__.Notify(243844)
	}
	__antithesis_instrumentation__.Notify(243834)

	if err := schemadesc.IsSchemaNameValid(newName); err != nil {
		__antithesis_instrumentation__.Notify(243845)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243846)
	}
	__antithesis_instrumentation__.Notify(243835)

	oldName := oldNameKey.GetName()
	desc.SetName(newName)

	b := p.txn.NewBatch()
	p.renameNamespaceEntry(ctx, b, oldNameKey, desc)
	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(243847)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243848)
	}
	__antithesis_instrumentation__.Notify(243836)

	_, oldPresent := db.Schemas[oldName]
	_, newPresent := db.Schemas[newName]
	if !oldPresent {
		__antithesis_instrumentation__.Notify(243849)
		return errors.AssertionFailedf(
			"old name %q not present in database schema mapping",
			oldName,
		)
	} else {
		__antithesis_instrumentation__.Notify(243850)
	}
	__antithesis_instrumentation__.Notify(243837)
	if newPresent {
		__antithesis_instrumentation__.Notify(243851)
		return errors.AssertionFailedf(
			"new name %q already present in database schema mapping",
			newName,
		)
	} else {
		__antithesis_instrumentation__.Notify(243852)
	}
	__antithesis_instrumentation__.Notify(243838)

	if p.execCfg.Settings.Version.IsActive(ctx, clusterversion.AvoidDrainingNames) {
		__antithesis_instrumentation__.Notify(243853)
		delete(db.Schemas, oldName)
	} else {
		__antithesis_instrumentation__.Notify(243854)
		db.Schemas[oldName] = descpb.DatabaseDescriptor_SchemaInfo{
			ID:      desc.ID,
			Dropped: true,
		}
	}
	__antithesis_instrumentation__.Notify(243839)

	db.Schemas[newName] = descpb.DatabaseDescriptor_SchemaInfo{ID: desc.ID}
	if err := p.writeNonDropDatabaseChange(
		ctx, db,
		fmt.Sprintf("updating parent database %s for %s", db.GetName(), jobDesc),
	); err != nil {
		__antithesis_instrumentation__.Notify(243855)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243856)
	}
	__antithesis_instrumentation__.Notify(243840)

	return p.writeSchemaDescChange(ctx, desc, jobDesc)
}

func (n *alterSchemaNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(243857)
	return false, nil
}
func (n *alterSchemaNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(243858)
	return tree.Datums{}
}
func (n *alterSchemaNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(243859) }
func (n *alterSchemaNode) ReadingOwnWrites()         { __antithesis_instrumentation__.Notify(243860) }
