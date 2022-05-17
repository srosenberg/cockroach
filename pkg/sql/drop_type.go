package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type dropTypeNode struct {
	n      *tree.DropType
	toDrop map[descpb.ID]*typedesc.Mutable
}

var _ planNode = &dropTypeNode{n: nil}

func (p *planner) DropType(ctx context.Context, n *tree.DropType) (planNode, error) {
	__antithesis_instrumentation__.Notify(469756)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"DROP TYPE",
	); err != nil {
		__antithesis_instrumentation__.Notify(469760)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469761)
	}
	__antithesis_instrumentation__.Notify(469757)

	node := &dropTypeNode{
		n:      n,
		toDrop: make(map[descpb.ID]*typedesc.Mutable),
	}
	if n.DropBehavior == tree.DropCascade {
		__antithesis_instrumentation__.Notify(469762)
		return nil, unimplemented.NewWithIssue(51480, "DROP TYPE CASCADE is not yet supported")
	} else {
		__antithesis_instrumentation__.Notify(469763)
	}
	__antithesis_instrumentation__.Notify(469758)
	for _, name := range n.Names {
		__antithesis_instrumentation__.Notify(469764)

		_, typeDesc, err := p.ResolveMutableTypeDescriptor(ctx, name, !n.IfExists)
		if err != nil {
			__antithesis_instrumentation__.Notify(469772)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469773)
		}
		__antithesis_instrumentation__.Notify(469765)
		if typeDesc == nil {
			__antithesis_instrumentation__.Notify(469774)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469775)
		}
		__antithesis_instrumentation__.Notify(469766)

		if _, ok := node.toDrop[typeDesc.ID]; ok {
			__antithesis_instrumentation__.Notify(469776)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469777)
		}
		__antithesis_instrumentation__.Notify(469767)
		switch typeDesc.Kind {
		case descpb.TypeDescriptor_ALIAS:
			__antithesis_instrumentation__.Notify(469778)

			return nil, pgerror.Newf(
				pgcode.DependentObjectsStillExist,
				"%q is an implicit array type and cannot be modified",
				name,
			)
		case descpb.TypeDescriptor_MULTIREGION_ENUM:
			__antithesis_instrumentation__.Notify(469779)

			return nil, errors.WithHintf(
				pgerror.Newf(
					pgcode.DependentObjectsStillExist,
					"%q is a multi-region enum and cannot be modified directly",
					name,
				),
				"try ALTER DATABASE DROP REGION %s", name)
		case descpb.TypeDescriptor_ENUM:
			__antithesis_instrumentation__.Notify(469780)
			sqltelemetry.IncrementEnumCounter(sqltelemetry.EnumDrop)
		case descpb.TypeDescriptor_TABLE_IMPLICIT_RECORD_TYPE:
			__antithesis_instrumentation__.Notify(469781)
			return nil, pgerror.Newf(
				pgcode.DependentObjectsStillExist,
				"cannot drop type %q because table %q requires it",
				name, name,
			)
		default:
			__antithesis_instrumentation__.Notify(469782)
		}
		__antithesis_instrumentation__.Notify(469768)

		if err := p.canDropTypeDesc(ctx, typeDesc, n.DropBehavior); err != nil {
			__antithesis_instrumentation__.Notify(469783)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469784)
		}
		__antithesis_instrumentation__.Notify(469769)

		mutArrayDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, typeDesc.ArrayTypeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(469785)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469786)
		}
		__antithesis_instrumentation__.Notify(469770)

		if err := p.canDropTypeDesc(ctx, mutArrayDesc, n.DropBehavior); err != nil {
			__antithesis_instrumentation__.Notify(469787)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469788)
		}
		__antithesis_instrumentation__.Notify(469771)

		node.toDrop[typeDesc.ID] = typeDesc
		node.toDrop[mutArrayDesc.ID] = mutArrayDesc
	}
	__antithesis_instrumentation__.Notify(469759)
	return node, nil
}

func (p *planner) canDropTypeDesc(
	ctx context.Context, desc *typedesc.Mutable, behavior tree.DropBehavior,
) error {
	__antithesis_instrumentation__.Notify(469789)
	if err := p.canModifyType(ctx, desc); err != nil {
		__antithesis_instrumentation__.Notify(469792)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469793)
	}
	__antithesis_instrumentation__.Notify(469790)
	if len(desc.ReferencingDescriptorIDs) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(469794)
		return behavior != tree.DropCascade == true
	}() == true {
		__antithesis_instrumentation__.Notify(469795)
		dependentNames, err := p.getFullyQualifiedTableNamesFromIDs(ctx, desc.ReferencingDescriptorIDs)
		if err != nil {
			__antithesis_instrumentation__.Notify(469797)
			return errors.Wrapf(err, "type %q has dependent objects", desc.Name)
		} else {
			__antithesis_instrumentation__.Notify(469798)
		}
		__antithesis_instrumentation__.Notify(469796)
		return pgerror.Newf(
			pgcode.DependentObjectsStillExist,
			"cannot drop type %q because other objects (%v) still depend on it",
			desc.Name,
			dependentNames,
		)
	} else {
		__antithesis_instrumentation__.Notify(469799)
	}
	__antithesis_instrumentation__.Notify(469791)
	return nil
}

