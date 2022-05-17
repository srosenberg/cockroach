package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/limit"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type dbAdapter struct {
	db         *kv.DB
	st         *cluster.Settings
	distSender *kvcoord.DistSender
}

var _ DB = (*dbAdapter)(nil)

var maxScanParallelism = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.rangefeed.max_scan_parallelism",
	"maximum number of concurrent scan requests that can be issued during initial scan",
	64,
)

func newDBAdapter(db *kv.DB, st *cluster.Settings) (*dbAdapter, error) {
	__antithesis_instrumentation__.Notify(89671)
	var distSender *kvcoord.DistSender
	{
		__antithesis_instrumentation__.Notify(89673)
		txnWrapperSender, ok := db.NonTransactionalSender().(*kv.CrossRangeTxnWrapperSender)
		if !ok {
			__antithesis_instrumentation__.Notify(89675)
			return nil, errors.Errorf("failed to extract a %T from %T",
				(*kv.CrossRangeTxnWrapperSender)(nil), db.NonTransactionalSender())
		} else {
			__antithesis_instrumentation__.Notify(89676)
		}
		__antithesis_instrumentation__.Notify(89674)
		distSender, ok = txnWrapperSender.Wrapped().(*kvcoord.DistSender)
		if !ok {
			__antithesis_instrumentation__.Notify(89677)
			return nil, errors.Errorf("failed to extract a %T from %T",
				(*kvcoord.DistSender)(nil), txnWrapperSender.Wrapped())
		} else {
			__antithesis_instrumentation__.Notify(89678)
		}
	}
	__antithesis_instrumentation__.Notify(89672)
	return &dbAdapter{
		db:         db,
		st:         st,
		distSender: distSender,
	}, nil
}

func (dbc *dbAdapter) RangeFeed(
	ctx context.Context,
	spans []roachpb.Span,
	startFrom hlc.Timestamp,
	withDiff bool,
	eventC chan<- *roachpb.RangeFeedEvent,
) error {
	__antithesis_instrumentation__.Notify(89679)
	return dbc.distSender.RangeFeed(ctx, spans, startFrom, withDiff, eventC)
}

type concurrentBoundAccount struct {
	syncutil.Mutex
	*mon.BoundAccount
}

func (ba *concurrentBoundAccount) Grow(ctx context.Context, x int64) error {
	__antithesis_instrumentation__.Notify(89680)
	ba.Lock()
	defer ba.Unlock()
	return ba.BoundAccount.Grow(ctx, x)
}

func (ba *concurrentBoundAccount) Shrink(ctx context.Context, x int64) {
	__antithesis_instrumentation__.Notify(89681)
	ba.Lock()
	defer ba.Unlock()
	ba.BoundAccount.Shrink(ctx, x)
}

