package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
)

type createSchemaNode struct {
	n *tree.CreateSchema
}

func (n *createSchemaNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(463374)
	return params.p.createUserDefinedSchema(params, n.n)
}

func CreateUserDefinedSchemaDescriptor(
	ctx context.Context,
	sessionData *sessiondata.SessionData,
	n *tree.CreateSchema,
	txn *kv.Txn,
	descriptors *descs.Collection,
	execCfg *ExecutorConfig,
	db catalog.DatabaseDescriptor,
	allocateID bool,
) (*schemadesc.Mutable, *catpb.PrivilegeDescriptor, error) {
	__antithesis_instrumentation__.Notify(463375)
	authRole, err := n.AuthRole.ToSQLUsername(sessionData, security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(463383)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(463384)
	}
	__antithesis_instrumentation__.Notify(463376)
	user := sessionData.User()
	var schemaName string
	if !n.Schema.ExplicitSchema {
		__antithesis_instrumentation__.Notify(463385)
		schemaName = authRole.Normalized()
	} else {
		__antithesis_instrumentation__.Notify(463386)
		schemaName = n.Schema.Schema()
	}
	__antithesis_instrumentation__.Notify(463377)

	exists, schemaID, err := schemaExists(ctx, txn, descriptors, db.GetID(), schemaName)
	if err != nil {
		__antithesis_instrumentation__.Notify(463387)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(463388)
	}
	__antithesis_instrumentation__.Notify(463378)

	if exists {
		__antithesis_instrumentation__.Notify(463389)
		if n.IfNotExists {
			__antithesis_instrumentation__.Notify(463391)

			if schemaID != descpb.InvalidID {
				__antithesis_instrumentation__.Notify(463393)

				sc, err := descriptors.GetImmutableSchemaByID(ctx, txn, schemaID, tree.SchemaLookupFlags{
					Required:       true,
					AvoidLeased:    true,
					IncludeOffline: true,
					IncludeDropped: true,
				})
				if err != nil || func() bool {
					__antithesis_instrumentation__.Notify(463395)
					return sc.SchemaKind() != catalog.SchemaUserDefined == true
				}() == true {
					__antithesis_instrumentation__.Notify(463396)
					return nil, nil, err
				} else {
					__antithesis_instrumentation__.Notify(463397)
				}
				__antithesis_instrumentation__.Notify(463394)
				if sc.Dropped() {
					__antithesis_instrumentation__.Notify(463398)
					return nil, nil, pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"schema %q is being dropped, try again later",
						schemaName)
				} else {
					__antithesis_instrumentation__.Notify(463399)
				}
			} else {
				__antithesis_instrumentation__.Notify(463400)
			}
			__antithesis_instrumentation__.Notify(463392)
			return nil, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(463401)
		}
		__antithesis_instrumentation__.Notify(463390)
		return nil, nil, sqlerrors.NewSchemaAlreadyExistsError(schemaName)
	} else {
		__antithesis_instrumentation__.Notify(463402)
	}
	__antithesis_instrumentation__.Notify(463379)

	if err := schemadesc.IsSchemaNameValid(schemaName); err != nil {
		__antithesis_instrumentation__.Notify(463403)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(463404)
	}
	__antithesis_instrumentation__.Notify(463380)

	owner := user
	if !n.AuthRole.Undefined() {
		__antithesis_instrumentation__.Notify(463405)
		exists, err := RoleExists(ctx, execCfg, txn, authRole)
		if err != nil {
			__antithesis_instrumentation__.Notify(463408)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(463409)
		}
		__antithesis_instrumentation__.Notify(463406)
		if !exists {
			__antithesis_instrumentation__.Notify(463410)
			return nil, nil, pgerror.Newf(pgcode.UndefinedObject, "role/user %q does not exist",
				n.AuthRole)
		} else {
			__antithesis_instrumentation__.Notify(463411)
		}
		__antithesis_instrumentation__.Notify(463407)
		owner = authRole
	} else {
		__antithesis_instrumentation__.Notify(463412)
	}
	__antithesis_instrumentation__.Notify(463381)

	desc, privs, err := CreateSchemaDescriptorWithPrivileges(ctx, execCfg.DB, execCfg.Codec, db, schemaName, user, owner, allocateID)
	if err != nil {
		__antithesis_instrumentation__.Notify(463413)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(463414)
	}
	__antithesis_instrumentation__.Notify(463382)

	return desc, privs, nil
}