func (n *dropTypeNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(469800)
	for _, typeDesc := range n.toDrop {
		__antithesis_instrumentation__.Notify(469802)
		typeFQName, err := getTypeNameFromTypeDescriptor(
			oneAtATimeSchemaResolver{params.ctx, params.p},
			typeDesc,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469805)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469806)
		}
		__antithesis_instrumentation__.Notify(469803)
		err = params.p.dropTypeImpl(params.ctx, typeDesc, "dropping type "+typeFQName.FQString(), true)
		if err != nil {
			__antithesis_instrumentation__.Notify(469807)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469808)
		}
		__antithesis_instrumentation__.Notify(469804)
		event := &eventpb.DropType{
			TypeName: typeFQName.FQString(),
		}

		if err := params.p.logEvent(params.ctx, typeDesc.ID, event); err != nil {
			__antithesis_instrumentation__.Notify(469809)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469810)
		}
	}
	__antithesis_instrumentation__.Notify(469801)
	return nil
}

func (p *planner) addTypeBackReference(
	ctx context.Context, typeID, ref descpb.ID, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(469811)
	mutDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, typeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(469814)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469815)
	}
	__antithesis_instrumentation__.Notify(469812)

	if err := p.CheckPrivilege(ctx, mutDesc, privilege.USAGE); err != nil {
		__antithesis_instrumentation__.Notify(469816)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469817)
	}
	__antithesis_instrumentation__.Notify(469813)

	mutDesc.AddReferencingDescriptorID(ref)
	return p.writeTypeSchemaChange(ctx, mutDesc, jobDesc)
}

func (p *planner) removeTypeBackReferences(
	ctx context.Context, typeIDs []descpb.ID, ref descpb.ID, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(469818)
	for _, typeID := range typeIDs {
		__antithesis_instrumentation__.Notify(469820)
		mutDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, typeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(469822)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469823)
		}
		__antithesis_instrumentation__.Notify(469821)
		mutDesc.RemoveReferencingDescriptorID(ref)
		if err := p.writeTypeSchemaChange(ctx, mutDesc, jobDesc); err != nil {
			__antithesis_instrumentation__.Notify(469824)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469825)
		}
	}
	__antithesis_instrumentation__.Notify(469819)
	return nil
}

func (p *planner) addBackRefsFromAllTypesInTable(
	ctx context.Context, desc *tabledesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(469826)
	_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(
		ctx, p.txn, desc.GetParentID(), tree.DatabaseLookupFlags{
			Required:    true,
			AvoidLeased: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(469831)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469832)
	}
	__antithesis_instrumentation__.Notify(469827)
	typeIDs, _, err := desc.GetAllReferencedTypeIDs(dbDesc, func(id descpb.ID) (catalog.TypeDescriptor, error) {
		__antithesis_instrumentation__.Notify(469833)
		mutDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(469835)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469836)
		}
		__antithesis_instrumentation__.Notify(469834)
		return mutDesc, nil
	})
	__antithesis_instrumentation__.Notify(469828)
	if err != nil {
		__antithesis_instrumentation__.Notify(469837)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469838)
	}
	__antithesis_instrumentation__.Notify(469829)
	for _, id := range typeIDs {
		__antithesis_instrumentation__.Notify(469839)
		jobDesc := fmt.Sprintf("updating type back reference %d for table %d", id, desc.ID)
		if err := p.addTypeBackReference(ctx, id, desc.ID, jobDesc); err != nil {
			__antithesis_instrumentation__.Notify(469840)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469841)
		}
	}
	__antithesis_instrumentation__.Notify(469830)
	return nil
}

func (p *planner) removeBackRefsFromAllTypesInTable(
	ctx context.Context, desc *tabledesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(469842)
	_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(
		ctx, p.txn, desc.GetParentID(), tree.DatabaseLookupFlags{
			Required:    true,
			AvoidLeased: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(469846)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469847)
	}
	__antithesis_instrumentation__.Notify(469843)
	typeIDs, _, err := desc.GetAllReferencedTypeIDs(dbDesc, func(id descpb.ID) (catalog.TypeDescriptor, error) {
		__antithesis_instrumentation__.Notify(469848)
		mutDesc, err := p.Descriptors().GetMutableTypeVersionByID(ctx, p.txn, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(469850)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469851)
		}
		__antithesis_instrumentation__.Notify(469849)
		return mutDesc, nil
	})
	__antithesis_instrumentation__.Notify(469844)
	if err != nil {
		__antithesis_instrumentation__.Notify(469852)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469853)
	}
	__antithesis_instrumentation__.Notify(469845)
	jobDesc := fmt.Sprintf("updating type back references %v for table %d", typeIDs, desc.ID)
	return p.removeTypeBackReferences(ctx, typeIDs, desc.ID, jobDesc)
}

func (p *planner) dropTypeImpl(
	ctx context.Context, typeDesc *typedesc.Mutable, jobDesc string, queueJob bool,
) error {
	__antithesis_instrumentation__.Notify(469854)
	if typeDesc.Dropped() {
		__antithesis_instrumentation__.Notify(469858)
		return errors.Errorf("type %q is already being dropped", typeDesc.Name)
	} else {
		__antithesis_instrumentation__.Notify(469859)
	}
	__antithesis_instrumentation__.Notify(469855)

	typeDesc.SetDropped()

	b := p.txn.NewBatch()
	p.dropNamespaceEntry(ctx, b, typeDesc)
	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(469860)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469861)
	}
	__antithesis_instrumentation__.Notify(469856)

	if queueJob {
		__antithesis_instrumentation__.Notify(469862)
		return p.writeTypeSchemaChange(ctx, typeDesc, jobDesc)
	} else {
		__antithesis_instrumentation__.Notify(469863)
	}
	__antithesis_instrumentation__.Notify(469857)
	return p.writeTypeDesc(ctx, typeDesc)
}

func (n *dropTypeNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(469864)
	return false, nil
}
func (n *dropTypeNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(469865)
	return tree.Datums{}
}
func (n *dropTypeNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(469866) }
func (n *dropTypeNode) ReadingOwnWrites()         { __antithesis_instrumentation__.Notify(469867) }
