package enginepb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"sort"

	"github.com/cockroachdb/redact"
)

type TxnEpoch int32

func (TxnEpoch) SafeValue() { __antithesis_instrumentation__.Notify(634533) }

type TxnSeq int32

func (TxnSeq) SafeValue() { __antithesis_instrumentation__.Notify(634534) }

type TxnPriority int32

const (
	MinTxnPriority TxnPriority = 0

	MaxTxnPriority TxnPriority = math.MaxInt32
)

func TxnSeqIsIgnored(seq TxnSeq, ignored []IgnoredSeqNumRange) bool {
	__antithesis_instrumentation__.Notify(634535)

	for i := len(ignored) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(634537)
		if seq < ignored[i].Start {
			__antithesis_instrumentation__.Notify(634540)

			continue
		} else {
			__antithesis_instrumentation__.Notify(634541)
		}
		__antithesis_instrumentation__.Notify(634538)

		if seq > ignored[i].End {
			__antithesis_instrumentation__.Notify(634542)

			return false
		} else {
			__antithesis_instrumentation__.Notify(634543)
		}
		__antithesis_instrumentation__.Notify(634539)

		return true
	}
	__antithesis_instrumentation__.Notify(634536)

	return false
}

func (t TxnMeta) Short() redact.SafeString {
	__antithesis_instrumentation__.Notify(634544)
	return redact.SafeString(t.ID.Short())
}

func (ms MVCCStats) Total() int64 {
	__antithesis_instrumentation__.Notify(634545)
	return ms.KeyBytes + ms.ValBytes
}

func (ms MVCCStats) GCBytes() int64 {
	__antithesis_instrumentation__.Notify(634546)
	return ms.KeyBytes + ms.ValBytes - ms.LiveBytes
}

func (ms MVCCStats) AvgIntentAge(nowNanos int64) float64 {
	__antithesis_instrumentation__.Notify(634547)
	if ms.IntentCount == 0 {
		__antithesis_instrumentation__.Notify(634549)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(634550)
	}
	__antithesis_instrumentation__.Notify(634548)

	ms.AgeTo(nowNanos)
	return float64(ms.IntentAge) / float64(ms.IntentCount)
}

func (ms MVCCStats) GCByteAge(nowNanos int64) int64 {
	__antithesis_instrumentation__.Notify(634551)
	ms.AgeTo(nowNanos)
	return ms.GCBytesAge
}

func (ms *MVCCStats) Forward(nowNanos int64) {
	__antithesis_instrumentation__.Notify(634552)
	if ms.LastUpdateNanos >= nowNanos {
		__antithesis_instrumentation__.Notify(634554)
		return
	} else {
		__antithesis_instrumentation__.Notify(634555)
	}
	__antithesis_instrumentation__.Notify(634553)
	ms.AgeTo(nowNanos)
}

func (ms *MVCCStats) AgeTo(nowNanos int64) {
	__antithesis_instrumentation__.Notify(634556)

	diffSeconds := nowNanos/1e9 - ms.LastUpdateNanos/1e9

	ms.GCBytesAge += ms.GCBytes() * diffSeconds
	ms.IntentAge += ms.IntentCount * diffSeconds
	ms.LastUpdateNanos = nowNanos
}

func (ms *MVCCStats) Add(oms MVCCStats) {
	__antithesis_instrumentation__.Notify(634557)

	ms.Forward(oms.LastUpdateNanos)
	oms.Forward(ms.LastUpdateNanos)

	ms.ContainsEstimates += oms.ContainsEstimates

	ms.IntentAge += oms.IntentAge
	ms.GCBytesAge += oms.GCBytesAge
	ms.LiveBytes += oms.LiveBytes
	ms.KeyBytes += oms.KeyBytes
	ms.ValBytes += oms.ValBytes
	ms.IntentBytes += oms.IntentBytes
	ms.LiveCount += oms.LiveCount
	ms.KeyCount += oms.KeyCount
	ms.ValCount += oms.ValCount
	ms.IntentCount += oms.IntentCount
	ms.SeparatedIntentCount += oms.SeparatedIntentCount
	ms.SysBytes += oms.SysBytes
	ms.SysCount += oms.SysCount
	ms.AbortSpanBytes += oms.AbortSpanBytes
}

