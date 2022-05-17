package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/backfill"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const (
	indexTruncateChunkSize = row.TableTruncateChunkSize

	indexTxnBackfillChunkSize = 100

	checkpointInterval = 2 * time.Minute
)

var indexBackfillBatchSize = settings.RegisterIntSetting(
	settings.TenantWritable,
	"bulkio.index_backfill.batch_size",
	"the number of rows for which we construct index entries in a single batch",
	50000,
	settings.NonNegativeInt,
)

var columnBackfillBatchSize = settings.RegisterIntSetting(
	settings.TenantWritable,
	"bulkio.column_backfill.batch_size",
	"the number of rows updated at a time to add/remove columns",
	200,
	settings.NonNegativeInt,
)

var _ sort.Interface = columnsByID{}
var _ sort.Interface = indexesByID{}

type columnsByID []descpb.ColumnDescriptor

func (cds columnsByID) Len() int {
	__antithesis_instrumentation__.Notify(246035)
	return len(cds)
}
func (cds columnsByID) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(246036)
	return cds[i].ID < cds[j].ID
}
func (cds columnsByID) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(246037)
	cds[i], cds[j] = cds[j], cds[i]
}

type indexesByID []descpb.IndexDescriptor

func (ids indexesByID) Len() int {
	__antithesis_instrumentation__.Notify(246038)
	return len(ids)
}
func (ids indexesByID) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(246039)
	return ids[i].ID < ids[j].ID
}
func (ids indexesByID) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(246040)
	ids[i], ids[j] = ids[j], ids[i]
}

func (sc *SchemaChanger) getChunkSize(chunkSize int64) int64 {
	__antithesis_instrumentation__.Notify(246041)
	if sc.testingKnobs.BackfillChunkSize > 0 {
		__antithesis_instrumentation__.Notify(246043)
		return sc.testingKnobs.BackfillChunkSize
	} else {
		__antithesis_instrumentation__.Notify(246044)
	}
	__antithesis_instrumentation__.Notify(246042)
	return chunkSize
}

type scTxnFn func(ctx context.Context, txn *kv.Txn, evalCtx *extendedEvalContext) error

type historicalTxnRunner func(ctx context.Context, fn scTxnFn) error

func (sc *SchemaChanger) makeFixedTimestampRunner(readAsOf hlc.Timestamp) historicalTxnRunner {
	__antithesis_instrumentation__.Notify(246045)
	runner := func(ctx context.Context, retryable scTxnFn) error {
		__antithesis_instrumentation__.Notify(246047)
		return sc.fixedTimestampTxn(ctx, readAsOf, func(
			ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
		) error {
			__antithesis_instrumentation__.Notify(246048)

			evalCtx := createSchemaChangeEvalCtx(ctx, sc.execCfg, readAsOf, descriptors)
			return retryable(ctx, txn, &evalCtx)
		})
	}
	__antithesis_instrumentation__.Notify(246046)
	return runner
}

func (sc *SchemaChanger) makeFixedTimestampInternalExecRunner(
	readAsOf hlc.Timestamp,
) sqlutil.HistoricalInternalExecTxnRunner {
	__antithesis_instrumentation__.Notify(246049)
	runner := func(ctx context.Context, retryable sqlutil.InternalExecFn) error {
		__antithesis_instrumentation__.Notify(246051)
		return sc.fixedTimestampTxn(ctx, readAsOf, func(
			ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
		) error {
			__antithesis_instrumentation__.Notify(246052)

			ie := sc.ieFactory(ctx, NewFakeSessionData(sc.execCfg.SV()))
			return retryable(ctx, txn, ie)
		})
	}
	__antithesis_instrumentation__.Notify(246050)
	return runner
}

func (sc *SchemaChanger) fixedTimestampTxn(
	ctx context.Context,
	readAsOf hlc.Timestamp,
	retryable func(ctx context.Context, txn *kv.Txn, descriptors *descs.Collection) error,
) error {
	__antithesis_instrumentation__.Notify(246053)
	return sc.txn(ctx, func(ctx context.Context, txn *kv.Txn, descriptors *descs.Collection) error {
		__antithesis_instrumentation__.Notify(246054)
		if err := txn.SetFixedTimestamp(ctx, readAsOf); err != nil {
			__antithesis_instrumentation__.Notify(246056)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246057)
		}
		__antithesis_instrumentation__.Notify(246055)
		return retryable(ctx, txn, descriptors)
	})
}

func (sc *SchemaChanger) runBackfill(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(246058)
	if sc.testingKnobs.RunBeforeBackfill != nil {
		__antithesis_instrumentation__.Notify(246070)
		if err := sc.testingKnobs.RunBeforeBackfill(); err != nil {
			__antithesis_instrumentation__.Notify(246071)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246072)
		}
	} else {
		__antithesis_instrumentation__.Notify(246073)
	}
	__antithesis_instrumentation__.Notify(246059)

	var addedIndexSpans []roachpb.Span
	var addedIndexes []descpb.IndexID
	var temporaryIndexes []descpb.IndexID

	var constraintsToDrop []catalog.ConstraintToUpdate
	var constraintsToAddBeforeValidation []catalog.ConstraintToUpdate
	var constraintsToValidate []catalog.ConstraintToUpdate

	var viewToRefresh catalog.MaterializedViewRefresh

	tableDesc, err := sc.updateJobRunningStatus(ctx, RunningStatusBackfill)
	if err != nil {
		__antithesis_instrumentation__.Notify(246074)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246075)
	}
	__antithesis_instrumentation__.Notify(246060)

	if tableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(246076)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(246077)
	}
	__antithesis_instrumentation__.Notify(246061)
	version := tableDesc.GetVersion()

	log.Infof(ctx, "running backfill for %q, v=%d", tableDesc.GetName(), tableDesc.GetVersion())

	needColumnBackfill := false
	for _, m := range tableDesc.AllMutations() {
		__antithesis_instrumentation__.Notify(246078)
		if m.MutationID() != sc.mutationID {
			__antithesis_instrumentation__.Notify(246081)
			break
		} else {
			__antithesis_instrumentation__.Notify(246082)
		}
		__antithesis_instrumentation__.Notify(246079)

		if discarded, _ := isCurrentMutationDiscarded(tableDesc, m, m.MutationOrdinal()+1); discarded {
			__antithesis_instrumentation__.Notify(246083)
			continue
		} else {
			__antithesis_instrumentation__.Notify(246084)
		}
		__antithesis_instrumentation__.Notify(246080)

		if m.Adding() {
			__antithesis_instrumentation__.Notify(246085)
			if col := m.AsColumn(); col != nil {
				__antithesis_instrumentation__.Notify(246086)

				needColumnBackfill = needColumnBackfill || func() bool {
					__antithesis_instrumentation__.Notify(246087)
					return catalog.ColumnNeedsBackfill(col) == true
				}() == true
			} else {
				__antithesis_instrumentation__.Notify(246088)
				if idx := m.AsIndex(); idx != nil {
					__antithesis_instrumentation__.Notify(246089)
					if idx.IsTemporaryIndexForBackfill() {
						__antithesis_instrumentation__.Notify(246090)
						temporaryIndexes = append(temporaryIndexes, idx.GetID())
					} else {
						__antithesis_instrumentation__.Notify(246091)
						addedIndexSpans = append(addedIndexSpans, tableDesc.IndexSpan(sc.execCfg.Codec, idx.GetID()))
						addedIndexes = append(addedIndexes, idx.GetID())
					}
				} else {
					__antithesis_instrumentation__.Notify(246092)
					if c := m.AsConstraint(); c != nil {
						__antithesis_instrumentation__.Notify(246093)
						isValidating := c.IsCheck() && func() bool {
							__antithesis_instrumentation__.Notify(246096)
							return c.Check().Validity == descpb.ConstraintValidity_Validating == true
						}() == true || func() bool {
							__antithesis_instrumentation__.Notify(246097)
							return (c.IsForeignKey() && func() bool {
								__antithesis_instrumentation__.Notify(246098)
								return c.ForeignKey().Validity == descpb.ConstraintValidity_Validating == true
							}() == true) == true
						}() == true || func() bool {
							__antithesis_instrumentation__.Notify(246099)
							return (c.IsUniqueWithoutIndex() && func() bool {
								__antithesis_instrumentation__.Notify(246100)
								return c.UniqueWithoutIndex().Validity == descpb.ConstraintValidity_Validating == true
							}() == true) == true
						}() == true || func() bool {
							__antithesis_instrumentation__.Notify(246101)
							return c.IsNotNull() == true
						}() == true
						isSkippingValidation, err := shouldSkipConstraintValidation(tableDesc, c)
						if err != nil {
							__antithesis_instrumentation__.Notify(246102)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246103)
						}
						__antithesis_instrumentation__.Notify(246094)
						if isValidating {
							__antithesis_instrumentation__.Notify(246104)
							constraintsToAddBeforeValidation = append(constraintsToAddBeforeValidation, c)
						} else {
							__antithesis_instrumentation__.Notify(246105)
						}
						__antithesis_instrumentation__.Notify(246095)
						if isValidating && func() bool {
							__antithesis_instrumentation__.Notify(246106)
							return !isSkippingValidation == true
						}() == true {
							__antithesis_instrumentation__.Notify(246107)
							constraintsToValidate = append(constraintsToValidate, c)
						} else {
							__antithesis_instrumentation__.Notify(246108)
						}
					} else {
						__antithesis_instrumentation__.Notify(246109)
						if mvRefresh := m.AsMaterializedViewRefresh(); mvRefresh != nil {
							__antithesis_instrumentation__.Notify(246110)
							viewToRefresh = mvRefresh
						} else {
							__antithesis_instrumentation__.Notify(246111)
							if m.AsPrimaryKeySwap() != nil || func() bool {
								__antithesis_instrumentation__.Notify(246112)
								return m.AsComputedColumnSwap() != nil == true
							}() == true || func() bool {
								__antithesis_instrumentation__.Notify(246113)
								return m.AsModifyRowLevelTTL() != nil == true
							}() == true {
								__antithesis_instrumentation__.Notify(246114)

							} else {
								__antithesis_instrumentation__.Notify(246115)
								return errors.AssertionFailedf("unsupported mutation: %+v", m)
							}
						}
					}
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(246116)
			if m.Dropped() {
				__antithesis_instrumentation__.Notify(246117)
				if col := m.AsColumn(); col != nil {
					__antithesis_instrumentation__.Notify(246118)

					needColumnBackfill = needColumnBackfill || func() bool {
						__antithesis_instrumentation__.Notify(246119)
						return catalog.ColumnNeedsBackfill(col) == true
					}() == true
				} else {
					__antithesis_instrumentation__.Notify(246120)
					if idx := m.AsIndex(); idx != nil {
						__antithesis_instrumentation__.Notify(246121)

					} else {
						__antithesis_instrumentation__.Notify(246122)
						if c := m.AsConstraint(); c != nil {
							__antithesis_instrumentation__.Notify(246123)
							constraintsToDrop = append(constraintsToDrop, c)
						} else {
							__antithesis_instrumentation__.Notify(246124)
							if m.AsPrimaryKeySwap() != nil || func() bool {
								__antithesis_instrumentation__.Notify(246125)
								return m.AsComputedColumnSwap() != nil == true
							}() == true || func() bool {
								__antithesis_instrumentation__.Notify(246126)
								return m.AsMaterializedViewRefresh() != nil == true
							}() == true || func() bool {
								__antithesis_instrumentation__.Notify(246127)
								return m.AsModifyRowLevelTTL() != nil == true
							}() == true {
								__antithesis_instrumentation__.Notify(246128)

							} else {
								__antithesis_instrumentation__.Notify(246129)
								return errors.AssertionFailedf("unsupported mutation: %+v", m)
							}
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(246130)
			}
		}
	}
	__antithesis_instrumentation__.Notify(246062)

	if viewToRefresh != nil {
		__antithesis_instrumentation__.Notify(246131)
		if err := sc.refreshMaterializedView(ctx, tableDesc, viewToRefresh); err != nil {
			__antithesis_instrumentation__.Notify(246132)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246133)
		}
	} else {
		__antithesis_instrumentation__.Notify(246134)
	}
	__antithesis_instrumentation__.Notify(246063)

	if len(constraintsToDrop) > 0 {
		__antithesis_instrumentation__.Notify(246135)
		descs, err := sc.dropConstraints(ctx, constraintsToDrop)
		if err != nil {
			__antithesis_instrumentation__.Notify(246137)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246138)
		}
		__antithesis_instrumentation__.Notify(246136)
		version = descs[tableDesc.GetID()].GetVersion()
	} else {
		__antithesis_instrumentation__.Notify(246139)
	}
	__antithesis_instrumentation__.Notify(246064)

	if needColumnBackfill {
		__antithesis_instrumentation__.Notify(246140)
		if err := sc.truncateAndBackfillColumns(ctx, version); err != nil {
			__antithesis_instrumentation__.Notify(246141)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246142)
		}
	} else {
		__antithesis_instrumentation__.Notify(246143)
	}
	__antithesis_instrumentation__.Notify(246065)

	if len(addedIndexSpans) > 0 {
		__antithesis_instrumentation__.Notify(246144)

		if err := sc.backfillIndexes(ctx, version, addedIndexSpans, addedIndexes, temporaryIndexes); err != nil {
			__antithesis_instrumentation__.Notify(246145)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246146)
		}
	} else {
		__antithesis_instrumentation__.Notify(246147)
	}
	__antithesis_instrumentation__.Notify(246066)

	if len(constraintsToAddBeforeValidation) > 0 {
		__antithesis_instrumentation__.Notify(246148)
		if err := sc.addConstraints(ctx, constraintsToAddBeforeValidation); err != nil {
			__antithesis_instrumentation__.Notify(246149)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246150)
		}
	} else {
		__antithesis_instrumentation__.Notify(246151)
	}
	__antithesis_instrumentation__.Notify(246067)

	if len(constraintsToValidate) > 0 {
		__antithesis_instrumentation__.Notify(246152)
		if err := sc.validateConstraints(ctx, constraintsToValidate); err != nil {
			__antithesis_instrumentation__.Notify(246153)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246154)
		}
	} else {
		__antithesis_instrumentation__.Notify(246155)
	}
	__antithesis_instrumentation__.Notify(246068)

	log.Infof(ctx, "completed backfill for %q, v=%d", tableDesc.GetName(), tableDesc.GetVersion())

	if sc.testingKnobs.RunAfterBackfill != nil {
		__antithesis_instrumentation__.Notify(246156)
		if err := sc.testingKnobs.RunAfterBackfill(sc.job.ID()); err != nil {
			__antithesis_instrumentation__.Notify(246157)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246158)
		}
	} else {
		__antithesis_instrumentation__.Notify(246159)
	}
	__antithesis_instrumentation__.Notify(246069)

	return nil
}

