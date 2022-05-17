package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type dropIndexNode struct {
	n        *tree.DropIndex
	idxNames []fullIndexName
}

func (p *planner) DropIndex(ctx context.Context, n *tree.DropIndex) (planNode, error) {
	__antithesis_instrumentation__.Notify(468898)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"DROP INDEX",
	); err != nil {
		__antithesis_instrumentation__.Notify(468901)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(468902)
	}
	__antithesis_instrumentation__.Notify(468899)

	idxNames := make([]fullIndexName, 0, len(n.IndexList))
	for _, index := range n.IndexList {
		__antithesis_instrumentation__.Notify(468903)
		tn, tableDesc, err := expandMutableIndexName(ctx, p, index, !n.IfExists)
		if err != nil {
			__antithesis_instrumentation__.Notify(468907)

			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(468908)
		}
		__antithesis_instrumentation__.Notify(468904)
		if tableDesc == nil {
			__antithesis_instrumentation__.Notify(468909)

			continue
		} else {
			__antithesis_instrumentation__.Notify(468910)
		}
		__antithesis_instrumentation__.Notify(468905)

		if err := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE); err != nil {
			__antithesis_instrumentation__.Notify(468911)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(468912)
		}
		__antithesis_instrumentation__.Notify(468906)

		idxNames = append(idxNames, fullIndexName{tn: tn, idxName: index.Index})
	}
	__antithesis_instrumentation__.Notify(468900)
	return &dropIndexNode{n: n, idxNames: idxNames}, nil
}

func (n *dropIndexNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(468913) }

