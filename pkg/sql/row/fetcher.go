package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/scrub"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const DebugRowFetch = false

const noOutputColumn = -1

type KVBatchFetcher interface {
	nextBatch(ctx context.Context) (ok bool, kvs []roachpb.KeyValue, batchResponse []byte, err error)

	close(ctx context.Context)
}

type tableInfo struct {
	spec descpb.IndexFetchSpec

	neededValueColsByIdx util.FastIntSet

	neededValueCols int

	colIdxMap catalog.TableColMap

	indexColIdx []int

	keyVals    []rowenc.EncDatum
	extraVals  []rowenc.EncDatum
	row        rowenc.EncDatumRow
	decodedRow tree.Datums

	rowLastModified hlc.Timestamp

	timestampOutputIdx int

	tableOid     tree.Datum
	oidOutputIdx int

	rowIsDeleted bool
}

type Fetcher struct {
	table tableInfo

	reverse bool

	mustDecodeIndexKey bool

	lockStrength descpb.ScanLockingStrength

	lockWaitPolicy descpb.ScanLockingWaitPolicy

	lockTimeout time.Duration

	traceKV bool

	mvccDecodeStrategy MVCCDecodingStrategy

	kvFetcher *KVFetcher

	indexKey       []byte
	prettyValueBuf *bytes.Buffer

	valueColsFound int

	kv                roachpb.KeyValue
	keyRemainingBytes []byte
	kvEnd             bool

	IgnoreUnexpectedNulls bool

	alloc *tree.DatumAlloc

	mon             *mon.BytesMonitor
	kvFetcherMemAcc *mon.BoundAccount
}

func (rf *Fetcher) Reset() {
	__antithesis_instrumentation__.Notify(567652)
	*rf = Fetcher{
		table: rf.table,
	}
}

func (rf *Fetcher) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(567653)
	if rf.kvFetcher != nil {
		__antithesis_instrumentation__.Notify(567655)
		rf.kvFetcher.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(567656)
	}
	__antithesis_instrumentation__.Notify(567654)
	if rf.mon != nil {
		__antithesis_instrumentation__.Notify(567657)
		rf.kvFetcherMemAcc.Close(ctx)
		rf.mon.Stop(ctx)
	} else {
		__antithesis_instrumentation__.Notify(567658)
	}
}

func (rf *Fetcher) Init(
	ctx context.Context,
	reverse bool,
	lockStrength descpb.ScanLockingStrength,
	lockWaitPolicy descpb.ScanLockingWaitPolicy,
	lockTimeout time.Duration,
	alloc *tree.DatumAlloc,
	memMonitor *mon.BytesMonitor,
	spec *descpb.IndexFetchSpec,
) error {
	__antithesis_instrumentation__.Notify(567659)
	if spec.Version != descpb.IndexFetchSpecVersionInitial {
		__antithesis_instrumentation__.Notify(567669)
		return errors.Newf("unsupported IndexFetchSpec version %d", spec.Version)
	} else {
		__antithesis_instrumentation__.Notify(567670)
	}
	__antithesis_instrumentation__.Notify(567660)
	rf.reverse = reverse
	rf.lockStrength = lockStrength
	rf.lockWaitPolicy = lockWaitPolicy
	rf.lockTimeout = lockTimeout
	rf.alloc = alloc

	if memMonitor != nil {
		__antithesis_instrumentation__.Notify(567671)
		rf.mon = mon.NewMonitorInheritWithLimit("fetcher-mem", 0, memMonitor)
		rf.mon.Start(ctx, memMonitor, mon.BoundAccount{})
		memAcc := rf.mon.MakeBoundAccount()
		rf.kvFetcherMemAcc = &memAcc
	} else {
		__antithesis_instrumentation__.Notify(567672)
	}
	__antithesis_instrumentation__.Notify(567661)

	table := &rf.table
	*table = tableInfo{
		spec:       *spec,
		row:        make(rowenc.EncDatumRow, len(spec.FetchedColumns)),
		decodedRow: make(tree.Datums, len(spec.FetchedColumns)),

		indexColIdx:        rf.table.indexColIdx[:0],
		keyVals:            rf.table.keyVals[:0],
		extraVals:          rf.table.extraVals[:0],
		timestampOutputIdx: noOutputColumn,
		oidOutputIdx:       noOutputColumn,
	}

	for idx := range spec.FetchedColumns {
		__antithesis_instrumentation__.Notify(567673)
		colID := spec.FetchedColumns[idx].ColumnID
		table.colIdxMap.Set(colID, idx)
		if colinfo.IsColIDSystemColumn(colID) {
			__antithesis_instrumentation__.Notify(567674)
			switch colinfo.GetSystemColumnKindFromColumnID(colID) {
			case catpb.SystemColumnKind_MVCCTIMESTAMP:
				__antithesis_instrumentation__.Notify(567675)
				table.timestampOutputIdx = idx
				rf.mvccDecodeStrategy = MVCCDecodingRequired

			case catpb.SystemColumnKind_TABLEOID:
				__antithesis_instrumentation__.Notify(567676)
				table.oidOutputIdx = idx
				table.tableOid = tree.NewDOid(tree.DInt(spec.TableID))
			default:
				__antithesis_instrumentation__.Notify(567677)
			}
		} else {
			__antithesis_instrumentation__.Notify(567678)
		}
	}
	__antithesis_instrumentation__.Notify(567662)

	if len(spec.FetchedColumns) > 0 {
		__antithesis_instrumentation__.Notify(567679)
		table.neededValueColsByIdx.AddRange(0, len(spec.FetchedColumns)-1)
	} else {
		__antithesis_instrumentation__.Notify(567680)
	}
	__antithesis_instrumentation__.Notify(567663)

	nExtraCols := 0

	if table.spec.IsSecondaryIndex && func() bool {
		__antithesis_instrumentation__.Notify(567681)
		return table.spec.IsUniqueIndex == true
	}() == true {
		__antithesis_instrumentation__.Notify(567682)
		nExtraCols = int(table.spec.NumKeySuffixColumns)
	} else {
		__antithesis_instrumentation__.Notify(567683)
	}
	__antithesis_instrumentation__.Notify(567664)
	nIndexCols := len(spec.KeyAndSuffixColumns) - nExtraCols

	neededIndexCols := 0
	compositeIndexCols := 0
	if cap(table.indexColIdx) >= nIndexCols {
		__antithesis_instrumentation__.Notify(567684)
		table.indexColIdx = table.indexColIdx[:nIndexCols]
	} else {
		__antithesis_instrumentation__.Notify(567685)
		table.indexColIdx = make([]int, nIndexCols)
	}
	__antithesis_instrumentation__.Notify(567665)
	for i := 0; i < nIndexCols; i++ {
		__antithesis_instrumentation__.Notify(567686)
		id := spec.KeyAndSuffixColumns[i].ColumnID
		colIdx, ok := table.colIdxMap.Get(id)
		if ok {
			__antithesis_instrumentation__.Notify(567688)
			table.indexColIdx[i] = colIdx
			neededIndexCols++
			table.neededValueColsByIdx.Remove(colIdx)
		} else {
			__antithesis_instrumentation__.Notify(567689)
			table.indexColIdx[i] = -1
		}
		__antithesis_instrumentation__.Notify(567687)
		if spec.KeyAndSuffixColumns[i].IsComposite {
			__antithesis_instrumentation__.Notify(567690)
			compositeIndexCols++
		} else {
			__antithesis_instrumentation__.Notify(567691)
		}
	}
	__antithesis_instrumentation__.Notify(567666)

	rf.mustDecodeIndexKey = neededIndexCols > 0

	table.neededValueCols = len(spec.FetchedColumns) - neededIndexCols + compositeIndexCols

	if cap(table.keyVals) >= nIndexCols {
		__antithesis_instrumentation__.Notify(567692)
		table.keyVals = table.keyVals[:nIndexCols]
	} else {
		__antithesis_instrumentation__.Notify(567693)
		table.keyVals = make([]rowenc.EncDatum, nIndexCols)
	}
	__antithesis_instrumentation__.Notify(567667)

	if nExtraCols > 0 {
		__antithesis_instrumentation__.Notify(567694)

		if cap(table.extraVals) >= nExtraCols {
			__antithesis_instrumentation__.Notify(567695)
			table.extraVals = table.extraVals[:nExtraCols]
		} else {
			__antithesis_instrumentation__.Notify(567696)
			table.extraVals = make([]rowenc.EncDatum, nExtraCols)
		}
	} else {
		__antithesis_instrumentation__.Notify(567697)
	}
	__antithesis_instrumentation__.Notify(567668)

	return nil
}

