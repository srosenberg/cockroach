package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type alterTypeNode struct {
	n      *tree.AlterType
	prefix catalog.ResolvedObjectPrefix
	desc   *typedesc.Mutable
}

var _ planNode = &alterTypeNode{n: nil}

func (p *planner) AlterType(ctx context.Context, n *tree.AlterType) (planNode, error) {
	__antithesis_instrumentation__.Notify(245049)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER TYPE",
	); err != nil {
		__antithesis_instrumentation__.Notify(245054)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245055)
	}
	__antithesis_instrumentation__.Notify(245050)

	prefix, desc, err := p.ResolveMutableTypeDescriptor(ctx, n.Type, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(245056)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245057)
	}
	__antithesis_instrumentation__.Notify(245051)

	if err := p.canModifyType(ctx, desc); err != nil {
		__antithesis_instrumentation__.Notify(245058)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245059)
	}
	__antithesis_instrumentation__.Notify(245052)

	switch desc.Kind {
	case descpb.TypeDescriptor_ALIAS:
		__antithesis_instrumentation__.Notify(245060)

		return nil, pgerror.Newf(
			pgcode.WrongObjectType,
			"%q is an implicit array type and cannot be modified",
			tree.AsStringWithFQNames(n.Type, &p.semaCtx.Annotations),
		)
	case descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(245061)

		if _, isAlterTypeOwner := n.Cmd.(*tree.AlterTypeOwner); !isAlterTypeOwner {
			__antithesis_instrumentation__.Notify(245065)
			return nil, errors.WithHint(
				pgerror.Newf(
					pgcode.WrongObjectType,
					"%q is a multi-region enum and can't be modified using the alter type command",
					tree.AsStringWithFQNames(n.Type, &p.semaCtx.Annotations)),
				"try adding/removing the region using ALTER DATABASE")
		} else {
			__antithesis_instrumentation__.Notify(245066)
		}
	case descpb.TypeDescriptor_ENUM:
		__antithesis_instrumentation__.Notify(245062)
		sqltelemetry.IncrementEnumCounter(sqltelemetry.EnumAlter)
	case descpb.TypeDescriptor_TABLE_IMPLICIT_RECORD_TYPE:
		__antithesis_instrumentation__.Notify(245063)
		return nil, pgerror.Newf(
			pgcode.WrongObjectType,
			"%q is a table's record type and cannot be modified",
			tree.AsStringWithFQNames(n.Type, &p.semaCtx.Annotations),
		)
	default:
		__antithesis_instrumentation__.Notify(245064)
	}
	__antithesis_instrumentation__.Notify(245053)

	return &alterTypeNode{
		n:      n,
		prefix: prefix,
		desc:   desc,
	}, nil
}

