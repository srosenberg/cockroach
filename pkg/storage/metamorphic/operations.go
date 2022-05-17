package metamorphic

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

type opReference struct {
	generator *opGenerator
	args      []string
}

type opRun struct {
	name string
	op   mvccOp

	args []string

	lineNum        uint64
	expectedOutput string
}

type mvccOp interface {
	run(ctx context.Context) string
}

type opGenerator struct {
	name string

	generate func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp

	dependentOps func(m *metaTestRunner, args ...string) []opReference

	operands []operandType

	weight int

	isOpener bool
}

func closeItersOnBatch(m *metaTestRunner, reader readWriterID) (results []opReference) {
	__antithesis_instrumentation__.Notify(639953)

	if reader == "engine" {
		__antithesis_instrumentation__.Notify(639956)
		return
	} else {
		__antithesis_instrumentation__.Notify(639957)
	}
	__antithesis_instrumentation__.Notify(639954)

	for _, iter := range m.iterGenerator.readerToIter[reader] {
		__antithesis_instrumentation__.Notify(639958)
		results = append(results, opReference{
			generator: m.nameToGenerator["iterator_close"],
			args:      []string{string(iter)},
		})
	}
	__antithesis_instrumentation__.Notify(639955)
	return
}

func generateMVCCScan(
	ctx context.Context, m *metaTestRunner, reverse bool, inconsistent bool, args []string,
) *mvccScanOp {
	__antithesis_instrumentation__.Notify(639959)
	key := m.keyGenerator.parse(args[0])
	endKey := m.keyGenerator.parse(args[1])
	if endKey.Less(key) {
		__antithesis_instrumentation__.Notify(639962)
		key, endKey = endKey, key
	} else {
		__antithesis_instrumentation__.Notify(639963)
	}
	__antithesis_instrumentation__.Notify(639960)
	var ts hlc.Timestamp
	var txn txnID
	if inconsistent {
		__antithesis_instrumentation__.Notify(639964)
		ts = m.pastTSGenerator.parse(args[2])
	} else {
		__antithesis_instrumentation__.Notify(639965)
		txn = txnID(args[2])
	}
	__antithesis_instrumentation__.Notify(639961)
	maxKeys := int64(m.floatGenerator.parse(args[3]) * 32)
	targetBytes := int64(m.floatGenerator.parse(args[4]) * (1 << 20))
	targetBytesAvoidExcess := m.boolGenerator.parse(args[5])
	allowEmpty := m.boolGenerator.parse(args[6])
	return &mvccScanOp{
		m:                      m,
		key:                    key.Key,
		endKey:                 endKey.Key,
		ts:                     ts,
		txn:                    txn,
		inconsistent:           inconsistent,
		reverse:                reverse,
		maxKeys:                maxKeys,
		targetBytes:            targetBytes,
		targetBytesAvoidExcess: targetBytesAvoidExcess,
		allowEmpty:             allowEmpty,
	}
}

