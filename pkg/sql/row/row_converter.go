package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/transform"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type KVInserter func(roachpb.KeyValue)

func (i KVInserter) CPut(key, value interface{}, expValue []byte) {
	__antithesis_instrumentation__.Notify(568502)
	panic("unimplemented")
}

func (i KVInserter) Del(key ...interface{}) {
	__antithesis_instrumentation__.Notify(568503)

}

func (i KVInserter) Put(key, value interface{}) {
	__antithesis_instrumentation__.Notify(568504)
	i(roachpb.KeyValue{
		Key:   *key.(*roachpb.Key),
		Value: *value.(*roachpb.Value),
	})
}

func (i KVInserter) InitPut(key, value interface{}, failOnTombstones bool) {
	__antithesis_instrumentation__.Notify(568505)
	i(roachpb.KeyValue{
		Key:   *key.(*roachpb.Key),
		Value: *value.(*roachpb.Value),
	})
}

func GenerateInsertRow(
	defaultExprs []tree.TypedExpr,
	computeExprs []tree.TypedExpr,
	insertCols []catalog.Column,
	computedColsLookup []catalog.Column,
	evalCtx *tree.EvalContext,
	tableDesc catalog.TableDescriptor,
	rowVals tree.Datums,
	rowContainerForComputedVals *schemaexpr.RowIndexedVarContainer,
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(568506)

	if len(rowVals) < len(insertCols) {
		__antithesis_instrumentation__.Notify(568511)

		oldVals := rowVals
		rowVals = make(tree.Datums, len(insertCols))
		copy(rowVals, oldVals)

		for i := len(oldVals); i < len(insertCols); i++ {
			__antithesis_instrumentation__.Notify(568512)
			if defaultExprs == nil {
				__antithesis_instrumentation__.Notify(568515)
				rowVals[i] = tree.DNull
				continue
			} else {
				__antithesis_instrumentation__.Notify(568516)
			}
			__antithesis_instrumentation__.Notify(568513)
			d, err := defaultExprs[i].Eval(evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(568517)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(568518)
			}
			__antithesis_instrumentation__.Notify(568514)
			rowVals[i] = d
		}
	} else {
		__antithesis_instrumentation__.Notify(568519)
	}
	__antithesis_instrumentation__.Notify(568507)

	if len(computeExprs) > 0 {
		__antithesis_instrumentation__.Notify(568520)
		rowContainerForComputedVals.CurSourceRow = rowVals
		evalCtx.PushIVarContainer(rowContainerForComputedVals)
		for i := range computedColsLookup {
			__antithesis_instrumentation__.Notify(568522)

			col := computedColsLookup[i]
			computeIdx := rowContainerForComputedVals.Mapping.GetDefault(col.GetID())
			if !col.IsComputed() {
				__antithesis_instrumentation__.Notify(568525)
				continue
			} else {
				__antithesis_instrumentation__.Notify(568526)
			}
			__antithesis_instrumentation__.Notify(568523)
			d, err := computeExprs[computeIdx].Eval(evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(568527)
				name := col.GetName()
				return nil, errors.Wrapf(err,
					"computed column %s",
					tree.ErrString((*tree.Name)(&name)))
			} else {
				__antithesis_instrumentation__.Notify(568528)
			}
			__antithesis_instrumentation__.Notify(568524)
			rowVals[computeIdx] = d
		}
		__antithesis_instrumentation__.Notify(568521)
		evalCtx.PopIVarContainer()
	} else {
		__antithesis_instrumentation__.Notify(568529)
	}
	__antithesis_instrumentation__.Notify(568508)

	for _, col := range tableDesc.WritableColumns() {
		__antithesis_instrumentation__.Notify(568530)
		if !col.IsNullable() {
			__antithesis_instrumentation__.Notify(568531)
			if i, ok := rowContainerForComputedVals.Mapping.Get(col.GetID()); !ok || func() bool {
				__antithesis_instrumentation__.Notify(568532)
				return rowVals[i] == tree.DNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(568533)
				return nil, sqlerrors.NewNonNullViolationError(col.GetName())
			} else {
				__antithesis_instrumentation__.Notify(568534)
			}
		} else {
			__antithesis_instrumentation__.Notify(568535)
		}
	}
	__antithesis_instrumentation__.Notify(568509)

	for i := 0; i < len(insertCols); i++ {
		__antithesis_instrumentation__.Notify(568536)
		outVal, err := tree.AdjustValueToType(insertCols[i].GetType(), rowVals[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(568538)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568539)
		}
		__antithesis_instrumentation__.Notify(568537)
		rowVals[i] = outVal
	}
	__antithesis_instrumentation__.Notify(568510)

	return rowVals, nil
}

