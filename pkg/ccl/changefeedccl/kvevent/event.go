// Package kvevent defines kvfeed events and buffers to communicate them
// locally.
package kvevent

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type ErrBufferClosed struct {
	reason error
}

func (e ErrBufferClosed) Error() string {
	__antithesis_instrumentation__.Notify(17111)
	return "buffer closed"
}

func (e ErrBufferClosed) Unwrap() error {
	__antithesis_instrumentation__.Notify(17112)
	return e.reason
}

var ErrNormalRestartReason = errors.New("writer can restart")

type Buffer interface {
	Reader
	Writer
}

type Reader interface {
	Get(ctx context.Context) (Event, error)
}

type Writer interface {
	Add(ctx context.Context, event Event) error

	Drain(ctx context.Context) error

	CloseWithReason(ctx context.Context, reason error) error
}

type Type int

const (
	TypeKV Type = iota

	TypeResolved

	TypeFlush

	TypeUnknown
)

type Event struct {
	kv                 roachpb.KeyValue
	prevVal            roachpb.Value
	flush              bool
	resolved           *jobspb.ResolvedSpan
	backfillTimestamp  hlc.Timestamp
	bufferAddTimestamp time.Time
	approxSize         int
	alloc              Alloc
}

func (b *Event) Type() Type {
	__antithesis_instrumentation__.Notify(17113)
	if b.kv.Key != nil {
		__antithesis_instrumentation__.Notify(17117)
		return TypeKV
	} else {
		__antithesis_instrumentation__.Notify(17118)
	}
	__antithesis_instrumentation__.Notify(17114)
	if b.resolved != nil {
		__antithesis_instrumentation__.Notify(17119)
		return TypeResolved
	} else {
		__antithesis_instrumentation__.Notify(17120)
	}
	__antithesis_instrumentation__.Notify(17115)
	if b.flush {
		__antithesis_instrumentation__.Notify(17121)
		return TypeFlush
	} else {
		__antithesis_instrumentation__.Notify(17122)
	}
	__antithesis_instrumentation__.Notify(17116)
	return TypeUnknown
}

func (b *Event) ApproximateSize() int {
	__antithesis_instrumentation__.Notify(17123)
	return b.approxSize
}

func (b *Event) KV() roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(17124)
	return b.kv
}

func (b *Event) PrevValue() roachpb.Value {
	__antithesis_instrumentation__.Notify(17125)
	return b.prevVal
}

func (b *Event) Resolved() *jobspb.ResolvedSpan {
	__antithesis_instrumentation__.Notify(17126)
	return b.resolved
}

func (b *Event) BackfillTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(17127)
	return b.backfillTimestamp
}

func (b *Event) BufferAddTimestamp() time.Time {
	__antithesis_instrumentation__.Notify(17128)
	return b.bufferAddTimestamp
}

func (b *Event) Timestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(17129)
	switch b.Type() {
	case TypeResolved:
		__antithesis_instrumentation__.Notify(17130)
		return b.resolved.Timestamp
	case TypeKV:
		__antithesis_instrumentation__.Notify(17131)
		if !b.backfillTimestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(17135)
			return b.backfillTimestamp
		} else {
			__antithesis_instrumentation__.Notify(17136)
		}
		__antithesis_instrumentation__.Notify(17132)
		return b.kv.Value.Timestamp
	case TypeFlush:
		__antithesis_instrumentation__.Notify(17133)
		return hlc.Timestamp{}
	default:
		__antithesis_instrumentation__.Notify(17134)
		log.Warningf(context.TODO(),
			"setting empty timestamp for unknown event type")
		return hlc.Timestamp{}
	}
}

func (b *Event) MVCCTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(17137)
	switch b.Type() {
	case TypeResolved:
		__antithesis_instrumentation__.Notify(17138)
		return b.resolved.Timestamp
	case TypeKV:
		__antithesis_instrumentation__.Notify(17139)
		return b.kv.Value.Timestamp
	case TypeFlush:
		__antithesis_instrumentation__.Notify(17140)
		return hlc.Timestamp{}
	default:
		__antithesis_instrumentation__.Notify(17141)
		log.Warningf(context.TODO(),
			"setting empty timestamp for unknown event type")
		return hlc.Timestamp{}
	}
}

func (b *Event) DetachAlloc() Alloc {
	__antithesis_instrumentation__.Notify(17142)
	a := b.alloc
	b.alloc.clear()
	return a
}

func MakeResolvedEvent(
	span roachpb.Span, ts hlc.Timestamp, boundaryType jobspb.ResolvedSpan_BoundaryType,
) Event {
	__antithesis_instrumentation__.Notify(17143)
	return Event{
		resolved: &jobspb.ResolvedSpan{
			Span:         span,
			Timestamp:    ts,
			BoundaryType: boundaryType,
		},
		approxSize: span.Size() + ts.Size() + 4,
	}
}

func MakeKVEvent(
	kv roachpb.KeyValue, prevVal roachpb.Value, backfillTimestamp hlc.Timestamp,
) Event {
	__antithesis_instrumentation__.Notify(17144)
	return Event{
		kv:                kv,
		prevVal:           prevVal,
		backfillTimestamp: backfillTimestamp,
		approxSize:        kv.Size() + prevVal.Size() + backfillTimestamp.Size(),
	}
}
