package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type truncateNode struct {
	n *tree.Truncate
}

func (p *planner) Truncate(ctx context.Context, n *tree.Truncate) (planNode, error) {
	__antithesis_instrumentation__.Notify(628406)
	return &truncateNode{n: n}, nil
}

func (t *truncateNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(628407)
	p := params.p
	n := t.n
	ctx := params.ctx

	toTruncate := make(map[descpb.ID]string, len(n.Tables))

	toTraverse := make([]tabledesc.Mutable, 0, len(n.Tables))

	for i := range n.Tables {
		__antithesis_instrumentation__.Notify(628412)
		tn := &n.Tables[i]
		_, tableDesc, err := p.ResolveMutableTableDescriptor(
			ctx, tn, true, tree.ResolveRequireTableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(628415)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628416)
		}
		__antithesis_instrumentation__.Notify(628413)

		if err := p.CheckPrivilege(ctx, tableDesc, privilege.DROP); err != nil {
			__antithesis_instrumentation__.Notify(628417)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628418)
		}
		__antithesis_instrumentation__.Notify(628414)

		toTruncate[tableDesc.ID] = tn.FQString()
		toTraverse = append(toTraverse, *tableDesc)
	}
	__antithesis_instrumentation__.Notify(628408)

	for len(toTraverse) > 0 {
		__antithesis_instrumentation__.Notify(628419)

		idx := len(toTraverse) - 1
		tableDesc := toTraverse[idx]
		toTraverse = toTraverse[:idx]

		maybeEnqueue := func(tableID descpb.ID, msg string) error {
			__antithesis_instrumentation__.Notify(628421)

			if _, ok := toTruncate[tableID]; ok {
				__antithesis_instrumentation__.Notify(628427)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(628428)
			}
			__antithesis_instrumentation__.Notify(628422)
			other, err := p.Descriptors().GetMutableTableVersionByID(ctx, tableID, p.txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(628429)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628430)
			}
			__antithesis_instrumentation__.Notify(628423)

			if n.DropBehavior != tree.DropCascade {
				__antithesis_instrumentation__.Notify(628431)
				return errors.Errorf("%q is %s table %q", tableDesc.Name, msg, other.Name)
			} else {
				__antithesis_instrumentation__.Notify(628432)
			}
			__antithesis_instrumentation__.Notify(628424)
			if err := p.CheckPrivilege(ctx, other, privilege.DROP); err != nil {
				__antithesis_instrumentation__.Notify(628433)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628434)
			}
			__antithesis_instrumentation__.Notify(628425)
			otherName, err := p.getQualifiedTableName(ctx, other)
			if err != nil {
				__antithesis_instrumentation__.Notify(628435)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628436)
			}
			__antithesis_instrumentation__.Notify(628426)
			toTruncate[other.ID] = otherName.FQString()
			toTraverse = append(toTraverse, *other)
			return nil
		}
		__antithesis_instrumentation__.Notify(628420)

		for i := range tableDesc.InboundFKs {
			__antithesis_instrumentation__.Notify(628437)
			fk := &tableDesc.InboundFKs[i]
			if err := maybeEnqueue(fk.OriginTableID, "referenced by foreign key from"); err != nil {
				__antithesis_instrumentation__.Notify(628438)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628439)
			}
		}
	}
	__antithesis_instrumentation__.Notify(628409)

	if err := p.cancelChecker.Check(); err != nil {
		__antithesis_instrumentation__.Notify(628440)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628441)
	}
	__antithesis_instrumentation__.Notify(628410)

	for id, name := range toTruncate {
		__antithesis_instrumentation__.Notify(628442)
		if err := p.truncateTable(ctx, id, tree.AsStringWithFQNames(t.n, params.Ann())); err != nil {
			__antithesis_instrumentation__.Notify(628444)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628445)
		}
		__antithesis_instrumentation__.Notify(628443)

		if err := params.p.logEvent(ctx,
			id,
			&eventpb.TruncateTable{
				TableName: name,
			}); err != nil {
			__antithesis_instrumentation__.Notify(628446)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628447)
		}
	}
	__antithesis_instrumentation__.Notify(628411)

	return nil
}

func (t *truncateNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(628448)
	return false, nil
}
func (t *truncateNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(628449)
	return tree.Datums{}
}
func (t *truncateNode) Close(context.Context) { __antithesis_instrumentation__.Notify(628450) }

var PreservedSplitCountMultiple = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.truncate.preserved_split_count_multiple",
	"set to non-zero to cause TRUNCATE to preserve range splits from the "+
		"table's indexes. The multiple given will be multiplied with the number of "+
		"nodes in the cluster to produce the number of preserved range splits. This "+
		"can improve performance when truncating a table with significant write traffic.",
	4)

