package colfetcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/colconv"
	"github.com/cockroachdb/cockroach/pkg/sql/colencoding"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/kvstreamer"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/scrub"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type cTableInfo struct {
	*cFetcherTableArgs

	neededValueColsByIdx util.FastIntSet

	orderedColIdxMap *colIdxMap

	indexColOrdinals []int

	compositeIndexColOrdinals util.FastIntSet

	extraValColOrdinals []int

	rowLastModified hlc.Timestamp

	timestampOutputIdx int

	oidOutputIdx int

	da tree.DatumAlloc
}

var _ execinfra.Releasable = &cTableInfo{}

var cTableInfoPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(455175)
		return &cTableInfo{
			orderedColIdxMap: &colIdxMap{},
		}
	},
}

func newCTableInfo() *cTableInfo {
	__antithesis_instrumentation__.Notify(455176)
	return cTableInfoPool.Get().(*cTableInfo)
}

func (c *cTableInfo) Release() {
	__antithesis_instrumentation__.Notify(455177)
	c.cFetcherTableArgs.Release()

	c.orderedColIdxMap.ords = c.orderedColIdxMap.ords[:0]
	c.orderedColIdxMap.vals = c.orderedColIdxMap.vals[:0]
	*c = cTableInfo{
		orderedColIdxMap:    c.orderedColIdxMap,
		indexColOrdinals:    c.indexColOrdinals[:0],
		extraValColOrdinals: c.extraValColOrdinals[:0],
	}
	cTableInfoPool.Put(c)
}

type colIdxMap struct {
	vals descpb.ColumnIDs

	ords []int
}

func (m colIdxMap) Len() int {
	__antithesis_instrumentation__.Notify(455178)
	return len(m.vals)
}

func (m colIdxMap) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(455179)
	return m.vals[i] < m.vals[j]
}

func (m colIdxMap) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(455180)
	m.vals[i], m.vals[j] = m.vals[j], m.vals[i]
	m.ords[i], m.ords[j] = m.ords[j], m.ords[i]
}

type cFetcherArgs struct {
	lockStrength descpb.ScanLockingStrength

	lockWaitPolicy descpb.ScanLockingWaitPolicy

	lockTimeout time.Duration

	memoryLimit int64

	estimatedRowCount uint64

	reverse bool

	traceKV bool
}

const noOutputColumn = -1

type cFetcher struct {
	cFetcherArgs

	table *cTableInfo

	mustDecodeIndexKey bool

	mvccDecodeStrategy row.MVCCDecodingStrategy

	fetcher *row.KVFetcher

	bytesRead int64

	machine struct {
		state [3]fetcherState

		rowIdx int

		nextKV roachpb.KeyValue

		limitHint int

		remainingValueColsByIdx util.FastIntSet

		lastRowPrefix roachpb.Key

		prettyValueBuf *bytes.Buffer

		batch coldata.Batch

		colvecs coldata.TypedVecs

		timestampCol []apd.Decimal

		tableoidCol coldata.DatumVec
	}

	scratch []byte

	accountingHelper colmem.SetAccountingHelper

	kvFetcherMemAcc *mon.BoundAccount

	maxCapacity int
}

