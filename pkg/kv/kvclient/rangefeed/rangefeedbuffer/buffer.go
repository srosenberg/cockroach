package rangefeedbuffer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

var ErrBufferLimitExceeded = errors.New("rangefeed buffer limit exceeded")

type Event interface {
	Timestamp() hlc.Timestamp
}

type Buffer struct {
	mu struct {
		syncutil.Mutex

		events
		frontier hlc.Timestamp
		limit    int
	}
}

func New(limit int) *Buffer {
	__antithesis_instrumentation__.Notify(89862)
	b := &Buffer{}
	b.mu.limit = limit
	return b
}

func (b *Buffer) Add(ev Event) error {
	__antithesis_instrumentation__.Notify(89863)
	b.mu.Lock()
	defer b.mu.Unlock()

	if ev.Timestamp().LessEq(b.mu.frontier) {
		__antithesis_instrumentation__.Notify(89866)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(89867)
	}
	__antithesis_instrumentation__.Notify(89864)

	if b.mu.events.Len()+1 > b.mu.limit {
		__antithesis_instrumentation__.Notify(89868)
		return ErrBufferLimitExceeded
	} else {
		__antithesis_instrumentation__.Notify(89869)
	}
	__antithesis_instrumentation__.Notify(89865)

	b.mu.events = append(b.mu.events, ev)
	return nil
}

func (b *Buffer) Flush(ctx context.Context, frontier hlc.Timestamp) (events []Event) {
	__antithesis_instrumentation__.Notify(89870)
	b.mu.Lock()
	defer b.mu.Unlock()

	if frontier.Less(b.mu.frontier) {
		__antithesis_instrumentation__.Notify(89873)
		log.Fatalf(ctx, "frontier timestamp regressed: saw %s, previously %s", frontier, b.mu.frontier)
	} else {
		__antithesis_instrumentation__.Notify(89874)
	}
	__antithesis_instrumentation__.Notify(89871)

	sort.Sort(&b.mu.events)
	idx := sort.Search(len(b.mu.events), func(i int) bool {
		__antithesis_instrumentation__.Notify(89875)
		return !b.mu.events[i].Timestamp().LessEq(frontier)
	})
	__antithesis_instrumentation__.Notify(89872)

	events = b.mu.events[:idx]
	b.mu.events = b.mu.events[idx:]
	b.mu.frontier = frontier
	return events
}

func (b *Buffer) SetLimit(limit int) {
	__antithesis_instrumentation__.Notify(89876)
	b.mu.Lock()
	defer b.mu.Unlock()

	b.mu.limit = limit
}

type events []Event

var _ sort.Interface = (*events)(nil)

func (es *events) Len() int { __antithesis_instrumentation__.Notify(89877); return len(*es) }
func (es *events) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(89878)
	return (*es)[i].Timestamp().Less((*es)[j].Timestamp())
}
func (es *events) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(89879)
	(*es)[i], (*es)[j] = (*es)[j], (*es)[i]
}
