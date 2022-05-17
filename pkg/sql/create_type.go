package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/enum"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type createTypeNode struct {
	n        *tree.CreateType
	typeName *tree.TypeName
	dbDesc   catalog.DatabaseDescriptor
}

type EnumType int

const (
	EnumTypeUserDefined = iota

	EnumTypeMultiRegion
)

var _ planNode = &createTypeNode{n: nil}

func (p *planner) CreateType(ctx context.Context, n *tree.CreateType) (planNode, error) {
	__antithesis_instrumentation__.Notify(464902)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"CREATE TYPE",
	); err != nil {
		__antithesis_instrumentation__.Notify(464905)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464906)
	}
	__antithesis_instrumentation__.Notify(464903)

	typeName, db, err := resolveNewTypeName(p.RunParams(ctx), n.TypeName)
	if err != nil {
		__antithesis_instrumentation__.Notify(464907)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464908)
	}
	__antithesis_instrumentation__.Notify(464904)
	n.TypeName.SetAnnotation(&p.semaCtx.Annotations, typeName)
	return &createTypeNode{
		n:        n,
		typeName: typeName,
		dbDesc:   db,
	}, nil
}

func (n *createTypeNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(464909)

	flags := tree.ObjectLookupFlags{CommonLookupFlags: tree.CommonLookupFlags{
		Required:    false,
		AvoidLeased: true,
	}}
	found, _, err := params.p.Descriptors().GetImmutableTypeByName(params.ctx, params.p.Txn(), n.typeName, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(464912)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464913)
	}
	__antithesis_instrumentation__.Notify(464910)

	if found && func() bool {
		__antithesis_instrumentation__.Notify(464914)
		return n.n.IfNotExists == true
	}() == true {
		__antithesis_instrumentation__.Notify(464915)
		params.p.BufferClientNotice(
			params.ctx,
			pgnotice.Newf("type %q already exists, skipping", n.typeName),
		)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(464916)
	}
	__antithesis_instrumentation__.Notify(464911)

	switch n.n.Variety {
	case tree.Enum:
		__antithesis_instrumentation__.Notify(464917)
		return params.p.createUserDefinedEnum(params, n)
	default:
		__antithesis_instrumentation__.Notify(464918)
		return unimplemented.NewWithIssue(25123, "CREATE TYPE")
	}
}

func resolveNewTypeName(
	params runParams, name *tree.UnresolvedObjectName,
) (*tree.TypeName, catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(464919)

	db, _, prefix, err := params.p.ResolveTargetObject(params.ctx, name)
	if err != nil {
		__antithesis_instrumentation__.Notify(464923)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(464924)
	}
	__antithesis_instrumentation__.Notify(464920)

	if err := params.p.CheckPrivilege(params.ctx, db, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(464925)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(464926)
	}
	__antithesis_instrumentation__.Notify(464921)

	if db.GetID() == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(464927)
		return nil, nil, errors.New("cannot create a type in the system database")
	} else {
		__antithesis_instrumentation__.Notify(464928)
	}
	__antithesis_instrumentation__.Notify(464922)

	typename := tree.NewUnqualifiedTypeName(name.Object())
	typename.ObjectNamePrefix = prefix
	return typename, db, nil
}

func getCreateTypeParams(
	params runParams, name *tree.TypeName, db catalog.DatabaseDescriptor,
) (schema catalog.SchemaDescriptor, err error) {
	__antithesis_instrumentation__.Notify(464929)

	if name.Schema() == tree.PublicSchema {
		__antithesis_instrumentation__.Notify(464935)
		if _, ok := types.PublicSchemaAliases[name.Object()]; ok {
			__antithesis_instrumentation__.Notify(464936)
			return nil, sqlerrors.NewTypeAlreadyExistsError(name.String())
		} else {
			__antithesis_instrumentation__.Notify(464937)
		}
	} else {
		__antithesis_instrumentation__.Notify(464938)
	}
	__antithesis_instrumentation__.Notify(464930)

	dbID := db.GetID()
	schema, err = params.p.getNonTemporarySchemaForCreate(params.ctx, db, name.Schema())
	if err != nil {
		__antithesis_instrumentation__.Notify(464939)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464940)
	}
	__antithesis_instrumentation__.Notify(464931)

	if err := params.p.canCreateOnSchema(
		params.ctx, schema.GetID(), dbID, params.p.User(), skipCheckPublicSchema); err != nil {
		__antithesis_instrumentation__.Notify(464941)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464942)
	}
	__antithesis_instrumentation__.Notify(464932)

	if schema.SchemaKind() == catalog.SchemaUserDefined {
		__antithesis_instrumentation__.Notify(464943)
		sqltelemetry.IncrementUserDefinedSchemaCounter(sqltelemetry.UserDefinedSchemaUsedByObject)
	} else {
		__antithesis_instrumentation__.Notify(464944)
	}
	__antithesis_instrumentation__.Notify(464933)

	err = params.p.Descriptors().Direct().CheckObjectCollision(
		params.ctx,
		params.p.txn,
		db.GetID(),
		schema.GetID(),
		name,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(464945)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(464946)
	}
	__antithesis_instrumentation__.Notify(464934)

	return schema, nil
}

