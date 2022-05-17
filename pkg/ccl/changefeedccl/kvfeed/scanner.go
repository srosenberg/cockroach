package kvfeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/covering"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/limit"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type kvScanner interface {
	Scan(ctx context.Context, sink kvevent.Writer, cfg physicalConfig) error
}

type scanRequestScanner struct {
	settings                *cluster.Settings
	gossip                  gossip.OptionalGossip
	db                      *kv.DB
	onBackfillRangeCallback func(int64) (func(), func())
}

var _ kvScanner = (*scanRequestScanner)(nil)

func (p *scanRequestScanner) Scan(
	ctx context.Context, sink kvevent.Writer, cfg physicalConfig,
) error {
	__antithesis_instrumentation__.Notify(17373)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if log.V(2) {
		__antithesis_instrumentation__.Notify(17378)
		log.Infof(ctx, "performing scan on %v at %v withDiff %v",
			cfg.Spans, cfg.Timestamp, cfg.WithDiff)
	} else {
		__antithesis_instrumentation__.Notify(17379)
	}
	__antithesis_instrumentation__.Notify(17374)

	sender := p.db.NonTransactionalSender()
	distSender := sender.(*kv.CrossRangeTxnWrapperSender).Wrapped().(*kvcoord.DistSender)
	spans, err := getSpansToProcess(ctx, distSender, cfg.Spans)
	if err != nil {
		__antithesis_instrumentation__.Notify(17380)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17381)
	}
	__antithesis_instrumentation__.Notify(17375)

	var backfillDec, backfillClear func()
	if p.onBackfillRangeCallback != nil {
		__antithesis_instrumentation__.Notify(17382)
		backfillDec, backfillClear = p.onBackfillRangeCallback(int64(len(spans)))
		defer backfillClear()
	} else {
		__antithesis_instrumentation__.Notify(17383)
	}
	__antithesis_instrumentation__.Notify(17376)

	maxConcurrentScans := maxConcurrentScanRequests(p.gossip, &p.settings.SV)
	exportLim := limit.MakeConcurrentRequestLimiter("changefeedScanRequestLimiter", maxConcurrentScans)

	lastScanLimitUserSetting := changefeedbase.ScanRequestLimit.Get(&p.settings.SV)

	g := ctxgroup.WithContext(ctx)

	var atomicFinished int64
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(17384)
		span := span

		if currentUserScanLimit := changefeedbase.ScanRequestLimit.Get(&p.settings.SV); currentUserScanLimit != lastScanLimitUserSetting {
			__antithesis_instrumentation__.Notify(17387)
			lastScanLimitUserSetting = currentUserScanLimit
			exportLim.SetLimit(maxConcurrentScanRequests(p.gossip, &p.settings.SV))
		} else {
			__antithesis_instrumentation__.Notify(17388)
		}
		__antithesis_instrumentation__.Notify(17385)

		limAlloc, err := exportLim.Begin(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(17389)
			cancel()
			return errors.CombineErrors(err, g.Wait())
		} else {
			__antithesis_instrumentation__.Notify(17390)
		}
		__antithesis_instrumentation__.Notify(17386)

		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(17391)
			defer limAlloc.Release()
			err := p.exportSpan(ctx, span, cfg.Timestamp, cfg.WithDiff, sink, cfg.Knobs)
			finished := atomic.AddInt64(&atomicFinished, 1)
			if backfillDec != nil {
				__antithesis_instrumentation__.Notify(17394)
				backfillDec()
			} else {
				__antithesis_instrumentation__.Notify(17395)
			}
			__antithesis_instrumentation__.Notify(17392)
			if log.V(2) {
				__antithesis_instrumentation__.Notify(17396)
				log.Infof(ctx, `exported %d of %d: %v`, finished, len(spans), err)
			} else {
				__antithesis_instrumentation__.Notify(17397)
			}
			__antithesis_instrumentation__.Notify(17393)
			return err
		})
	}
	__antithesis_instrumentation__.Notify(17377)
	return g.Wait()
}

