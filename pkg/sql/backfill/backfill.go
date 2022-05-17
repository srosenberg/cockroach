package backfill

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/transform"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

var IndexBackfillCheckpointInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"bulkio.index_backfill.checkpoint_interval",
	"the amount of time between index backfill checkpoint updates",
	30*time.Second,
	settings.NonNegativeDuration,
)

type MutationFilter func(catalog.Mutation) bool

func ColumnMutationFilter(m catalog.Mutation) bool {
	__antithesis_instrumentation__.Notify(245618)
	return m.AsColumn() != nil && func() bool {
		__antithesis_instrumentation__.Notify(245619)
		return (m.Adding() || func() bool {
			__antithesis_instrumentation__.Notify(245620)
			return m.Dropped() == true
		}() == true) == true
	}() == true
}

func IndexMutationFilter(m catalog.Mutation) bool {
	__antithesis_instrumentation__.Notify(245621)
	idx := m.AsIndex()
	return idx != nil && func() bool {
		__antithesis_instrumentation__.Notify(245622)
		return !idx.IsTemporaryIndexForBackfill() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(245623)
		return m.Adding() == true
	}() == true
}

type ColumnBackfiller struct {
	added   []catalog.Column
	dropped []catalog.Column

	updateCols  []catalog.Column
	updateExprs []tree.TypedExpr
	evalCtx     *tree.EvalContext

	fetcher     row.Fetcher
	fetcherCols []descpb.ColumnID
	colIdxMap   catalog.TableColMap
	alloc       tree.DatumAlloc

	mon *mon.BytesMonitor

	rowMetrics *row.Metrics
}

func (cb *ColumnBackfiller) initCols(desc catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(245624)
	for _, m := range desc.AllMutations() {
		__antithesis_instrumentation__.Notify(245625)
		if ColumnMutationFilter(m) {
			__antithesis_instrumentation__.Notify(245626)
			col := m.AsColumn()
			if m.Adding() {
				__antithesis_instrumentation__.Notify(245627)
				cb.added = append(cb.added, col)
			} else {
				__antithesis_instrumentation__.Notify(245628)
				if m.Dropped() {
					__antithesis_instrumentation__.Notify(245629)
					cb.dropped = append(cb.dropped, col)
				} else {
					__antithesis_instrumentation__.Notify(245630)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(245631)
		}
	}
}

func (cb *ColumnBackfiller) init(
	evalCtx *tree.EvalContext,
	defaultExprs []tree.TypedExpr,
	computedExprs []tree.TypedExpr,
	desc catalog.TableDescriptor,
	mon *mon.BytesMonitor,
	rowMetrics *row.Metrics,
) error {
	cb.evalCtx = evalCtx
	cb.updateCols = append(cb.added, cb.dropped...)

	cb.updateExprs = make([]tree.TypedExpr, len(cb.updateCols))
	for j, col := range cb.added {
		if col.IsComputed() {
			cb.updateExprs[j] = computedExprs[j]
		} else if defaultExprs == nil || defaultExprs[j] == nil {
			cb.updateExprs[j] = tree.DNull
		} else {
			cb.updateExprs[j] = defaultExprs[j]
		}
	}
	for j := range cb.dropped {
		cb.updateExprs[j+len(cb.added)] = tree.DNull
	}

	for _, c := range desc.PublicColumns() {
		if !c.IsVirtual() {
			cb.fetcherCols = append(cb.fetcherCols, c.GetID())
		}
	}

	cb.colIdxMap = catalog.ColumnIDToOrdinalMap(desc.PublicColumns())
	var spec descpb.IndexFetchSpec
	if err := rowenc.InitIndexFetchSpec(&spec, evalCtx.Codec, desc, desc.GetPrimaryIndex(), cb.fetcherCols); err != nil {
		return err
	}

	if mon == nil {
		return errors.AssertionFailedf("no memory monitor linked to ColumnBackfiller during init")
	}
	cb.mon = mon
	cb.rowMetrics = rowMetrics

	return cb.fetcher.Init(
		evalCtx.Context,
		false,
		descpb.ScanLockingStrength_FOR_NONE,
		descpb.ScanLockingWaitPolicy_BLOCK,
		0,
		&cb.alloc,
		cb.mon,
		&spec,
	)
}

func (cb *ColumnBackfiller) InitForLocalUse(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
	desc catalog.TableDescriptor,
	mon *mon.BytesMonitor,
	rowMetrics *row.Metrics,
) error {
	__antithesis_instrumentation__.Notify(245632)
	cb.initCols(desc)
	defaultExprs, err := schemaexpr.MakeDefaultExprs(
		ctx, cb.added, &transform.ExprTransformContext{}, evalCtx, semaCtx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245635)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245636)
	}
	__antithesis_instrumentation__.Notify(245633)
	computedExprs, _, err := schemaexpr.MakeComputedExprs(
		ctx,
		cb.added,
		desc.PublicColumns(),
		desc,
		tree.NewUnqualifiedTableName(tree.Name(desc.GetName())),
		evalCtx,
		semaCtx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245637)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245638)
	}
	__antithesis_instrumentation__.Notify(245634)
	return cb.init(evalCtx, defaultExprs, computedExprs, desc, mon, rowMetrics)
}

