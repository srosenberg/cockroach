package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/binary"
	"sort"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/uncertainty"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

const (
	maxItersBeforeSeek = 10

	kvLenSize = 8
)

type pebbleResults struct {
	count int64
	bytes int64
	repr  []byte
	bufs  [][]byte

	lastOffsetsEnabled bool
	lastOffsets        []int
	lastOffsetIdx      int
}

func (p *pebbleResults) clear() {
	__antithesis_instrumentation__.Notify(643243)
	*p = pebbleResults{}
}

func (p *pebbleResults) put(
	ctx context.Context, key []byte, value []byte, memAccount *mon.BoundAccount,
) error {
	__antithesis_instrumentation__.Notify(643244)
	const minSize = 16
	const maxSize = 128 << 20

	lenKey := len(key)
	lenValue := len(value)
	lenToAdd := p.sizeOf(lenKey, lenValue)
	if len(p.repr)+lenToAdd > cap(p.repr) {
		__antithesis_instrumentation__.Notify(643247)
		newSize := 2 * cap(p.repr)
		if newSize == 0 || func() bool {
			__antithesis_instrumentation__.Notify(643252)
			return newSize > maxSize == true
		}() == true {
			__antithesis_instrumentation__.Notify(643253)

			newSize = minSize
		} else {
			__antithesis_instrumentation__.Notify(643254)
		}
		__antithesis_instrumentation__.Notify(643248)
		if lenToAdd >= maxSize {
			__antithesis_instrumentation__.Notify(643255)
			newSize = lenToAdd
		} else {
			__antithesis_instrumentation__.Notify(643256)
			for newSize < lenToAdd && func() bool {
				__antithesis_instrumentation__.Notify(643257)
				return newSize < maxSize == true
			}() == true {
				__antithesis_instrumentation__.Notify(643258)
				newSize *= 2
			}
		}
		__antithesis_instrumentation__.Notify(643249)
		if len(p.repr) > 0 {
			__antithesis_instrumentation__.Notify(643259)
			p.bufs = append(p.bufs, p.repr)
		} else {
			__antithesis_instrumentation__.Notify(643260)
		}
		__antithesis_instrumentation__.Notify(643250)
		if err := memAccount.Grow(ctx, int64(newSize)); err != nil {
			__antithesis_instrumentation__.Notify(643261)
			return err
		} else {
			__antithesis_instrumentation__.Notify(643262)
		}
		__antithesis_instrumentation__.Notify(643251)
		p.repr = nonZeroingMakeByteSlice(newSize)[:0]
	} else {
		__antithesis_instrumentation__.Notify(643263)
	}
	__antithesis_instrumentation__.Notify(643245)

	startIdx := len(p.repr)
	p.repr = p.repr[:startIdx+lenToAdd]
	binary.LittleEndian.PutUint32(p.repr[startIdx:], uint32(lenValue))
	binary.LittleEndian.PutUint32(p.repr[startIdx+4:], uint32(lenKey))
	copy(p.repr[startIdx+kvLenSize:], key)
	copy(p.repr[startIdx+kvLenSize+lenKey:], value)
	p.count++
	p.bytes += int64(lenToAdd)

	if p.lastOffsetsEnabled {
		__antithesis_instrumentation__.Notify(643264)
		p.lastOffsets[p.lastOffsetIdx] = startIdx
		p.lastOffsetIdx++

		if p.lastOffsetIdx == len(p.lastOffsets) {
			__antithesis_instrumentation__.Notify(643265)
			p.lastOffsetIdx = 0
		} else {
			__antithesis_instrumentation__.Notify(643266)
		}
	} else {
		__antithesis_instrumentation__.Notify(643267)
	}
	__antithesis_instrumentation__.Notify(643246)

	return nil
}

func (p *pebbleResults) sizeOf(lenKey, lenValue int) int {
	__antithesis_instrumentation__.Notify(643268)
	return kvLenSize + lenKey + lenValue
}

func (p *pebbleResults) continuesFirstRow(key roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(643269)
	repr := p.repr
	if len(p.bufs) > 0 {
		__antithesis_instrumentation__.Notify(643273)
		repr = p.bufs[0]
	} else {
		__antithesis_instrumentation__.Notify(643274)
	}
	__antithesis_instrumentation__.Notify(643270)
	if len(repr) == 0 {
		__antithesis_instrumentation__.Notify(643275)
		return true
	} else {
		__antithesis_instrumentation__.Notify(643276)
	}
	__antithesis_instrumentation__.Notify(643271)

	rowPrefix := getRowPrefix(key)
	if rowPrefix == nil {
		__antithesis_instrumentation__.Notify(643277)
		return false
	} else {
		__antithesis_instrumentation__.Notify(643278)
	}
	__antithesis_instrumentation__.Notify(643272)
	return bytes.Equal(rowPrefix, getRowPrefix(extractResultKey(repr)))
}

func (p *pebbleResults) lastRowHasFinalColumnFamily() bool {
	__antithesis_instrumentation__.Notify(643279)
	if !p.lastOffsetsEnabled || func() bool {
		__antithesis_instrumentation__.Notify(643283)
		return p.count == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(643284)
		return false
	} else {
		__antithesis_instrumentation__.Notify(643285)
	}
	__antithesis_instrumentation__.Notify(643280)

	lastOffsetIdx := p.lastOffsetIdx - 1
	if lastOffsetIdx < 0 {
		__antithesis_instrumentation__.Notify(643286)
		lastOffsetIdx = len(p.lastOffsets) - 1
	} else {
		__antithesis_instrumentation__.Notify(643287)
	}
	__antithesis_instrumentation__.Notify(643281)
	lastOffset := p.lastOffsets[lastOffsetIdx]

	key := extractResultKey(p.repr[lastOffset:])
	colFamilyID, err := keys.DecodeFamilyKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(643288)
		return false
	} else {
		__antithesis_instrumentation__.Notify(643289)
	}
	__antithesis_instrumentation__.Notify(643282)
	return int(colFamilyID) == len(p.lastOffsets)-1
}