func shouldSkipConstraintValidation(
	tableDesc catalog.TableDescriptor, c catalog.ConstraintToUpdate,
) (bool, error) {
	__antithesis_instrumentation__.Notify(246160)
	if !c.IsCheck() {
		__antithesis_instrumentation__.Notify(246164)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(246165)
	}
	__antithesis_instrumentation__.Notify(246161)

	check := c.Check()

	if len(check.ColumnIDs) != 1 {
		__antithesis_instrumentation__.Notify(246166)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(246167)
	}
	__antithesis_instrumentation__.Notify(246162)

	checkCol, err := tableDesc.FindColumnWithID(check.ColumnIDs[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(246168)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(246169)
	}
	__antithesis_instrumentation__.Notify(246163)

	return tableDesc.IsShardColumn(checkCol) && func() bool {
		__antithesis_instrumentation__.Notify(246170)
		return checkCol.Adding() == true
	}() == true, nil
}

func (sc *SchemaChanger) dropConstraints(
	ctx context.Context, constraints []catalog.ConstraintToUpdate,
) (map[descpb.ID]catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(246171)
	log.Infof(ctx, "dropping %d constraints", len(constraints))

	fksByBackrefTable := make(map[descpb.ID][]catalog.ConstraintToUpdate)
	for _, c := range constraints {
		__antithesis_instrumentation__.Notify(246175)
		if c.IsForeignKey() {
			__antithesis_instrumentation__.Notify(246176)
			id := c.ForeignKey().ReferencedTableID
			if id != sc.descID {
				__antithesis_instrumentation__.Notify(246177)
				fksByBackrefTable[id] = append(fksByBackrefTable[id], c)
			} else {
				__antithesis_instrumentation__.Notify(246178)
			}
		} else {
			__antithesis_instrumentation__.Notify(246179)
		}
	}
	__antithesis_instrumentation__.Notify(246172)

	if err := sc.txn(ctx, func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
		__antithesis_instrumentation__.Notify(246180)
		scTable, err := descsCol.GetMutableTableVersionByID(ctx, sc.descID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(246184)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246185)
		}
		__antithesis_instrumentation__.Notify(246181)
		b := txn.NewBatch()
		for _, constraint := range constraints {
			__antithesis_instrumentation__.Notify(246186)
			if constraint.IsCheck() || func() bool {
				__antithesis_instrumentation__.Notify(246187)
				return constraint.IsNotNull() == true
			}() == true {
				__antithesis_instrumentation__.Notify(246188)
				found := false
				for j, c := range scTable.Checks {
					__antithesis_instrumentation__.Notify(246190)
					if c.Name == constraint.GetName() {
						__antithesis_instrumentation__.Notify(246191)
						scTable.Checks = append(scTable.Checks[:j], scTable.Checks[j+1:]...)
						found = true
						break
					} else {
						__antithesis_instrumentation__.Notify(246192)
					}
				}
				__antithesis_instrumentation__.Notify(246189)
				if !found {
					__antithesis_instrumentation__.Notify(246193)
					log.VEventf(
						ctx, 2,
						"backfiller tried to drop constraint %+v but it was not found, "+
							"presumably due to a retry or rollback",
						constraint.ConstraintToUpdateDesc(),
					)
				} else {
					__antithesis_instrumentation__.Notify(246194)
				}
			} else {
				__antithesis_instrumentation__.Notify(246195)
				if constraint.IsForeignKey() {
					__antithesis_instrumentation__.Notify(246196)
					var foundExisting bool
					for j := range scTable.OutboundFKs {
						__antithesis_instrumentation__.Notify(246198)
						def := &scTable.OutboundFKs[j]
						if def.Name != constraint.GetName() {
							__antithesis_instrumentation__.Notify(246203)
							continue
						} else {
							__antithesis_instrumentation__.Notify(246204)
						}
						__antithesis_instrumentation__.Notify(246199)
						backrefTable, err := descsCol.GetMutableTableVersionByID(ctx,
							constraint.ForeignKey().ReferencedTableID, txn)
						if err != nil {
							__antithesis_instrumentation__.Notify(246205)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246206)
						}
						__antithesis_instrumentation__.Notify(246200)
						if err := removeFKBackReferenceFromTable(
							backrefTable, def.Name, scTable,
						); err != nil {
							__antithesis_instrumentation__.Notify(246207)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246208)
						}
						__antithesis_instrumentation__.Notify(246201)
						if err := descsCol.WriteDescToBatch(
							ctx, true, backrefTable, b,
						); err != nil {
							__antithesis_instrumentation__.Notify(246209)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246210)
						}
						__antithesis_instrumentation__.Notify(246202)
						scTable.OutboundFKs = append(scTable.OutboundFKs[:j], scTable.OutboundFKs[j+1:]...)
						foundExisting = true
						break
					}
					__antithesis_instrumentation__.Notify(246197)
					if !foundExisting {
						__antithesis_instrumentation__.Notify(246211)
						log.VEventf(
							ctx, 2,
							"backfiller tried to drop constraint %+v but it was not found, "+
								"presumably due to a retry or rollback",
							constraint.ConstraintToUpdateDesc(),
						)
					} else {
						__antithesis_instrumentation__.Notify(246212)
					}
				} else {
					__antithesis_instrumentation__.Notify(246213)
					if constraint.IsUniqueWithoutIndex() {
						__antithesis_instrumentation__.Notify(246214)
						found := false
						for j, c := range scTable.UniqueWithoutIndexConstraints {
							__antithesis_instrumentation__.Notify(246216)
							if c.Name == constraint.GetName() {
								__antithesis_instrumentation__.Notify(246217)
								scTable.UniqueWithoutIndexConstraints = append(
									scTable.UniqueWithoutIndexConstraints[:j],
									scTable.UniqueWithoutIndexConstraints[j+1:]...,
								)
								found = true
								break
							} else {
								__antithesis_instrumentation__.Notify(246218)
							}
						}
						__antithesis_instrumentation__.Notify(246215)
						if !found {
							__antithesis_instrumentation__.Notify(246219)
							log.VEventf(
								ctx, 2,
								"backfiller tried to drop constraint %+v but it was not found, "+
									"presumably due to a retry or rollback",
								constraint.ConstraintToUpdateDesc(),
							)
						} else {
							__antithesis_instrumentation__.Notify(246220)
						}
					} else {
						__antithesis_instrumentation__.Notify(246221)
					}
				}
			}
		}
		__antithesis_instrumentation__.Notify(246182)
		if err := descsCol.WriteDescToBatch(
			ctx, true, scTable, b,
		); err != nil {
			__antithesis_instrumentation__.Notify(246222)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246223)
		}
		__antithesis_instrumentation__.Notify(246183)
		return txn.Run(ctx, b)
	}); err != nil {
		__antithesis_instrumentation__.Notify(246224)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(246225)
	}
	__antithesis_instrumentation__.Notify(246173)

	log.Info(ctx, "finished dropping constraints")
	tableDescs := make(map[descpb.ID]catalog.TableDescriptor, len(fksByBackrefTable)+1)
	if err := sc.txn(ctx, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) (err error) {
		__antithesis_instrumentation__.Notify(246226)
		if tableDescs[sc.descID], err = descsCol.GetImmutableTableByID(
			ctx, txn, sc.descID, tree.ObjectLookupFlags{},
		); err != nil {
			__antithesis_instrumentation__.Notify(246229)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246230)
		}
		__antithesis_instrumentation__.Notify(246227)
		for id := range fksByBackrefTable {
			__antithesis_instrumentation__.Notify(246231)
			if tableDescs[id], err = descsCol.GetImmutableTableByID(
				ctx, txn, id, tree.ObjectLookupFlags{},
			); err != nil {
				__antithesis_instrumentation__.Notify(246232)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246233)
			}
		}
		__antithesis_instrumentation__.Notify(246228)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(246234)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(246235)
	}
	__antithesis_instrumentation__.Notify(246174)
	return tableDescs, nil
}

func (sc *SchemaChanger) addConstraints(
	ctx context.Context, constraints []catalog.ConstraintToUpdate,
) error {
	__antithesis_instrumentation__.Notify(246236)
	log.Infof(ctx, "adding %d constraints", len(constraints))

	fksByBackrefTable := make(map[descpb.ID][]catalog.ConstraintToUpdate)
	for _, c := range constraints {
		__antithesis_instrumentation__.Notify(246239)
		if c.IsForeignKey() {
			__antithesis_instrumentation__.Notify(246240)
			id := c.ForeignKey().ReferencedTableID
			if id != sc.descID {
				__antithesis_instrumentation__.Notify(246241)
				fksByBackrefTable[id] = append(fksByBackrefTable[id], c)
			} else {
				__antithesis_instrumentation__.Notify(246242)
			}
		} else {
			__antithesis_instrumentation__.Notify(246243)
		}
	}
	__antithesis_instrumentation__.Notify(246237)

	if err := sc.txn(ctx, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(246244)
		scTable, err := descsCol.GetMutableTableVersionByID(ctx, sc.descID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(246248)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246249)
		}
		__antithesis_instrumentation__.Notify(246245)

		b := txn.NewBatch()
		for _, constraint := range constraints {
			__antithesis_instrumentation__.Notify(246250)
			if constraint.IsCheck() || func() bool {
				__antithesis_instrumentation__.Notify(246251)
				return constraint.IsNotNull() == true
			}() == true {
				__antithesis_instrumentation__.Notify(246252)
				found := false
				for _, c := range scTable.Checks {
					__antithesis_instrumentation__.Notify(246254)
					if c.Name == constraint.GetName() {
						__antithesis_instrumentation__.Notify(246255)
						log.VEventf(
							ctx, 2,
							"backfiller tried to add constraint %+v but found existing constraint %+v, "+
								"presumably due to a retry or rollback",
							constraint, c,
						)

						c.Validity = descpb.ConstraintValidity_Validating
						found = true
						break
					} else {
						__antithesis_instrumentation__.Notify(246256)
					}
				}
				__antithesis_instrumentation__.Notify(246253)
				if !found {
					__antithesis_instrumentation__.Notify(246257)
					scTable.Checks = append(scTable.Checks, &constraint.ConstraintToUpdateDesc().Check)
				} else {
					__antithesis_instrumentation__.Notify(246258)
				}
			} else {
				__antithesis_instrumentation__.Notify(246259)
				if constraint.IsForeignKey() {
					__antithesis_instrumentation__.Notify(246260)
					var foundExisting bool
					for j := range scTable.OutboundFKs {
						__antithesis_instrumentation__.Notify(246262)
						def := &scTable.OutboundFKs[j]
						if def.Name == constraint.GetName() {
							__antithesis_instrumentation__.Notify(246263)
							if log.V(2) {
								__antithesis_instrumentation__.Notify(246265)
								log.VEventf(
									ctx, 2,
									"backfiller tried to add constraint %+v but found existing constraint %+v, "+
										"presumably due to a retry or rollback",
									constraint.ConstraintToUpdateDesc(), def,
								)
							} else {
								__antithesis_instrumentation__.Notify(246266)
							}
							__antithesis_instrumentation__.Notify(246264)

							def.Validity = descpb.ConstraintValidity_Validating
							foundExisting = true
							break
						} else {
							__antithesis_instrumentation__.Notify(246267)
						}
					}
					__antithesis_instrumentation__.Notify(246261)
					if !foundExisting {
						__antithesis_instrumentation__.Notify(246268)
						scTable.OutboundFKs = append(scTable.OutboundFKs, constraint.ForeignKey())
						backrefTable, err := descsCol.GetMutableTableVersionByID(ctx, constraint.ForeignKey().ReferencedTableID, txn)
						if err != nil {
							__antithesis_instrumentation__.Notify(246271)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246272)
						}
						__antithesis_instrumentation__.Notify(246269)

						_, err = tabledesc.FindFKReferencedUniqueConstraint(backrefTable, constraint.ForeignKey().ReferencedColumnIDs)
						if err != nil {
							__antithesis_instrumentation__.Notify(246273)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246274)
						}
						__antithesis_instrumentation__.Notify(246270)
						backrefTable.InboundFKs = append(backrefTable.InboundFKs, constraint.ForeignKey())

						if backrefTable != scTable {
							__antithesis_instrumentation__.Notify(246275)
							if err := descsCol.WriteDescToBatch(
								ctx, true, backrefTable, b,
							); err != nil {
								__antithesis_instrumentation__.Notify(246276)
								return err
							} else {
								__antithesis_instrumentation__.Notify(246277)
							}
						} else {
							__antithesis_instrumentation__.Notify(246278)
						}
					} else {
						__antithesis_instrumentation__.Notify(246279)
					}
				} else {
					__antithesis_instrumentation__.Notify(246280)
					if constraint.IsUniqueWithoutIndex() {
						__antithesis_instrumentation__.Notify(246281)
						found := false
						for _, c := range scTable.UniqueWithoutIndexConstraints {
							__antithesis_instrumentation__.Notify(246283)
							if c.Name == constraint.GetName() {
								__antithesis_instrumentation__.Notify(246284)
								log.VEventf(
									ctx, 2,
									"backfiller tried to add constraint %+v but found existing constraint %+v, "+
										"presumably due to a retry or rollback",
									constraint.ConstraintToUpdateDesc(), c,
								)

								c.Validity = descpb.ConstraintValidity_Validating
								found = true
								break
							} else {
								__antithesis_instrumentation__.Notify(246285)
							}
						}
						__antithesis_instrumentation__.Notify(246282)
						if !found {
							__antithesis_instrumentation__.Notify(246286)
							scTable.UniqueWithoutIndexConstraints = append(scTable.UniqueWithoutIndexConstraints,
								constraint.UniqueWithoutIndex())
						} else {
							__antithesis_instrumentation__.Notify(246287)
						}
					} else {
						__antithesis_instrumentation__.Notify(246288)
					}
				}
			}
		}
		__antithesis_instrumentation__.Notify(246246)
		if err := descsCol.WriteDescToBatch(
			ctx, true, scTable, b,
		); err != nil {
			__antithesis_instrumentation__.Notify(246289)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246290)
		}
		__antithesis_instrumentation__.Notify(246247)
		return txn.Run(ctx, b)
	}); err != nil {
		__antithesis_instrumentation__.Notify(246291)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246292)
	}
	__antithesis_instrumentation__.Notify(246238)
	log.Info(ctx, "finished adding constraints")
	return nil
}

func (sc *SchemaChanger) validateConstraints(
	ctx context.Context, constraints []catalog.ConstraintToUpdate,
) error {
	__antithesis_instrumentation__.Notify(246293)
	if lease.TestingTableLeasesAreDisabled() {
		__antithesis_instrumentation__.Notify(246300)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(246301)
	}
	__antithesis_instrumentation__.Notify(246294)
	log.Infof(ctx, "validating %d new constraints", len(constraints))

	_, err := sc.updateJobRunningStatus(ctx, RunningStatusValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(246302)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246303)
	}
	__antithesis_instrumentation__.Notify(246295)

	if fn := sc.testingKnobs.RunBeforeConstraintValidation; fn != nil {
		__antithesis_instrumentation__.Notify(246304)
		if err := fn(constraints); err != nil {
			__antithesis_instrumentation__.Notify(246305)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246306)
		}
	} else {
		__antithesis_instrumentation__.Notify(246307)
	}
	__antithesis_instrumentation__.Notify(246296)

	readAsOf := sc.clock.Now()
	var tableDesc catalog.TableDescriptor

	if err := sc.fixedTimestampTxn(ctx, readAsOf, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(246308)
		flags := tree.ObjectLookupFlagsWithRequired()
		flags.AvoidLeased = true
		tableDesc, err = descriptors.GetImmutableTableByID(ctx, txn, sc.descID, flags)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(246309)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246310)
	}
	__antithesis_instrumentation__.Notify(246297)

	grp := ctxgroup.WithContext(ctx)

	runHistoricalTxn := sc.makeFixedTimestampRunner(readAsOf)

	for i := range constraints {
		__antithesis_instrumentation__.Notify(246311)
		c := constraints[i]
		grp.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(246312)

			descI, err := tableDesc.MakeFirstMutationPublic(catalog.IgnoreConstraints)
			if err != nil {
				__antithesis_instrumentation__.Notify(246314)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246315)
			}
			__antithesis_instrumentation__.Notify(246313)
			desc := descI.(*tabledesc.Mutable)

			return runHistoricalTxn(ctx, func(ctx context.Context, txn *kv.Txn, evalCtx *extendedEvalContext) error {
				__antithesis_instrumentation__.Notify(246316)

				evalCtx.Txn = txn

				collection := evalCtx.Descs
				resolver := descs.NewDistSQLTypeResolver(collection, txn)
				semaCtx := tree.MakeSemaContext()
				semaCtx.TypeResolver = &resolver

				defer func() { __antithesis_instrumentation__.Notify(246319); collection.ReleaseAll(ctx) }()
				__antithesis_instrumentation__.Notify(246317)
				if c.IsCheck() {
					__antithesis_instrumentation__.Notify(246320)
					if err := validateCheckInTxn(
						ctx, &semaCtx, sc.ieFactory, evalCtx.SessionData(), desc, txn, c.Check().Expr,
					); err != nil {
						__antithesis_instrumentation__.Notify(246321)
						return err
					} else {
						__antithesis_instrumentation__.Notify(246322)
					}
				} else {
					__antithesis_instrumentation__.Notify(246323)
					if c.IsForeignKey() {
						__antithesis_instrumentation__.Notify(246324)
						if err := validateFkInTxn(ctx, sc.ieFactory, evalCtx.SessionData(), desc, txn, collection, c.GetName()); err != nil {
							__antithesis_instrumentation__.Notify(246325)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246326)
						}
					} else {
						__antithesis_instrumentation__.Notify(246327)
						if c.IsUniqueWithoutIndex() {
							__antithesis_instrumentation__.Notify(246328)
							if err := validateUniqueWithoutIndexConstraintInTxn(ctx, sc.ieFactory(ctx, evalCtx.SessionData()), desc, txn, c.GetName()); err != nil {
								__antithesis_instrumentation__.Notify(246329)
								return err
							} else {
								__antithesis_instrumentation__.Notify(246330)
							}
						} else {
							__antithesis_instrumentation__.Notify(246331)
							if c.IsNotNull() {
								__antithesis_instrumentation__.Notify(246332)
								if err := validateCheckInTxn(
									ctx, &semaCtx, sc.ieFactory, evalCtx.SessionData(), desc, txn, c.Check().Expr,
								); err != nil {
									__antithesis_instrumentation__.Notify(246333)

									return errors.Wrap(err, "validation of NOT NULL constraint failed")
								} else {
									__antithesis_instrumentation__.Notify(246334)
								}
							} else {
								__antithesis_instrumentation__.Notify(246335)
								return errors.Errorf("unsupported constraint type: %d", c.ConstraintToUpdateDesc().ConstraintType)
							}
						}
					}
				}
				__antithesis_instrumentation__.Notify(246318)
				return nil
			})
		})
	}
	__antithesis_instrumentation__.Notify(246298)
	if err := grp.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(246336)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246337)
	}
	__antithesis_instrumentation__.Notify(246299)
	log.Info(ctx, "finished validating new constraints")
	return nil
}

