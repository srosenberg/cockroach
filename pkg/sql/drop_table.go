package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type dropTableNode struct {
	n *tree.DropTable

	td map[descpb.ID]toDelete
}

type toDelete struct {
	tn   tree.ObjectName
	desc *tabledesc.Mutable
}

func (p *planner) DropTable(ctx context.Context, n *tree.DropTable) (planNode, error) {
	__antithesis_instrumentation__.Notify(469529)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"DROP TABLE",
	); err != nil {
		__antithesis_instrumentation__.Notify(469534)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469535)
	}
	__antithesis_instrumentation__.Notify(469530)

	td := make(map[descpb.ID]toDelete, len(n.Names))
	for i := range n.Names {
		__antithesis_instrumentation__.Notify(469536)
		tn := &n.Names[i]
		droppedDesc, err := p.prepareDrop(ctx, tn, !n.IfExists, tree.ResolveRequireTableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(469539)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469540)
		}
		__antithesis_instrumentation__.Notify(469537)
		if droppedDesc == nil {
			__antithesis_instrumentation__.Notify(469541)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469542)
		}
		__antithesis_instrumentation__.Notify(469538)

		td[droppedDesc.ID] = toDelete{tn, droppedDesc}
	}
	__antithesis_instrumentation__.Notify(469531)

	for _, toDel := range td {
		__antithesis_instrumentation__.Notify(469543)
		droppedDesc := toDel.desc
		for i := range droppedDesc.InboundFKs {
			__antithesis_instrumentation__.Notify(469546)
			ref := &droppedDesc.InboundFKs[i]
			if _, ok := td[ref.OriginTableID]; !ok {
				__antithesis_instrumentation__.Notify(469547)
				if err := p.canRemoveFKBackreference(ctx, droppedDesc.Name, ref, n.DropBehavior); err != nil {
					__antithesis_instrumentation__.Notify(469548)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(469549)
				}
			} else {
				__antithesis_instrumentation__.Notify(469550)
			}
		}
		__antithesis_instrumentation__.Notify(469544)
		for _, ref := range droppedDesc.DependedOnBy {
			__antithesis_instrumentation__.Notify(469551)
			if _, ok := td[ref.ID]; !ok {
				__antithesis_instrumentation__.Notify(469552)
				if err := p.canRemoveDependentView(ctx, droppedDesc, ref, n.DropBehavior); err != nil {
					__antithesis_instrumentation__.Notify(469553)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(469554)
				}
			} else {
				__antithesis_instrumentation__.Notify(469555)
			}
		}
		__antithesis_instrumentation__.Notify(469545)
		if err := p.canRemoveAllTableOwnedSequences(ctx, droppedDesc, n.DropBehavior); err != nil {
			__antithesis_instrumentation__.Notify(469556)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(469557)
		}

	}
	__antithesis_instrumentation__.Notify(469532)

	if len(td) == 0 {
		__antithesis_instrumentation__.Notify(469558)
		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(469559)
	}
	__antithesis_instrumentation__.Notify(469533)
	return &dropTableNode{n: n, td: td}, nil
}

func (n *dropTableNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(469560) }

func (n *dropTableNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(469561)
	telemetry.Inc(sqltelemetry.SchemaChangeDropCounter("table"))

	ctx := params.ctx
	for _, toDel := range n.td {
		__antithesis_instrumentation__.Notify(469563)
		droppedDesc := toDel.desc
		if droppedDesc == nil {
			__antithesis_instrumentation__.Notify(469566)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469567)
		}
		__antithesis_instrumentation__.Notify(469564)

		droppedViews, err := params.p.dropTableImpl(
			ctx,
			droppedDesc,
			false,
			tree.AsStringWithFQNames(n.n, params.Ann()),
			n.n.DropBehavior,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469568)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469569)
		}
		__antithesis_instrumentation__.Notify(469565)

		if err := params.p.logEvent(params.ctx,
			droppedDesc.ID,
			&eventpb.DropTable{
				TableName:           toDel.tn.FQString(),
				CascadeDroppedViews: droppedViews,
			}); err != nil {
			__antithesis_instrumentation__.Notify(469570)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469571)
		}
	}
	__antithesis_instrumentation__.Notify(469562)
	return nil
}