func (cb *ColumnBackfiller) InitForDistributedUse(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	desc catalog.TableDescriptor,
	mon *mon.BytesMonitor,
) error {
	__antithesis_instrumentation__.Notify(245639)
	cb.initCols(desc)
	evalCtx := flowCtx.NewEvalCtx()
	var defaultExprs, computedExprs []tree.TypedExpr

	if err := flowCtx.Cfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(245641)
		resolver := flowCtx.NewTypeResolver(txn)

		if err := typedesc.HydrateTypesInTableDescriptor(ctx, desc.TableDesc(), &resolver); err != nil {
			__antithesis_instrumentation__.Notify(245645)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245646)
		}
		__antithesis_instrumentation__.Notify(245642)

		semaCtx := tree.MakeSemaContext()
		semaCtx.TypeResolver = &resolver
		var err error
		defaultExprs, err = schemaexpr.MakeDefaultExprs(
			ctx, cb.added, &transform.ExprTransformContext{}, evalCtx, &semaCtx,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(245647)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245648)
		}
		__antithesis_instrumentation__.Notify(245643)
		computedExprs, _, err = schemaexpr.MakeComputedExprs(
			ctx,
			cb.added,
			desc.PublicColumns(),
			desc,
			tree.NewUnqualifiedTableName(tree.Name(desc.GetName())),
			evalCtx,
			&semaCtx,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(245649)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245650)
		}
		__antithesis_instrumentation__.Notify(245644)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(245651)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245652)
	}
	__antithesis_instrumentation__.Notify(245640)

	flowCtx.Descriptors.ReleaseAll(ctx)

	rowMetrics := flowCtx.GetRowMetrics()
	return cb.init(evalCtx, defaultExprs, computedExprs, desc, mon, rowMetrics)
}

func (cb *ColumnBackfiller) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(245653)
	cb.fetcher.Close(ctx)
	if cb.mon != nil {
		__antithesis_instrumentation__.Notify(245654)
		cb.mon.Stop(ctx)
	} else {
		__antithesis_instrumentation__.Notify(245655)
	}
}