func findFreeArrayTypeName(
	ctx context.Context,
	txn *kv.Txn,
	col *descs.Collection,
	parentID, schemaID descpb.ID,
	name string,
) (string, error) {
	__antithesis_instrumentation__.Notify(464947)
	arrayName := "_" + name
	for {
		__antithesis_instrumentation__.Notify(464949)

		objectID, err := col.Direct().LookupObjectID(ctx, txn, parentID, schemaID, arrayName)
		if err != nil {
			__antithesis_instrumentation__.Notify(464952)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(464953)
		}
		__antithesis_instrumentation__.Notify(464950)

		if objectID == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(464954)
			break
		} else {
			__antithesis_instrumentation__.Notify(464955)
		}
		__antithesis_instrumentation__.Notify(464951)

		arrayName = "_" + arrayName
	}
	__antithesis_instrumentation__.Notify(464948)
	return arrayName, nil
}

func CreateEnumArrayTypeDesc(
	params runParams,
	typDesc *typedesc.Mutable,
	db catalog.DatabaseDescriptor,
	schemaID descpb.ID,
	id descpb.ID,
	arrayTypeName string,
) (*typedesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(464956)

	var elemTyp *types.T
	switch t := typDesc.Kind; t {
	case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(464958)
		elemTyp = types.MakeEnum(typedesc.TypeIDToOID(typDesc.GetID()), typedesc.TypeIDToOID(id))
	default:
		__antithesis_instrumentation__.Notify(464959)
		return nil, errors.AssertionFailedf("cannot make array type for kind %s", t.String())
	}
	__antithesis_instrumentation__.Notify(464957)

	return typedesc.NewBuilder(&descpb.TypeDescriptor{
		Name:           arrayTypeName,
		ID:             id,
		ParentID:       db.GetID(),
		ParentSchemaID: schemaID,
		Kind:           descpb.TypeDescriptor_ALIAS,
		Alias:          types.MakeArray(elemTyp),
		Version:        1,
		Privileges:     typDesc.Privileges,
	}).BuildCreatedMutableType(), nil
}

func (p *planner) createArrayType(
	params runParams,
	typ *tree.TypeName,
	typDesc *typedesc.Mutable,
	db catalog.DatabaseDescriptor,
	schemaID descpb.ID,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(464960)
	arrayTypeName, err := findFreeArrayTypeName(
		params.ctx,
		params.p.txn,
		params.p.Descriptors(),
		db.GetID(),
		schemaID,
		typ.Type(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(464965)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(464966)
	}
	__antithesis_instrumentation__.Notify(464961)
	arrayTypeKey := catalogkeys.MakeObjectNameKey(params.ExecCfg().Codec, db.GetID(), schemaID, arrayTypeName)

	id, err := descidgen.GenerateUniqueDescID(params.ctx, params.ExecCfg().DB, params.ExecCfg().Codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(464967)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(464968)
	}
	__antithesis_instrumentation__.Notify(464962)

	arrayTypDesc, err := CreateEnumArrayTypeDesc(
		params,
		typDesc,
		db,
		schemaID,
		id,
		arrayTypeName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(464969)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(464970)
	}
	__antithesis_instrumentation__.Notify(464963)

	jobStr := fmt.Sprintf("implicit array type creation for %s", typ)
	if err := p.createDescriptorWithID(
		params.ctx,
		arrayTypeKey,
		id,
		arrayTypDesc,
		jobStr,
	); err != nil {
		__antithesis_instrumentation__.Notify(464971)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(464972)
	}
	__antithesis_instrumentation__.Notify(464964)
	return id, nil
}

func (p *planner) createUserDefinedEnum(params runParams, n *createTypeNode) error {
	__antithesis_instrumentation__.Notify(464973)

	id, err := descidgen.GenerateUniqueDescID(
		params.ctx, params.ExecCfg().DB, params.ExecCfg().Codec,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(464975)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464976)
	}
	__antithesis_instrumentation__.Notify(464974)
	return params.p.createEnumWithID(
		params, id, n.n.EnumLabels, n.dbDesc, n.typeName, EnumTypeUserDefined,
	)
}

