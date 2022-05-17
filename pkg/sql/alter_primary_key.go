package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

type alterPrimaryKeyLocalitySwap struct {
	localityConfigSwap descpb.PrimaryKeySwap_LocalityConfigSwap

	mutationIdxAllowedInSameTxn *int

	newColumnName *tree.Name
}

func (p *planner) AlterPrimaryKey(
	ctx context.Context,
	tableDesc *tabledesc.Mutable,
	alterPKNode tree.AlterTableAlterPrimaryKey,
	alterPrimaryKeyLocalitySwap *alterPrimaryKeyLocalitySwap,
) error {
	__antithesis_instrumentation__.Notify(243229)
	if err := paramparse.ValidateUniqueConstraintParams(
		alterPKNode.StorageParams,
		paramparse.UniqueConstraintParamContext{
			IsPrimaryKey: true,
			IsSharded:    alterPKNode.Sharded != nil,
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(243252)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243253)
	}
	__antithesis_instrumentation__.Notify(243230)

	if alterPrimaryKeyLocalitySwap != nil {
		__antithesis_instrumentation__.Notify(243254)
		if err := p.checkNoRegionChangeUnderway(
			ctx,
			tableDesc.GetParentID(),
			"perform this locality change",
		); err != nil {
			__antithesis_instrumentation__.Notify(243255)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243256)
		}
	} else {
		__antithesis_instrumentation__.Notify(243257)
		if tableDesc.IsLocalityRegionalByRow() {
			__antithesis_instrumentation__.Notify(243258)
			if err := p.checkNoRegionChangeUnderway(
				ctx,
				tableDesc.GetParentID(),
				"perform a primary key change on a REGIONAL BY ROW table",
			); err != nil {
				__antithesis_instrumentation__.Notify(243259)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243260)
			}
		} else {
			__antithesis_instrumentation__.Notify(243261)
		}
	}
	__antithesis_instrumentation__.Notify(243231)

	currentMutationID := tableDesc.ClusterVersion().NextMutationID
	for i := range tableDesc.Mutations {
		__antithesis_instrumentation__.Notify(243262)
		mut := &tableDesc.Mutations[i]
		if mut.MutationID == currentMutationID {
			__antithesis_instrumentation__.Notify(243264)
			errBase := "primary key"
			if alterPrimaryKeyLocalitySwap != nil {
				__antithesis_instrumentation__.Notify(243266)
				allowIdx := alterPrimaryKeyLocalitySwap.mutationIdxAllowedInSameTxn
				if allowIdx != nil && func() bool {
					__antithesis_instrumentation__.Notify(243268)
					return *allowIdx == i == true
				}() == true {
					__antithesis_instrumentation__.Notify(243269)
					continue
				} else {
					__antithesis_instrumentation__.Notify(243270)
				}
				__antithesis_instrumentation__.Notify(243267)
				errBase = "locality"
			} else {
				__antithesis_instrumentation__.Notify(243271)
			}
			__antithesis_instrumentation__.Notify(243265)
			return unimplemented.NewWithIssuef(
				45510,
				"cannot perform a %[1]s change on %[2]s with other schema changes on %[2]s in the same transaction",
				errBase,
				tableDesc.Name,
			)
		} else {
			__antithesis_instrumentation__.Notify(243272)
		}
		__antithesis_instrumentation__.Notify(243263)
		if mut.MutationID < currentMutationID {
			__antithesis_instrumentation__.Notify(243273)

			if mut.GetIndex() != nil && func() bool {
				__antithesis_instrumentation__.Notify(243275)
				return mut.Direction == descpb.DescriptorMutation_DROP == true
			}() == true {
				__antithesis_instrumentation__.Notify(243276)
				continue
			} else {
				__antithesis_instrumentation__.Notify(243277)
			}
			__antithesis_instrumentation__.Notify(243274)
			return unimplemented.NewWithIssuef(
				45510, "table %s is currently undergoing a schema change", tableDesc.Name)
		} else {
			__antithesis_instrumentation__.Notify(243278)
		}
	}
	__antithesis_instrumentation__.Notify(243232)

	for _, elem := range alterPKNode.Columns {
		__antithesis_instrumentation__.Notify(243279)
		col, err := tableDesc.FindColumnWithName(elem.Column)
		if err != nil {
			__antithesis_instrumentation__.Notify(243284)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243285)
		}
		__antithesis_instrumentation__.Notify(243280)
		if col.Dropped() {
			__antithesis_instrumentation__.Notify(243286)
			return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"column %q is being dropped", col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(243287)
		}
		__antithesis_instrumentation__.Notify(243281)
		if col.IsInaccessible() {
			__antithesis_instrumentation__.Notify(243288)
			return pgerror.Newf(pgcode.InvalidSchemaDefinition, "cannot use inaccessible column %q in primary key", col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(243289)
		}
		__antithesis_instrumentation__.Notify(243282)
		if col.IsNullable() {
			__antithesis_instrumentation__.Notify(243290)
			return pgerror.Newf(pgcode.InvalidSchemaDefinition, "cannot use nullable column %q in primary key", col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(243291)
		}
		__antithesis_instrumentation__.Notify(243283)
		if !p.EvalContext().Settings.Version.IsActive(ctx, clusterversion.Start22_1) {
			__antithesis_instrumentation__.Notify(243292)
			if col.IsVirtual() {
				__antithesis_instrumentation__.Notify(243293)
				return pgerror.Newf(pgcode.FeatureNotSupported, "cannot use virtual column %q in primary key", col.GetName())
			} else {
				__antithesis_instrumentation__.Notify(243294)
			}
		} else {
			__antithesis_instrumentation__.Notify(243295)
		}
	}

	{
		__antithesis_instrumentation__.Notify(243296)
		requiresIndexChange, err := p.shouldCreateIndexes(ctx, tableDesc, &alterPKNode, alterPrimaryKeyLocalitySwap)
		if err != nil {
			__antithesis_instrumentation__.Notify(243298)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243299)
		}
		__antithesis_instrumentation__.Notify(243297)
		if !requiresIndexChange {
			__antithesis_instrumentation__.Notify(243300)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(243301)
		}
	}
	__antithesis_instrumentation__.Notify(243233)

	nameExists := func(name string) bool {
		__antithesis_instrumentation__.Notify(243302)
		_, err := tableDesc.FindIndexWithName(name)
		return err == nil
	}
	__antithesis_instrumentation__.Notify(243234)

	name := tabledesc.GenerateUniqueName(
		"new_primary_key",
		nameExists,
	)
	if alterPKNode.Name != "" && func() bool {
		__antithesis_instrumentation__.Notify(243303)
		return tableDesc.PrimaryIndex.Name != string(alterPKNode.Name) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(243304)
		return nameExists(string(alterPKNode.Name)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(243305)
		return pgerror.Newf(pgcode.DuplicateRelation, "index with name %s already exists", alterPKNode.Name)
	} else {
		__antithesis_instrumentation__.Notify(243306)
	}
	__antithesis_instrumentation__.Notify(243235)
	newPrimaryIndexDesc := &descpb.IndexDescriptor{
		Name:              name,
		Unique:            true,
		CreatedExplicitly: true,
		EncodingType:      descpb.PrimaryIndexEncoding,
		Type:              descpb.IndexDescriptor_FORWARD,

		Version:        descpb.StrictIndexColumnIDGuaranteesVersion,
		ConstraintID:   tableDesc.GetNextConstraintID(),
		CreatedAtNanos: p.EvalContext().GetTxnTimestamp(time.Microsecond).UnixNano(),
	}
	tableDesc.NextConstraintID++

	if alterPKNode.Sharded != nil {
		__antithesis_instrumentation__.Notify(243307)
		shardCol, newColumns, err := setupShardedIndex(
			ctx,
			p.EvalContext(),
			&p.semaCtx,
			alterPKNode.Columns,
			alterPKNode.Sharded.ShardBuckets,
			tableDesc,
			newPrimaryIndexDesc,
			alterPKNode.StorageParams,
			false,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(243310)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243311)
		}
		__antithesis_instrumentation__.Notify(243308)
		alterPKNode.Columns = newColumns
		if err := p.maybeSetupConstraintForShard(
			ctx,
			tableDesc,
			shardCol,
			newPrimaryIndexDesc.Sharded.ShardBuckets,
		); err != nil {
			__antithesis_instrumentation__.Notify(243312)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243313)
		}
		__antithesis_instrumentation__.Notify(243309)
		telemetry.Inc(sqltelemetry.HashShardedIndexCounter)
	} else {
		__antithesis_instrumentation__.Notify(243314)
	}
	__antithesis_instrumentation__.Notify(243236)

	if err := newPrimaryIndexDesc.FillColumns(alterPKNode.Columns); err != nil {
		__antithesis_instrumentation__.Notify(243315)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243316)
	}

	{
		__antithesis_instrumentation__.Notify(243317)

		names := make(map[string]struct{}, len(newPrimaryIndexDesc.KeyColumnNames))
		for _, name := range newPrimaryIndexDesc.KeyColumnNames {
			__antithesis_instrumentation__.Notify(243320)
			names[name] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(243318)
		deletable := tableDesc.DeletableColumns()
		newPrimaryIndexDesc.StoreColumnNames = make([]string, 0, len(deletable))
		for _, col := range deletable {
			__antithesis_instrumentation__.Notify(243321)
			if _, found := names[col.GetName()]; found || func() bool {
				__antithesis_instrumentation__.Notify(243323)
				return col.IsVirtual() == true
			}() == true {
				__antithesis_instrumentation__.Notify(243324)
				continue
			} else {
				__antithesis_instrumentation__.Notify(243325)
			}
			__antithesis_instrumentation__.Notify(243322)
			newPrimaryIndexDesc.StoreColumnNames = append(newPrimaryIndexDesc.StoreColumnNames, col.GetName())
		}
		__antithesis_instrumentation__.Notify(243319)
		if len(newPrimaryIndexDesc.StoreColumnNames) == 0 {
			__antithesis_instrumentation__.Notify(243326)
			newPrimaryIndexDesc.StoreColumnNames = nil
		} else {
			__antithesis_instrumentation__.Notify(243327)
		}
	}
	__antithesis_instrumentation__.Notify(243237)

	if err := tableDesc.AddIndexMutation(ctx, newPrimaryIndexDesc, descpb.DescriptorMutation_ADD, p.ExecCfg().Settings); err != nil {
		__antithesis_instrumentation__.Notify(243328)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243329)
	}
	__antithesis_instrumentation__.Notify(243238)
	version := p.ExecCfg().Settings.Version.ActiveVersion(ctx)
	if err := tableDesc.AllocateIDs(ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(243330)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243331)
	}
	__antithesis_instrumentation__.Notify(243239)

	var allowedNewColumnNames []tree.Name
	var err error

	isNewPartitionAllBy := false
	var partitionAllBy *tree.PartitionBy

	dropPartitionAllBy := false

	allowImplicitPartitioning := false

	if alterPrimaryKeyLocalitySwap != nil {
		__antithesis_instrumentation__.Notify(243332)
		localityConfigSwap := alterPrimaryKeyLocalitySwap.localityConfigSwap
		switch to := localityConfigSwap.NewLocalityConfig.Locality.(type) {
		case *catpb.LocalityConfig_RegionalByRow_:
			__antithesis_instrumentation__.Notify(243333)

			switch localityConfigSwap.OldLocalityConfig.Locality.(type) {
			case *catpb.LocalityConfig_RegionalByRow_:
				__antithesis_instrumentation__.Notify(243340)

				dropPartitionAllBy = true
			case *catpb.LocalityConfig_Global_,
				*catpb.LocalityConfig_RegionalByTable_:
				__antithesis_instrumentation__.Notify(243341)
			default:
				__antithesis_instrumentation__.Notify(243342)
				return errors.AssertionFailedf(
					"unknown locality config swap: %T to %T",
					localityConfigSwap.OldLocalityConfig.Locality,
					localityConfigSwap.NewLocalityConfig.Locality,
				)
			}
			__antithesis_instrumentation__.Notify(243334)

			isNewPartitionAllBy = true
			allowImplicitPartitioning = true
			colName := tree.RegionalByRowRegionDefaultColName
			if as := to.RegionalByRow.As; as != nil {
				__antithesis_instrumentation__.Notify(243343)
				colName = tree.Name(*as)
			} else {
				__antithesis_instrumentation__.Notify(243344)
			}
			__antithesis_instrumentation__.Notify(243335)
			regionConfig, err := SynthesizeRegionConfig(
				ctx, p.txn, tableDesc.GetParentID(), p.Descriptors(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(243345)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243346)
			}
			__antithesis_instrumentation__.Notify(243336)
			partitionAllBy = partitionByForRegionalByRow(
				regionConfig,
				colName,
			)
			if alterPrimaryKeyLocalitySwap.newColumnName != nil {
				__antithesis_instrumentation__.Notify(243347)
				allowedNewColumnNames = append(
					allowedNewColumnNames,
					*alterPrimaryKeyLocalitySwap.newColumnName,
				)
			} else {
				__antithesis_instrumentation__.Notify(243348)
			}
		case *catpb.LocalityConfig_Global_,
			*catpb.LocalityConfig_RegionalByTable_:
			__antithesis_instrumentation__.Notify(243337)

			if localityConfigSwap.OldLocalityConfig.GetRegionalByRow() == nil {
				__antithesis_instrumentation__.Notify(243349)
				return errors.AssertionFailedf(
					"unknown locality config swap: %T to %T",
					localityConfigSwap.OldLocalityConfig.Locality,
					localityConfigSwap.NewLocalityConfig.Locality,
				)
			} else {
				__antithesis_instrumentation__.Notify(243350)
			}
			__antithesis_instrumentation__.Notify(243338)

			dropPartitionAllBy = true
		default:
			__antithesis_instrumentation__.Notify(243339)
			return errors.AssertionFailedf(
				"unknown locality config swap: %T to %T",
				localityConfigSwap.OldLocalityConfig.Locality,
				localityConfigSwap.NewLocalityConfig.Locality,
			)
		}
	} else {
		__antithesis_instrumentation__.Notify(243351)
		if tableDesc.IsPartitionAllBy() {
			__antithesis_instrumentation__.Notify(243352)
			allowImplicitPartitioning = true
			partitionAllBy, err = partitionByFromTableDesc(p.ExecCfg().Codec, tableDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(243353)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243354)
			}
		} else {
			__antithesis_instrumentation__.Notify(243355)
		}
	}
	__antithesis_instrumentation__.Notify(243240)

	if partitionAllBy != nil {
		__antithesis_instrumentation__.Notify(243356)
		newImplicitCols, newPartitioning, err := CreatePartitioning(
			ctx,
			p.ExecCfg().Settings,
			p.EvalContext(),
			tableDesc,
			*newPrimaryIndexDesc,
			partitionAllBy,
			allowedNewColumnNames,
			allowImplicitPartitioning,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(243358)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243359)
		}
		__antithesis_instrumentation__.Notify(243357)
		tabledesc.UpdateIndexPartitioning(newPrimaryIndexDesc, true, newImplicitCols, newPartitioning)
	} else {
		__antithesis_instrumentation__.Notify(243360)
	}
	__antithesis_instrumentation__.Notify(243241)

	if shouldCopyPrimaryKey(tableDesc, newPrimaryIndexDesc, alterPrimaryKeyLocalitySwap) {
		__antithesis_instrumentation__.Notify(243361)
		newUniqueIdx := tableDesc.GetPrimaryIndex().IndexDescDeepCopy()

		newUniqueIdx.ID = 0
		newUniqueIdx.Name = ""
		newUniqueIdx.StoreColumnIDs = nil
		newUniqueIdx.StoreColumnNames = nil
		newUniqueIdx.KeySuffixColumnIDs = nil
		newUniqueIdx.CompositeColumnIDs = nil
		newUniqueIdx.KeyColumnIDs = nil

		newUniqueIdx.Version = descpb.StrictIndexColumnIDGuaranteesVersion
		newUniqueIdx.EncodingType = descpb.SecondaryIndexEncoding
		if err := addIndexMutationWithSpecificPrimaryKey(ctx, tableDesc, &newUniqueIdx, newPrimaryIndexDesc, p.ExecCfg().Settings); err != nil {
			__antithesis_instrumentation__.Notify(243363)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243364)
		}
		__antithesis_instrumentation__.Notify(243362)

		if tableDesc.IsLocalityRegionalByRow() {
			__antithesis_instrumentation__.Notify(243365)
			if err := p.configureZoneConfigForNewIndexPartitioning(
				ctx,
				tableDesc,
				newUniqueIdx,
			); err != nil {
				__antithesis_instrumentation__.Notify(243366)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243367)
			}
		} else {
			__antithesis_instrumentation__.Notify(243368)
		}
	} else {
		__antithesis_instrumentation__.Notify(243369)
	}
	__antithesis_instrumentation__.Notify(243242)

	shouldRewriteIndex := func(idx catalog.Index) (bool, error) {
		__antithesis_instrumentation__.Notify(243370)
		if alterPrimaryKeyLocalitySwap != nil {
			__antithesis_instrumentation__.Notify(243376)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(243377)
		}
		__antithesis_instrumentation__.Notify(243371)
		colIDs := idx.CollectKeyColumnIDs()
		if !idx.Primary() {
			__antithesis_instrumentation__.Notify(243378)
			colIDs.UnionWith(idx.CollectSecondaryStoredColumnIDs())
			colIDs.UnionWith(idx.CollectKeySuffixColumnIDs())
		} else {
			__antithesis_instrumentation__.Notify(243379)
		}
		__antithesis_instrumentation__.Notify(243372)
		for _, colID := range newPrimaryIndexDesc.KeyColumnIDs {
			__antithesis_instrumentation__.Notify(243380)
			if !colIDs.Contains(colID) {
				__antithesis_instrumentation__.Notify(243381)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(243382)
			}
		}
		__antithesis_instrumentation__.Notify(243373)
		if idx.IsUnique() {
			__antithesis_instrumentation__.Notify(243383)
			for i := 0; i < idx.NumKeyColumns(); i++ {
				__antithesis_instrumentation__.Notify(243384)
				colID := idx.GetKeyColumnID(i)
				col, err := tableDesc.FindColumnWithID(colID)
				if err != nil {
					__antithesis_instrumentation__.Notify(243386)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(243387)
				}
				__antithesis_instrumentation__.Notify(243385)
				if col.IsNullable() {
					__antithesis_instrumentation__.Notify(243388)
					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(243389)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(243390)
		}
		__antithesis_instrumentation__.Notify(243374)

		if !idx.Primary() {
			__antithesis_instrumentation__.Notify(243391)
			newPrimaryKeyColIDs := catalog.MakeTableColSet(newPrimaryIndexDesc.KeyColumnIDs...)
			for i := 0; i < idx.NumKeySuffixColumns(); i++ {
				__antithesis_instrumentation__.Notify(243392)
				colID := idx.GetKeySuffixColumnID(i)
				col, err := tableDesc.FindColumnWithID(colID)
				if err != nil {
					__antithesis_instrumentation__.Notify(243394)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(243395)
				}
				__antithesis_instrumentation__.Notify(243393)
				if col.IsVirtual() && func() bool {
					__antithesis_instrumentation__.Notify(243396)
					return !newPrimaryKeyColIDs.Contains(colID) == true
				}() == true {
					__antithesis_instrumentation__.Notify(243397)
					return true, err
				} else {
					__antithesis_instrumentation__.Notify(243398)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(243399)
		}
		__antithesis_instrumentation__.Notify(243375)

		return !idx.IsUnique() || func() bool {
			__antithesis_instrumentation__.Notify(243400)
			return idx.GetType() == descpb.IndexDescriptor_INVERTED == true
		}() == true, nil
	}
	__antithesis_instrumentation__.Notify(243243)
	var indexesToRewrite []catalog.Index
	for _, idx := range tableDesc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(243401)
		shouldRewrite, err := shouldRewriteIndex(idx)
		if err != nil {
			__antithesis_instrumentation__.Notify(243403)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243404)
		}
		__antithesis_instrumentation__.Notify(243402)
		if idx.GetID() != newPrimaryIndexDesc.ID && func() bool {
			__antithesis_instrumentation__.Notify(243405)
			return shouldRewrite == true
		}() == true {
			__antithesis_instrumentation__.Notify(243406)
			indexesToRewrite = append(indexesToRewrite, idx)
		} else {
			__antithesis_instrumentation__.Notify(243407)
		}
	}
	__antithesis_instrumentation__.Notify(243244)

	for _, mut := range tableDesc.AllMutations() {
		__antithesis_instrumentation__.Notify(243408)

		if idx := mut.AsIndex(); mut.MutationID() < currentMutationID && func() bool {
			__antithesis_instrumentation__.Notify(243409)
			return idx != nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(243410)
			return mut.Adding() == true
		}() == true {
			__antithesis_instrumentation__.Notify(243411)
			shouldRewrite, err := shouldRewriteIndex(idx)
			if err != nil {
				__antithesis_instrumentation__.Notify(243413)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243414)
			}
			__antithesis_instrumentation__.Notify(243412)
			if shouldRewrite {
				__antithesis_instrumentation__.Notify(243415)
				indexesToRewrite = append(indexesToRewrite, idx)
			} else {
				__antithesis_instrumentation__.Notify(243416)
			}
		} else {
			__antithesis_instrumentation__.Notify(243417)
		}
	}
	__antithesis_instrumentation__.Notify(243245)

	var oldIndexIDs, newIndexIDs []descpb.IndexID
	for _, idx := range indexesToRewrite {
		__antithesis_instrumentation__.Notify(243418)

		newIndex := idx.IndexDescDeepCopy()
		basename := newIndex.Name + "_rewrite_for_primary_key_change"

		if dropPartitionAllBy {
			__antithesis_instrumentation__.Notify(243422)
			tabledesc.UpdateIndexPartitioning(&newIndex, idx.Primary(), nil, catpb.PartitioningDescriptor{})
		} else {
			__antithesis_instrumentation__.Notify(243423)
		}
		__antithesis_instrumentation__.Notify(243419)

		if isNewPartitionAllBy {
			__antithesis_instrumentation__.Notify(243424)
			newImplicitCols, newPartitioning, err := CreatePartitioning(
				ctx,
				p.ExecCfg().Settings,
				p.EvalContext(),
				tableDesc,
				newIndex,
				partitionAllBy,
				allowedNewColumnNames,
				allowImplicitPartitioning,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(243426)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243427)
			}
			__antithesis_instrumentation__.Notify(243425)
			tabledesc.UpdateIndexPartitioning(&newIndex, idx.Primary(), newImplicitCols, newPartitioning)
		} else {
			__antithesis_instrumentation__.Notify(243428)
		}
		__antithesis_instrumentation__.Notify(243420)

		newIndex.Name = tabledesc.GenerateUniqueName(basename, nameExists)

		newIndex.Version = descpb.StrictIndexColumnIDGuaranteesVersion
		newIndex.EncodingType = descpb.SecondaryIndexEncoding
		if err := addIndexMutationWithSpecificPrimaryKey(ctx, tableDesc, &newIndex, newPrimaryIndexDesc, p.ExecCfg().Settings); err != nil {
			__antithesis_instrumentation__.Notify(243429)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243430)
		}
		__antithesis_instrumentation__.Notify(243421)

		oldIndexIDs = append(oldIndexIDs, idx.GetID())
		newIndexIDs = append(newIndexIDs, newIndex.ID)
	}
	__antithesis_instrumentation__.Notify(243246)

	nonDropIndexes := tableDesc.NonDropIndexes()
	remainingIndexes := make([]descpb.UniqueConstraint, 0, len(nonDropIndexes))
	for i := range nonDropIndexes {
		__antithesis_instrumentation__.Notify(243431)

		if nonDropIndexes[i].GetID() == tableDesc.GetPrimaryIndex().GetID() {
			__antithesis_instrumentation__.Notify(243433)
			continue
		} else {
			__antithesis_instrumentation__.Notify(243434)
		}
		__antithesis_instrumentation__.Notify(243432)
		remainingIndexes = append(remainingIndexes, nonDropIndexes[i])
	}
	__antithesis_instrumentation__.Notify(243247)
	remainingIndexes = append(remainingIndexes, newPrimaryIndexDesc)
	err = p.tryRemoveFKBackReferences(
		ctx,
		tableDesc,
		tableDesc.GetPrimaryIndex(),
		tree.DropRestrict,
		remainingIndexes)
	if err != nil {
		__antithesis_instrumentation__.Notify(243435)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243436)
	}
	__antithesis_instrumentation__.Notify(243248)

	swapArgs := &descpb.PrimaryKeySwap{
		OldPrimaryIndexId:   tableDesc.GetPrimaryIndexID(),
		NewPrimaryIndexId:   newPrimaryIndexDesc.ID,
		NewIndexes:          newIndexIDs,
		OldIndexes:          oldIndexIDs,
		NewPrimaryIndexName: string(alterPKNode.Name),
	}
	if alterPrimaryKeyLocalitySwap != nil {
		__antithesis_instrumentation__.Notify(243437)
		swapArgs.LocalityConfigSwap = &alterPrimaryKeyLocalitySwap.localityConfigSwap
	} else {
		__antithesis_instrumentation__.Notify(243438)
	}
	__antithesis_instrumentation__.Notify(243249)
	tableDesc.AddPrimaryKeySwapMutation(swapArgs)

	if err := descbuilder.ValidateSelf(tableDesc, version); err != nil {
		__antithesis_instrumentation__.Notify(243439)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243440)
	}

	{
		__antithesis_instrumentation__.Notify(243441)
		primaryIndex := *tableDesc.GetPrimaryIndex().IndexDesc()
		primaryIndex.Disabled = false
		tableDesc.SetPrimaryIndex(primaryIndex)
	}
	__antithesis_instrumentation__.Notify(243250)

	noticeStr := "primary key changes are finalized asynchronously"
	if alterPrimaryKeyLocalitySwap != nil {
		__antithesis_instrumentation__.Notify(243442)
		noticeStr = "LOCALITY changes will be finalized asynchronously"
	} else {
		__antithesis_instrumentation__.Notify(243443)
	}
	__antithesis_instrumentation__.Notify(243251)
	p.BufferClientNotice(
		ctx,
		pgnotice.Newf(
			"%s; further schema changes on this table may be restricted "+
				"until the job completes",
			noticeStr,
		),
	)

	return nil
}