func (cb *ColumnBackfiller) RunColumnBackfillChunk(
	ctx context.Context,
	txn *kv.Txn,
	tableDesc catalog.TableDescriptor,
	sp roachpb.Span,
	chunkSize rowinfra.RowLimit,
	alsoCommit bool,
	traceKV bool,
) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(245656)

	requestedCols := make([]catalog.Column, 0, len(tableDesc.PublicColumns())+len(cb.added)+len(cb.dropped))
	requestedCols = append(requestedCols, tableDesc.PublicColumns()...)
	requestedCols = append(requestedCols, cb.added...)
	requestedCols = append(requestedCols, cb.dropped...)
	ru, err := row.MakeUpdater(
		ctx,
		txn,
		cb.evalCtx.Codec,
		tableDesc,
		cb.updateCols,
		requestedCols,
		row.UpdaterOnlyColumns,
		&cb.alloc,
		&cb.evalCtx.Settings.SV,
		cb.evalCtx.SessionData().Internal,
		cb.rowMetrics,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245664)
		return roachpb.Key{}, err
	} else {
		__antithesis_instrumentation__.Notify(245665)
	}
	__antithesis_instrumentation__.Notify(245657)

	if !ru.IsColumnOnlyUpdate() {
		__antithesis_instrumentation__.Notify(245666)
		panic("only column data should be modified, but the rowUpdater is configured otherwise")
	} else {
		__antithesis_instrumentation__.Notify(245667)
	}
	__antithesis_instrumentation__.Notify(245658)

	if err := cb.fetcher.StartScan(
		ctx, txn, []roachpb.Span{sp}, rowinfra.DefaultBatchBytesLimit, chunkSize,
		traceKV, false,
	); err != nil {
		__antithesis_instrumentation__.Notify(245668)
		log.Errorf(ctx, "scan error: %s", err)
		return roachpb.Key{}, err
	} else {
		__antithesis_instrumentation__.Notify(245669)
	}
	__antithesis_instrumentation__.Notify(245659)

	updateValues := make(tree.Datums, len(cb.updateExprs))
	b := txn.NewBatch()
	iv := &schemaexpr.RowIndexedVarContainer{
		Cols:    make([]catalog.Column, 0, len(tableDesc.PublicColumns())+len(cb.added)),
		Mapping: ru.FetchColIDtoRowIndex,
	}
	iv.Cols = append(iv.Cols, tableDesc.PublicColumns()...)
	iv.Cols = append(iv.Cols, cb.added...)
	cb.evalCtx.IVarContainer = iv

	fetchedValues := make(tree.Datums, cb.colIdxMap.Len())
	iv.CurSourceRow = make(tree.Datums, len(iv.Cols))

	oldValues := make(tree.Datums, len(ru.FetchCols))
	for i := range oldValues {
		__antithesis_instrumentation__.Notify(245670)
		oldValues[i] = tree.DNull
	}
	__antithesis_instrumentation__.Notify(245660)

	for i := int64(0); i < int64(chunkSize); i++ {
		__antithesis_instrumentation__.Notify(245671)
		ok, err := cb.fetcher.NextRowDecodedInto(ctx, fetchedValues, cb.colIdxMap)
		if err != nil {
			__antithesis_instrumentation__.Notify(245675)
			return roachpb.Key{}, err
		} else {
			__antithesis_instrumentation__.Notify(245676)
		}
		__antithesis_instrumentation__.Notify(245672)
		if !ok {
			__antithesis_instrumentation__.Notify(245677)
			break
		} else {
			__antithesis_instrumentation__.Notify(245678)
		}
		__antithesis_instrumentation__.Notify(245673)

		iv.CurSourceRow = append(iv.CurSourceRow[:0], fetchedValues...)

		for j, e := range cb.updateExprs {
			__antithesis_instrumentation__.Notify(245679)
			val, err := e.Eval(cb.evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(245683)
				return roachpb.Key{}, sqlerrors.NewInvalidSchemaDefinitionError(err)
			} else {
				__antithesis_instrumentation__.Notify(245684)
			}
			__antithesis_instrumentation__.Notify(245680)
			if j < len(cb.added) && func() bool {
				__antithesis_instrumentation__.Notify(245685)
				return !cb.added[j].IsNullable() == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(245686)
				return val == tree.DNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(245687)
				return roachpb.Key{}, sqlerrors.NewNonNullViolationError(cb.added[j].GetName())
			} else {
				__antithesis_instrumentation__.Notify(245688)
			}
			__antithesis_instrumentation__.Notify(245681)

			if j < len(cb.added) {
				__antithesis_instrumentation__.Notify(245689)
				iv.CurSourceRow = append(iv.CurSourceRow, val)
			} else {
				__antithesis_instrumentation__.Notify(245690)
			}
			__antithesis_instrumentation__.Notify(245682)
			updateValues[j] = val
		}
		__antithesis_instrumentation__.Notify(245674)
		copy(oldValues, fetchedValues)

		var pm row.PartialIndexUpdateHelper
		if _, err := ru.UpdateRow(
			ctx, b, oldValues, updateValues, pm, traceKV,
		); err != nil {
			__antithesis_instrumentation__.Notify(245691)
			return roachpb.Key{}, err
		} else {
			__antithesis_instrumentation__.Notify(245692)
		}
	}
	__antithesis_instrumentation__.Notify(245661)

	writeBatch := txn.Run
	if alsoCommit {
		__antithesis_instrumentation__.Notify(245693)
		writeBatch = txn.CommitInBatch
	} else {
		__antithesis_instrumentation__.Notify(245694)
	}
	__antithesis_instrumentation__.Notify(245662)
	if err := writeBatch(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(245695)
		return roachpb.Key{}, ConvertBackfillError(ctx, tableDesc, b)
	} else {
		__antithesis_instrumentation__.Notify(245696)
	}
	__antithesis_instrumentation__.Notify(245663)
	return cb.fetcher.Key(), nil
}

func ConvertBackfillError(
	ctx context.Context, tableDesc catalog.TableDescriptor, b *kv.Batch,
) error {
	__antithesis_instrumentation__.Notify(245697)

	desc, err := tableDesc.MakeFirstMutationPublic(catalog.IncludeConstraints)
	if err != nil {
		__antithesis_instrumentation__.Notify(245699)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245700)
	}
	__antithesis_instrumentation__.Notify(245698)
	return row.ConvertBatchError(ctx, desc, b)
}

type muBoundAccount struct {
	syncutil.Mutex

	boundAccount mon.BoundAccount
}

type IndexBackfiller struct {
	added []catalog.Index

	colIdxMap catalog.TableColMap

	types   []*types.T
	rowVals tree.Datums
	evalCtx *tree.EvalContext

	cols []catalog.Column

	addedCols []catalog.Column

	computedCols []catalog.Column

	colExprs map[descpb.ColumnID]tree.TypedExpr

	predicates map[descpb.IndexID]tree.TypedExpr

	indexesToEncode []catalog.Index

	valNeededForCol util.FastIntSet

	alloc tree.DatumAlloc

	mon            *mon.BytesMonitor
	muBoundAccount muBoundAccount
}

func (ib *IndexBackfiller) ContainsInvertedIndex() bool {
	__antithesis_instrumentation__.Notify(245701)
	for _, idx := range ib.added {
		__antithesis_instrumentation__.Notify(245703)
		if idx.GetType() == descpb.IndexDescriptor_INVERTED {
			__antithesis_instrumentation__.Notify(245704)
			return true
		} else {
			__antithesis_instrumentation__.Notify(245705)
		}
	}
	__antithesis_instrumentation__.Notify(245702)
	return false
}