func (ms *MVCCStats) Subtract(oms MVCCStats) {
	__antithesis_instrumentation__.Notify(634558)

	ms.Forward(oms.LastUpdateNanos)
	oms.Forward(ms.LastUpdateNanos)

	ms.ContainsEstimates -= oms.ContainsEstimates

	ms.IntentAge -= oms.IntentAge
	ms.GCBytesAge -= oms.GCBytesAge
	ms.LiveBytes -= oms.LiveBytes
	ms.KeyBytes -= oms.KeyBytes
	ms.ValBytes -= oms.ValBytes
	ms.IntentBytes -= oms.IntentBytes
	ms.LiveCount -= oms.LiveCount
	ms.KeyCount -= oms.KeyCount
	ms.ValCount -= oms.ValCount
	ms.IntentCount -= oms.IntentCount
	ms.SeparatedIntentCount -= oms.SeparatedIntentCount
	ms.SysBytes -= oms.SysBytes
	ms.SysCount -= oms.SysCount
	ms.AbortSpanBytes -= oms.AbortSpanBytes
}

func (meta MVCCMetadata) IsInline() bool {
	__antithesis_instrumentation__.Notify(634559)
	return meta.RawBytes != nil
}

func (meta *MVCCMetadata) AddToIntentHistory(seq TxnSeq, val []byte) {
	__antithesis_instrumentation__.Notify(634560)
	meta.IntentHistory = append(meta.IntentHistory,
		MVCCMetadata_SequencedIntent{Sequence: seq, Value: val})
}

func (meta *MVCCMetadata) GetPrevIntentSeq(
	seq TxnSeq, ignored []IgnoredSeqNumRange,
) (MVCCMetadata_SequencedIntent, bool) {
	__antithesis_instrumentation__.Notify(634561)
	end := len(meta.IntentHistory)
	found := 0
	for {
		__antithesis_instrumentation__.Notify(634563)
		index := sort.Search(end, func(i int) bool {
			__antithesis_instrumentation__.Notify(634567)
			return meta.IntentHistory[i].Sequence >= seq
		})
		__antithesis_instrumentation__.Notify(634564)
		if index == 0 {
			__antithesis_instrumentation__.Notify(634568)

			return MVCCMetadata_SequencedIntent{}, false
		} else {
			__antithesis_instrumentation__.Notify(634569)
		}
		__antithesis_instrumentation__.Notify(634565)
		candidate := index - 1
		if TxnSeqIsIgnored(meta.IntentHistory[candidate].Sequence, ignored) {
			__antithesis_instrumentation__.Notify(634570)

			end = candidate
			continue
		} else {
			__antithesis_instrumentation__.Notify(634571)
		}
		__antithesis_instrumentation__.Notify(634566)

		found = candidate
		break
	}
	__antithesis_instrumentation__.Notify(634562)
	return meta.IntentHistory[found], true
}

func (meta *MVCCMetadata) GetIntentValue(seq TxnSeq) ([]byte, bool) {
	__antithesis_instrumentation__.Notify(634572)
	index := sort.Search(len(meta.IntentHistory), func(i int) bool {
		__antithesis_instrumentation__.Notify(634575)
		return meta.IntentHistory[i].Sequence >= seq
	})
	__antithesis_instrumentation__.Notify(634573)
	if index < len(meta.IntentHistory) && func() bool {
		__antithesis_instrumentation__.Notify(634576)
		return meta.IntentHistory[index].Sequence == seq == true
	}() == true {
		__antithesis_instrumentation__.Notify(634577)
		return meta.IntentHistory[index].Value, true
	} else {
		__antithesis_instrumentation__.Notify(634578)
	}
	__antithesis_instrumentation__.Notify(634574)
	return nil, false
}