func (p *pebbleResults) maybeTrimPartialLastRow(nextKey roachpb.Key) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(643290)
	if !p.lastOffsetsEnabled || func() bool {
		__antithesis_instrumentation__.Notify(643294)
		return len(p.repr) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(643295)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(643296)
	}
	__antithesis_instrumentation__.Notify(643291)
	trimRowPrefix := getRowPrefix(nextKey)
	if trimRowPrefix == nil {
		__antithesis_instrumentation__.Notify(643297)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(643298)
	}
	__antithesis_instrumentation__.Notify(643292)

	var firstTrimmedKey roachpb.Key

	for i := 0; i < len(p.lastOffsets); i++ {
		__antithesis_instrumentation__.Notify(643299)
		lastOffsetIdx := p.lastOffsetIdx - 1
		if lastOffsetIdx < 0 {
			__antithesis_instrumentation__.Notify(643302)
			lastOffsetIdx = len(p.lastOffsets) - 1
		} else {
			__antithesis_instrumentation__.Notify(643303)
		}
		__antithesis_instrumentation__.Notify(643300)
		lastOffset := p.lastOffsets[lastOffsetIdx]

		repr := p.repr[lastOffset:]
		key := extractResultKey(repr)
		rowPrefix := getRowPrefix(key)

		if !bytes.Equal(rowPrefix, trimRowPrefix) {
			__antithesis_instrumentation__.Notify(643304)
			return firstTrimmedKey, nil
		} else {
			__antithesis_instrumentation__.Notify(643305)
		}
		__antithesis_instrumentation__.Notify(643301)

		p.repr = p.repr[:lastOffset]
		p.count--
		p.bytes -= int64(len(repr))
		firstTrimmedKey = key

		p.lastOffsetIdx = lastOffsetIdx
		p.lastOffsets[lastOffsetIdx] = 0

		if len(p.repr) == 0 {
			__antithesis_instrumentation__.Notify(643306)
			if len(p.bufs) == 0 {
				__antithesis_instrumentation__.Notify(643308)

				return firstTrimmedKey, nil
			} else {
				__antithesis_instrumentation__.Notify(643309)
			}
			__antithesis_instrumentation__.Notify(643307)

			p.repr = p.bufs[len(p.bufs)-1]
			p.bufs = p.bufs[:len(p.bufs)-1]
		} else {
			__antithesis_instrumentation__.Notify(643310)
		}
	}
	__antithesis_instrumentation__.Notify(643293)
	return nil, errors.Errorf("row exceeds expected max size (%d): %s", len(p.lastOffsets), nextKey)
}

func (p *pebbleResults) finish() [][]byte {
	__antithesis_instrumentation__.Notify(643311)
	if len(p.repr) > 0 {
		__antithesis_instrumentation__.Notify(643313)
		p.bufs = append(p.bufs, p.repr)
		p.repr = nil
	} else {
		__antithesis_instrumentation__.Notify(643314)
	}
	__antithesis_instrumentation__.Notify(643312)
	return p.bufs
}

func getRowPrefix(key roachpb.Key) []byte {
	__antithesis_instrumentation__.Notify(643315)
	if len(key) == 0 {
		__antithesis_instrumentation__.Notify(643318)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(643319)
	}
	__antithesis_instrumentation__.Notify(643316)
	n, err := keys.GetRowPrefixLength(key)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(643320)
		return n <= 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(643321)
		return n >= len(key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(643322)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(643323)
	}
	__antithesis_instrumentation__.Notify(643317)
	return key[:n]
}

func extractResultKey(repr []byte) roachpb.Key {
	__antithesis_instrumentation__.Notify(643324)
	keyLen := binary.LittleEndian.Uint32(repr[4:8])
	key, _, ok := enginepb.SplitMVCCKey(repr[8 : 8+keyLen])
	if !ok {
		__antithesis_instrumentation__.Notify(643326)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(643327)
	}
	__antithesis_instrumentation__.Notify(643325)
	return key
}

type pebbleMVCCScanner struct {
	parent MVCCIterator

	memAccount *mon.BoundAccount
	reverse    bool
	peeked     bool

	start, end roachpb.Key

	ts hlc.Timestamp

	maxKeys int64

	targetBytes int64

	targetBytesAvoidExcess bool

	allowEmpty bool

	wholeRows bool

	maxIntents int64

	resumeReason    roachpb.ResumeReason
	resumeKey       roachpb.Key
	resumeNextBytes int64

	txn               *roachpb.Transaction
	txnEpoch          enginepb.TxnEpoch
	txnSequence       enginepb.TxnSeq
	txnIgnoredSeqNums []enginepb.IgnoredSeqNumRange

	uncertainty      uncertainty.Interval
	checkUncertainty bool

	meta enginepb.MVCCMetadata

	inconsistent, tombstones bool
	failOnMoreRecent         bool
	isGet                    bool
	keyBuf                   []byte
	savedBuf                 []byte

	curUnsafeKey MVCCKey
	curRawKey    []byte
	curValue     []byte
	results      pebbleResults
	intents      pebble.Batch

	mostRecentTS  hlc.Timestamp
	mostRecentKey roachpb.Key

	err error

	itersBeforeSeek int
}

var pebbleMVCCScannerPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(643328)
		return &pebbleMVCCScanner{}
	},
}

func (p *pebbleMVCCScanner) release() {
	__antithesis_instrumentation__.Notify(643329)

	*p = pebbleMVCCScanner{
		keyBuf: p.keyBuf,
	}
	pebbleMVCCScannerPool.Put(p)
}

func (p *pebbleMVCCScanner) init(
	txn *roachpb.Transaction, ui uncertainty.Interval, trackLastOffsets int,
) {
	p.itersBeforeSeek = maxItersBeforeSeek / 2
	if trackLastOffsets > 0 {
		p.results.lastOffsetsEnabled = true
		p.results.lastOffsets = make([]int, trackLastOffsets)
	}

	if txn != nil {
		p.txn = txn
		p.txnEpoch = txn.Epoch
		p.txnSequence = txn.Sequence
		p.txnIgnoredSeqNums = txn.IgnoredSeqNums
	}

	p.uncertainty = ui

	p.checkUncertainty = p.ts.Less(p.uncertainty.GlobalLimit)
}