func (ib *IndexBackfiller) InitForLocalUse(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
	desc catalog.TableDescriptor,
	mon *mon.BytesMonitor,
) error {
	__antithesis_instrumentation__.Notify(245706)

	ib.initCols(desc)

	ib.valNeededForCol = ib.initIndexes(desc)

	predicates, colExprs, referencedColumns, err := constructExprs(
		ctx, desc, ib.added, ib.cols, ib.addedCols, ib.computedCols, evalCtx, semaCtx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245709)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245710)
	}
	__antithesis_instrumentation__.Notify(245707)

	referencedColumns.ForEach(func(col descpb.ColumnID) {
		__antithesis_instrumentation__.Notify(245711)
		ib.valNeededForCol.Add(ib.colIdxMap.GetDefault(col))
	})
	__antithesis_instrumentation__.Notify(245708)

	return ib.init(evalCtx, predicates, colExprs, mon)
}

func constructExprs(
	ctx context.Context,
	desc catalog.TableDescriptor,
	addedIndexes []catalog.Index,
	cols, addedCols, computedCols []catalog.Column,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
) (
	predicates map[descpb.IndexID]tree.TypedExpr,
	colExprs map[descpb.ColumnID]tree.TypedExpr,
	referencedColumns catalog.TableColSet,
	_ error,
) {
	__antithesis_instrumentation__.Notify(245712)

	predicates, predicateRefColIDs, err := schemaexpr.MakePartialIndexExprs(
		ctx,
		addedIndexes,
		cols,
		desc,
		evalCtx,
		semaCtx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245721)
		return nil, nil, catalog.TableColSet{}, err
	} else {
		__antithesis_instrumentation__.Notify(245722)
	}
	__antithesis_instrumentation__.Notify(245713)

	defaultExprs, err := schemaexpr.MakeDefaultExprs(
		ctx, addedCols, &transform.ExprTransformContext{}, evalCtx, semaCtx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245723)
		return nil, nil, catalog.TableColSet{}, err
	} else {
		__antithesis_instrumentation__.Notify(245724)
	}
	__antithesis_instrumentation__.Notify(245714)

	tn := tree.NewUnqualifiedTableName(tree.Name(desc.GetName()))
	computedExprs, computedExprRefColIDs, err := schemaexpr.MakeComputedExprs(
		ctx,
		computedCols,
		cols,
		desc,
		tn,
		evalCtx,
		semaCtx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245725)
		return nil, nil, catalog.TableColSet{}, err
	} else {
		__antithesis_instrumentation__.Notify(245726)
	}
	__antithesis_instrumentation__.Notify(245715)

	numColExprs := len(addedCols) + len(computedCols)
	colExprs = make(map[descpb.ColumnID]tree.TypedExpr, numColExprs)
	var addedColSet catalog.TableColSet
	for i := range defaultExprs {
		__antithesis_instrumentation__.Notify(245727)
		id := addedCols[i].GetID()
		colExprs[id] = defaultExprs[i]
		addedColSet.Add(id)
	}
	__antithesis_instrumentation__.Notify(245716)
	for i := range computedCols {
		__antithesis_instrumentation__.Notify(245728)
		id := computedCols[i].GetID()
		colExprs[id] = computedExprs[i]
	}
	__antithesis_instrumentation__.Notify(245717)

	addToReferencedColumns := func(cols catalog.TableColSet) error {
		__antithesis_instrumentation__.Notify(245729)
		for colID, ok := cols.Next(0); ok; colID, ok = cols.Next(colID + 1) {
			__antithesis_instrumentation__.Notify(245731)
			if addedColSet.Contains(colID) {
				__antithesis_instrumentation__.Notify(245735)
				continue
			} else {
				__antithesis_instrumentation__.Notify(245736)
			}
			__antithesis_instrumentation__.Notify(245732)
			col, err := desc.FindColumnWithID(colID)
			if err != nil {
				__antithesis_instrumentation__.Notify(245737)
				return errors.AssertionFailedf("column %d does not exist", colID)
			} else {
				__antithesis_instrumentation__.Notify(245738)
			}
			__antithesis_instrumentation__.Notify(245733)
			if col.IsVirtual() {
				__antithesis_instrumentation__.Notify(245739)
				continue
			} else {
				__antithesis_instrumentation__.Notify(245740)
			}
			__antithesis_instrumentation__.Notify(245734)
			referencedColumns.Add(colID)
		}
		__antithesis_instrumentation__.Notify(245730)
		return nil
	}
	__antithesis_instrumentation__.Notify(245718)
	if err := addToReferencedColumns(predicateRefColIDs); err != nil {
		__antithesis_instrumentation__.Notify(245741)
		return nil, nil, catalog.TableColSet{}, err
	} else {
		__antithesis_instrumentation__.Notify(245742)
	}
	__antithesis_instrumentation__.Notify(245719)
	if err := addToReferencedColumns(computedExprRefColIDs); err != nil {
		__antithesis_instrumentation__.Notify(245743)
		return nil, nil, catalog.TableColSet{}, err
	} else {
		__antithesis_instrumentation__.Notify(245744)
	}
	__antithesis_instrumentation__.Notify(245720)
	return predicates, colExprs, referencedColumns, nil
}