func (*dropTableNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(469572)
	return false, nil
}
func (*dropTableNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(469573)
	return tree.Datums{}
}
func (*dropTableNode) Close(context.Context) { __antithesis_instrumentation__.Notify(469574) }

func (p *planner) prepareDrop(
	ctx context.Context, name *tree.TableName, required bool, requiredType tree.RequiredTableKind,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(469575)
	_, tableDesc, err := p.ResolveMutableTableDescriptor(ctx, name, required, requiredType)
	if err != nil {
		__antithesis_instrumentation__.Notify(469579)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469580)
	}
	__antithesis_instrumentation__.Notify(469576)
	if tableDesc == nil {
		__antithesis_instrumentation__.Notify(469581)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469582)
	}
	__antithesis_instrumentation__.Notify(469577)
	if err := p.canDropTable(ctx, tableDesc, true); err != nil {
		__antithesis_instrumentation__.Notify(469583)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469584)
	}
	__antithesis_instrumentation__.Notify(469578)
	return tableDesc, nil
}

func (p *planner) canDropTable(
	ctx context.Context, tableDesc *tabledesc.Mutable, checkOwnership bool,
) error {
	__antithesis_instrumentation__.Notify(469585)
	var err error
	hasOwnership := false

	if checkOwnership {
		__antithesis_instrumentation__.Notify(469588)

		hasOwnership, err = p.HasOwnershipOnSchema(
			ctx, tableDesc.GetParentSchemaID(), tableDesc.GetParentID())
		if err != nil {
			__antithesis_instrumentation__.Notify(469589)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469590)
		}
	} else {
		__antithesis_instrumentation__.Notify(469591)
	}
	__antithesis_instrumentation__.Notify(469586)
	if !hasOwnership {
		__antithesis_instrumentation__.Notify(469592)
		return p.CheckPrivilege(ctx, tableDesc, privilege.DROP)
	} else {
		__antithesis_instrumentation__.Notify(469593)
	}
	__antithesis_instrumentation__.Notify(469587)

	return nil
}

func (p *planner) canRemoveFKBackreference(
	ctx context.Context, from string, ref *descpb.ForeignKeyConstraint, behavior tree.DropBehavior,
) error {
	__antithesis_instrumentation__.Notify(469594)
	table, err := p.Descriptors().GetMutableTableVersionByID(ctx, ref.OriginTableID, p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(469597)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469598)
	}
	__antithesis_instrumentation__.Notify(469595)
	if behavior != tree.DropCascade {
		__antithesis_instrumentation__.Notify(469599)
		return fmt.Errorf("%q is referenced by foreign key from table %q", from, table.Name)
	} else {
		__antithesis_instrumentation__.Notify(469600)
	}
	__antithesis_instrumentation__.Notify(469596)

	return p.CheckPrivilege(ctx, table, privilege.CREATE)
}