func (rf *Fetcher) StartScan(
	ctx context.Context,
	txn *kv.Txn,
	spans roachpb.Spans,
	batchBytesLimit rowinfra.BytesLimit,
	rowLimitHint rowinfra.RowLimit,
	traceKV bool,
	forceProductionKVBatchSize bool,
) error {
	__antithesis_instrumentation__.Notify(567698)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(567701)
		return errors.AssertionFailedf("no spans")
	} else {
		__antithesis_instrumentation__.Notify(567702)
	}
	__antithesis_instrumentation__.Notify(567699)

	f, err := makeKVBatchFetcher(
		ctx,
		makeKVBatchFetcherDefaultSendFunc(txn),
		spans,
		rf.reverse,
		batchBytesLimit,
		rf.rowLimitToKeyLimit(rowLimitHint),
		rf.lockStrength,
		rf.lockWaitPolicy,
		rf.lockTimeout,
		rf.kvFetcherMemAcc,
		forceProductionKVBatchSize,
		txn.AdmissionHeader(),
		txn.DB().SQLKVResponseAdmissionQ,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(567703)
		return err
	} else {
		__antithesis_instrumentation__.Notify(567704)
	}
	__antithesis_instrumentation__.Notify(567700)
	return rf.StartScanFrom(ctx, &f, traceKV)
}

var TestingInconsistentScanSleep time.Duration