func (n *alterTypeNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(245067)
	telemetry.Inc(n.n.Cmd.TelemetryCounter())

	typeName := tree.AsStringWithFQNames(n.n.Type, params.p.Ann())
	eventLogDone := false
	var err error
	switch t := n.n.Cmd.(type) {
	case *tree.AlterTypeAddValue:
		__antithesis_instrumentation__.Notify(245071)
		err = params.p.addEnumValue(params.ctx, n.desc, t, tree.AsStringWithFQNames(n.n, params.p.Ann()))
	case *tree.AlterTypeRenameValue:
		__antithesis_instrumentation__.Notify(245072)
		err = params.p.renameTypeValue(params.ctx, n, string(t.OldVal), string(t.NewVal))
	case *tree.AlterTypeRename:
		__antithesis_instrumentation__.Notify(245073)
		if err = params.p.renameType(params.ctx, n, string(t.NewName)); err != nil {
			__antithesis_instrumentation__.Notify(245081)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245082)
		}
		__antithesis_instrumentation__.Notify(245074)
		err = params.p.logEvent(params.ctx, n.desc.ID, &eventpb.RenameType{
			TypeName:    typeName,
			NewTypeName: string(t.NewName),
		})
		eventLogDone = true
	case *tree.AlterTypeSetSchema:
		__antithesis_instrumentation__.Notify(245075)

		err = params.p.setTypeSchema(params.ctx, n, string(t.Schema))
	case *tree.AlterTypeOwner:
		__antithesis_instrumentation__.Notify(245076)
		owner, err := t.Owner.ToSQLUsername(params.SessionData(), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(245083)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245084)
		}
		__antithesis_instrumentation__.Notify(245077)
		if err = params.p.alterTypeOwner(params.ctx, n, owner); err != nil {
			__antithesis_instrumentation__.Notify(245085)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245086)
		}
		__antithesis_instrumentation__.Notify(245078)
		eventLogDone = true
	case *tree.AlterTypeDropValue:
		__antithesis_instrumentation__.Notify(245079)
		err = params.p.dropEnumValue(params.ctx, n.desc, t.Val)
	default:
		__antithesis_instrumentation__.Notify(245080)
		err = errors.AssertionFailedf("unknown alter type cmd %s", t)
	}
	__antithesis_instrumentation__.Notify(245068)
	if err != nil {
		__antithesis_instrumentation__.Notify(245087)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245088)
	}
	__antithesis_instrumentation__.Notify(245069)

	if !eventLogDone {
		__antithesis_instrumentation__.Notify(245089)

		if err := params.p.logEvent(params.ctx,
			n.desc.ID,
			&eventpb.AlterType{
				TypeName: typeName,
			}); err != nil {
			__antithesis_instrumentation__.Notify(245090)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245091)
		}
	} else {
		__antithesis_instrumentation__.Notify(245092)
	}
	__antithesis_instrumentation__.Notify(245070)
	return nil
}

func findEnumMemberByName(
	desc *typedesc.Mutable, val tree.EnumValue,
) (bool, *descpb.TypeDescriptor_EnumMember) {
	__antithesis_instrumentation__.Notify(245093)
	for _, member := range desc.EnumMembers {
		__antithesis_instrumentation__.Notify(245095)
		if member.LogicalRepresentation == string(val) {
			__antithesis_instrumentation__.Notify(245096)
			return true, &member
		} else {
			__antithesis_instrumentation__.Notify(245097)
		}
	}
	__antithesis_instrumentation__.Notify(245094)
	return false, nil
}

func (p *planner) addEnumValue(
	ctx context.Context, desc *typedesc.Mutable, node *tree.AlterTypeAddValue, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(245098)
	if desc.Kind != descpb.TypeDescriptor_ENUM && func() bool {
		__antithesis_instrumentation__.Notify(245102)
		return desc.Kind != descpb.TypeDescriptor_MULTIREGION_ENUM == true
	}() == true {
		__antithesis_instrumentation__.Notify(245103)
		return pgerror.Newf(pgcode.WrongObjectType, "%q is not an enum", desc.Name)
	} else {
		__antithesis_instrumentation__.Notify(245104)
	}
	__antithesis_instrumentation__.Notify(245099)

	found, member := findEnumMemberByName(desc, node.NewVal)
	if found {
		__antithesis_instrumentation__.Notify(245105)
		if enumMemberIsRemoving(member) {
			__antithesis_instrumentation__.Notify(245108)
			return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"enum value %q is being dropped, try again later", node.NewVal)
		} else {
			__antithesis_instrumentation__.Notify(245109)
		}
		__antithesis_instrumentation__.Notify(245106)
		if node.IfNotExists {
			__antithesis_instrumentation__.Notify(245110)
			p.BufferClientNotice(
				ctx,
				pgnotice.Newf("enum value %q already exists, skipping", node.NewVal),
			)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(245111)
		}
		__antithesis_instrumentation__.Notify(245107)
		return pgerror.Newf(pgcode.DuplicateObject, "enum value %q already exists", node.NewVal)
	} else {
		__antithesis_instrumentation__.Notify(245112)
	}
	__antithesis_instrumentation__.Notify(245100)

	if err := desc.AddEnumValue(node); err != nil {
		__antithesis_instrumentation__.Notify(245113)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245114)
	}
	__antithesis_instrumentation__.Notify(245101)
	return p.writeTypeSchemaChange(ctx, desc, jobDesc)
}