func (p *planner) truncateTable(ctx context.Context, id descpb.ID, jobDesc string) error {
	__antithesis_instrumentation__.Notify(628451)

	tableDesc, err := p.Descriptors().GetMutableTableVersionByID(ctx, id, p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(628469)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628470)
	}
	__antithesis_instrumentation__.Notify(628452)

	if err := checkTableForDisallowedMutationsWithTruncate(tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(628471)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628472)
	}
	__antithesis_instrumentation__.Notify(628453)

	if tableDesc.GetDeclarativeSchemaChangerState() != nil {
		__antithesis_instrumentation__.Notify(628473)
		return scerrors.ConcurrentSchemaChangeError(tableDesc)
	} else {
		__antithesis_instrumentation__.Notify(628474)
	}
	__antithesis_instrumentation__.Notify(628454)

	tempIndexMutations := []descpb.DescriptorMutation{}
	for _, m := range tableDesc.Mutations {
		__antithesis_instrumentation__.Notify(628475)
		if idx := m.GetIndex(); idx != nil && func() bool {
			__antithesis_instrumentation__.Notify(628476)
			return idx.UseDeletePreservingEncoding == true
		}() == true {
			__antithesis_instrumentation__.Notify(628477)
			tempIndexMutations = append(tempIndexMutations, m)
		} else {
			__antithesis_instrumentation__.Notify(628478)
			if err := tableDesc.MakeMutationComplete(m); err != nil {
				__antithesis_instrumentation__.Notify(628479)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628480)
			}
		}
	}
	__antithesis_instrumentation__.Notify(628455)

	tableDesc.Mutations = nil
	tableDesc.GCMutations = nil

	oldIndexes := make([]descpb.IndexDescriptor, len(tableDesc.ActiveIndexes()))
	for _, idx := range tableDesc.ActiveIndexes() {
		__antithesis_instrumentation__.Notify(628481)
		oldIndexes[idx.Ordinal()] = idx.IndexDescDeepCopy()
		newIndex := *idx.IndexDesc()
		newIndex.ID = descpb.IndexID(0)
		if idx.Primary() {
			__antithesis_instrumentation__.Notify(628482)
			tableDesc.SetPrimaryIndex(newIndex)
		} else {
			__antithesis_instrumentation__.Notify(628483)
			tableDesc.SetPublicNonPrimaryIndex(idx.Ordinal(), newIndex)
		}
	}
	__antithesis_instrumentation__.Notify(628456)

	version := p.ExecCfg().Settings.Version.ActiveVersion(ctx)
	if err := tableDesc.AllocateIDs(ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(628484)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628485)
	}
	__antithesis_instrumentation__.Notify(628457)

	indexIDMapping := make(map[descpb.IndexID]descpb.IndexID, len(oldIndexes))
	for _, idx := range tableDesc.ActiveIndexes() {
		__antithesis_instrumentation__.Notify(628486)
		indexIDMapping[oldIndexes[idx.Ordinal()].ID] = idx.GetID()
	}
	__antithesis_instrumentation__.Notify(628458)

	dropTime := timeutil.Now().UnixNano()
	droppedIndexes := make([]jobspb.SchemaChangeGCDetails_DroppedIndex, 0, len(oldIndexes))
	for i := range oldIndexes {
		__antithesis_instrumentation__.Notify(628487)
		idx := oldIndexes[i]
		droppedIndexes = append(droppedIndexes, jobspb.SchemaChangeGCDetails_DroppedIndex{
			IndexID:  idx.ID,
			DropTime: dropTime,
		})
	}
	__antithesis_instrumentation__.Notify(628459)

	minimumDropTime := int64(1)
	for _, m := range tempIndexMutations {
		__antithesis_instrumentation__.Notify(628488)
		droppedIndexes = append(droppedIndexes, jobspb.SchemaChangeGCDetails_DroppedIndex{
			IndexID:  m.GetIndex().ID,
			DropTime: minimumDropTime,
		})
	}
	__antithesis_instrumentation__.Notify(628460)

	details := jobspb.SchemaChangeGCDetails{
		Indexes:  droppedIndexes,
		ParentID: tableDesc.ID,
	}
	record := CreateGCJobRecord(jobDesc, p.User(), details)
	if _, err := p.ExecCfg().JobRegistry.CreateAdoptableJobWithTxn(
		ctx, record, p.ExecCfg().JobRegistry.MakeJobID(), p.txn); err != nil {
		__antithesis_instrumentation__.Notify(628489)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628490)
	}
	__antithesis_instrumentation__.Notify(628461)

	st := p.EvalContext().Settings
	if !st.Version.IsActive(ctx, clusterversion.UnsplitRangesInAsyncGCJobs) {
		__antithesis_instrumentation__.Notify(628491)

		if err := p.unsplitRangesForTable(ctx, tableDesc); err != nil {
			__antithesis_instrumentation__.Notify(628492)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628493)
		}
	} else {
		__antithesis_instrumentation__.Notify(628494)
	}
	__antithesis_instrumentation__.Notify(628462)

	oldIndexIDs := make([]descpb.IndexID, len(oldIndexes))
	for i := range oldIndexIDs {
		__antithesis_instrumentation__.Notify(628495)
		oldIndexIDs[i] = oldIndexes[i].ID
	}
	__antithesis_instrumentation__.Notify(628463)
	newIndexIDs := make([]descpb.IndexID, len(tableDesc.ActiveIndexes()))
	newIndexes := tableDesc.ActiveIndexes()
	for i := range newIndexIDs {
		__antithesis_instrumentation__.Notify(628496)
		newIndexIDs[i] = newIndexes[i].GetID()
	}
	__antithesis_instrumentation__.Notify(628464)

	if err := p.copySplitPointsToNewIndexes(ctx, id, oldIndexIDs, newIndexIDs); err != nil {
		__antithesis_instrumentation__.Notify(628497)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628498)
	}
	__antithesis_instrumentation__.Notify(628465)

	swapInfo := &descpb.PrimaryKeySwap{
		OldPrimaryIndexId: oldIndexIDs[0],
		OldIndexes:        oldIndexIDs[1:],
		NewPrimaryIndexId: newIndexIDs[0],
		NewIndexes:        newIndexIDs[1:],
	}
	if err := maybeUpdateZoneConfigsForPKChange(ctx, p.txn, p.ExecCfg(), tableDesc, swapInfo); err != nil {
		__antithesis_instrumentation__.Notify(628499)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628500)
	}
	__antithesis_instrumentation__.Notify(628466)

	if err := p.reassignIndexComments(ctx, tableDesc, indexIDMapping); err != nil {
		__antithesis_instrumentation__.Notify(628501)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628502)
	}
	__antithesis_instrumentation__.Notify(628467)

	if err := p.markTableMutationJobsSuccessful(ctx, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(628503)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628504)
	}
	__antithesis_instrumentation__.Notify(628468)
	tableDesc.MutationJobs = nil

	return p.writeSchemaChange(ctx, tableDesc, descpb.InvalidMutationID, jobDesc)
}