func printIterState(iter storage.MVCCIterator) string {
	__antithesis_instrumentation__.Notify(639966)
	if ok, err := iter.Valid(); !ok || func() bool {
		__antithesis_instrumentation__.Notify(639968)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(639969)
		if err != nil {
			__antithesis_instrumentation__.Notify(639971)
			return fmt.Sprintf("valid = %v, err = %s", ok, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(639972)
		}
		__antithesis_instrumentation__.Notify(639970)
		return "valid = false"
	} else {
		__antithesis_instrumentation__.Notify(639973)
	}
	__antithesis_instrumentation__.Notify(639967)
	return fmt.Sprintf("key = %s", iter.UnsafeKey().String())
}

func addKeyToLockSpans(txn *roachpb.Transaction, key roachpb.Key) {
	__antithesis_instrumentation__.Notify(639974)

	newLockSpans := make([]roachpb.Span, 0, len(txn.LockSpans)+1)
	newLockSpans = append(newLockSpans, txn.LockSpans...)
	newLockSpans = append(newLockSpans, roachpb.Span{
		Key: key,
	})
	txn.LockSpans, _ = roachpb.MergeSpans(&newLockSpans)
}

type mvccGetOp struct {
	m            *metaTestRunner
	reader       readWriterID
	key          roachpb.Key
	ts           hlc.Timestamp
	txn          txnID
	inconsistent bool
}

func (m mvccGetOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(639975)
	reader := m.m.getReadWriter(m.reader)
	var txn *roachpb.Transaction
	if !m.inconsistent {
		__antithesis_instrumentation__.Notify(639978)
		txn = m.m.getTxn(m.txn)
		m.ts = txn.ReadTimestamp
	} else {
		__antithesis_instrumentation__.Notify(639979)
	}
	__antithesis_instrumentation__.Notify(639976)

	val, intent, err := storage.MVCCGet(ctx, reader, m.key, m.ts, storage.MVCCGetOptions{
		Inconsistent: m.inconsistent,
		Tombstones:   true,
		Txn:          txn,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(639980)
		return fmt.Sprintf("error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(639981)
	}
	__antithesis_instrumentation__.Notify(639977)
	return fmt.Sprintf("val = %v, intent = %v", val, intent)
}

type mvccPutOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	value  roachpb.Value
	txn    txnID
}

func (m mvccPutOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(639982)
	txn := m.m.getTxn(m.txn)
	txn.Sequence++
	writer := m.m.getReadWriter(m.writer)

	err := storage.MVCCPut(ctx, writer, nil, m.key, txn.WriteTimestamp, m.value, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(639984)
		return fmt.Sprintf("error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(639985)
	}
	__antithesis_instrumentation__.Notify(639983)

	addKeyToLockSpans(txn, m.key)
	return "ok"
}

type mvccCPutOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	value  roachpb.Value
	expVal []byte
	txn    txnID
}

func (m mvccCPutOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(639986)
	txn := m.m.getTxn(m.txn)
	writer := m.m.getReadWriter(m.writer)
	txn.Sequence++

	err := storage.MVCCConditionalPut(ctx, writer, nil, m.key, txn.WriteTimestamp, m.value, m.expVal, true, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(639988)
		return fmt.Sprintf("error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(639989)
	}
	__antithesis_instrumentation__.Notify(639987)

	addKeyToLockSpans(txn, m.key)
	return "ok"
}

type mvccInitPutOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	value  roachpb.Value
	txn    txnID
}

func (m mvccInitPutOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(639990)
	txn := m.m.getTxn(m.txn)
	writer := m.m.getReadWriter(m.writer)
	txn.Sequence++

	err := storage.MVCCInitPut(ctx, writer, nil, m.key, txn.WriteTimestamp, m.value, false, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(639992)
		return fmt.Sprintf("error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(639993)
	}
	__antithesis_instrumentation__.Notify(639991)

	addKeyToLockSpans(txn, m.key)
	return "ok"
}

type mvccDeleteRangeOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	endKey roachpb.Key
	txn    txnID
}

func (m mvccDeleteRangeOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(639994)
	txn := m.m.getTxn(m.txn)
	writer := m.m.getReadWriter(m.writer)
	if m.key.Compare(m.endKey) >= 0 {
		__antithesis_instrumentation__.Notify(639999)

		return "no-op due to no non-conflicting key range"
	} else {
		__antithesis_instrumentation__.Notify(640000)
	}
	__antithesis_instrumentation__.Notify(639995)

	txn.Sequence++

	keys, _, _, err := storage.MVCCDeleteRange(ctx, writer, nil, m.key, m.endKey, 0, txn.WriteTimestamp, txn, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(640001)
		return fmt.Sprintf("error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(640002)
	}
	__antithesis_instrumentation__.Notify(639996)

	for _, key := range keys {
		__antithesis_instrumentation__.Notify(640003)
		addKeyToLockSpans(txn, key)
	}
	__antithesis_instrumentation__.Notify(639997)
	var builder strings.Builder
	fmt.Fprintf(&builder, "truncated range to delete = %s - %s, deleted keys = ", m.key, m.endKey)
	for i, key := range keys {
		__antithesis_instrumentation__.Notify(640004)
		fmt.Fprintf(&builder, "%s", key)
		if i < len(keys)-1 {
			__antithesis_instrumentation__.Notify(640005)
			fmt.Fprintf(&builder, ", ")
		} else {
			__antithesis_instrumentation__.Notify(640006)
		}
	}
	__antithesis_instrumentation__.Notify(639998)
	return builder.String()
}

type mvccClearTimeRangeOp struct {
	m         *metaTestRunner
	key       roachpb.Key
	endKey    roachpb.Key
	startTime hlc.Timestamp
	endTime   hlc.Timestamp
}

func (m mvccClearTimeRangeOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640007)
	if m.key.Compare(m.endKey) >= 0 {
		__antithesis_instrumentation__.Notify(640010)

		return "no-op due to no non-conflicting key range"
	} else {
		__antithesis_instrumentation__.Notify(640011)
	}
	__antithesis_instrumentation__.Notify(640008)
	span, err := storage.MVCCClearTimeRange(ctx, m.m.engine, &enginepb.MVCCStats{}, m.key, m.endKey,
		m.startTime, m.endTime, math.MaxInt64, math.MaxInt64, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(640012)
		return fmt.Sprintf("error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(640013)
	}
	__antithesis_instrumentation__.Notify(640009)
	return fmt.Sprintf("ok, deleted span = %s - %s, resumeSpan = %v", m.key, m.endKey, span)
}

type mvccDeleteOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	txn    txnID
}

func (m mvccDeleteOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640014)
	txn := m.m.getTxn(m.txn)
	writer := m.m.getReadWriter(m.writer)
	txn.Sequence++

	err := storage.MVCCDelete(ctx, writer, nil, m.key, txn.WriteTimestamp, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(640016)
		return fmt.Sprintf("error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(640017)
	}
	__antithesis_instrumentation__.Notify(640015)

	addKeyToLockSpans(txn, m.key)
	return "ok"
}

type mvccFindSplitKeyOp struct {
	m         *metaTestRunner
	key       roachpb.RKey
	endKey    roachpb.RKey
	splitSize int64
}

func (m mvccFindSplitKeyOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640018)
	splitKey, err := storage.MVCCFindSplitKey(ctx, m.m.engine, m.key, m.endKey, m.splitSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(640020)
		return fmt.Sprintf("error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(640021)
	}
	__antithesis_instrumentation__.Notify(640019)

	return fmt.Sprintf("ok, splitSize = %d, splitKey = %v", m.splitSize, splitKey)
}

type mvccScanOp struct {
	m                      *metaTestRunner
	key                    roachpb.Key
	endKey                 roachpb.Key
	ts                     hlc.Timestamp
	txn                    txnID
	inconsistent           bool
	reverse                bool
	maxKeys                int64
	targetBytes            int64
	targetBytesAvoidExcess bool
	allowEmpty             bool
}

func (m mvccScanOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640022)
	var txn *roachpb.Transaction
	if !m.inconsistent {
		__antithesis_instrumentation__.Notify(640025)
		txn = m.m.getTxn(m.txn)
		m.ts = txn.ReadTimestamp
	} else {
		__antithesis_instrumentation__.Notify(640026)
	}
	__antithesis_instrumentation__.Notify(640023)

	result, err := storage.MVCCScan(ctx, m.m.engine, m.key, m.endKey, m.ts, storage.MVCCScanOptions{
		Inconsistent:           m.inconsistent,
		Tombstones:             true,
		Reverse:                m.reverse,
		Txn:                    txn,
		MaxKeys:                m.maxKeys,
		TargetBytes:            m.targetBytes,
		TargetBytesAvoidExcess: m.targetBytesAvoidExcess,
		AllowEmpty:             m.allowEmpty,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(640027)
		return fmt.Sprintf("error: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(640028)
	}
	__antithesis_instrumentation__.Notify(640024)
	return fmt.Sprintf("kvs = %v, intents = %v, resumeSpan = %v, numBytes = %d, numKeys = %d",
		result.KVs, result.Intents, result.ResumeSpan, result.NumBytes, result.NumKeys)
}

type txnOpenOp struct {
	m  *metaTestRunner
	id txnID
	ts hlc.Timestamp
}

func (t txnOpenOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640029)
	var id uint64
	if _, err := fmt.Sscanf(string(t.id), "t%d", &id); err != nil {
		__antithesis_instrumentation__.Notify(640031)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(640032)
	}
	__antithesis_instrumentation__.Notify(640030)
	txn := &roachpb.Transaction{
		TxnMeta: enginepb.TxnMeta{
			ID:             uuid.FromUint128(uint128.FromInts(0, id)),
			Key:            roachpb.KeyMin,
			WriteTimestamp: t.ts,
			Sequence:       0,
		},
		Name:          string(t.id),
		ReadTimestamp: t.ts,
		Status:        roachpb.PENDING,
	}
	t.m.setTxn(t.id, txn)
	return txn.Name
}

type txnCommitOp struct {
	m  *metaTestRunner
	id txnID
}

func (t txnCommitOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640033)
	txn := t.m.getTxn(t.id)
	txn.Status = roachpb.COMMITTED
	txn.Sequence++

	for _, span := range txn.LockSpans {
		__antithesis_instrumentation__.Notify(640035)
		intent := roachpb.MakeLockUpdate(txn, span)
		intent.Status = roachpb.COMMITTED
		_, err := storage.MVCCResolveWriteIntent(context.TODO(), t.m.engine, nil, intent)
		if err != nil {
			__antithesis_instrumentation__.Notify(640036)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(640037)
		}
	}
	__antithesis_instrumentation__.Notify(640034)
	delete(t.m.openTxns, t.id)
	delete(t.m.openSavepoints, t.id)

	return "ok"
}

type txnAbortOp struct {
	m  *metaTestRunner
	id txnID
}

func (t txnAbortOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640038)
	txn := t.m.getTxn(t.id)
	txn.Status = roachpb.ABORTED

	for _, span := range txn.LockSpans {
		__antithesis_instrumentation__.Notify(640040)
		intent := roachpb.MakeLockUpdate(txn, span)
		intent.Status = roachpb.ABORTED
		_, err := storage.MVCCResolveWriteIntent(context.TODO(), t.m.engine, nil, intent)
		if err != nil {
			__antithesis_instrumentation__.Notify(640041)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(640042)
		}
	}
	__antithesis_instrumentation__.Notify(640039)
	delete(t.m.openTxns, t.id)
	delete(t.m.openSavepoints, t.id)

	return "ok"
}

