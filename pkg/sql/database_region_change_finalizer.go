package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type databaseRegionChangeFinalizer struct {
	dbID   descpb.ID
	typeID descpb.ID

	localPlanner        *planner
	cleanupFunc         func()
	regionalByRowTables []*tabledesc.Mutable
}

func newDatabaseRegionChangeFinalizer(
	ctx context.Context,
	txn *kv.Txn,
	execCfg *ExecutorConfig,
	descsCol *descs.Collection,
	dbID descpb.ID,
	typeID descpb.ID,
) (*databaseRegionChangeFinalizer, error) {
	__antithesis_instrumentation__.Notify(465279)
	p, cleanup := NewInternalPlanner(
		"repartition-regional-by-row-tables",
		txn,
		security.RootUserName(),
		&MemoryMetrics{},
		execCfg,
		sessiondatapb.SessionData{},
		WithDescCollection(descsCol),
	)
	localPlanner := p.(*planner)

	var regionalByRowTables []*tabledesc.Mutable
	if err := func() error {
		__antithesis_instrumentation__.Notify(465281)
		_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
			ctx,
			txn,
			dbID,
			tree.DatabaseLookupFlags{
				Required:    true,
				AvoidLeased: true,
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(465283)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465284)
		}
		__antithesis_instrumentation__.Notify(465282)

		return localPlanner.forEachMutableTableInDatabase(
			ctx,
			dbDesc,
			func(ctx context.Context, scName string, tableDesc *tabledesc.Mutable) error {
				__antithesis_instrumentation__.Notify(465285)
				if !tableDesc.IsLocalityRegionalByRow() || func() bool {
					__antithesis_instrumentation__.Notify(465287)
					return tableDesc.Dropped() == true
				}() == true {
					__antithesis_instrumentation__.Notify(465288)

					return nil
				} else {
					__antithesis_instrumentation__.Notify(465289)
				}
				__antithesis_instrumentation__.Notify(465286)
				regionalByRowTables = append(regionalByRowTables, tableDesc)
				return nil
			},
		)
	}(); err != nil {
		__antithesis_instrumentation__.Notify(465290)
		cleanup()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465291)
	}
	__antithesis_instrumentation__.Notify(465280)

	return &databaseRegionChangeFinalizer{
		dbID:                dbID,
		typeID:              typeID,
		localPlanner:        localPlanner,
		cleanupFunc:         cleanup,
		regionalByRowTables: regionalByRowTables,
	}, nil
}

func (r *databaseRegionChangeFinalizer) cleanup() {
	__antithesis_instrumentation__.Notify(465292)
	if r.cleanupFunc != nil {
		__antithesis_instrumentation__.Notify(465293)
		r.cleanupFunc()
		r.cleanupFunc = nil
	} else {
		__antithesis_instrumentation__.Notify(465294)
	}
}

func (r *databaseRegionChangeFinalizer) finalize(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(465295)
	if err := r.updateDatabaseZoneConfig(ctx, txn); err != nil {
		__antithesis_instrumentation__.Notify(465298)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465299)
	}
	__antithesis_instrumentation__.Notify(465296)
	if err := r.preDrop(ctx, txn); err != nil {
		__antithesis_instrumentation__.Notify(465300)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465301)
	}
	__antithesis_instrumentation__.Notify(465297)
	return r.updateGlobalTablesZoneConfig(ctx, txn)
}

func (r *databaseRegionChangeFinalizer) preDrop(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(465302)
	repartitioned, zoneConfigUpdates, err := r.repartitionRegionalByRowTables(ctx, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(465306)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465307)
	}
	__antithesis_instrumentation__.Notify(465303)
	for _, update := range zoneConfigUpdates {
		__antithesis_instrumentation__.Notify(465308)
		if _, err := writeZoneConfigUpdate(
			ctx, txn, r.localPlanner.ExecCfg(), update,
		); err != nil {
			__antithesis_instrumentation__.Notify(465309)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465310)
		}
	}
	__antithesis_instrumentation__.Notify(465304)
	b := txn.NewBatch()
	for _, t := range repartitioned {
		__antithesis_instrumentation__.Notify(465311)
		const kvTrace = false
		if err := r.localPlanner.Descriptors().WriteDescToBatch(
			ctx, kvTrace, t, b,
		); err != nil {
			__antithesis_instrumentation__.Notify(465312)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465313)
		}
	}
	__antithesis_instrumentation__.Notify(465305)
	return txn.Run(ctx, b)
}

func (r *databaseRegionChangeFinalizer) updateGlobalTablesZoneConfig(
	ctx context.Context, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(465314)
	regionConfig, err := SynthesizeRegionConfig(ctx, txn, r.dbID, r.localPlanner.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(465319)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465320)
	}
	__antithesis_instrumentation__.Notify(465315)

	if !regionConfig.IsPlacementRestricted() {
		__antithesis_instrumentation__.Notify(465321)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(465322)
	}
	__antithesis_instrumentation__.Notify(465316)

	descsCol := r.localPlanner.Descriptors()

	_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
		ctx,
		txn,
		r.dbID,
		tree.DatabaseLookupFlags{
			Required:    true,
			AvoidLeased: true,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(465323)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465324)
	}
	__antithesis_instrumentation__.Notify(465317)

	err = r.localPlanner.updateZoneConfigsForTables(ctx, dbDesc, WithOnlyGlobalTables)
	if err != nil {
		__antithesis_instrumentation__.Notify(465325)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465326)
	}
	__antithesis_instrumentation__.Notify(465318)

	return nil
}