func (p *pebbleMVCCScanner) get(ctx context.Context) {
	__antithesis_instrumentation__.Notify(643330)
	p.isGet = true
	p.parent.SeekGE(MVCCKey{Key: p.start})
	if !p.updateCurrent() {
		__antithesis_instrumentation__.Notify(643332)
		return
	} else {
		__antithesis_instrumentation__.Notify(643333)
	}
	__antithesis_instrumentation__.Notify(643331)
	p.getAndAdvance(ctx)
	p.maybeFailOnMoreRecent()
}

func (p *pebbleMVCCScanner) scan(
	ctx context.Context,
) (*roachpb.Span, roachpb.ResumeReason, int64, error) {
	__antithesis_instrumentation__.Notify(643334)
	if p.wholeRows && func() bool {
		__antithesis_instrumentation__.Notify(643340)
		return !p.results.lastOffsetsEnabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(643341)
		return nil, 0, 0, errors.AssertionFailedf("cannot use wholeRows without trackLastOffsets")
	} else {
		__antithesis_instrumentation__.Notify(643342)
	}
	__antithesis_instrumentation__.Notify(643335)

	p.isGet = false
	if p.reverse {
		__antithesis_instrumentation__.Notify(643343)
		if !p.iterSeekReverse(MVCCKey{Key: p.end}) {
			__antithesis_instrumentation__.Notify(643344)
			return nil, 0, 0, p.err
		} else {
			__antithesis_instrumentation__.Notify(643345)
		}
	} else {
		__antithesis_instrumentation__.Notify(643346)
		if !p.iterSeek(MVCCKey{Key: p.start}) {
			__antithesis_instrumentation__.Notify(643347)
			return nil, 0, 0, p.err
		} else {
			__antithesis_instrumentation__.Notify(643348)
		}
	}
	__antithesis_instrumentation__.Notify(643336)

	for p.getAndAdvance(ctx) {
		__antithesis_instrumentation__.Notify(643349)
	}
	__antithesis_instrumentation__.Notify(643337)
	p.maybeFailOnMoreRecent()

	if p.err != nil {
		__antithesis_instrumentation__.Notify(643350)
		return nil, 0, 0, p.err
	} else {
		__antithesis_instrumentation__.Notify(643351)
	}
	__antithesis_instrumentation__.Notify(643338)

	if p.resumeReason != 0 {
		__antithesis_instrumentation__.Notify(643352)
		resumeKey := p.resumeKey
		if len(resumeKey) == 0 {
			__antithesis_instrumentation__.Notify(643355)
			if !p.advanceKey() {
				__antithesis_instrumentation__.Notify(643357)
				return nil, 0, 0, nil
			} else {
				__antithesis_instrumentation__.Notify(643358)
			}
			__antithesis_instrumentation__.Notify(643356)
			resumeKey = p.curUnsafeKey.Key
		} else {
			__antithesis_instrumentation__.Notify(643359)
		}
		__antithesis_instrumentation__.Notify(643353)

		var resumeSpan *roachpb.Span
		if p.reverse {
			__antithesis_instrumentation__.Notify(643360)

			resumeKeyCopy := make(roachpb.Key, len(resumeKey), len(resumeKey)+1)
			copy(resumeKeyCopy, resumeKey)
			resumeSpan = &roachpb.Span{
				Key:    p.start,
				EndKey: resumeKeyCopy.Next(),
			}
		} else {
			__antithesis_instrumentation__.Notify(643361)
			resumeSpan = &roachpb.Span{
				Key:    append(roachpb.Key(nil), resumeKey...),
				EndKey: p.end,
			}
		}
		__antithesis_instrumentation__.Notify(643354)
		return resumeSpan, p.resumeReason, p.resumeNextBytes, nil
	} else {
		__antithesis_instrumentation__.Notify(643362)
	}
	__antithesis_instrumentation__.Notify(643339)
	return nil, 0, 0, nil
}

func (p *pebbleMVCCScanner) incrementItersBeforeSeek() {
	__antithesis_instrumentation__.Notify(643363)
	p.itersBeforeSeek++
	if p.itersBeforeSeek > maxItersBeforeSeek {
		__antithesis_instrumentation__.Notify(643364)
		p.itersBeforeSeek = maxItersBeforeSeek
	} else {
		__antithesis_instrumentation__.Notify(643365)
	}
}

func (p *pebbleMVCCScanner) decrementItersBeforeSeek() {
	__antithesis_instrumentation__.Notify(643366)
	p.itersBeforeSeek--
	if p.itersBeforeSeek < 1 {
		__antithesis_instrumentation__.Notify(643367)
		p.itersBeforeSeek = 1
	} else {
		__antithesis_instrumentation__.Notify(643368)
	}
}

func (p *pebbleMVCCScanner) getFromIntentHistory() (value []byte, found bool) {
	__antithesis_instrumentation__.Notify(643369)
	intentHistory := p.meta.IntentHistory

	upIdx := sort.Search(len(intentHistory), func(i int) bool {
		__antithesis_instrumentation__.Notify(643373)
		return intentHistory[i].Sequence > p.txnSequence
	})
	__antithesis_instrumentation__.Notify(643370)

	for upIdx > 0 && func() bool {
		__antithesis_instrumentation__.Notify(643374)
		return enginepb.TxnSeqIsIgnored(p.meta.IntentHistory[upIdx-1].Sequence, p.txnIgnoredSeqNums) == true
	}() == true {
		__antithesis_instrumentation__.Notify(643375)
		upIdx--
	}
	__antithesis_instrumentation__.Notify(643371)
	if upIdx == 0 {
		__antithesis_instrumentation__.Notify(643376)

		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(643377)
	}
	__antithesis_instrumentation__.Notify(643372)
	intent := &p.meta.IntentHistory[upIdx-1]
	return intent.Value, true
}