func (p *planner) dropEnumValue(
	ctx context.Context, desc *typedesc.Mutable, val tree.EnumValue,
) error {
	__antithesis_instrumentation__.Notify(245115)
	if desc.Kind != descpb.TypeDescriptor_ENUM && func() bool {
		__antithesis_instrumentation__.Notify(245120)
		return desc.Kind != descpb.TypeDescriptor_MULTIREGION_ENUM == true
	}() == true {
		__antithesis_instrumentation__.Notify(245121)
		return pgerror.Newf(pgcode.WrongObjectType, "%q is not an enum", desc.Name)
	} else {
		__antithesis_instrumentation__.Notify(245122)
	}
	__antithesis_instrumentation__.Notify(245116)

	found, member := findEnumMemberByName(desc, val)
	if !found {
		__antithesis_instrumentation__.Notify(245123)
		return pgerror.Newf(pgcode.UndefinedObject, "enum value %q does not exist", val)
	} else {
		__antithesis_instrumentation__.Notify(245124)
	}
	__antithesis_instrumentation__.Notify(245117)

	if enumMemberIsRemoving(member) {
		__antithesis_instrumentation__.Notify(245125)
		return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
			"enum value %q is already being dropped", val)
	} else {
		__antithesis_instrumentation__.Notify(245126)
	}
	__antithesis_instrumentation__.Notify(245118)
	if enumMemberIsAdding(member) {
		__antithesis_instrumentation__.Notify(245127)
		return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
			"enum value %q is being added, try again later", val)
	} else {
		__antithesis_instrumentation__.Notify(245128)
	}
	__antithesis_instrumentation__.Notify(245119)

	desc.DropEnumValue(val)
	return p.writeTypeSchemaChange(ctx, desc, desc.Name)
}

func (p *planner) renameType(ctx context.Context, n *alterTypeNode, newName string) error {
	__antithesis_instrumentation__.Notify(245129)
	err := p.Descriptors().Direct().CheckObjectCollision(
		ctx,
		p.txn,
		n.desc.ParentID,
		n.desc.ParentSchemaID,
		tree.NewUnqualifiedTypeName(newName),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245135)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245136)
	}
	__antithesis_instrumentation__.Notify(245130)

	if err := p.performRenameTypeDesc(
		ctx,
		n.desc,
		newName,
		n.desc.ParentSchemaID,
		tree.AsStringWithFQNames(n.n, p.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(245137)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245138)
	}
	__antithesis_instrumentation__.Notify(245131)

	newArrayName, err := findFreeArrayTypeName(
		ctx,
		p.txn,
		p.Descriptors(),
		n.desc.ParentID,
		n.desc.ParentSchemaID,
		newName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245139)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245140)
	}
	__antithesis_instrumentation__.Notify(245132)
	arrayDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, n.desc.ArrayTypeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(245141)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245142)
	}
	__antithesis_instrumentation__.Notify(245133)
	if err := p.performRenameTypeDesc(
		ctx,
		arrayDesc,
		newArrayName,
		arrayDesc.ParentSchemaID,
		tree.AsStringWithFQNames(n.n, p.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(245143)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245144)
	}
	__antithesis_instrumentation__.Notify(245134)
	return nil
}

func (p *planner) performRenameTypeDesc(
	ctx context.Context,
	desc *typedesc.Mutable,
	newName string,
	newSchemaID descpb.ID,
	jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(245145)
	oldNameKey := descpb.NameInfo{
		ParentID:       desc.GetParentID(),
		ParentSchemaID: desc.GetParentSchemaID(),
		Name:           desc.GetName(),
	}

	desc.SetName(newName)
	desc.SetParentSchemaID(newSchemaID)

	b := p.txn.NewBatch()
	p.renameNamespaceEntry(ctx, b, oldNameKey, desc)

	if err := p.writeTypeSchemaChange(ctx, desc, jobDesc); err != nil {
		__antithesis_instrumentation__.Notify(245147)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245148)
	}
	__antithesis_instrumentation__.Notify(245146)

	return p.txn.Run(ctx, b)
}