func (rf *Fetcher) StartInconsistentScan(
	ctx context.Context,
	db *kv.DB,
	initialTimestamp hlc.Timestamp,
	maxTimestampAge time.Duration,
	spans roachpb.Spans,
	batchBytesLimit rowinfra.BytesLimit,
	rowLimitHint rowinfra.RowLimit,
	traceKV bool,
	forceProductionKVBatchSize bool,
	qualityOfService sessiondatapb.QoSLevel,
) error {
	__antithesis_instrumentation__.Notify(567705)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(567712)
		return errors.AssertionFailedf("no spans")
	} else {
		__antithesis_instrumentation__.Notify(567713)
	}
	__antithesis_instrumentation__.Notify(567706)

	txnTimestamp := initialTimestamp
	txnStartTime := timeutil.Now()
	if txnStartTime.Sub(txnTimestamp.GoTime()) >= maxTimestampAge {
		__antithesis_instrumentation__.Notify(567714)
		return errors.Errorf(
			"AS OF SYSTEM TIME: cannot specify timestamp older than %s for this operation",
			maxTimestampAge,
		)
	} else {
		__antithesis_instrumentation__.Notify(567715)
	}
	__antithesis_instrumentation__.Notify(567707)
	txn := kv.NewTxnWithSteppingEnabled(ctx, db, 0, qualityOfService)
	if err := txn.SetFixedTimestamp(ctx, txnTimestamp); err != nil {
		__antithesis_instrumentation__.Notify(567716)
		return err
	} else {
		__antithesis_instrumentation__.Notify(567717)
	}
	__antithesis_instrumentation__.Notify(567708)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(567718)
		log.Infof(ctx, "starting inconsistent scan at timestamp %v", txnTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(567719)
	}
	__antithesis_instrumentation__.Notify(567709)

	sendFn := func(ctx context.Context, ba roachpb.BatchRequest) (*roachpb.BatchResponse, error) {
		__antithesis_instrumentation__.Notify(567720)
		if now := timeutil.Now(); now.Sub(txnTimestamp.GoTime()) >= maxTimestampAge {
			__antithesis_instrumentation__.Notify(567724)

			if err := txn.Commit(ctx); err != nil {
				__antithesis_instrumentation__.Notify(567727)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(567728)
			}
			__antithesis_instrumentation__.Notify(567725)

			txnTimestamp = txnTimestamp.Add(now.Sub(txnStartTime).Nanoseconds(), 0)
			txnStartTime = now
			txn = kv.NewTxnWithSteppingEnabled(ctx, db, 0, qualityOfService)
			if err := txn.SetFixedTimestamp(ctx, txnTimestamp); err != nil {
				__antithesis_instrumentation__.Notify(567729)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(567730)
			}
			__antithesis_instrumentation__.Notify(567726)

			if log.V(1) {
				__antithesis_instrumentation__.Notify(567731)
				log.Infof(ctx, "bumped inconsistent scan timestamp to %v", txnTimestamp)
			} else {
				__antithesis_instrumentation__.Notify(567732)
			}
		} else {
			__antithesis_instrumentation__.Notify(567733)
		}
		__antithesis_instrumentation__.Notify(567721)

		res, err := txn.Send(ctx, ba)
		if err != nil {
			__antithesis_instrumentation__.Notify(567734)
			return nil, err.GoError()
		} else {
			__antithesis_instrumentation__.Notify(567735)
		}
		__antithesis_instrumentation__.Notify(567722)
		if TestingInconsistentScanSleep != 0 {
			__antithesis_instrumentation__.Notify(567736)
			time.Sleep(TestingInconsistentScanSleep)
		} else {
			__antithesis_instrumentation__.Notify(567737)
		}
		__antithesis_instrumentation__.Notify(567723)
		return res, nil
	}
	__antithesis_instrumentation__.Notify(567710)

	f, err := makeKVBatchFetcher(
		ctx,
		sendFunc(sendFn),
		spans,
		rf.reverse,
		batchBytesLimit,
		rf.rowLimitToKeyLimit(rowLimitHint),
		rf.lockStrength,
		rf.lockWaitPolicy,
		rf.lockTimeout,
		rf.kvFetcherMemAcc,
		forceProductionKVBatchSize,
		txn.AdmissionHeader(),
		txn.DB().SQLKVResponseAdmissionQ,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(567738)
		return err
	} else {
		__antithesis_instrumentation__.Notify(567739)
	}
	__antithesis_instrumentation__.Notify(567711)
	return rf.StartScanFrom(ctx, &f, traceKV)
}

func (rf *Fetcher) rowLimitToKeyLimit(rowLimitHint rowinfra.RowLimit) rowinfra.KeyLimit {
	__antithesis_instrumentation__.Notify(567740)
	if rowLimitHint == 0 {
		__antithesis_instrumentation__.Notify(567742)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(567743)
	}
	__antithesis_instrumentation__.Notify(567741)

	return rowinfra.KeyLimit(int64(rowLimitHint)*int64(rf.table.spec.MaxKeysPerRow) + 1)
}

func (rf *Fetcher) StartScanFrom(ctx context.Context, f KVBatchFetcher, traceKV bool) error {
	__antithesis_instrumentation__.Notify(567744)
	rf.traceKV = traceKV
	rf.indexKey = nil
	if rf.kvFetcher != nil {
		__antithesis_instrumentation__.Notify(567746)
		rf.kvFetcher.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(567747)
	}
	__antithesis_instrumentation__.Notify(567745)
	rf.kvFetcher = newKVFetcher(f)
	rf.kvEnd = false

	_, err := rf.nextKey(ctx)
	return err
}

func (rf *Fetcher) setNextKV(kv roachpb.KeyValue, needsCopy bool) {
	__antithesis_instrumentation__.Notify(567748)
	if !needsCopy {
		__antithesis_instrumentation__.Notify(567750)
		rf.kv = kv
		return
	} else {
		__antithesis_instrumentation__.Notify(567751)
	}
	__antithesis_instrumentation__.Notify(567749)

	kvCopy := roachpb.KeyValue{}
	kvCopy.Key = make(roachpb.Key, len(kv.Key))
	copy(kvCopy.Key, kv.Key)
	kvCopy.Value.RawBytes = make([]byte, len(kv.Value.RawBytes))
	copy(kvCopy.Value.RawBytes, kv.Value.RawBytes)
	kvCopy.Value.Timestamp = kv.Value.Timestamp
	rf.kv = kvCopy
}