func (p *pebbleMVCCScanner) maybeFailOnMoreRecent() {
	__antithesis_instrumentation__.Notify(643378)
	if p.err != nil || func() bool {
		__antithesis_instrumentation__.Notify(643380)
		return p.mostRecentTS.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(643381)
		return
	} else {
		__antithesis_instrumentation__.Notify(643382)
	}
	__antithesis_instrumentation__.Notify(643379)

	p.err = roachpb.NewWriteTooOldError(p.ts, p.mostRecentTS.Next(), p.mostRecentKey)
	p.results.clear()
	p.intents.Reset()
}

func (p *pebbleMVCCScanner) uncertaintyError(ts hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(643383)
	p.err = roachpb.NewReadWithinUncertaintyIntervalError(
		p.ts, ts, p.uncertainty.LocalLimit.ToTimestamp(), p.txn)
	p.results.clear()
	p.intents.Reset()
	return false
}

func (p *pebbleMVCCScanner) getAndAdvance(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(643384)
	if !p.curUnsafeKey.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(643396)

		if p.curUnsafeKey.Timestamp.Less(p.ts) {
			__antithesis_instrumentation__.Notify(643401)

			return p.addAndAdvance(ctx, p.curUnsafeKey.Key, p.curRawKey, p.curValue)
		} else {
			__antithesis_instrumentation__.Notify(643402)
		}
		__antithesis_instrumentation__.Notify(643397)

		if p.curUnsafeKey.Timestamp.EqOrdering(p.ts) {
			__antithesis_instrumentation__.Notify(643403)
			if p.failOnMoreRecent {
				__antithesis_instrumentation__.Notify(643405)

				p.mostRecentTS.Forward(p.curUnsafeKey.Timestamp)
				if len(p.mostRecentKey) == 0 {
					__antithesis_instrumentation__.Notify(643407)
					p.mostRecentKey = append(p.mostRecentKey, p.curUnsafeKey.Key...)
				} else {
					__antithesis_instrumentation__.Notify(643408)
				}
				__antithesis_instrumentation__.Notify(643406)
				return p.advanceKey()
			} else {
				__antithesis_instrumentation__.Notify(643409)
			}
			__antithesis_instrumentation__.Notify(643404)

			return p.addAndAdvance(ctx, p.curUnsafeKey.Key, p.curRawKey, p.curValue)
		} else {
			__antithesis_instrumentation__.Notify(643410)
		}
		__antithesis_instrumentation__.Notify(643398)

		if p.failOnMoreRecent {
			__antithesis_instrumentation__.Notify(643411)

			p.mostRecentTS.Forward(p.curUnsafeKey.Timestamp)
			if len(p.mostRecentKey) == 0 {
				__antithesis_instrumentation__.Notify(643413)
				p.mostRecentKey = append(p.mostRecentKey, p.curUnsafeKey.Key...)
			} else {
				__antithesis_instrumentation__.Notify(643414)
			}
			__antithesis_instrumentation__.Notify(643412)
			return p.advanceKey()
		} else {
			__antithesis_instrumentation__.Notify(643415)
		}
		__antithesis_instrumentation__.Notify(643399)

		if p.checkUncertainty {
			__antithesis_instrumentation__.Notify(643416)

			if p.uncertainty.IsUncertain(p.curUnsafeKey.Timestamp) {
				__antithesis_instrumentation__.Notify(643418)
				return p.uncertaintyError(p.curUnsafeKey.Timestamp)
			} else {
				__antithesis_instrumentation__.Notify(643419)
			}
			__antithesis_instrumentation__.Notify(643417)

			return p.seekVersion(ctx, p.uncertainty.GlobalLimit, true)
		} else {
			__antithesis_instrumentation__.Notify(643420)
		}
		__antithesis_instrumentation__.Notify(643400)

		return p.seekVersion(ctx, p.ts, false)
	} else {
		__antithesis_instrumentation__.Notify(643421)
	}
	__antithesis_instrumentation__.Notify(643385)

	if len(p.curValue) == 0 {
		__antithesis_instrumentation__.Notify(643422)
		p.err = errors.Errorf("zero-length mvcc metadata")
		return false
	} else {
		__antithesis_instrumentation__.Notify(643423)
	}
	__antithesis_instrumentation__.Notify(643386)
	err := protoutil.Unmarshal(p.curValue, &p.meta)
	if err != nil {
		__antithesis_instrumentation__.Notify(643424)
		p.err = errors.Wrap(err, "unable to decode MVCCMetadata")
		return false
	} else {
		__antithesis_instrumentation__.Notify(643425)
	}
	__antithesis_instrumentation__.Notify(643387)
	if len(p.meta.RawBytes) != 0 {
		__antithesis_instrumentation__.Notify(643426)

		return p.addAndAdvance(ctx, p.curUnsafeKey.Key, p.curRawKey, p.meta.RawBytes)
	} else {
		__antithesis_instrumentation__.Notify(643427)
	}
	__antithesis_instrumentation__.Notify(643388)

	if p.meta.Txn == nil {
		__antithesis_instrumentation__.Notify(643428)
		p.err = errors.Errorf("intent without transaction")
		return false
	} else {
		__antithesis_instrumentation__.Notify(643429)
	}
	__antithesis_instrumentation__.Notify(643389)
	metaTS := p.meta.Timestamp.ToTimestamp()

	prevTS := p.ts
	if metaTS.LessEq(p.ts) {
		__antithesis_instrumentation__.Notify(643430)
		prevTS = metaTS.Prev()
	} else {
		__antithesis_instrumentation__.Notify(643431)
	}
	__antithesis_instrumentation__.Notify(643390)

	ownIntent := p.txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(643432)
		return p.meta.Txn.ID.Equal(p.txn.ID) == true
	}() == true
	conflictingIntent := metaTS.LessEq(p.ts) || func() bool {
		__antithesis_instrumentation__.Notify(643433)
		return p.failOnMoreRecent == true
	}() == true

	if !ownIntent && func() bool {
		__antithesis_instrumentation__.Notify(643434)
		return !conflictingIntent == true
	}() == true {
		__antithesis_instrumentation__.Notify(643435)

		if p.checkUncertainty {
			__antithesis_instrumentation__.Notify(643437)
			if p.uncertainty.IsUncertain(metaTS) {
				__antithesis_instrumentation__.Notify(643439)
				return p.uncertaintyError(metaTS)
			} else {
				__antithesis_instrumentation__.Notify(643440)
			}
			__antithesis_instrumentation__.Notify(643438)

			return p.seekVersion(ctx, p.uncertainty.GlobalLimit, true)
		} else {
			__antithesis_instrumentation__.Notify(643441)
		}
		__antithesis_instrumentation__.Notify(643436)
		return p.seekVersion(ctx, p.ts, false)
	} else {
		__antithesis_instrumentation__.Notify(643442)
	}
	__antithesis_instrumentation__.Notify(643391)

	if p.inconsistent {
		__antithesis_instrumentation__.Notify(643443)

		if p.err = p.memAccount.Grow(ctx, int64(len(p.curRawKey)+len(p.curValue))); p.err != nil {
			__antithesis_instrumentation__.Notify(643446)
			p.err = errors.Wrapf(p.err, "scan with start key %s", p.start)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643447)
		}
		__antithesis_instrumentation__.Notify(643444)
		p.err = p.intents.Set(p.curRawKey, p.curValue, nil)
		if p.err != nil {
			__antithesis_instrumentation__.Notify(643448)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643449)
		}
		__antithesis_instrumentation__.Notify(643445)

		return p.seekVersion(ctx, prevTS, false)
	} else {
		__antithesis_instrumentation__.Notify(643450)
	}
	__antithesis_instrumentation__.Notify(643392)

	if !ownIntent {
		__antithesis_instrumentation__.Notify(643451)

		if p.err = p.memAccount.Grow(ctx, int64(len(p.curRawKey)+len(p.curValue))); p.err != nil {
			__antithesis_instrumentation__.Notify(643455)
			p.err = errors.Wrapf(p.err, "scan with start key %s", p.start)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643456)
		}
		__antithesis_instrumentation__.Notify(643452)
		p.err = p.intents.Set(p.curRawKey, p.curValue, nil)
		if p.err != nil {
			__antithesis_instrumentation__.Notify(643457)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643458)
		}
		__antithesis_instrumentation__.Notify(643453)

		if p.maxIntents > 0 && func() bool {
			__antithesis_instrumentation__.Notify(643459)
			return int64(p.intents.Count()) >= p.maxIntents == true
		}() == true {
			__antithesis_instrumentation__.Notify(643460)
			p.resumeReason = roachpb.RESUME_INTENT_LIMIT
			return false
		} else {
			__antithesis_instrumentation__.Notify(643461)
		}
		__antithesis_instrumentation__.Notify(643454)
		return p.advanceKey()
	} else {
		__antithesis_instrumentation__.Notify(643462)
	}
	__antithesis_instrumentation__.Notify(643393)

	if p.txnEpoch == p.meta.Txn.Epoch {
		__antithesis_instrumentation__.Notify(643463)
		if p.txnSequence >= p.meta.Txn.Sequence && func() bool {
			__antithesis_instrumentation__.Notify(643466)
			return !enginepb.TxnSeqIsIgnored(p.meta.Txn.Sequence, p.txnIgnoredSeqNums) == true
		}() == true {
			__antithesis_instrumentation__.Notify(643467)

			return p.seekVersion(ctx, metaTS, false)
		} else {
			__antithesis_instrumentation__.Notify(643468)
		}
		__antithesis_instrumentation__.Notify(643464)

		if value, found := p.getFromIntentHistory(); found {
			__antithesis_instrumentation__.Notify(643469)

			p.curUnsafeKey.Timestamp = metaTS
			p.keyBuf = EncodeMVCCKeyToBuf(p.keyBuf[:0], p.curUnsafeKey)
			return p.addAndAdvance(ctx, p.curUnsafeKey.Key, p.keyBuf, value)
		} else {
			__antithesis_instrumentation__.Notify(643470)
		}
		__antithesis_instrumentation__.Notify(643465)

		return p.seekVersion(ctx, prevTS, false)
	} else {
		__antithesis_instrumentation__.Notify(643471)
	}
	__antithesis_instrumentation__.Notify(643394)

	if p.txnEpoch < p.meta.Txn.Epoch {
		__antithesis_instrumentation__.Notify(643472)

		p.err = errors.Errorf("failed to read with epoch %d due to a write intent with epoch %d",
			p.txnEpoch, p.meta.Txn.Epoch)
		return false
	} else {
		__antithesis_instrumentation__.Notify(643473)
	}
	__antithesis_instrumentation__.Notify(643395)

	return p.seekVersion(ctx, prevTS, false)
}

