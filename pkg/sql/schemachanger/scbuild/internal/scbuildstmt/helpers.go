package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
)

func qualifiedName(b BuildCtx, id catid.DescID) string {
	__antithesis_instrumentation__.Notify(580173)
	_, _, ns := scpb.FindNamespace(b.QueryByID(id))
	_, _, sc := scpb.FindNamespace(b.QueryByID(ns.SchemaID))
	_, _, db := scpb.FindNamespace(b.QueryByID(ns.DatabaseID))
	if db == nil {
		__antithesis_instrumentation__.Notify(580176)
		return ns.Name
	} else {
		__antithesis_instrumentation__.Notify(580177)
	}
	__antithesis_instrumentation__.Notify(580174)
	if sc == nil {
		__antithesis_instrumentation__.Notify(580178)
		return db.Name + "." + ns.Name
	} else {
		__antithesis_instrumentation__.Notify(580179)
	}
	__antithesis_instrumentation__.Notify(580175)
	return db.Name + "." + sc.Name + "." + ns.Name
}

func simpleName(b BuildCtx, id catid.DescID) string {
	__antithesis_instrumentation__.Notify(580180)
	_, _, ns := scpb.FindNamespace(b.QueryByID(id))
	return ns.Name
}

func dropRestrictDescriptor(b BuildCtx, id catid.DescID) (hasChanged bool) {
	__antithesis_instrumentation__.Notify(580181)
	b.QueryByID(id).ForEachElementStatus(func(_ scpb.Status, target scpb.TargetStatus, e scpb.Element) {
		__antithesis_instrumentation__.Notify(580183)
		if target == scpb.ToAbsent {
			__antithesis_instrumentation__.Notify(580185)
			return
		} else {
			__antithesis_instrumentation__.Notify(580186)
		}
		__antithesis_instrumentation__.Notify(580184)
		b.CheckPrivilege(e, privilege.DROP)
		dropElement(b, e)
		hasChanged = true
	})
	__antithesis_instrumentation__.Notify(580182)
	return hasChanged
}

func dropElement(b BuildCtx, e scpb.Element) {
	__antithesis_instrumentation__.Notify(580187)

	switch t := e.(type) {
	case *scpb.ColumnType:
		__antithesis_instrumentation__.Notify(580189)
		t.IsRelationBeingDropped = true
	case *scpb.SecondaryIndexPartial:
		__antithesis_instrumentation__.Notify(580190)
		t.IsRelationBeingDropped = true
	}
	__antithesis_instrumentation__.Notify(580188)
	b.Drop(e)
}