type KVBatch struct {
	Source int32

	LastRow int64

	Progress float32

	KVs     []roachpb.KeyValue
	MemSize int64
}

type DatumRowConverter struct {
	Datums []tree.Datum

	KvCh     chan<- KVBatch
	KvBatch  KVBatch
	BatchCap int

	tableDesc catalog.TableDescriptor

	TargetColOrds util.FastIntSet

	ri                    Inserter
	EvalCtx               *tree.EvalContext
	cols                  []catalog.Column
	VisibleCols           []catalog.Column
	VisibleColTypes       []*types.T
	computedExprs         []tree.TypedExpr
	defaultCache          []tree.TypedExpr
	computedIVarContainer schemaexpr.RowIndexedVarContainer

	CompletedRowFn func() int64
	FractionFn     func() float32
}

var kvDatumRowConverterBatchSize = util.ConstantWithMetamorphicTestValue(
	"datum-row-converter-batch-size",
	5000,
	1,
)

const kvDatumRowConverterBatchMemSize = 4 << 20

func TestingSetDatumRowConverterBatchSize(newSize int) func() {
	__antithesis_instrumentation__.Notify(568540)
	oldSize := kvDatumRowConverterBatchSize
	kvDatumRowConverterBatchSize = newSize
	return func() {
		__antithesis_instrumentation__.Notify(568541)
		kvDatumRowConverterBatchSize = oldSize
	}
}

func (c *DatumRowConverter) getSequenceAnnotation(
	evalCtx *tree.EvalContext, cols []catalog.Column,
) (map[string]*SequenceMetadata, map[descpb.ID]*SequenceMetadata, error) {
	__antithesis_instrumentation__.Notify(568542)

	sequenceIDs := make(map[descpb.ID]struct{})
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(568546)
		for i := 0; i < col.NumUsesSequences(); i++ {
			__antithesis_instrumentation__.Notify(568547)
			id := col.GetUsesSequenceID(i)
			sequenceIDs[id] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(568543)

	if len(sequenceIDs) == 0 {
		__antithesis_instrumentation__.Notify(568548)
		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(568549)
	}
	__antithesis_instrumentation__.Notify(568544)

	var seqNameToMetadata map[string]*SequenceMetadata
	var seqIDToMetadata map[descpb.ID]*SequenceMetadata

	cf := descs.NewBareBonesCollectionFactory(evalCtx.Settings, evalCtx.Codec)
	descsCol := cf.MakeCollection(evalCtx.Context, descs.NewTemporarySchemaProvider(evalCtx.SessionDataStack))
	err := evalCtx.DB.Txn(evalCtx.Context, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(568550)
		seqNameToMetadata = make(map[string]*SequenceMetadata)
		seqIDToMetadata = make(map[descpb.ID]*SequenceMetadata)
		if err := txn.SetFixedTimestamp(ctx, hlc.Timestamp{WallTime: evalCtx.TxnTimestamp.UnixNano()}); err != nil {
			__antithesis_instrumentation__.Notify(568553)
			return err
		} else {
			__antithesis_instrumentation__.Notify(568554)
		}
		__antithesis_instrumentation__.Notify(568551)
		for seqID := range sequenceIDs {
			__antithesis_instrumentation__.Notify(568555)
			seqDesc, err := descsCol.Direct().MustGetTableDescByID(ctx, txn, seqID)
			if err != nil {
				__antithesis_instrumentation__.Notify(568558)
				return err
			} else {
				__antithesis_instrumentation__.Notify(568559)
			}
			__antithesis_instrumentation__.Notify(568556)
			if seqDesc.GetSequenceOpts() == nil {
				__antithesis_instrumentation__.Notify(568560)
				return errors.Errorf("relation %q (%d) is not a sequence", seqDesc.GetName(), seqDesc.GetID())
			} else {
				__antithesis_instrumentation__.Notify(568561)
			}
			__antithesis_instrumentation__.Notify(568557)
			seqMetadata := &SequenceMetadata{seqDesc: seqDesc}
			seqNameToMetadata[seqDesc.GetName()] = seqMetadata
			seqIDToMetadata[seqID] = seqMetadata
		}
		__antithesis_instrumentation__.Notify(568552)
		return nil
	})
	__antithesis_instrumentation__.Notify(568545)
	return seqNameToMetadata, seqIDToMetadata, err
}