func (sc *SchemaChanger) getTableVersion(
	ctx context.Context, txn *kv.Txn, tc *descs.Collection, version descpb.DescriptorVersion,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(246338)
	tableDesc, err := tc.GetImmutableTableByID(ctx, txn, sc.descID, tree.ObjectLookupFlags{})
	if err != nil {
		__antithesis_instrumentation__.Notify(246341)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(246342)
	}
	__antithesis_instrumentation__.Notify(246339)
	if version != tableDesc.GetVersion() {
		__antithesis_instrumentation__.Notify(246343)
		return nil, makeErrTableVersionMismatch(tableDesc.GetVersion(), version)
	} else {
		__antithesis_instrumentation__.Notify(246344)
	}
	__antithesis_instrumentation__.Notify(246340)
	return tableDesc, nil
}

func getJobIDForMutationWithDescriptor(
	ctx context.Context, tableDesc catalog.TableDescriptor, mutationID descpb.MutationID,
) (jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(246345)
	for _, job := range tableDesc.GetMutationJobs() {
		__antithesis_instrumentation__.Notify(246347)
		if job.MutationID == mutationID {
			__antithesis_instrumentation__.Notify(246348)
			return job.JobID, nil
		} else {
			__antithesis_instrumentation__.Notify(246349)
		}
	}
	__antithesis_instrumentation__.Notify(246346)

	return jobspb.InvalidJobID, errors.AssertionFailedf(
		"job not found for table id %d, mutation %d", tableDesc.GetID(), mutationID)
}

func numRangesInSpans(
	ctx context.Context, db *kv.DB, distSQLPlanner *DistSQLPlanner, spans []roachpb.Span,
) (int, error) {
	__antithesis_instrumentation__.Notify(246350)
	txn := db.NewTxn(ctx, "num-ranges-in-spans")
	spanResolver := distSQLPlanner.spanResolver.NewSpanResolverIterator(txn)
	rangeIds := make(map[int64]struct{})
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(246352)

		spanResolver.Seek(ctx, span, kvcoord.Ascending)
		for {
			__antithesis_instrumentation__.Notify(246353)
			if !spanResolver.Valid() {
				__antithesis_instrumentation__.Notify(246356)
				return 0, spanResolver.Error()
			} else {
				__antithesis_instrumentation__.Notify(246357)
			}
			__antithesis_instrumentation__.Notify(246354)
			rangeIds[int64(spanResolver.Desc().RangeID)] = struct{}{}
			if !spanResolver.NeedAnother() {
				__antithesis_instrumentation__.Notify(246358)
				break
			} else {
				__antithesis_instrumentation__.Notify(246359)
			}
			__antithesis_instrumentation__.Notify(246355)
			spanResolver.Next(ctx)
		}
	}
	__antithesis_instrumentation__.Notify(246351)

	return len(rangeIds), nil
}

func NumRangesInSpanContainedBy(
	ctx context.Context,
	db *kv.DB,
	distSQLPlanner *DistSQLPlanner,
	outerSpan roachpb.Span,
	containedBy []roachpb.Span,
) (total, inContainedBy int, _ error) {
	__antithesis_instrumentation__.Notify(246360)
	txn := db.NewTxn(ctx, "num-ranges-in-spans")
	spanResolver := distSQLPlanner.spanResolver.NewSpanResolverIterator(txn)

	spanResolver.Seek(ctx, outerSpan, kvcoord.Ascending)
	var g roachpb.SpanGroup
	g.Add(containedBy...)
	for {
		__antithesis_instrumentation__.Notify(246362)
		if !spanResolver.Valid() {
			__antithesis_instrumentation__.Notify(246366)
			return 0, 0, spanResolver.Error()
		} else {
			__antithesis_instrumentation__.Notify(246367)
		}
		__antithesis_instrumentation__.Notify(246363)
		total++
		desc := spanResolver.Desc()
		if g.Encloses(desc.RSpan().AsRawSpanWithNoLocals().Intersect(outerSpan)) {
			__antithesis_instrumentation__.Notify(246368)
			inContainedBy++
		} else {
			__antithesis_instrumentation__.Notify(246369)
		}
		__antithesis_instrumentation__.Notify(246364)
		if !spanResolver.NeedAnother() {
			__antithesis_instrumentation__.Notify(246370)
			break
		} else {
			__antithesis_instrumentation__.Notify(246371)
		}
		__antithesis_instrumentation__.Notify(246365)
		spanResolver.Next(ctx)
	}
	__antithesis_instrumentation__.Notify(246361)
	return total, inContainedBy, nil
}

func (sc *SchemaChanger) distIndexBackfill(
	ctx context.Context,
	version descpb.DescriptorVersion,
	targetSpans []roachpb.Span,
	addedIndexes []descpb.IndexID,
	writeAtRequestTimestamp bool,
	filter backfill.MutationFilter,
	fractionScaler *multiStageFractionScaler,
) error {
	__antithesis_instrumentation__.Notify(246372)

	var todoSpans []roachpb.Span
	var mutationIdx int

	if err := DescsTxn(ctx, sc.execCfg, func(ctx context.Context, txn *kv.Txn, col *descs.Collection) (err error) {
		__antithesis_instrumentation__.Notify(246387)
		todoSpans, _, mutationIdx, err = rowexec.GetResumeSpans(
			ctx, sc.jobRegistry, txn, sc.execCfg.Codec, col, sc.descID, sc.mutationID, filter)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(246388)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246389)
	}
	__antithesis_instrumentation__.Notify(246373)

	log.VEventf(ctx, 2, "indexbackfill: initial resume spans %+v", todoSpans)

	if todoSpans == nil {
		__antithesis_instrumentation__.Notify(246390)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(246391)
	}
	__antithesis_instrumentation__.Notify(246374)

	writeAsOf := sc.job.Details().(jobspb.SchemaChangeDetails).WriteTimestamp
	if writeAsOf.IsEmpty() {
		__antithesis_instrumentation__.Notify(246392)
		if err := sc.job.RunningStatus(ctx, nil, func(_ context.Context, _ jobspb.Details) (jobs.RunningStatus, error) {
			__antithesis_instrumentation__.Notify(246397)
			return jobs.RunningStatus("scanning target index for in-progress transactions"), nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(246398)
			return errors.Wrapf(err, "failed to update running status of job %d", errors.Safe(sc.job.ID()))
		} else {
			__antithesis_instrumentation__.Notify(246399)
		}
		__antithesis_instrumentation__.Notify(246393)
		writeAsOf = sc.clock.Now()
		log.Infof(ctx, "starting scan of target index as of %v...", writeAsOf)

		const pageSize = 10000
		noop := func(_ []kv.KeyValue) error { __antithesis_instrumentation__.Notify(246400); return nil }
		__antithesis_instrumentation__.Notify(246394)
		if err := sc.fixedTimestampTxn(ctx, writeAsOf, func(
			ctx context.Context, txn *kv.Txn, _ *descs.Collection,
		) error {
			__antithesis_instrumentation__.Notify(246401)
			for _, span := range targetSpans {
				__antithesis_instrumentation__.Notify(246403)

				if err := txn.Iterate(ctx, span.Key, span.EndKey, pageSize, noop); err != nil {
					__antithesis_instrumentation__.Notify(246404)
					return err
				} else {
					__antithesis_instrumentation__.Notify(246405)
				}
			}
			__antithesis_instrumentation__.Notify(246402)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(246406)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246407)
		}
		__antithesis_instrumentation__.Notify(246395)
		log.Infof(ctx, "persisting target safe write time %v...", writeAsOf)
		if err := sc.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(246408)
			details := sc.job.Details().(jobspb.SchemaChangeDetails)
			details.WriteTimestamp = writeAsOf
			return sc.job.SetDetails(ctx, txn, details)
		}); err != nil {
			__antithesis_instrumentation__.Notify(246409)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246410)
		}
		__antithesis_instrumentation__.Notify(246396)
		if err := sc.job.RunningStatus(ctx, nil, func(_ context.Context, _ jobspb.Details) (jobs.RunningStatus, error) {
			__antithesis_instrumentation__.Notify(246411)
			return RunningStatusBackfill, nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(246412)
			return errors.Wrapf(err, "failed to update running status of job %d", errors.Safe(sc.job.ID()))
		} else {
			__antithesis_instrumentation__.Notify(246413)
		}
	} else {
		__antithesis_instrumentation__.Notify(246414)
		log.Infof(ctx, "writing at persisted safe write time %v...", writeAsOf)
	}
	__antithesis_instrumentation__.Notify(246375)

	readAsOf := sc.clock.Now()

	var p *PhysicalPlan
	var evalCtx extendedEvalContext
	var planCtx *PlanningCtx

	if err := sc.txn(ctx, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(246415)

		tableDesc, err := sc.getTableVersion(ctx, txn, descriptors, version)
		if err != nil {
			__antithesis_instrumentation__.Notify(246418)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246419)
		}
		__antithesis_instrumentation__.Notify(246416)
		evalCtx = createSchemaChangeEvalCtx(ctx, sc.execCfg, txn.ReadTimestamp(), descriptors)
		planCtx = sc.distSQLPlanner.NewPlanningCtx(ctx, &evalCtx, nil,
			txn, DistributionTypeSystemTenantOnly)
		indexBatchSize := indexBackfillBatchSize.Get(&sc.execCfg.Settings.SV)
		chunkSize := sc.getChunkSize(indexBatchSize)
		spec, err := initIndexBackfillerSpec(*tableDesc.TableDesc(), writeAsOf, readAsOf, writeAtRequestTimestamp, chunkSize, addedIndexes)
		if err != nil {
			__antithesis_instrumentation__.Notify(246420)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246421)
		}
		__antithesis_instrumentation__.Notify(246417)
		p, err = sc.distSQLPlanner.createBackfillerPhysicalPlan(ctx, planCtx, spec, todoSpans)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(246422)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246423)
	}
	__antithesis_instrumentation__.Notify(246376)

	mu := struct {
		syncutil.Mutex
		updatedTodoSpans []roachpb.Span
	}{}
	var updateJobProgress func() error
	var updateJobDetails func() error
	metaFn := func(_ context.Context, meta *execinfrapb.ProducerMetadata) error {
		__antithesis_instrumentation__.Notify(246424)
		if meta.BulkProcessorProgress != nil {
			__antithesis_instrumentation__.Notify(246426)
			todoSpans = roachpb.SubtractSpans(todoSpans,
				meta.BulkProcessorProgress.CompletedSpans)
			mu.Lock()
			mu.updatedTodoSpans = make([]roachpb.Span, len(todoSpans))
			copy(mu.updatedTodoSpans, todoSpans)
			mu.Unlock()

			if sc.testingKnobs.AlwaysUpdateIndexBackfillDetails {
				__antithesis_instrumentation__.Notify(246428)
				if err := updateJobDetails(); err != nil {
					__antithesis_instrumentation__.Notify(246429)
					return err
				} else {
					__antithesis_instrumentation__.Notify(246430)
				}
			} else {
				__antithesis_instrumentation__.Notify(246431)
			}
			__antithesis_instrumentation__.Notify(246427)

			if sc.testingKnobs.AlwaysUpdateIndexBackfillProgress {
				__antithesis_instrumentation__.Notify(246432)
				if err := updateJobProgress(); err != nil {
					__antithesis_instrumentation__.Notify(246433)
					return err
				} else {
					__antithesis_instrumentation__.Notify(246434)
				}
			} else {
				__antithesis_instrumentation__.Notify(246435)
			}
		} else {
			__antithesis_instrumentation__.Notify(246436)
		}
		__antithesis_instrumentation__.Notify(246425)
		return nil
	}
	__antithesis_instrumentation__.Notify(246377)
	cbw := MetadataCallbackWriter{rowResultWriter: &errOnlyResultWriter{}, fn: metaFn}
	recv := MakeDistSQLReceiver(
		ctx,
		&cbw,
		tree.Rows,
		sc.rangeDescriptorCache,
		nil,
		sc.clock,
		evalCtx.Tracing,
		sc.execCfg.ContentionRegistry,
		nil,
	)
	defer recv.Release()

	getTodoSpansForUpdate := func() []roachpb.Span {
		__antithesis_instrumentation__.Notify(246437)
		mu.Lock()
		defer mu.Unlock()
		if mu.updatedTodoSpans == nil {
			__antithesis_instrumentation__.Notify(246439)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(246440)
		}
		__antithesis_instrumentation__.Notify(246438)
		return append(
			make([]roachpb.Span, 0, len(mu.updatedTodoSpans)),
			mu.updatedTodoSpans...,
		)
	}
	__antithesis_instrumentation__.Notify(246378)

	origNRanges := -1
	updateJobProgress = func() error {
		__antithesis_instrumentation__.Notify(246441)

		updatedTodoSpans := getTodoSpansForUpdate()
		if updatedTodoSpans == nil {
			__antithesis_instrumentation__.Notify(246445)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(246446)
		}
		__antithesis_instrumentation__.Notify(246442)
		nRanges, err := numRangesInSpans(ctx, sc.db, sc.distSQLPlanner, updatedTodoSpans)
		if err != nil {
			__antithesis_instrumentation__.Notify(246447)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246448)
		}
		__antithesis_instrumentation__.Notify(246443)
		if origNRanges == -1 {
			__antithesis_instrumentation__.Notify(246449)
			origNRanges = nRanges
		} else {
			__antithesis_instrumentation__.Notify(246450)
		}
		__antithesis_instrumentation__.Notify(246444)
		return sc.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(246451)

			if nRanges < origNRanges {
				__antithesis_instrumentation__.Notify(246453)
				fractionRangesFinished := float32(origNRanges-nRanges) / float32(origNRanges)
				fractionCompleted, err := fractionScaler.fractionCompleteFromStageFraction(stageBackfill, fractionRangesFinished)
				if err != nil {
					__antithesis_instrumentation__.Notify(246455)
					return err
				} else {
					__antithesis_instrumentation__.Notify(246456)
				}
				__antithesis_instrumentation__.Notify(246454)
				if err := sc.job.FractionProgressed(ctx, txn,
					jobs.FractionUpdater(fractionCompleted)); err != nil {
					__antithesis_instrumentation__.Notify(246457)
					return jobs.SimplifyInvalidStatusError(err)
				} else {
					__antithesis_instrumentation__.Notify(246458)
				}
			} else {
				__antithesis_instrumentation__.Notify(246459)
			}
			__antithesis_instrumentation__.Notify(246452)
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(246379)

	var updateJobMu syncutil.Mutex
	updateJobDetails = func() error {
		__antithesis_instrumentation__.Notify(246460)
		updatedTodoSpans := getTodoSpansForUpdate()
		return sc.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(246461)
			updateJobMu.Lock()
			defer updateJobMu.Unlock()

			if updatedTodoSpans == nil {
				__antithesis_instrumentation__.Notify(246463)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(246464)
			}
			__antithesis_instrumentation__.Notify(246462)
			log.VEventf(ctx, 2, "writing todo spans to job details: %+v", updatedTodoSpans)
			return rowexec.SetResumeSpansInJob(ctx, updatedTodoSpans, mutationIdx, txn, sc.job)
		})
	}
	__antithesis_instrumentation__.Notify(246380)

	stopProgress := make(chan struct{})
	duration := 10 * time.Second
	g := ctxgroup.WithContext(ctx)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(246465)
		tick := time.NewTicker(duration)
		defer tick.Stop()
		done := ctx.Done()
		for {
			__antithesis_instrumentation__.Notify(246466)
			select {
			case <-stopProgress:
				__antithesis_instrumentation__.Notify(246467)
				return nil
			case <-done:
				__antithesis_instrumentation__.Notify(246468)
				return ctx.Err()
			case <-tick.C:
				__antithesis_instrumentation__.Notify(246469)
				if err := updateJobProgress(); err != nil {
					__antithesis_instrumentation__.Notify(246470)
					return err
				} else {
					__antithesis_instrumentation__.Notify(246471)
				}
			}
		}
	})
	__antithesis_instrumentation__.Notify(246381)

	stopJobDetailsUpdate := make(chan struct{})
	detailsDuration := backfill.IndexBackfillCheckpointInterval.Get(&sc.settings.SV)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(246472)
		tick := time.NewTicker(detailsDuration)
		defer tick.Stop()
		done := ctx.Done()
		for {
			__antithesis_instrumentation__.Notify(246473)
			select {
			case <-stopJobDetailsUpdate:
				__antithesis_instrumentation__.Notify(246474)
				return nil
			case <-done:
				__antithesis_instrumentation__.Notify(246475)
				return ctx.Err()
			case <-tick.C:
				__antithesis_instrumentation__.Notify(246476)
				if err := updateJobDetails(); err != nil {
					__antithesis_instrumentation__.Notify(246477)
					return err
				} else {
					__antithesis_instrumentation__.Notify(246478)
				}
			}
		}
	})
	__antithesis_instrumentation__.Notify(246382)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(246479)
		defer close(stopProgress)
		defer close(stopJobDetailsUpdate)
		if err := sc.jobRegistry.CheckPausepoint("indexbackfill.before_flow"); err != nil {
			__antithesis_instrumentation__.Notify(246481)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246482)
		}
		__antithesis_instrumentation__.Notify(246480)

		evalCtxCopy := evalCtx
		sc.distSQLPlanner.Run(
			ctx,
			planCtx,
			nil,
			p, recv, &evalCtxCopy,
			nil,
		)()
		return cbw.Err()
	})
	__antithesis_instrumentation__.Notify(246383)

	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(246483)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246484)
	}
	__antithesis_instrumentation__.Notify(246384)

	if err := updateJobDetails(); err != nil {
		__antithesis_instrumentation__.Notify(246485)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246486)
	}
	__antithesis_instrumentation__.Notify(246385)
	if err := updateJobProgress(); err != nil {
		__antithesis_instrumentation__.Notify(246487)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246488)
	}
	__antithesis_instrumentation__.Notify(246386)

	return nil
}