func (cf *cFetcher) resetBatch() {
	__antithesis_instrumentation__.Notify(455181)
	var reallocated bool
	var minDesiredCapacity int
	if cf.maxCapacity > 0 {
		__antithesis_instrumentation__.Notify(455183)

		minDesiredCapacity = cf.maxCapacity
	} else {
		__antithesis_instrumentation__.Notify(455184)
		if cf.machine.limitHint > 0 && func() bool {
			__antithesis_instrumentation__.Notify(455185)
			return (cf.estimatedRowCount == 0 || func() bool {
				__antithesis_instrumentation__.Notify(455186)
				return uint64(cf.machine.limitHint) < cf.estimatedRowCount == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(455187)

			minDesiredCapacity = cf.machine.limitHint
		} else {
			__antithesis_instrumentation__.Notify(455188)

			if cf.estimatedRowCount > uint64(coldata.BatchSize()) {
				__antithesis_instrumentation__.Notify(455189)
				minDesiredCapacity = coldata.BatchSize()
			} else {
				__antithesis_instrumentation__.Notify(455190)
				minDesiredCapacity = int(cf.estimatedRowCount)
			}
		}
	}
	__antithesis_instrumentation__.Notify(455182)
	cf.machine.batch, reallocated = cf.accountingHelper.ResetMaybeReallocate(
		cf.table.typs, cf.machine.batch, minDesiredCapacity, cf.memoryLimit,
	)
	if reallocated {
		__antithesis_instrumentation__.Notify(455191)
		cf.machine.colvecs.SetBatch(cf.machine.batch)

		if cf.table.timestampOutputIdx != noOutputColumn {
			__antithesis_instrumentation__.Notify(455194)
			cf.machine.timestampCol = cf.machine.colvecs.DecimalCols[cf.machine.colvecs.ColsMap[cf.table.timestampOutputIdx]]
		} else {
			__antithesis_instrumentation__.Notify(455195)
		}
		__antithesis_instrumentation__.Notify(455192)
		if cf.table.oidOutputIdx != noOutputColumn {
			__antithesis_instrumentation__.Notify(455196)
			cf.machine.tableoidCol = cf.machine.colvecs.DatumCols[cf.machine.colvecs.ColsMap[cf.table.oidOutputIdx]]
		} else {
			__antithesis_instrumentation__.Notify(455197)
		}
		__antithesis_instrumentation__.Notify(455193)

		cf.table.da.AllocSize = cf.machine.batch.Capacity()
	} else {
		__antithesis_instrumentation__.Notify(455198)
	}
}

func (cf *cFetcher) Init(
	allocator *colmem.Allocator, kvFetcherMemAcc *mon.BoundAccount, tableArgs *cFetcherTableArgs,
) error {
	__antithesis_instrumentation__.Notify(455199)
	if tableArgs.spec.Version != descpb.IndexFetchSpecVersionInitial {
		__antithesis_instrumentation__.Notify(455209)
		return errors.Newf("unsupported IndexFetchSpec version %d", tableArgs.spec.Version)
	} else {
		__antithesis_instrumentation__.Notify(455210)
	}
	__antithesis_instrumentation__.Notify(455200)
	cf.kvFetcherMemAcc = kvFetcherMemAcc
	table := newCTableInfo()
	nCols := tableArgs.ColIdxMap.Len()
	if cap(table.orderedColIdxMap.vals) < nCols {
		__antithesis_instrumentation__.Notify(455211)
		table.orderedColIdxMap.vals = make(descpb.ColumnIDs, 0, nCols)
		table.orderedColIdxMap.ords = make([]int, 0, nCols)
	} else {
		__antithesis_instrumentation__.Notify(455212)
	}
	__antithesis_instrumentation__.Notify(455201)
	for i := range tableArgs.spec.FetchedColumns {
		__antithesis_instrumentation__.Notify(455213)
		id := tableArgs.spec.FetchedColumns[i].ColumnID
		table.orderedColIdxMap.vals = append(table.orderedColIdxMap.vals, id)
		table.orderedColIdxMap.ords = append(table.orderedColIdxMap.ords, tableArgs.ColIdxMap.GetDefault(id))
	}
	__antithesis_instrumentation__.Notify(455202)
	sort.Sort(table.orderedColIdxMap)
	*table = cTableInfo{
		cFetcherTableArgs:   tableArgs,
		orderedColIdxMap:    table.orderedColIdxMap,
		indexColOrdinals:    table.indexColOrdinals[:0],
		extraValColOrdinals: table.extraValColOrdinals[:0],
		timestampOutputIdx:  noOutputColumn,
		oidOutputIdx:        noOutputColumn,
	}

	if nCols > 0 {
		__antithesis_instrumentation__.Notify(455214)
		table.neededValueColsByIdx.AddRange(0, nCols-1)
	} else {
		__antithesis_instrumentation__.Notify(455215)
	}
	__antithesis_instrumentation__.Notify(455203)

	for idx := range tableArgs.spec.FetchedColumns {
		__antithesis_instrumentation__.Notify(455216)
		colID := tableArgs.spec.FetchedColumns[idx].ColumnID
		if colinfo.IsColIDSystemColumn(colID) {
			__antithesis_instrumentation__.Notify(455217)

			switch colinfo.GetSystemColumnKindFromColumnID(colID) {
			case catpb.SystemColumnKind_MVCCTIMESTAMP:
				__antithesis_instrumentation__.Notify(455218)
				table.timestampOutputIdx = idx
				cf.mvccDecodeStrategy = row.MVCCDecodingRequired
				table.neededValueColsByIdx.Remove(idx)
			case catpb.SystemColumnKind_TABLEOID:
				__antithesis_instrumentation__.Notify(455219)
				table.oidOutputIdx = idx
				table.neededValueColsByIdx.Remove(idx)
			default:
				__antithesis_instrumentation__.Notify(455220)
			}
		} else {
			__antithesis_instrumentation__.Notify(455221)
		}
	}
	__antithesis_instrumentation__.Notify(455204)

	fullColumns := table.spec.KeyFullColumns()

	nIndexCols := len(fullColumns)
	if cap(table.indexColOrdinals) >= nIndexCols {
		__antithesis_instrumentation__.Notify(455222)
		table.indexColOrdinals = table.indexColOrdinals[:nIndexCols]
	} else {
		__antithesis_instrumentation__.Notify(455223)
		table.indexColOrdinals = make([]int, nIndexCols)
	}
	__antithesis_instrumentation__.Notify(455205)
	indexColOrdinals := table.indexColOrdinals
	_ = indexColOrdinals[len(fullColumns)-1]
	needToDecodeDecimalKey := false
	for i := range fullColumns {
		__antithesis_instrumentation__.Notify(455224)
		col := &fullColumns[i]
		colIdx, ok := tableArgs.ColIdxMap.Get(col.ColumnID)
		if ok {
			__antithesis_instrumentation__.Notify(455225)

			indexColOrdinals[i] = colIdx
			cf.mustDecodeIndexKey = true
			needToDecodeDecimalKey = needToDecodeDecimalKey || func() bool {
				__antithesis_instrumentation__.Notify(455226)
				return tableArgs.spec.FetchedColumns[colIdx].Type.Family() == types.DecimalFamily == true
			}() == true

			if col.IsComposite {
				__antithesis_instrumentation__.Notify(455227)
				table.compositeIndexColOrdinals.Add(colIdx)
			} else {
				__antithesis_instrumentation__.Notify(455228)
				table.neededValueColsByIdx.Remove(colIdx)
			}
		} else {
			__antithesis_instrumentation__.Notify(455229)

			indexColOrdinals[i] = -1
		}
	}
	__antithesis_instrumentation__.Notify(455206)
	if needToDecodeDecimalKey && func() bool {
		__antithesis_instrumentation__.Notify(455230)
		return cap(cf.scratch) < 64 == true
	}() == true {
		__antithesis_instrumentation__.Notify(455231)

		cf.scratch = make([]byte, 64)
	} else {
		__antithesis_instrumentation__.Notify(455232)
	}
	__antithesis_instrumentation__.Notify(455207)

	if table.spec.NumKeySuffixColumns > 0 && func() bool {
		__antithesis_instrumentation__.Notify(455233)
		return table.spec.IsSecondaryIndex == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(455234)
		return table.spec.IsUniqueIndex == true
	}() == true {
		__antithesis_instrumentation__.Notify(455235)
		suffixCols := table.spec.KeySuffixColumns()
		for i := range suffixCols {
			__antithesis_instrumentation__.Notify(455238)
			id := suffixCols[i].ColumnID
			colIdx, ok := tableArgs.ColIdxMap.Get(id)
			if ok {
				__antithesis_instrumentation__.Notify(455239)
				if suffixCols[i].IsComposite {
					__antithesis_instrumentation__.Notify(455240)
					table.compositeIndexColOrdinals.Add(colIdx)

					table.neededValueColsByIdx.Remove(colIdx)
				} else {
					__antithesis_instrumentation__.Notify(455241)
				}
			} else {
				__antithesis_instrumentation__.Notify(455242)
			}
		}
		__antithesis_instrumentation__.Notify(455236)

		if cap(table.extraValColOrdinals) >= len(suffixCols) {
			__antithesis_instrumentation__.Notify(455243)
			table.extraValColOrdinals = table.extraValColOrdinals[:len(suffixCols)]
		} else {
			__antithesis_instrumentation__.Notify(455244)
			table.extraValColOrdinals = make([]int, len(suffixCols))
		}
		__antithesis_instrumentation__.Notify(455237)

		extraValColOrdinals := table.extraValColOrdinals
		_ = extraValColOrdinals[len(suffixCols)-1]
		for i := range suffixCols {
			__antithesis_instrumentation__.Notify(455245)
			idx, ok := tableArgs.ColIdxMap.Get(suffixCols[i].ColumnID)
			if ok {
				__antithesis_instrumentation__.Notify(455246)

				extraValColOrdinals[i] = idx
			} else {
				__antithesis_instrumentation__.Notify(455247)

				extraValColOrdinals[i] = -1
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(455248)
	}
	__antithesis_instrumentation__.Notify(455208)

	cf.table = table
	cf.accountingHelper.Init(allocator, cf.table.typs)

	return nil
}

func (cf *cFetcher) setFetcher(f *row.KVFetcher, limitHint rowinfra.RowLimit) {
	__antithesis_instrumentation__.Notify(455249)
	cf.fetcher = f
	cf.machine.lastRowPrefix = nil
	cf.machine.limitHint = int(limitHint)
	cf.machine.state[0] = stateResetBatch
	cf.machine.state[1] = stateInitFetch
}

func (cf *cFetcher) StartScan(
	ctx context.Context,
	txn *kv.Txn,
	spans roachpb.Spans,
	bsHeader *roachpb.BoundedStalenessHeader,
	limitBatches bool,
	batchBytesLimit rowinfra.BytesLimit,
	limitHint rowinfra.RowLimit,
	forceProductionKVBatchSize bool,
) error {
	__antithesis_instrumentation__.Notify(455250)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(455255)
		return errors.AssertionFailedf("no spans")
	} else {
		__antithesis_instrumentation__.Notify(455256)
	}
	__antithesis_instrumentation__.Notify(455251)
	if !limitBatches && func() bool {
		__antithesis_instrumentation__.Notify(455257)
		return batchBytesLimit != rowinfra.NoBytesLimit == true
	}() == true {
		__antithesis_instrumentation__.Notify(455258)
		return errors.AssertionFailedf("batchBytesLimit set without limitBatches")
	} else {
		__antithesis_instrumentation__.Notify(455259)
	}
	__antithesis_instrumentation__.Notify(455252)

	firstBatchLimit := rowinfra.KeyLimit(limitHint)
	if firstBatchLimit != 0 {
		__antithesis_instrumentation__.Notify(455260)

		firstBatchLimit = rowinfra.KeyLimit(int(limitHint) * int(cf.table.spec.MaxKeysPerRow))
	} else {
		__antithesis_instrumentation__.Notify(455261)
	}
	__antithesis_instrumentation__.Notify(455253)

	f, err := row.NewKVFetcher(
		ctx,
		txn,
		spans,
		bsHeader,
		cf.reverse,
		batchBytesLimit,
		firstBatchLimit,
		cf.lockStrength,
		cf.lockWaitPolicy,
		cf.lockTimeout,
		cf.kvFetcherMemAcc,
		forceProductionKVBatchSize,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(455262)
		return err
	} else {
		__antithesis_instrumentation__.Notify(455263)
	}
	__antithesis_instrumentation__.Notify(455254)
	cf.setFetcher(f, limitHint)
	return nil
}

func (cf *cFetcher) StartScanStreaming(
	ctx context.Context,
	streamer *kvstreamer.Streamer,
	spans roachpb.Spans,
	limitHint rowinfra.RowLimit,
) error {
	__antithesis_instrumentation__.Notify(455264)
	kvBatchFetcher, err := row.NewTxnKVStreamer(ctx, streamer, spans, cf.lockStrength)
	if err != nil {
		__antithesis_instrumentation__.Notify(455266)
		return err
	} else {
		__antithesis_instrumentation__.Notify(455267)
	}
	__antithesis_instrumentation__.Notify(455265)
	f := row.NewKVStreamingFetcher(kvBatchFetcher)
	cf.setFetcher(f, limitHint)
	return nil
}

type fetcherState int

const (
	stateInvalid fetcherState = iota

	stateInitFetch

	stateResetBatch

	stateDecodeFirstKVOfRow

	stateFetchNextKVWithUnfinishedRow

	stateFinalizeRow

	stateEmitLastBatch

	stateFinished
)

const debugState = false

func (cf *cFetcher) setEstimatedRowCount(estimatedRowCount uint64) {
	__antithesis_instrumentation__.Notify(455268)
	cf.estimatedRowCount = estimatedRowCount
}

func (cf *cFetcher) setNextKV(kv roachpb.KeyValue, needsCopy bool) {
	__antithesis_instrumentation__.Notify(455269)
	if !needsCopy {
		__antithesis_instrumentation__.Notify(455271)
		cf.machine.nextKV = kv
		return
	} else {
		__antithesis_instrumentation__.Notify(455272)
	}
	__antithesis_instrumentation__.Notify(455270)

	kvCopy := roachpb.KeyValue{}
	kvCopy.Key = make(roachpb.Key, len(kv.Key))
	copy(kvCopy.Key, kv.Key)
	kvCopy.Value.RawBytes = make([]byte, len(kv.Value.RawBytes))
	copy(kvCopy.Value.RawBytes, kv.Value.RawBytes)
	kvCopy.Value.Timestamp = kv.Value.Timestamp
	cf.machine.nextKV = kvCopy
}

func (cf *cFetcher) NextBatch(ctx context.Context) (coldata.Batch, error) {
	__antithesis_instrumentation__.Notify(455273)
	for {
		__antithesis_instrumentation__.Notify(455274)
		if debugState {
			__antithesis_instrumentation__.Notify(455276)
			log.Infof(ctx, "State %s", cf.machine.state[0])
		} else {
			__antithesis_instrumentation__.Notify(455277)
		}
		__antithesis_instrumentation__.Notify(455275)
		switch cf.machine.state[0] {
		case stateInvalid:
			__antithesis_instrumentation__.Notify(455278)
			return nil, errors.New("invalid fetcher state")
		case stateInitFetch:
			__antithesis_instrumentation__.Notify(455279)
			moreKVs, kv, finalReferenceToBatch, err := cf.fetcher.NextKV(ctx, cf.mvccDecodeStrategy)
			if err != nil {
				__antithesis_instrumentation__.Notify(455306)
				return nil, cf.convertFetchError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(455307)
			}
			__antithesis_instrumentation__.Notify(455280)
			if !moreKVs {
				__antithesis_instrumentation__.Notify(455308)
				cf.machine.state[0] = stateEmitLastBatch
				continue
			} else {
				__antithesis_instrumentation__.Notify(455309)
			}
			__antithesis_instrumentation__.Notify(455281)

			cf.setNextKV(kv, finalReferenceToBatch)
			cf.machine.state[0] = stateDecodeFirstKVOfRow

		case stateResetBatch:
			__antithesis_instrumentation__.Notify(455282)
			cf.resetBatch()
			cf.shiftState()
		case stateDecodeFirstKVOfRow:
			__antithesis_instrumentation__.Notify(455283)

			cf.table.rowLastModified = hlc.Timestamp{}

			var foundNull bool
			if cf.mustDecodeIndexKey {
				__antithesis_instrumentation__.Notify(455310)
				if debugState {
					__antithesis_instrumentation__.Notify(455313)
					log.Infof(ctx, "decoding first key %s", cf.machine.nextKV.Key)
				} else {
					__antithesis_instrumentation__.Notify(455314)
				}
				__antithesis_instrumentation__.Notify(455311)
				var (
					key []byte
					err error
				)

				checkAllColsForNull := cf.table.spec.IsSecondaryIndex && func() bool {
					__antithesis_instrumentation__.Notify(455315)
					return cf.table.spec.IsUniqueIndex == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(455316)
					return cf.table.spec.MaxKeysPerRow != 1 == true
				}() == true
				key, foundNull, cf.scratch, err = colencoding.DecodeKeyValsToCols(
					&cf.table.da,
					&cf.machine.colvecs,
					cf.machine.rowIdx,
					cf.table.indexColOrdinals,
					checkAllColsForNull,
					cf.table.spec.KeyFullColumns(),
					nil,
					cf.machine.nextKV.Key[cf.table.spec.KeyPrefixLength:],
					cf.scratch,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(455317)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(455318)
				}
				__antithesis_instrumentation__.Notify(455312)
				prefix := cf.machine.nextKV.Key[:len(cf.machine.nextKV.Key)-len(key)]
				cf.machine.lastRowPrefix = prefix
			} else {
				__antithesis_instrumentation__.Notify(455319)
				prefixLen, err := keys.GetRowPrefixLength(cf.machine.nextKV.Key)
				if err != nil {
					__antithesis_instrumentation__.Notify(455321)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(455322)
				}
				__antithesis_instrumentation__.Notify(455320)
				cf.machine.lastRowPrefix = cf.machine.nextKV.Key[:prefixLen]
			}
			__antithesis_instrumentation__.Notify(455284)

			if foundNull && func() bool {
				__antithesis_instrumentation__.Notify(455323)
				return cf.table.spec.IsSecondaryIndex == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(455324)
				return cf.table.spec.IsUniqueIndex == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(455325)
				return cf.table.spec.MaxKeysPerRow != 1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(455326)

				prefixLen := len(cf.machine.lastRowPrefix)
				remainingBytes := cf.machine.nextKV.Key[prefixLen:]
				origRemainingBytesLen := len(remainingBytes)
				for i := 0; i < int(cf.table.spec.NumKeySuffixColumns); i++ {
					__antithesis_instrumentation__.Notify(455328)
					var err error

					remainingBytes, err = keyside.Skip(remainingBytes)
					if err != nil {
						__antithesis_instrumentation__.Notify(455329)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(455330)
					}
				}
				__antithesis_instrumentation__.Notify(455327)
				cf.machine.lastRowPrefix = cf.machine.nextKV.Key[:prefixLen+(origRemainingBytesLen-len(remainingBytes))]
			} else {
				__antithesis_instrumentation__.Notify(455331)
			}
			__antithesis_instrumentation__.Notify(455285)

			familyID, err := cf.getCurrentColumnFamilyID()
			if err != nil {
				__antithesis_instrumentation__.Notify(455332)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(455333)
			}
			__antithesis_instrumentation__.Notify(455286)
			cf.machine.remainingValueColsByIdx.CopyFrom(cf.table.neededValueColsByIdx)

			if err := cf.processValue(ctx, familyID); err != nil {
				__antithesis_instrumentation__.Notify(455334)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(455335)
			}
			__antithesis_instrumentation__.Notify(455287)

			if cf.table.rowLastModified.Less(cf.machine.nextKV.Value.Timestamp) {
				__antithesis_instrumentation__.Notify(455336)
				cf.table.rowLastModified = cf.machine.nextKV.Value.Timestamp
			} else {
				__antithesis_instrumentation__.Notify(455337)
			}
			__antithesis_instrumentation__.Notify(455288)

			if cf.table.spec.MaxKeysPerRow == 1 {
				__antithesis_instrumentation__.Notify(455338)
				cf.machine.state[0] = stateFinalizeRow
				cf.machine.state[1] = stateInitFetch
				continue
			} else {
				__antithesis_instrumentation__.Notify(455339)
			}
			__antithesis_instrumentation__.Notify(455289)

			cf.machine.state[0] = stateFetchNextKVWithUnfinishedRow

		case stateFetchNextKVWithUnfinishedRow:
			__antithesis_instrumentation__.Notify(455290)
			moreKVs, kv, finalReferenceToBatch, err := cf.fetcher.NextKV(ctx, cf.mvccDecodeStrategy)
			if err != nil {
				__antithesis_instrumentation__.Notify(455340)
				return nil, cf.convertFetchError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(455341)
			}
			__antithesis_instrumentation__.Notify(455291)
			if !moreKVs {
				__antithesis_instrumentation__.Notify(455342)

				cf.machine.state[0] = stateFinalizeRow
				cf.machine.state[1] = stateEmitLastBatch
				continue
			} else {
				__antithesis_instrumentation__.Notify(455343)
			}
			__antithesis_instrumentation__.Notify(455292)

			cf.setNextKV(kv, finalReferenceToBatch)
			if debugState {
				__antithesis_instrumentation__.Notify(455344)
				log.Infof(ctx, "decoding next key %s", cf.machine.nextKV.Key)
			} else {
				__antithesis_instrumentation__.Notify(455345)
			}
			__antithesis_instrumentation__.Notify(455293)

			if !bytes.HasPrefix(kv.Key[cf.table.spec.KeyPrefixLength:], cf.machine.lastRowPrefix[cf.table.spec.KeyPrefixLength:]) {
				__antithesis_instrumentation__.Notify(455346)

				cf.machine.state[0] = stateFinalizeRow
				cf.machine.state[1] = stateDecodeFirstKVOfRow
				continue
			} else {
				__antithesis_instrumentation__.Notify(455347)
			}
			__antithesis_instrumentation__.Notify(455294)

			familyID, err := cf.getCurrentColumnFamilyID()
			if err != nil {
				__antithesis_instrumentation__.Notify(455348)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(455349)
			}
			__antithesis_instrumentation__.Notify(455295)

			if err := cf.processValue(ctx, familyID); err != nil {
				__antithesis_instrumentation__.Notify(455350)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(455351)
			}
			__antithesis_instrumentation__.Notify(455296)

			if cf.table.rowLastModified.Less(cf.machine.nextKV.Value.Timestamp) {
				__antithesis_instrumentation__.Notify(455352)
				cf.table.rowLastModified = cf.machine.nextKV.Value.Timestamp
			} else {
				__antithesis_instrumentation__.Notify(455353)
			}
			__antithesis_instrumentation__.Notify(455297)

			if familyID == cf.table.spec.MaxFamilyID {
				__antithesis_instrumentation__.Notify(455354)

				cf.machine.state[0] = stateFinalizeRow
				cf.machine.state[1] = stateInitFetch
			} else {
				__antithesis_instrumentation__.Notify(455355)

				cf.machine.state[0] = stateFetchNextKVWithUnfinishedRow
			}

		case stateFinalizeRow:
			__antithesis_instrumentation__.Notify(455298)

			if cf.table.timestampOutputIdx != noOutputColumn {
				__antithesis_instrumentation__.Notify(455356)
				cf.machine.timestampCol[cf.machine.rowIdx] = tree.TimestampToDecimal(cf.table.rowLastModified)
			} else {
				__antithesis_instrumentation__.Notify(455357)
			}
			__antithesis_instrumentation__.Notify(455299)

			if err := cf.fillNulls(); err != nil {
				__antithesis_instrumentation__.Notify(455358)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(455359)
			}
			__antithesis_instrumentation__.Notify(455300)

			cf.accountingHelper.AccountForSet(cf.machine.rowIdx)
			cf.machine.rowIdx++
			cf.shiftState()

			var emitBatch bool
			if cf.maxCapacity == 0 && func() bool {
				__antithesis_instrumentation__.Notify(455360)
				return cf.accountingHelper.Allocator.Used() >= cf.memoryLimit == true
			}() == true {
				__antithesis_instrumentation__.Notify(455361)
				cf.maxCapacity = cf.machine.rowIdx
			} else {
				__antithesis_instrumentation__.Notify(455362)
			}
			__antithesis_instrumentation__.Notify(455301)
			if cf.machine.rowIdx >= cf.machine.batch.Capacity() || func() bool {
				__antithesis_instrumentation__.Notify(455363)
				return (cf.maxCapacity > 0 && func() bool {
					__antithesis_instrumentation__.Notify(455364)
					return cf.machine.rowIdx >= cf.maxCapacity == true
				}() == true) == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(455365)
				return (cf.machine.limitHint > 0 && func() bool {
					__antithesis_instrumentation__.Notify(455366)
					return cf.machine.rowIdx >= cf.machine.limitHint == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(455367)

				emitBatch = true

				cf.machine.limitHint -= cf.machine.rowIdx
			} else {
				__antithesis_instrumentation__.Notify(455368)
			}
			__antithesis_instrumentation__.Notify(455302)

			if emitBatch {
				__antithesis_instrumentation__.Notify(455369)
				cf.pushState(stateResetBatch)
				cf.finalizeBatch()
				return cf.machine.batch, nil
			} else {
				__antithesis_instrumentation__.Notify(455370)
			}

		case stateEmitLastBatch:
			__antithesis_instrumentation__.Notify(455303)
			cf.machine.state[0] = stateFinished
			cf.finalizeBatch()

			cf.Close(ctx)
			return cf.machine.batch, nil

		case stateFinished:
			__antithesis_instrumentation__.Notify(455304)

			cf.Close(ctx)
			return coldata.ZeroBatch, nil
		default:
			__antithesis_instrumentation__.Notify(455305)
		}
	}
}

func (cf *cFetcher) shiftState() {
	__antithesis_instrumentation__.Notify(455371)
	copy(cf.machine.state[:2], cf.machine.state[1:])
	cf.machine.state[2] = stateInvalid
}

func (cf *cFetcher) pushState(state fetcherState) {
	__antithesis_instrumentation__.Notify(455372)
	copy(cf.machine.state[1:], cf.machine.state[:2])
	cf.machine.state[0] = state
}

func (cf *cFetcher) getDatumAt(colIdx int, rowIdx int) tree.Datum {
	__antithesis_instrumentation__.Notify(455373)
	res := []tree.Datum{nil}
	colconv.ColVecToDatumAndDeselect(res, cf.machine.colvecs.Vecs[colIdx], 1, []int{rowIdx}, &cf.table.da)
	return res[0]
}

func (cf *cFetcher) writeDecodedCols(buf *strings.Builder, colOrdinals []int, separator byte) {
	__antithesis_instrumentation__.Notify(455374)
	for i, idx := range colOrdinals {
		__antithesis_instrumentation__.Notify(455375)
		if i > 0 {
			__antithesis_instrumentation__.Notify(455377)
			buf.WriteByte(separator)
		} else {
			__antithesis_instrumentation__.Notify(455378)
		}
		__antithesis_instrumentation__.Notify(455376)
		if idx != -1 {
			__antithesis_instrumentation__.Notify(455379)
			buf.WriteString(cf.getDatumAt(idx, cf.machine.rowIdx).String())
		} else {
			__antithesis_instrumentation__.Notify(455380)
			buf.WriteByte('?')
		}
	}
}

func (cf *cFetcher) processValue(ctx context.Context, familyID descpb.FamilyID) (err error) {
	__antithesis_instrumentation__.Notify(455381)
	table := cf.table

	var prettyKey, prettyValue string
	if cf.traceKV {
		__antithesis_instrumentation__.Notify(455386)
		defer func() {
			__antithesis_instrumentation__.Notify(455388)
			if err == nil {
				__antithesis_instrumentation__.Notify(455389)
				log.VEventf(ctx, 2, "fetched: %s -> %s", prettyKey, prettyValue)
			} else {
				__antithesis_instrumentation__.Notify(455390)
			}
		}()
		__antithesis_instrumentation__.Notify(455387)

		var buf strings.Builder
		buf.WriteByte('/')
		buf.WriteString(cf.table.spec.TableName)
		buf.WriteByte('/')
		buf.WriteString(cf.table.spec.IndexName)
		buf.WriteByte('/')
		cf.writeDecodedCols(&buf, cf.table.indexColOrdinals, '/')
		prettyKey = buf.String()
	} else {
		__antithesis_instrumentation__.Notify(455391)
	}
	__antithesis_instrumentation__.Notify(455382)

	if len(table.spec.FetchedColumns) == 0 {
		__antithesis_instrumentation__.Notify(455392)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(455393)
	}
	__antithesis_instrumentation__.Notify(455383)

	val := cf.machine.nextKV.Value
	if !table.spec.IsSecondaryIndex || func() bool {
		__antithesis_instrumentation__.Notify(455394)
		return table.spec.EncodingType == descpb.PrimaryIndexEncoding == true
	}() == true {
		__antithesis_instrumentation__.Notify(455395)

		switch val.GetTag() {
		case roachpb.ValueType_TUPLE:
			__antithesis_instrumentation__.Notify(455397)

			var tupleBytes []byte
			tupleBytes, err = val.GetTuple()
			if err != nil {
				__antithesis_instrumentation__.Notify(455403)
				break
			} else {
				__antithesis_instrumentation__.Notify(455404)
			}
			__antithesis_instrumentation__.Notify(455398)
			prettyKey, prettyValue, err = cf.processValueBytes(ctx, table, tupleBytes, prettyKey)

		default:
			__antithesis_instrumentation__.Notify(455399)

			if familyID == 0 {
				__antithesis_instrumentation__.Notify(455405)
				break
			} else {
				__antithesis_instrumentation__.Notify(455406)
			}
			__antithesis_instrumentation__.Notify(455400)

			var defaultColumnID descpb.ColumnID
			for _, f := range table.spec.FamilyDefaultColumns {
				__antithesis_instrumentation__.Notify(455407)
				if f.FamilyID == familyID {
					__antithesis_instrumentation__.Notify(455408)
					defaultColumnID = f.DefaultColumnID
					break
				} else {
					__antithesis_instrumentation__.Notify(455409)
				}
			}
			__antithesis_instrumentation__.Notify(455401)
			if defaultColumnID == 0 {
				__antithesis_instrumentation__.Notify(455410)
				return scrub.WrapError(
					scrub.IndexKeyDecodingError,
					errors.Errorf("single entry value with no default column id"),
				)
			} else {
				__antithesis_instrumentation__.Notify(455411)
			}
			__antithesis_instrumentation__.Notify(455402)
			prettyKey, prettyValue, err = cf.processValueSingle(ctx, table, defaultColumnID, prettyKey)
		}
		__antithesis_instrumentation__.Notify(455396)
		if err != nil {
			__antithesis_instrumentation__.Notify(455412)
			return scrub.WrapError(scrub.IndexValueDecodingError, err)
		} else {
			__antithesis_instrumentation__.Notify(455413)
		}
	} else {
		__antithesis_instrumentation__.Notify(455414)
		tag := val.GetTag()
		var valueBytes []byte
		switch tag {
		case roachpb.ValueType_BYTES:
			__antithesis_instrumentation__.Notify(455416)

			valueBytes, err = val.GetBytes()
			if err != nil {
				__antithesis_instrumentation__.Notify(455420)
				return scrub.WrapError(scrub.IndexValueDecodingError, err)
			} else {
				__antithesis_instrumentation__.Notify(455421)
			}
			__antithesis_instrumentation__.Notify(455417)

			if table.spec.IsSecondaryIndex && func() bool {
				__antithesis_instrumentation__.Notify(455422)
				return table.spec.IsUniqueIndex == true
			}() == true {
				__antithesis_instrumentation__.Notify(455423)

				valueBytes, _, cf.scratch, err = colencoding.DecodeKeyValsToCols(
					&table.da,
					&cf.machine.colvecs,
					cf.machine.rowIdx,
					table.extraValColOrdinals,
					false,
					table.spec.KeySuffixColumns(),
					&cf.machine.remainingValueColsByIdx,
					valueBytes,
					cf.scratch,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(455425)
					return scrub.WrapError(scrub.SecondaryIndexKeyExtraValueDecodingError, err)
				} else {
					__antithesis_instrumentation__.Notify(455426)
				}
				__antithesis_instrumentation__.Notify(455424)
				if cf.traceKV && func() bool {
					__antithesis_instrumentation__.Notify(455427)
					return len(table.extraValColOrdinals) > 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(455428)
					var buf strings.Builder
					buf.WriteByte('/')
					cf.writeDecodedCols(&buf, table.extraValColOrdinals, '/')
					prettyValue = buf.String()
				} else {
					__antithesis_instrumentation__.Notify(455429)
				}
			} else {
				__antithesis_instrumentation__.Notify(455430)
			}
		case roachpb.ValueType_TUPLE:
			__antithesis_instrumentation__.Notify(455418)
			valueBytes, err = val.GetTuple()
			if err != nil {
				__antithesis_instrumentation__.Notify(455431)
				return scrub.WrapError(scrub.IndexValueDecodingError, err)
			} else {
				__antithesis_instrumentation__.Notify(455432)
			}
		default:
			__antithesis_instrumentation__.Notify(455419)
		}
		__antithesis_instrumentation__.Notify(455415)

		if len(valueBytes) > 0 {
			__antithesis_instrumentation__.Notify(455433)
			prettyKey, prettyValue, err = cf.processValueBytes(
				ctx, table, valueBytes, prettyKey,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(455434)
				return scrub.WrapError(scrub.IndexValueDecodingError, err)
			} else {
				__antithesis_instrumentation__.Notify(455435)
			}
		} else {
			__antithesis_instrumentation__.Notify(455436)
		}
	}
	__antithesis_instrumentation__.Notify(455384)

	if cf.traceKV && func() bool {
		__antithesis_instrumentation__.Notify(455437)
		return prettyValue == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(455438)
		prettyValue = "<undecoded>"
	} else {
		__antithesis_instrumentation__.Notify(455439)
	}
	__antithesis_instrumentation__.Notify(455385)

	return nil
}

func (cf *cFetcher) processValueSingle(
	ctx context.Context, table *cTableInfo, colID descpb.ColumnID, prettyKeyPrefix string,
) (prettyKey string, prettyValue string, err error) {
	__antithesis_instrumentation__.Notify(455440)
	prettyKey = prettyKeyPrefix

	if idx, ok := table.ColIdxMap.Get(colID); ok {
		__antithesis_instrumentation__.Notify(455443)
		if cf.traceKV {
			__antithesis_instrumentation__.Notify(455449)
			prettyKey = fmt.Sprintf("%s/%s", prettyKey, table.spec.FetchedColumns[idx].Name)
		} else {
			__antithesis_instrumentation__.Notify(455450)
		}
		__antithesis_instrumentation__.Notify(455444)
		val := cf.machine.nextKV.Value
		if len(val.RawBytes) == 0 {
			__antithesis_instrumentation__.Notify(455451)
			return prettyKey, "", nil
		} else {
			__antithesis_instrumentation__.Notify(455452)
		}
		__antithesis_instrumentation__.Notify(455445)
		typ := cf.table.spec.FetchedColumns[idx].Type
		err := colencoding.UnmarshalColumnValueToCol(
			&table.da, &cf.machine.colvecs, idx, cf.machine.rowIdx, typ, val,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(455453)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(455454)
		}
		__antithesis_instrumentation__.Notify(455446)
		cf.machine.remainingValueColsByIdx.Remove(idx)

		if cf.traceKV {
			__antithesis_instrumentation__.Notify(455455)
			prettyValue = cf.getDatumAt(idx, cf.machine.rowIdx).String()
		} else {
			__antithesis_instrumentation__.Notify(455456)
		}
		__antithesis_instrumentation__.Notify(455447)
		if row.DebugRowFetch {
			__antithesis_instrumentation__.Notify(455457)
			log.Infof(ctx, "Scan %s -> %v", cf.machine.nextKV.Key, "?")
		} else {
			__antithesis_instrumentation__.Notify(455458)
		}
		__antithesis_instrumentation__.Notify(455448)
		return prettyKey, prettyValue, nil
	} else {
		__antithesis_instrumentation__.Notify(455459)
	}
	__antithesis_instrumentation__.Notify(455441)

	if row.DebugRowFetch {
		__antithesis_instrumentation__.Notify(455460)
		log.Infof(ctx, "Scan %s -> [%d] (skipped)", cf.machine.nextKV.Key, colID)
	} else {
		__antithesis_instrumentation__.Notify(455461)
	}
	__antithesis_instrumentation__.Notify(455442)
	return prettyKey, prettyValue, nil
}

func (cf *cFetcher) processValueBytes(
	ctx context.Context, table *cTableInfo, valueBytes []byte, prettyKeyPrefix string,
) (prettyKey string, prettyValue string, err error) {
	__antithesis_instrumentation__.Notify(455462)
	prettyKey = prettyKeyPrefix
	if cf.traceKV {
		__antithesis_instrumentation__.Notify(455466)
		if cf.machine.prettyValueBuf == nil {
			__antithesis_instrumentation__.Notify(455468)
			cf.machine.prettyValueBuf = &bytes.Buffer{}
		} else {
			__antithesis_instrumentation__.Notify(455469)
		}
		__antithesis_instrumentation__.Notify(455467)
		cf.machine.prettyValueBuf.Reset()
	} else {
		__antithesis_instrumentation__.Notify(455470)
	}
	__antithesis_instrumentation__.Notify(455463)

	cf.machine.remainingValueColsByIdx.UnionWith(cf.table.compositeIndexColOrdinals)

	var (
		colIDDiff      uint32
		lastColID      descpb.ColumnID
		dataOffset     int
		typ            encoding.Type
		lastColIDIndex int
	)

	for len(valueBytes) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(455471)
		return cf.machine.remainingValueColsByIdx.Len() > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(455472)
		_, dataOffset, colIDDiff, typ, err = encoding.DecodeValueTag(valueBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(455478)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(455479)
		}
		__antithesis_instrumentation__.Notify(455473)
		colID := lastColID + descpb.ColumnID(colIDDiff)
		lastColID = colID
		vecIdx := -1

		for ; lastColIDIndex < len(table.orderedColIdxMap.vals); lastColIDIndex++ {
			__antithesis_instrumentation__.Notify(455480)
			nextID := table.orderedColIdxMap.vals[lastColIDIndex]
			if nextID == colID {
				__antithesis_instrumentation__.Notify(455481)
				vecIdx = table.orderedColIdxMap.ords[lastColIDIndex]

				lastColIDIndex++
				break
			} else {
				__antithesis_instrumentation__.Notify(455482)
				if nextID > colID {
					__antithesis_instrumentation__.Notify(455483)
					break
				} else {
					__antithesis_instrumentation__.Notify(455484)
				}
			}
		}
		__antithesis_instrumentation__.Notify(455474)
		if vecIdx == -1 {
			__antithesis_instrumentation__.Notify(455485)

			len, err := encoding.PeekValueLengthWithOffsetsAndType(valueBytes, dataOffset, typ)
			if err != nil {
				__antithesis_instrumentation__.Notify(455488)
				return "", "", err
			} else {
				__antithesis_instrumentation__.Notify(455489)
			}
			__antithesis_instrumentation__.Notify(455486)
			valueBytes = valueBytes[len:]
			if row.DebugRowFetch {
				__antithesis_instrumentation__.Notify(455490)
				log.Infof(ctx, "Scan %s -> [%d] (skipped)", cf.machine.nextKV.Key, colID)
			} else {
				__antithesis_instrumentation__.Notify(455491)
			}
			__antithesis_instrumentation__.Notify(455487)
			continue
		} else {
			__antithesis_instrumentation__.Notify(455492)
		}
		__antithesis_instrumentation__.Notify(455475)

		if cf.traceKV {
			__antithesis_instrumentation__.Notify(455493)
			prettyKey = fmt.Sprintf("%s/%s", prettyKey, table.spec.FetchedColumns[vecIdx].Name)
		} else {
			__antithesis_instrumentation__.Notify(455494)
		}
		__antithesis_instrumentation__.Notify(455476)

		valueBytes, err = colencoding.DecodeTableValueToCol(
			&table.da, &cf.machine.colvecs, vecIdx, cf.machine.rowIdx, typ,
			dataOffset, cf.table.typs[vecIdx], valueBytes,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(455495)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(455496)
		}
		__antithesis_instrumentation__.Notify(455477)
		cf.machine.remainingValueColsByIdx.Remove(vecIdx)
		if cf.traceKV {
			__antithesis_instrumentation__.Notify(455497)
			dVal := cf.getDatumAt(vecIdx, cf.machine.rowIdx)
			if _, err := fmt.Fprintf(cf.machine.prettyValueBuf, "/%v", dVal.String()); err != nil {
				__antithesis_instrumentation__.Notify(455498)
				return "", "", err
			} else {
				__antithesis_instrumentation__.Notify(455499)
			}
		} else {
			__antithesis_instrumentation__.Notify(455500)
		}
	}
	__antithesis_instrumentation__.Notify(455464)
	if cf.traceKV {
		__antithesis_instrumentation__.Notify(455501)
		prettyValue = cf.machine.prettyValueBuf.String()
	} else {
		__antithesis_instrumentation__.Notify(455502)
	}
	__antithesis_instrumentation__.Notify(455465)
	return prettyKey, prettyValue, nil
}

func (cf *cFetcher) fillNulls() error {
	__antithesis_instrumentation__.Notify(455503)
	table := cf.table
	if cf.machine.remainingValueColsByIdx.Empty() {
		__antithesis_instrumentation__.Notify(455506)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(455507)
	}
	__antithesis_instrumentation__.Notify(455504)
	for i, ok := cf.machine.remainingValueColsByIdx.Next(0); ok; i, ok = cf.machine.remainingValueColsByIdx.Next(i + 1) {
		__antithesis_instrumentation__.Notify(455508)

		if table.compositeIndexColOrdinals.Contains(i) {
			__antithesis_instrumentation__.Notify(455511)
			continue
		} else {
			__antithesis_instrumentation__.Notify(455512)
		}
		__antithesis_instrumentation__.Notify(455509)
		if table.spec.FetchedColumns[i].IsNonNullable {
			__antithesis_instrumentation__.Notify(455513)
			var indexColValues strings.Builder
			cf.writeDecodedCols(&indexColValues, table.indexColOrdinals, ',')
			var indexColNames []string
			for i := range table.spec.KeyFullColumns() {
				__antithesis_instrumentation__.Notify(455515)
				indexColNames = append(indexColNames, table.spec.KeyAndSuffixColumns[i].Name)
			}
			__antithesis_instrumentation__.Notify(455514)
			return scrub.WrapError(scrub.UnexpectedNullValueError, errors.Errorf(
				"non-nullable column \"%s:%s\" with no value! Index scanned was %q with the index key columns (%s) and the values (%s)",
				table.spec.TableName, table.spec.FetchedColumns[i].Name, table.spec.IndexName,
				strings.Join(indexColNames, ","), indexColValues.String()))
		} else {
			__antithesis_instrumentation__.Notify(455516)
		}
		__antithesis_instrumentation__.Notify(455510)
		cf.machine.colvecs.Nulls[i].SetNull(cf.machine.rowIdx)
	}
	__antithesis_instrumentation__.Notify(455505)
	return nil
}

func (cf *cFetcher) finalizeBatch() {
	__antithesis_instrumentation__.Notify(455517)

	if cf.table.oidOutputIdx != noOutputColumn {
		__antithesis_instrumentation__.Notify(455519)
		id := cf.table.spec.TableID
		for i := 0; i < cf.machine.rowIdx; i++ {
			__antithesis_instrumentation__.Notify(455520)

			cf.machine.tableoidCol.Set(i, cf.table.da.NewDOid(tree.MakeDOid(tree.DInt(id))))
		}
	} else {
		__antithesis_instrumentation__.Notify(455521)
	}
	__antithesis_instrumentation__.Notify(455518)
	cf.machine.batch.SetLength(cf.machine.rowIdx)
	cf.machine.rowIdx = 0
}

func (cf *cFetcher) getCurrentColumnFamilyID() (descpb.FamilyID, error) {
	__antithesis_instrumentation__.Notify(455522)

	if cf.table.spec.MaxFamilyID == 0 {
		__antithesis_instrumentation__.Notify(455525)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(455526)
	}
	__antithesis_instrumentation__.Notify(455523)

	var id uint64
	_, id, err := encoding.DecodeUvarintAscending(cf.machine.nextKV.Key[len(cf.machine.lastRowPrefix):])
	if err != nil {
		__antithesis_instrumentation__.Notify(455527)
		return 0, scrub.WrapError(scrub.IndexKeyDecodingError, err)
	} else {
		__antithesis_instrumentation__.Notify(455528)
	}
	__antithesis_instrumentation__.Notify(455524)
	return descpb.FamilyID(id), nil
}

func (cf *cFetcher) convertFetchError(ctx context.Context, err error) error {
	__antithesis_instrumentation__.Notify(455529)
	err = row.ConvertFetchError(&cf.table.spec, err)
	err = colexecerror.NewStorageError(err)
	return err
}

func (cf *cFetcher) getBytesRead() int64 {
	__antithesis_instrumentation__.Notify(455530)
	if cf.fetcher != nil {
		__antithesis_instrumentation__.Notify(455532)
		cf.bytesRead += cf.fetcher.ResetBytesRead()
	} else {
		__antithesis_instrumentation__.Notify(455533)
	}
	__antithesis_instrumentation__.Notify(455531)
	return cf.bytesRead
}

var cFetcherPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(455534)
		return &cFetcher{}
	},
}

func (cf *cFetcher) Release() {
	__antithesis_instrumentation__.Notify(455535)
	cf.accountingHelper.Release()
	if cf.table != nil {
		__antithesis_instrumentation__.Notify(455537)
		cf.table.Release()
	} else {
		__antithesis_instrumentation__.Notify(455538)
	}
	__antithesis_instrumentation__.Notify(455536)
	colvecs := cf.machine.colvecs
	colvecs.Reset()
	*cf = cFetcher{
		scratch: cf.scratch[:0],
	}
	cf.machine.colvecs = colvecs
	cFetcherPool.Put(cf)
}

func (cf *cFetcher) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(455539)
	if cf != nil && func() bool {
		__antithesis_instrumentation__.Notify(455540)
		return cf.fetcher != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(455541)
		cf.bytesRead += cf.fetcher.GetBytesRead()
		cf.fetcher.Close(ctx)
		cf.fetcher = nil
	} else {
		__antithesis_instrumentation__.Notify(455542)
	}
}