func (p *planner) dropTableImpl(
	ctx context.Context,
	tableDesc *tabledesc.Mutable,
	droppingParent bool,
	jobDesc string,
	behavior tree.DropBehavior,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(469601)
	var droppedViews []string

	outboundFKs := append([]descpb.ForeignKeyConstraint(nil), tableDesc.OutboundFKs...)
	for i := range outboundFKs {
		__antithesis_instrumentation__.Notify(469609)
		ref := &tableDesc.OutboundFKs[i]
		if err := p.removeFKBackReference(ctx, tableDesc, ref); err != nil {
			__antithesis_instrumentation__.Notify(469610)
			return droppedViews, err
		} else {
			__antithesis_instrumentation__.Notify(469611)
		}
	}
	__antithesis_instrumentation__.Notify(469602)
	tableDesc.OutboundFKs = nil

	inboundFKs := append([]descpb.ForeignKeyConstraint(nil), tableDesc.InboundFKs...)
	for i := range inboundFKs {
		__antithesis_instrumentation__.Notify(469612)
		ref := &tableDesc.InboundFKs[i]
		if err := p.removeFKForBackReference(ctx, tableDesc, ref); err != nil {
			__antithesis_instrumentation__.Notify(469613)
			return droppedViews, err
		} else {
			__antithesis_instrumentation__.Notify(469614)
		}
	}
	__antithesis_instrumentation__.Notify(469603)
	tableDesc.InboundFKs = nil

	for _, col := range tableDesc.PublicColumns() {
		__antithesis_instrumentation__.Notify(469615)
		if err := p.removeSequenceDependencies(ctx, tableDesc, col); err != nil {
			__antithesis_instrumentation__.Notify(469616)
			return droppedViews, err
		} else {
			__antithesis_instrumentation__.Notify(469617)
		}
	}
	__antithesis_instrumentation__.Notify(469604)

	for _, col := range tableDesc.PublicColumns() {
		__antithesis_instrumentation__.Notify(469618)
		if err := p.dropSequencesOwnedByCol(ctx, col, !droppingParent, behavior); err != nil {
			__antithesis_instrumentation__.Notify(469619)
			return droppedViews, err
		} else {
			__antithesis_instrumentation__.Notify(469620)
		}
	}
	__antithesis_instrumentation__.Notify(469605)

	dependedOnBy := append([]descpb.TableDescriptor_Reference(nil), tableDesc.DependedOnBy...)
	for _, ref := range dependedOnBy {
		__antithesis_instrumentation__.Notify(469621)
		viewDesc, err := p.getViewDescForCascade(
			ctx, string(tableDesc.DescriptorType()), tableDesc.Name, tableDesc.ParentID, ref.ID, tree.DropCascade,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469626)
			return droppedViews, err
		} else {
			__antithesis_instrumentation__.Notify(469627)
		}
		__antithesis_instrumentation__.Notify(469622)

		if viewDesc.Dropped() {
			__antithesis_instrumentation__.Notify(469628)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469629)
		}
		__antithesis_instrumentation__.Notify(469623)
		cascadedViews, err := p.dropViewImpl(ctx, viewDesc, !droppingParent, "dropping dependent view", tree.DropCascade)
		if err != nil {
			__antithesis_instrumentation__.Notify(469630)
			return droppedViews, err
		} else {
			__antithesis_instrumentation__.Notify(469631)
		}
		__antithesis_instrumentation__.Notify(469624)

		qualifiedView, err := p.getQualifiedTableName(ctx, viewDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(469632)
			return droppedViews, err
		} else {
			__antithesis_instrumentation__.Notify(469633)
		}
		__antithesis_instrumentation__.Notify(469625)

		droppedViews = append(droppedViews, cascadedViews...)
		droppedViews = append(droppedViews, qualifiedView.FQString())
	}
	__antithesis_instrumentation__.Notify(469606)

	err := p.removeTableComments(ctx, tableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(469634)
		return droppedViews, err
	} else {
		__antithesis_instrumentation__.Notify(469635)
	}
	__antithesis_instrumentation__.Notify(469607)

	if err := p.removeBackRefsFromAllTypesInTable(ctx, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(469636)
		return droppedViews, err
	} else {
		__antithesis_instrumentation__.Notify(469637)
	}
	__antithesis_instrumentation__.Notify(469608)

	err = p.initiateDropTable(ctx, tableDesc, !droppingParent, jobDesc)
	return droppedViews, err
}

func UnsplitRangesInSpan(ctx context.Context, kvDB *kv.DB, span roachpb.Span) error {
	__antithesis_instrumentation__.Notify(469638)
	ranges, err := kvclient.ScanMetaKVs(ctx, kvDB.NewTxn(ctx, "unsplit-ranges-in-span"), span)
	if err != nil {
		__antithesis_instrumentation__.Notify(469641)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469642)
	}
	__antithesis_instrumentation__.Notify(469639)
	for _, r := range ranges {
		__antithesis_instrumentation__.Notify(469643)
		var desc roachpb.RangeDescriptor
		if err := r.ValueProto(&desc); err != nil {
			__antithesis_instrumentation__.Notify(469646)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469647)
		}
		__antithesis_instrumentation__.Notify(469644)

		if !span.ContainsKey(desc.StartKey.AsRawKey()) {
			__antithesis_instrumentation__.Notify(469648)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469649)
		}
		__antithesis_instrumentation__.Notify(469645)

		if !desc.GetStickyBit().IsEmpty() {
			__antithesis_instrumentation__.Notify(469650)

			if err := kvDB.AdminUnsplit(ctx, desc.StartKey); err != nil && func() bool {
				__antithesis_instrumentation__.Notify(469651)
				return !strings.Contains(err.Error(), "is not the start of a range") == true
			}() == true {
				__antithesis_instrumentation__.Notify(469652)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469653)
			}
		} else {
			__antithesis_instrumentation__.Notify(469654)
		}
	}
	__antithesis_instrumentation__.Notify(469640)

	return nil
}

