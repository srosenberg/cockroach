package idalloc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

type Incrementer func(_ context.Context, _ roachpb.Key, inc int64) (updated int64, _ error)

func DBIncrementer(
	db interface {
		Inc(ctx context.Context, key interface{}, value int64) (kv.KeyValue, error)
	},
) Incrementer {
	__antithesis_instrumentation__.Notify(101506)
	return func(ctx context.Context, key roachpb.Key, inc int64) (int64, error) {
		__antithesis_instrumentation__.Notify(101507)
		res, err := db.Inc(ctx, key, inc)
		if err != nil {
			__antithesis_instrumentation__.Notify(101509)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(101510)
		}
		__antithesis_instrumentation__.Notify(101508)
		return res.Value.GetInt()
	}
}

type Options struct {
	AmbientCtx  log.AmbientContext
	Key         roachpb.Key
	Incrementer Incrementer
	BlockSize   int64
	Stopper     *stop.Stopper
	Fatalf      func(context.Context, string, ...interface{})
}

type Allocator struct {
	log.AmbientContext
	opts Options

	ids  chan int64
	once sync.Once
}

func NewAllocator(opts Options) (*Allocator, error) {
	__antithesis_instrumentation__.Notify(101511)
	if opts.BlockSize == 0 {
		__antithesis_instrumentation__.Notify(101514)
		return nil, errors.Errorf("blockSize must be a positive integer: %d", opts.BlockSize)
	} else {
		__antithesis_instrumentation__.Notify(101515)
	}
	__antithesis_instrumentation__.Notify(101512)
	if opts.Fatalf == nil {
		__antithesis_instrumentation__.Notify(101516)
		opts.Fatalf = log.Fatalf
	} else {
		__antithesis_instrumentation__.Notify(101517)
	}
	__antithesis_instrumentation__.Notify(101513)
	opts.AmbientCtx.AddLogTag("idalloc", nil)
	return &Allocator{
		AmbientContext: opts.AmbientCtx,
		opts:           opts,
		ids:            make(chan int64, opts.BlockSize/2+1),
	}, nil
}

func (ia *Allocator) Allocate(ctx context.Context) (int64, error) {
	__antithesis_instrumentation__.Notify(101518)
	ia.once.Do(ia.start)

	select {
	case id := <-ia.ids:
		__antithesis_instrumentation__.Notify(101519)

		if id == 0 {
			__antithesis_instrumentation__.Notify(101522)
			return id, errors.Errorf("could not allocate ID; system is draining")
		} else {
			__antithesis_instrumentation__.Notify(101523)
		}
		__antithesis_instrumentation__.Notify(101520)
		return id, nil
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(101521)
		return 0, ctx.Err()
	}
}

func (ia *Allocator) start() {
	__antithesis_instrumentation__.Notify(101524)
	ctx := ia.AnnotateCtx(context.Background())
	if err := ia.opts.Stopper.RunAsyncTask(ctx, "id-alloc", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(101525)
		defer close(ia.ids)

		var prevValue int64
		for {
			__antithesis_instrumentation__.Notify(101526)
			var newValue int64
			var err error
			for r := retry.Start(base.DefaultRetryOptions()); r.Next(); {
				__antithesis_instrumentation__.Notify(101531)
				if stopperErr := ia.opts.Stopper.RunTask(ctx, "idalloc: allocating block",
					func(ctx context.Context) {
						__antithesis_instrumentation__.Notify(101534)
						newValue, err = ia.opts.Incrementer(ctx, ia.opts.Key, ia.opts.BlockSize)
					}); stopperErr != nil {
					__antithesis_instrumentation__.Notify(101535)
					return
				} else {
					__antithesis_instrumentation__.Notify(101536)
				}
				__antithesis_instrumentation__.Notify(101532)
				if err == nil {
					__antithesis_instrumentation__.Notify(101537)
					break
				} else {
					__antithesis_instrumentation__.Notify(101538)
				}
				__antithesis_instrumentation__.Notify(101533)

				log.Warningf(
					ctx,
					"unable to allocate %d ids from %s: %+v",
					ia.opts.BlockSize,
					ia.opts.Key,
					err,
				)
			}
			__antithesis_instrumentation__.Notify(101527)
			if err != nil {
				__antithesis_instrumentation__.Notify(101539)
				ia.opts.Fatalf(ctx, "unexpectedly exited id allocation retry loop: %s", err)
				return
			} else {
				__antithesis_instrumentation__.Notify(101540)
			}
			__antithesis_instrumentation__.Notify(101528)
			if prevValue != 0 && func() bool {
				__antithesis_instrumentation__.Notify(101541)
				return newValue < prevValue+ia.opts.BlockSize == true
			}() == true {
				__antithesis_instrumentation__.Notify(101542)
				ia.opts.Fatalf(
					ctx,
					"counter corrupt: incremented to %d, expected at least %d + %d",
					newValue, prevValue, ia.opts.BlockSize,
				)
				return
			} else {
				__antithesis_instrumentation__.Notify(101543)
			}
			__antithesis_instrumentation__.Notify(101529)

			end := newValue + 1
			start := end - ia.opts.BlockSize
			if start <= 0 {
				__antithesis_instrumentation__.Notify(101544)
				ia.opts.Fatalf(ctx, "allocator initialized with negative key")
				return
			} else {
				__antithesis_instrumentation__.Notify(101545)
			}
			__antithesis_instrumentation__.Notify(101530)
			prevValue = newValue

			for i := start; i < end; i++ {
				__antithesis_instrumentation__.Notify(101546)
				select {
				case ia.ids <- i:
					__antithesis_instrumentation__.Notify(101547)
				case <-ia.opts.Stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(101548)
					return
				}
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(101549)
		close(ia.ids)
	} else {
		__antithesis_instrumentation__.Notify(101550)
	}
}
