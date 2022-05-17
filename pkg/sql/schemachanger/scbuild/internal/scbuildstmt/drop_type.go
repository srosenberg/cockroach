package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

func DropType(b BuildCtx, n *tree.DropType) {
	__antithesis_instrumentation__.Notify(580108)
	if n.DropBehavior == tree.DropCascade {
		__antithesis_instrumentation__.Notify(580111)
		panic(scerrors.NotImplementedErrorf(n, "DROP TYPE CASCADE is not yet supported"))
	} else {
		__antithesis_instrumentation__.Notify(580112)
	}
	__antithesis_instrumentation__.Notify(580109)
	var toCheckBackrefs []catid.DescID
	arrayTypesToAlsoCheck := make(map[catid.DescID]catid.DescID)
	for _, name := range n.Names {
		__antithesis_instrumentation__.Notify(580113)
		elts := b.ResolveEnumType(name, ResolveParams{
			IsExistenceOptional: n.IfExists,
			RequiredPrivilege:   privilege.DROP,
		})
		_, _, typ := scpb.FindEnumType(elts)
		if typ == nil {
			__antithesis_instrumentation__.Notify(580116)
			continue
		} else {
			__antithesis_instrumentation__.Notify(580117)
		}
		__antithesis_instrumentation__.Notify(580114)
		prefix := b.NamePrefix(typ)

		tn := tree.MakeTypeNameWithPrefix(prefix, name.Object())
		b.SetUnresolvedNameAnnotation(name, &tn)

		if n.DropBehavior == tree.DropCascade {
			__antithesis_instrumentation__.Notify(580118)
			dropCascadeDescriptor(b, typ.TypeID)
		} else {
			__antithesis_instrumentation__.Notify(580119)
			if dropRestrictDescriptor(b, typ.TypeID) {
				__antithesis_instrumentation__.Notify(580121)
				toCheckBackrefs = append(toCheckBackrefs, typ.TypeID)
			} else {
				__antithesis_instrumentation__.Notify(580122)
			}
			__antithesis_instrumentation__.Notify(580120)
			b.IncrementSubWorkID()
			if dropRestrictDescriptor(b.WithNewSourceElementID(), typ.ArrayTypeID) {
				__antithesis_instrumentation__.Notify(580123)
				arrayTypesToAlsoCheck[typ.TypeID] = typ.ArrayTypeID
			} else {
				__antithesis_instrumentation__.Notify(580124)
			}
		}
		__antithesis_instrumentation__.Notify(580115)
		b.IncrementSubWorkID()
		b.IncrementEnumCounter(sqltelemetry.EnumDrop)
	}
	__antithesis_instrumentation__.Notify(580110)

	for _, typeID := range toCheckBackrefs {
		__antithesis_instrumentation__.Notify(580125)
		dependentNames := dependentTypeNames(b, typeID)
		if arrayTypeID, found := arrayTypesToAlsoCheck[typeID]; len(dependentNames) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(580127)
			return found == true
		}() == true {
			__antithesis_instrumentation__.Notify(580128)
			dependentNames = dependentTypeNames(b, arrayTypeID)
		} else {
			__antithesis_instrumentation__.Notify(580129)
		}
		__antithesis_instrumentation__.Notify(580126)
		if len(dependentNames) > 0 {
			__antithesis_instrumentation__.Notify(580130)
			panic(pgerror.Newf(
				pgcode.DependentObjectsStillExist,
				"cannot drop type %q because other objects (%v) still depend on it",
				simpleName(b, typeID), dependentNames,
			))
		} else {
			__antithesis_instrumentation__.Notify(580131)
		}
	}
}

func dependentTypeNames(b BuildCtx, typeID catid.DescID) (dependentNames []string) {
	__antithesis_instrumentation__.Notify(580132)
	backrefs := undroppedBackrefs(b, typeID)
	if backrefs.IsEmpty() {
		__antithesis_instrumentation__.Notify(580135)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(580136)
	}
	__antithesis_instrumentation__.Notify(580133)
	descIDs(backrefs).ForEach(func(depID descpb.ID) {
		__antithesis_instrumentation__.Notify(580137)
		dependentNames = append(dependentNames, qualifiedName(b, depID))
	})
	__antithesis_instrumentation__.Notify(580134)
	return dependentNames
}