func (p *planner) unsplitRangesForTable(ctx context.Context, tableDesc *tabledesc.Mutable) error {
	__antithesis_instrumentation__.Notify(469655)

	if !p.ExecCfg().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(469657)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(469658)
	}
	__antithesis_instrumentation__.Notify(469656)

	span := tableDesc.TableSpan(p.ExecCfg().Codec)
	return UnsplitRangesInSpan(ctx, p.execCfg.DB, span)
}

func (p *planner) initiateDropTable(
	ctx context.Context, tableDesc *tabledesc.Mutable, queueJob bool, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(469659)
	if tableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(469666)
		return errors.Errorf("table %q is already being dropped", tableDesc.Name)
	} else {
		__antithesis_instrumentation__.Notify(469667)
	}
	__antithesis_instrumentation__.Notify(469660)

	if tableDesc.GetDeclarativeSchemaChangerState() != nil {
		__antithesis_instrumentation__.Notify(469668)
		return scerrors.ConcurrentSchemaChangeError(tableDesc)
	} else {
		__antithesis_instrumentation__.Notify(469669)
	}
	__antithesis_instrumentation__.Notify(469661)

	if tableDesc.IsTable() {
		__antithesis_instrumentation__.Notify(469670)
		tableDesc.DropTime = timeutil.Now().UnixNano()
	} else {
		__antithesis_instrumentation__.Notify(469671)
	}
	__antithesis_instrumentation__.Notify(469662)

	st := p.EvalContext().Settings
	if !st.Version.IsActive(ctx, clusterversion.UnsplitRangesInAsyncGCJobs) {
		__antithesis_instrumentation__.Notify(469672)

		if err := p.unsplitRangesForTable(ctx, tableDesc); err != nil {
			__antithesis_instrumentation__.Notify(469673)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469674)
		}
	} else {
		__antithesis_instrumentation__.Notify(469675)
	}
	__antithesis_instrumentation__.Notify(469663)

	tableDesc.SetDropped()

	b := p.txn.NewBatch()
	p.dropNamespaceEntry(ctx, b, tableDesc)
	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(469676)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469677)
	}
	__antithesis_instrumentation__.Notify(469664)

	if err := p.markTableMutationJobsSuccessful(ctx, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(469678)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469679)
	}
	__antithesis_instrumentation__.Notify(469665)

	return p.writeDropTable(ctx, tableDesc, queueJob, jobDesc)
}