func (rf *Fetcher) nextKey(ctx context.Context) (newRow bool, _ error) {
	__antithesis_instrumentation__.Notify(567752)
	ok, kv, finalReferenceToBatch, err := rf.kvFetcher.NextKV(ctx, rf.mvccDecodeStrategy)
	if err != nil {
		__antithesis_instrumentation__.Notify(567757)
		return false, ConvertFetchError(&rf.table.spec, err)
	} else {
		__antithesis_instrumentation__.Notify(567758)
	}
	__antithesis_instrumentation__.Notify(567753)
	rf.setNextKV(kv, finalReferenceToBatch)

	if !ok {
		__antithesis_instrumentation__.Notify(567759)

		rf.kvEnd = true
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(567760)
	}
	__antithesis_instrumentation__.Notify(567754)

	unchangedPrefix := rf.table.spec.MaxKeysPerRow > 1 && func() bool {
		__antithesis_instrumentation__.Notify(567761)
		return rf.indexKey != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(567762)
		return bytes.HasPrefix(rf.kv.Key, rf.indexKey) == true
	}() == true
	if unchangedPrefix {
		__antithesis_instrumentation__.Notify(567763)

		rf.keyRemainingBytes = rf.kv.Key[len(rf.indexKey):]
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(567764)
	}
	__antithesis_instrumentation__.Notify(567755)

	if rf.mustDecodeIndexKey {
		__antithesis_instrumentation__.Notify(567765)
		var foundNull bool
		rf.keyRemainingBytes, foundNull, err = rf.DecodeIndexKey(rf.kv.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(567767)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(567768)
		}
		__antithesis_instrumentation__.Notify(567766)

		if foundNull && func() bool {
			__antithesis_instrumentation__.Notify(567769)
			return rf.table.spec.IsSecondaryIndex == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(567770)
			return rf.table.spec.IsUniqueIndex == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(567771)
			return rf.table.spec.MaxKeysPerRow > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(567772)
			for i := 0; i < int(rf.table.spec.NumKeySuffixColumns); i++ {
				__antithesis_instrumentation__.Notify(567773)
				var err error

				rf.keyRemainingBytes, err = keyside.Skip(rf.keyRemainingBytes)
				if err != nil {
					__antithesis_instrumentation__.Notify(567774)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(567775)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(567776)
		}
	} else {
		__antithesis_instrumentation__.Notify(567777)

		prefixLen, err := keys.GetRowPrefixLength(rf.kv.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(567779)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(567780)
		}
		__antithesis_instrumentation__.Notify(567778)

		rf.keyRemainingBytes = rf.kv.Key[prefixLen:]
	}
	__antithesis_instrumentation__.Notify(567756)

	rf.indexKey = nil
	return true, nil
}

func (rf *Fetcher) prettyKeyDatums(
	cols []descpb.IndexFetchSpec_KeyColumn, vals []rowenc.EncDatum,
) string {
	__antithesis_instrumentation__.Notify(567781)
	var buf strings.Builder
	for i, v := range vals {
		__antithesis_instrumentation__.Notify(567783)
		buf.WriteByte('/')
		if err := v.EnsureDecoded(cols[i].Type, rf.alloc); err != nil {
			__antithesis_instrumentation__.Notify(567784)
			buf.WriteByte('?')
		} else {
			__antithesis_instrumentation__.Notify(567785)
			buf.WriteString(v.Datum.String())
		}
	}
	__antithesis_instrumentation__.Notify(567782)
	return buf.String()
}

func (rf *Fetcher) DecodeIndexKey(key roachpb.Key) (remaining []byte, foundNull bool, err error) {
	__antithesis_instrumentation__.Notify(567786)
	key = key[rf.table.spec.KeyPrefixLength:]
	return rowenc.DecodeKeyValsUsingSpec(rf.table.spec.KeyAndSuffixColumns, key, rf.table.keyVals)
}

func (rf *Fetcher) processKV(
	ctx context.Context, kv roachpb.KeyValue,
) (prettyKey string, prettyValue string, err error) {
	__antithesis_instrumentation__.Notify(567787)
	table := &rf.table

	if rf.traceKV {
		__antithesis_instrumentation__.Notify(567794)
		prettyKey = fmt.Sprintf(
			"/%s/%s%s",
			table.spec.TableName,
			table.spec.IndexName,
			rf.prettyKeyDatums(table.spec.KeyAndSuffixColumns, table.keyVals),
		)
	} else {
		__antithesis_instrumentation__.Notify(567795)
	}
	__antithesis_instrumentation__.Notify(567788)

	if rf.indexKey == nil {
		__antithesis_instrumentation__.Notify(567796)

		rf.indexKey = []byte(kv.Key[:len(kv.Key)-len(rf.keyRemainingBytes)])

		for idx, ok := table.neededValueColsByIdx.Next(0); ok; idx, ok = table.neededValueColsByIdx.Next(idx + 1) {
			__antithesis_instrumentation__.Notify(567799)
			table.row[idx].UnsetDatum()
		}
		__antithesis_instrumentation__.Notify(567797)

		for i := range table.keyVals {
			__antithesis_instrumentation__.Notify(567800)
			if idx := table.indexColIdx[i]; idx != -1 {
				__antithesis_instrumentation__.Notify(567801)
				table.row[idx] = table.keyVals[i]
			} else {
				__antithesis_instrumentation__.Notify(567802)
			}
		}
		__antithesis_instrumentation__.Notify(567798)

		rf.valueColsFound = 0

		table.rowLastModified = hlc.Timestamp{}

		table.rowIsDeleted = len(kv.Value.RawBytes) == 0
	} else {
		__antithesis_instrumentation__.Notify(567803)
	}
	__antithesis_instrumentation__.Notify(567789)

	if table.rowLastModified.Less(kv.Value.Timestamp) {
		__antithesis_instrumentation__.Notify(567804)
		table.rowLastModified = kv.Value.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(567805)
	}
	__antithesis_instrumentation__.Notify(567790)

	if len(table.spec.FetchedColumns) == 0 {
		__antithesis_instrumentation__.Notify(567806)

		if rf.traceKV {
			__antithesis_instrumentation__.Notify(567808)
			prettyValue = "<undecoded>"
		} else {
			__antithesis_instrumentation__.Notify(567809)
		}
		__antithesis_instrumentation__.Notify(567807)
		return prettyKey, prettyValue, nil
	} else {
		__antithesis_instrumentation__.Notify(567810)
	}
	__antithesis_instrumentation__.Notify(567791)

	if table.spec.EncodingType == descpb.PrimaryIndexEncoding && func() bool {
		__antithesis_instrumentation__.Notify(567811)
		return len(rf.keyRemainingBytes) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(567812)

		switch kv.Value.GetTag() {
		case roachpb.ValueType_TUPLE:
			__antithesis_instrumentation__.Notify(567814)

			var tupleBytes []byte
			tupleBytes, err = kv.Value.GetTuple()
			if err != nil {
				__antithesis_instrumentation__.Notify(567818)
				break
			} else {
				__antithesis_instrumentation__.Notify(567819)
			}
			__antithesis_instrumentation__.Notify(567815)
			prettyKey, prettyValue, err = rf.processValueBytes(ctx, table, kv, tupleBytes, prettyKey)
		default:
			__antithesis_instrumentation__.Notify(567816)
			var familyID uint64
			_, familyID, err = encoding.DecodeUvarintAscending(rf.keyRemainingBytes)
			if err != nil {
				__antithesis_instrumentation__.Notify(567820)
				return "", "", scrub.WrapError(scrub.IndexKeyDecodingError, err)
			} else {
				__antithesis_instrumentation__.Notify(567821)
			}
			__antithesis_instrumentation__.Notify(567817)

			if familyID != 0 {
				__antithesis_instrumentation__.Notify(567822)

				var defaultColumnID descpb.ColumnID
				for _, f := range table.spec.FamilyDefaultColumns {
					__antithesis_instrumentation__.Notify(567825)
					if f.FamilyID == descpb.FamilyID(familyID) {
						__antithesis_instrumentation__.Notify(567826)
						defaultColumnID = f.DefaultColumnID
						break
					} else {
						__antithesis_instrumentation__.Notify(567827)
					}
				}
				__antithesis_instrumentation__.Notify(567823)
				if defaultColumnID == 0 {
					__antithesis_instrumentation__.Notify(567828)
					return "", "", errors.Errorf("single entry value with no default column id")
				} else {
					__antithesis_instrumentation__.Notify(567829)
				}
				__antithesis_instrumentation__.Notify(567824)

				prettyKey, prettyValue, err = rf.processValueSingle(ctx, table, defaultColumnID, kv, prettyKey)
			} else {
				__antithesis_instrumentation__.Notify(567830)
			}
		}
		__antithesis_instrumentation__.Notify(567813)
		if err != nil {
			__antithesis_instrumentation__.Notify(567831)
			return "", "", scrub.WrapError(scrub.IndexValueDecodingError, err)
		} else {
			__antithesis_instrumentation__.Notify(567832)
		}
	} else {
		__antithesis_instrumentation__.Notify(567833)
		tag := kv.Value.GetTag()
		var valueBytes []byte
		switch tag {
		case roachpb.ValueType_BYTES:
			__antithesis_instrumentation__.Notify(567835)

			valueBytes, err = kv.Value.GetBytes()
			if err != nil {
				__antithesis_instrumentation__.Notify(567839)
				return "", "", scrub.WrapError(scrub.IndexValueDecodingError, err)
			} else {
				__antithesis_instrumentation__.Notify(567840)
			}
			__antithesis_instrumentation__.Notify(567836)
			if len(table.extraVals) > 0 {
				__antithesis_instrumentation__.Notify(567841)
				extraCols := table.spec.KeySuffixColumns()

				var err error
				valueBytes, _, err = rowenc.DecodeKeyValsUsingSpec(
					extraCols,
					valueBytes,
					table.extraVals,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(567844)
					return "", "", scrub.WrapError(scrub.SecondaryIndexKeyExtraValueDecodingError, err)
				} else {
					__antithesis_instrumentation__.Notify(567845)
				}
				__antithesis_instrumentation__.Notify(567842)
				for i := range extraCols {
					__antithesis_instrumentation__.Notify(567846)
					if idx, ok := table.colIdxMap.Get(extraCols[i].ColumnID); ok {
						__antithesis_instrumentation__.Notify(567847)
						table.row[idx] = table.extraVals[i]
					} else {
						__antithesis_instrumentation__.Notify(567848)
					}
				}
				__antithesis_instrumentation__.Notify(567843)
				if rf.traceKV {
					__antithesis_instrumentation__.Notify(567849)
					prettyValue = rf.prettyKeyDatums(extraCols, table.extraVals)
				} else {
					__antithesis_instrumentation__.Notify(567850)
				}
			} else {
				__antithesis_instrumentation__.Notify(567851)
			}
		case roachpb.ValueType_TUPLE:
			__antithesis_instrumentation__.Notify(567837)
			valueBytes, err = kv.Value.GetTuple()
			if err != nil {
				__antithesis_instrumentation__.Notify(567852)
				return "", "", scrub.WrapError(scrub.IndexValueDecodingError, err)
			} else {
				__antithesis_instrumentation__.Notify(567853)
			}
		default:
			__antithesis_instrumentation__.Notify(567838)
		}
		__antithesis_instrumentation__.Notify(567834)

		if len(valueBytes) > 0 {
			__antithesis_instrumentation__.Notify(567854)
			prettyKey, prettyValue, err = rf.processValueBytes(
				ctx, table, kv, valueBytes, prettyKey,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(567855)
				return "", "", scrub.WrapError(scrub.IndexValueDecodingError, err)
			} else {
				__antithesis_instrumentation__.Notify(567856)
			}
		} else {
			__antithesis_instrumentation__.Notify(567857)
		}
	}
	__antithesis_instrumentation__.Notify(567792)

	if rf.traceKV && func() bool {
		__antithesis_instrumentation__.Notify(567858)
		return prettyValue == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(567859)
		prettyValue = "<undecoded>"
	} else {
		__antithesis_instrumentation__.Notify(567860)
	}
	__antithesis_instrumentation__.Notify(567793)

	return prettyKey, prettyValue, nil
}

func (rf *Fetcher) processValueSingle(
	ctx context.Context,
	table *tableInfo,
	colID descpb.ColumnID,
	kv roachpb.KeyValue,
	prettyKeyPrefix string,
) (prettyKey string, prettyValue string, err error) {
	__antithesis_instrumentation__.Notify(567861)
	prettyKey = prettyKeyPrefix
	idx, ok := table.colIdxMap.Get(colID)
	if !ok {
		__antithesis_instrumentation__.Notify(567868)

		if DebugRowFetch {
			__antithesis_instrumentation__.Notify(567870)
			log.Infof(ctx, "Scan %s -> [%d] (skipped)", kv.Key, colID)
		} else {
			__antithesis_instrumentation__.Notify(567871)
		}
		__antithesis_instrumentation__.Notify(567869)
		return prettyKey, "", nil
	} else {
		__antithesis_instrumentation__.Notify(567872)
	}
	__antithesis_instrumentation__.Notify(567862)

	if rf.traceKV {
		__antithesis_instrumentation__.Notify(567873)
		prettyKey = fmt.Sprintf("%s/%s", prettyKey, table.spec.FetchedColumns[idx].Name)
	} else {
		__antithesis_instrumentation__.Notify(567874)
	}
	__antithesis_instrumentation__.Notify(567863)
	if len(kv.Value.RawBytes) == 0 {
		__antithesis_instrumentation__.Notify(567875)
		return prettyKey, "", nil
	} else {
		__antithesis_instrumentation__.Notify(567876)
	}
	__antithesis_instrumentation__.Notify(567864)
	typ := table.spec.FetchedColumns[idx].Type

	value, err := valueside.UnmarshalLegacy(rf.alloc, typ, kv.Value)
	if err != nil {
		__antithesis_instrumentation__.Notify(567877)
		return "", "", err
	} else {
		__antithesis_instrumentation__.Notify(567878)
	}
	__antithesis_instrumentation__.Notify(567865)
	if rf.traceKV {
		__antithesis_instrumentation__.Notify(567879)
		prettyValue = value.String()
	} else {
		__antithesis_instrumentation__.Notify(567880)
	}
	__antithesis_instrumentation__.Notify(567866)
	table.row[idx] = rowenc.DatumToEncDatum(typ, value)
	if DebugRowFetch {
		__antithesis_instrumentation__.Notify(567881)
		log.Infof(ctx, "Scan %s -> %v", kv.Key, value)
	} else {
		__antithesis_instrumentation__.Notify(567882)
	}
	__antithesis_instrumentation__.Notify(567867)
	return prettyKey, prettyValue, nil
}

func (rf *Fetcher) processValueBytes(
	ctx context.Context,
	table *tableInfo,
	kv roachpb.KeyValue,
	valueBytes []byte,
	prettyKeyPrefix string,
) (prettyKey string, prettyValue string, err error) {
	__antithesis_instrumentation__.Notify(567883)
	prettyKey = prettyKeyPrefix
	if rf.traceKV {
		__antithesis_instrumentation__.Notify(567887)
		if rf.prettyValueBuf == nil {
			__antithesis_instrumentation__.Notify(567889)
			rf.prettyValueBuf = &bytes.Buffer{}
		} else {
			__antithesis_instrumentation__.Notify(567890)
		}
		__antithesis_instrumentation__.Notify(567888)
		rf.prettyValueBuf.Reset()
	} else {
		__antithesis_instrumentation__.Notify(567891)
	}
	__antithesis_instrumentation__.Notify(567884)

	var colIDDiff uint32
	var lastColID descpb.ColumnID
	var typeOffset, dataOffset int
	var typ encoding.Type
	for len(valueBytes) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(567892)
		return rf.valueColsFound < table.neededValueCols == true
	}() == true {
		__antithesis_instrumentation__.Notify(567893)
		typeOffset, dataOffset, colIDDiff, typ, err = encoding.DecodeValueTag(valueBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(567899)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(567900)
		}
		__antithesis_instrumentation__.Notify(567894)
		colID := lastColID + descpb.ColumnID(colIDDiff)
		lastColID = colID
		idx, ok := table.colIdxMap.Get(colID)
		if !ok {
			__antithesis_instrumentation__.Notify(567901)

			numBytes, err := encoding.PeekValueLengthWithOffsetsAndType(valueBytes, dataOffset, typ)
			if err != nil {
				__antithesis_instrumentation__.Notify(567904)
				return "", "", err
			} else {
				__antithesis_instrumentation__.Notify(567905)
			}
			__antithesis_instrumentation__.Notify(567902)
			valueBytes = valueBytes[numBytes:]
			if DebugRowFetch {
				__antithesis_instrumentation__.Notify(567906)
				log.Infof(ctx, "Scan %s -> [%d] (skipped)", kv.Key, colID)
			} else {
				__antithesis_instrumentation__.Notify(567907)
			}
			__antithesis_instrumentation__.Notify(567903)
			continue
		} else {
			__antithesis_instrumentation__.Notify(567908)
		}
		__antithesis_instrumentation__.Notify(567895)

		if rf.traceKV {
			__antithesis_instrumentation__.Notify(567909)
			prettyKey = fmt.Sprintf("%s/%s", prettyKey, table.spec.FetchedColumns[idx].Name)
		} else {
			__antithesis_instrumentation__.Notify(567910)
		}
		__antithesis_instrumentation__.Notify(567896)

		var encValue rowenc.EncDatum
		encValue, valueBytes, err = rowenc.EncDatumValueFromBufferWithOffsetsAndType(valueBytes, typeOffset,
			dataOffset, typ)
		if err != nil {
			__antithesis_instrumentation__.Notify(567911)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(567912)
		}
		__antithesis_instrumentation__.Notify(567897)
		if rf.traceKV {
			__antithesis_instrumentation__.Notify(567913)
			err := encValue.EnsureDecoded(table.spec.FetchedColumns[idx].Type, rf.alloc)
			if err != nil {
				__antithesis_instrumentation__.Notify(567915)
				return "", "", err
			} else {
				__antithesis_instrumentation__.Notify(567916)
			}
			__antithesis_instrumentation__.Notify(567914)
			fmt.Fprintf(rf.prettyValueBuf, "/%v", encValue.Datum)
		} else {
			__antithesis_instrumentation__.Notify(567917)
		}
		__antithesis_instrumentation__.Notify(567898)
		table.row[idx] = encValue
		rf.valueColsFound++
		if DebugRowFetch {
			__antithesis_instrumentation__.Notify(567918)
			log.Infof(ctx, "Scan %d -> %v", idx, encValue)
		} else {
			__antithesis_instrumentation__.Notify(567919)
		}
	}
	__antithesis_instrumentation__.Notify(567885)
	if rf.traceKV {
		__antithesis_instrumentation__.Notify(567920)
		prettyValue = rf.prettyValueBuf.String()
	} else {
		__antithesis_instrumentation__.Notify(567921)
	}
	__antithesis_instrumentation__.Notify(567886)
	return prettyKey, prettyValue, nil
}

func (rf *Fetcher) NextRow(ctx context.Context) (row rowenc.EncDatumRow, err error) {
	__antithesis_instrumentation__.Notify(567922)
	if rf.kvEnd {
		__antithesis_instrumentation__.Notify(567924)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(567925)
	}
	__antithesis_instrumentation__.Notify(567923)

	for {
		__antithesis_instrumentation__.Notify(567926)
		prettyKey, prettyVal, err := rf.processKV(ctx, rf.kv)
		if err != nil {
			__antithesis_instrumentation__.Notify(567930)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(567931)
		}
		__antithesis_instrumentation__.Notify(567927)
		if rf.traceKV {
			__antithesis_instrumentation__.Notify(567932)
			log.VEventf(ctx, 2, "fetched: %s -> %s", prettyKey, prettyVal)
		} else {
			__antithesis_instrumentation__.Notify(567933)
		}
		__antithesis_instrumentation__.Notify(567928)

		rowDone, err := rf.nextKey(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(567934)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(567935)
		}
		__antithesis_instrumentation__.Notify(567929)
		if rowDone {
			__antithesis_instrumentation__.Notify(567936)
			err := rf.finalizeRow()
			return rf.table.row, err
		} else {
			__antithesis_instrumentation__.Notify(567937)
		}
	}
}

func (rf *Fetcher) NextRowInto(
	ctx context.Context, destination rowenc.EncDatumRow, colIdxMap catalog.TableColMap,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(567938)
	row, err := rf.NextRow(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(567942)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(567943)
	}
	__antithesis_instrumentation__.Notify(567939)
	if row == nil {
		__antithesis_instrumentation__.Notify(567944)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(567945)
	}
	__antithesis_instrumentation__.Notify(567940)

	for i := range rf.table.spec.FetchedColumns {
		__antithesis_instrumentation__.Notify(567946)
		if ord, ok := colIdxMap.Get(rf.table.spec.FetchedColumns[i].ColumnID); ok {
			__antithesis_instrumentation__.Notify(567947)
			destination[ord] = row[i]
		} else {
			__antithesis_instrumentation__.Notify(567948)
		}
	}
	__antithesis_instrumentation__.Notify(567941)
	return true, nil
}

func (rf *Fetcher) NextRowDecoded(ctx context.Context) (datums tree.Datums, err error) {
	__antithesis_instrumentation__.Notify(567949)
	row, err := rf.NextRow(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(567953)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567954)
	}
	__antithesis_instrumentation__.Notify(567950)
	if row == nil {
		__antithesis_instrumentation__.Notify(567955)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(567956)
	}
	__antithesis_instrumentation__.Notify(567951)

	for i, encDatum := range row {
		__antithesis_instrumentation__.Notify(567957)
		if encDatum.IsUnset() {
			__antithesis_instrumentation__.Notify(567960)
			rf.table.decodedRow[i] = tree.DNull
			continue
		} else {
			__antithesis_instrumentation__.Notify(567961)
		}
		__antithesis_instrumentation__.Notify(567958)
		if err := encDatum.EnsureDecoded(rf.table.spec.FetchedColumns[i].Type, rf.alloc); err != nil {
			__antithesis_instrumentation__.Notify(567962)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(567963)
		}
		__antithesis_instrumentation__.Notify(567959)
		rf.table.decodedRow[i] = encDatum.Datum
	}
	__antithesis_instrumentation__.Notify(567952)

	return rf.table.decodedRow, nil
}