func (p *pebbleMVCCScanner) nextKey() bool {
	__antithesis_instrumentation__.Notify(643474)
	p.keyBuf = append(p.keyBuf[:0], p.curUnsafeKey.Key...)

	for i := 0; i < p.itersBeforeSeek; i++ {
		__antithesis_instrumentation__.Notify(643476)
		if !p.iterNext() {
			__antithesis_instrumentation__.Notify(643478)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643479)
		}
		__antithesis_instrumentation__.Notify(643477)
		if !bytes.Equal(p.curUnsafeKey.Key, p.keyBuf) {
			__antithesis_instrumentation__.Notify(643480)
			p.incrementItersBeforeSeek()
			return true
		} else {
			__antithesis_instrumentation__.Notify(643481)
		}
	}
	__antithesis_instrumentation__.Notify(643475)

	p.decrementItersBeforeSeek()

	p.keyBuf = append(p.keyBuf, 0)
	return p.iterSeek(MVCCKey{Key: p.keyBuf})
}

func (p *pebbleMVCCScanner) backwardLatestVersion(key []byte, i int) bool {
	__antithesis_instrumentation__.Notify(643482)
	p.keyBuf = append(p.keyBuf[:0], key...)

	for ; i < p.itersBeforeSeek; i++ {
		__antithesis_instrumentation__.Notify(643484)
		peekedKey, ok := p.iterPeekPrev()
		if !ok {
			__antithesis_instrumentation__.Notify(643487)

			return true
		} else {
			__antithesis_instrumentation__.Notify(643488)
		}
		__antithesis_instrumentation__.Notify(643485)
		if !bytes.Equal(peekedKey, p.keyBuf) {
			__antithesis_instrumentation__.Notify(643489)
			p.incrementItersBeforeSeek()
			return true
		} else {
			__antithesis_instrumentation__.Notify(643490)
		}
		__antithesis_instrumentation__.Notify(643486)
		if !p.iterPrev() {
			__antithesis_instrumentation__.Notify(643491)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643492)
		}
	}
	__antithesis_instrumentation__.Notify(643483)

	p.decrementItersBeforeSeek()
	return p.iterSeek(MVCCKey{Key: p.keyBuf})
}

