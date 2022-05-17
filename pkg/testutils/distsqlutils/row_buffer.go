package distsqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type BufferedRecord struct {
	row  rowenc.EncDatumRow
	Meta *execinfrapb.ProducerMetadata
}

type RowBuffer struct {
	Mu struct {
		syncutil.Mutex

		producerClosed bool

		Records []BufferedRecord
	}

	Done bool

	ConsumerStatus execinfra.ConsumerStatus

	types []*types.T

	args RowBufferArgs
}

var _ execinfra.RowReceiver = &RowBuffer{}
var _ execinfra.RowSource = &RowBuffer{}

type RowBufferArgs struct {
	AccumulateRowsWhileDraining bool

	OnConsumerDone func(*RowBuffer)

	OnConsumerClosed func(*RowBuffer)

	OnNext func(*RowBuffer) (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata)

	OnPush func(rowenc.EncDatumRow, *execinfrapb.ProducerMetadata)
}

func NewRowBuffer(types []*types.T, rows rowenc.EncDatumRows, hooks RowBufferArgs) *RowBuffer {
	__antithesis_instrumentation__.Notify(644080)
	if types == nil {
		__antithesis_instrumentation__.Notify(644083)
		panic("types required")
	} else {
		__antithesis_instrumentation__.Notify(644084)
	}
	__antithesis_instrumentation__.Notify(644081)
	wrappedRows := make([]BufferedRecord, len(rows))
	for i, row := range rows {
		__antithesis_instrumentation__.Notify(644085)
		wrappedRows[i].row = row
	}
	__antithesis_instrumentation__.Notify(644082)
	rb := &RowBuffer{types: types, args: hooks}
	rb.Mu.Records = wrappedRows
	return rb
}

func (rb *RowBuffer) Push(
	row rowenc.EncDatumRow, meta *execinfrapb.ProducerMetadata,
) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(644086)
	if rb.args.OnPush != nil {
		__antithesis_instrumentation__.Notify(644091)
		rb.args.OnPush(row, meta)
	} else {
		__antithesis_instrumentation__.Notify(644092)
	}
	__antithesis_instrumentation__.Notify(644087)
	rb.Mu.Lock()
	defer rb.Mu.Unlock()
	if rb.Mu.producerClosed {
		__antithesis_instrumentation__.Notify(644093)
		panic("Push called after ProducerDone")
	} else {
		__antithesis_instrumentation__.Notify(644094)
	}
	__antithesis_instrumentation__.Notify(644088)

	storeRow := func() {
		__antithesis_instrumentation__.Notify(644095)
		var rowCopy rowenc.EncDatumRow
		if len(row) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(644097)
			return row != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(644098)

			rowCopy = make(rowenc.EncDatumRow, 0)
		} else {
			__antithesis_instrumentation__.Notify(644099)
			rowCopy = append(rowCopy, row...)
		}
		__antithesis_instrumentation__.Notify(644096)
		rb.Mu.Records = append(rb.Mu.Records, BufferedRecord{row: rowCopy, Meta: meta})
	}
	__antithesis_instrumentation__.Notify(644089)
	status := execinfra.ConsumerStatus(atomic.LoadUint32((*uint32)(&rb.ConsumerStatus)))
	if rb.args.AccumulateRowsWhileDraining {
		__antithesis_instrumentation__.Notify(644100)
		storeRow()
	} else {
		__antithesis_instrumentation__.Notify(644101)
		switch status {
		case execinfra.NeedMoreRows:
			__antithesis_instrumentation__.Notify(644102)
			storeRow()
		case execinfra.DrainRequested:
			__antithesis_instrumentation__.Notify(644103)
			if meta != nil {
				__antithesis_instrumentation__.Notify(644106)
				storeRow()
			} else {
				__antithesis_instrumentation__.Notify(644107)
			}
		case execinfra.ConsumerClosed:
			__antithesis_instrumentation__.Notify(644104)
		default:
			__antithesis_instrumentation__.Notify(644105)
		}
	}
	__antithesis_instrumentation__.Notify(644090)
	return status
}

func (rb *RowBuffer) ProducerClosed() bool {
	__antithesis_instrumentation__.Notify(644108)
	rb.Mu.Lock()
	c := rb.Mu.producerClosed
	rb.Mu.Unlock()
	return c
}