func (p *planner) renameTypeValue(
	ctx context.Context, n *alterTypeNode, oldVal string, newVal string,
) error {
	__antithesis_instrumentation__.Notify(245149)
	enumMemberIndex := -1

	for i := range n.desc.EnumMembers {
		__antithesis_instrumentation__.Notify(245154)
		member := n.desc.EnumMembers[i]
		if member.LogicalRepresentation == oldVal {
			__antithesis_instrumentation__.Notify(245155)
			enumMemberIndex = i
		} else {
			__antithesis_instrumentation__.Notify(245156)
			if member.LogicalRepresentation == newVal {
				__antithesis_instrumentation__.Notify(245157)
				return pgerror.Newf(pgcode.DuplicateObject,
					"enum value %s already exists", newVal)
			} else {
				__antithesis_instrumentation__.Notify(245158)
			}
		}
	}
	__antithesis_instrumentation__.Notify(245150)

	if enumMemberIndex == -1 {
		__antithesis_instrumentation__.Notify(245159)
		return pgerror.Newf(pgcode.InvalidParameterValue,
			"%s is not an existing enum value", oldVal)
	} else {
		__antithesis_instrumentation__.Notify(245160)
	}
	__antithesis_instrumentation__.Notify(245151)

	if enumMemberIsRemoving(&n.desc.EnumMembers[enumMemberIndex]) {
		__antithesis_instrumentation__.Notify(245161)
		return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
			"enum value %q is being dropped", oldVal)
	} else {
		__antithesis_instrumentation__.Notify(245162)
	}
	__antithesis_instrumentation__.Notify(245152)
	if enumMemberIsAdding(&n.desc.EnumMembers[enumMemberIndex]) {
		__antithesis_instrumentation__.Notify(245163)
		return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
			"enum value %q is being added, try again later", oldVal)

	} else {
		__antithesis_instrumentation__.Notify(245164)
	}
	__antithesis_instrumentation__.Notify(245153)

	n.desc.EnumMembers[enumMemberIndex].LogicalRepresentation = newVal

	return p.writeTypeSchemaChange(
		ctx,
		n.desc,
		tree.AsStringWithFQNames(n.n, p.Ann()),
	)
}

func (p *planner) setTypeSchema(ctx context.Context, n *alterTypeNode, schema string) error {
	__antithesis_instrumentation__.Notify(245165)
	typeDesc := n.desc
	schemaID := typeDesc.GetParentSchemaID()

	oldName, err := p.getQualifiedTypeName(ctx, typeDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(245173)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245174)
	}
	__antithesis_instrumentation__.Notify(245166)

	desiredSchemaID, err := p.prepareSetSchema(ctx, n.prefix.Database, typeDesc, schema)
	if err != nil {
		__antithesis_instrumentation__.Notify(245175)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245176)
	}
	__antithesis_instrumentation__.Notify(245167)

	if desiredSchemaID == schemaID {
		__antithesis_instrumentation__.Notify(245177)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245178)
	}
	__antithesis_instrumentation__.Notify(245168)

	err = p.performRenameTypeDesc(
		ctx, typeDesc, typeDesc.Name, desiredSchemaID, tree.AsStringWithFQNames(n.n, p.Ann()),
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(245179)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245180)
	}
	__antithesis_instrumentation__.Notify(245169)

	arrayDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, n.desc.ArrayTypeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(245181)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245182)
	}
	__antithesis_instrumentation__.Notify(245170)

	if err := p.performRenameTypeDesc(
		ctx, arrayDesc, arrayDesc.Name, desiredSchemaID, tree.AsStringWithFQNames(n.n, p.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(245183)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245184)
	}
	__antithesis_instrumentation__.Notify(245171)

	newName, err := p.getQualifiedTypeName(ctx, typeDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(245185)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245186)
	}
	__antithesis_instrumentation__.Notify(245172)

	return p.logEvent(ctx,
		desiredSchemaID,
		&eventpb.SetSchema{
			CommonEventDetails:    eventpb.CommonEventDetails{},
			CommonSQLEventDetails: eventpb.CommonSQLEventDetails{},
			DescriptorName:        oldName.FQString(),
			NewDescriptorName:     newName.FQString(),
			DescriptorType:        "type",
		},
	)
}