func NewDatumRowConverter(
	ctx context.Context,
	baseSemaCtx *tree.SemaContext,
	tableDesc catalog.TableDescriptor,
	targetColNames tree.NameList,
	evalCtx *tree.EvalContext,
	kvCh chan<- KVBatch,
	seqChunkProvider *SeqChunkProvider,
	metrics *Metrics,
) (*DatumRowConverter, error) {
	__antithesis_instrumentation__.Notify(568562)
	c := &DatumRowConverter{
		tableDesc: tableDesc,
		KvCh:      kvCh,
		EvalCtx:   evalCtx.Copy(),
	}

	var targetCols []catalog.Column
	var err error

	if len(targetColNames) != 0 {
		__antithesis_instrumentation__.Notify(568574)
		if targetCols, err = colinfo.ProcessTargetColumns(tableDesc, targetColNames,
			true, false); err != nil {
			__antithesis_instrumentation__.Notify(568575)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568576)
		}
	} else {
		__antithesis_instrumentation__.Notify(568577)
		targetCols = tableDesc.VisibleColumns()
	}
	__antithesis_instrumentation__.Notify(568563)

	var targetColIDs catalog.TableColSet
	for i, col := range targetCols {
		__antithesis_instrumentation__.Notify(568578)
		c.TargetColOrds.Add(i)
		targetColIDs.Add(col.GetID())
	}
	__antithesis_instrumentation__.Notify(568564)

	var txCtx transform.ExprTransformContext
	relevantColumns := func(col catalog.Column) bool {
		__antithesis_instrumentation__.Notify(568579)
		return col.HasDefault() || func() bool {
			__antithesis_instrumentation__.Notify(568580)
			return col.IsComputed() == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(568565)

	semaCtxCopy := *baseSemaCtx
	cols := schemaexpr.ProcessColumnSet(targetCols, tableDesc, relevantColumns)
	defaultExprs, err := schemaexpr.MakeDefaultExprs(ctx, cols, &txCtx, c.EvalCtx, &semaCtxCopy)
	if err != nil {
		__antithesis_instrumentation__.Notify(568581)
		return nil, errors.Wrap(err, "process default and computed columns")
	} else {
		__antithesis_instrumentation__.Notify(568582)
	}
	__antithesis_instrumentation__.Notify(568566)

	ri, err := MakeInserter(
		ctx,
		nil,
		evalCtx.Codec,
		tableDesc,
		cols,
		&tree.DatumAlloc{},
		&evalCtx.Settings.SV,
		evalCtx.SessionData().Internal,
		metrics,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(568583)
		return nil, errors.Wrap(err, "make row inserter")
	} else {
		__antithesis_instrumentation__.Notify(568584)
	}
	__antithesis_instrumentation__.Notify(568567)

	c.ri = ri
	c.cols = cols

	c.VisibleCols = targetCols
	c.VisibleColTypes = make([]*types.T, len(c.VisibleCols))
	for i := range c.VisibleCols {
		__antithesis_instrumentation__.Notify(568585)
		c.VisibleColTypes[i] = c.VisibleCols[i].GetType()
	}
	__antithesis_instrumentation__.Notify(568568)

	c.Datums = make([]tree.Datum, len(targetCols), len(cols))
	c.defaultCache = make([]tree.TypedExpr, len(cols))

	annot := make(tree.Annotations, 1)
	var cellInfoAnnot CellInfoAnnotation

	if seqChunkProvider != nil {
		__antithesis_instrumentation__.Notify(568586)
		seqNameToMetadata, seqIDToMetadata, err := c.getSequenceAnnotation(evalCtx, c.cols)
		if err != nil {
			__antithesis_instrumentation__.Notify(568588)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568589)
		}
		__antithesis_instrumentation__.Notify(568587)
		cellInfoAnnot.seqNameToMetadata = seqNameToMetadata
		cellInfoAnnot.seqIDToMetadata = seqIDToMetadata
		cellInfoAnnot.seqChunkProvider = seqChunkProvider
	} else {
		__antithesis_instrumentation__.Notify(568590)
	}
	__antithesis_instrumentation__.Notify(568569)
	cellInfoAnnot.uniqueRowIDInstance = 0
	annot.Set(cellInfoAddr, &cellInfoAnnot)
	c.EvalCtx.Annotations = &annot

	for i, col := range cols {
		__antithesis_instrumentation__.Notify(568591)
		if col.HasDefault() {
			__antithesis_instrumentation__.Notify(568593)

			typedExpr, volatile, err := sanitizeExprsForImport(ctx, c.EvalCtx, defaultExprs[i], col.GetType())
			if err != nil {
				__antithesis_instrumentation__.Notify(568595)

				c.defaultCache[i] = &unsafeErrExpr{
					err: errors.Wrapf(err, "default expression %s unsafe for import", defaultExprs[i].String()),
				}
			} else {
				__antithesis_instrumentation__.Notify(568596)
				c.defaultCache[i] = typedExpr
				if volatile == overrideImmutable {
					__antithesis_instrumentation__.Notify(568597)

					c.defaultCache[i], err = c.defaultCache[i].Eval(c.EvalCtx)
					if err != nil {
						__antithesis_instrumentation__.Notify(568598)
						return nil, errors.Wrapf(err, "error evaluating default expression")
					} else {
						__antithesis_instrumentation__.Notify(568599)
					}
				} else {
					__antithesis_instrumentation__.Notify(568600)
				}
			}
			__antithesis_instrumentation__.Notify(568594)
			if !targetColIDs.Contains(col.GetID()) {
				__antithesis_instrumentation__.Notify(568601)
				c.Datums = append(c.Datums, nil)
			} else {
				__antithesis_instrumentation__.Notify(568602)
			}
		} else {
			__antithesis_instrumentation__.Notify(568603)
		}
		__antithesis_instrumentation__.Notify(568592)
		if col.IsComputed() && func() bool {
			__antithesis_instrumentation__.Notify(568604)
			return !targetColIDs.Contains(col.GetID()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(568605)
			c.Datums = append(c.Datums, nil)
		} else {
			__antithesis_instrumentation__.Notify(568606)
		}
	}
	__antithesis_instrumentation__.Notify(568570)
	if len(c.Datums) != len(cols) {
		__antithesis_instrumentation__.Notify(568607)
		return nil, errors.New("unexpected hidden column")
	} else {
		__antithesis_instrumentation__.Notify(568608)
	}
	__antithesis_instrumentation__.Notify(568571)

	padding := 2 * (len(tableDesc.PublicNonPrimaryIndexes()) + len(tableDesc.GetFamilies()))
	c.BatchCap = kvDatumRowConverterBatchSize + padding
	c.KvBatch.KVs = make([]roachpb.KeyValue, 0, c.BatchCap)
	c.KvBatch.MemSize = 0

	colsOrdered := make([]catalog.Column, len(cols))
	for _, col := range c.tableDesc.PublicColumns() {
		__antithesis_instrumentation__.Notify(568609)

		colsOrdered[ri.InsertColIDtoRowIndex.GetDefault(col.GetID())] = col
	}
	__antithesis_instrumentation__.Notify(568572)

	c.computedExprs, _, err = schemaexpr.MakeComputedExprs(
		ctx,
		colsOrdered,
		c.tableDesc.PublicColumns(),
		c.tableDesc,
		tree.NewUnqualifiedTableName(tree.Name(c.tableDesc.GetName())),
		c.EvalCtx,
		&semaCtxCopy)
	if err != nil {
		__antithesis_instrumentation__.Notify(568610)
		return nil, errors.Wrapf(err, "error evaluating computed expression for IMPORT INTO")
	} else {
		__antithesis_instrumentation__.Notify(568611)
	}
	__antithesis_instrumentation__.Notify(568573)

	c.computedIVarContainer = schemaexpr.RowIndexedVarContainer{
		Mapping: ri.InsertColIDtoRowIndex,
		Cols:    tableDesc.PublicColumns(),
	}
	return c, nil
}

const rowIDBits = 64 - builtins.NodeIDBits

func (c *DatumRowConverter) Row(ctx context.Context, sourceID int32, rowIndex int64) error {
	__antithesis_instrumentation__.Notify(568612)
	getCellInfoAnnotation(c.EvalCtx.Annotations).reset(sourceID, rowIndex)
	for i, col := range c.cols {
		__antithesis_instrumentation__.Notify(568618)
		if col.HasDefault() {
			__antithesis_instrumentation__.Notify(568619)

			datum, err := c.defaultCache[i].Eval(c.EvalCtx)
			if !c.TargetColOrds.Contains(i) {
				__antithesis_instrumentation__.Notify(568620)
				if err != nil {
					__antithesis_instrumentation__.Notify(568622)
					return errors.Wrapf(
						err, "error evaluating default expression %q", col.GetDefaultExpr())
				} else {
					__antithesis_instrumentation__.Notify(568623)
				}
				__antithesis_instrumentation__.Notify(568621)
				c.Datums[i] = datum
			} else {
				__antithesis_instrumentation__.Notify(568624)
			}
		} else {
			__antithesis_instrumentation__.Notify(568625)
		}
	}
	__antithesis_instrumentation__.Notify(568613)

	var computedColsLookup []catalog.Column
	if len(c.computedExprs) > 0 {
		__antithesis_instrumentation__.Notify(568626)
		computedColsLookup = c.tableDesc.PublicColumns()
	} else {
		__antithesis_instrumentation__.Notify(568627)
	}
	__antithesis_instrumentation__.Notify(568614)

	insertRow, err := GenerateInsertRow(
		c.defaultCache, c.computedExprs, c.cols, computedColsLookup, c.EvalCtx,
		c.tableDesc, c.Datums, &c.computedIVarContainer)
	if err != nil {
		__antithesis_instrumentation__.Notify(568628)
		return errors.Wrap(err, "generate insert row")
	} else {
		__antithesis_instrumentation__.Notify(568629)
	}
	__antithesis_instrumentation__.Notify(568615)

	var pm PartialIndexUpdateHelper
	if err := c.ri.InsertRow(
		ctx,
		KVInserter(func(kv roachpb.KeyValue) {
			__antithesis_instrumentation__.Notify(568630)
			kv.Value.InitChecksum(kv.Key)
			c.KvBatch.KVs = append(c.KvBatch.KVs, kv)
			c.KvBatch.MemSize += int64(cap(kv.Key) + cap(kv.Value.RawBytes))
		}),
		insertRow,
		pm,
		true,
		false,
	); err != nil {
		__antithesis_instrumentation__.Notify(568631)
		return errors.Wrap(err, "insert row")
	} else {
		__antithesis_instrumentation__.Notify(568632)
	}
	__antithesis_instrumentation__.Notify(568616)

	if len(c.KvBatch.KVs) >= kvDatumRowConverterBatchSize || func() bool {
		__antithesis_instrumentation__.Notify(568633)
		return c.KvBatch.MemSize > kvDatumRowConverterBatchMemSize == true
	}() == true {
		__antithesis_instrumentation__.Notify(568634)
		if err := c.SendBatch(ctx); err != nil {
			__antithesis_instrumentation__.Notify(568635)
			return err
		} else {
			__antithesis_instrumentation__.Notify(568636)
		}
	} else {
		__antithesis_instrumentation__.Notify(568637)
	}
	__antithesis_instrumentation__.Notify(568617)
	return nil
}

func (c *DatumRowConverter) SendBatch(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(568638)
	if len(c.KvBatch.KVs) == 0 {
		__antithesis_instrumentation__.Notify(568643)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(568644)
	}
	__antithesis_instrumentation__.Notify(568639)
	if c.FractionFn != nil {
		__antithesis_instrumentation__.Notify(568645)
		c.KvBatch.Progress = c.FractionFn()
	} else {
		__antithesis_instrumentation__.Notify(568646)
	}
	__antithesis_instrumentation__.Notify(568640)
	if c.CompletedRowFn != nil {
		__antithesis_instrumentation__.Notify(568647)
		c.KvBatch.LastRow = c.CompletedRowFn()
	} else {
		__antithesis_instrumentation__.Notify(568648)
	}
	__antithesis_instrumentation__.Notify(568641)
	select {
	case c.KvCh <- c.KvBatch:
		__antithesis_instrumentation__.Notify(568649)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(568650)
		return ctx.Err()
	}
	__antithesis_instrumentation__.Notify(568642)
	c.KvBatch.KVs = make([]roachpb.KeyValue, 0, c.BatchCap)
	c.KvBatch.MemSize = 0
	return nil
}