func (sc *SchemaChanger) distColumnBackfill(
	ctx context.Context,
	version descpb.DescriptorVersion,
	backfillChunkSize int64,
	filter backfill.MutationFilter,
) error {
	__antithesis_instrumentation__.Notify(246489)
	duration := checkpointInterval
	if sc.testingKnobs.WriteCheckpointInterval > 0 {
		__antithesis_instrumentation__.Notify(246493)
		duration = sc.testingKnobs.WriteCheckpointInterval
	} else {
		__antithesis_instrumentation__.Notify(246494)
	}
	__antithesis_instrumentation__.Notify(246490)
	chunkSize := sc.getChunkSize(backfillChunkSize)

	origNRanges := -1
	origFractionCompleted := sc.job.FractionCompleted()
	fractionLeft := 1 - origFractionCompleted
	readAsOf := sc.clock.Now()

	var todoSpans []roachpb.Span
	var mutationIdx int
	if err := DescsTxn(ctx, sc.execCfg, func(ctx context.Context, txn *kv.Txn, col *descs.Collection) (err error) {
		__antithesis_instrumentation__.Notify(246495)
		todoSpans, _, mutationIdx, err = rowexec.GetResumeSpans(
			ctx, sc.jobRegistry, txn, sc.execCfg.Codec, col, sc.descID, sc.mutationID, filter)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(246496)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246497)
	}
	__antithesis_instrumentation__.Notify(246491)

	for len(todoSpans) > 0 {
		__antithesis_instrumentation__.Notify(246498)
		log.VEventf(ctx, 2, "backfill: process %+v spans", todoSpans)

		var updatedTodoSpans []roachpb.Span
		if err := sc.txn(ctx, func(
			ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
		) error {
			__antithesis_instrumentation__.Notify(246500)
			updatedTodoSpans = todoSpans

			nRanges, err := numRangesInSpans(ctx, sc.db, sc.distSQLPlanner, todoSpans)
			if err != nil {
				__antithesis_instrumentation__.Notify(246508)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246509)
			}
			__antithesis_instrumentation__.Notify(246501)
			if origNRanges == -1 {
				__antithesis_instrumentation__.Notify(246510)
				origNRanges = nRanges
			} else {
				__antithesis_instrumentation__.Notify(246511)
			}
			__antithesis_instrumentation__.Notify(246502)

			if nRanges < origNRanges {
				__antithesis_instrumentation__.Notify(246512)
				fractionRangesFinished := float32(origNRanges-nRanges) / float32(origNRanges)
				fractionCompleted := origFractionCompleted + fractionLeft*fractionRangesFinished
				if err := sc.job.FractionProgressed(ctx, txn, jobs.FractionUpdater(fractionCompleted)); err != nil {
					__antithesis_instrumentation__.Notify(246513)
					return jobs.SimplifyInvalidStatusError(err)
				} else {
					__antithesis_instrumentation__.Notify(246514)
				}
			} else {
				__antithesis_instrumentation__.Notify(246515)
			}
			__antithesis_instrumentation__.Notify(246503)

			tableDesc, err := sc.getTableVersion(ctx, txn, descriptors, version)
			if err != nil {
				__antithesis_instrumentation__.Notify(246516)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246517)
			}
			__antithesis_instrumentation__.Notify(246504)
			metaFn := func(_ context.Context, meta *execinfrapb.ProducerMetadata) error {
				__antithesis_instrumentation__.Notify(246518)
				if meta.BulkProcessorProgress != nil {
					__antithesis_instrumentation__.Notify(246520)
					updatedTodoSpans = roachpb.SubtractSpans(updatedTodoSpans,
						meta.BulkProcessorProgress.CompletedSpans)
				} else {
					__antithesis_instrumentation__.Notify(246521)
				}
				__antithesis_instrumentation__.Notify(246519)
				return nil
			}
			__antithesis_instrumentation__.Notify(246505)
			cbw := MetadataCallbackWriter{rowResultWriter: &errOnlyResultWriter{}, fn: metaFn}
			evalCtx := createSchemaChangeEvalCtx(ctx, sc.execCfg, txn.ReadTimestamp(), descriptors)
			recv := MakeDistSQLReceiver(
				ctx,
				&cbw,
				tree.Rows,
				sc.rangeDescriptorCache,
				nil,
				sc.clock,
				evalCtx.Tracing,
				sc.execCfg.ContentionRegistry,
				nil,
			)
			defer recv.Release()

			planCtx := sc.distSQLPlanner.NewPlanningCtx(ctx, &evalCtx, nil, txn,
				DistributionTypeSystemTenantOnly)
			spec, err := initColumnBackfillerSpec(*tableDesc.TableDesc(), duration, chunkSize, readAsOf)
			if err != nil {
				__antithesis_instrumentation__.Notify(246522)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246523)
			}
			__antithesis_instrumentation__.Notify(246506)
			plan, err := sc.distSQLPlanner.createBackfillerPhysicalPlan(ctx, planCtx, spec, todoSpans)
			if err != nil {
				__antithesis_instrumentation__.Notify(246524)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246525)
			}
			__antithesis_instrumentation__.Notify(246507)
			sc.distSQLPlanner.Run(
				ctx,
				planCtx,
				nil,
				plan, recv, &evalCtx,
				nil,
			)()
			return cbw.Err()
		}); err != nil {
			__antithesis_instrumentation__.Notify(246526)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246527)
		}
		__antithesis_instrumentation__.Notify(246499)
		todoSpans = updatedTodoSpans

		if err := sc.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(246528)
			return rowexec.SetResumeSpansInJob(ctx, todoSpans, mutationIdx, txn, sc.job)
		}); err != nil {
			__antithesis_instrumentation__.Notify(246529)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246530)
		}
	}
	__antithesis_instrumentation__.Notify(246492)
	return nil
}

func (sc *SchemaChanger) updateJobRunningStatus(
	ctx context.Context, status jobs.RunningStatus,
) (tableDesc catalog.TableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(246531)
	err = DescsTxn(ctx, sc.execCfg, func(ctx context.Context, txn *kv.Txn, col *descs.Collection) (err error) {
		__antithesis_instrumentation__.Notify(246533)

		tableDesc, err = col.Direct().MustGetTableDescByID(ctx, txn, sc.descID)
		if err != nil {
			__antithesis_instrumentation__.Notify(246537)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246538)
		}
		__antithesis_instrumentation__.Notify(246534)

		updateJobRunningProgress := false
		for _, mutation := range tableDesc.AllMutations() {
			__antithesis_instrumentation__.Notify(246539)
			if mutation.MutationID() != sc.mutationID {
				__antithesis_instrumentation__.Notify(246541)

				break
			} else {
				__antithesis_instrumentation__.Notify(246542)
			}
			__antithesis_instrumentation__.Notify(246540)

			if mutation.Adding() && func() bool {
				__antithesis_instrumentation__.Notify(246543)
				return mutation.WriteAndDeleteOnly() == true
			}() == true {
				__antithesis_instrumentation__.Notify(246544)
				updateJobRunningProgress = true
			} else {
				__antithesis_instrumentation__.Notify(246545)
				if mutation.Dropped() && func() bool {
					__antithesis_instrumentation__.Notify(246546)
					return mutation.DeleteOnly() == true
				}() == true {
					__antithesis_instrumentation__.Notify(246547)
					updateJobRunningProgress = true
				} else {
					__antithesis_instrumentation__.Notify(246548)
				}
			}
		}
		__antithesis_instrumentation__.Notify(246535)
		if updateJobRunningProgress && func() bool {
			__antithesis_instrumentation__.Notify(246549)
			return !tableDesc.Dropped() == true
		}() == true {
			__antithesis_instrumentation__.Notify(246550)
			if err := sc.job.RunningStatus(ctx, txn, func(
				ctx context.Context, details jobspb.Details) (jobs.RunningStatus, error) {
				__antithesis_instrumentation__.Notify(246551)
				return status, nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(246552)
				return errors.Wrapf(err, "failed to update running status of job %d", errors.Safe(sc.job.ID()))
			} else {
				__antithesis_instrumentation__.Notify(246553)
			}
		} else {
			__antithesis_instrumentation__.Notify(246554)
		}
		__antithesis_instrumentation__.Notify(246536)
		return nil
	})
	__antithesis_instrumentation__.Notify(246532)
	return tableDesc, err
}

func (sc *SchemaChanger) validateIndexes(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(246555)
	if lease.TestingTableLeasesAreDisabled() {
		__antithesis_instrumentation__.Notify(246565)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(246566)
	}
	__antithesis_instrumentation__.Notify(246556)
	log.Info(ctx, "validating new indexes")

	_, err := sc.updateJobRunningStatus(ctx, RunningStatusValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(246567)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246568)
	}
	__antithesis_instrumentation__.Notify(246557)

	if fn := sc.testingKnobs.RunBeforeIndexValidation; fn != nil {
		__antithesis_instrumentation__.Notify(246569)
		if err := fn(); err != nil {
			__antithesis_instrumentation__.Notify(246570)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246571)
		}
	} else {
		__antithesis_instrumentation__.Notify(246572)
	}
	__antithesis_instrumentation__.Notify(246558)

	readAsOf := sc.clock.Now()
	var tableDesc catalog.TableDescriptor
	if err := sc.fixedTimestampTxn(ctx, readAsOf, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) (err error) {
		__antithesis_instrumentation__.Notify(246573)
		flags := tree.ObjectLookupFlagsWithRequired()
		flags.AvoidLeased = true
		tableDesc, err = descriptors.GetImmutableTableByID(ctx, txn, sc.descID, flags)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(246574)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246575)
	}
	__antithesis_instrumentation__.Notify(246559)

	var forwardIndexes, invertedIndexes []catalog.Index

	for _, m := range tableDesc.AllMutations() {
		__antithesis_instrumentation__.Notify(246576)
		if sc.mutationID != m.MutationID() {
			__antithesis_instrumentation__.Notify(246579)
			break
		} else {
			__antithesis_instrumentation__.Notify(246580)
		}
		__antithesis_instrumentation__.Notify(246577)
		idx := m.AsIndex()

		if idx == nil || func() bool {
			__antithesis_instrumentation__.Notify(246581)
			return idx.Dropped() == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(246582)
			return idx.IsTemporaryIndexForBackfill() == true
		}() == true {
			__antithesis_instrumentation__.Notify(246583)
			continue
		} else {
			__antithesis_instrumentation__.Notify(246584)
		}
		__antithesis_instrumentation__.Notify(246578)
		switch idx.GetType() {
		case descpb.IndexDescriptor_FORWARD:
			__antithesis_instrumentation__.Notify(246585)
			forwardIndexes = append(forwardIndexes, idx)
		case descpb.IndexDescriptor_INVERTED:
			__antithesis_instrumentation__.Notify(246586)
			invertedIndexes = append(invertedIndexes, idx)
		default:
			__antithesis_instrumentation__.Notify(246587)
		}
	}
	__antithesis_instrumentation__.Notify(246560)
	if len(forwardIndexes) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(246588)
		return len(invertedIndexes) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(246589)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(246590)
	}
	__antithesis_instrumentation__.Notify(246561)

	grp := ctxgroup.WithContext(ctx)
	runHistoricalTxn := sc.makeFixedTimestampInternalExecRunner(readAsOf)

	if len(forwardIndexes) > 0 {
		__antithesis_instrumentation__.Notify(246591)
		grp.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(246592)
			return ValidateForwardIndexes(
				ctx,
				tableDesc,
				forwardIndexes,
				runHistoricalTxn,
				true,
				false,
				sessiondata.InternalExecutorOverride{},
			)
		})
	} else {
		__antithesis_instrumentation__.Notify(246593)
	}
	__antithesis_instrumentation__.Notify(246562)
	if len(invertedIndexes) > 0 {
		__antithesis_instrumentation__.Notify(246594)
		grp.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(246595)
			return ValidateInvertedIndexes(
				ctx,
				sc.execCfg.Codec,
				tableDesc,
				invertedIndexes,
				runHistoricalTxn,
				true,
				false,
				sessiondata.InternalExecutorOverride{},
			)
		})
	} else {
		__antithesis_instrumentation__.Notify(246596)
	}
	__antithesis_instrumentation__.Notify(246563)
	if err := grp.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(246597)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246598)
	}
	__antithesis_instrumentation__.Notify(246564)
	log.Info(ctx, "finished validating new indexes")
	return nil
}