func (ib *IndexBackfiller) InitForDistributedUse(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	desc catalog.TableDescriptor,
	mon *mon.BytesMonitor,
) error {
	__antithesis_instrumentation__.Notify(245745)

	ib.initCols(desc)

	ib.valNeededForCol = ib.initIndexes(desc)

	evalCtx := flowCtx.NewEvalCtx()
	var predicates map[descpb.IndexID]tree.TypedExpr
	var colExprs map[descpb.ColumnID]tree.TypedExpr
	var referencedColumns catalog.TableColSet

	if err := flowCtx.Cfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(245748)
		resolver := flowCtx.NewTypeResolver(txn)

		if err = typedesc.HydrateTypesInTableDescriptor(
			ctx, desc.TableDesc(), &resolver,
		); err != nil {
			__antithesis_instrumentation__.Notify(245750)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245751)
		}
		__antithesis_instrumentation__.Notify(245749)

		semaCtx := tree.MakeSemaContext()
		semaCtx.TypeResolver = &resolver

		predicates, colExprs, referencedColumns, err = constructExprs(
			ctx, desc, ib.added, ib.cols, ib.addedCols, ib.computedCols, evalCtx, &semaCtx,
		)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(245752)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245753)
	}
	__antithesis_instrumentation__.Notify(245746)

	flowCtx.Descriptors.ReleaseAll(ctx)

	referencedColumns.ForEach(func(col descpb.ColumnID) {
		__antithesis_instrumentation__.Notify(245754)
		ib.valNeededForCol.Add(ib.colIdxMap.GetDefault(col))
	})
	__antithesis_instrumentation__.Notify(245747)

	return ib.init(evalCtx, predicates, colExprs, mon)
}

func (ib *IndexBackfiller) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(245755)
	if ib.mon != nil {
		__antithesis_instrumentation__.Notify(245756)
		ib.muBoundAccount.Lock()
		ib.muBoundAccount.boundAccount.Close(ctx)
		ib.muBoundAccount.Unlock()
		ib.mon.Stop(ctx)
	} else {
		__antithesis_instrumentation__.Notify(245757)
	}
}

func (ib *IndexBackfiller) GrowBoundAccount(ctx context.Context, growBy int64) error {
	__antithesis_instrumentation__.Notify(245758)
	defer ib.muBoundAccount.Unlock()
	ib.muBoundAccount.Lock()
	err := ib.muBoundAccount.boundAccount.Grow(ctx, growBy)
	return err
}

func (ib *IndexBackfiller) ShrinkBoundAccount(ctx context.Context, shrinkBy int64) {
	__antithesis_instrumentation__.Notify(245759)
	defer ib.muBoundAccount.Unlock()
	ib.muBoundAccount.Lock()
	ib.muBoundAccount.boundAccount.Shrink(ctx, shrinkBy)
}

func (ib *IndexBackfiller) initCols(desc catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(245760)
	ib.cols = make([]catalog.Column, 0, len(desc.DeletableColumns()))
	for _, column := range desc.DeletableColumns() {
		__antithesis_instrumentation__.Notify(245761)
		if column.Public() {
			__antithesis_instrumentation__.Notify(245763)
			if column.IsComputed() && func() bool {
				__antithesis_instrumentation__.Notify(245764)
				return column.IsVirtual() == true
			}() == true {
				__antithesis_instrumentation__.Notify(245765)
				ib.computedCols = append(ib.computedCols, column)
			} else {
				__antithesis_instrumentation__.Notify(245766)
			}
		} else {
			__antithesis_instrumentation__.Notify(245767)
			if column.Adding() && func() bool {
				__antithesis_instrumentation__.Notify(245768)
				return column.WriteAndDeleteOnly() == true
			}() == true {
				__antithesis_instrumentation__.Notify(245769)

				if column.IsComputed() {
					__antithesis_instrumentation__.Notify(245770)
					ib.computedCols = append(ib.computedCols, column)
				} else {
					__antithesis_instrumentation__.Notify(245771)
					ib.addedCols = append(ib.addedCols, column)
				}
			} else {
				__antithesis_instrumentation__.Notify(245772)
				continue
			}
		}
		__antithesis_instrumentation__.Notify(245762)

		ib.colIdxMap.Set(column.GetID(), len(ib.cols))
		ib.cols = append(ib.cols, column)
	}
}