func (rf *Fetcher) NextRowDecodedInto(
	ctx context.Context, destination tree.Datums, colIdxMap catalog.TableColMap,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(567964)
	row, err := rf.NextRow(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(567968)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(567969)
	}
	__antithesis_instrumentation__.Notify(567965)
	if row == nil {
		__antithesis_instrumentation__.Notify(567970)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(567971)
	}
	__antithesis_instrumentation__.Notify(567966)

	for i := range rf.table.spec.FetchedColumns {
		__antithesis_instrumentation__.Notify(567972)
		col := &rf.table.spec.FetchedColumns[i]
		ord, ok := colIdxMap.Get(col.ColumnID)
		if !ok {
			__antithesis_instrumentation__.Notify(567976)

			continue
		} else {
			__antithesis_instrumentation__.Notify(567977)
		}
		__antithesis_instrumentation__.Notify(567973)
		encDatum := row[i]
		if encDatum.IsUnset() {
			__antithesis_instrumentation__.Notify(567978)
			destination[ord] = tree.DNull
			continue
		} else {
			__antithesis_instrumentation__.Notify(567979)
		}
		__antithesis_instrumentation__.Notify(567974)
		if err := encDatum.EnsureDecoded(col.Type, rf.alloc); err != nil {
			__antithesis_instrumentation__.Notify(567980)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(567981)
		}
		__antithesis_instrumentation__.Notify(567975)
		destination[ord] = encDatum.Datum
	}
	__antithesis_instrumentation__.Notify(567967)

	return true, nil
}