type txnCreateSavepointOp struct {
	m  *metaTestRunner
	id txnID

	savepoint int
}

func (t txnCreateSavepointOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640043)
	txn := t.m.getTxn(t.id)
	txn.Sequence++

	if len(t.m.openSavepoints[t.id]) == t.savepoint {
		__antithesis_instrumentation__.Notify(640045)
		t.m.openSavepoints[t.id] = append(t.m.openSavepoints[t.id], txn.Sequence)
	} else {
		__antithesis_instrumentation__.Notify(640046)
		panic(fmt.Sprintf("mismatching savepoint index: %d != %d", len(t.m.openSavepoints[t.id]), t.savepoint))
	}
	__antithesis_instrumentation__.Notify(640044)

	return fmt.Sprintf("savepoint %d", t.savepoint)
}

type txnRollbackSavepointOp struct {
	m  *metaTestRunner
	id txnID

	savepoint int
}

func (t txnRollbackSavepointOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640047)
	txn := t.m.getTxn(t.id)
	txn.Sequence++

	savepoints := t.m.openSavepoints[t.id]
	if len(savepoints) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(640049)
		return t.savepoint > len(savepoints) == true
	}() == true {
		__antithesis_instrumentation__.Notify(640050)
		panic(fmt.Sprintf("got a higher savepoint idx %d than allowed for txn %s", t.savepoint, t.id))
	} else {
		__antithesis_instrumentation__.Notify(640051)
	}
	__antithesis_instrumentation__.Notify(640048)

	ignoredSeqNumRange := enginepb.IgnoredSeqNumRange{
		Start: savepoints[t.savepoint],
		End:   txn.Sequence,
	}
	txn.AddIgnoredSeqNumRange(ignoredSeqNumRange)
	return "ok"
}

