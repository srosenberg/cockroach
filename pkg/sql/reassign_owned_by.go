package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/errors"
)

type reassignOwnedByNode struct {
	n                  *tree.ReassignOwnedBy
	normalizedOldRoles []security.SQLUsername
}

func (p *planner) ReassignOwnedBy(ctx context.Context, n *tree.ReassignOwnedBy) (planNode, error) {
	__antithesis_instrumentation__.Notify(564778)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"REASSIGN OWNED BY",
	); err != nil {
		__antithesis_instrumentation__.Notify(564787)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(564788)
	}
	__antithesis_instrumentation__.Notify(564779)

	normalizedOldRoles, err := n.OldRoles.ToSQLUsernames(p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(564789)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(564790)
	}
	__antithesis_instrumentation__.Notify(564780)

	for _, oldRole := range normalizedOldRoles {
		__antithesis_instrumentation__.Notify(564791)
		roleExists, err := RoleExists(ctx, p.ExecCfg(), p.Txn(), oldRole)
		if err != nil {
			__antithesis_instrumentation__.Notify(564793)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(564794)
		}
		__antithesis_instrumentation__.Notify(564792)
		if !roleExists {
			__antithesis_instrumentation__.Notify(564795)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "role/user %q does not exist", oldRole)
		} else {
			__antithesis_instrumentation__.Notify(564796)
		}
	}
	__antithesis_instrumentation__.Notify(564781)
	newRole, err := n.NewRole.ToSQLUsername(p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(564797)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(564798)
	}
	__antithesis_instrumentation__.Notify(564782)
	roleExists, err := RoleExists(ctx, p.ExecCfg(), p.Txn(), newRole)
	if !roleExists {
		__antithesis_instrumentation__.Notify(564799)
		return nil, pgerror.Newf(pgcode.UndefinedObject, "role/user %q does not exist", newRole)
	} else {
		__antithesis_instrumentation__.Notify(564800)
	}
	__antithesis_instrumentation__.Notify(564783)
	if err != nil {
		__antithesis_instrumentation__.Notify(564801)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(564802)
	}
	__antithesis_instrumentation__.Notify(564784)

	hasAdminRole, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(564803)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(564804)
	}
	__antithesis_instrumentation__.Notify(564785)

	if !hasAdminRole {
		__antithesis_instrumentation__.Notify(564805)
		memberOf, err := p.MemberOfWithAdminOption(ctx, p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(564808)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(564809)
		}
		__antithesis_instrumentation__.Notify(564806)
		if p.User() != newRole {
			__antithesis_instrumentation__.Notify(564810)
			if _, ok := memberOf[newRole]; !ok {
				__antithesis_instrumentation__.Notify(564811)
				return nil, errors.WithHint(
					pgerror.Newf(pgcode.InsufficientPrivilege,
						"permission denied to reassign objects"),
					"user must be a member of the new role")
			} else {
				__antithesis_instrumentation__.Notify(564812)
			}
		} else {
			__antithesis_instrumentation__.Notify(564813)
		}
		__antithesis_instrumentation__.Notify(564807)
		for _, oldRole := range normalizedOldRoles {
			__antithesis_instrumentation__.Notify(564814)
			if p.User() != oldRole {
				__antithesis_instrumentation__.Notify(564815)
				if _, ok := memberOf[oldRole]; !ok {
					__antithesis_instrumentation__.Notify(564816)
					return nil, errors.WithHint(
						pgerror.Newf(pgcode.InsufficientPrivilege,
							"permission denied to reassign objects"),
						"user must be a member of the old roles")
				} else {
					__antithesis_instrumentation__.Notify(564817)
				}
			} else {
				__antithesis_instrumentation__.Notify(564818)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(564819)
	}
	__antithesis_instrumentation__.Notify(564786)
	return &reassignOwnedByNode{n: n, normalizedOldRoles: normalizedOldRoles}, nil
}

func (n *reassignOwnedByNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(564820)
	telemetry.Inc(sqltelemetry.CreateReassignOwnedByCounter())

	all, err := params.p.Descriptors().GetAllDescriptors(params.ctx, params.p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(564824)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564825)
	}
	__antithesis_instrumentation__.Notify(564821)

	currentDatabase := params.p.CurrentDatabase()
	currentDbDesc, err := params.p.Descriptors().GetMutableDatabaseByName(
		params.ctx, params.p.txn, currentDatabase, tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(564826)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564827)
	}
	__antithesis_instrumentation__.Notify(564822)

	lCtx := newInternalLookupCtx(all.OrderedDescriptors(), currentDbDesc.ImmutableCopy().(catalog.DatabaseDescriptor))

	for _, oldRole := range n.normalizedOldRoles {
		__antithesis_instrumentation__.Notify(564828)

		for _, dbID := range lCtx.dbIDs {
			__antithesis_instrumentation__.Notify(564832)
			if IsOwner(lCtx.dbDescs[dbID], oldRole) {
				__antithesis_instrumentation__.Notify(564833)
				if err := n.reassignDatabaseOwner(lCtx.dbDescs[dbID], params); err != nil {
					__antithesis_instrumentation__.Notify(564834)
					return err
				} else {
					__antithesis_instrumentation__.Notify(564835)
				}
			} else {
				__antithesis_instrumentation__.Notify(564836)
			}
		}
		__antithesis_instrumentation__.Notify(564829)
		for _, schemaID := range lCtx.schemaIDs {
			__antithesis_instrumentation__.Notify(564837)
			if IsOwner(lCtx.schemaDescs[schemaID], oldRole) {
				__antithesis_instrumentation__.Notify(564838)

				if lCtx.schemaDescs[schemaID].GetName() == tree.PublicSchema {
					__antithesis_instrumentation__.Notify(564840)
					continue
				} else {
					__antithesis_instrumentation__.Notify(564841)
				}
				__antithesis_instrumentation__.Notify(564839)
				if err := n.reassignSchemaOwner(lCtx.schemaDescs[schemaID], currentDbDesc, params); err != nil {
					__antithesis_instrumentation__.Notify(564842)
					return err
				} else {
					__antithesis_instrumentation__.Notify(564843)
				}
			} else {
				__antithesis_instrumentation__.Notify(564844)
			}
		}
		__antithesis_instrumentation__.Notify(564830)

		for _, tbID := range lCtx.tbIDs {
			__antithesis_instrumentation__.Notify(564845)
			if IsOwner(lCtx.tbDescs[tbID], oldRole) {
				__antithesis_instrumentation__.Notify(564846)
				if err := n.reassignTableOwner(lCtx.tbDescs[tbID], params); err != nil {
					__antithesis_instrumentation__.Notify(564847)
					return err
				} else {
					__antithesis_instrumentation__.Notify(564848)
				}
			} else {
				__antithesis_instrumentation__.Notify(564849)
			}
		}
		__antithesis_instrumentation__.Notify(564831)
		for _, typID := range lCtx.typIDs {
			__antithesis_instrumentation__.Notify(564850)
			if IsOwner(lCtx.typDescs[typID], oldRole) && func() bool {
				__antithesis_instrumentation__.Notify(564851)
				return (lCtx.typDescs[typID].GetKind() != descpb.TypeDescriptor_ALIAS) == true
			}() == true {
				__antithesis_instrumentation__.Notify(564852)
				if err := n.reassignTypeOwner(lCtx.typDescs[typID], params); err != nil {
					__antithesis_instrumentation__.Notify(564853)
					return err
				} else {
					__antithesis_instrumentation__.Notify(564854)
				}
			} else {
				__antithesis_instrumentation__.Notify(564855)
			}
		}
	}
	__antithesis_instrumentation__.Notify(564823)
	return nil
}