func (dbc *dbAdapter) Scan(
	ctx context.Context,
	spans []roachpb.Span,
	asOf hlc.Timestamp,
	rowFn func(value roachpb.KeyValue),
	cfg scanConfig,
) error {
	__antithesis_instrumentation__.Notify(89682)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(89688)
		return errors.AssertionFailedf("expected at least 1 span, got none")
	} else {
		__antithesis_instrumentation__.Notify(89689)
	}
	__antithesis_instrumentation__.Notify(89683)

	var acc *concurrentBoundAccount
	if cfg.mon != nil {
		__antithesis_instrumentation__.Notify(89690)
		ba := cfg.mon.MakeBoundAccount()
		defer ba.Close(ctx)
		acc = &concurrentBoundAccount{BoundAccount: &ba}
	} else {
		__antithesis_instrumentation__.Notify(89691)
	}
	__antithesis_instrumentation__.Notify(89684)

	if cfg.scanParallelism == nil {
		__antithesis_instrumentation__.Notify(89692)
		for _, sp := range spans {
			__antithesis_instrumentation__.Notify(89694)
			if err := dbc.scanSpan(ctx, sp, asOf, rowFn, cfg.targetScanBytes, cfg.onSpanDone, acc); err != nil {
				__antithesis_instrumentation__.Notify(89695)
				return err
			} else {
				__antithesis_instrumentation__.Notify(89696)
			}
		}
		__antithesis_instrumentation__.Notify(89693)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(89697)
	}
	__antithesis_instrumentation__.Notify(89685)

	parallelismFn := cfg.scanParallelism
	if parallelismFn == nil {
		__antithesis_instrumentation__.Notify(89698)
		parallelismFn = func() int { __antithesis_instrumentation__.Notify(89699); return 1 }
	} else {
		__antithesis_instrumentation__.Notify(89700)
		highParallelism := log.Every(30 * time.Second)
		userSuppliedFn := parallelismFn
		parallelismFn = func() int {
			__antithesis_instrumentation__.Notify(89701)
			p := userSuppliedFn()
			if p < 1 {
				__antithesis_instrumentation__.Notify(89704)
				p = 1
			} else {
				__antithesis_instrumentation__.Notify(89705)
			}
			__antithesis_instrumentation__.Notify(89702)
			maxP := int(maxScanParallelism.Get(&dbc.st.SV))
			if p > maxP {
				__antithesis_instrumentation__.Notify(89706)
				if highParallelism.ShouldLog() {
					__antithesis_instrumentation__.Notify(89708)
					log.Warningf(ctx,
						"high scan parallelism %d limited via 'kv.rangefeed.max_scan_parallelism' to %d", p, maxP)
				} else {
					__antithesis_instrumentation__.Notify(89709)
				}
				__antithesis_instrumentation__.Notify(89707)
				p = maxP
			} else {
				__antithesis_instrumentation__.Notify(89710)
			}
			__antithesis_instrumentation__.Notify(89703)
			return p
		}
	}
	__antithesis_instrumentation__.Notify(89686)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g := ctxgroup.WithContext(ctx)
	err := dbc.divideAndSendScanRequests(
		ctx, &g, spans, asOf, rowFn,
		parallelismFn, cfg.targetScanBytes, cfg.onSpanDone, acc)
	if err != nil {
		__antithesis_instrumentation__.Notify(89711)
		cancel()
	} else {
		__antithesis_instrumentation__.Notify(89712)
	}
	__antithesis_instrumentation__.Notify(89687)
	return errors.CombineErrors(err, g.Wait())
}

func (dbc *dbAdapter) scanSpan(
	ctx context.Context,
	span roachpb.Span,
	asOf hlc.Timestamp,
	rowFn func(value roachpb.KeyValue),
	targetScanBytes int64,
	onScanDone OnScanCompleted,
	acc *concurrentBoundAccount,
) error {
	__antithesis_instrumentation__.Notify(89713)
	if acc != nil {
		__antithesis_instrumentation__.Notify(89715)
		if err := acc.Grow(ctx, targetScanBytes); err != nil {
			__antithesis_instrumentation__.Notify(89717)
			return err
		} else {
			__antithesis_instrumentation__.Notify(89718)
		}
		__antithesis_instrumentation__.Notify(89716)
		defer acc.Shrink(ctx, targetScanBytes)
	} else {
		__antithesis_instrumentation__.Notify(89719)
	}
	__antithesis_instrumentation__.Notify(89714)

	return dbc.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(89720)
		if err := txn.SetFixedTimestamp(ctx, asOf); err != nil {
			__antithesis_instrumentation__.Notify(89722)
			return err
		} else {
			__antithesis_instrumentation__.Notify(89723)
		}
		__antithesis_instrumentation__.Notify(89721)
		sp := span
		var b kv.Batch
		for {
			__antithesis_instrumentation__.Notify(89724)
			b.Header.TargetBytes = targetScanBytes
			b.Scan(sp.Key, sp.EndKey)
			if err := txn.Run(ctx, &b); err != nil {
				__antithesis_instrumentation__.Notify(89729)
				return err
			} else {
				__antithesis_instrumentation__.Notify(89730)
			}
			__antithesis_instrumentation__.Notify(89725)
			res := b.Results[0]
			for _, row := range res.Rows {
				__antithesis_instrumentation__.Notify(89731)
				rowFn(roachpb.KeyValue{Key: row.Key, Value: *row.Value})
			}
			__antithesis_instrumentation__.Notify(89726)
			if res.ResumeSpan == nil {
				__antithesis_instrumentation__.Notify(89732)
				if onScanDone != nil {
					__antithesis_instrumentation__.Notify(89734)
					return onScanDone(ctx, sp)
				} else {
					__antithesis_instrumentation__.Notify(89735)
				}
				__antithesis_instrumentation__.Notify(89733)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(89736)
			}
			__antithesis_instrumentation__.Notify(89727)

			if onScanDone != nil {
				__antithesis_instrumentation__.Notify(89737)
				if err := onScanDone(ctx, roachpb.Span{Key: sp.Key, EndKey: res.ResumeSpan.Key}); err != nil {
					__antithesis_instrumentation__.Notify(89738)
					return err
				} else {
					__antithesis_instrumentation__.Notify(89739)
				}
			} else {
				__antithesis_instrumentation__.Notify(89740)
			}
			__antithesis_instrumentation__.Notify(89728)

			sp = res.ResumeSpanAsValue()
			b = kv.Batch{}
		}
	})
}