type batchOpenOp struct {
	m  *metaTestRunner
	id readWriterID
}

func (b batchOpenOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640052)
	batch := b.m.engine.NewBatch()
	b.m.setReadWriter(b.id, batch)
	return string(b.id)
}

type batchCommitOp struct {
	m  *metaTestRunner
	id readWriterID
}

func (b batchCommitOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640053)
	if b.id == "engine" {
		__antithesis_instrumentation__.Notify(640056)
		return "noop"
	} else {
		__antithesis_instrumentation__.Notify(640057)
	}
	__antithesis_instrumentation__.Notify(640054)
	batch := b.m.getReadWriter(b.id).(storage.Batch)
	if err := batch.Commit(true); err != nil {
		__antithesis_instrumentation__.Notify(640058)
		return err.Error()
	} else {
		__antithesis_instrumentation__.Notify(640059)
	}
	__antithesis_instrumentation__.Notify(640055)
	batch.Close()
	delete(b.m.openBatches, b.id)
	return "ok"
}

type iterOpenOp struct {
	m      *metaTestRunner
	rw     readWriterID
	key    roachpb.Key
	endKey roachpb.Key
	id     iteratorID
}

func (i iterOpenOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640060)
	rw := i.m.getReadWriter(i.rw)
	iter := rw.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{
		Prefix:     false,
		LowerBound: i.key,
		UpperBound: i.endKey.Next(),
	})

	i.m.setIterInfo(i.id, iteratorInfo{
		id:          i.id,
		lowerBound:  i.key,
		iter:        iter,
		isBatchIter: i.rw != "engine",
	})

	if _, ok := rw.(storage.Batch); ok {
		__antithesis_instrumentation__.Notify(640062)

		iter.SeekGE(storage.MakeMVCCMetadataKey(i.key))
	} else {
		__antithesis_instrumentation__.Notify(640063)
	}
	__antithesis_instrumentation__.Notify(640061)

	return string(i.id)
}

type iterCloseOp struct {
	m  *metaTestRunner
	id iteratorID
}

func (i iterCloseOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640064)
	iterInfo := i.m.getIterInfo(i.id)
	iterInfo.iter.Close()
	delete(i.m.openIters, i.id)
	return "ok"
}

type iterSeekOp struct {
	m      *metaTestRunner
	iter   iteratorID
	key    storage.MVCCKey
	seekLT bool
}

func (i iterSeekOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640065)
	iterInfo := i.m.getIterInfo(i.iter)
	iter := iterInfo.iter
	if iterInfo.isBatchIter {
		__antithesis_instrumentation__.Notify(640068)
		if i.seekLT {
			__antithesis_instrumentation__.Notify(640070)
			return "noop due to missing seekLT support in rocksdb batch iterators"
		} else {
			__antithesis_instrumentation__.Notify(640071)
		}
		__antithesis_instrumentation__.Notify(640069)

		lowerBound := iterInfo.lowerBound
		if i.key.Key.Compare(lowerBound) < 0 {
			__antithesis_instrumentation__.Notify(640072)
			i.key.Key = lowerBound
		} else {
			__antithesis_instrumentation__.Notify(640073)
		}
	} else {
		__antithesis_instrumentation__.Notify(640074)
	}
	__antithesis_instrumentation__.Notify(640066)
	if i.seekLT {
		__antithesis_instrumentation__.Notify(640075)
		iter.SeekLT(i.key)
	} else {
		__antithesis_instrumentation__.Notify(640076)
		iter.SeekGE(i.key)
	}
	__antithesis_instrumentation__.Notify(640067)

	return printIterState(iter)
}

type iterNextOp struct {
	m       *metaTestRunner
	iter    iteratorID
	nextKey bool
}