func (ib *IndexBackfiller) initIndexes(desc catalog.TableDescriptor) util.FastIntSet {
	__antithesis_instrumentation__.Notify(245773)
	var valNeededForCol util.FastIntSet
	mutations := desc.AllMutations()
	mutationID := mutations[0].MutationID()

	for _, m := range mutations {
		__antithesis_instrumentation__.Notify(245775)
		if m.MutationID() != mutationID {
			__antithesis_instrumentation__.Notify(245777)
			break
		} else {
			__antithesis_instrumentation__.Notify(245778)
		}
		__antithesis_instrumentation__.Notify(245776)
		if IndexMutationFilter(m) {
			__antithesis_instrumentation__.Notify(245779)
			idx := m.AsIndex()
			colIDs := idx.CollectKeyColumnIDs()
			if idx.GetEncodingType() == descpb.PrimaryIndexEncoding {
				__antithesis_instrumentation__.Notify(245781)
				for _, col := range ib.cols {
					__antithesis_instrumentation__.Notify(245782)
					if !col.IsVirtual() {
						__antithesis_instrumentation__.Notify(245783)
						colIDs.Add(col.GetID())
					} else {
						__antithesis_instrumentation__.Notify(245784)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(245785)
				colIDs.UnionWith(idx.CollectSecondaryStoredColumnIDs())
				colIDs.UnionWith(idx.CollectKeySuffixColumnIDs())
			}
			__antithesis_instrumentation__.Notify(245780)

			ib.added = append(ib.added, idx)
			for i := range ib.cols {
				__antithesis_instrumentation__.Notify(245786)
				id := ib.cols[i].GetID()
				if colIDs.Contains(id) && func() bool {
					__antithesis_instrumentation__.Notify(245787)
					return i < len(desc.PublicColumns()) == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(245788)
					return !ib.cols[i].IsVirtual() == true
				}() == true {
					__antithesis_instrumentation__.Notify(245789)
					valNeededForCol.Add(i)
				} else {
					__antithesis_instrumentation__.Notify(245790)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(245791)
		}
	}
	__antithesis_instrumentation__.Notify(245774)

	return valNeededForCol
}

func (ib *IndexBackfiller) init(
	evalCtx *tree.EvalContext,
	predicateExprs map[descpb.IndexID]tree.TypedExpr,
	colExprs map[descpb.ColumnID]tree.TypedExpr,
	mon *mon.BytesMonitor,
) error {
	ib.evalCtx = evalCtx
	ib.predicates = predicateExprs
	ib.colExprs = colExprs

	ib.indexesToEncode = ib.added
	if len(ib.predicates) > 0 {
		ib.indexesToEncode = make([]catalog.Index, 0, len(ib.added))
	}

	ib.types = make([]*types.T, len(ib.cols))
	for i := range ib.cols {
		ib.types[i] = ib.cols[i].GetType()
	}

	if mon == nil {
		return errors.AssertionFailedf("no memory monitor linked to IndexBackfiller during init")
	}
	ib.mon = mon
	ib.muBoundAccount.boundAccount = mon.MakeBoundAccount()
	return nil
}

func (ib *IndexBackfiller) BuildIndexEntriesChunk(
	ctx context.Context,
	txn *kv.Txn,
	tableDesc catalog.TableDescriptor,
	sp roachpb.Span,
	chunkSize int64,
	traceKV bool,
) ([]rowenc.IndexEntry, roachpb.Key, int64, error) {
	__antithesis_instrumentation__.Notify(245792)

	const initBufferSize = 1000
	const sizeOfIndexEntry = int64(unsafe.Sizeof(rowenc.IndexEntry{}))
	var memUsedPerChunk int64

	indexEntriesInChunkInitialBufferSize :=
		sizeOfIndexEntry * initBufferSize * int64(len(ib.added))
	if err := ib.GrowBoundAccount(ctx, indexEntriesInChunkInitialBufferSize); err != nil {
		__antithesis_instrumentation__.Notify(245803)
		return nil, nil, 0, errors.Wrap(err,
			"failed to initialize empty buffer to store the index entries of all rows in the chunk")
	} else {
		__antithesis_instrumentation__.Notify(245804)
	}
	__antithesis_instrumentation__.Notify(245793)
	memUsedPerChunk += indexEntriesInChunkInitialBufferSize
	entries := make([]rowenc.IndexEntry, 0, initBufferSize*int64(len(ib.added)))

	var fetcherCols []descpb.ColumnID
	for i, c := range ib.cols {
		__antithesis_instrumentation__.Notify(245805)
		if ib.valNeededForCol.Contains(i) {
			__antithesis_instrumentation__.Notify(245806)
			fetcherCols = append(fetcherCols, c.GetID())
		} else {
			__antithesis_instrumentation__.Notify(245807)
		}
	}
	__antithesis_instrumentation__.Notify(245794)
	if ib.rowVals == nil {
		__antithesis_instrumentation__.Notify(245808)
		ib.rowVals = make(tree.Datums, len(ib.cols))

		for i := range ib.rowVals {
			__antithesis_instrumentation__.Notify(245809)
			ib.rowVals[i] = tree.DNull
		}
	} else {
		__antithesis_instrumentation__.Notify(245810)
	}
	__antithesis_instrumentation__.Notify(245795)

	var spec descpb.IndexFetchSpec
	if err := rowenc.InitIndexFetchSpec(
		&spec, ib.evalCtx.Codec, tableDesc, tableDesc.GetPrimaryIndex(), fetcherCols,
	); err != nil {
		__antithesis_instrumentation__.Notify(245811)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(245812)
	}
	__antithesis_instrumentation__.Notify(245796)
	var fetcher row.Fetcher
	if err := fetcher.Init(
		ib.evalCtx.Context,
		false,
		descpb.ScanLockingStrength_FOR_NONE,
		descpb.ScanLockingWaitPolicy_BLOCK,
		0,
		&ib.alloc,
		ib.mon,
		&spec,
	); err != nil {
		__antithesis_instrumentation__.Notify(245813)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(245814)
	}
	__antithesis_instrumentation__.Notify(245797)
	defer fetcher.Close(ctx)
	if err := fetcher.StartScan(
		ctx, txn, []roachpb.Span{sp}, rowinfra.DefaultBatchBytesLimit, initBufferSize,
		traceKV, false,
	); err != nil {
		__antithesis_instrumentation__.Notify(245815)
		log.Errorf(ctx, "scan error: %s", err)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(245816)
	}
	__antithesis_instrumentation__.Notify(245798)

	iv := &schemaexpr.RowIndexedVarContainer{
		Cols:    ib.cols,
		Mapping: ib.colIdxMap,
	}
	ib.evalCtx.IVarContainer = iv

	indexEntriesPerRowInitialBufferSize := int64(len(ib.added)) * sizeOfIndexEntry
	if err := ib.GrowBoundAccount(ctx, indexEntriesPerRowInitialBufferSize); err != nil {
		__antithesis_instrumentation__.Notify(245817)
		return nil, nil, 0, errors.Wrap(err,
			"failed to initialize empty buffer to store the index entries of a single row")
	} else {
		__antithesis_instrumentation__.Notify(245818)
	}
	__antithesis_instrumentation__.Notify(245799)
	memUsedPerChunk += indexEntriesPerRowInitialBufferSize
	buffer := make([]rowenc.IndexEntry, len(ib.added))
	evaluateExprs := func(cols []catalog.Column) error {
		__antithesis_instrumentation__.Notify(245819)
		for i := range cols {
			__antithesis_instrumentation__.Notify(245821)
			colID := cols[i].GetID()
			texpr, ok := ib.colExprs[colID]
			if !ok {
				__antithesis_instrumentation__.Notify(245825)
				continue
			} else {
				__antithesis_instrumentation__.Notify(245826)
			}
			__antithesis_instrumentation__.Notify(245822)
			val, err := texpr.Eval(ib.evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(245827)
				return err
			} else {
				__antithesis_instrumentation__.Notify(245828)
			}
			__antithesis_instrumentation__.Notify(245823)
			colIdx, ok := ib.colIdxMap.Get(colID)
			if !ok {
				__antithesis_instrumentation__.Notify(245829)
				return errors.AssertionFailedf(
					"failed to find index for column %d in %d",
					colID, tableDesc.GetID(),
				)
			} else {
				__antithesis_instrumentation__.Notify(245830)
			}
			__antithesis_instrumentation__.Notify(245824)
			ib.rowVals[colIdx] = val
		}
		__antithesis_instrumentation__.Notify(245820)
		return nil
	}
	__antithesis_instrumentation__.Notify(245800)
	for i := int64(0); i < chunkSize; i++ {
		__antithesis_instrumentation__.Notify(245831)
		ok, err := fetcher.NextRowDecodedInto(ctx, ib.rowVals, ib.colIdxMap)
		if err != nil {
			__antithesis_instrumentation__.Notify(245838)
			return nil, nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(245839)
		}
		__antithesis_instrumentation__.Notify(245832)
		if !ok {
			__antithesis_instrumentation__.Notify(245840)
			break
		} else {
			__antithesis_instrumentation__.Notify(245841)
		}
		__antithesis_instrumentation__.Notify(245833)
		iv.CurSourceRow = ib.rowVals

		if len(ib.colExprs) > 0 {
			__antithesis_instrumentation__.Notify(245842)
			if err := evaluateExprs(ib.addedCols); err != nil {
				__antithesis_instrumentation__.Notify(245844)
				return nil, nil, 0, err
			} else {
				__antithesis_instrumentation__.Notify(245845)
			}
			__antithesis_instrumentation__.Notify(245843)
			if err := evaluateExprs(ib.computedCols); err != nil {
				__antithesis_instrumentation__.Notify(245846)
				return nil, nil, 0, err
			} else {
				__antithesis_instrumentation__.Notify(245847)
			}
		} else {
			__antithesis_instrumentation__.Notify(245848)
		}
		__antithesis_instrumentation__.Notify(245834)

		if len(ib.predicates) > 0 {
			__antithesis_instrumentation__.Notify(245849)
			ib.indexesToEncode = ib.indexesToEncode[:0]
			for _, idx := range ib.added {
				__antithesis_instrumentation__.Notify(245850)
				if !idx.IsPartial() {
					__antithesis_instrumentation__.Notify(245853)

					ib.indexesToEncode = append(ib.indexesToEncode, idx)
					continue
				} else {
					__antithesis_instrumentation__.Notify(245854)
				}
				__antithesis_instrumentation__.Notify(245851)

				texpr := ib.predicates[idx.GetID()]

				val, err := texpr.Eval(ib.evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(245855)
					return nil, nil, 0, err
				} else {
					__antithesis_instrumentation__.Notify(245856)
				}
				__antithesis_instrumentation__.Notify(245852)

				if val == tree.DBoolTrue {
					__antithesis_instrumentation__.Notify(245857)
					ib.indexesToEncode = append(ib.indexesToEncode, idx)
				} else {
					__antithesis_instrumentation__.Notify(245858)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(245859)
		}
		__antithesis_instrumentation__.Notify(245835)

		buffer = buffer[:0]

		var memUsedDuringEncoding int64
		ib.muBoundAccount.Lock()
		if buffer, memUsedDuringEncoding, err = rowenc.EncodeSecondaryIndexes(
			ctx,
			ib.evalCtx.Codec,
			tableDesc,
			ib.indexesToEncode,
			ib.colIdxMap,
			ib.rowVals,
			buffer,
			false,
			&ib.muBoundAccount.boundAccount,
		); err != nil {
			__antithesis_instrumentation__.Notify(245860)
			ib.muBoundAccount.Unlock()
			return nil, nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(245861)
		}
		__antithesis_instrumentation__.Notify(245836)
		ib.muBoundAccount.Unlock()
		memUsedPerChunk += memUsedDuringEncoding

		if cap(entries)-len(entries) < len(buffer) {
			__antithesis_instrumentation__.Notify(245862)
			resliceSize := sizeOfIndexEntry * int64(cap(entries))
			if err := ib.GrowBoundAccount(ctx, resliceSize); err != nil {
				__antithesis_instrumentation__.Notify(245864)
				return nil, nil, 0, err
			} else {
				__antithesis_instrumentation__.Notify(245865)
			}
			__antithesis_instrumentation__.Notify(245863)
			memUsedPerChunk += resliceSize
		} else {
			__antithesis_instrumentation__.Notify(245866)
		}
		__antithesis_instrumentation__.Notify(245837)

		entries = append(entries, buffer...)
	}
	__antithesis_instrumentation__.Notify(245801)

	shrinkSize := sizeOfIndexEntry * int64(cap(buffer))
	ib.ShrinkBoundAccount(ctx, shrinkSize)
	memUsedPerChunk -= shrinkSize

	var resumeKey roachpb.Key
	if fetcher.Key() != nil {
		__antithesis_instrumentation__.Notify(245867)
		resumeKey = make(roachpb.Key, len(fetcher.Key()))
		copy(resumeKey, fetcher.Key())
	} else {
		__antithesis_instrumentation__.Notify(245868)
	}
	__antithesis_instrumentation__.Notify(245802)
	return entries, resumeKey, memUsedPerChunk, nil
}

func (ib *IndexBackfiller) RunIndexBackfillChunk(
	ctx context.Context,
	txn *kv.Txn,
	tableDesc catalog.TableDescriptor,
	sp roachpb.Span,
	chunkSize int64,
	alsoCommit bool,
	traceKV bool,
) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(245869)
	entries, key, memUsedBuildingChunk, err := ib.BuildIndexEntriesChunk(ctx, txn, tableDesc, sp,
		chunkSize, traceKV)
	if err != nil {
		__antithesis_instrumentation__.Notify(245874)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245875)
	}
	__antithesis_instrumentation__.Notify(245870)
	batch := txn.NewBatch()

	for _, entry := range entries {
		__antithesis_instrumentation__.Notify(245876)
		if traceKV {
			__antithesis_instrumentation__.Notify(245878)
			log.VEventf(ctx, 2, "InitPut %s -> %s", entry.Key, entry.Value.PrettyPrint())
		} else {
			__antithesis_instrumentation__.Notify(245879)
		}
		__antithesis_instrumentation__.Notify(245877)
		batch.InitPut(entry.Key, &entry.Value, false)
	}
	__antithesis_instrumentation__.Notify(245871)
	writeBatch := txn.Run
	if alsoCommit {
		__antithesis_instrumentation__.Notify(245880)
		writeBatch = txn.CommitInBatch
	} else {
		__antithesis_instrumentation__.Notify(245881)
	}
	__antithesis_instrumentation__.Notify(245872)
	if err := writeBatch(ctx, batch); err != nil {
		__antithesis_instrumentation__.Notify(245882)
		return nil, ConvertBackfillError(ctx, tableDesc, batch)
	} else {
		__antithesis_instrumentation__.Notify(245883)
	}
	__antithesis_instrumentation__.Notify(245873)

	entries = nil
	ib.ShrinkBoundAccount(ctx, memUsedBuildingChunk)

	return key, nil
}
