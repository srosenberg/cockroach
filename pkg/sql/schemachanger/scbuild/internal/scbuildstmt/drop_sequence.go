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
)

func DropSequence(b BuildCtx, n *tree.DropSequence) {
	__antithesis_instrumentation__.Notify(580057)
	var toCheckBackrefs []catid.DescID
	for idx := range n.Names {
		__antithesis_instrumentation__.Notify(580059)
		name := &n.Names[idx]
		elts := b.ResolveSequence(name.ToUnresolvedObjectName(), ResolveParams{
			IsExistenceOptional: n.IfExists,
			RequiredPrivilege:   privilege.DROP,
		})
		_, _, seq := scpb.FindSequence(elts)
		if seq == nil {
			__antithesis_instrumentation__.Notify(580063)
			b.MarkNameAsNonExistent(name)
			continue
		} else {
			__antithesis_instrumentation__.Notify(580064)
		}
		__antithesis_instrumentation__.Notify(580060)

		name.ObjectNamePrefix = b.NamePrefix(seq)

		if seq.IsTemporary {
			__antithesis_instrumentation__.Notify(580065)
			panic(scerrors.NotImplementedErrorf(n, "dropping a temporary sequence"))
		} else {
			__antithesis_instrumentation__.Notify(580066)
		}
		__antithesis_instrumentation__.Notify(580061)
		if n.DropBehavior == tree.DropCascade {
			__antithesis_instrumentation__.Notify(580067)
			dropCascadeDescriptor(b, seq.SequenceID)
		} else {
			__antithesis_instrumentation__.Notify(580068)
			if dropRestrictDescriptor(b, seq.SequenceID) {
				__antithesis_instrumentation__.Notify(580069)

				scpb.ForEachSequenceOwner(
					undroppedBackrefs(b, seq.SequenceID),
					func(_ scpb.Status, _ scpb.TargetStatus, so *scpb.SequenceOwner) {
						__antithesis_instrumentation__.Notify(580071)
						dropElement(b, so)
					},
				)
				__antithesis_instrumentation__.Notify(580070)
				toCheckBackrefs = append(toCheckBackrefs, seq.SequenceID)
			} else {
				__antithesis_instrumentation__.Notify(580072)
			}
		}
		__antithesis_instrumentation__.Notify(580062)
		b.IncrementSubWorkID()
		b.IncrementSchemaChangeDropCounter("sequence")
	}
	__antithesis_instrumentation__.Notify(580058)

	for _, sequenceID := range toCheckBackrefs {
		__antithesis_instrumentation__.Notify(580073)
		backrefs := undroppedBackrefs(b, sequenceID)
		if backrefs.IsEmpty() {
			__antithesis_instrumentation__.Notify(580075)
			continue
		} else {
			__antithesis_instrumentation__.Notify(580076)
		}
		__antithesis_instrumentation__.Notify(580074)
		_, _, ns := scpb.FindNamespace(b.QueryByID(sequenceID))
		panic(pgerror.Newf(pgcode.DependentObjectsStillExist,
			"cannot drop sequence %s because other objects depend on it", ns.Name))
	}
}