type InvalidIndexesError struct {
	Indexes []descpb.IndexID
}

func (e InvalidIndexesError) Error() string {
	__antithesis_instrumentation__.Notify(246599)
	return fmt.Sprintf("found %d invalid indexes", len(e.Indexes))
}

func ValidateInvertedIndexes(
	ctx context.Context,
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	indexes []catalog.Index,
	runHistoricalTxn sqlutil.HistoricalInternalExecTxnRunner,
	withFirstMutationPublic bool,
	gatherAllInvalid bool,
	execOverride sessiondata.InternalExecutorOverride,
) error {
	__antithesis_instrumentation__.Notify(246600)
	grp := ctxgroup.WithContext(ctx)
	invalid := make(chan descpb.IndexID, len(indexes))

	expectedCount := make([]int64, len(indexes))
	countReady := make([]chan struct{}, len(indexes))

	for i, idx := range indexes {
		__antithesis_instrumentation__.Notify(246605)

		i, idx := i, idx
		countReady[i] = make(chan struct{})

		grp.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(246607)

			start := timeutil.Now()
			var idxLen int64
			span := tableDesc.IndexSpan(codec, idx.GetID())
			key := span.Key
			endKey := span.EndKey
			if err := runHistoricalTxn(ctx, func(ctx context.Context, txn *kv.Txn, _ sqlutil.InternalExecutor) error {
				__antithesis_instrumentation__.Notify(246610)
				for {
					__antithesis_instrumentation__.Notify(246612)
					kvs, err := txn.Scan(ctx, key, endKey, 1000000)
					if err != nil {
						__antithesis_instrumentation__.Notify(246615)
						return err
					} else {
						__antithesis_instrumentation__.Notify(246616)
					}
					__antithesis_instrumentation__.Notify(246613)
					if len(kvs) == 0 {
						__antithesis_instrumentation__.Notify(246617)
						break
					} else {
						__antithesis_instrumentation__.Notify(246618)
					}
					__antithesis_instrumentation__.Notify(246614)
					idxLen += int64(len(kvs))
					key = kvs[len(kvs)-1].Key.PrefixEnd()
				}
				__antithesis_instrumentation__.Notify(246611)
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(246619)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246620)
			}
			__antithesis_instrumentation__.Notify(246608)
			log.Infof(ctx, "inverted index %s/%s count = %d, took %s",
				tableDesc.GetName(), idx.GetName(), idxLen, timeutil.Since(start))
			select {
			case <-countReady[i]:
				__antithesis_instrumentation__.Notify(246621)
				if idxLen != expectedCount[i] {
					__antithesis_instrumentation__.Notify(246623)
					if gatherAllInvalid {
						__antithesis_instrumentation__.Notify(246625)
						invalid <- idx.GetID()
						return nil
					} else {
						__antithesis_instrumentation__.Notify(246626)
					}
					__antithesis_instrumentation__.Notify(246624)

					return errors.AssertionFailedf(
						"validation of index %s failed: expected %d rows, found %d",
						idx.GetName(), errors.Safe(expectedCount[i]), errors.Safe(idxLen))
				} else {
					__antithesis_instrumentation__.Notify(246627)
				}
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(246622)
				return ctx.Err()
			}
			__antithesis_instrumentation__.Notify(246609)
			return nil
		})
		__antithesis_instrumentation__.Notify(246606)

		grp.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(246628)
			c, err := countExpectedRowsForInvertedIndex(ctx, tableDesc, idx, runHistoricalTxn, withFirstMutationPublic, execOverride)
			if err != nil {
				__antithesis_instrumentation__.Notify(246630)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246631)
			}
			__antithesis_instrumentation__.Notify(246629)
			expectedCount[i] = c
			close(countReady[i])
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(246601)

	if err := grp.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(246632)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246633)
	}
	__antithesis_instrumentation__.Notify(246602)
	close(invalid)
	invalidErr := InvalidIndexesError{}
	for i := range invalid {
		__antithesis_instrumentation__.Notify(246634)
		invalidErr.Indexes = append(invalidErr.Indexes, i)
	}
	__antithesis_instrumentation__.Notify(246603)
	if len(invalidErr.Indexes) > 0 {
		__antithesis_instrumentation__.Notify(246635)
		return invalidErr
	} else {
		__antithesis_instrumentation__.Notify(246636)
	}
	__antithesis_instrumentation__.Notify(246604)
	return nil
}

func countExpectedRowsForInvertedIndex(
	ctx context.Context,
	tableDesc catalog.TableDescriptor,
	idx catalog.Index,
	runHistoricalTxn sqlutil.HistoricalInternalExecTxnRunner,
	withFirstMutationPublic bool,
	execOverride sessiondata.InternalExecutorOverride,
) (int64, error) {
	__antithesis_instrumentation__.Notify(246637)
	desc := tableDesc
	start := timeutil.Now()
	if withFirstMutationPublic {
		__antithesis_instrumentation__.Notify(246642)

		fakeDesc, err := tableDesc.MakeFirstMutationPublic(catalog.IgnoreConstraints)
		if err != nil {
			__antithesis_instrumentation__.Notify(246644)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(246645)
		}
		__antithesis_instrumentation__.Notify(246643)
		desc = fakeDesc
	} else {
		__antithesis_instrumentation__.Notify(246646)
	}
	__antithesis_instrumentation__.Notify(246638)

	colID := idx.InvertedColumnID()
	col, err := desc.FindColumnWithID(colID)
	if err != nil {
		__antithesis_instrumentation__.Notify(246647)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(246648)
	}
	__antithesis_instrumentation__.Notify(246639)

	var colNameOrExpr string
	if col.IsExpressionIndexColumn() {
		__antithesis_instrumentation__.Notify(246649)
		colNameOrExpr = col.GetComputeExpr()
	} else {
		__antithesis_instrumentation__.Notify(246650)

		colNameOrExpr = fmt.Sprintf("%q", col.ColName())
	}
	__antithesis_instrumentation__.Notify(246640)

	var expectedCount int64
	if err := runHistoricalTxn(ctx, func(ctx context.Context, txn *kv.Txn, ie sqlutil.InternalExecutor) error {
		__antithesis_instrumentation__.Notify(246651)
		var stmt string
		geoConfig := idx.GetGeoConfig()
		if geoindex.IsEmptyConfig(&geoConfig) {
			__antithesis_instrumentation__.Notify(246654)
			stmt = fmt.Sprintf(
				`SELECT coalesce(sum_int(crdb_internal.num_inverted_index_entries(%s, %d)), 0) FROM [%d AS t]`,
				colNameOrExpr, idx.GetVersion(), desc.GetID(),
			)
		} else {
			__antithesis_instrumentation__.Notify(246655)
			stmt = fmt.Sprintf(
				`SELECT coalesce(sum_int(crdb_internal.num_geo_inverted_index_entries(%d, %d, %s)), 0) FROM [%d AS t]`,
				desc.GetID(), idx.GetID(), colNameOrExpr, desc.GetID(),
			)
		}
		__antithesis_instrumentation__.Notify(246652)

		if idx.IsPartial() {
			__antithesis_instrumentation__.Notify(246656)
			stmt = fmt.Sprintf(`%s WHERE %s`, stmt, idx.GetPredicate())
		} else {
			__antithesis_instrumentation__.Notify(246657)
		}
		__antithesis_instrumentation__.Notify(246653)
		return ie.WithSyntheticDescriptors([]catalog.Descriptor{desc}, func() error {
			__antithesis_instrumentation__.Notify(246658)
			row, err := ie.QueryRowEx(ctx, "verify-inverted-idx-count", txn, execOverride, stmt)
			if err != nil {
				__antithesis_instrumentation__.Notify(246661)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246662)
			}
			__antithesis_instrumentation__.Notify(246659)
			if row == nil {
				__antithesis_instrumentation__.Notify(246663)
				return errors.New("failed to verify inverted index count")
			} else {
				__antithesis_instrumentation__.Notify(246664)
			}
			__antithesis_instrumentation__.Notify(246660)
			expectedCount = int64(tree.MustBeDInt(row[0]))
			return nil
		})
	}); err != nil {
		__antithesis_instrumentation__.Notify(246665)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(246666)
	}
	__antithesis_instrumentation__.Notify(246641)
	log.Infof(ctx, "%s %s expected inverted index count = %d, took %s",
		desc.GetName(), colNameOrExpr, expectedCount, timeutil.Since(start))
	return expectedCount, nil

}

func ValidateForwardIndexes(
	ctx context.Context,
	tableDesc catalog.TableDescriptor,
	indexes []catalog.Index,
	runHistoricalTxn sqlutil.HistoricalInternalExecTxnRunner,
	withFirstMutationPublic bool,
	gatherAllInvalid bool,
	execOverride sessiondata.InternalExecutorOverride,
) error {
	__antithesis_instrumentation__.Notify(246667)
	grp := ctxgroup.WithContext(ctx)

	invalid := make(chan descpb.IndexID, len(indexes))
	var tableRowCount int64
	partialIndexExpectedCounts := make(map[descpb.IndexID]int64, len(indexes))

	tableCountsReady := make(chan struct{})

	for _, idx := range indexes {
		__antithesis_instrumentation__.Notify(246673)

		idx := idx

		grp.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(246674)
			start := timeutil.Now()
			idxLen, err := countIndexRowsAndMaybeCheckUniqueness(ctx, tableDesc, idx, withFirstMutationPublic, runHistoricalTxn, execOverride)
			if err != nil {
				__antithesis_instrumentation__.Notify(246677)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246678)
			}
			__antithesis_instrumentation__.Notify(246675)
			log.Infof(ctx, "validation: index %s/%s row count = %d, time so far %s",
				tableDesc.GetName(), idx.GetName(), idxLen, timeutil.Since(start))

			select {
			case <-tableCountsReady:
				__antithesis_instrumentation__.Notify(246679)
				expectedCount := tableRowCount

				if idx.IsPartial() {
					__antithesis_instrumentation__.Notify(246682)
					expectedCount = partialIndexExpectedCounts[idx.GetID()]
				} else {
					__antithesis_instrumentation__.Notify(246683)
				}
				__antithesis_instrumentation__.Notify(246680)

				if idxLen != expectedCount {
					__antithesis_instrumentation__.Notify(246684)
					if gatherAllInvalid {
						__antithesis_instrumentation__.Notify(246686)
						invalid <- idx.GetID()
						return nil
					} else {
						__antithesis_instrumentation__.Notify(246687)
					}
					__antithesis_instrumentation__.Notify(246685)

					return pgerror.WithConstraintName(pgerror.Newf(pgcode.UniqueViolation,
						"duplicate key value violates unique constraint %q",
						idx.GetName()),
						idx.GetName())

				} else {
					__antithesis_instrumentation__.Notify(246688)
				}
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(246681)
				return ctx.Err()
			}
			__antithesis_instrumentation__.Notify(246676)

			return nil
		})
	}
	__antithesis_instrumentation__.Notify(246668)

	grp.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(246689)
		start := timeutil.Now()
		c, err := populateExpectedCounts(ctx, tableDesc, indexes, partialIndexExpectedCounts, withFirstMutationPublic, runHistoricalTxn, execOverride)
		if err != nil {
			__antithesis_instrumentation__.Notify(246691)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246692)
		}
		__antithesis_instrumentation__.Notify(246690)
		log.Infof(ctx, "validation: table %s row count = %d, took %s",
			tableDesc.GetName(), c, timeutil.Since(start))
		tableRowCount = c
		defer close(tableCountsReady)
		return nil
	})
	__antithesis_instrumentation__.Notify(246669)

	if err := grp.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(246693)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246694)
	}
	__antithesis_instrumentation__.Notify(246670)
	close(invalid)
	invalidErr := InvalidIndexesError{}
	for i := range invalid {
		__antithesis_instrumentation__.Notify(246695)
		invalidErr.Indexes = append(invalidErr.Indexes, i)
	}
	__antithesis_instrumentation__.Notify(246671)
	if len(invalidErr.Indexes) > 0 {
		__antithesis_instrumentation__.Notify(246696)
		return invalidErr
	} else {
		__antithesis_instrumentation__.Notify(246697)
	}
	__antithesis_instrumentation__.Notify(246672)
	return nil
}

func populateExpectedCounts(
	ctx context.Context,
	tableDesc catalog.TableDescriptor,
	indexes []catalog.Index,
	partialIndexExpectedCounts map[descpb.IndexID]int64,
	withFirstMutationPublic bool,
	runHistoricalTxn sqlutil.HistoricalInternalExecTxnRunner,
	execOverride sessiondata.InternalExecutorOverride,
) (int64, error) {
	__antithesis_instrumentation__.Notify(246698)
	desc := tableDesc
	if withFirstMutationPublic {
		__antithesis_instrumentation__.Notify(246701)

		fakeDesc, err := tableDesc.MakeFirstMutationPublic(catalog.IgnoreConstraintsAndPKSwaps)
		if err != nil {
			__antithesis_instrumentation__.Notify(246703)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(246704)
		}
		__antithesis_instrumentation__.Notify(246702)
		desc = fakeDesc
	} else {
		__antithesis_instrumentation__.Notify(246705)
	}
	__antithesis_instrumentation__.Notify(246699)
	var tableRowCount int64
	if err := runHistoricalTxn(ctx, func(ctx context.Context, txn *kv.Txn, ie sqlutil.InternalExecutor) error {
		__antithesis_instrumentation__.Notify(246706)
		var s strings.Builder
		for _, idx := range indexes {
			__antithesis_instrumentation__.Notify(246708)

			if idx.IsPartial() {
				__antithesis_instrumentation__.Notify(246709)
				s.WriteString(fmt.Sprintf(`, count(1) FILTER (WHERE %s)`, idx.GetPredicate()))
			} else {
				__antithesis_instrumentation__.Notify(246710)
			}
		}
		__antithesis_instrumentation__.Notify(246707)
		partialIndexCounts := s.String()

		query := fmt.Sprintf(`SELECT count(1)%s FROM [%d AS t]@[%d]`, partialIndexCounts, desc.GetID(), desc.GetPrimaryIndexID())

		return ie.WithSyntheticDescriptors([]catalog.Descriptor{desc}, func() error {
			__antithesis_instrumentation__.Notify(246711)
			cnt, err := ie.QueryRowEx(ctx, "VERIFY INDEX", txn, execOverride, query)
			if err != nil {
				__antithesis_instrumentation__.Notify(246715)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246716)
			}
			__antithesis_instrumentation__.Notify(246712)
			if cnt == nil {
				__antithesis_instrumentation__.Notify(246717)
				return errors.New("failed to verify index")
			} else {
				__antithesis_instrumentation__.Notify(246718)
			}
			__antithesis_instrumentation__.Notify(246713)

			tableRowCount = int64(tree.MustBeDInt(cnt[0]))
			cntIdx := 1
			for _, idx := range indexes {
				__antithesis_instrumentation__.Notify(246719)
				if idx.IsPartial() {
					__antithesis_instrumentation__.Notify(246720)
					partialIndexExpectedCounts[idx.GetID()] = int64(tree.MustBeDInt(cnt[cntIdx]))
					cntIdx++
				} else {
					__antithesis_instrumentation__.Notify(246721)
				}
			}
			__antithesis_instrumentation__.Notify(246714)

			return nil
		})
	}); err != nil {
		__antithesis_instrumentation__.Notify(246722)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(246723)
	}
	__antithesis_instrumentation__.Notify(246700)
	return tableRowCount, nil
}