func (p *planner) markTableMutationJobsSuccessful(
	ctx context.Context, tableDesc *tabledesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(469680)
	for _, mj := range tableDesc.MutationJobs {
		__antithesis_instrumentation__.Notify(469682)
		jobID := mj.JobID

		if record, exists := p.ExtendedEvalContext().SchemaChangeJobRecords[tableDesc.ID]; exists && func() bool {
			__antithesis_instrumentation__.Notify(469685)
			return record.JobID == jobID == true
		}() == true {
			__antithesis_instrumentation__.Notify(469686)
			delete(p.ExtendedEvalContext().SchemaChangeJobRecords, tableDesc.ID)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469687)
		}
		__antithesis_instrumentation__.Notify(469683)
		mutationJob, err := p.execCfg.JobRegistry.LoadJobWithTxn(ctx, jobID, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(469688)
			if jobs.HasJobNotFoundError(err) {
				__antithesis_instrumentation__.Notify(469690)
				log.Warningf(ctx, "mutation job %d not found", jobID)
				continue
			} else {
				__antithesis_instrumentation__.Notify(469691)
			}
			__antithesis_instrumentation__.Notify(469689)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469692)
		}
		__antithesis_instrumentation__.Notify(469684)
		if err := mutationJob.Update(
			ctx, p.txn, func(txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater) error {
				__antithesis_instrumentation__.Notify(469693)
				status := md.Status
				switch status {
				case jobs.StatusSucceeded, jobs.StatusCanceled, jobs.StatusFailed, jobs.StatusRevertFailed:
					__antithesis_instrumentation__.Notify(469695)
					log.Warningf(ctx, "mutation job %d in unexpected state %s", jobID, status)
					return nil
				case jobs.StatusRunning, jobs.StatusPending:
					__antithesis_instrumentation__.Notify(469696)
					status = jobs.StatusSucceeded
				default:
					__antithesis_instrumentation__.Notify(469697)

					status = jobs.StatusFailed
				}
				__antithesis_instrumentation__.Notify(469694)
				log.Infof(ctx, "marking mutation job %d for dropped table as %s", jobID, status)
				ju.UpdateStatus(status)
				return nil
			}); err != nil {
			__antithesis_instrumentation__.Notify(469698)
			return errors.Wrap(err, "updating mutation job for dropped table")
		} else {
			__antithesis_instrumentation__.Notify(469699)
		}
	}
	__antithesis_instrumentation__.Notify(469681)
	return nil
}

func (p *planner) removeFKForBackReference(
	ctx context.Context, tableDesc *tabledesc.Mutable, ref *descpb.ForeignKeyConstraint,
) error {
	__antithesis_instrumentation__.Notify(469700)
	var originTableDesc *tabledesc.Mutable

	if tableDesc.ID == ref.OriginTableID {
		__antithesis_instrumentation__.Notify(469705)
		originTableDesc = tableDesc
	} else {
		__antithesis_instrumentation__.Notify(469706)
		lookup, err := p.Descriptors().GetMutableTableVersionByID(ctx, ref.OriginTableID, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(469708)
			return errors.Wrapf(err, "error resolving origin table ID %d", ref.OriginTableID)
		} else {
			__antithesis_instrumentation__.Notify(469709)
		}
		__antithesis_instrumentation__.Notify(469707)
		originTableDesc = lookup
	}
	__antithesis_instrumentation__.Notify(469701)
	if originTableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(469710)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(469711)
	}
	__antithesis_instrumentation__.Notify(469702)

	if err := removeFKForBackReferenceFromTable(originTableDesc, ref, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(469712)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469713)
	}
	__antithesis_instrumentation__.Notify(469703)

	name, err := p.getQualifiedTableName(ctx, tableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(469714)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469715)
	}
	__antithesis_instrumentation__.Notify(469704)
	jobDesc := fmt.Sprintf("updating table %q after removing constraint %q from table %q", originTableDesc.GetName(), ref.Name, name.FQString())
	return p.writeSchemaChange(ctx, originTableDesc, descpb.InvalidMutationID, jobDesc)
}

func removeFKForBackReferenceFromTable(
	originTableDesc *tabledesc.Mutable,
	backref *descpb.ForeignKeyConstraint,
	referencedTableDesc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(469716)
	matchIdx := -1
	for i, fk := range originTableDesc.OutboundFKs {
		__antithesis_instrumentation__.Notify(469719)
		if fk.ReferencedTableID == referencedTableDesc.GetID() && func() bool {
			__antithesis_instrumentation__.Notify(469720)
			return fk.Name == backref.Name == true
		}() == true {
			__antithesis_instrumentation__.Notify(469721)

			matchIdx = i
			break
		} else {
			__antithesis_instrumentation__.Notify(469722)
		}
	}
	__antithesis_instrumentation__.Notify(469717)
	if matchIdx == -1 {
		__antithesis_instrumentation__.Notify(469723)

		return errors.AssertionFailedf("there was no foreign key constraint "+
			"for backreference %v on table %q", backref, originTableDesc.Name)
	} else {
		__antithesis_instrumentation__.Notify(469724)
	}
	__antithesis_instrumentation__.Notify(469718)

	originTableDesc.OutboundFKs = append(
		originTableDesc.OutboundFKs[:matchIdx],
		originTableDesc.OutboundFKs[matchIdx+1:]...)
	return nil
}

