package txnidcache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

type Reader interface {
	Lookup(txnID uuid.UUID) (result roachpb.TransactionFingerprintID, found bool)
}

type Writer interface {
	Record(resolvedTxnID contentionpb.ResolvedTxnID)

	DrainWriteBuffer()
}

type blockSink interface {
	push(*block)
}

const channelSize = 128

type Cache struct {
	st *cluster.Settings

	blockCh chan *block
	closeCh chan struct{}

	store  *fifoCache
	writer Writer

	metrics *Metrics
}

var (
	entrySize = int64(uuid.UUID{}.Size()) +
		roachpb.TransactionFingerprintID(0).Size()
)

var (
	_ Reader    = &Cache{}
	_ Writer    = &Cache{}
	_ blockSink = &Cache{}
)

func NewTxnIDCache(st *cluster.Settings, metrics *Metrics) *Cache {
	__antithesis_instrumentation__.Notify(459381)
	t := &Cache{
		st:      st,
		metrics: metrics,
		blockCh: make(chan *block, channelSize),
		closeCh: make(chan struct{}),
	}

	t.store = newFIFOCache(func() int64 { __antithesis_instrumentation__.Notify(459383); return MaxSize.Get(&st.SV) })
	__antithesis_instrumentation__.Notify(459382)
	t.writer = newWriter(st, t)
	return t
}

func (t *Cache) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(459384)
	err := stopper.RunAsyncTask(ctx, "txn-id-cache-ingest", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(459386)
		for {
			__antithesis_instrumentation__.Notify(459387)
			select {
			case b := <-t.blockCh:
				__antithesis_instrumentation__.Notify(459388)
				t.store.add(b)
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(459389)
				close(t.closeCh)
				return
			}
		}
	})
	__antithesis_instrumentation__.Notify(459385)
	if err != nil {
		__antithesis_instrumentation__.Notify(459390)
		close(t.closeCh)
	} else {
		__antithesis_instrumentation__.Notify(459391)
	}
}

func (t *Cache) Lookup(txnID uuid.UUID) (result roachpb.TransactionFingerprintID, found bool) {
	__antithesis_instrumentation__.Notify(459392)
	t.metrics.CacheReadCounter.Inc(1)

	txnFingerprintID, found := t.store.get(txnID)
	if !found {
		__antithesis_instrumentation__.Notify(459394)
		t.metrics.CacheMissCounter.Inc(1)
		return roachpb.InvalidTransactionFingerprintID, found
	} else {
		__antithesis_instrumentation__.Notify(459395)
	}
	__antithesis_instrumentation__.Notify(459393)

	return txnFingerprintID, found
}

func (t *Cache) Record(resolvedTxnID contentionpb.ResolvedTxnID) {
	__antithesis_instrumentation__.Notify(459396)
	t.writer.Record(resolvedTxnID)
}

func (t *Cache) push(b *block) {
	__antithesis_instrumentation__.Notify(459397)
	select {
	case t.blockCh <- b:
		__antithesis_instrumentation__.Notify(459398)
	case <-t.closeCh:
		__antithesis_instrumentation__.Notify(459399)
	}
}

func (t *Cache) DrainWriteBuffer() {
	__antithesis_instrumentation__.Notify(459400)
	t.writer.DrainWriteBuffer()
}

func (t *Cache) Size() int64 {
	__antithesis_instrumentation__.Notify(459401)
	return t.store.size()
}