func countIndexRowsAndMaybeCheckUniqueness(
	ctx context.Context,
	tableDesc catalog.TableDescriptor,
	idx catalog.Index,
	withFirstMutationPublic bool,
	runHistoricalTxn sqlutil.HistoricalInternalExecTxnRunner,
	execOverride sessiondata.InternalExecutorOverride,
) (int64, error) {
	__antithesis_instrumentation__.Notify(246724)

	skipUniquenessChecks := false
	if withFirstMutationPublic {
		__antithesis_instrumentation__.Notify(246728)
		mutations := tableDesc.AllMutations()
		if len(mutations) > 0 {
			__antithesis_instrumentation__.Notify(246729)
			mutationID := mutations[0].MutationID()
			for _, mut := range tableDesc.AllMutations() {
				__antithesis_instrumentation__.Notify(246730)

				if mut.MutationID() != mutationID {
					__antithesis_instrumentation__.Notify(246732)
					break
				} else {
					__antithesis_instrumentation__.Notify(246733)
				}
				__antithesis_instrumentation__.Notify(246731)
				if pkSwap := mut.AsPrimaryKeySwap(); pkSwap != nil {
					__antithesis_instrumentation__.Notify(246734)
					if lcSwap := pkSwap.PrimaryKeySwapDesc().LocalityConfigSwap; lcSwap != nil {
						__antithesis_instrumentation__.Notify(246735)
						if lcSwap.OldLocalityConfig.GetRegionalByRow() != nil || func() bool {
							__antithesis_instrumentation__.Notify(246736)
							return lcSwap.NewLocalityConfig.GetRegionalByRow() != nil == true
						}() == true {
							__antithesis_instrumentation__.Notify(246737)
							skipUniquenessChecks = true
							break
						} else {
							__antithesis_instrumentation__.Notify(246738)
						}
					} else {
						__antithesis_instrumentation__.Notify(246739)
					}
				} else {
					__antithesis_instrumentation__.Notify(246740)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(246741)
		}
	} else {
		__antithesis_instrumentation__.Notify(246742)
	}
	__antithesis_instrumentation__.Notify(246725)

	desc := tableDesc
	if withFirstMutationPublic {
		__antithesis_instrumentation__.Notify(246743)

		fakeDesc, err := tableDesc.MakeFirstMutationPublic(catalog.IgnoreConstraints)
		if err != nil {
			__antithesis_instrumentation__.Notify(246745)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(246746)
		}
		__antithesis_instrumentation__.Notify(246744)
		desc = fakeDesc
	} else {
		__antithesis_instrumentation__.Notify(246747)
	}
	__antithesis_instrumentation__.Notify(246726)

	var idxLen int64
	if err := runHistoricalTxn(ctx, func(ctx context.Context, txn *kv.Txn, ie sqlutil.InternalExecutor) error {
		__antithesis_instrumentation__.Notify(246748)
		query := fmt.Sprintf(`SELECT count(1) FROM [%d AS t]@[%d]`, desc.GetID(), idx.GetID())

		if idx.IsPartial() {
			__antithesis_instrumentation__.Notify(246750)
			query = fmt.Sprintf(`%s WHERE %s`, query, idx.GetPredicate())
		} else {
			__antithesis_instrumentation__.Notify(246751)
		}
		__antithesis_instrumentation__.Notify(246749)
		return ie.WithSyntheticDescriptors([]catalog.Descriptor{desc}, func() error {
			__antithesis_instrumentation__.Notify(246752)
			row, err := ie.QueryRowEx(ctx, "verify-idx-count", txn, execOverride, query)
			if err != nil {
				__antithesis_instrumentation__.Notify(246756)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246757)
			}
			__antithesis_instrumentation__.Notify(246753)
			if row == nil {
				__antithesis_instrumentation__.Notify(246758)
				return errors.New("failed to verify index count")
			} else {
				__antithesis_instrumentation__.Notify(246759)
			}
			__antithesis_instrumentation__.Notify(246754)
			idxLen = int64(tree.MustBeDInt(row[0]))

			if idx.IsUnique() && func() bool {
				__antithesis_instrumentation__.Notify(246760)
				return idx.GetPartitioning().NumImplicitColumns() > 0 == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(246761)
				return !skipUniquenessChecks == true
			}() == true {
				__antithesis_instrumentation__.Notify(246762)
				if err := validateUniqueConstraint(
					ctx,
					tableDesc,
					idx.GetName(),
					idx.IndexDesc().KeyColumnIDs[idx.GetPartitioning().NumImplicitColumns():],
					idx.GetPredicate(),
					ie,
					txn,
					false,
				); err != nil {
					__antithesis_instrumentation__.Notify(246763)
					return err
				} else {
					__antithesis_instrumentation__.Notify(246764)
				}
			} else {
				__antithesis_instrumentation__.Notify(246765)
			}
			__antithesis_instrumentation__.Notify(246755)
			return nil
		})
	}); err != nil {
		__antithesis_instrumentation__.Notify(246766)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(246767)
	}
	__antithesis_instrumentation__.Notify(246727)
	return idxLen, nil
}

func (sc *SchemaChanger) backfillIndexes(
	ctx context.Context,
	version descpb.DescriptorVersion,
	addingSpans []roachpb.Span,
	addedIndexes []descpb.IndexID,
	temporaryIndexes []descpb.IndexID,
) error {
	__antithesis_instrumentation__.Notify(246768)

	writeAtRequestTimestamp := len(temporaryIndexes) != 0
	log.Infof(ctx, "backfilling %d indexes: %v (writeAtRequestTimestamp: %v)", len(addingSpans), addingSpans, writeAtRequestTimestamp)

	if sc.execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(246775)
		expirationTime := sc.db.Clock().Now().Add(time.Hour.Nanoseconds(), 0)
		for _, span := range addingSpans {
			__antithesis_instrumentation__.Notify(246776)
			if err := sc.db.AdminSplit(ctx, span.Key, expirationTime); err != nil {
				__antithesis_instrumentation__.Notify(246777)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246778)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(246779)
	}
	__antithesis_instrumentation__.Notify(246769)

	if fn := sc.testingKnobs.RunBeforeIndexBackfill; fn != nil {
		__antithesis_instrumentation__.Notify(246780)
		fn()
	} else {
		__antithesis_instrumentation__.Notify(246781)
	}
	__antithesis_instrumentation__.Notify(246770)

	fractionScaler := &multiStageFractionScaler{initial: sc.job.FractionCompleted(), stages: backfillStageFractions}
	if writeAtRequestTimestamp {
		__antithesis_instrumentation__.Notify(246782)
		fractionScaler.stages = mvccCompatibleBackfillStageFractions
	} else {
		__antithesis_instrumentation__.Notify(246783)
	}
	__antithesis_instrumentation__.Notify(246771)

	if err := sc.distIndexBackfill(
		ctx, version, addingSpans, addedIndexes, writeAtRequestTimestamp, backfill.IndexMutationFilter, fractionScaler,
	); err != nil {
		__antithesis_instrumentation__.Notify(246784)
		if errors.HasType(err, &roachpb.InsufficientSpaceError{}) {
			__antithesis_instrumentation__.Notify(246786)
			return jobs.MarkPauseRequestError(errors.UnwrapAll(err))
		} else {
			__antithesis_instrumentation__.Notify(246787)
		}
		__antithesis_instrumentation__.Notify(246785)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246788)
	}
	__antithesis_instrumentation__.Notify(246772)

	if writeAtRequestTimestamp {
		__antithesis_instrumentation__.Notify(246789)
		if fn := sc.testingKnobs.RunBeforeTempIndexMerge; fn != nil {
			__antithesis_instrumentation__.Notify(246794)
			fn()
		} else {
			__antithesis_instrumentation__.Notify(246795)
		}
		__antithesis_instrumentation__.Notify(246790)

		if err := sc.RunStateMachineAfterIndexBackfill(ctx); err != nil {
			__antithesis_instrumentation__.Notify(246796)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246797)
		}
		__antithesis_instrumentation__.Notify(246791)

		if err := sc.mergeFromTemporaryIndex(ctx, addedIndexes, temporaryIndexes, fractionScaler); err != nil {
			__antithesis_instrumentation__.Notify(246798)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246799)
		}
		__antithesis_instrumentation__.Notify(246792)

		if fn := sc.testingKnobs.RunAfterTempIndexMerge; fn != nil {
			__antithesis_instrumentation__.Notify(246800)
			fn()
		} else {
			__antithesis_instrumentation__.Notify(246801)
		}
		__antithesis_instrumentation__.Notify(246793)

		if err := sc.runStateMachineAfterTempIndexMerge(ctx); err != nil {
			__antithesis_instrumentation__.Notify(246802)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246803)
		}
	} else {
		__antithesis_instrumentation__.Notify(246804)
	}
	__antithesis_instrumentation__.Notify(246773)

	if fn := sc.testingKnobs.RunAfterIndexBackfill; fn != nil {
		__antithesis_instrumentation__.Notify(246805)
		fn()
	} else {
		__antithesis_instrumentation__.Notify(246806)
	}
	__antithesis_instrumentation__.Notify(246774)

	log.Info(ctx, "finished backfilling indexes")
	return sc.validateIndexes(ctx)
}

func (sc *SchemaChanger) mergeFromTemporaryIndex(
	ctx context.Context,
	addingIndexes []descpb.IndexID,
	temporaryIndexes []descpb.IndexID,
	fractionScaler *multiStageFractionScaler,
) error {
	__antithesis_instrumentation__.Notify(246807)
	var tbl *tabledesc.Mutable
	if err := sc.txn(ctx, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(246810)
		var err error
		tbl, err = descsCol.GetMutableTableVersionByID(ctx, sc.descID, txn)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(246811)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246812)
	}
	__antithesis_instrumentation__.Notify(246808)
	clusterVersion := tbl.ClusterVersion()
	tableDesc := tabledesc.NewBuilder(&clusterVersion).BuildImmutableTable()
	if err := sc.distIndexMerge(ctx, tableDesc, addingIndexes, temporaryIndexes, fractionScaler); err != nil {
		__antithesis_instrumentation__.Notify(246813)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246814)
	}
	__antithesis_instrumentation__.Notify(246809)
	return nil
}

func (sc *SchemaChanger) runStateMachineAfterTempIndexMerge(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(246815)
	var runStatus jobs.RunningStatus
	return sc.txn(ctx, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(246816)
		tbl, err := descsCol.GetMutableTableVersionByID(ctx, sc.descID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(246822)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246823)
		}
		__antithesis_instrumentation__.Notify(246817)
		runStatus = ""

		for _, m := range tbl.AllMutations() {
			__antithesis_instrumentation__.Notify(246824)
			if m.MutationID() != sc.mutationID {
				__antithesis_instrumentation__.Notify(246827)

				break
			} else {
				__antithesis_instrumentation__.Notify(246828)
			}
			__antithesis_instrumentation__.Notify(246825)
			idx := m.AsIndex()
			if idx == nil {
				__antithesis_instrumentation__.Notify(246829)

				continue
			} else {
				__antithesis_instrumentation__.Notify(246830)
			}
			__antithesis_instrumentation__.Notify(246826)

			if idx.IsTemporaryIndexForBackfill() && func() bool {
				__antithesis_instrumentation__.Notify(246831)
				return m.Adding() == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(246832)
				return m.WriteAndDeleteOnly() == true
			}() == true {
				__antithesis_instrumentation__.Notify(246833)
				log.Infof(ctx, "dropping temporary index: %d", idx.IndexDesc().ID)
				tbl.Mutations[m.MutationOrdinal()].State = descpb.DescriptorMutation_DELETE_ONLY
				tbl.Mutations[m.MutationOrdinal()].Direction = descpb.DescriptorMutation_DROP
				runStatus = RunningStatusDeleteOnly
			} else {
				__antithesis_instrumentation__.Notify(246834)
				if m.Merging() {
					__antithesis_instrumentation__.Notify(246835)
					tbl.Mutations[m.MutationOrdinal()].State = descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY
				} else {
					__antithesis_instrumentation__.Notify(246836)
				}
			}
		}
		__antithesis_instrumentation__.Notify(246818)
		if runStatus == "" || func() bool {
			__antithesis_instrumentation__.Notify(246837)
			return tbl.Dropped() == true
		}() == true {
			__antithesis_instrumentation__.Notify(246838)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(246839)
		}
		__antithesis_instrumentation__.Notify(246819)
		if err := descsCol.WriteDesc(
			ctx, true, tbl, txn,
		); err != nil {
			__antithesis_instrumentation__.Notify(246840)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246841)
		}
		__antithesis_instrumentation__.Notify(246820)
		if sc.job != nil {
			__antithesis_instrumentation__.Notify(246842)
			if err := sc.job.RunningStatus(ctx, txn, func(
				ctx context.Context, details jobspb.Details,
			) (jobs.RunningStatus, error) {
				__antithesis_instrumentation__.Notify(246843)
				return runStatus, nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(246844)
				return errors.Wrap(err, "failed to update job status")
			} else {
				__antithesis_instrumentation__.Notify(246845)
			}
		} else {
			__antithesis_instrumentation__.Notify(246846)
		}
		__antithesis_instrumentation__.Notify(246821)
		return nil
	})
}

func (sc *SchemaChanger) truncateAndBackfillColumns(
	ctx context.Context, version descpb.DescriptorVersion,
) error {
	__antithesis_instrumentation__.Notify(246847)
	log.Infof(ctx, "clearing and backfilling columns")

	if err := sc.distColumnBackfill(
		ctx, version, columnBackfillBatchSize.Get(&sc.settings.SV),
		backfill.ColumnMutationFilter); err != nil {
		__antithesis_instrumentation__.Notify(246849)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246850)
	}
	__antithesis_instrumentation__.Notify(246848)
	log.Info(ctx, "finished clearing and backfilling columns")
	return nil
}

