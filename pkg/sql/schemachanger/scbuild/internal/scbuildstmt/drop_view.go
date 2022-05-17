package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/errors"
)

func DropView(b BuildCtx, n *tree.DropView) {
	__antithesis_instrumentation__.Notify(580138)
	var toCheckBackrefs []catid.DescID
	for i := range n.Names {
		__antithesis_instrumentation__.Notify(580140)
		name := &n.Names[i]
		elts := b.ResolveView(name.ToUnresolvedObjectName(), ResolveParams{
			IsExistenceOptional: n.IfExists,
			RequiredPrivilege:   privilege.DROP,
		})
		_, _, view := scpb.FindView(elts)
		if view == nil {
			__antithesis_instrumentation__.Notify(580146)
			b.MarkNameAsNonExistent(name)
			continue
		} else {
			__antithesis_instrumentation__.Notify(580147)
		}
		__antithesis_instrumentation__.Notify(580141)

		name.ObjectNamePrefix = b.NamePrefix(view)

		if view.IsTemporary {
			__antithesis_instrumentation__.Notify(580148)
			panic(scerrors.NotImplementedErrorf(n, "dropping a temporary view"))
		} else {
			__antithesis_instrumentation__.Notify(580149)
		}
		__antithesis_instrumentation__.Notify(580142)
		if view.IsMaterialized && func() bool {
			__antithesis_instrumentation__.Notify(580150)
			return !n.IsMaterialized == true
		}() == true {
			__antithesis_instrumentation__.Notify(580151)
			panic(errors.WithHint(pgerror.Newf(pgcode.WrongObjectType, "%q is a materialized view", name.ObjectName),
				"use the corresponding MATERIALIZED VIEW command"))
		} else {
			__antithesis_instrumentation__.Notify(580152)
		}
		__antithesis_instrumentation__.Notify(580143)
		if !view.IsMaterialized && func() bool {
			__antithesis_instrumentation__.Notify(580153)
			return n.IsMaterialized == true
		}() == true {
			__antithesis_instrumentation__.Notify(580154)
			panic(pgerror.Newf(pgcode.WrongObjectType, "%q is not a materialized view", name.ObjectName))
		} else {
			__antithesis_instrumentation__.Notify(580155)
		}
		__antithesis_instrumentation__.Notify(580144)
		if n.DropBehavior == tree.DropCascade {
			__antithesis_instrumentation__.Notify(580156)
			dropCascadeDescriptor(b, view.ViewID)
		} else {
			__antithesis_instrumentation__.Notify(580157)
			if dropRestrictDescriptor(b, view.ViewID) {
				__antithesis_instrumentation__.Notify(580158)
				toCheckBackrefs = append(toCheckBackrefs, view.ViewID)
			} else {
				__antithesis_instrumentation__.Notify(580159)
			}
		}
		__antithesis_instrumentation__.Notify(580145)
		b.IncrementSubWorkID()
		if view.IsMaterialized {
			__antithesis_instrumentation__.Notify(580160)
			b.IncrementSchemaChangeDropCounter("materialized_view")
		} else {
			__antithesis_instrumentation__.Notify(580161)
			b.IncrementSchemaChangeDropCounter("view")
		}
	}
	__antithesis_instrumentation__.Notify(580139)

	for _, viewID := range toCheckBackrefs {
		__antithesis_instrumentation__.Notify(580162)
		backrefs := undroppedBackrefs(b, viewID)
		if backrefs.IsEmpty() {
			__antithesis_instrumentation__.Notify(580164)
			continue
		} else {
			__antithesis_instrumentation__.Notify(580165)
		}
		__antithesis_instrumentation__.Notify(580163)
		_, _, ns := scpb.FindNamespace(b.QueryByID(viewID))
		maybePanicOnDependentView(b, ns, backrefs)
		panic(pgerror.Newf(pgcode.DependentObjectsStillExist,
			"cannot drop view %s because other objects depend on it", ns.Name))
	}
}

func maybePanicOnDependentView(b BuildCtx, ns *scpb.Namespace, backrefs ElementResultSet) {
	__antithesis_instrumentation__.Notify(580166)
	_, _, depView := scpb.FindView(backrefs)
	if depView == nil {
		__antithesis_instrumentation__.Notify(580169)
		return
	} else {
		__antithesis_instrumentation__.Notify(580170)
	}
	__antithesis_instrumentation__.Notify(580167)
	_, _, nsDep := scpb.FindNamespace(b.QueryByID(depView.ViewID))
	if nsDep.DatabaseID != ns.DatabaseID {
		__antithesis_instrumentation__.Notify(580171)
		panic(errors.WithHintf(sqlerrors.NewDependentObjectErrorf("cannot drop relation %q because view %q depends on it",
			ns.Name, qualifiedName(b, depView.ViewID)),
			"you can drop %s instead.", nsDep.Name))
	} else {
		__antithesis_instrumentation__.Notify(580172)
	}
	__antithesis_instrumentation__.Notify(580168)
	panic(sqlerrors.NewDependentObjectErrorf("cannot drop relation %q because view %q depends on it",
		ns.Name, nsDep.Name))
}