func (m MVCCMetadata_SequencedIntent) String() string {
	__antithesis_instrumentation__.Notify(634579)
	return redact.StringWithoutMarkers(m)
}

func (m MVCCMetadata_SequencedIntent) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(634580)
	w.Printf(
		"{%d %s}",
		m.Sequence,
		FormatBytesAsValue(m.Value))
}

func (meta *MVCCMetadata) String() string {
	__antithesis_instrumentation__.Notify(634581)
	return redact.StringWithoutMarkers(meta)
}

func (meta *MVCCMetadata) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(634582)
	expand := w.Flag('+')

	w.Printf("txn={%s} ts=%s del=%t klen=%d vlen=%d",
		meta.Txn,
		meta.Timestamp,
		meta.Deleted,
		meta.KeyBytes,
		meta.ValBytes,
	)

	if len(meta.RawBytes) > 0 {
		__antithesis_instrumentation__.Notify(634586)
		if expand {
			__antithesis_instrumentation__.Notify(634587)
			w.Printf(" raw=%s", FormatBytesAsValue(meta.RawBytes))
		} else {
			__antithesis_instrumentation__.Notify(634588)
			w.Printf(" rawlen=%d", len(meta.RawBytes))
		}
	} else {
		__antithesis_instrumentation__.Notify(634589)
	}
	__antithesis_instrumentation__.Notify(634583)
	if nih := len(meta.IntentHistory); nih > 0 {
		__antithesis_instrumentation__.Notify(634590)
		if expand {
			__antithesis_instrumentation__.Notify(634591)
			w.Printf(" ih={")
			for i := range meta.IntentHistory {
				__antithesis_instrumentation__.Notify(634593)
				w.Print(meta.IntentHistory[i])
			}
			__antithesis_instrumentation__.Notify(634592)
			w.Printf("}")
		} else {
			__antithesis_instrumentation__.Notify(634594)
			w.Printf(" nih=%d", nih)
		}
	} else {
		__antithesis_instrumentation__.Notify(634595)
	}
	__antithesis_instrumentation__.Notify(634584)

	var txnDidNotUpdateMeta bool
	if meta.TxnDidNotUpdateMeta != nil {
		__antithesis_instrumentation__.Notify(634596)
		txnDidNotUpdateMeta = *meta.TxnDidNotUpdateMeta
	} else {
		__antithesis_instrumentation__.Notify(634597)
	}
	__antithesis_instrumentation__.Notify(634585)
	w.Printf(" mergeTs=%s txnDidNotUpdateMeta=%t", meta.MergeTimestamp, txnDidNotUpdateMeta)
}

func (meta *MVCCMetadataSubsetForMergeSerialization) String() string {
	__antithesis_instrumentation__.Notify(634598)
	var m MVCCMetadata
	m.RawBytes = meta.RawBytes
	m.MergeTimestamp = meta.MergeTimestamp
	return m.String()
}

func (t TxnMeta) String() string {
	__antithesis_instrumentation__.Notify(634599)
	return redact.StringWithoutMarkers(t)
}

func (t TxnMeta) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(634600)

	floatPri := 100 * float64(t.Priority) / float64(math.MaxInt32)
	w.Printf(
		"id=%s key=%s pri=%.8f epo=%d ts=%s min=%s seq=%d",
		t.Short(),
		FormatBytesAsKey(t.Key),
		floatPri,
		t.Epoch,
		t.WriteTimestamp,
		t.MinTimestamp,
		t.Sequence)
}

var FormatBytesAsKey = func(k []byte) string {
	__antithesis_instrumentation__.Notify(634601)
	return string(k)
}

var FormatBytesAsValue = func(v []byte) string {
	__antithesis_instrumentation__.Notify(634602)
	return string(v)
}