func runSchemaChangesInTxn(
	ctx context.Context, planner *planner, tableDesc *tabledesc.Mutable, traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(246851)

	if len(tableDesc.GetDrainingNames()) > 0 {
		__antithesis_instrumentation__.Notify(246858)

		for _, drain := range tableDesc.GetDrainingNames() {
			__antithesis_instrumentation__.Notify(246860)
			key := catalogkeys.EncodeNameKey(planner.ExecCfg().Codec, drain)
			if err := planner.Txn().Del(ctx, key); err != nil {
				__antithesis_instrumentation__.Notify(246861)
				return err
			} else {
				__antithesis_instrumentation__.Notify(246862)
			}
		}
		__antithesis_instrumentation__.Notify(246859)

		tableDesc.SetDrainingNames(nil)
	} else {
		__antithesis_instrumentation__.Notify(246863)
	}
	__antithesis_instrumentation__.Notify(246852)

	if tableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(246864)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(246865)
	}
	__antithesis_instrumentation__.Notify(246853)

	doneColumnBackfill := false

	var constraintAdditionMutations []catalog.ConstraintToUpdate

	for _, m := range tableDesc.AllMutations() {
		__antithesis_instrumentation__.Notify(246866)

		if discarded, _ := isCurrentMutationDiscarded(tableDesc, m, m.MutationOrdinal()+1); discarded {
			__antithesis_instrumentation__.Notify(246870)
			continue
		} else {
			__antithesis_instrumentation__.Notify(246871)
		}
		__antithesis_instrumentation__.Notify(246867)

		if idx := m.AsIndex(); idx != nil && func() bool {
			__antithesis_instrumentation__.Notify(246872)
			return idx.IsTemporaryIndexForBackfill() == true
		}() == true {
			__antithesis_instrumentation__.Notify(246873)
			continue
		} else {
			__antithesis_instrumentation__.Notify(246874)
		}
		__antithesis_instrumentation__.Notify(246868)

		immutDesc := tabledesc.NewBuilder(tableDesc.TableDesc()).BuildImmutableTable()

		if m.Adding() {
			__antithesis_instrumentation__.Notify(246875)
			if m.AsPrimaryKeySwap() != nil || func() bool {
				__antithesis_instrumentation__.Notify(246876)
				return m.AsModifyRowLevelTTL() != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(246877)

			} else {
				__antithesis_instrumentation__.Notify(246878)
				if m.AsComputedColumnSwap() != nil {
					__antithesis_instrumentation__.Notify(246879)
					return AlterColTypeInTxnNotSupportedErr
				} else {
					__antithesis_instrumentation__.Notify(246880)
					if col := m.AsColumn(); col != nil {
						__antithesis_instrumentation__.Notify(246881)
						if !doneColumnBackfill && func() bool {
							__antithesis_instrumentation__.Notify(246882)
							return catalog.ColumnNeedsBackfill(col) == true
						}() == true {
							__antithesis_instrumentation__.Notify(246883)
							if err := columnBackfillInTxn(
								ctx, planner.Txn(), planner.ExecCfg(), planner.EvalContext(), planner.SemaCtx(),
								immutDesc, traceKV,
							); err != nil {
								__antithesis_instrumentation__.Notify(246885)
								return err
							} else {
								__antithesis_instrumentation__.Notify(246886)
							}
							__antithesis_instrumentation__.Notify(246884)
							doneColumnBackfill = true
						} else {
							__antithesis_instrumentation__.Notify(246887)
						}
					} else {
						__antithesis_instrumentation__.Notify(246888)
						if idx := m.AsIndex(); idx != nil {
							__antithesis_instrumentation__.Notify(246889)
							if err := indexBackfillInTxn(ctx, planner.Txn(), planner.EvalContext(), planner.SemaCtx(), immutDesc, traceKV); err != nil {
								__antithesis_instrumentation__.Notify(246890)
								return err
							} else {
								__antithesis_instrumentation__.Notify(246891)
							}
						} else {
							__antithesis_instrumentation__.Notify(246892)
							if c := m.AsConstraint(); c != nil {
								__antithesis_instrumentation__.Notify(246893)

								constraintAdditionMutations = append(constraintAdditionMutations, c)
								continue
							} else {
								__antithesis_instrumentation__.Notify(246894)
								return errors.AssertionFailedf("unsupported mutation: %+v", m)
							}
						}
					}
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(246895)
			if m.Dropped() {
				__antithesis_instrumentation__.Notify(246896)

				if col := m.AsColumn(); col != nil {
					__antithesis_instrumentation__.Notify(246897)
					if !doneColumnBackfill && func() bool {
						__antithesis_instrumentation__.Notify(246898)
						return catalog.ColumnNeedsBackfill(col) == true
					}() == true {
						__antithesis_instrumentation__.Notify(246899)
						if err := columnBackfillInTxn(
							ctx, planner.Txn(), planner.ExecCfg(), planner.EvalContext(), planner.SemaCtx(),
							immutDesc, traceKV,
						); err != nil {
							__antithesis_instrumentation__.Notify(246901)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246902)
						}
						__antithesis_instrumentation__.Notify(246900)
						doneColumnBackfill = true
					} else {
						__antithesis_instrumentation__.Notify(246903)
					}
				} else {
					__antithesis_instrumentation__.Notify(246904)
					if idx := m.AsIndex(); idx != nil {
						__antithesis_instrumentation__.Notify(246905)
						if err := indexTruncateInTxn(
							ctx, planner.Txn(), planner.ExecCfg(), planner.EvalContext(), immutDesc, idx, traceKV,
						); err != nil {
							__antithesis_instrumentation__.Notify(246906)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246907)
						}
					} else {
						__antithesis_instrumentation__.Notify(246908)
						if c := m.AsConstraint(); c != nil {
							__antithesis_instrumentation__.Notify(246909)
							if c.IsCheck() || func() bool {
								__antithesis_instrumentation__.Notify(246910)
								return c.IsNotNull() == true
							}() == true {
								__antithesis_instrumentation__.Notify(246911)
								for i := range tableDesc.Checks {
									__antithesis_instrumentation__.Notify(246912)
									if tableDesc.Checks[i].Name == c.GetName() {
										__antithesis_instrumentation__.Notify(246913)
										tableDesc.Checks = append(tableDesc.Checks[:i], tableDesc.Checks[i+1:]...)
										break
									} else {
										__antithesis_instrumentation__.Notify(246914)
									}
								}
							} else {
								__antithesis_instrumentation__.Notify(246915)
								if c.IsForeignKey() {
									__antithesis_instrumentation__.Notify(246916)
									for i := range tableDesc.OutboundFKs {
										__antithesis_instrumentation__.Notify(246917)
										fk := &tableDesc.OutboundFKs[i]
										if fk.Name == c.GetName() {
											__antithesis_instrumentation__.Notify(246918)
											if err := planner.removeFKBackReference(ctx, tableDesc, fk); err != nil {
												__antithesis_instrumentation__.Notify(246920)
												return err
											} else {
												__antithesis_instrumentation__.Notify(246921)
											}
											__antithesis_instrumentation__.Notify(246919)
											tableDesc.OutboundFKs = append(tableDesc.OutboundFKs[:i], tableDesc.OutboundFKs[i+1:]...)
											break
										} else {
											__antithesis_instrumentation__.Notify(246922)
										}
									}
								} else {
									__antithesis_instrumentation__.Notify(246923)
									if c.IsUniqueWithoutIndex() {
										__antithesis_instrumentation__.Notify(246924)
										for i := range tableDesc.UniqueWithoutIndexConstraints {
											__antithesis_instrumentation__.Notify(246925)
											if tableDesc.UniqueWithoutIndexConstraints[i].Name == c.GetName() {
												__antithesis_instrumentation__.Notify(246926)
												tableDesc.UniqueWithoutIndexConstraints = append(
													tableDesc.UniqueWithoutIndexConstraints[:i],
													tableDesc.UniqueWithoutIndexConstraints[i+1:]...,
												)
												break
											} else {
												__antithesis_instrumentation__.Notify(246927)
											}
										}
									} else {
										__antithesis_instrumentation__.Notify(246928)
										return errors.AssertionFailedf("unsupported constraint type: %d", c.ConstraintToUpdateDesc().ConstraintType)
									}
								}
							}
						} else {
							__antithesis_instrumentation__.Notify(246929)
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(246930)
			}
		}
		__antithesis_instrumentation__.Notify(246869)

		if err := tableDesc.MakeMutationComplete(tableDesc.Mutations[m.MutationOrdinal()]); err != nil {
			__antithesis_instrumentation__.Notify(246931)
			return err
		} else {
			__antithesis_instrumentation__.Notify(246932)
		}
	}
	__antithesis_instrumentation__.Notify(246854)

	tableDesc.Mutations = make([]descpb.DescriptorMutation, len(constraintAdditionMutations))
	for i, c := range constraintAdditionMutations {
		__antithesis_instrumentation__.Notify(246933)
		tableDesc.Mutations[i] = descpb.DescriptorMutation{
			Descriptor_: &descpb.DescriptorMutation_Constraint{Constraint: c.ConstraintToUpdateDesc()},
			Direction:   descpb.DescriptorMutation_ADD,
			MutationID:  c.MutationID(),
		}
		if c.DeleteOnly() {
			__antithesis_instrumentation__.Notify(246934)
			tableDesc.Mutations[i].State = descpb.DescriptorMutation_DELETE_ONLY
		} else {
			__antithesis_instrumentation__.Notify(246935)
			if c.WriteAndDeleteOnly() {
				__antithesis_instrumentation__.Notify(246936)
				tableDesc.Mutations[i].State = descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY
			} else {
				__antithesis_instrumentation__.Notify(246937)
			}
		}
	}
	__antithesis_instrumentation__.Notify(246855)

	for _, c := range constraintAdditionMutations {
		__antithesis_instrumentation__.Notify(246938)
		if c.IsCheck() || func() bool {
			__antithesis_instrumentation__.Notify(246939)
			return c.IsNotNull() == true
		}() == true {
			__antithesis_instrumentation__.Notify(246940)
			check := &c.ConstraintToUpdateDesc().Check
			if check.Validity == descpb.ConstraintValidity_Validating {
				__antithesis_instrumentation__.Notify(246941)
				if err := validateCheckInTxn(
					ctx, &planner.semaCtx, planner.ExecCfg().InternalExecutorFactory,
					planner.SessionData(), tableDesc, planner.txn, check.Expr,
				); err != nil {
					__antithesis_instrumentation__.Notify(246943)
					return err
				} else {
					__antithesis_instrumentation__.Notify(246944)
				}
				__antithesis_instrumentation__.Notify(246942)
				check.Validity = descpb.ConstraintValidity_Validated
			} else {
				__antithesis_instrumentation__.Notify(246945)
			}
		} else {
			__antithesis_instrumentation__.Notify(246946)
			if c.IsForeignKey() {
				__antithesis_instrumentation__.Notify(246947)

				c.ConstraintToUpdateDesc().ForeignKey.Validity = descpb.ConstraintValidity_Unvalidated
			} else {
				__antithesis_instrumentation__.Notify(246948)
				if c.IsUniqueWithoutIndex() {
					__antithesis_instrumentation__.Notify(246949)
					uwi := &c.ConstraintToUpdateDesc().UniqueWithoutIndexConstraint
					if uwi.Validity == descpb.ConstraintValidity_Validating {
						__antithesis_instrumentation__.Notify(246950)
						if err := validateUniqueWithoutIndexConstraintInTxn(
							ctx, planner.ExecCfg().InternalExecutor, tableDesc, planner.txn, c.GetName(),
						); err != nil {
							__antithesis_instrumentation__.Notify(246952)
							return err
						} else {
							__antithesis_instrumentation__.Notify(246953)
						}
						__antithesis_instrumentation__.Notify(246951)
						uwi.Validity = descpb.ConstraintValidity_Validated
					} else {
						__antithesis_instrumentation__.Notify(246954)
					}
				} else {
					__antithesis_instrumentation__.Notify(246955)
					return errors.AssertionFailedf("unsupported constraint type: %d", c.ConstraintToUpdateDesc().ConstraintType)
				}
			}
		}

	}
	__antithesis_instrumentation__.Notify(246856)

	for _, c := range constraintAdditionMutations {
		__antithesis_instrumentation__.Notify(246956)
		if c.IsCheck() || func() bool {
			__antithesis_instrumentation__.Notify(246957)
			return c.IsNotNull() == true
		}() == true {
			__antithesis_instrumentation__.Notify(246958)
			tableDesc.Checks = append(tableDesc.Checks, &c.ConstraintToUpdateDesc().Check)
		} else {
			__antithesis_instrumentation__.Notify(246959)
			if c.IsForeignKey() {
				__antithesis_instrumentation__.Notify(246960)
				fk := c.ConstraintToUpdateDesc().ForeignKey
				var referencedTableDesc *tabledesc.Mutable

				selfReference := tableDesc.ID == fk.ReferencedTableID
				if selfReference {
					__antithesis_instrumentation__.Notify(246962)
					referencedTableDesc = tableDesc
				} else {
					__antithesis_instrumentation__.Notify(246963)
					lookup, err := planner.Descriptors().GetMutableTableVersionByID(ctx, fk.ReferencedTableID, planner.Txn())
					if err != nil {
						__antithesis_instrumentation__.Notify(246965)
						return errors.Wrapf(err, "error resolving referenced table ID %d", fk.ReferencedTableID)
					} else {
						__antithesis_instrumentation__.Notify(246966)
					}
					__antithesis_instrumentation__.Notify(246964)
					referencedTableDesc = lookup
				}
				__antithesis_instrumentation__.Notify(246961)
				referencedTableDesc.InboundFKs = append(referencedTableDesc.InboundFKs, fk)
				tableDesc.OutboundFKs = append(tableDesc.OutboundFKs, fk)

				if !selfReference {
					__antithesis_instrumentation__.Notify(246967)
					if err := planner.writeSchemaChange(
						ctx, referencedTableDesc, descpb.InvalidMutationID,
						fmt.Sprintf("updating referenced FK table %s(%d) table %s(%d)",
							referencedTableDesc.Name, referencedTableDesc.ID, tableDesc.Name, tableDesc.ID),
					); err != nil {
						__antithesis_instrumentation__.Notify(246968)
						return err
					} else {
						__antithesis_instrumentation__.Notify(246969)
					}
				} else {
					__antithesis_instrumentation__.Notify(246970)
				}
			} else {
				__antithesis_instrumentation__.Notify(246971)
				if c.IsUniqueWithoutIndex() {
					__antithesis_instrumentation__.Notify(246972)
					tableDesc.UniqueWithoutIndexConstraints = append(
						tableDesc.UniqueWithoutIndexConstraints, c.ConstraintToUpdateDesc().UniqueWithoutIndexConstraint,
					)
				} else {
					__antithesis_instrumentation__.Notify(246973)
					return errors.AssertionFailedf("unsupported constraint type: %d", c.ConstraintToUpdateDesc().ConstraintType)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(246857)
	tableDesc.Mutations = nil
	return nil
}

func validateCheckInTxn(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	ief sqlutil.SessionBoundInternalExecutorFactory,
	sessionData *sessiondata.SessionData,
	tableDesc *tabledesc.Mutable,
	txn *kv.Txn,
	checkExpr string,
) error {
	__antithesis_instrumentation__.Notify(246974)
	var syntheticDescs []catalog.Descriptor
	if tableDesc.Version > tableDesc.ClusterVersion().Version {
		__antithesis_instrumentation__.Notify(246976)
		syntheticDescs = append(syntheticDescs, tableDesc)
	} else {
		__antithesis_instrumentation__.Notify(246977)
	}
	__antithesis_instrumentation__.Notify(246975)
	ie := ief(ctx, sessionData)
	return ie.WithSyntheticDescriptors(syntheticDescs, func() error {
		__antithesis_instrumentation__.Notify(246978)
		return validateCheckExpr(ctx, semaCtx, sessionData, checkExpr, tableDesc, ie, txn)
	})
}

func validateFkInTxn(
	ctx context.Context,
	ief sqlutil.SessionBoundInternalExecutorFactory,
	sd *sessiondata.SessionData,
	srcTable *tabledesc.Mutable,
	txn *kv.Txn,
	descsCol *descs.Collection,
	fkName string,
) error {
	__antithesis_instrumentation__.Notify(246979)
	var syntheticTable catalog.TableDescriptor
	if srcTable.Version > srcTable.ClusterVersion().Version {
		__antithesis_instrumentation__.Notify(246985)
		syntheticTable = srcTable
	} else {
		__antithesis_instrumentation__.Notify(246986)
	}
	__antithesis_instrumentation__.Notify(246980)
	var fk *descpb.ForeignKeyConstraint
	for i := range srcTable.OutboundFKs {
		__antithesis_instrumentation__.Notify(246987)
		def := &srcTable.OutboundFKs[i]
		if def.Name == fkName {
			__antithesis_instrumentation__.Notify(246988)
			fk = def
			break
		} else {
			__antithesis_instrumentation__.Notify(246989)
		}
	}
	__antithesis_instrumentation__.Notify(246981)
	if fk == nil {
		__antithesis_instrumentation__.Notify(246990)
		return errors.AssertionFailedf("foreign key %s does not exist", fkName)
	} else {
		__antithesis_instrumentation__.Notify(246991)
	}
	__antithesis_instrumentation__.Notify(246982)
	targetTable, err := descsCol.Direct().MustGetTableDescByID(ctx, txn, fk.ReferencedTableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(246992)
		return err
	} else {
		__antithesis_instrumentation__.Notify(246993)
	}
	__antithesis_instrumentation__.Notify(246983)
	var syntheticDescs []catalog.Descriptor
	if syntheticTable != nil {
		__antithesis_instrumentation__.Notify(246994)
		syntheticDescs = append(syntheticDescs, syntheticTable)
		if targetTable.GetID() == syntheticTable.GetID() {
			__antithesis_instrumentation__.Notify(246995)
			targetTable = syntheticTable
		} else {
			__antithesis_instrumentation__.Notify(246996)
		}
	} else {
		__antithesis_instrumentation__.Notify(246997)
	}
	__antithesis_instrumentation__.Notify(246984)
	ie := ief(ctx, sd)
	return ie.WithSyntheticDescriptors(syntheticDescs, func() error {
		__antithesis_instrumentation__.Notify(246998)
		return validateForeignKey(ctx, srcTable, targetTable, fk, ie, txn)
	})
}

func validateUniqueWithoutIndexConstraintInTxn(
	ctx context.Context,
	ie sqlutil.InternalExecutor,
	tableDesc *tabledesc.Mutable,
	txn *kv.Txn,
	constraintName string,
) error {
	__antithesis_instrumentation__.Notify(246999)
	var syntheticDescs []catalog.Descriptor
	if tableDesc.Version > tableDesc.ClusterVersion().Version {
		__antithesis_instrumentation__.Notify(247003)
		syntheticDescs = append(syntheticDescs, tableDesc)
	} else {
		__antithesis_instrumentation__.Notify(247004)
	}
	__antithesis_instrumentation__.Notify(247000)

	var uc *descpb.UniqueWithoutIndexConstraint
	for i := range tableDesc.UniqueWithoutIndexConstraints {
		__antithesis_instrumentation__.Notify(247005)
		def := &tableDesc.UniqueWithoutIndexConstraints[i]
		if def.Name == constraintName {
			__antithesis_instrumentation__.Notify(247006)
			uc = def
			break
		} else {
			__antithesis_instrumentation__.Notify(247007)
		}
	}
	__antithesis_instrumentation__.Notify(247001)
	if uc == nil {
		__antithesis_instrumentation__.Notify(247008)
		return errors.AssertionFailedf("unique constraint %s does not exist", constraintName)
	} else {
		__antithesis_instrumentation__.Notify(247009)
	}
	__antithesis_instrumentation__.Notify(247002)

	return ie.WithSyntheticDescriptors(syntheticDescs, func() error {
		__antithesis_instrumentation__.Notify(247010)
		return validateUniqueConstraint(
			ctx, tableDesc, uc.Name, uc.ColumnIDs, uc.Predicate, ie, txn, false,
		)
	})
}

func columnBackfillInTxn(
	ctx context.Context,
	txn *kv.Txn,
	execCfg *ExecutorConfig,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
	tableDesc catalog.TableDescriptor,
	traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(247011)

	if tableDesc.Adding() {
		__antithesis_instrumentation__.Notify(247016)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(247017)
	}
	__antithesis_instrumentation__.Notify(247012)
	var columnBackfillerMon *mon.BytesMonitor

	if evalCtx.Mon != nil {
		__antithesis_instrumentation__.Notify(247018)
		columnBackfillerMon = execinfra.NewMonitor(ctx, evalCtx.Mon, "local-column-backfill-mon")
	} else {
		__antithesis_instrumentation__.Notify(247019)
	}
	__antithesis_instrumentation__.Notify(247013)

	rowMetrics := execCfg.GetRowMetrics(evalCtx.SessionData().Internal)
	var backfiller backfill.ColumnBackfiller
	if err := backfiller.InitForLocalUse(
		ctx, evalCtx, semaCtx, tableDesc, columnBackfillerMon, rowMetrics,
	); err != nil {
		__antithesis_instrumentation__.Notify(247020)
		return err
	} else {
		__antithesis_instrumentation__.Notify(247021)
	}
	__antithesis_instrumentation__.Notify(247014)
	defer backfiller.Close(ctx)
	sp := tableDesc.PrimaryIndexSpan(evalCtx.Codec)
	for sp.Key != nil {
		__antithesis_instrumentation__.Notify(247022)
		var err error
		sp.Key, err = backfiller.RunColumnBackfillChunk(ctx,
			txn, tableDesc, sp, rowinfra.RowLimit(columnBackfillBatchSize.Get(&evalCtx.Settings.SV)),
			false, traceKV)
		if err != nil {
			__antithesis_instrumentation__.Notify(247023)
			return err
		} else {
			__antithesis_instrumentation__.Notify(247024)
		}
	}
	__antithesis_instrumentation__.Notify(247015)

	return nil
}

func indexBackfillInTxn(
	ctx context.Context,
	txn *kv.Txn,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
	tableDesc catalog.TableDescriptor,
	traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(247025)
	var indexBackfillerMon *mon.BytesMonitor

	if evalCtx.Mon != nil {
		__antithesis_instrumentation__.Notify(247029)
		indexBackfillerMon = execinfra.NewMonitor(ctx, evalCtx.Mon, "local-index-backfill-mon")
	} else {
		__antithesis_instrumentation__.Notify(247030)
	}
	__antithesis_instrumentation__.Notify(247026)

	var backfiller backfill.IndexBackfiller
	if err := backfiller.InitForLocalUse(
		ctx, evalCtx, semaCtx, tableDesc, indexBackfillerMon,
	); err != nil {
		__antithesis_instrumentation__.Notify(247031)
		return err
	} else {
		__antithesis_instrumentation__.Notify(247032)
	}
	__antithesis_instrumentation__.Notify(247027)
	defer backfiller.Close(ctx)
	sp := tableDesc.PrimaryIndexSpan(evalCtx.Codec)
	for sp.Key != nil {
		__antithesis_instrumentation__.Notify(247033)
		var err error
		sp.Key, err = backfiller.RunIndexBackfillChunk(ctx,
			txn, tableDesc, sp, indexTxnBackfillChunkSize, false, traceKV)
		if err != nil {
			__antithesis_instrumentation__.Notify(247034)
			return err
		} else {
			__antithesis_instrumentation__.Notify(247035)
		}
	}
	__antithesis_instrumentation__.Notify(247028)

	return nil
}

func indexTruncateInTxn(
	ctx context.Context,
	txn *kv.Txn,
	execCfg *ExecutorConfig,
	evalCtx *tree.EvalContext,
	tableDesc catalog.TableDescriptor,
	idx catalog.Index,
	traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(247036)
	alloc := &tree.DatumAlloc{}
	var sp roachpb.Span
	for done := false; !done; done = sp.Key == nil {
		__antithesis_instrumentation__.Notify(247038)
		internal := evalCtx.SessionData().Internal
		rd := row.MakeDeleter(
			execCfg.Codec, tableDesc, nil, &execCfg.Settings.SV, internal,
			execCfg.GetRowMetrics(internal),
		)
		td := tableDeleter{rd: rd, alloc: alloc}
		if err := td.init(ctx, txn, evalCtx, &evalCtx.Settings.SV); err != nil {
			__antithesis_instrumentation__.Notify(247040)
			return err
		} else {
			__antithesis_instrumentation__.Notify(247041)
		}
		__antithesis_instrumentation__.Notify(247039)
		var err error
		sp, err = td.deleteIndex(
			ctx, idx, sp, indexTruncateChunkSize, traceKV,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(247042)
			return err
		} else {
			__antithesis_instrumentation__.Notify(247043)
		}
	}
	__antithesis_instrumentation__.Notify(247037)

	return RemoveIndexZoneConfigs(ctx, txn, execCfg, tableDesc, []uint32{uint32(idx.GetID())})
}

func (sc *SchemaChanger) distIndexMerge(
	ctx context.Context,
	tableDesc catalog.TableDescriptor,
	addedIndexes []descpb.IndexID,
	temporaryIndexes []descpb.IndexID,
	fractionScaler *multiStageFractionScaler,
) error {
	__antithesis_instrumentation__.Notify(247044)

	mergeTimestamp := sc.clock.Now()
	log.Infof(ctx, "merging all keys in temporary index before time %v", mergeTimestamp)

	progress, err := extractMergeProgress(sc.job, tableDesc, addedIndexes, temporaryIndexes)
	if err != nil {
		__antithesis_instrumentation__.Notify(247054)
		return err
	} else {
		__antithesis_instrumentation__.Notify(247055)
	}
	__antithesis_instrumentation__.Notify(247045)

	log.VEventf(ctx, 2, "indexbackfill merge: initial resume spans %+v", progress.TodoSpans)
	if progress.TodoSpans == nil {
		__antithesis_instrumentation__.Notify(247056)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(247057)
	}
	__antithesis_instrumentation__.Notify(247046)

	planner := NewIndexBackfillerMergePlanner(sc.execCfg, sc.execCfg.InternalExecutorFactory)
	rc := func(ctx context.Context, spans []roachpb.Span) (int, error) {
		__antithesis_instrumentation__.Notify(247058)
		return numRangesInSpans(ctx, sc.db, sc.distSQLPlanner, spans)
	}
	__antithesis_instrumentation__.Notify(247047)
	tracker := NewIndexMergeTracker(progress, sc.job, rc, fractionScaler)
	periodicFlusher := newPeriodicProgressFlusher(sc.settings)

	metaFn := func(ctx context.Context, meta *execinfrapb.ProducerMetadata) error {
		__antithesis_instrumentation__.Notify(247059)
		if meta.BulkProcessorProgress != nil {
			__antithesis_instrumentation__.Notify(247061)
			idxCompletedSpans := make(map[int32][]roachpb.Span)
			for i, sp := range meta.BulkProcessorProgress.CompletedSpans {
				__antithesis_instrumentation__.Notify(247065)
				spanIdx := meta.BulkProcessorProgress.CompletedSpanIdx[i]
				idxCompletedSpans[spanIdx] = append(idxCompletedSpans[spanIdx], sp)
			}
			__antithesis_instrumentation__.Notify(247062)
			tracker.UpdateMergeProgress(ctx, func(_ context.Context, currentProgress *MergeProgress) {
				__antithesis_instrumentation__.Notify(247066)
				for idx, completedSpans := range idxCompletedSpans {
					__antithesis_instrumentation__.Notify(247067)
					currentProgress.TodoSpans[idx] = roachpb.SubtractSpans(currentProgress.TodoSpans[idx], completedSpans)
				}
			})
			__antithesis_instrumentation__.Notify(247063)
			if sc.testingKnobs.AlwaysUpdateIndexBackfillDetails {
				__antithesis_instrumentation__.Notify(247068)
				if err := tracker.FlushCheckpoint(ctx); err != nil {
					__antithesis_instrumentation__.Notify(247069)
					return err
				} else {
					__antithesis_instrumentation__.Notify(247070)
				}
			} else {
				__antithesis_instrumentation__.Notify(247071)
			}
			__antithesis_instrumentation__.Notify(247064)
			if sc.testingKnobs.AlwaysUpdateIndexBackfillProgress {
				__antithesis_instrumentation__.Notify(247072)
				if err := tracker.FlushFractionCompleted(ctx); err != nil {
					__antithesis_instrumentation__.Notify(247073)
					return err
				} else {
					__antithesis_instrumentation__.Notify(247074)
				}
			} else {
				__antithesis_instrumentation__.Notify(247075)
			}
		} else {
			__antithesis_instrumentation__.Notify(247076)
		}
		__antithesis_instrumentation__.Notify(247060)
		return nil
	}
	__antithesis_instrumentation__.Notify(247048)

	stop := periodicFlusher.StartPeriodicUpdates(ctx, tracker)
	defer func() { __antithesis_instrumentation__.Notify(247077); _ = stop() }()
	__antithesis_instrumentation__.Notify(247049)

	run, err := planner.plan(ctx, tableDesc, progress.TodoSpans, progress.AddedIndexes,
		progress.TemporaryIndexes, metaFn, mergeTimestamp)
	if err != nil {
		__antithesis_instrumentation__.Notify(247078)
		return err
	} else {
		__antithesis_instrumentation__.Notify(247079)
	}
	__antithesis_instrumentation__.Notify(247050)

	if err := run(ctx); err != nil {
		__antithesis_instrumentation__.Notify(247080)
		return err
	} else {
		__antithesis_instrumentation__.Notify(247081)
	}
	__antithesis_instrumentation__.Notify(247051)

	if err := stop(); err != nil {
		__antithesis_instrumentation__.Notify(247082)
		return err
	} else {
		__antithesis_instrumentation__.Notify(247083)
	}
	__antithesis_instrumentation__.Notify(247052)

	if err := tracker.FlushCheckpoint(ctx); err != nil {
		__antithesis_instrumentation__.Notify(247084)
		return err
	} else {
		__antithesis_instrumentation__.Notify(247085)
	}
	__antithesis_instrumentation__.Notify(247053)

	return tracker.FlushFractionCompleted(ctx)
}

func extractMergeProgress(
	job *jobs.Job, tableDesc catalog.TableDescriptor, addedIndexes, temporaryIndexes []descpb.IndexID,
) (*MergeProgress, error) {
	__antithesis_instrumentation__.Notify(247086)
	resumeSpanList := job.Details().(jobspb.SchemaChangeDetails).ResumeSpanList
	progress := MergeProgress{}
	progress.TemporaryIndexes = temporaryIndexes
	progress.AddedIndexes = addedIndexes

	const noIdx = -1
	findMutIdx := func(id descpb.IndexID) int {
		__antithesis_instrumentation__.Notify(247089)
		for mutIdx, mut := range tableDesc.AllMutations() {
			__antithesis_instrumentation__.Notify(247091)
			if mut.AsIndex() != nil && func() bool {
				__antithesis_instrumentation__.Notify(247092)
				return mut.AsIndex().GetID() == id == true
			}() == true {
				__antithesis_instrumentation__.Notify(247093)
				return mutIdx
			} else {
				__antithesis_instrumentation__.Notify(247094)
			}
		}
		__antithesis_instrumentation__.Notify(247090)

		return noIdx
	}
	__antithesis_instrumentation__.Notify(247087)

	for _, tempIdx := range temporaryIndexes {
		__antithesis_instrumentation__.Notify(247095)
		mutIdx := findMutIdx(tempIdx)
		if mutIdx == noIdx {
			__antithesis_instrumentation__.Notify(247097)
			return nil, errors.AssertionFailedf("no corresponding mutation for temporary index %d", tempIdx)
		} else {
			__antithesis_instrumentation__.Notify(247098)
		}
		__antithesis_instrumentation__.Notify(247096)

		progress.TodoSpans = append(progress.TodoSpans, resumeSpanList[mutIdx].ResumeSpans)
		progress.MutationIdx = append(progress.MutationIdx, mutIdx)
	}
	__antithesis_instrumentation__.Notify(247088)

	return &progress, nil
}

type backfillStage int

const (
	stageBackfill backfillStage = iota
	stageMerge
)

var (
	mvccCompatibleBackfillStageFractions = []float32{
		.60,
		1.0,
	}
	backfillStageFractions = []float32{
		1.0,
	}
)

type multiStageFractionScaler struct {
	initial float32
	stages  []float32
}

func (m *multiStageFractionScaler) fractionCompleteFromStageFraction(
	stage backfillStage, fraction float32,
) (float32, error) {
	__antithesis_instrumentation__.Notify(247099)
	if fraction > 1.0 || func() bool {
		__antithesis_instrumentation__.Notify(247105)
		return fraction < 0.0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(247106)
		return 0, errors.AssertionFailedf("fraction %f outside allowed range [0.0, 1.0]", fraction)
	} else {
		__antithesis_instrumentation__.Notify(247107)
	}
	__antithesis_instrumentation__.Notify(247100)

	if int(stage) >= len(m.stages) {
		__antithesis_instrumentation__.Notify(247108)
		return 0, errors.AssertionFailedf("unknown stage %d", stage)
	} else {
		__antithesis_instrumentation__.Notify(247109)
	}
	__antithesis_instrumentation__.Notify(247101)

	max := m.stages[stage]
	if max > 1.0 {
		__antithesis_instrumentation__.Notify(247110)
		return 0, errors.AssertionFailedf("stage %d max percentage larger than 1: %f", stage, max)
	} else {
		__antithesis_instrumentation__.Notify(247111)
	}
	__antithesis_instrumentation__.Notify(247102)

	min := m.initial
	if stage > 0 {
		__antithesis_instrumentation__.Notify(247112)
		min = m.stages[stage-1]
	} else {
		__antithesis_instrumentation__.Notify(247113)
	}
	__antithesis_instrumentation__.Notify(247103)

	v := min + (max-min)*fraction
	if v < m.initial {
		__antithesis_instrumentation__.Notify(247114)
		return m.initial, nil
	} else {
		__antithesis_instrumentation__.Notify(247115)
	}
	__antithesis_instrumentation__.Notify(247104)
	return v, nil
}
