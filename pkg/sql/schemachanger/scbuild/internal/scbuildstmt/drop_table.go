package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func DropTable(b BuildCtx, n *tree.DropTable) {
	__antithesis_instrumentation__.Notify(580077)
	var toCheckBackrefs []catid.DescID
	droppedOwnedSequences := make(map[catid.DescID]catalog.DescriptorIDSet)
	for idx := range n.Names {
		__antithesis_instrumentation__.Notify(580079)
		name := &n.Names[idx]
		elts := b.ResolveTable(name.ToUnresolvedObjectName(), ResolveParams{
			IsExistenceOptional: n.IfExists,
			RequiredPrivilege:   privilege.DROP,
		})
		_, _, tbl := scpb.FindTable(elts)
		if tbl == nil {
			__antithesis_instrumentation__.Notify(580083)
			b.MarkNameAsNonExistent(name)
			continue
		} else {
			__antithesis_instrumentation__.Notify(580084)
		}
		__antithesis_instrumentation__.Notify(580080)

		name.ObjectNamePrefix = b.NamePrefix(tbl)

		if tbl.IsTemporary {
			__antithesis_instrumentation__.Notify(580085)
			panic(scerrors.NotImplementedErrorf(n, "dropping a temporary table"))
		} else {
			__antithesis_instrumentation__.Notify(580086)
		}
		__antithesis_instrumentation__.Notify(580081)

		if n.DropBehavior == tree.DropCascade {
			__antithesis_instrumentation__.Notify(580087)
			dropCascadeDescriptor(b, tbl.TableID)
		} else {
			__antithesis_instrumentation__.Notify(580088)

			var ownedIDs catalog.DescriptorIDSet
			scpb.ForEachSequenceOwner(
				b.QueryByID(tbl.TableID),
				func(_ scpb.Status, target scpb.TargetStatus, so *scpb.SequenceOwner) {
					__antithesis_instrumentation__.Notify(580090)
					if target == scpb.ToPublic {
						__antithesis_instrumentation__.Notify(580091)
						ownedIDs.Add(so.SequenceID)
					} else {
						__antithesis_instrumentation__.Notify(580092)
					}
				},
			)
			__antithesis_instrumentation__.Notify(580089)
			if dropRestrictDescriptor(b, tbl.TableID) {
				__antithesis_instrumentation__.Notify(580093)
				toCheckBackrefs = append(toCheckBackrefs, tbl.TableID)
				ownedIDs.ForEach(func(ownedSequenceID descpb.ID) {
					__antithesis_instrumentation__.Notify(580095)
					dropRestrictDescriptor(b, ownedSequenceID)
				})
				__antithesis_instrumentation__.Notify(580094)
				droppedOwnedSequences[tbl.TableID] = ownedIDs
			} else {
				__antithesis_instrumentation__.Notify(580096)
			}
		}
		__antithesis_instrumentation__.Notify(580082)
		b.IncrementSubWorkID()
		b.IncrementSchemaChangeDropCounter("table")
	}
	__antithesis_instrumentation__.Notify(580078)

	for _, tableID := range toCheckBackrefs {
		__antithesis_instrumentation__.Notify(580097)
		backrefs := undroppedBackrefs(b, tableID)
		hasUndroppedBackrefs := !backrefs.IsEmpty()
		droppedOwnedSequences[tableID].ForEach(func(seqID descpb.ID) {
			__antithesis_instrumentation__.Notify(580101)
			if !undroppedBackrefs(b, seqID).IsEmpty() {
				__antithesis_instrumentation__.Notify(580102)
				hasUndroppedBackrefs = true
			} else {
				__antithesis_instrumentation__.Notify(580103)
			}
		})
		__antithesis_instrumentation__.Notify(580098)
		if !hasUndroppedBackrefs {
			__antithesis_instrumentation__.Notify(580104)
			continue
		} else {
			__antithesis_instrumentation__.Notify(580105)
		}
		__antithesis_instrumentation__.Notify(580099)
		_, _, ns := scpb.FindNamespace(b.QueryByID(tableID))
		maybePanicOnDependentView(b, ns, backrefs)
		if _, _, fk := scpb.FindForeignKeyConstraint(backrefs); fk != nil {
			__antithesis_instrumentation__.Notify(580106)
			panic(pgerror.Newf(pgcode.DependentObjectsStillExist,
				"%q is referenced by foreign key from table %q", ns.Name, simpleName(b, fk.TableID)))
		} else {
			__antithesis_instrumentation__.Notify(580107)
		}
		__antithesis_instrumentation__.Notify(580100)
		panic(pgerror.Newf(pgcode.DependentObjectsStillExist,
			"cannot drop table %s because other objects depend on it", ns.Name))
	}
}