func (i iterNextOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640077)
	iter := i.m.getIterInfo(i.iter).iter

	if ok, err := iter.Valid(); !ok || func() bool {
		__antithesis_instrumentation__.Notify(640080)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(640081)
		if err != nil {
			__antithesis_instrumentation__.Notify(640083)
			return fmt.Sprintf("valid = %v, err = %s", ok, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(640084)
		}
		__antithesis_instrumentation__.Notify(640082)
		return "valid = false"
	} else {
		__antithesis_instrumentation__.Notify(640085)
	}
	__antithesis_instrumentation__.Notify(640078)
	if i.nextKey {
		__antithesis_instrumentation__.Notify(640086)
		iter.NextKey()
	} else {
		__antithesis_instrumentation__.Notify(640087)
		iter.Next()
	}
	__antithesis_instrumentation__.Notify(640079)

	return printIterState(iter)
}

type iterPrevOp struct {
	m    *metaTestRunner
	iter iteratorID
}

func (i iterPrevOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640088)
	iterInfo := i.m.getIterInfo(i.iter)
	iter := iterInfo.iter
	if iterInfo.isBatchIter {
		__antithesis_instrumentation__.Notify(640091)
		return "noop due to missing Prev support in rocksdb batch iterators"
	} else {
		__antithesis_instrumentation__.Notify(640092)
	}
	__antithesis_instrumentation__.Notify(640089)

	if ok, err := iter.Valid(); !ok || func() bool {
		__antithesis_instrumentation__.Notify(640093)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(640094)
		if err != nil {
			__antithesis_instrumentation__.Notify(640096)
			return fmt.Sprintf("valid = %v, err = %s", ok, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(640097)
		}
		__antithesis_instrumentation__.Notify(640095)
		return "valid = false"
	} else {
		__antithesis_instrumentation__.Notify(640098)
	}
	__antithesis_instrumentation__.Notify(640090)
	iter.Prev()

	return printIterState(iter)
}

type clearRangeOp struct {
	m      *metaTestRunner
	key    roachpb.Key
	endKey roachpb.Key
}

func (c clearRangeOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640099)

	if c.key.Compare(c.endKey) >= 0 {
		__antithesis_instrumentation__.Notify(640102)

		return "no-op due to no non-conflicting key range"
	} else {
		__antithesis_instrumentation__.Notify(640103)
	}
	__antithesis_instrumentation__.Notify(640100)
	err := c.m.engine.ClearMVCCRangeAndIntents(c.key, c.endKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(640104)
		return fmt.Sprintf("error: %s", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(640105)
	}
	__antithesis_instrumentation__.Notify(640101)
	return fmt.Sprintf("deleted range = %s - %s", c.key, c.endKey)
}

type compactOp struct {
	m      *metaTestRunner
	key    roachpb.Key
	endKey roachpb.Key
}

func (c compactOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640106)
	err := c.m.engine.CompactRange(c.key, c.endKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(640108)
		return fmt.Sprintf("error: %s", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(640109)
	}
	__antithesis_instrumentation__.Notify(640107)
	return "ok"
}

type ingestOp struct {
	m    *metaTestRunner
	keys []storage.MVCCKey
}

