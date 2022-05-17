package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type reparentDatabaseNode struct {
	n         *tree.ReparentDatabase
	db        *dbdesc.Mutable
	newParent *dbdesc.Mutable
}

func (p *planner) ReparentDatabase(
	ctx context.Context, n *tree.ReparentDatabase,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(566486)
	if p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) {
		__antithesis_instrumentation__.Notify(566497)
		return nil, pgerror.Newf(pgcode.FeatureNotSupported,
			"cannot perform ALTER DATABASE CONVERT TO SCHEMA in version %v and beyond",
			clusterversion.PublicSchemasWithDescriptors)
	} else {
		__antithesis_instrumentation__.Notify(566498)
	}
	__antithesis_instrumentation__.Notify(566487)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"REPARENT DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(566499)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566500)
	}
	__antithesis_instrumentation__.Notify(566488)

	if err := p.RequireAdminRole(ctx, "ALTER DATABASE ... CONVERT TO SCHEMA"); err != nil {
		__antithesis_instrumentation__.Notify(566501)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566502)
	}
	__antithesis_instrumentation__.Notify(566489)

	if string(n.Name) == p.CurrentDatabase() {
		__antithesis_instrumentation__.Notify(566503)
		return nil, pgerror.DangerousStatementf("CONVERT TO SCHEMA on current database")
	} else {
		__antithesis_instrumentation__.Notify(566504)
	}
	__antithesis_instrumentation__.Notify(566490)

	sqltelemetry.IncrementUserDefinedSchemaCounter(sqltelemetry.UserDefinedSchemaReparentDatabase)

	db, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Name),
		tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(566505)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566506)
	}
	__antithesis_instrumentation__.Notify(566491)

	parent, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Parent),
		tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(566507)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566508)
	}
	__antithesis_instrumentation__.Notify(566492)

	exists, _, err := schemaExists(ctx, p.txn, p.Descriptors(), parent.ID, db.Name)
	if err != nil {
		__antithesis_instrumentation__.Notify(566509)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566510)
	}
	__antithesis_instrumentation__.Notify(566493)
	if exists {
		__antithesis_instrumentation__.Notify(566511)
		return nil, sqlerrors.NewSchemaAlreadyExistsError(db.Name)
	} else {
		__antithesis_instrumentation__.Notify(566512)
	}
	__antithesis_instrumentation__.Notify(566494)

	if err := schemadesc.IsSchemaNameValid(db.Name); err != nil {
		__antithesis_instrumentation__.Notify(566513)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566514)
	}
	__antithesis_instrumentation__.Notify(566495)

	if len(db.Schemas) > 0 {
		__antithesis_instrumentation__.Notify(566515)
		return nil, pgerror.Newf(pgcode.ObjectNotInPrerequisiteState, "cannot convert database with schemas into schema")
	} else {
		__antithesis_instrumentation__.Notify(566516)
	}
	__antithesis_instrumentation__.Notify(566496)

	return &reparentDatabaseNode{
		n:         n,
		db:        db,
		newParent: parent,
	}, nil
}