func (n *dropIndexNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(468914)
	telemetry.Inc(sqltelemetry.SchemaChangeDropCounter("index"))

	if n.n.Concurrently {
		__antithesis_instrumentation__.Notify(468917)
		params.p.BufferClientNotice(
			params.ctx,
			pgnotice.Newf("CONCURRENTLY is not required as all indexes are dropped concurrently"),
		)
	} else {
		__antithesis_instrumentation__.Notify(468918)
	}
	__antithesis_instrumentation__.Notify(468915)

	ctx := params.ctx
	for _, index := range n.idxNames {
		__antithesis_instrumentation__.Notify(468919)

		_, tableDesc, err := params.p.ResolveMutableTableDescriptor(
			ctx, index.tn, true, tree.ResolveRequireTableOrViewDesc)
		if sqlerrors.IsUndefinedRelationError(err) {
			__antithesis_instrumentation__.Notify(468928)

			return errors.NewAssertionErrorWithWrappedErrf(err,
				"table descriptor for %q became unavailable within same txn",
				tree.ErrString(index.tn))
		} else {
			__antithesis_instrumentation__.Notify(468929)
		}
		__antithesis_instrumentation__.Notify(468920)
		if err != nil {
			__antithesis_instrumentation__.Notify(468930)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468931)
		}
		__antithesis_instrumentation__.Notify(468921)

		if tableDesc.IsView() && func() bool {
			__antithesis_instrumentation__.Notify(468932)
			return !tableDesc.MaterializedView() == true
		}() == true {
			__antithesis_instrumentation__.Notify(468933)
			return pgerror.Newf(pgcode.WrongObjectType, "%q is not a table or materialized view", tableDesc.Name)
		} else {
			__antithesis_instrumentation__.Notify(468934)
		}
		__antithesis_instrumentation__.Notify(468922)

		idx, _ := tableDesc.FindIndexWithName(string(index.idxName))
		var shardColName string

		if idx != nil && func() bool {
			__antithesis_instrumentation__.Notify(468935)
			return idx.IsSharded() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(468936)
			return !idx.Dropped() == true
		}() == true {
			__antithesis_instrumentation__.Notify(468937)
			shardColName = idx.GetShardColumnName()
		} else {
			__antithesis_instrumentation__.Notify(468938)
		}
		__antithesis_instrumentation__.Notify(468923)

		keyColumnOfOtherIndex := func(colID descpb.ColumnID) bool {
			__antithesis_instrumentation__.Notify(468939)
			for _, otherIdx := range tableDesc.AllIndexes() {
				__antithesis_instrumentation__.Notify(468941)
				if otherIdx.GetID() == idx.GetID() {
					__antithesis_instrumentation__.Notify(468943)
					continue
				} else {
					__antithesis_instrumentation__.Notify(468944)
				}
				__antithesis_instrumentation__.Notify(468942)
				if otherIdx.CollectKeyColumnIDs().Contains(colID) {
					__antithesis_instrumentation__.Notify(468945)
					return true
				} else {
					__antithesis_instrumentation__.Notify(468946)
				}
			}
			__antithesis_instrumentation__.Notify(468940)
			return false
		}
		__antithesis_instrumentation__.Notify(468924)

		columnsDropped := false
		if idx != nil {
			__antithesis_instrumentation__.Notify(468947)
			for i, count := 0, idx.NumKeyColumns(); i < count; i++ {
				__antithesis_instrumentation__.Notify(468948)
				id := idx.GetKeyColumnID(i)
				col, err := tableDesc.FindColumnWithID(id)
				if err != nil {
					__antithesis_instrumentation__.Notify(468950)
					return err
				} else {
					__antithesis_instrumentation__.Notify(468951)
				}
				__antithesis_instrumentation__.Notify(468949)
				if col.IsExpressionIndexColumn() && func() bool {
					__antithesis_instrumentation__.Notify(468952)
					return !keyColumnOfOtherIndex(col.GetID()) == true
				}() == true {
					__antithesis_instrumentation__.Notify(468953)
					n.queueDropColumn(tableDesc, col)
					columnsDropped = true
				} else {
					__antithesis_instrumentation__.Notify(468954)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(468955)
		}
		__antithesis_instrumentation__.Notify(468925)

		if err := params.p.dropIndexByName(
			ctx, index.tn, index.idxName, tableDesc, n.n.IfExists, n.n.DropBehavior, checkIdxConstraint,
			tree.AsStringWithFQNames(n.n, params.Ann()),
		); err != nil {
			__antithesis_instrumentation__.Notify(468956)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468957)
		}
		__antithesis_instrumentation__.Notify(468926)

		if shardColName != "" {
			__antithesis_instrumentation__.Notify(468958)
			ok, err := n.maybeQueueDropShardColumn(tableDesc, shardColName)
			if err != nil {
				__antithesis_instrumentation__.Notify(468960)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468961)
			}
			__antithesis_instrumentation__.Notify(468959)
			columnsDropped = columnsDropped || func() bool {
				__antithesis_instrumentation__.Notify(468962)
				return ok == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(468963)
		}
		__antithesis_instrumentation__.Notify(468927)

		if columnsDropped {
			__antithesis_instrumentation__.Notify(468964)
			if err := n.finalizeDropColumn(params, tableDesc); err != nil {
				__antithesis_instrumentation__.Notify(468965)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468966)
			}
		} else {
			__antithesis_instrumentation__.Notify(468967)
		}

	}
	__antithesis_instrumentation__.Notify(468916)
	return nil
}

func (n *dropIndexNode) queueDropColumn(tableDesc *tabledesc.Mutable, col catalog.Column) {
	__antithesis_instrumentation__.Notify(468968)
	tableDesc.AddColumnMutation(col.ColumnDesc(), descpb.DescriptorMutation_DROP)
	for i := range tableDesc.Columns {
		__antithesis_instrumentation__.Notify(468969)
		if tableDesc.Columns[i].ID == col.GetID() {
			__antithesis_instrumentation__.Notify(468970)

			tableDesc.Columns = append(tableDesc.Columns[:i:i],
				tableDesc.Columns[i+1:]...)
			break
		} else {
			__antithesis_instrumentation__.Notify(468971)
		}
	}
}