func (n *reassignOwnedByNode) reassignDatabaseOwner(
	dbDesc catalog.DatabaseDescriptor, params runParams,
) error {
	__antithesis_instrumentation__.Notify(564856)
	mutableDbDesc, err := params.p.Descriptors().GetMutableDescriptorByID(params.ctx, params.p.txn, dbDesc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(564861)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564862)
	}
	__antithesis_instrumentation__.Notify(564857)
	owner, err := n.n.NewRole.ToSQLUsername(params.p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(564863)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564864)
	}
	__antithesis_instrumentation__.Notify(564858)
	if err := params.p.setNewDatabaseOwner(params.ctx, mutableDbDesc, owner); err != nil {
		__antithesis_instrumentation__.Notify(564865)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564866)
	}
	__antithesis_instrumentation__.Notify(564859)
	if err := params.p.writeNonDropDatabaseChange(
		params.ctx,
		mutableDbDesc.(*dbdesc.Mutable),
		tree.AsStringWithFQNames(n.n, params.p.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(564867)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564868)
	}
	__antithesis_instrumentation__.Notify(564860)
	return nil
}

func (n *reassignOwnedByNode) reassignSchemaOwner(
	schemaDesc catalog.SchemaDescriptor, dbDesc *dbdesc.Mutable, params runParams,
) error {
	__antithesis_instrumentation__.Notify(564869)
	mutableSchemaDesc, err := params.p.Descriptors().GetMutableDescriptorByID(params.ctx, params.p.txn, schemaDesc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(564874)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564875)
	}
	__antithesis_instrumentation__.Notify(564870)
	owner, err := n.n.NewRole.ToSQLUsername(params.p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(564876)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564877)
	}
	__antithesis_instrumentation__.Notify(564871)
	if err := params.p.setNewSchemaOwner(
		params.ctx, dbDesc, mutableSchemaDesc.(*schemadesc.Mutable), owner); err != nil {
		__antithesis_instrumentation__.Notify(564878)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564879)
	}
	__antithesis_instrumentation__.Notify(564872)
	if err := params.p.writeSchemaDescChange(params.ctx,
		mutableSchemaDesc.(*schemadesc.Mutable),
		tree.AsStringWithFQNames(n.n, params.p.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(564880)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564881)
	}
	__antithesis_instrumentation__.Notify(564873)
	return nil
}