func checkTableForDisallowedMutationsWithTruncate(desc *tabledesc.Mutable) error {
	__antithesis_instrumentation__.Notify(628505)

	for i, m := range desc.AllMutations() {
		__antithesis_instrumentation__.Notify(628507)
		if idx := m.AsIndex(); idx != nil {
			__antithesis_instrumentation__.Notify(628508)

			if !m.Adding() && func() bool {
				__antithesis_instrumentation__.Notify(628509)
				return !idx.IsTemporaryIndexForBackfill() == true
			}() == true {
				__antithesis_instrumentation__.Notify(628510)
				return unimplemented.Newf(
					"TRUNCATE concurrent with ongoing schema change",
					"cannot perform TRUNCATE on %q which has indexes being dropped", desc.GetName())
			} else {
				__antithesis_instrumentation__.Notify(628511)
			}
		} else {
			__antithesis_instrumentation__.Notify(628512)
			if col := m.AsColumn(); col != nil {
				__antithesis_instrumentation__.Notify(628513)
				if col.Dropped() && func() bool {
					__antithesis_instrumentation__.Notify(628514)
					return col.GetType().UserDefined() == true
				}() == true {
					__antithesis_instrumentation__.Notify(628515)
					return unimplemented.Newf(
						"TRUNCATE concurrent with ongoing schema change",
						"cannot perform TRUNCATE on %q which has a column (%q) being "+
							"dropped which depends on another object", desc.GetName(), col.GetName())
				} else {
					__antithesis_instrumentation__.Notify(628516)
				}
			} else {
				__antithesis_instrumentation__.Notify(628517)
				if c := m.AsConstraint(); c != nil {
					__antithesis_instrumentation__.Notify(628518)
					if c.IsCheck() || func() bool {
						__antithesis_instrumentation__.Notify(628520)
						return c.IsNotNull() == true
					}() == true || func() bool {
						__antithesis_instrumentation__.Notify(628521)
						return c.IsForeignKey() == true
					}() == true || func() bool {
						__antithesis_instrumentation__.Notify(628522)
						return c.IsUniqueWithoutIndex() == true
					}() == true {
						__antithesis_instrumentation__.Notify(628523)
						return unimplemented.Newf(
							"TRUNCATE concurrent with ongoing schema change",
							"cannot perform TRUNCATE on %q which has an ongoing %s "+
								"constraint change", desc.GetName(), c.ConstraintToUpdateDesc().ConstraintType)
					} else {
						__antithesis_instrumentation__.Notify(628524)
					}
					__antithesis_instrumentation__.Notify(628519)
					return errors.AssertionFailedf("cannot perform TRUNCATE due to "+
						"unknown constraint type %v on mutation %d in %v", c.ConstraintToUpdateDesc().ConstraintType, i, desc)
				} else {
					__antithesis_instrumentation__.Notify(628525)
					if s := m.AsPrimaryKeySwap(); s != nil {
						__antithesis_instrumentation__.Notify(628526)
						return unimplemented.Newf(
							"TRUNCATE concurrent with ongoing schema change",
							"cannot perform TRUNCATE on %q which has an ongoing primary key "+
								"change", desc.GetName())
					} else {
						__antithesis_instrumentation__.Notify(628527)
						if m.AsComputedColumnSwap() != nil {
							__antithesis_instrumentation__.Notify(628528)
							return unimplemented.Newf(
								"TRUNCATE concurrent with ongoing schema change",
								"cannot perform TRUNCATE on %q which has an ongoing column type "+
									"change", desc.GetName())
						} else {
							__antithesis_instrumentation__.Notify(628529)
							return errors.AssertionFailedf("cannot perform TRUNCATE due to "+
								"concurrent unknown mutation of type %T for mutation %d in %v", m, i, desc)
						}
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(628506)
	return nil
}

func (p *planner) copySplitPointsToNewIndexes(
	ctx context.Context,
	tableID descpb.ID,
	oldIndexIDs []descpb.IndexID,
	newIndexIDs []descpb.IndexID,
) error {
	__antithesis_instrumentation__.Notify(628530)
	if !p.EvalContext().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(628540)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(628541)
	}
	__antithesis_instrumentation__.Notify(628531)

	preservedSplitsMultiple := int(PreservedSplitCountMultiple.Get(p.execCfg.SV()))
	if preservedSplitsMultiple <= 0 {
		__antithesis_instrumentation__.Notify(628542)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(628543)
	}
	__antithesis_instrumentation__.Notify(628532)
	row, err := p.execCfg.InternalExecutor.QueryRowEx(

		ctx, "count-active-nodes", nil, sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"SELECT count(*) FROM crdb_internal.kv_node_status")
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(628544)
		return row == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(628545)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628546)
	}
	__antithesis_instrumentation__.Notify(628533)
	nNodes := int(tree.MustBeDInt(row[0]))
	nSplits := preservedSplitsMultiple * nNodes

	log.Infof(ctx, "making %d new truncate split points (%d * %d)", nSplits, preservedSplitsMultiple, nNodes)

	var b kv.Batch
	tablePrefix := p.execCfg.Codec.TablePrefix(uint32(tableID))

	ranges, err := kvclient.ScanMetaKVs(ctx, p.execCfg.DB.NewTxn(ctx, "truncate-copy-splits"), roachpb.Span{
		Key:    tablePrefix,
		EndKey: tablePrefix.PrefixEnd(),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(628547)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628548)
	}
	__antithesis_instrumentation__.Notify(628534)

	var desc roachpb.RangeDescriptor
	splitPoints := make([][]byte, 0, len(ranges))
	for i := range ranges {
		__antithesis_instrumentation__.Notify(628549)
		if err := ranges[i].ValueProto(&desc); err != nil {
			__antithesis_instrumentation__.Notify(628555)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628556)
		}
		__antithesis_instrumentation__.Notify(628550)

		startKey := desc.StartKey

		restOfKey, foundTable, foundIndex, err := p.execCfg.Codec.DecodeIndexPrefix(roachpb.Key(startKey))
		if err != nil {
			__antithesis_instrumentation__.Notify(628557)

			continue
		} else {
			__antithesis_instrumentation__.Notify(628558)
		}
		__antithesis_instrumentation__.Notify(628551)
		if foundTable != uint32(tableID) {
			__antithesis_instrumentation__.Notify(628559)

			continue
		} else {
			__antithesis_instrumentation__.Notify(628560)
		}
		__antithesis_instrumentation__.Notify(628552)
		var newIndexID descpb.IndexID
		var found bool
		for k := range oldIndexIDs {
			__antithesis_instrumentation__.Notify(628561)
			if oldIndexIDs[k] == descpb.IndexID(foundIndex) {
				__antithesis_instrumentation__.Notify(628562)
				newIndexID = newIndexIDs[k]
				found = true
			} else {
				__antithesis_instrumentation__.Notify(628563)
			}
		}
		__antithesis_instrumentation__.Notify(628553)
		if !found {
			__antithesis_instrumentation__.Notify(628564)

			continue
		} else {
			__antithesis_instrumentation__.Notify(628565)
		}
		__antithesis_instrumentation__.Notify(628554)

		newStartKey := append(p.execCfg.Codec.IndexPrefix(uint32(tableID), uint32(newIndexID)), restOfKey...)
		splitPoints = append(splitPoints, newStartKey)
	}
	__antithesis_instrumentation__.Notify(628535)

	if len(splitPoints) == 0 {
		__antithesis_instrumentation__.Notify(628566)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(628567)
	}
	__antithesis_instrumentation__.Notify(628536)

	step := float64(len(splitPoints)) / float64(nSplits)
	if step < 1 {
		__antithesis_instrumentation__.Notify(628568)
		step = 1
	} else {
		__antithesis_instrumentation__.Notify(628569)
	}
	__antithesis_instrumentation__.Notify(628537)
	expirationTime := kvserverbase.SplitByLoadMergeDelay.Get(p.execCfg.SV()).Nanoseconds()
	for i := 0; i < nSplits; i++ {
		__antithesis_instrumentation__.Notify(628570)

		idx := int(step * float64(i))
		if idx >= len(splitPoints) {
			__antithesis_instrumentation__.Notify(628572)
			break
		} else {
			__antithesis_instrumentation__.Notify(628573)
		}
		__antithesis_instrumentation__.Notify(628571)
		sp := splitPoints[idx]

		maxJitter := expirationTime / 5
		jitter := rand.Int63n(maxJitter*2) - maxJitter
		expirationTime += jitter

		log.Infof(ctx, "truncate sending split request for key %s", sp)
		b.AddRawRequest(&roachpb.AdminSplitRequest{
			RequestHeader: roachpb.RequestHeader{
				Key: sp,
			},
			SplitKey:       sp,
			ExpirationTime: p.execCfg.Clock.Now().Add(expirationTime, 0),
		})
	}
	__antithesis_instrumentation__.Notify(628538)

	if err = p.txn.DB().Run(ctx, &b); err != nil {
		__antithesis_instrumentation__.Notify(628574)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628575)
	}
	__antithesis_instrumentation__.Notify(628539)

	b = kv.Batch{}
	b.AddRawRequest(&roachpb.AdminScatterRequest{

		RequestHeader: roachpb.RequestHeader{
			Key:    p.execCfg.Codec.IndexPrefix(uint32(tableID), uint32(newIndexIDs[0])),
			EndKey: p.execCfg.Codec.IndexPrefix(uint32(tableID), uint32(newIndexIDs[len(newIndexIDs)-1])).PrefixEnd(),
		},
		RandomizeLeases: true,
	})

	return p.txn.DB().Run(ctx, &b)
}