func (rf *Fetcher) RowLastModified() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(567982)
	return rf.table.rowLastModified
}

func (rf *Fetcher) RowIsDeleted() bool {
	__antithesis_instrumentation__.Notify(567983)
	return rf.table.rowIsDeleted
}

func (rf *Fetcher) finalizeRow() error {
	__antithesis_instrumentation__.Notify(567984)
	table := &rf.table

	if table.timestampOutputIdx != noOutputColumn {
		__antithesis_instrumentation__.Notify(567988)

		dec := rf.alloc.NewDDecimal(tree.DDecimal{Decimal: tree.TimestampToDecimal(rf.RowLastModified())})
		table.row[table.timestampOutputIdx] = rowenc.EncDatum{Datum: dec}
	} else {
		__antithesis_instrumentation__.Notify(567989)
	}
	__antithesis_instrumentation__.Notify(567985)
	if table.oidOutputIdx != noOutputColumn {
		__antithesis_instrumentation__.Notify(567990)
		table.row[table.oidOutputIdx] = rowenc.EncDatum{Datum: table.tableOid}
	} else {
		__antithesis_instrumentation__.Notify(567991)
	}
	__antithesis_instrumentation__.Notify(567986)

	for i := range table.spec.FetchedColumns {
		__antithesis_instrumentation__.Notify(567992)
		col := &table.spec.FetchedColumns[i]
		if rf.valueColsFound == table.neededValueCols {
			__antithesis_instrumentation__.Notify(567994)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(567995)
		}
		__antithesis_instrumentation__.Notify(567993)
		if table.row[i].IsUnset() {
			__antithesis_instrumentation__.Notify(567996)

			if col.IsNonNullable && func() bool {
				__antithesis_instrumentation__.Notify(567998)
				return !table.rowIsDeleted == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(567999)
				return !rf.IgnoreUnexpectedNulls == true
			}() == true {
				__antithesis_instrumentation__.Notify(568000)
				var indexColValues []string
				for _, idx := range table.indexColIdx {
					__antithesis_instrumentation__.Notify(568003)
					if idx != -1 {
						__antithesis_instrumentation__.Notify(568004)
						indexColValues = append(indexColValues, table.row[idx].String(table.spec.FetchedColumns[idx].Type))
					} else {
						__antithesis_instrumentation__.Notify(568005)
						indexColValues = append(indexColValues, "?")
					}
				}
				__antithesis_instrumentation__.Notify(568001)
				var indexColNames []string
				for i := range table.spec.KeyFullColumns() {
					__antithesis_instrumentation__.Notify(568006)
					indexColNames = append(indexColNames, table.spec.KeyAndSuffixColumns[i].Name)
				}
				__antithesis_instrumentation__.Notify(568002)
				return errors.AssertionFailedf(
					"Non-nullable column \"%s:%s\" with no value! Index scanned was %q with the index key columns (%s) and the values (%s)",
					table.spec.TableName, col.Name, table.spec.IndexName,
					strings.Join(indexColNames, ","), strings.Join(indexColValues, ","))
			} else {
				__antithesis_instrumentation__.Notify(568007)
			}
			__antithesis_instrumentation__.Notify(567997)
			table.row[i] = rowenc.EncDatum{
				Datum: tree.DNull,
			}

			rf.valueColsFound++
		} else {
			__antithesis_instrumentation__.Notify(568008)
		}
	}
	__antithesis_instrumentation__.Notify(567987)
	return nil
}

func (rf *Fetcher) Key() roachpb.Key {
	__antithesis_instrumentation__.Notify(568009)
	return rf.kv.Key
}

func (rf *Fetcher) PartialKey(nCols int) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(568010)
	if rf.kv.Key == nil {
		__antithesis_instrumentation__.Notify(568013)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(568014)
	}
	__antithesis_instrumentation__.Notify(568011)
	partialKeyLength := int(rf.table.spec.KeyPrefixLength)
	for consumedCols := 0; consumedCols < nCols; consumedCols++ {
		__antithesis_instrumentation__.Notify(568015)
		l, err := encoding.PeekLength(rf.kv.Key[partialKeyLength:])
		if err != nil {
			__antithesis_instrumentation__.Notify(568017)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568018)
		}
		__antithesis_instrumentation__.Notify(568016)
		partialKeyLength += l
	}
	__antithesis_instrumentation__.Notify(568012)
	return rf.kv.Key[:partialKeyLength], nil
}

func (rf *Fetcher) GetBytesRead() int64 {
	__antithesis_instrumentation__.Notify(568019)
	return rf.kvFetcher.GetBytesRead()
}