func CreateEnumTypeDesc(
	params runParams,
	id descpb.ID,
	enumLabels tree.EnumValueList,
	dbDesc catalog.DatabaseDescriptor,
	schema catalog.SchemaDescriptor,
	typeName *tree.TypeName,
	enumType EnumType,
) (*typedesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(464977)

	seenVals := make(map[tree.EnumValue]struct{})
	for _, value := range enumLabels {
		__antithesis_instrumentation__.Notify(464981)
		_, ok := seenVals[value]
		if ok {
			__antithesis_instrumentation__.Notify(464983)
			return nil, pgerror.Newf(pgcode.InvalidObjectDefinition,
				"enum definition contains duplicate value %q", value)
		} else {
			__antithesis_instrumentation__.Notify(464984)
		}
		__antithesis_instrumentation__.Notify(464982)
		seenVals[value] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(464978)

	members := make([]descpb.TypeDescriptor_EnumMember, len(enumLabels))
	physReps := enum.GenerateNEvenlySpacedBytes(len(enumLabels))
	for i := range enumLabels {
		__antithesis_instrumentation__.Notify(464985)
		members[i] = descpb.TypeDescriptor_EnumMember{
			LogicalRepresentation:  string(enumLabels[i]),
			PhysicalRepresentation: physReps[i],
			Capability:             descpb.TypeDescriptor_EnumMember_ALL,
		}
	}
	__antithesis_instrumentation__.Notify(464979)

	privs := catprivilege.CreatePrivilegesFromDefaultPrivileges(
		dbDesc.GetDefaultPrivilegeDescriptor(),
		schema.GetDefaultPrivilegeDescriptor(),
		dbDesc.GetID(),
		params.SessionData().User(),
		tree.Types,
		dbDesc.GetPrivileges(),
	)

	enumKind := descpb.TypeDescriptor_ENUM
	var regionConfig *descpb.TypeDescriptor_RegionConfig
	if enumType == EnumTypeMultiRegion {
		__antithesis_instrumentation__.Notify(464986)
		enumKind = descpb.TypeDescriptor_MULTIREGION_ENUM
		primaryRegion, err := dbDesc.PrimaryRegionName()
		if err != nil {
			__antithesis_instrumentation__.Notify(464988)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(464989)
		}
		__antithesis_instrumentation__.Notify(464987)
		regionConfig = &descpb.TypeDescriptor_RegionConfig{
			PrimaryRegion: primaryRegion,
		}
	} else {
		__antithesis_instrumentation__.Notify(464990)
	}
	__antithesis_instrumentation__.Notify(464980)

	return typedesc.NewBuilder(&descpb.TypeDescriptor{
		Name:           typeName.Type(),
		ID:             id,
		ParentID:       dbDesc.GetID(),
		ParentSchemaID: schema.GetID(),
		Kind:           enumKind,
		EnumMembers:    members,
		Version:        1,
		Privileges:     privs,
		RegionConfig:   regionConfig,
	}).BuildCreatedMutableType(), nil
}

func (p *planner) createEnumWithID(
	params runParams,
	id descpb.ID,
	enumLabels tree.EnumValueList,
	dbDesc catalog.DatabaseDescriptor,
	typeName *tree.TypeName,
	enumType EnumType,
) error {
	__antithesis_instrumentation__.Notify(464991)
	sqltelemetry.IncrementEnumCounter(sqltelemetry.EnumCreate)

	schema, err := getCreateTypeParams(params, typeName, dbDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(464996)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464997)
	}
	__antithesis_instrumentation__.Notify(464992)

	typeDesc, err := CreateEnumTypeDesc(params, id, enumLabels, dbDesc, schema, typeName, enumType)
	if err != nil {
		__antithesis_instrumentation__.Notify(464998)
		return err
	} else {
		__antithesis_instrumentation__.Notify(464999)
	}
	__antithesis_instrumentation__.Notify(464993)

	arrayTypeID, err := p.createArrayType(params, typeName, typeDesc, dbDesc, schema.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(465000)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465001)
	}
	__antithesis_instrumentation__.Notify(464994)

	typeDesc.ArrayTypeID = arrayTypeID

	if err := p.createDescriptorWithID(
		params.ctx,
		catalogkeys.MakeObjectNameKey(params.ExecCfg().Codec, dbDesc.GetID(), schema.GetID(), typeName.Type()),
		id,
		typeDesc,
		typeName.String(),
	); err != nil {
		__antithesis_instrumentation__.Notify(465002)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465003)
	}
	__antithesis_instrumentation__.Notify(464995)

	return p.logEvent(params.ctx,
		typeDesc.GetID(),
		&eventpb.CreateType{
			TypeName: typeName.FQString(),
		})
}

func (n *createTypeNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(465004)
	return false, nil
}
func (n *createTypeNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(465005)
	return tree.Datums{}
}
func (n *createTypeNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(465006) }
func (n *createTypeNode) ReadingOwnWrites()         { __antithesis_instrumentation__.Notify(465007) }