func (p *planner) removeFKBackReference(
	ctx context.Context, tableDesc *tabledesc.Mutable, ref *descpb.ForeignKeyConstraint,
) error {
	__antithesis_instrumentation__.Notify(469725)
	var referencedTableDesc *tabledesc.Mutable

	if tableDesc.ID == ref.ReferencedTableID {
		__antithesis_instrumentation__.Notify(469730)
		referencedTableDesc = tableDesc
	} else {
		__antithesis_instrumentation__.Notify(469731)
		lookup, err := p.Descriptors().GetMutableTableVersionByID(ctx, ref.ReferencedTableID, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(469733)
			return errors.Wrapf(err, "error resolving referenced table ID %d", ref.ReferencedTableID)
		} else {
			__antithesis_instrumentation__.Notify(469734)
		}
		__antithesis_instrumentation__.Notify(469732)
		referencedTableDesc = lookup
	}
	__antithesis_instrumentation__.Notify(469726)
	if referencedTableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(469735)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(469736)
	}
	__antithesis_instrumentation__.Notify(469727)

	if err := removeFKBackReferenceFromTable(referencedTableDesc, ref.Name, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(469737)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469738)
	}
	__antithesis_instrumentation__.Notify(469728)

	name, err := p.getQualifiedTableName(ctx, tableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(469739)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469740)
	}
	__antithesis_instrumentation__.Notify(469729)
	jobDesc := fmt.Sprintf("updating table %q after removing constraint %q from table %q", referencedTableDesc.GetName(), ref.Name, name.FQString())

	return p.writeSchemaChange(ctx, referencedTableDesc, descpb.InvalidMutationID, jobDesc)
}

func removeFKBackReferenceFromTable(
	referencedTableDesc *tabledesc.Mutable, fkName string, originTableDesc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(469741)
	matchIdx := -1
	for i, backref := range referencedTableDesc.InboundFKs {
		__antithesis_instrumentation__.Notify(469744)
		if backref.OriginTableID == originTableDesc.GetID() && func() bool {
			__antithesis_instrumentation__.Notify(469745)
			return backref.Name == fkName == true
		}() == true {
			__antithesis_instrumentation__.Notify(469746)

			matchIdx = i
			break
		} else {
			__antithesis_instrumentation__.Notify(469747)
		}
	}
	__antithesis_instrumentation__.Notify(469742)
	if matchIdx == -1 {
		__antithesis_instrumentation__.Notify(469748)

		return errors.AssertionFailedf("there was no foreign key backreference "+
			"for constraint %q on table %q", fkName, originTableDesc.GetName())
	} else {
		__antithesis_instrumentation__.Notify(469749)
	}
	__antithesis_instrumentation__.Notify(469743)

	referencedTableDesc.InboundFKs = append(referencedTableDesc.InboundFKs[:matchIdx],
		referencedTableDesc.InboundFKs[matchIdx+1:]...)
	return nil
}

func removeMatchingReferences(
	refs []descpb.TableDescriptor_Reference, id descpb.ID,
) []descpb.TableDescriptor_Reference {
	__antithesis_instrumentation__.Notify(469750)
	updatedRefs := refs[:0]
	for _, ref := range refs {
		__antithesis_instrumentation__.Notify(469752)
		if ref.ID != id {
			__antithesis_instrumentation__.Notify(469753)
			updatedRefs = append(updatedRefs, ref)
		} else {
			__antithesis_instrumentation__.Notify(469754)
		}
	}
	__antithesis_instrumentation__.Notify(469751)
	return updatedRefs
}

func (p *planner) removeTableComments(ctx context.Context, tableDesc *tabledesc.Mutable) error {
	__antithesis_instrumentation__.Notify(469755)
	return p.execCfg.DescMetadaUpdaterFactory.NewMetadataUpdater(
		ctx, p.Txn(), p.SessionData(),
	).DeleteAllCommentsForTables(catalog.MakeDescriptorIDSet(tableDesc.GetID()))
}
