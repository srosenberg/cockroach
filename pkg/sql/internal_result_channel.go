package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

type ieResultReader interface {
	firstResult(ctx context.Context) (_ ieIteratorResult, done bool, err error)

	nextResult(ctx context.Context) (_ ieIteratorResult, done bool, err error)

	close() error
}

type ieResultWriter interface {
	addResult(ctx context.Context, result ieIteratorResult) error

	finish()
}

var asyncIEResultChannelBufferSize = util.ConstantWithMetamorphicTestRange(
	"async-IE-result-channel-buffer-size",
	32,
	1,
	32,
)

func newAsyncIEResultChannel() *ieResultChannel {
	__antithesis_instrumentation__.Notify(497980)
	return &ieResultChannel{
		dataCh: make(chan ieIteratorResult, asyncIEResultChannelBufferSize),
		doneCh: make(chan struct{}),
	}
}

type ieResultChannel struct {
	dataCh chan ieIteratorResult

	waitCh chan struct{}

	doneCh   chan struct{}
	doneErr  error
	doneOnce sync.Once
}

func newSyncIEResultChannel() *ieResultChannel {
	__antithesis_instrumentation__.Notify(497981)
	return &ieResultChannel{
		dataCh: make(chan ieIteratorResult),
		waitCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

func (i *ieResultChannel) firstResult(
	ctx context.Context,
) (_ ieIteratorResult, done bool, err error) {
	__antithesis_instrumentation__.Notify(497982)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(497983)
		return ieIteratorResult{}, true, ctx.Err()
	case <-i.doneCh:
		__antithesis_instrumentation__.Notify(497984)
		return ieIteratorResult{}, true, ctx.Err()
	case res, ok := <-i.dataCh:
		__antithesis_instrumentation__.Notify(497985)
		if !ok {
			__antithesis_instrumentation__.Notify(497987)
			return ieIteratorResult{}, true, ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(497988)
		}
		__antithesis_instrumentation__.Notify(497986)
		return res, false, nil
	}
}

func (i *ieResultChannel) maybeUnblockWriter(ctx context.Context) (done bool, err error) {
	__antithesis_instrumentation__.Notify(497989)
	if i.async() {
		__antithesis_instrumentation__.Notify(497991)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(497992)
	}
	__antithesis_instrumentation__.Notify(497990)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(497993)
		return true, ctx.Err()
	case <-i.doneCh:
		__antithesis_instrumentation__.Notify(497994)
		return true, ctx.Err()
	case i.waitCh <- struct{}{}:
		__antithesis_instrumentation__.Notify(497995)
		return false, nil
	}
}

func (i *ieResultChannel) async() bool {
	__antithesis_instrumentation__.Notify(497996)
	return i.waitCh == nil
}

func (i *ieResultChannel) nextResult(
	ctx context.Context,
) (_ ieIteratorResult, done bool, err error) {
	__antithesis_instrumentation__.Notify(497997)
	if done, err = i.maybeUnblockWriter(ctx); done {
		__antithesis_instrumentation__.Notify(497999)
		return ieIteratorResult{}, done, err
	} else {
		__antithesis_instrumentation__.Notify(498000)
	}
	__antithesis_instrumentation__.Notify(497998)
	return i.firstResult(ctx)
}

func (i *ieResultChannel) close() error {
	__antithesis_instrumentation__.Notify(498001)
	i.doneOnce.Do(func() {
		__antithesis_instrumentation__.Notify(498003)
		close(i.doneCh)
		for {
			__antithesis_instrumentation__.Notify(498004)

			res, done, err := i.nextResult(context.TODO())
			if i.doneErr == nil {
				__antithesis_instrumentation__.Notify(498006)
				if res.err != nil {
					__antithesis_instrumentation__.Notify(498007)
					i.doneErr = res.err
				} else {
					__antithesis_instrumentation__.Notify(498008)
					if err != nil {
						__antithesis_instrumentation__.Notify(498009)
						i.doneErr = err
					} else {
						__antithesis_instrumentation__.Notify(498010)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(498011)
			}
			__antithesis_instrumentation__.Notify(498005)
			if done {
				__antithesis_instrumentation__.Notify(498012)
				return
			} else {
				__antithesis_instrumentation__.Notify(498013)
			}
		}
	})
	__antithesis_instrumentation__.Notify(498002)
	return i.doneErr
}

var errIEResultChannelClosed = errors.New("ieResultReader closed")

func (i *ieResultChannel) addResult(ctx context.Context, result ieIteratorResult) error {
	__antithesis_instrumentation__.Notify(498014)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(498016)
		return ctx.Err()
	case <-i.doneCh:
		__antithesis_instrumentation__.Notify(498017)

		if ctxErr := ctx.Err(); ctxErr != nil {
			__antithesis_instrumentation__.Notify(498020)
			return ctxErr
		} else {
			__antithesis_instrumentation__.Notify(498021)
		}
		__antithesis_instrumentation__.Notify(498018)
		return errIEResultChannelClosed
	case i.dataCh <- result:
		__antithesis_instrumentation__.Notify(498019)
	}
	__antithesis_instrumentation__.Notify(498015)
	return i.maybeBlock(ctx)
}

func (i *ieResultChannel) maybeBlock(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(498022)
	if i.async() {
		__antithesis_instrumentation__.Notify(498024)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(498025)
	}
	__antithesis_instrumentation__.Notify(498023)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(498026)
		return ctx.Err()
	case <-i.doneCh:
		__antithesis_instrumentation__.Notify(498027)

		if ctxErr := ctx.Err(); ctxErr != nil {
			__antithesis_instrumentation__.Notify(498030)
			return ctxErr
		} else {
			__antithesis_instrumentation__.Notify(498031)
		}
		__antithesis_instrumentation__.Notify(498028)
		return errIEResultChannelClosed
	case <-i.waitCh:
		__antithesis_instrumentation__.Notify(498029)
		return nil
	}
}

func (i *ieResultChannel) finish() {
	__antithesis_instrumentation__.Notify(498032)
	close(i.dataCh)
}