func (dbc *dbAdapter) divideAndSendScanRequests(
	ctx context.Context,
	workGroup *ctxgroup.Group,
	spans []roachpb.Span,
	asOf hlc.Timestamp,
	rowFn func(value roachpb.KeyValue),
	parallelismFn func() int,
	targetScanBytes int64,
	onSpanDone OnScanCompleted,
	acc *concurrentBoundAccount,
) error {
	__antithesis_instrumentation__.Notify(89741)

	var sg roachpb.SpanGroup
	sg.Add(spans...)

	currentScanLimit := parallelismFn()
	exportLim := limit.MakeConcurrentRequestLimiter("rangefeedScanLimiter", parallelismFn())
	ri := kvcoord.MakeRangeIterator(dbc.distSender)

	for _, sp := range sg.Slice() {
		__antithesis_instrumentation__.Notify(89743)
		nextRS, err := keys.SpanAddr(sp)
		if err != nil {
			__antithesis_instrumentation__.Notify(89746)
			return err
		} else {
			__antithesis_instrumentation__.Notify(89747)
		}
		__antithesis_instrumentation__.Notify(89744)

		for ri.Seek(ctx, nextRS.Key, kvcoord.Ascending); ri.Valid(); ri.Next(ctx) {
			__antithesis_instrumentation__.Notify(89748)
			desc := ri.Desc()
			partialRS, err := nextRS.Intersect(desc)
			if err != nil {
				__antithesis_instrumentation__.Notify(89753)
				return err
			} else {
				__antithesis_instrumentation__.Notify(89754)
			}
			__antithesis_instrumentation__.Notify(89749)
			nextRS.Key = partialRS.EndKey

			if newLimit := parallelismFn(); newLimit != currentScanLimit {
				__antithesis_instrumentation__.Notify(89755)
				currentScanLimit = newLimit
				exportLim.SetLimit(newLimit)
			} else {
				__antithesis_instrumentation__.Notify(89756)
			}
			__antithesis_instrumentation__.Notify(89750)

			limAlloc, err := exportLim.Begin(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(89757)
				return err
			} else {
				__antithesis_instrumentation__.Notify(89758)
			}
			__antithesis_instrumentation__.Notify(89751)

			sp := partialRS.AsRawSpanWithNoLocals()
			workGroup.GoCtx(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(89759)
				defer limAlloc.Release()
				return dbc.scanSpan(ctx, sp, asOf, rowFn, targetScanBytes, onSpanDone, acc)
			})
			__antithesis_instrumentation__.Notify(89752)

			if !ri.NeedAnother(nextRS) {
				__antithesis_instrumentation__.Notify(89760)
				break
			} else {
				__antithesis_instrumentation__.Notify(89761)
			}
		}
		__antithesis_instrumentation__.Notify(89745)
		if err := ri.Error(); err != nil {
			__antithesis_instrumentation__.Notify(89762)
			return ri.Error()
		} else {
			__antithesis_instrumentation__.Notify(89763)
		}
	}
	__antithesis_instrumentation__.Notify(89742)

	return nil
}