func (p *scanRequestScanner) exportSpan(
	ctx context.Context,
	span roachpb.Span,
	ts hlc.Timestamp,
	withDiff bool,
	sink kvevent.Writer,
	knobs TestingKnobs,
) error {
	__antithesis_instrumentation__.Notify(17398)
	txn := p.db.NewTxn(ctx, "changefeed backfill")
	if log.V(2) {
		__antithesis_instrumentation__.Notify(17404)
		log.Infof(ctx, `sending ScanRequest %s at %s`, span, ts)
	} else {
		__antithesis_instrumentation__.Notify(17405)
	}
	__antithesis_instrumentation__.Notify(17399)
	if err := txn.SetFixedTimestamp(ctx, ts); err != nil {
		__antithesis_instrumentation__.Notify(17406)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17407)
	}
	__antithesis_instrumentation__.Notify(17400)
	stopwatchStart := timeutil.Now()
	var scanDuration, bufferDuration time.Duration
	targetBytesPerScan := changefeedbase.ScanRequestSize.Get(&p.settings.SV)
	for remaining := &span; remaining != nil; {
		__antithesis_instrumentation__.Notify(17408)
		start := timeutil.Now()
		b := txn.NewBatch()
		r := roachpb.NewScan(remaining.Key, remaining.EndKey, false).(*roachpb.ScanRequest)
		r.ScanFormat = roachpb.BATCH_RESPONSE
		b.Header.TargetBytes = targetBytesPerScan

		b.AddRawRequest(r)
		if knobs.BeforeScanRequest != nil {
			__antithesis_instrumentation__.Notify(17413)
			if err := knobs.BeforeScanRequest(b); err != nil {
				__antithesis_instrumentation__.Notify(17414)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17415)
			}
		} else {
			__antithesis_instrumentation__.Notify(17416)
		}
		__antithesis_instrumentation__.Notify(17409)

		if err := txn.Run(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(17417)
			return errors.Wrapf(err, `fetching changes for %s`, span)
		} else {
			__antithesis_instrumentation__.Notify(17418)
		}
		__antithesis_instrumentation__.Notify(17410)
		afterScan := timeutil.Now()
		res := b.RawResponse().Responses[0].GetScan()
		if err := slurpScanResponse(ctx, sink, res, ts, withDiff, *remaining); err != nil {
			__antithesis_instrumentation__.Notify(17419)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17420)
		}
		__antithesis_instrumentation__.Notify(17411)
		afterBuffer := timeutil.Now()
		scanDuration += afterScan.Sub(start)
		bufferDuration += afterBuffer.Sub(afterScan)
		if res.ResumeSpan != nil {
			__antithesis_instrumentation__.Notify(17421)
			consumed := roachpb.Span{Key: remaining.Key, EndKey: res.ResumeSpan.Key}
			if err := sink.Add(
				ctx, kvevent.MakeResolvedEvent(consumed, ts, jobspb.ResolvedSpan_NONE),
			); err != nil {
				__antithesis_instrumentation__.Notify(17422)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17423)
			}
		} else {
			__antithesis_instrumentation__.Notify(17424)
		}
		__antithesis_instrumentation__.Notify(17412)
		remaining = res.ResumeSpan
	}
	__antithesis_instrumentation__.Notify(17401)

	if err := sink.Add(
		ctx, kvevent.MakeResolvedEvent(span, ts, jobspb.ResolvedSpan_NONE),
	); err != nil {
		__antithesis_instrumentation__.Notify(17425)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17426)
	}
	__antithesis_instrumentation__.Notify(17402)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(17427)
		log.Infof(ctx, `finished Scan of %s at %s took %s`,
			span, ts.AsOfSystemTime(), timeutil.Since(stopwatchStart))
	} else {
		__antithesis_instrumentation__.Notify(17428)
	}
	__antithesis_instrumentation__.Notify(17403)
	return nil
}

func getSpansToProcess(
	ctx context.Context, ds *kvcoord.DistSender, targetSpans []roachpb.Span,
) ([]roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(17429)
	ranges, err := allRangeSpans(ctx, ds, targetSpans)
	if err != nil {
		__antithesis_instrumentation__.Notify(17434)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(17435)
	}
	__antithesis_instrumentation__.Notify(17430)

	type spanMarker struct{}
	type rangeMarker struct{}

	var spanCovering covering.Covering
	for _, span := range targetSpans {
		__antithesis_instrumentation__.Notify(17436)
		spanCovering = append(spanCovering, covering.Range{
			Start:   []byte(span.Key),
			End:     []byte(span.EndKey),
			Payload: spanMarker{},
		})
	}
	__antithesis_instrumentation__.Notify(17431)

	var rangeCovering covering.Covering
	for _, r := range ranges {
		__antithesis_instrumentation__.Notify(17437)
		rangeCovering = append(rangeCovering, covering.Range{
			Start:   []byte(r.Key),
			End:     []byte(r.EndKey),
			Payload: rangeMarker{},
		})
	}
	__antithesis_instrumentation__.Notify(17432)

	chunks := covering.OverlapCoveringMerge(
		[]covering.Covering{spanCovering, rangeCovering},
	)

	var requests []roachpb.Span
	for _, chunk := range chunks {
		__antithesis_instrumentation__.Notify(17438)
		if _, ok := chunk.Payload.([]interface{})[0].(spanMarker); !ok {
			__antithesis_instrumentation__.Notify(17440)
			continue
		} else {
			__antithesis_instrumentation__.Notify(17441)
		}
		__antithesis_instrumentation__.Notify(17439)
		requests = append(requests, roachpb.Span{Key: chunk.Start, EndKey: chunk.End})
	}
	__antithesis_instrumentation__.Notify(17433)
	return requests, nil
}