func CreateSchemaDescriptorWithPrivileges(
	ctx context.Context,
	kvDB *kv.DB,
	codec keys.SQLCodec,
	db catalog.DatabaseDescriptor,
	schemaName string,
	user, owner security.SQLUsername,
	allocateID bool,
) (*schemadesc.Mutable, *catpb.PrivilegeDescriptor, error) {
	__antithesis_instrumentation__.Notify(463415)

	var id descpb.ID
	var err error
	if allocateID {
		__antithesis_instrumentation__.Notify(463417)
		id, err = descidgen.GenerateUniqueDescID(ctx, kvDB, codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(463418)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(463419)
		}
	} else {
		__antithesis_instrumentation__.Notify(463420)
	}
	__antithesis_instrumentation__.Notify(463416)

	privs := catprivilege.CreatePrivilegesFromDefaultPrivileges(
		db.GetDefaultPrivilegeDescriptor(),
		nil,
		db.GetID(),
		user,
		tree.Schemas,
		db.GetPrivileges(),
	)

	privs.SetOwner(owner)

	desc := schemadesc.NewBuilder(&descpb.SchemaDescriptor{
		ParentID:   db.GetID(),
		Name:       schemaName,
		ID:         id,
		Privileges: privs,
		Version:    1,
	}).BuildCreatedMutableSchema()

	return desc, privs, nil
}

func (p *planner) createUserDefinedSchema(params runParams, n *tree.CreateSchema) error {
	__antithesis_instrumentation__.Notify(463421)
	if err := checkSchemaChangeEnabled(
		p.EvalContext().Context,
		p.ExecCfg(),
		"CREATE SCHEMA",
	); err != nil {
		__antithesis_instrumentation__.Notify(463433)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463434)
	}
	__antithesis_instrumentation__.Notify(463422)

	if p.CurrentDatabase() == "" {
		__antithesis_instrumentation__.Notify(463435)
		return pgerror.New(pgcode.UndefinedDatabase,
			"cannot create schema without being connected to a database")
	} else {
		__antithesis_instrumentation__.Notify(463436)
	}
	__antithesis_instrumentation__.Notify(463423)

	sqltelemetry.IncrementUserDefinedSchemaCounter(sqltelemetry.UserDefinedSchemaCreate)
	dbName := p.CurrentDatabase()
	if n.Schema.ExplicitCatalog {
		__antithesis_instrumentation__.Notify(463437)
		dbName = n.Schema.Catalog()
	} else {
		__antithesis_instrumentation__.Notify(463438)
	}
	__antithesis_instrumentation__.Notify(463424)

	db, err := p.Descriptors().GetMutableDatabaseByName(params.ctx, p.txn, dbName,
		tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(463439)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463440)
	}
	__antithesis_instrumentation__.Notify(463425)

	if db.ID == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(463441)
		return pgerror.New(pgcode.InvalidObjectDefinition, "cannot create schemas in the system database")
	} else {
		__antithesis_instrumentation__.Notify(463442)
	}
	__antithesis_instrumentation__.Notify(463426)

	if err := p.CheckPrivilege(params.ctx, db, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(463443)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463444)
	}
	__antithesis_instrumentation__.Notify(463427)

	desc, privs, err := CreateUserDefinedSchemaDescriptor(params.ctx, params.SessionData(), n,
		p.Txn(), p.Descriptors(), p.ExecCfg(), db, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(463445)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463446)
	}
	__antithesis_instrumentation__.Notify(463428)

	if desc == nil {
		__antithesis_instrumentation__.Notify(463447)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(463448)
	}
	__antithesis_instrumentation__.Notify(463429)

	db.AddSchemaToDatabase(desc.Name, descpb.DatabaseDescriptor_SchemaInfo{ID: desc.ID})

	if err := p.writeNonDropDatabaseChange(
		params.ctx, db,
		fmt.Sprintf("updating parent database %s for %s", db.GetName(), tree.AsStringWithFQNames(n, params.Ann())),
	); err != nil {
		__antithesis_instrumentation__.Notify(463449)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463450)
	}
	__antithesis_instrumentation__.Notify(463430)

	if err := p.createDescriptorWithID(
		params.ctx,
		catalogkeys.MakeSchemaNameKey(p.ExecCfg().Codec, db.ID, desc.Name),
		desc.ID,
		desc,
		tree.AsStringWithFQNames(n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(463451)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463452)
	}
	__antithesis_instrumentation__.Notify(463431)

	qualifiedSchemaName, err := p.getQualifiedSchemaName(params.ctx, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(463453)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463454)
	}
	__antithesis_instrumentation__.Notify(463432)

	return params.p.logEvent(params.ctx,
		desc.GetID(),
		&eventpb.CreateSchema{
			SchemaName: qualifiedSchemaName.String(),
			Owner:      privs.Owner().Normalized(),
		})
}

func (*createSchemaNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(463455)
	return false, nil
}
func (*createSchemaNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(463456)
	return tree.Datums{}
}
func (n *createSchemaNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(463457) }

func (p *planner) CreateSchema(ctx context.Context, n *tree.CreateSchema) (planNode, error) {
	__antithesis_instrumentation__.Notify(463458)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"CREATE SCHEMA",
	); err != nil {
		__antithesis_instrumentation__.Notify(463460)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463461)
	}
	__antithesis_instrumentation__.Notify(463459)

	return &createSchemaNode{
		n: n,
	}, nil
}