func (n *dropIndexNode) maybeQueueDropShardColumn(
	tableDesc *tabledesc.Mutable, shardColName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(468972)
	shardColDesc, err := tableDesc.FindColumnWithName(tree.Name(shardColName))
	if err != nil {
		__antithesis_instrumentation__.Notify(468977)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(468978)
	}
	__antithesis_instrumentation__.Notify(468973)
	if shardColDesc.Dropped() {
		__antithesis_instrumentation__.Notify(468979)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(468980)
	}
	__antithesis_instrumentation__.Notify(468974)
	if catalog.FindNonDropIndex(tableDesc, func(otherIdx catalog.Index) bool {
		__antithesis_instrumentation__.Notify(468981)
		colIDs := otherIdx.CollectKeyColumnIDs()
		if !otherIdx.Primary() {
			__antithesis_instrumentation__.Notify(468983)
			colIDs.UnionWith(otherIdx.CollectSecondaryStoredColumnIDs())
			colIDs.UnionWith(otherIdx.CollectKeySuffixColumnIDs())
		} else {
			__antithesis_instrumentation__.Notify(468984)
		}
		__antithesis_instrumentation__.Notify(468982)
		return colIDs.Contains(shardColDesc.GetID())
	}) != nil {
		__antithesis_instrumentation__.Notify(468985)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(468986)
	}
	__antithesis_instrumentation__.Notify(468975)
	if err := n.dropShardColumnAndConstraint(tableDesc, shardColDesc); err != nil {
		__antithesis_instrumentation__.Notify(468987)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(468988)
	}
	__antithesis_instrumentation__.Notify(468976)
	return true, nil
}

func (n *dropIndexNode) dropShardColumnAndConstraint(
	tableDesc *tabledesc.Mutable, shardCol catalog.Column,
) error {
	__antithesis_instrumentation__.Notify(468989)
	validChecks := tableDesc.Checks[:0]
	for _, check := range tableDesc.AllActiveAndInactiveChecks() {
		__antithesis_instrumentation__.Notify(468992)
		if used, err := tableDesc.CheckConstraintUsesColumn(check, shardCol.GetID()); err != nil {
			__antithesis_instrumentation__.Notify(468993)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468994)
			if used {
				__antithesis_instrumentation__.Notify(468995)
				if check.Validity == descpb.ConstraintValidity_Validating {
					__antithesis_instrumentation__.Notify(468996)
					return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"referencing constraint %q in the middle of being added, try again later", check.Name)
				} else {
					__antithesis_instrumentation__.Notify(468997)
				}
			} else {
				__antithesis_instrumentation__.Notify(468998)
				validChecks = append(validChecks, check)
			}
		}
	}
	__antithesis_instrumentation__.Notify(468990)

	if len(validChecks) != len(tableDesc.Checks) {
		__antithesis_instrumentation__.Notify(468999)
		tableDesc.Checks = validChecks
	} else {
		__antithesis_instrumentation__.Notify(469000)
	}
	__antithesis_instrumentation__.Notify(468991)

	n.queueDropColumn(tableDesc, shardCol)
	return nil
}

func (n *dropIndexNode) finalizeDropColumn(params runParams, tableDesc *tabledesc.Mutable) error {
	__antithesis_instrumentation__.Notify(469001)
	version := params.ExecCfg().Settings.Version.ActiveVersion(params.ctx)
	if err := tableDesc.AllocateIDs(params.ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(469004)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469005)
	}
	__antithesis_instrumentation__.Notify(469002)
	mutationID := tableDesc.ClusterVersion().NextMutationID
	if err := params.p.writeSchemaChange(
		params.ctx, tableDesc, mutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(469006)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469007)
	}
	__antithesis_instrumentation__.Notify(469003)
	return nil
}

func (*dropIndexNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(469008)
	return false, nil
}
func (*dropIndexNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(469009)
	return tree.Datums{}
}
func (*dropIndexNode) Close(context.Context) { __antithesis_instrumentation__.Notify(469010) }

type fullIndexName struct {
	tn      *tree.TableName
	idxName tree.UnrestrictedName
}

type dropIndexConstraintBehavior bool

const (
	checkIdxConstraint  dropIndexConstraintBehavior = true
	ignoreIdxConstraint dropIndexConstraintBehavior = false
)