func (r *databaseRegionChangeFinalizer) updateDatabaseZoneConfig(
	ctx context.Context, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(465327)
	regionConfig, err := SynthesizeRegionConfig(ctx, txn, r.dbID, r.localPlanner.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(465329)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465330)
	}
	__antithesis_instrumentation__.Notify(465328)
	return ApplyZoneConfigFromDatabaseRegionConfig(
		ctx,
		r.dbID,
		regionConfig,
		txn,
		r.localPlanner.ExecCfg(),
	)
}

func (r *databaseRegionChangeFinalizer) repartitionRegionalByRowTables(
	ctx context.Context, txn *kv.Txn,
) (repartitioned []*tabledesc.Mutable, zoneConfigUpdates []*zoneConfigUpdate, _ error) {
	__antithesis_instrumentation__.Notify(465331)
	regionConfig, err := SynthesizeRegionConfig(ctx, txn, r.dbID, r.localPlanner.Descriptors())
	if err != nil {
		__antithesis_instrumentation__.Notify(465334)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(465335)
	}
	__antithesis_instrumentation__.Notify(465332)

	for _, tableDesc := range r.regionalByRowTables {
		__antithesis_instrumentation__.Notify(465336)

		for i := range tableDesc.Columns {
			__antithesis_instrumentation__.Notify(465344)
			col := &tableDesc.Columns[i]
			if col.Type.UserDefined() {
				__antithesis_instrumentation__.Notify(465345)
				tid, err := typedesc.UserDefinedTypeOIDToID(col.Type.Oid())
				if err != nil {
					__antithesis_instrumentation__.Notify(465347)
					return nil, nil, err
				} else {
					__antithesis_instrumentation__.Notify(465348)
				}
				__antithesis_instrumentation__.Notify(465346)
				if tid == r.typeID {
					__antithesis_instrumentation__.Notify(465349)
					col.Type.TypeMeta = types.UserDefinedTypeMetadata{}
				} else {
					__antithesis_instrumentation__.Notify(465350)
				}
			} else {
				__antithesis_instrumentation__.Notify(465351)
			}
		}
		__antithesis_instrumentation__.Notify(465337)
		if err := typedesc.HydrateTypesInTableDescriptor(
			ctx,
			tableDesc.TableDesc(),
			r.localPlanner,
		); err != nil {
			__antithesis_instrumentation__.Notify(465352)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(465353)
		}
		__antithesis_instrumentation__.Notify(465338)

		colName, err := tableDesc.GetRegionalByRowTableRegionColumnName()
		if err != nil {
			__antithesis_instrumentation__.Notify(465354)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(465355)
		}
		__antithesis_instrumentation__.Notify(465339)
		partitionAllBy := partitionByForRegionalByRow(regionConfig, colName)

		oldPartitionings := make(map[descpb.IndexID]catalog.Partitioning)

		for _, index := range tableDesc.NonDropIndexes() {
			__antithesis_instrumentation__.Notify(465356)
			oldPartitionings[index.GetID()] = index.GetPartitioning().DeepCopy()
			newImplicitCols, newPartitioning, err := CreatePartitioning(
				ctx,
				r.localPlanner.extendedEvalCtx.Settings,
				r.localPlanner.EvalContext(),
				tableDesc,
				*index.IndexDesc(),
				partitionAllBy,
				nil,
				true,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(465358)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(465359)
			}
			__antithesis_instrumentation__.Notify(465357)
			tabledesc.UpdateIndexPartitioning(index.IndexDesc(), index.Primary(), newImplicitCols, newPartitioning)
		}
		__antithesis_instrumentation__.Notify(465340)

		for _, index := range tableDesc.NonDropIndexes() {
			__antithesis_instrumentation__.Notify(465360)

			update, err := prepareRemovedPartitionZoneConfigs(
				ctx,
				txn,
				tableDesc,
				index.GetID(),
				oldPartitionings[index.GetID()],
				index.GetPartitioning(),
				r.localPlanner.ExecCfg(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(465362)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(465363)
			}
			__antithesis_instrumentation__.Notify(465361)
			if update != nil {
				__antithesis_instrumentation__.Notify(465364)
				zoneConfigUpdates = append(zoneConfigUpdates, update)
			} else {
				__antithesis_instrumentation__.Notify(465365)
			}
		}
		__antithesis_instrumentation__.Notify(465341)

		update, err := prepareZoneConfigForMultiRegionTable(
			ctx,
			txn,
			r.localPlanner.ExecCfg(),
			regionConfig,
			tableDesc,
			ApplyZoneConfigForMultiRegionTableOptionTableAndIndexes,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(465366)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(465367)
		}
		__antithesis_instrumentation__.Notify(465342)
		if update != nil {
			__antithesis_instrumentation__.Notify(465368)
			zoneConfigUpdates = append(zoneConfigUpdates, update)
		} else {
			__antithesis_instrumentation__.Notify(465369)
		}
		__antithesis_instrumentation__.Notify(465343)
		repartitioned = append(repartitioned, tableDesc)
	}
	__antithesis_instrumentation__.Notify(465333)

	return repartitioned, zoneConfigUpdates, nil
}