func (n *reparentDatabaseNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(566517)
	ctx, p, codec := params.ctx, params.p, params.ExecCfg().Codec

	id, err := descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(566527)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566528)
	}
	__antithesis_instrumentation__.Notify(566518)

	schemaPrivs := privilege.GetValidPrivilegesForObject(privilege.Schema).ToBitField()
	privs := n.db.GetPrivileges()
	for i, u := range privs.Users {
		__antithesis_instrumentation__.Notify(566529)

		privs.Users[i].Privileges = u.Privileges & schemaPrivs
	}
	__antithesis_instrumentation__.Notify(566519)

	schema := schemadesc.NewBuilder(&descpb.SchemaDescriptor{
		ParentID:   n.newParent.ID,
		Name:       n.db.Name,
		ID:         id,
		Privileges: protoutil.Clone(n.db.Privileges).(*catpb.PrivilegeDescriptor),
		Version:    1,
	}).BuildCreatedMutable()

	n.newParent.AddSchemaToDatabase(n.db.Name, descpb.DatabaseDescriptor_SchemaInfo{ID: schema.GetID()})

	if err := p.createDescriptorWithID(
		ctx,
		catalogkeys.MakeSchemaNameKey(p.ExecCfg().Codec, n.newParent.ID, schema.GetName()),
		id,
		schema,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(566530)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566531)
	}
	__antithesis_instrumentation__.Notify(566520)

	b := p.txn.NewBatch()

	objNames, _, err := resolver.GetObjectNamesAndIDs(ctx, p.txn, p, codec, n.db, tree.PublicSchema, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(566532)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566533)
	}
	__antithesis_instrumentation__.Notify(566521)

	for _, objName := range objNames {
		__antithesis_instrumentation__.Notify(566534)

		found, _, desc, err := p.LookupObject(
			ctx,
			tree.ObjectLookupFlags{

				CommonLookupFlags: tree.CommonLookupFlags{
					Required:       false,
					RequireMutable: true,
					IncludeOffline: true,
				},
				DesiredObjectKind: tree.TableObject,
			},
			objName.Catalog(),
			objName.Schema(),
			objName.Object(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(566536)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566537)
		}
		__antithesis_instrumentation__.Notify(566535)
		if found {
			__antithesis_instrumentation__.Notify(566538)
			oldNameKey := descpb.NameInfo{
				ParentID:       desc.GetParentID(),
				ParentSchemaID: desc.GetParentSchemaID(),
				Name:           desc.GetName(),
			}

			tbl, ok := desc.(*tabledesc.Mutable)
			if !ok {
				__antithesis_instrumentation__.Notify(566541)
				return errors.AssertionFailedf("%q was not a Mutable", objName.Object())
			} else {
				__antithesis_instrumentation__.Notify(566542)
			}
			__antithesis_instrumentation__.Notify(566539)

			if len(tbl.GetDependedOnBy()) > 0 {
				__antithesis_instrumentation__.Notify(566543)
				var names []string
				const errStr = "cannot convert database %q into schema because %q has dependent objects"
				tblName, err := p.getQualifiedTableName(ctx, tbl)
				if err != nil {
					__antithesis_instrumentation__.Notify(566546)
					return errors.Wrapf(err, errStr, n.db.Name, tbl.Name)
				} else {
					__antithesis_instrumentation__.Notify(566547)
				}
				__antithesis_instrumentation__.Notify(566544)
				for _, ref := range tbl.GetDependedOnBy() {
					__antithesis_instrumentation__.Notify(566548)
					dep, err := p.Descriptors().GetMutableTableVersionByID(ctx, ref.ID, p.txn)
					if err != nil {
						__antithesis_instrumentation__.Notify(566551)
						return errors.Wrapf(err, errStr, n.db.Name, tblName.FQString())
					} else {
						__antithesis_instrumentation__.Notify(566552)
					}
					__antithesis_instrumentation__.Notify(566549)
					fqName, err := p.getQualifiedTableName(ctx, dep)
					if err != nil {
						__antithesis_instrumentation__.Notify(566553)
						return errors.Wrapf(err, errStr, n.db.Name, dep.Name)
					} else {
						__antithesis_instrumentation__.Notify(566554)
					}
					__antithesis_instrumentation__.Notify(566550)
					names = append(names, fqName.FQString())
				}
				__antithesis_instrumentation__.Notify(566545)
				return sqlerrors.NewDependentObjectErrorf(
					"could not convert database %q into schema because %q has dependent objects %v",
					n.db.Name,
					tblName.FQString(),
					names,
				)
			} else {
				__antithesis_instrumentation__.Notify(566555)
			}
			__antithesis_instrumentation__.Notify(566540)

			tbl.ParentID = n.newParent.ID
			tbl.UnexposedParentSchemaID = schema.GetID()
			p.renameNamespaceEntry(ctx, b, oldNameKey, tbl)
			if err := p.writeSchemaChange(ctx, tbl, descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann())); err != nil {
				__antithesis_instrumentation__.Notify(566556)
				return err
			} else {
				__antithesis_instrumentation__.Notify(566557)
			}
		} else {
			__antithesis_instrumentation__.Notify(566558)

			found, _, desc, err := p.LookupObject(
				ctx,
				tree.ObjectLookupFlags{
					CommonLookupFlags: tree.CommonLookupFlags{
						Required:       true,
						RequireMutable: true,
						IncludeOffline: true,
					},
					DesiredObjectKind: tree.TypeObject,
				},
				objName.Catalog(),
				objName.Schema(),
				objName.Object(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(566562)
				return err
			} else {
				__antithesis_instrumentation__.Notify(566563)
			}
			__antithesis_instrumentation__.Notify(566559)

			if !found {
				__antithesis_instrumentation__.Notify(566564)
				continue
			} else {
				__antithesis_instrumentation__.Notify(566565)
			}
			__antithesis_instrumentation__.Notify(566560)
			oldNameKey := descpb.NameInfo{
				ParentID:       desc.GetParentID(),
				ParentSchemaID: desc.GetParentSchemaID(),
				Name:           desc.GetName(),
			}

			typ, ok := desc.(*typedesc.Mutable)
			if !ok {
				__antithesis_instrumentation__.Notify(566566)
				return errors.AssertionFailedf("%q was not a Mutable", objName.Object())
			} else {
				__antithesis_instrumentation__.Notify(566567)
			}
			__antithesis_instrumentation__.Notify(566561)
			typ.ParentID = n.newParent.ID
			typ.ParentSchemaID = schema.GetID()
			p.renameNamespaceEntry(ctx, b, oldNameKey, typ)
			if err := p.writeTypeSchemaChange(ctx, typ, tree.AsStringWithFQNames(n.n, params.Ann())); err != nil {
				__antithesis_instrumentation__.Notify(566568)
				return err
			} else {
				__antithesis_instrumentation__.Notify(566569)
			}
		}
	}
	__antithesis_instrumentation__.Notify(566522)

	b.Del(catalogkeys.MakeSchemaNameKey(codec, n.db.ID, tree.PublicSchema))

	p.dropNamespaceEntry(ctx, b, n.db)

	n.db.SetDropped()
	if err := p.writeDatabaseChangeToBatch(ctx, n.db, b); err != nil {
		__antithesis_instrumentation__.Notify(566570)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566571)
	}
	__antithesis_instrumentation__.Notify(566523)

	if err := p.writeDatabaseChangeToBatch(ctx, n.newParent, b); err != nil {
		__antithesis_instrumentation__.Notify(566572)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566573)
	}
	__antithesis_instrumentation__.Notify(566524)

	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(566574)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566575)
	}
	__antithesis_instrumentation__.Notify(566525)

	if err := p.createDropDatabaseJob(
		ctx,
		n.db.ID,
		nil,
		nil,
		nil,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(566576)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566577)
	}
	__antithesis_instrumentation__.Notify(566526)

	return p.logEvent(ctx,
		n.db.ID,
		&eventpb.ConvertToSchema{
			DatabaseName:      n.db.Name,
			NewDatabaseParent: n.newParent.Name,
		})
}

func (n *reparentDatabaseNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(566578)
	return false, nil
}
func (n *reparentDatabaseNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(566579)
	return tree.Datums{}
}
func (n *reparentDatabaseNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(566580)
}
func (n *reparentDatabaseNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(566581) }
