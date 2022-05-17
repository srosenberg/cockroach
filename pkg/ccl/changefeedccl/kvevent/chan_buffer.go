package kvevent

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "context"

type chanBuffer struct {
	entriesCh    chan Event
	closedReason error
}

func MakeChanBuffer() Buffer {
	__antithesis_instrumentation__.Notify(17094)
	return &chanBuffer{entriesCh: make(chan Event)}
}

func (b *chanBuffer) Add(ctx context.Context, event Event) error {
	__antithesis_instrumentation__.Notify(17095)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(17096)
		return ctx.Err()
	case b.entriesCh <- event:
		__antithesis_instrumentation__.Notify(17097)
		return nil
	}
}

func (b *chanBuffer) Drain(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17098)

	return nil
}

func (b *chanBuffer) CloseWithReason(_ context.Context, reason error) error {
	__antithesis_instrumentation__.Notify(17099)
	b.closedReason = reason
	close(b.entriesCh)
	return nil
}

func (b *chanBuffer) Get(ctx context.Context) (Event, error) {
	__antithesis_instrumentation__.Notify(17100)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(17101)
		return Event{}, ctx.Err()
	case e, ok := <-b.entriesCh:
		__antithesis_instrumentation__.Notify(17102)
		if !ok {
			__antithesis_instrumentation__.Notify(17104)

			return e, ErrBufferClosed{reason: b.closedReason}
		} else {
			__antithesis_instrumentation__.Notify(17105)
		}
		__antithesis_instrumentation__.Notify(17103)
		return e, nil
	}
}