func (rb *RowBuffer) ProducerDone() {
	__antithesis_instrumentation__.Notify(644109)
	rb.Mu.Lock()
	defer rb.Mu.Unlock()
	if rb.Mu.producerClosed {
		__antithesis_instrumentation__.Notify(644111)
		panic("RowBuffer already closed")
	} else {
		__antithesis_instrumentation__.Notify(644112)
	}
	__antithesis_instrumentation__.Notify(644110)
	rb.Mu.producerClosed = true
}

func (rb *RowBuffer) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(644113)
	if rb.types == nil {
		__antithesis_instrumentation__.Notify(644115)
		panic("not initialized")
	} else {
		__antithesis_instrumentation__.Notify(644116)
	}
	__antithesis_instrumentation__.Notify(644114)
	return rb.types
}

func (rb *RowBuffer) Start(ctx context.Context) { __antithesis_instrumentation__.Notify(644117) }

func (rb *RowBuffer) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(644118)
	if rb.args.OnNext != nil {
		__antithesis_instrumentation__.Notify(644121)
		row, meta := rb.args.OnNext(rb)
		if row != nil || func() bool {
			__antithesis_instrumentation__.Notify(644122)
			return meta != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(644123)
			return row, meta
		} else {
			__antithesis_instrumentation__.Notify(644124)
		}
	} else {
		__antithesis_instrumentation__.Notify(644125)
	}
	__antithesis_instrumentation__.Notify(644119)
	if len(rb.Mu.Records) == 0 {
		__antithesis_instrumentation__.Notify(644126)
		rb.Done = true
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(644127)
	}
	__antithesis_instrumentation__.Notify(644120)
	rec := rb.Mu.Records[0]
	rb.Mu.Records = rb.Mu.Records[1:]
	return rec.row, rec.Meta
}

func (rb *RowBuffer) ConsumerDone() {
	__antithesis_instrumentation__.Notify(644128)
	if atomic.CompareAndSwapUint32((*uint32)(&rb.ConsumerStatus),
		uint32(execinfra.NeedMoreRows), uint32(execinfra.DrainRequested)) {
		__antithesis_instrumentation__.Notify(644129)
		if rb.args.OnConsumerDone != nil {
			__antithesis_instrumentation__.Notify(644130)
			rb.args.OnConsumerDone(rb)
		} else {
			__antithesis_instrumentation__.Notify(644131)
		}
	} else {
		__antithesis_instrumentation__.Notify(644132)
	}
}

func (rb *RowBuffer) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(644133)
	status := execinfra.ConsumerStatus(atomic.LoadUint32((*uint32)(&rb.ConsumerStatus)))
	if status == execinfra.ConsumerClosed {
		__antithesis_instrumentation__.Notify(644135)
		log.Fatalf(context.Background(), "RowBuffer already closed")
	} else {
		__antithesis_instrumentation__.Notify(644136)
	}
	__antithesis_instrumentation__.Notify(644134)
	atomic.StoreUint32((*uint32)(&rb.ConsumerStatus), uint32(execinfra.ConsumerClosed))
	if rb.args.OnConsumerClosed != nil {
		__antithesis_instrumentation__.Notify(644137)
		rb.args.OnConsumerClosed(rb)
	} else {
		__antithesis_instrumentation__.Notify(644138)
	}
}

func (rb *RowBuffer) NextNoMeta(tb testing.TB) rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(644139)
	row, meta := rb.Next()
	if meta != nil {
		__antithesis_instrumentation__.Notify(644141)
		tb.Fatalf("unexpected metadata: %v", meta)
	} else {
		__antithesis_instrumentation__.Notify(644142)
	}
	__antithesis_instrumentation__.Notify(644140)
	return row
}

func (rb *RowBuffer) GetRowsNoMeta(t *testing.T) rowenc.EncDatumRows {
	__antithesis_instrumentation__.Notify(644143)
	var res rowenc.EncDatumRows
	for {
		__antithesis_instrumentation__.Notify(644145)
		row := rb.NextNoMeta(t)
		if row == nil {
			__antithesis_instrumentation__.Notify(644147)
			break
		} else {
			__antithesis_instrumentation__.Notify(644148)
		}
		__antithesis_instrumentation__.Notify(644146)
		res = append(res, row)
	}
	__antithesis_instrumentation__.Notify(644144)
	return res
}