func (p *pebbleMVCCScanner) prevKey(key []byte) bool {
	__antithesis_instrumentation__.Notify(643493)
	p.keyBuf = append(p.keyBuf[:0], key...)

	for i := 0; i < p.itersBeforeSeek; i++ {
		__antithesis_instrumentation__.Notify(643495)
		peekedKey, ok := p.iterPeekPrev()
		if !ok {
			__antithesis_instrumentation__.Notify(643498)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643499)
		}
		__antithesis_instrumentation__.Notify(643496)
		if !bytes.Equal(peekedKey, p.keyBuf) {
			__antithesis_instrumentation__.Notify(643500)
			return p.backwardLatestVersion(peekedKey, i+1)
		} else {
			__antithesis_instrumentation__.Notify(643501)
		}
		__antithesis_instrumentation__.Notify(643497)
		if !p.iterPrev() {
			__antithesis_instrumentation__.Notify(643502)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643503)
		}
	}
	__antithesis_instrumentation__.Notify(643494)

	p.decrementItersBeforeSeek()
	return p.iterSeekReverse(MVCCKey{Key: p.keyBuf})
}

func (p *pebbleMVCCScanner) advanceKey() bool {
	__antithesis_instrumentation__.Notify(643504)
	if p.isGet {
		__antithesis_instrumentation__.Notify(643507)
		return false
	} else {
		__antithesis_instrumentation__.Notify(643508)
	}
	__antithesis_instrumentation__.Notify(643505)
	if p.reverse {
		__antithesis_instrumentation__.Notify(643509)
		return p.prevKey(p.curUnsafeKey.Key)
	} else {
		__antithesis_instrumentation__.Notify(643510)
	}
	__antithesis_instrumentation__.Notify(643506)
	return p.nextKey()
}

func (p *pebbleMVCCScanner) advanceKeyAtEnd() bool {
	__antithesis_instrumentation__.Notify(643511)
	if p.reverse {
		__antithesis_instrumentation__.Notify(643513)

		p.peeked = false
		p.parent.SeekLT(MVCCKey{Key: p.end})
		if !p.updateCurrent() {
			__antithesis_instrumentation__.Notify(643515)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643516)
		}
		__antithesis_instrumentation__.Notify(643514)
		return p.advanceKey()
	} else {
		__antithesis_instrumentation__.Notify(643517)
	}
	__antithesis_instrumentation__.Notify(643512)

	return false
}

func (p *pebbleMVCCScanner) advanceKeyAtNewKey(key []byte) bool {
	__antithesis_instrumentation__.Notify(643518)
	if p.reverse {
		__antithesis_instrumentation__.Notify(643520)

		return p.prevKey(key)
	} else {
		__antithesis_instrumentation__.Notify(643521)
	}
	__antithesis_instrumentation__.Notify(643519)

	return true
}