func (p *planner) dropIndexByName(
	ctx context.Context,
	tn *tree.TableName,
	idxName tree.UnrestrictedName,
	tableDesc *tabledesc.Mutable,
	ifExists bool,
	behavior tree.DropBehavior,
	constraintBehavior dropIndexConstraintBehavior,
	jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(469011)
	idx, err := tableDesc.FindIndexWithName(string(idxName))
	if err != nil {
		__antithesis_instrumentation__.Notify(469032)

		if ifExists {
			__antithesis_instrumentation__.Notify(469034)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(469035)
		}
		__antithesis_instrumentation__.Notify(469033)

		return pgerror.WithCandidateCode(err, pgcode.UndefinedObject)
	} else {
		__antithesis_instrumentation__.Notify(469036)
	}
	__antithesis_instrumentation__.Notify(469012)
	if idx.Dropped() {
		__antithesis_instrumentation__.Notify(469037)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(469038)
	}
	__antithesis_instrumentation__.Notify(469013)

	if tableDesc.IsLocalityRegionalByRow() {
		__antithesis_instrumentation__.Notify(469039)
		if err := p.checkNoRegionChangeUnderway(
			ctx,
			tableDesc.GetParentID(),
			"DROP INDEX on a REGIONAL BY ROW table",
		); err != nil {
			__antithesis_instrumentation__.Notify(469040)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469041)
		}
	} else {
		__antithesis_instrumentation__.Notify(469042)
	}
	__antithesis_instrumentation__.Notify(469014)

	if idx.IsUnique() && func() bool {
		__antithesis_instrumentation__.Notify(469043)
		return behavior != tree.DropCascade == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(469044)
		return constraintBehavior != ignoreIdxConstraint == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(469045)
		return !idx.IsCreatedExplicitly() == true
	}() == true {
		__antithesis_instrumentation__.Notify(469046)
		return errors.WithHint(
			pgerror.Newf(pgcode.DependentObjectsStillExist,
				"index %q is in use as unique constraint", idx.GetName()),
			"use CASCADE if you really want to drop it.",
		)
	} else {
		__antithesis_instrumentation__.Notify(469047)
	}
	__antithesis_instrumentation__.Notify(469015)

	_, zone, _, err := GetZoneConfigInTxn(
		ctx, p.txn, p.ExecCfg().Codec, tableDesc.ID, nil, "", false,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(469048)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469049)
	}
	__antithesis_instrumentation__.Notify(469016)

	for _, s := range zone.Subzones {
		__antithesis_instrumentation__.Notify(469050)
		if s.IndexID != uint32(idx.GetID()) {
			__antithesis_instrumentation__.Notify(469051)
			_, err = GenerateSubzoneSpans(
				p.ExecCfg().Settings,
				p.ExecCfg().LogicalClusterID(),
				p.ExecCfg().Codec,
				tableDesc,
				zone.Subzones,
				false,
			)
			if sqlerrors.IsCCLRequiredError(err) {
				__antithesis_instrumentation__.Notify(469053)
				return sqlerrors.NewCCLRequiredError(fmt.Errorf("schema change requires a CCL binary "+
					"because table %q has at least one remaining index or partition with a zone config",
					tableDesc.Name))
			} else {
				__antithesis_instrumentation__.Notify(469054)
			}
			__antithesis_instrumentation__.Notify(469052)
			break
		} else {
			__antithesis_instrumentation__.Notify(469055)
		}
	}
	__antithesis_instrumentation__.Notify(469017)

	remainingIndexes := make([]catalog.Index, 1, len(tableDesc.ActiveIndexes()))
	remainingIndexes[0] = tableDesc.GetPrimaryIndex()
	for _, index := range tableDesc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(469056)
		if index.GetID() != idx.GetID() {
			__antithesis_instrumentation__.Notify(469057)
			remainingIndexes = append(remainingIndexes, index)
		} else {
			__antithesis_instrumentation__.Notify(469058)
		}
	}
	__antithesis_instrumentation__.Notify(469018)

	indexHasReplacementCandidate := func(isValidIndex func(index catalog.Index) bool) bool {
		__antithesis_instrumentation__.Notify(469059)
		foundReplacement := false
		for _, index := range remainingIndexes {
			__antithesis_instrumentation__.Notify(469061)
			if isValidIndex(index) {
				__antithesis_instrumentation__.Notify(469062)
				foundReplacement = true
				break
			} else {
				__antithesis_instrumentation__.Notify(469063)
			}
		}
		__antithesis_instrumentation__.Notify(469060)
		return foundReplacement
	}
	__antithesis_instrumentation__.Notify(469019)

	for _, m := range tableDesc.Mutations {
		__antithesis_instrumentation__.Notify(469064)
		if c := m.GetConstraint(); c != nil && func() bool {
			__antithesis_instrumentation__.Notify(469065)
			return c.ConstraintType == descpb.ConstraintToUpdate_FOREIGN_KEY == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(469066)
			return idx.IsValidOriginIndex(c.ForeignKey.OriginColumnIDs) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(469067)
			return !indexHasReplacementCandidate(func(idx catalog.Index) bool {
				__antithesis_instrumentation__.Notify(469068)
				return idx.IsValidOriginIndex(c.ForeignKey.OriginColumnIDs)
			}) == true
		}() == true {
			__antithesis_instrumentation__.Notify(469069)
			return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"referencing constraint %q in the middle of being added, try again later", c.ForeignKey.Name)
		} else {
			__antithesis_instrumentation__.Notify(469070)
		}
	}
	__antithesis_instrumentation__.Notify(469020)

	candidateConstraints := make([]descpb.UniqueConstraint, len(remainingIndexes))
	for i := range remainingIndexes {
		__antithesis_instrumentation__.Notify(469071)

		candidateConstraints[i] = remainingIndexes[i]
	}
	__antithesis_instrumentation__.Notify(469021)
	if err := p.tryRemoveFKBackReferences(
		ctx, tableDesc, idx, behavior, candidateConstraints,
	); err != nil {
		__antithesis_instrumentation__.Notify(469072)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469073)
	}
	__antithesis_instrumentation__.Notify(469022)

	var droppedViews []string
	for _, tableRef := range tableDesc.DependedOnBy {
		__antithesis_instrumentation__.Notify(469074)
		if tableRef.IndexID == idx.GetID() {
			__antithesis_instrumentation__.Notify(469075)

			err := p.canRemoveDependentViewGeneric(
				ctx, "index", idx.GetName(), tableDesc.ParentID, tableRef, behavior)
			if err != nil {
				__antithesis_instrumentation__.Notify(469080)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469081)
			}
			__antithesis_instrumentation__.Notify(469076)
			viewDesc, err := p.getViewDescForCascade(
				ctx, "index", idx.GetName(), tableDesc.ParentID, tableRef.ID, behavior,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(469082)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469083)
			}
			__antithesis_instrumentation__.Notify(469077)
			viewJobDesc := fmt.Sprintf("removing view %q dependent on index %q which is being dropped",
				viewDesc.Name, idx.GetName())
			cascadedViews, err := p.removeDependentView(ctx, tableDesc, viewDesc, viewJobDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(469084)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469085)
			}
			__antithesis_instrumentation__.Notify(469078)

			qualifiedView, err := p.getQualifiedTableName(ctx, viewDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(469086)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469087)
			}
			__antithesis_instrumentation__.Notify(469079)

			droppedViews = append(droppedViews, qualifiedView.FQString())
			droppedViews = append(droppedViews, cascadedViews...)
		} else {
			__antithesis_instrumentation__.Notify(469088)
		}
	}
	__antithesis_instrumentation__.Notify(469023)

	idxCopy := *idx.IndexDesc()
	idxDesc := &idxCopy

	if idxDesc.ID == tableDesc.GetPrimaryIndexID() {
		__antithesis_instrumentation__.Notify(469089)
		return errors.WithHint(
			pgerror.Newf(pgcode.FeatureNotSupported, "cannot drop the primary index of a table using DROP INDEX"),
			"instead, use ALTER TABLE ... ALTER PRIMARY KEY or"+
				"use DROP CONSTRAINT ... PRIMARY KEY followed by ADD CONSTRAINT ... PRIMARY KEY in a transaction",
		)
	} else {
		__antithesis_instrumentation__.Notify(469090)
	}
	__antithesis_instrumentation__.Notify(469024)

	foundIndex := catalog.FindPublicNonPrimaryIndex(tableDesc, func(idxEntry catalog.Index) bool {
		__antithesis_instrumentation__.Notify(469091)
		return idxEntry.GetID() == idxDesc.ID
	})
	__antithesis_instrumentation__.Notify(469025)

	if foundIndex == nil {
		__antithesis_instrumentation__.Notify(469092)
		return pgerror.Newf(
			pgcode.ObjectNotInPrerequisiteState,
			"index %q in the middle of being added, try again later",
			idxName,
		)
	} else {
		__antithesis_instrumentation__.Notify(469093)
	}
	__antithesis_instrumentation__.Notify(469026)

	idxEntry := *foundIndex.IndexDesc()
	idxOrdinal := foundIndex.Ordinal()

	st := p.EvalContext().Settings
	if p.ExecCfg().Codec.ForSystemTenant() && func() bool {
		__antithesis_instrumentation__.Notify(469094)
		return !st.Version.IsActive(ctx, clusterversion.UnsplitRangesInAsyncGCJobs) == true
	}() == true {
		__antithesis_instrumentation__.Notify(469095)

		span := tableDesc.IndexSpan(p.ExecCfg().Codec, idxEntry.ID)
		txn := p.ExecCfg().DB.NewTxn(ctx, "scan-ranges-for-index-drop")
		ranges, err := kvclient.ScanMetaKVs(ctx, txn, span)
		if err != nil {
			__antithesis_instrumentation__.Notify(469097)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469098)
		}
		__antithesis_instrumentation__.Notify(469096)
		for _, r := range ranges {
			__antithesis_instrumentation__.Notify(469099)
			var desc roachpb.RangeDescriptor
			if err := r.ValueProto(&desc); err != nil {
				__antithesis_instrumentation__.Notify(469101)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469102)
			}
			__antithesis_instrumentation__.Notify(469100)

			if !desc.GetStickyBit().IsEmpty() && func() bool {
				__antithesis_instrumentation__.Notify(469103)
				return span.Key.Compare(desc.StartKey.AsRawKey()) <= 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(469104)

				if err := p.ExecCfg().DB.AdminUnsplit(ctx, desc.StartKey); err != nil && func() bool {
					__antithesis_instrumentation__.Notify(469105)
					return !strings.Contains(err.Error(), "is not the start of a range") == true
				}() == true {
					__antithesis_instrumentation__.Notify(469106)
					return err
				} else {
					__antithesis_instrumentation__.Notify(469107)
				}
			} else {
				__antithesis_instrumentation__.Notify(469108)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(469109)
	}
	__antithesis_instrumentation__.Notify(469027)

	if err := tableDesc.AddDropIndexMutation(&idxEntry); err != nil {
		__antithesis_instrumentation__.Notify(469110)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469111)
	}
	__antithesis_instrumentation__.Notify(469028)
	tableDesc.RemovePublicNonPrimaryIndex(idxOrdinal)

	commentUpdater := p.execCfg.DescMetadaUpdaterFactory.NewMetadataUpdater(
		ctx,
		p.txn,
		p.SessionData(),
	)
	if err := commentUpdater.DeleteDescriptorComment(
		int64(tableDesc.ID), int64(idxDesc.ID), keys.IndexCommentType); err != nil {
		__antithesis_instrumentation__.Notify(469112)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469113)
	}
	__antithesis_instrumentation__.Notify(469029)

	if err := validateDescriptor(ctx, p, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(469114)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469115)
	}
	__antithesis_instrumentation__.Notify(469030)

	mutationID := tableDesc.ClusterVersion().NextMutationID
	if err := p.writeSchemaChange(ctx, tableDesc, mutationID, jobDesc); err != nil {
		__antithesis_instrumentation__.Notify(469116)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469117)
	}
	__antithesis_instrumentation__.Notify(469031)
	p.BufferClientNotice(
		ctx,
		errors.WithHint(
			pgnotice.Newf("the data for dropped indexes is reclaimed asynchronously"),
			"The reclamation delay can be customized in the zone configuration for the table.",
		),
	)

	return p.logEvent(ctx,
		tableDesc.ID,
		&eventpb.DropIndex{
			TableName:           tn.FQString(),
			IndexName:           string(idxName),
			MutationID:          uint32(mutationID),
			CascadeDroppedViews: droppedViews,
		})
}
