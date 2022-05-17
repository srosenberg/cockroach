package txnidcache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

const shardCount = 16

type writer struct {
	st *cluster.Settings

	shards [shardCount]*concurrentWriteBuffer

	sink blockSink
}

var _ Writer = &writer{}

func newWriter(st *cluster.Settings, sink blockSink) *writer {
	__antithesis_instrumentation__.Notify(459402)
	w := &writer{
		st:   st,
		sink: sink,
	}

	for shardIdx := 0; shardIdx < shardCount; shardIdx++ {
		__antithesis_instrumentation__.Notify(459404)
		w.shards[shardIdx] = newConcurrentWriteBuffer(sink)
	}
	__antithesis_instrumentation__.Notify(459403)

	return w
}

func (w *writer) Record(resolvedTxnID contentionpb.ResolvedTxnID) {
	__antithesis_instrumentation__.Notify(459405)
	if MaxSize.Get(&w.st.SV) == 0 {
		__antithesis_instrumentation__.Notify(459408)
		return
	} else {
		__antithesis_instrumentation__.Notify(459409)
	}
	__antithesis_instrumentation__.Notify(459406)

	if resolvedTxnID.TxnID.Equal(uuid.Nil) {
		__antithesis_instrumentation__.Notify(459410)
		return
	} else {
		__antithesis_instrumentation__.Notify(459411)
	}
	__antithesis_instrumentation__.Notify(459407)
	shardIdx := hashTxnID(resolvedTxnID.TxnID)
	buffer := w.shards[shardIdx]
	buffer.Record(resolvedTxnID)
}

func (w *writer) DrainWriteBuffer() {
	__antithesis_instrumentation__.Notify(459412)
	for shardIdx := 0; shardIdx < shardCount; shardIdx++ {
		__antithesis_instrumentation__.Notify(459413)
		w.shards[shardIdx].DrainWriteBuffer()
	}
}

func hashTxnID(txnID uuid.UUID) int {
	__antithesis_instrumentation__.Notify(459414)
	b := txnID.GetBytes()
	_, val, err := encoding.DecodeUint64Descending(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(459416)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(459417)
	}
	__antithesis_instrumentation__.Notify(459415)
	return int(val % shardCount)
}