func slurpScanResponse(
	ctx context.Context,
	sink kvevent.Writer,
	res *roachpb.ScanResponse,
	ts hlc.Timestamp,
	withDiff bool,
	span roachpb.Span,
) error {
	__antithesis_instrumentation__.Notify(17442)
	for _, br := range res.BatchResponses {
		__antithesis_instrumentation__.Notify(17444)
		for len(br) > 0 {
			__antithesis_instrumentation__.Notify(17445)
			var kv roachpb.KeyValue
			var err error
			kv.Key, kv.Value.Timestamp, kv.Value.RawBytes, br, err = enginepb.ScanDecodeKeyValue(br)
			if err != nil {
				__antithesis_instrumentation__.Notify(17448)
				return errors.Wrapf(err, `decoding changes for %s`, span)
			} else {
				__antithesis_instrumentation__.Notify(17449)
			}
			__antithesis_instrumentation__.Notify(17446)
			var prevVal roachpb.Value
			if withDiff {
				__antithesis_instrumentation__.Notify(17450)

				prevVal = kv.Value
			} else {
				__antithesis_instrumentation__.Notify(17451)
			}
			__antithesis_instrumentation__.Notify(17447)
			if err = sink.Add(ctx, kvevent.MakeKVEvent(kv, prevVal, ts)); err != nil {
				__antithesis_instrumentation__.Notify(17452)
				return errors.Wrapf(err, `buffering changes for %s`, span)
			} else {
				__antithesis_instrumentation__.Notify(17453)
			}
		}
	}
	__antithesis_instrumentation__.Notify(17443)
	return nil
}

func allRangeSpans(
	ctx context.Context, ds *kvcoord.DistSender, spans []roachpb.Span,
) ([]roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(17454)

	ranges := make([]roachpb.Span, 0, len(spans))

	it := kvcoord.MakeRangeIterator(ds)

	for i := range spans {
		__antithesis_instrumentation__.Notify(17456)
		rSpan, err := keys.SpanAddr(spans[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(17458)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(17459)
		}
		__antithesis_instrumentation__.Notify(17457)
		for it.Seek(ctx, rSpan.Key, kvcoord.Ascending); ; it.Next(ctx) {
			__antithesis_instrumentation__.Notify(17460)
			if !it.Valid() {
				__antithesis_instrumentation__.Notify(17462)
				return nil, it.Error()
			} else {
				__antithesis_instrumentation__.Notify(17463)
			}
			__antithesis_instrumentation__.Notify(17461)
			ranges = append(ranges, roachpb.Span{
				Key: it.Desc().StartKey.AsRawKey(), EndKey: it.Desc().EndKey.AsRawKey(),
			})
			if !it.NeedAnother(rSpan) {
				__antithesis_instrumentation__.Notify(17464)
				break
			} else {
				__antithesis_instrumentation__.Notify(17465)
			}
		}
	}
	__antithesis_instrumentation__.Notify(17455)

	return ranges, nil
}

func clusterNodeCount(gw gossip.OptionalGossip) int {
	__antithesis_instrumentation__.Notify(17466)
	g, err := gw.OptionalErr(47971)
	if err != nil {
		__antithesis_instrumentation__.Notify(17469)

		return 1
	} else {
		__antithesis_instrumentation__.Notify(17470)
	}
	__antithesis_instrumentation__.Notify(17467)
	var nodes int
	_ = g.IterateInfos(gossip.KeyNodeIDPrefix, func(_ string, _ gossip.Info) error {
		__antithesis_instrumentation__.Notify(17471)
		nodes++
		return nil
	})
	__antithesis_instrumentation__.Notify(17468)
	return nodes
}

func maxConcurrentScanRequests(gw gossip.OptionalGossip, sv *settings.Values) int {
	__antithesis_instrumentation__.Notify(17472)

	if max := changefeedbase.ScanRequestLimit.Get(sv); max > 0 {
		__antithesis_instrumentation__.Notify(17475)
		return int(max)
	} else {
		__antithesis_instrumentation__.Notify(17476)
	}
	__antithesis_instrumentation__.Notify(17473)

	nodes := clusterNodeCount(gw)

	max := 3 * nodes
	if max > 100 {
		__antithesis_instrumentation__.Notify(17477)
		max = 100
	} else {
		__antithesis_instrumentation__.Notify(17478)
	}
	__antithesis_instrumentation__.Notify(17474)
	return max
}
