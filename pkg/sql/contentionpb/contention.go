package contentionpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

const singleIndentation = "  "
const doubleIndentation = singleIndentation + singleIndentation
const tripleIndentation = doubleIndentation + singleIndentation

const contentionEventsStr = "num contention events:"
const cumulativeContentionTimeStr = "cumulative contention time:"
const contendingTxnsStr = "contending txns:"

func (ice IndexContentionEvents) String() string {
	__antithesis_instrumentation__.Notify(459418)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("tableID=%d indexID=%d\n", ice.TableID, ice.IndexID))
	b.WriteString(fmt.Sprintf("%s%s %d\n", singleIndentation, contentionEventsStr, ice.NumContentionEvents))
	b.WriteString(fmt.Sprintf("%s%s %s\n", singleIndentation, cumulativeContentionTimeStr, ice.CumulativeContentionTime))
	b.WriteString(fmt.Sprintf("%skeys:\n", singleIndentation))
	for i := range ice.Events {
		__antithesis_instrumentation__.Notify(459420)
		b.WriteString(ice.Events[i].String())
	}
	__antithesis_instrumentation__.Notify(459419)
	return b.String()
}

func (skc SingleKeyContention) String() string {
	__antithesis_instrumentation__.Notify(459421)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s%s %s\n", doubleIndentation, skc.Key, contendingTxnsStr))
	for i := range skc.Txns {
		__antithesis_instrumentation__.Notify(459423)
		b.WriteString(skc.Txns[i].String())
	}
	__antithesis_instrumentation__.Notify(459422)
	return b.String()
}

func toString(stx SingleTxnContention, indentation string) string {
	__antithesis_instrumentation__.Notify(459424)
	return fmt.Sprintf("%sid=%s count=%d\n", indentation, stx.TxnID, stx.Count)
}

func (stx SingleTxnContention) String() string {
	__antithesis_instrumentation__.Notify(459425)
	return toString(stx, tripleIndentation)
}

func (skc SingleNonSQLKeyContention) String() string {
	__antithesis_instrumentation__.Notify(459426)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("non-SQL key %s %s\n", skc.Key, contendingTxnsStr))
	b.WriteString(fmt.Sprintf("%s%s %d\n", singleIndentation, contentionEventsStr, skc.NumContentionEvents))
	b.WriteString(fmt.Sprintf("%s%s %s\n", singleIndentation, cumulativeContentionTimeStr, skc.CumulativeContentionTime))
	for i := range skc.Txns {
		__antithesis_instrumentation__.Notify(459428)
		b.WriteString(toString(skc.Txns[i], doubleIndentation))
	}
	__antithesis_instrumentation__.Notify(459427)
	return b.String()
}

func (r *ResolvedTxnID) Valid() bool {
	__antithesis_instrumentation__.Notify(459429)
	return !uuid.Nil.Equal(r.TxnID)
}

func (e *ExtendedContentionEvent) Valid() bool {
	__antithesis_instrumentation__.Notify(459430)
	return !uuid.Nil.Equal(e.BlockingEvent.TxnMeta.ID)
}

func (e *ExtendedContentionEvent) Hash() uint64 {
	__antithesis_instrumentation__.Notify(459431)
	hash := util.MakeFNV64()
	hashUUID(e.BlockingEvent.TxnMeta.ID, &hash)
	hashUUID(e.WaitingTxnID, &hash)
	hash.Add(uint64(e.CollectionTs.UnixMilli()))
	return hash.Sum()
}

func hashUUID(u uuid.UUID, fnv *util.FNV64) {
	__antithesis_instrumentation__.Notify(459432)
	b := u.GetBytes()

	b, val, err := encoding.DecodeUint64Descending(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(459435)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(459436)
	}
	__antithesis_instrumentation__.Notify(459433)
	fnv.Add(val)
	_, val, err = encoding.DecodeUint64Descending(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(459437)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(459438)
	}
	__antithesis_instrumentation__.Notify(459434)
	fnv.Add(val)
}