func (p *planner) shouldCreateIndexes(
	ctx context.Context,
	desc *tabledesc.Mutable,
	alterPKNode *tree.AlterTableAlterPrimaryKey,
	alterPrimaryKeyLocalitySwap *alterPrimaryKeyLocalitySwap,
) (requiresIndexChange bool, err error) {
	__antithesis_instrumentation__.Notify(243444)
	oldPK := desc.GetPrimaryIndex()

	if oldPK.NumKeyColumns() != len(alterPKNode.Columns) || func() bool {
		__antithesis_instrumentation__.Notify(243450)
		return oldPK.IsSharded() != (alterPKNode.Sharded != nil) == true
	}() == true {
		__antithesis_instrumentation__.Notify(243451)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(243452)
	}
	__antithesis_instrumentation__.Notify(243445)

	if alterPKNode.Sharded != nil {
		__antithesis_instrumentation__.Notify(243453)
		shardBuckets, err := tabledesc.EvalShardBucketCount(ctx, &p.semaCtx, p.EvalContext(), alterPKNode.Sharded.ShardBuckets, alterPKNode.StorageParams)
		if err != nil {
			__antithesis_instrumentation__.Notify(243455)
			return true, err
		} else {
			__antithesis_instrumentation__.Notify(243456)
		}
		__antithesis_instrumentation__.Notify(243454)
		if oldPK.GetSharded().ShardBuckets != shardBuckets {
			__antithesis_instrumentation__.Notify(243457)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(243458)
		}
	} else {
		__antithesis_instrumentation__.Notify(243459)
	}
	__antithesis_instrumentation__.Notify(243446)

	if oldPK.IsDisabled() {
		__antithesis_instrumentation__.Notify(243460)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(243461)
	}
	__antithesis_instrumentation__.Notify(243447)

	for idx, elem := range alterPKNode.Columns {
		__antithesis_instrumentation__.Notify(243462)
		col, err := desc.FindColumnWithName(elem.Column)
		if err != nil {
			__antithesis_instrumentation__.Notify(243465)
			return true, err
		} else {
			__antithesis_instrumentation__.Notify(243466)
		}
		__antithesis_instrumentation__.Notify(243463)

		if col.GetID() != oldPK.GetKeyColumnID(idx) {
			__antithesis_instrumentation__.Notify(243467)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(243468)
		}
		__antithesis_instrumentation__.Notify(243464)
		if (elem.Direction == tree.Ascending && func() bool {
			__antithesis_instrumentation__.Notify(243469)
			return oldPK.GetKeyColumnDirection(idx) != descpb.IndexDescriptor_ASC == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(243470)
			return (elem.Direction == tree.Descending && func() bool {
				__antithesis_instrumentation__.Notify(243471)
				return oldPK.GetKeyColumnDirection(idx) != descpb.IndexDescriptor_DESC == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(243472)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(243473)
		}
	}
	__antithesis_instrumentation__.Notify(243448)

	if alterPrimaryKeyLocalitySwap != nil {
		__antithesis_instrumentation__.Notify(243474)
		localitySwapConfig := alterPrimaryKeyLocalitySwap.localityConfigSwap
		if !localitySwapConfig.NewLocalityConfig.Equal(localitySwapConfig.OldLocalityConfig) {
			__antithesis_instrumentation__.Notify(243476)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(243477)
		}
		__antithesis_instrumentation__.Notify(243475)
		if localitySwapConfig.NewRegionalByRowColumnID != nil && func() bool {
			__antithesis_instrumentation__.Notify(243478)
			return *localitySwapConfig.NewRegionalByRowColumnID != oldPK.GetKeyColumnID(0) == true
		}() == true {
			__antithesis_instrumentation__.Notify(243479)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(243480)
		}
	} else {
		__antithesis_instrumentation__.Notify(243481)
	}
	__antithesis_instrumentation__.Notify(243449)
	return false, nil
}

func shouldCopyPrimaryKey(
	desc *tabledesc.Mutable,
	newPK *descpb.IndexDescriptor,
	alterPrimaryKeyLocalitySwap *alterPrimaryKeyLocalitySwap,
) bool {
	__antithesis_instrumentation__.Notify(243482)

	columnIDsAndDirsWithoutSharded := func(idx *descpb.IndexDescriptor) (
		columnIDs descpb.ColumnIDs,
		columnDirs []descpb.IndexDescriptor_Direction,
	) {
		__antithesis_instrumentation__.Notify(243485)
		for i, colName := range idx.KeyColumnNames {
			__antithesis_instrumentation__.Notify(243487)
			if colName != idx.Sharded.Name {
				__antithesis_instrumentation__.Notify(243488)
				columnIDs = append(columnIDs, idx.KeyColumnIDs[i])
				columnDirs = append(columnDirs, idx.KeyColumnDirections[i])
			} else {
				__antithesis_instrumentation__.Notify(243489)
			}
		}
		__antithesis_instrumentation__.Notify(243486)
		return columnIDs, columnDirs
	}
	__antithesis_instrumentation__.Notify(243483)
	idsAndDirsMatch := func(old, new *descpb.IndexDescriptor) bool {
		__antithesis_instrumentation__.Notify(243490)
		oldIDs, oldDirs := columnIDsAndDirsWithoutSharded(old)
		newIDs, newDirs := columnIDsAndDirsWithoutSharded(new)
		if !oldIDs.Equals(newIDs) {
			__antithesis_instrumentation__.Notify(243493)
			return false
		} else {
			__antithesis_instrumentation__.Notify(243494)
		}
		__antithesis_instrumentation__.Notify(243491)
		for i := range oldDirs {
			__antithesis_instrumentation__.Notify(243495)
			if oldDirs[i] != newDirs[i] {
				__antithesis_instrumentation__.Notify(243496)
				return false
			} else {
				__antithesis_instrumentation__.Notify(243497)
			}
		}
		__antithesis_instrumentation__.Notify(243492)
		return true
	}
	__antithesis_instrumentation__.Notify(243484)
	oldPK := desc.GetPrimaryIndex().IndexDesc()
	return alterPrimaryKeyLocalitySwap == nil && func() bool {
		__antithesis_instrumentation__.Notify(243498)
		return desc.HasPrimaryKey() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(243499)
		return !desc.IsPrimaryIndexDefaultRowID() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(243500)
		return !idsAndDirsMatch(oldPK, newPK) == true
	}() == true
}

func addIndexMutationWithSpecificPrimaryKey(
	ctx context.Context,
	table *tabledesc.Mutable,
	toAdd *descpb.IndexDescriptor,
	primary *descpb.IndexDescriptor,
	settings *cluster.Settings,
) error {
	__antithesis_instrumentation__.Notify(243501)

	toAdd.ID = 0
	if err := table.AddIndexMutation(ctx, toAdd, descpb.DescriptorMutation_ADD, settings); err != nil {
		__antithesis_instrumentation__.Notify(243505)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243506)
	}
	__antithesis_instrumentation__.Notify(243502)
	if err := table.AllocateIDsWithoutValidation(ctx); err != nil {
		__antithesis_instrumentation__.Notify(243507)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243508)
	}
	__antithesis_instrumentation__.Notify(243503)

	setKeySuffixColumnIDsFromPrimary(toAdd, primary)
	if tempIdx := catalog.FindCorrespondingTemporaryIndexByID(table, toAdd.ID); tempIdx != nil {
		__antithesis_instrumentation__.Notify(243509)
		setKeySuffixColumnIDsFromPrimary(tempIdx.IndexDesc(), primary)
	} else {
		__antithesis_instrumentation__.Notify(243510)
	}
	__antithesis_instrumentation__.Notify(243504)

	return nil
}

func setKeySuffixColumnIDsFromPrimary(
	toAdd *descpb.IndexDescriptor, primary *descpb.IndexDescriptor,
) {
	__antithesis_instrumentation__.Notify(243511)
	presentColIDs := catalog.MakeTableColSet(toAdd.KeyColumnIDs...)
	presentColIDs.UnionWith(catalog.MakeTableColSet(toAdd.StoreColumnIDs...))
	toAdd.KeySuffixColumnIDs = nil
	for _, colID := range primary.KeyColumnIDs {
		__antithesis_instrumentation__.Notify(243512)
		if !presentColIDs.Contains(colID) {
			__antithesis_instrumentation__.Notify(243513)
			toAdd.KeySuffixColumnIDs = append(toAdd.KeySuffixColumnIDs, colID)
		} else {
			__antithesis_instrumentation__.Notify(243514)
		}
	}
}