func dropCascadeDescriptor(b BuildCtx, id catid.DescID) {
	__antithesis_instrumentation__.Notify(580191)
	undropped := b.QueryByID(id).Filter(func(_ scpb.Status, target scpb.TargetStatus, _ scpb.Element) bool {
		__antithesis_instrumentation__.Notify(580196)
		return target == scpb.ToPublic
	})
	__antithesis_instrumentation__.Notify(580192)

	if undropped.IsEmpty() {
		__antithesis_instrumentation__.Notify(580197)
		return
	} else {
		__antithesis_instrumentation__.Notify(580198)
	}
	__antithesis_instrumentation__.Notify(580193)

	var isVirtualSchema bool
	undropped.ForEachElementStatus(func(_ scpb.Status, _ scpb.TargetStatus, e scpb.Element) {
		__antithesis_instrumentation__.Notify(580199)
		switch t := e.(type) {
		case *scpb.Database:
			__antithesis_instrumentation__.Notify(580201)
			break
		case *scpb.Schema:
			__antithesis_instrumentation__.Notify(580202)
			if t.IsTemporary {
				__antithesis_instrumentation__.Notify(580210)
				panic(scerrors.NotImplementedErrorf(nil, "dropping a temporary schema"))
			} else {
				__antithesis_instrumentation__.Notify(580211)
			}
			__antithesis_instrumentation__.Notify(580203)
			isVirtualSchema = t.IsVirtual

			return
		case *scpb.Table:
			__antithesis_instrumentation__.Notify(580204)
			if t.IsTemporary {
				__antithesis_instrumentation__.Notify(580212)
				panic(scerrors.NotImplementedErrorf(nil, "dropping a temporary table"))
			} else {
				__antithesis_instrumentation__.Notify(580213)
			}
		case *scpb.Sequence:
			__antithesis_instrumentation__.Notify(580205)
			if t.IsTemporary {
				__antithesis_instrumentation__.Notify(580214)
				panic(scerrors.NotImplementedErrorf(nil, "dropping a temporary sequence"))
			} else {
				__antithesis_instrumentation__.Notify(580215)
			}
		case *scpb.View:
			__antithesis_instrumentation__.Notify(580206)
			if t.IsTemporary {
				__antithesis_instrumentation__.Notify(580216)
				panic(scerrors.NotImplementedErrorf(nil, "dropping a temporary view"))
			} else {
				__antithesis_instrumentation__.Notify(580217)
			}
		case *scpb.EnumType:
			__antithesis_instrumentation__.Notify(580207)
		case *scpb.AliasType:
			__antithesis_instrumentation__.Notify(580208)
			break
		default:
			__antithesis_instrumentation__.Notify(580209)
			return
		}
		__antithesis_instrumentation__.Notify(580200)
		b.CheckPrivilege(e, privilege.DROP)
	})
	__antithesis_instrumentation__.Notify(580194)

	next := b.WithNewSourceElementID()
	undropped.ForEachElementStatus(func(_ scpb.Status, target scpb.TargetStatus, e scpb.Element) {
		__antithesis_instrumentation__.Notify(580218)
		if isVirtualSchema {
			__antithesis_instrumentation__.Notify(580220)

			return
		} else {
			__antithesis_instrumentation__.Notify(580221)
		}
		__antithesis_instrumentation__.Notify(580219)
		dropElement(b, e)
		switch t := e.(type) {
		case *scpb.EnumType:
			__antithesis_instrumentation__.Notify(580222)
			dropCascadeDescriptor(next, t.ArrayTypeID)
		case *scpb.SequenceOwner:
			__antithesis_instrumentation__.Notify(580223)
			dropCascadeDescriptor(next, t.SequenceID)
		}
	})
	__antithesis_instrumentation__.Notify(580195)

	ub := undroppedBackrefs(b, id)
	ub.ForEachElementStatus(func(_ scpb.Status, target scpb.TargetStatus, e scpb.Element) {
		__antithesis_instrumentation__.Notify(580224)
		switch t := e.(type) {
		case *scpb.SchemaParent:
			__antithesis_instrumentation__.Notify(580225)
			dropCascadeDescriptor(next, t.SchemaID)
		case *scpb.ObjectParent:
			__antithesis_instrumentation__.Notify(580226)
			dropCascadeDescriptor(next, t.ObjectID)
		case *scpb.View:
			__antithesis_instrumentation__.Notify(580227)
			dropCascadeDescriptor(next, t.ViewID)
		case *scpb.Sequence:
			__antithesis_instrumentation__.Notify(580228)
			dropCascadeDescriptor(next, t.SequenceID)
		case *scpb.AliasType:
			__antithesis_instrumentation__.Notify(580229)
			dropCascadeDescriptor(next, t.TypeID)
		case *scpb.EnumType:
			__antithesis_instrumentation__.Notify(580230)
			dropCascadeDescriptor(next, t.TypeID)
		case *scpb.Column, *scpb.ColumnType, *scpb.SecondaryIndexPartial:
			__antithesis_instrumentation__.Notify(580231)

			break
		case
			*scpb.ColumnDefaultExpression,
			*scpb.ColumnOnUpdateExpression,
			*scpb.CheckConstraint,
			*scpb.ForeignKeyConstraint,
			*scpb.SequenceOwner,
			*scpb.DatabaseRegionConfig:
			__antithesis_instrumentation__.Notify(580232)
			dropElement(b, e)
		}
	})
}

func undroppedBackrefs(b BuildCtx, id catid.DescID) ElementResultSet {
	__antithesis_instrumentation__.Notify(580233)
	return b.BackReferences(id).Filter(func(_ scpb.Status, target scpb.TargetStatus, e scpb.Element) bool {
		__antithesis_instrumentation__.Notify(580234)
		return target != scpb.ToAbsent && func() bool {
			__antithesis_instrumentation__.Notify(580235)
			return screl.ContainsDescID(e, id) == true
		}() == true
	})
}

func descIDs(input ElementResultSet) (ids catalog.DescriptorIDSet) {
	__antithesis_instrumentation__.Notify(580236)
	input.ForEachElementStatus(func(_ scpb.Status, _ scpb.TargetStatus, e scpb.Element) {
		__antithesis_instrumentation__.Notify(580238)
		ids.Add(screl.GetDescID(e))
	})
	__antithesis_instrumentation__.Notify(580237)
	return ids
}