func (p *planner) alterTypeOwner(
	ctx context.Context, n *alterTypeNode, newOwner security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(245187)
	typeDesc := n.desc
	oldOwner := typeDesc.GetPrivileges().Owner()

	arrayDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, typeDesc.ArrayTypeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(245194)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245195)
	}
	__antithesis_instrumentation__.Notify(245188)

	if err := p.checkCanAlterToNewOwner(ctx, typeDesc, newOwner); err != nil {
		__antithesis_instrumentation__.Notify(245196)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245197)
	}
	__antithesis_instrumentation__.Notify(245189)

	if err := p.canCreateOnSchema(
		ctx, typeDesc.GetParentSchemaID(), typeDesc.ParentID, newOwner, checkPublicSchema); err != nil {
		__antithesis_instrumentation__.Notify(245198)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245199)
	}
	__antithesis_instrumentation__.Notify(245190)

	typeNameWithPrefix := tree.MakeTypeNameWithPrefix(n.prefix.NamePrefix(), typeDesc.GetName())

	arrayTypeNameWithPrefix := tree.MakeTypeNameWithPrefix(n.prefix.NamePrefix(), arrayDesc.GetName())

	if err := p.setNewTypeOwner(ctx, typeDesc, arrayDesc, typeNameWithPrefix,
		arrayTypeNameWithPrefix, newOwner); err != nil {
		__antithesis_instrumentation__.Notify(245200)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245201)
	}
	__antithesis_instrumentation__.Notify(245191)

	if newOwner == oldOwner {
		__antithesis_instrumentation__.Notify(245202)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245203)
	}
	__antithesis_instrumentation__.Notify(245192)

	if err := p.writeTypeSchemaChange(
		ctx, typeDesc, tree.AsStringWithFQNames(n.n, p.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(245204)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245205)
	}
	__antithesis_instrumentation__.Notify(245193)

	return p.writeTypeSchemaChange(
		ctx, arrayDesc, tree.AsStringWithFQNames(n.n, p.Ann()),
	)
}

func (p *planner) setNewTypeOwner(
	ctx context.Context,
	typeDesc *typedesc.Mutable,
	arrayTypeDesc *typedesc.Mutable,
	typeName tree.TypeName,
	arrayTypeName tree.TypeName,
	newOwner security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(245206)
	privs := typeDesc.GetPrivileges()
	privs.SetOwner(newOwner)

	arrayTypeDesc.Privileges.SetOwner(newOwner)

	if err := p.logEvent(ctx,
		typeDesc.GetID(),
		&eventpb.AlterTypeOwner{
			TypeName: typeName.FQString(),
			Owner:    newOwner.Normalized(),
		}); err != nil {
		__antithesis_instrumentation__.Notify(245208)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245209)
	}
	__antithesis_instrumentation__.Notify(245207)
	return p.logEvent(ctx,
		arrayTypeDesc.GetID(),
		&eventpb.AlterTypeOwner{
			TypeName: arrayTypeName.FQString(),
			Owner:    newOwner.Normalized(),
		})
}

func (n *alterTypeNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(245210)
	return false, nil
}
func (n *alterTypeNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(245211)
	return tree.Datums{}
}
func (n *alterTypeNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(245212) }
func (n *alterTypeNode) ReadingOwnWrites()         { __antithesis_instrumentation__.Notify(245213) }

func (p *planner) canModifyType(ctx context.Context, desc *typedesc.Mutable) error {
	__antithesis_instrumentation__.Notify(245214)
	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(245219)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245220)
	}
	__antithesis_instrumentation__.Notify(245215)
	if hasAdmin {
		__antithesis_instrumentation__.Notify(245221)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245222)
	}
	__antithesis_instrumentation__.Notify(245216)

	hasOwnership, err := p.HasOwnership(ctx, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(245223)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245224)
	}
	__antithesis_instrumentation__.Notify(245217)
	if !hasOwnership {
		__antithesis_instrumentation__.Notify(245225)
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"must be owner of type %s", tree.Name(desc.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(245226)
	}
	__antithesis_instrumentation__.Notify(245218)
	return nil
}
