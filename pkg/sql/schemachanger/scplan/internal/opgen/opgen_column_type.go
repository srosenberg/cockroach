package opgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

func init() {
	opRegistry.register((*scpb.ColumnType)(nil),
		toPublic(
			scpb.Status_ABSENT,
			to(scpb.Status_PUBLIC,
				minPhase(scop.PreCommitPhase),
				emit(func(this *scpb.ColumnType) scop.Op {
					return &scop.SetAddedColumnType{
						ColumnType: *protoutil.Clone(this).(*scpb.ColumnType),
					}
				}),
				emit(func(this *scpb.ColumnType) scop.Op {
					if ids := referencedTypeIDs(this); len(ids) > 0 {
						return &scop.UpdateTableBackReferencesInTypes{
							TypeIDs:               ids,
							BackReferencedTableID: this.TableID,
						}
					}
					return nil
				}),
			),
		),
		toAbsent(
			scpb.Status_PUBLIC,
			to(scpb.Status_ABSENT,
				minPhase(scop.PreCommitPhase),
				revertible(false),
				emit(func(this *scpb.ColumnType) scop.Op {
					return &scop.RemoveDroppedColumnType{
						TableID:  this.TableID,
						ColumnID: this.ColumnID,
					}
				}),
				emit(func(this *scpb.ColumnType) scop.Op {
					if ids := referencedTypeIDs(this); len(ids) > 0 {
						return &scop.UpdateTableBackReferencesInTypes{
							TypeIDs:               ids,
							BackReferencedTableID: this.TableID,
						}
					}
					return nil
				}),
			),
		),
	)
}

func referencedTypeIDs(this *scpb.ColumnType) []catid.DescID {
	__antithesis_instrumentation__.Notify(594061)
	var ids catalog.DescriptorIDSet
	if this.ComputeExpr != nil {
		__antithesis_instrumentation__.Notify(594064)
		for _, id := range this.ComputeExpr.UsesTypeIDs {
			__antithesis_instrumentation__.Notify(594065)
			ids.Add(id)
		}
	} else {
		__antithesis_instrumentation__.Notify(594066)
	}
	__antithesis_instrumentation__.Notify(594062)
	for _, id := range this.ClosedTypeIDs {
		__antithesis_instrumentation__.Notify(594067)
		ids.Add(id)
	}
	__antithesis_instrumentation__.Notify(594063)
	return ids.Ordered()
}