func (i ingestOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640110)
	sstPath := filepath.Join(i.m.path, "ingest.sst")
	f, err := i.m.engineFS.Create(sstPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(640115)
		return fmt.Sprintf("error = %s", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(640116)
	}
	__antithesis_instrumentation__.Notify(640111)

	sstWriter := storage.MakeIngestionSSTWriter(ctx, i.m.st, f)
	for _, key := range i.keys {
		__antithesis_instrumentation__.Notify(640117)
		_ = sstWriter.Put(key, []byte("ingested"))
	}
	__antithesis_instrumentation__.Notify(640112)
	if err := sstWriter.Finish(); err != nil {
		__antithesis_instrumentation__.Notify(640118)
		return fmt.Sprintf("error = %s", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(640119)
	}
	__antithesis_instrumentation__.Notify(640113)
	sstWriter.Close()

	if err := i.m.engine.IngestExternalFiles(ctx, []string{sstPath}); err != nil {
		__antithesis_instrumentation__.Notify(640120)
		return fmt.Sprintf("error = %s", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(640121)
	}
	__antithesis_instrumentation__.Notify(640114)

	return "ok"
}

type restartOp struct {
	m *metaTestRunner
}

func (r restartOp) run(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(640122)
	if !r.m.restarts {
		__antithesis_instrumentation__.Notify(640124)
		r.m.printComment("no-op due to restarts being disabled")
		return "ok"
	} else {
		__antithesis_instrumentation__.Notify(640125)
	}
	__antithesis_instrumentation__.Notify(640123)

	oldEngine, newEngine := r.m.restart()
	r.m.printComment(fmt.Sprintf("restarting: %s -> %s", oldEngine.name, newEngine.name))
	r.m.printComment(fmt.Sprintf("new options: %s", newEngine.opts.String()))
	return "ok"
}

var opGenerators = []opGenerator{
	{
		name: "mvcc_inconsistent_get",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640126)
			reader := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1])
			ts := m.pastTSGenerator.parse(args[2])
			return &mvccGetOp{
				m:            m,
				reader:       reader,
				key:          key.Key,
				ts:           ts,
				inconsistent: true,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandPastTS,
		},
		weight: 100,
	},
	{
		name: "mvcc_get",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640127)
			reader := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1])
			txn := txnID(args[2])
			return &mvccGetOp{
				m:      m,
				reader: reader,
				key:    key.Key,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandTransaction,
		},
		weight: 100,
	},
	{
		name: "mvcc_put",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640128)
			writer := readWriterID(args[0])
			txn := txnID(args[1])
			key := m.txnKeyGenerator.parse(args[2])
			value := roachpb.MakeValueFromBytes(m.valueGenerator.parse(args[3]))

			m.txnGenerator.trackTransactionalWrite(writer, txn, key.Key, nil)
			return &mvccPutOp{
				m:      m,
				writer: writer,
				key:    key.Key,
				value:  value,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandTransaction,
			operandUnusedMVCCKey,
			operandValue,
		},
		weight: 500,
	},
	{
		name: "mvcc_conditional_put",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640129)
			writer := readWriterID(args[0])
			txn := txnID(args[1])
			key := m.txnKeyGenerator.parse(args[2])
			value := roachpb.MakeValueFromBytes(m.valueGenerator.parse(args[3]))
			expVal := m.valueGenerator.parse(args[4])

			m.txnGenerator.trackTransactionalWrite(writer, txn, key.Key, nil)
			return &mvccCPutOp{
				m:      m,
				writer: writer,
				key:    key.Key,
				value:  value,
				expVal: expVal,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandTransaction,
			operandUnusedMVCCKey,
			operandValue,
			operandValue,
		},
		weight: 50,
	},
	{
		name: "mvcc_init_put",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640130)
			writer := readWriterID(args[0])
			txn := txnID(args[1])
			key := m.txnKeyGenerator.parse(args[2])
			value := roachpb.MakeValueFromBytes(m.valueGenerator.parse(args[3]))

			m.txnGenerator.trackTransactionalWrite(writer, txn, key.Key, nil)
			return &mvccInitPutOp{
				m:      m,
				writer: writer,
				key:    key.Key,
				value:  value,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandTransaction,
			operandUnusedMVCCKey,
			operandValue,
		},
		weight: 50,
	},
	{
		name: "mvcc_delete_range",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640131)
			writer := readWriterID(args[0])
			txn := txnID(args[1])
			key := m.keyGenerator.parse(args[2]).Key
			endKey := m.keyGenerator.parse(args[3]).Key

			if endKey.Compare(key) < 0 {
				__antithesis_instrumentation__.Notify(640133)
				key, endKey = endKey, key
			} else {
				__antithesis_instrumentation__.Notify(640134)
			}
			__antithesis_instrumentation__.Notify(640132)
			truncatedSpan := m.txnGenerator.truncateSpanForConflicts(writer, txn, key, endKey)
			key = truncatedSpan.Key
			endKey = truncatedSpan.EndKey

			m.txnGenerator.trackTransactionalWrite(writer, txn, key, endKey)
			return &mvccDeleteRangeOp{
				m:      m,
				writer: writer,
				key:    key,
				endKey: endKey,
				txn:    txn,
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) []opReference {
			__antithesis_instrumentation__.Notify(640135)
			return closeItersOnBatch(m, readWriterID(args[0]))
		},
		operands: []operandType{
			operandReadWriter,
			operandTransaction,
			operandUnusedMVCCKey,
			operandUnusedMVCCKey,
		},
		weight: 20,
	},
	{
		name: "mvcc_clear_time_range",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640136)
			key := m.keyGenerator.parse(args[0]).Key
			endKey := m.keyGenerator.parse(args[1]).Key
			startTime := m.pastTSGenerator.parse(args[2])
			endTime := m.pastTSGenerator.parse(args[3])

			if endKey.Compare(key) < 0 {
				__antithesis_instrumentation__.Notify(640139)
				key, endKey = endKey, key
			} else {
				__antithesis_instrumentation__.Notify(640140)
			}
			__antithesis_instrumentation__.Notify(640137)
			truncatedSpan := m.txnGenerator.truncateSpanForConflicts("engine", "", key, endKey)
			key = truncatedSpan.Key
			endKey = truncatedSpan.EndKey
			if endTime.Less(startTime) {
				__antithesis_instrumentation__.Notify(640141)
				startTime, endTime = endTime, startTime
			} else {
				__antithesis_instrumentation__.Notify(640142)
			}
			__antithesis_instrumentation__.Notify(640138)

			endTime = endTime.Next()
			return &mvccClearTimeRangeOp{
				m:         m,
				key:       key,
				endKey:    endKey,
				startTime: startTime,
				endTime:   endTime,
			}
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
			operandPastTS,
			operandPastTS,
		},
		weight: 20,
	},
	{
		name: "mvcc_delete",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640143)
			writer := readWriterID(args[0])
			txn := txnID(args[1])
			key := m.txnKeyGenerator.parse(args[2])

			m.txnGenerator.trackTransactionalWrite(writer, txn, key.Key, nil)
			return &mvccDeleteOp{
				m:      m,
				writer: writer,
				key:    key.Key,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandTransaction,
			operandUnusedMVCCKey,
		},
		weight: 100,
	},
	{
		name: "mvcc_find_split_key",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640144)
			key, _ := keys.Addr(m.keyGenerator.parse(args[0]).Key)
			endKey, _ := keys.Addr(m.keyGenerator.parse(args[1]).Key)
			splitSize := int64(1024)

			return &mvccFindSplitKeyOp{
				m:         m,
				key:       key,
				endKey:    endKey,
				splitSize: splitSize,
			}
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
		},
		weight: 20,
	},
	{
		name: "mvcc_scan",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640145)
			return generateMVCCScan(ctx, m, false, false, args)
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
			operandTransaction,
			operandFloat,
			operandFloat,
			operandBool,
			operandBool,
		},
		weight: 100,
	},
	{
		name: "mvcc_inconsistent_scan",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640146)
			return generateMVCCScan(ctx, m, false, true, args)
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
			operandPastTS,
			operandFloat,
			operandFloat,
			operandBool,
			operandBool,
		},
		weight: 100,
	},
	{
		name: "mvcc_reverse_scan",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640147)
			return generateMVCCScan(ctx, m, true, false, args)
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
			operandTransaction,
			operandFloat,
			operandFloat,
			operandBool,
			operandBool,
		},
		weight: 100,
	},
	{
		name: "txn_open",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640148)
			return &txnOpenOp{
				m:  m,
				id: txnID(args[1]),
				ts: m.nextTSGenerator.parse(args[0]),
			}
		},
		operands: []operandType{
			operandNextTS,
			operandTransaction,
		},
		weight:   40,
		isOpener: true,
	},
	{
		name: "txn_commit",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640149)
			m.txnGenerator.generateClose(txnID(args[0]))
			return &txnCommitOp{
				m:  m,
				id: txnID(args[0]),
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) (result []opReference) {
			__antithesis_instrumentation__.Notify(640150)
			txn := txnID(args[0])

			for batch := range m.txnGenerator.openBatches[txn] {
				__antithesis_instrumentation__.Notify(640152)
				result = append(result, opReference{
					generator: m.nameToGenerator["batch_commit"],
					args:      []string{string(batch)},
				})
			}
			__antithesis_instrumentation__.Notify(640151)
			return
		},
		operands: []operandType{
			operandTransaction,
		},
		weight: 50,
	},
	{
		name: "txn_abort",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640153)
			m.txnGenerator.generateClose(txnID(args[0]))
			return &txnAbortOp{
				m:  m,
				id: txnID(args[0]),
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) (result []opReference) {
			__antithesis_instrumentation__.Notify(640154)
			txn := txnID(args[0])

			for batch := range m.txnGenerator.openBatches[txn] {
				__antithesis_instrumentation__.Notify(640156)
				result = append(result, opReference{
					generator: m.nameToGenerator["batch_commit"],
					args:      []string{string(batch)},
				})
			}
			__antithesis_instrumentation__.Notify(640155)
			return
		},
		operands: []operandType{
			operandTransaction,
		},
		weight: 50,
	},
	{
		name: "txn_create_savepoint",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640157)
			savepoint, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				__antithesis_instrumentation__.Notify(640159)
				panic(err.Error())
			} else {
				__antithesis_instrumentation__.Notify(640160)
			}
			__antithesis_instrumentation__.Notify(640158)
			return &txnCreateSavepointOp{
				m:         m,
				id:        txnID(args[0]),
				savepoint: int(savepoint),
			}
		},
		operands: []operandType{
			operandTransaction,
			operandSavepoint,
		},
		isOpener: true,
		weight:   10,
	},
	{
		name: "txn_rollback_savepoint",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640161)
			savepoint, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				__antithesis_instrumentation__.Notify(640163)
				panic(err.Error())
			} else {
				__antithesis_instrumentation__.Notify(640164)
			}
			__antithesis_instrumentation__.Notify(640162)
			return &txnRollbackSavepointOp{
				m:         m,
				id:        txnID(args[0]),
				savepoint: int(savepoint),
			}
		},
		operands: []operandType{
			operandTransaction,
			operandSavepoint,
		},
		weight: 10,
	},
	{
		name: "batch_open",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640165)
			batchID := readWriterID(args[0])
			return &batchOpenOp{
				m:  m,
				id: batchID,
			}
		},
		operands: []operandType{
			operandReadWriter,
		},
		weight:   40,
		isOpener: true,
	},
	{
		name: "batch_commit",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640166)
			batchID := readWriterID(args[0])
			m.rwGenerator.generateClose(batchID)

			return &batchCommitOp{
				m:  m,
				id: batchID,
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) (results []opReference) {
			__antithesis_instrumentation__.Notify(640167)
			return closeItersOnBatch(m, readWriterID(args[0]))
		},
		operands: []operandType{
			operandReadWriter,
		},
		weight: 100,
	},
	{
		name: "iterator_open",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640168)
			key := m.keyGenerator.parse(args[1])
			endKey := m.keyGenerator.parse(args[2])
			iterID := iteratorID(args[3])
			if endKey.Less(key) {
				__antithesis_instrumentation__.Notify(640170)
				key, endKey = endKey, key
			} else {
				__antithesis_instrumentation__.Notify(640171)
			}
			__antithesis_instrumentation__.Notify(640169)
			rw := readWriterID(args[0])
			m.iterGenerator.generateOpen(rw, iterID)
			return &iterOpenOp{
				m:      m,
				rw:     rw,
				key:    key.Key,
				endKey: endKey.Key,
				id:     iterID,
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) (results []opReference) {
			__antithesis_instrumentation__.Notify(640172)
			return closeItersOnBatch(m, readWriterID(args[0]))
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandMVCCKey,
			operandIterator,
		},
		weight:   20,
		isOpener: true,
	},
	{
		name: "iterator_close",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640173)
			iter := iteratorID(args[0])
			m.iterGenerator.generateClose(iter)
			return &iterCloseOp{
				m:  m,
				id: iter,
			}
		},
		operands: []operandType{
			operandIterator,
		},
		weight: 50,
	},
	{
		name: "iterator_seekge",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640174)
			iter := iteratorID(args[0])
			key := m.keyGenerator.parse(args[1])
			return &iterSeekOp{
				m:      m,
				iter:   iter,
				key:    key,
				seekLT: false,
			}
		},
		operands: []operandType{
			operandIterator,
			operandMVCCKey,
		},
		weight: 50,
	},
	{
		name: "iterator_seeklt",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640175)
			iter := iteratorID(args[0])
			key := m.keyGenerator.parse(args[1])

			return &iterSeekOp{
				m:      m,
				iter:   iter,
				key:    key,
				seekLT: true,
			}
		},
		operands: []operandType{
			operandIterator,
			operandMVCCKey,
		},
		weight: 50,
	},
	{
		name: "iterator_next",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640176)
			iter := iteratorID(args[0])

			return &iterNextOp{
				m:       m,
				iter:    iter,
				nextKey: false,
			}
		},
		operands: []operandType{
			operandIterator,
		},
		weight: 100,
	},
	{
		name: "iterator_nextkey",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640177)
			iter := iteratorID(args[0])
			return &iterNextOp{
				m:       m,
				iter:    iter,
				nextKey: false,
			}
		},
		operands: []operandType{
			operandIterator,
		},
		weight: 100,
	},
	{
		name: "iterator_prev",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640178)
			iter := iteratorID(args[0])
			return &iterPrevOp{
				m:    m,
				iter: iter,
			}
		},
		operands: []operandType{
			operandIterator,
		},
		weight: 100,
	},
	{

		name: "delete_range",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640179)
			key := m.keyGenerator.parse(args[0]).Key
			endKey := m.keyGenerator.parse(args[1]).Key
			if endKey.Compare(key) < 0 {
				__antithesis_instrumentation__.Notify(640181)
				key, endKey = endKey, key
			} else {
				__antithesis_instrumentation__.Notify(640182)
				if endKey.Equal(key) {
					__antithesis_instrumentation__.Notify(640183)

					endKey = endKey.Next()
				} else {
					__antithesis_instrumentation__.Notify(640184)
				}
			}
			__antithesis_instrumentation__.Notify(640180)

			truncatedSpan := m.txnGenerator.truncateSpanForConflicts("engine", "", key, endKey)
			key = truncatedSpan.Key
			endKey = truncatedSpan.EndKey

			return &clearRangeOp{
				m:      m,
				key:    key,
				endKey: endKey,
			}
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
		},
		weight: 20,
	},
	{
		name: "compact",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640185)
			key := m.keyGenerator.parse(args[0]).Key
			endKey := m.keyGenerator.parse(args[1]).Key
			if endKey.Compare(key) < 0 {
				__antithesis_instrumentation__.Notify(640187)
				key, endKey = endKey, key
			} else {
				__antithesis_instrumentation__.Notify(640188)
			}
			__antithesis_instrumentation__.Notify(640186)
			return &compactOp{
				m:      m,
				key:    key,
				endKey: endKey,
			}
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
		},
		weight: 10,
	},
	{
		name: "ingest",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640189)
			var keys []storage.MVCCKey
			for _, arg := range args {
				__antithesis_instrumentation__.Notify(640192)
				key := m.keyGenerator.parse(arg)

				if key.Timestamp.IsEmpty() {
					__antithesis_instrumentation__.Notify(640194)
					key.Timestamp = key.Timestamp.Next()
				} else {
					__antithesis_instrumentation__.Notify(640195)
				}
				__antithesis_instrumentation__.Notify(640193)
				keys = append(keys, key)
			}
			__antithesis_instrumentation__.Notify(640190)

			sort.Slice(keys, func(i, j int) bool {
				__antithesis_instrumentation__.Notify(640196)
				return keys[i].Less(keys[j])
			})
			__antithesis_instrumentation__.Notify(640191)

			return &ingestOp{
				m:    m,
				keys: keys,
			}
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
			operandMVCCKey,
			operandMVCCKey,
			operandMVCCKey,
		},
		weight: 10,
	},
	{
		name: "restart",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			__antithesis_instrumentation__.Notify(640197)

			m.closeGenerators()
			return &restartOp{
				m: m,
			}
		},
		operands: nil,
		weight:   4,
	},
}