func (p *pebbleMVCCScanner) addAndAdvance(
	ctx context.Context, key roachpb.Key, rawKey []byte, val []byte,
) bool {
	__antithesis_instrumentation__.Notify(643522)

	if len(val) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(643528)
		return !p.tombstones == true
	}() == true {
		__antithesis_instrumentation__.Notify(643529)
		return p.advanceKey()
	} else {
		__antithesis_instrumentation__.Notify(643530)
	}
	__antithesis_instrumentation__.Notify(643523)

	if p.targetBytes > 0 && func() bool {
		__antithesis_instrumentation__.Notify(643531)
		return (p.results.bytes >= p.targetBytes || func() bool {
			__antithesis_instrumentation__.Notify(643532)
			return (p.targetBytesAvoidExcess && func() bool {
				__antithesis_instrumentation__.Notify(643533)
				return p.results.bytes+int64(p.results.sizeOf(len(rawKey), len(val))) > p.targetBytes == true
			}() == true) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(643534)
		p.resumeReason = roachpb.RESUME_BYTE_LIMIT
		p.resumeNextBytes = int64(p.results.sizeOf(len(rawKey), len(val)))

	} else {
		__antithesis_instrumentation__.Notify(643535)
		if p.maxKeys > 0 && func() bool {
			__antithesis_instrumentation__.Notify(643536)
			return p.results.count >= p.maxKeys == true
		}() == true {
			__antithesis_instrumentation__.Notify(643537)
			p.resumeReason = roachpb.RESUME_KEY_LIMIT
		} else {
			__antithesis_instrumentation__.Notify(643538)
		}
	}
	__antithesis_instrumentation__.Notify(643524)

	if p.resumeReason != 0 {
		__antithesis_instrumentation__.Notify(643539)

		if !p.allowEmpty && func() bool {
			__antithesis_instrumentation__.Notify(643540)
			return (p.results.count == 0 || func() bool {
				__antithesis_instrumentation__.Notify(643541)
				return (p.wholeRows && func() bool {
					__antithesis_instrumentation__.Notify(643542)
					return p.results.continuesFirstRow(key) == true
				}() == true) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(643543)
			p.resumeReason = 0
			p.resumeNextBytes = 0
		} else {
			__antithesis_instrumentation__.Notify(643544)
			p.resumeKey = key

			if p.wholeRows {
				__antithesis_instrumentation__.Notify(643546)
				trimmedKey, err := p.results.maybeTrimPartialLastRow(key)
				if err != nil {
					__antithesis_instrumentation__.Notify(643548)
					p.err = err
					return false
				} else {
					__antithesis_instrumentation__.Notify(643549)
				}
				__antithesis_instrumentation__.Notify(643547)
				if trimmedKey != nil {
					__antithesis_instrumentation__.Notify(643550)
					p.resumeKey = trimmedKey
				} else {
					__antithesis_instrumentation__.Notify(643551)
				}
			} else {
				__antithesis_instrumentation__.Notify(643552)
			}
			__antithesis_instrumentation__.Notify(643545)
			return false
		}
	} else {
		__antithesis_instrumentation__.Notify(643553)
	}
	__antithesis_instrumentation__.Notify(643525)

	if err := p.results.put(ctx, rawKey, val, p.memAccount); err != nil {
		__antithesis_instrumentation__.Notify(643554)
		p.err = errors.Wrapf(err, "scan with start key %s", p.start)
		return false
	} else {
		__antithesis_instrumentation__.Notify(643555)
	}
	__antithesis_instrumentation__.Notify(643526)

	if p.maxKeys > 0 && func() bool {
		__antithesis_instrumentation__.Notify(643556)
		return p.results.count >= p.maxKeys == true
	}() == true {
		__antithesis_instrumentation__.Notify(643557)

		if !p.wholeRows || func() bool {
			__antithesis_instrumentation__.Notify(643558)
			return p.results.lastRowHasFinalColumnFamily() == true
		}() == true {
			__antithesis_instrumentation__.Notify(643559)
			p.resumeReason = roachpb.RESUME_KEY_LIMIT
			return false
		} else {
			__antithesis_instrumentation__.Notify(643560)
		}
	} else {
		__antithesis_instrumentation__.Notify(643561)
	}
	__antithesis_instrumentation__.Notify(643527)
	return p.advanceKey()
}

func (p *pebbleMVCCScanner) seekVersion(
	ctx context.Context, seekTS hlc.Timestamp, uncertaintyCheck bool,
) bool {
	__antithesis_instrumentation__.Notify(643562)
	seekKey := MVCCKey{Key: p.curUnsafeKey.Key, Timestamp: seekTS}
	p.keyBuf = EncodeMVCCKeyToBuf(p.keyBuf[:0], seekKey)
	origKey := p.keyBuf[:len(p.curUnsafeKey.Key)]

	seekKey.Key = origKey

	for i := 0; i < p.itersBeforeSeek; i++ {
		__antithesis_instrumentation__.Notify(643565)
		if !p.iterNext() {
			__antithesis_instrumentation__.Notify(643568)
			return p.advanceKeyAtEnd()
		} else {
			__antithesis_instrumentation__.Notify(643569)
		}
		__antithesis_instrumentation__.Notify(643566)
		if !bytes.Equal(p.curUnsafeKey.Key, origKey) {
			__antithesis_instrumentation__.Notify(643570)
			p.incrementItersBeforeSeek()
			return p.advanceKeyAtNewKey(origKey)
		} else {
			__antithesis_instrumentation__.Notify(643571)
		}
		__antithesis_instrumentation__.Notify(643567)
		if p.curUnsafeKey.Timestamp.LessEq(seekTS) {
			__antithesis_instrumentation__.Notify(643572)
			p.incrementItersBeforeSeek()
			if !uncertaintyCheck || func() bool {
				__antithesis_instrumentation__.Notify(643574)
				return p.curUnsafeKey.Timestamp.LessEq(p.ts) == true
			}() == true {
				__antithesis_instrumentation__.Notify(643575)
				return p.addAndAdvance(ctx, p.curUnsafeKey.Key, p.curRawKey, p.curValue)
			} else {
				__antithesis_instrumentation__.Notify(643576)
			}
			__antithesis_instrumentation__.Notify(643573)

			if p.uncertainty.IsUncertain(p.curUnsafeKey.Timestamp) {
				__antithesis_instrumentation__.Notify(643577)
				return p.uncertaintyError(p.curUnsafeKey.Timestamp)
			} else {
				__antithesis_instrumentation__.Notify(643578)
			}
		} else {
			__antithesis_instrumentation__.Notify(643579)
		}
	}
	__antithesis_instrumentation__.Notify(643563)

	p.decrementItersBeforeSeek()
	if !p.iterSeek(seekKey) {
		__antithesis_instrumentation__.Notify(643580)
		return p.advanceKeyAtEnd()
	} else {
		__antithesis_instrumentation__.Notify(643581)
	}
	__antithesis_instrumentation__.Notify(643564)
	for {
		__antithesis_instrumentation__.Notify(643582)
		if !bytes.Equal(p.curUnsafeKey.Key, origKey) {
			__antithesis_instrumentation__.Notify(643586)
			return p.advanceKeyAtNewKey(origKey)
		} else {
			__antithesis_instrumentation__.Notify(643587)
		}
		__antithesis_instrumentation__.Notify(643583)
		if !uncertaintyCheck || func() bool {
			__antithesis_instrumentation__.Notify(643588)
			return p.curUnsafeKey.Timestamp.LessEq(p.ts) == true
		}() == true {
			__antithesis_instrumentation__.Notify(643589)
			return p.addAndAdvance(ctx, p.curUnsafeKey.Key, p.curRawKey, p.curValue)
		} else {
			__antithesis_instrumentation__.Notify(643590)
		}
		__antithesis_instrumentation__.Notify(643584)

		if p.uncertainty.IsUncertain(p.curUnsafeKey.Timestamp) {
			__antithesis_instrumentation__.Notify(643591)
			return p.uncertaintyError(p.curUnsafeKey.Timestamp)
		} else {
			__antithesis_instrumentation__.Notify(643592)
		}
		__antithesis_instrumentation__.Notify(643585)
		if !p.iterNext() {
			__antithesis_instrumentation__.Notify(643593)
			return p.advanceKeyAtEnd()
		} else {
			__antithesis_instrumentation__.Notify(643594)
		}
	}
}

func (p *pebbleMVCCScanner) updateCurrent() bool {
	__antithesis_instrumentation__.Notify(643595)
	if !p.iterValid() {
		__antithesis_instrumentation__.Notify(643598)
		return false
	} else {
		__antithesis_instrumentation__.Notify(643599)
	}
	__antithesis_instrumentation__.Notify(643596)

	p.curRawKey = p.parent.UnsafeRawMVCCKey()

	var err error
	p.curUnsafeKey, err = DecodeMVCCKey(p.curRawKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(643600)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(643601)
	}
	__antithesis_instrumentation__.Notify(643597)
	p.curValue = p.parent.UnsafeValue()
	return true
}

func (p *pebbleMVCCScanner) iterValid() bool {
	__antithesis_instrumentation__.Notify(643602)
	if valid, err := p.parent.Valid(); !valid {
		__antithesis_instrumentation__.Notify(643604)

		if err != nil {
			__antithesis_instrumentation__.Notify(643606)
			p.err = err
		} else {
			__antithesis_instrumentation__.Notify(643607)
		}
		__antithesis_instrumentation__.Notify(643605)
		return false
	} else {
		__antithesis_instrumentation__.Notify(643608)
	}
	__antithesis_instrumentation__.Notify(643603)
	return true
}

func (p *pebbleMVCCScanner) iterSeek(key MVCCKey) bool {
	__antithesis_instrumentation__.Notify(643609)
	p.clearPeeked()
	p.parent.SeekGE(key)
	return p.updateCurrent()
}

func (p *pebbleMVCCScanner) iterSeekReverse(key MVCCKey) bool {
	__antithesis_instrumentation__.Notify(643610)
	p.clearPeeked()
	p.parent.SeekLT(key)
	if !p.updateCurrent() {
		__antithesis_instrumentation__.Notify(643613)

		return false
	} else {
		__antithesis_instrumentation__.Notify(643614)
	}
	__antithesis_instrumentation__.Notify(643611)

	if p.curUnsafeKey.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(643615)

		return true
	} else {
		__antithesis_instrumentation__.Notify(643616)
	}
	__antithesis_instrumentation__.Notify(643612)

	return p.backwardLatestVersion(p.curUnsafeKey.Key, 0)
}

func (p *pebbleMVCCScanner) iterNext() bool {
	__antithesis_instrumentation__.Notify(643617)
	if p.reverse && func() bool {
		__antithesis_instrumentation__.Notify(643619)
		return p.peeked == true
	}() == true {
		__antithesis_instrumentation__.Notify(643620)

		p.peeked = false
		if !p.iterValid() {
			__antithesis_instrumentation__.Notify(643622)

			p.parent.SeekGE(MVCCKey{Key: p.start})
			if !p.iterValid() {
				__antithesis_instrumentation__.Notify(643624)
				return false
			} else {
				__antithesis_instrumentation__.Notify(643625)
			}
			__antithesis_instrumentation__.Notify(643623)
			p.parent.Next()
			return p.updateCurrent()
		} else {
			__antithesis_instrumentation__.Notify(643626)
		}
		__antithesis_instrumentation__.Notify(643621)
		p.parent.Next()
		if !p.iterValid() {
			__antithesis_instrumentation__.Notify(643627)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643628)
		}
	} else {
		__antithesis_instrumentation__.Notify(643629)
	}
	__antithesis_instrumentation__.Notify(643618)
	p.parent.Next()
	return p.updateCurrent()
}

func (p *pebbleMVCCScanner) iterPrev() bool {
	__antithesis_instrumentation__.Notify(643630)
	if p.peeked {
		__antithesis_instrumentation__.Notify(643632)
		p.peeked = false
		return p.updateCurrent()
	} else {
		__antithesis_instrumentation__.Notify(643633)
	}
	__antithesis_instrumentation__.Notify(643631)
	p.parent.Prev()
	return p.updateCurrent()
}

func (p *pebbleMVCCScanner) iterPeekPrev() ([]byte, bool) {
	__antithesis_instrumentation__.Notify(643634)
	if !p.peeked {
		__antithesis_instrumentation__.Notify(643636)
		p.peeked = true

		p.savedBuf = append(p.savedBuf[:0], p.curRawKey...)
		p.savedBuf = append(p.savedBuf, p.curValue...)
		p.curRawKey = p.savedBuf[:len(p.curRawKey)]
		p.curValue = p.savedBuf[len(p.curRawKey):]

		p.curUnsafeKey.Key = p.curRawKey[:len(p.curUnsafeKey.Key)]

		p.parent.Prev()
		if !p.iterValid() {
			__antithesis_instrumentation__.Notify(643637)

			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(643638)
		}
	} else {
		__antithesis_instrumentation__.Notify(643639)
		if !p.iterValid() {
			__antithesis_instrumentation__.Notify(643640)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(643641)
		}
	}
	__antithesis_instrumentation__.Notify(643635)

	peekedKey := p.parent.UnsafeKey()
	return peekedKey.Key, true
}

func (p *pebbleMVCCScanner) clearPeeked() {
	__antithesis_instrumentation__.Notify(643642)
	if p.reverse {
		__antithesis_instrumentation__.Notify(643643)
		p.peeked = false
	} else {
		__antithesis_instrumentation__.Notify(643644)
	}
}

func (p *pebbleMVCCScanner) intentsRepr() []byte {
	__antithesis_instrumentation__.Notify(643645)
	if p.intents.Count() == 0 {
		__antithesis_instrumentation__.Notify(643647)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(643648)
	}
	__antithesis_instrumentation__.Notify(643646)
	return p.intents.Repr()
}

func (p *pebbleMVCCScanner) stats() IteratorStats {
	__antithesis_instrumentation__.Notify(643649)
	return p.parent.Stats()
}