func (n *reassignOwnedByNode) reassignTableOwner(
	tbDesc catalog.TableDescriptor, params runParams,
) error {
	__antithesis_instrumentation__.Notify(564882)
	mutableTbDesc, err := params.p.Descriptors().GetMutableDescriptorByID(params.ctx, params.p.txn, tbDesc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(564888)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564889)
	}
	__antithesis_instrumentation__.Notify(564883)

	tableName, err := params.p.getQualifiedTableName(params.ctx, tbDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(564890)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564891)
	}
	__antithesis_instrumentation__.Notify(564884)

	owner, err := n.n.NewRole.ToSQLUsername(params.p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(564892)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564893)
	}
	__antithesis_instrumentation__.Notify(564885)
	if err := params.p.setNewTableOwner(
		params.ctx, mutableTbDesc.(*tabledesc.Mutable), *tableName, owner); err != nil {
		__antithesis_instrumentation__.Notify(564894)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564895)
	}
	__antithesis_instrumentation__.Notify(564886)
	if err := params.p.writeSchemaChange(
		params.ctx, mutableTbDesc.(*tabledesc.Mutable), descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(564896)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564897)
	}
	__antithesis_instrumentation__.Notify(564887)
	return nil
}

func (n *reassignOwnedByNode) reassignTypeOwner(
	typDesc catalog.TypeDescriptor, params runParams,
) error {
	__antithesis_instrumentation__.Notify(564898)
	mutableTypDesc, err := params.p.Descriptors().GetMutableDescriptorByID(params.ctx, params.p.txn, typDesc.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(564907)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564908)
	}
	__antithesis_instrumentation__.Notify(564899)
	arrayDesc, err := params.p.Descriptors().GetMutableTypeVersionByID(
		params.ctx, params.p.txn, typDesc.GetArrayTypeID())
	if err != nil {
		__antithesis_instrumentation__.Notify(564909)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564910)
	}
	__antithesis_instrumentation__.Notify(564900)

	typeName, err := params.p.getQualifiedTypeName(params.ctx, mutableTypDesc.(*typedesc.Mutable))
	if err != nil {
		__antithesis_instrumentation__.Notify(564911)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564912)
	}
	__antithesis_instrumentation__.Notify(564901)
	arrayTypeName, err := params.p.getQualifiedTypeName(params.ctx, arrayDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(564913)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564914)
	}
	__antithesis_instrumentation__.Notify(564902)

	owner, err := n.n.NewRole.ToSQLUsername(params.p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(564915)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564916)
	}
	__antithesis_instrumentation__.Notify(564903)
	if err := params.p.setNewTypeOwner(
		params.ctx, mutableTypDesc.(*typedesc.Mutable), arrayDesc, *typeName,
		*arrayTypeName, owner); err != nil {
		__antithesis_instrumentation__.Notify(564917)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564918)
	}
	__antithesis_instrumentation__.Notify(564904)
	if err := params.p.writeTypeSchemaChange(
		params.ctx, mutableTypDesc.(*typedesc.Mutable), tree.AsStringWithFQNames(n.n, params.p.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(564919)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564920)
	}
	__antithesis_instrumentation__.Notify(564905)
	if err := params.p.writeTypeSchemaChange(
		params.ctx, arrayDesc, tree.AsStringWithFQNames(n.n, params.p.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(564921)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564922)
	}
	__antithesis_instrumentation__.Notify(564906)
	return nil
}

func (n *reassignOwnedByNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(564923)
	return false, nil
}
func (n *reassignOwnedByNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(564924)
	return tree.Datums{}
}
func (n *reassignOwnedByNode) Close(context.Context) { __antithesis_instrumentation__.Notify(564925) }
