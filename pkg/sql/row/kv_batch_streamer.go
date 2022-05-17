package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/kvstreamer"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func CanUseStreamer(ctx context.Context, settings *cluster.Settings) bool {
	__antithesis_instrumentation__.Notify(568333)

	return settings.Version.IsActive(ctx, clusterversion.ScanWholeRows) && func() bool {
		__antithesis_instrumentation__.Notify(568334)
		return useStreamerEnabled.Get(&settings.SV) == true
	}() == true
}

var useStreamerEnabled = settings.RegisterBoolSetting(
	settings.TenantReadOnly,
	"sql.distsql.use_streamer.enabled",
	"determines whether the usage of the Streamer API is allowed. "+
		"Enabling this will increase the speed of lookup/index joins "+
		"while adhering to memory limits.",
	false,
)

type TxnKVStreamer struct {
	streamer *kvstreamer.Streamer
	spans    roachpb.Spans

	getResponseScratch [1]roachpb.KeyValue

	results         []kvstreamer.Result
	lastResultState struct {
		kvstreamer.Result

		numEmitted int

		remainingBatches [][]byte
	}
}

var _ KVBatchFetcher = &TxnKVStreamer{}

func NewTxnKVStreamer(
	ctx context.Context,
	streamer *kvstreamer.Streamer,
	spans roachpb.Spans,
	lockStrength descpb.ScanLockingStrength,
) (*TxnKVStreamer, error) {
	__antithesis_instrumentation__.Notify(568335)
	if log.ExpensiveLogEnabled(ctx, 2) {
		__antithesis_instrumentation__.Notify(568338)
		log.VEventf(ctx, 2, "Scan %s", spans)
	} else {
		__antithesis_instrumentation__.Notify(568339)
	}
	__antithesis_instrumentation__.Notify(568336)
	keyLocking := getKeyLockingStrength(lockStrength)
	reqs := spansToRequests(spans, false, keyLocking)
	if err := streamer.Enqueue(ctx, reqs, nil); err != nil {
		__antithesis_instrumentation__.Notify(568340)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(568341)
	}
	__antithesis_instrumentation__.Notify(568337)
	return &TxnKVStreamer{
		streamer: streamer,
		spans:    spans,
	}, nil
}

func (f *TxnKVStreamer) proceedWithLastResult(
	ctx context.Context,
) (skip bool, kvs []roachpb.KeyValue, batchResp []byte, err error) {
	__antithesis_instrumentation__.Notify(568342)
	result := f.lastResultState.Result
	if get := result.GetResp; get != nil {
		__antithesis_instrumentation__.Notify(568346)

		if get.Value == nil {
			__antithesis_instrumentation__.Notify(568348)

			f.releaseLastResult(ctx)
			return true, nil, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(568349)
		}
		__antithesis_instrumentation__.Notify(568347)
		pos := result.EnqueueKeysSatisfied[f.lastResultState.numEmitted]
		origSpan := f.spans[pos]
		f.lastResultState.numEmitted++
		f.getResponseScratch[0] = roachpb.KeyValue{Key: origSpan.Key, Value: *get.Value}
		return false, f.getResponseScratch[:], nil, nil
	} else {
		__antithesis_instrumentation__.Notify(568350)
	}
	__antithesis_instrumentation__.Notify(568343)
	scan := result.ScanResp
	if len(scan.BatchResponses) > 0 {
		__antithesis_instrumentation__.Notify(568351)
		batchResp, f.lastResultState.remainingBatches = scan.BatchResponses[0], scan.BatchResponses[1:]
	} else {
		__antithesis_instrumentation__.Notify(568352)
	}
	__antithesis_instrumentation__.Notify(568344)
	if len(f.lastResultState.remainingBatches) == 0 {
		__antithesis_instrumentation__.Notify(568353)
		f.lastResultState.numEmitted++
	} else {
		__antithesis_instrumentation__.Notify(568354)
	}
	__antithesis_instrumentation__.Notify(568345)

	return false, nil, batchResp, nil
}

func (f *TxnKVStreamer) releaseLastResult(ctx context.Context) {
	__antithesis_instrumentation__.Notify(568355)
	f.lastResultState.Release(ctx)
	f.lastResultState.Result = kvstreamer.Result{}
}

func (f *TxnKVStreamer) nextBatch(
	ctx context.Context,
) (ok bool, kvs []roachpb.KeyValue, batchResp []byte, err error) {
	__antithesis_instrumentation__.Notify(568356)

	if len(f.lastResultState.remainingBatches) > 0 {
		__antithesis_instrumentation__.Notify(568362)
		batchResp, f.lastResultState.remainingBatches = f.lastResultState.remainingBatches[0], f.lastResultState.remainingBatches[1:]
		if len(f.lastResultState.remainingBatches) == 0 {
			__antithesis_instrumentation__.Notify(568364)
			f.lastResultState.numEmitted++
		} else {
			__antithesis_instrumentation__.Notify(568365)
		}
		__antithesis_instrumentation__.Notify(568363)
		return true, nil, batchResp, nil
	} else {
		__antithesis_instrumentation__.Notify(568366)
	}
	__antithesis_instrumentation__.Notify(568357)

	if f.lastResultState.numEmitted < len(f.lastResultState.EnqueueKeysSatisfied) {
		__antithesis_instrumentation__.Notify(568367)

		_, kvs, batchResp, err = f.proceedWithLastResult(ctx)
		return true, kvs, batchResp, err
	} else {
		__antithesis_instrumentation__.Notify(568368)
	}
	__antithesis_instrumentation__.Notify(568358)

	if f.lastResultState.numEmitted == len(f.lastResultState.EnqueueKeysSatisfied) && func() bool {
		__antithesis_instrumentation__.Notify(568369)
		return f.lastResultState.numEmitted > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(568370)
		f.releaseLastResult(ctx)
	} else {
		__antithesis_instrumentation__.Notify(568371)
	}
	__antithesis_instrumentation__.Notify(568359)

	for len(f.results) > 0 {
		__antithesis_instrumentation__.Notify(568372)

		f.lastResultState.Result = f.results[0]
		f.lastResultState.numEmitted = 0
		f.lastResultState.remainingBatches = nil

		f.results[0] = kvstreamer.Result{}
		f.results = f.results[1:]
		var skip bool
		skip, kvs, batchResp, err = f.proceedWithLastResult(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(568375)
			return false, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(568376)
		}
		__antithesis_instrumentation__.Notify(568373)
		if skip {
			__antithesis_instrumentation__.Notify(568377)
			continue
		} else {
			__antithesis_instrumentation__.Notify(568378)
		}
		__antithesis_instrumentation__.Notify(568374)
		return true, kvs, batchResp, err
	}
	__antithesis_instrumentation__.Notify(568360)

	f.results, err = f.streamer.GetResults(ctx)
	if len(f.results) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(568379)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(568380)
		return false, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(568381)
	}
	__antithesis_instrumentation__.Notify(568361)
	return f.nextBatch(ctx)
}

func (f *TxnKVStreamer) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(568382)
	f.lastResultState.Release(ctx)
	for _, r := range f.results {
		__antithesis_instrumentation__.Notify(568384)
		r.Release(ctx)
	}
	__antithesis_instrumentation__.Notify(568383)
	*f = TxnKVStreamer{}
}
