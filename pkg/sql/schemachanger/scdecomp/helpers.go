package scdecomp

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/seqexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/lib/pq/oid"
)

func descriptorStatus(desc catalog.Descriptor) scpb.Status {
	__antithesis_instrumentation__.Notify(580393)
	if desc.Dropped() {
		__antithesis_instrumentation__.Notify(580395)
		return scpb.Status_DROPPED
	} else {
		__antithesis_instrumentation__.Notify(580396)
	}
	__antithesis_instrumentation__.Notify(580394)
	return scpb.Status_PUBLIC
}

func maybeMutationStatus(mm catalog.TableElementMaybeMutation) scpb.Status {
	__antithesis_instrumentation__.Notify(580397)
	switch {
	case mm.DeleteOnly():
		__antithesis_instrumentation__.Notify(580398)
		return scpb.Status_DELETE_ONLY
	case mm.WriteAndDeleteOnly():
		__antithesis_instrumentation__.Notify(580399)
		return scpb.Status_WRITE_ONLY
	case mm.Backfilling():
		__antithesis_instrumentation__.Notify(580400)
		return scpb.Status_BACKFILL_ONLY
	case mm.Merging():
		__antithesis_instrumentation__.Notify(580401)
		return scpb.Status_MERGE_ONLY
	default:
		__antithesis_instrumentation__.Notify(580402)
		return scpb.Status_PUBLIC
	}
}

func (w *walkCtx) newExpression(expr string) (*scpb.Expression, error) {
	__antithesis_instrumentation__.Notify(580403)
	e, err := parser.ParseExpr(expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(580407)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(580408)
	}
	__antithesis_instrumentation__.Notify(580404)
	var seqIDs catalog.DescriptorIDSet
	{
		__antithesis_instrumentation__.Notify(580409)
		seqIdents, err := seqexpr.GetUsedSequences(e)
		if err != nil {
			__antithesis_instrumentation__.Notify(580411)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(580412)
		}
		__antithesis_instrumentation__.Notify(580410)
		for _, si := range seqIdents {
			__antithesis_instrumentation__.Notify(580413)
			if !si.IsByID() {
				__antithesis_instrumentation__.Notify(580415)
				panic(scerrors.NotImplementedErrorf(nil,
					"sequence %q referenced by name", si.SeqName))
			} else {
				__antithesis_instrumentation__.Notify(580416)
			}
			__antithesis_instrumentation__.Notify(580414)
			seqIDs.Add(descpb.ID(si.SeqID))
		}
	}
	__antithesis_instrumentation__.Notify(580405)
	var typIDs catalog.DescriptorIDSet
	{
		__antithesis_instrumentation__.Notify(580417)
		visitor := &tree.TypeCollectorVisitor{OIDs: make(map[oid.Oid]struct{})}
		tree.WalkExpr(visitor, e)
		for oid := range visitor.OIDs {
			__antithesis_instrumentation__.Notify(580418)
			if !types.IsOIDUserDefinedType(oid) {
				__antithesis_instrumentation__.Notify(580422)
				continue
			} else {
				__antithesis_instrumentation__.Notify(580423)
			}
			__antithesis_instrumentation__.Notify(580419)
			id, err := typedesc.UserDefinedTypeOIDToID(oid)
			if err != nil {
				__antithesis_instrumentation__.Notify(580424)

				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(580425)
			}
			__antithesis_instrumentation__.Notify(580420)
			if _, found := w.cachedTypeIDClosures[id]; !found {
				__antithesis_instrumentation__.Notify(580426)
				desc := w.lookupFn(id)
				typ, err := catalog.AsTypeDescriptor(desc)
				if err != nil {
					__antithesis_instrumentation__.Notify(580428)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(580429)
				}
				__antithesis_instrumentation__.Notify(580427)
				w.cachedTypeIDClosures[id], err = typ.GetIDClosure()
				if err != nil {
					__antithesis_instrumentation__.Notify(580430)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(580431)
				}
			} else {
				__antithesis_instrumentation__.Notify(580432)
			}
			__antithesis_instrumentation__.Notify(580421)
			for id = range w.cachedTypeIDClosures[id] {
				__antithesis_instrumentation__.Notify(580433)
				typIDs.Add(id)
			}
		}
	}
	__antithesis_instrumentation__.Notify(580406)
	return &scpb.Expression{
		Expr:            catpb.Expression(expr),
		UsesTypeIDs:     typIDs.Ordered(),
		UsesSequenceIDs: seqIDs.Ordered(),
	}, nil
}

func newTypeT(t *types.T) (*scpb.TypeT, error) {
	__antithesis_instrumentation__.Notify(580434)
	ids, err := typedesc.GetTypeDescriptorClosure(t)
	if err != nil {
		__antithesis_instrumentation__.Notify(580437)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(580438)
	}
	__antithesis_instrumentation__.Notify(580435)
	var ret catalog.DescriptorIDSet
	for id := range ids {
		__antithesis_instrumentation__.Notify(580439)
		ret.Add(id)
	}
	__antithesis_instrumentation__.Notify(580436)
	ret.Remove(descpb.InvalidID)
	return &scpb.TypeT{
		Type:          t,
		ClosedTypeIDs: ret.Ordered(),
	}, nil
}