func (p *planner) reassignIndexComments(
	ctx context.Context, table *tabledesc.Mutable, indexIDMapping map[descpb.IndexID]descpb.IndexID,
) error {
	__antithesis_instrumentation__.Notify(628576)

	row, err := p.extendedEvalCtx.ExecCfg.InternalExecutor.QueryRowEx(
		ctx,
		"update-table-comments",
		p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		`SELECT count(*) FROM system.comments WHERE object_id = $1 AND type = $2`,
		table.ID,
		keys.IndexCommentType,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(628580)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628581)
	}
	__antithesis_instrumentation__.Notify(628577)
	if row == nil {
		__antithesis_instrumentation__.Notify(628582)
		return errors.New("failed to update table comments")
	} else {
		__antithesis_instrumentation__.Notify(628583)
	}
	__antithesis_instrumentation__.Notify(628578)
	if int(tree.MustBeDInt(row[0])) > 0 {
		__antithesis_instrumentation__.Notify(628584)
		for old, new := range indexIDMapping {
			__antithesis_instrumentation__.Notify(628585)
			if _, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.ExecEx(
				ctx,
				"update-table-comments",
				p.txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				`UPDATE system.comments SET sub_id=$1 WHERE sub_id=$2 AND object_id=$3 AND type=$4`,
				new,
				old,
				table.ID,
				keys.IndexCommentType,
			); err != nil {
				__antithesis_instrumentation__.Notify(628586)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628587)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(628588)
	}
	__antithesis_instrumentation__.Notify(628579)
	return nil
}
