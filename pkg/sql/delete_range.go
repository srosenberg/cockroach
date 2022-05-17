package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type deleteRangeNode struct {
	spans roachpb.Spans

	desc catalog.TableDescriptor

	fetcher row.Fetcher

	autoCommitEnabled bool

	rowCount int
}

var _ planNode = &deleteRangeNode{}
var _ planNodeFastPath = &deleteRangeNode{}
var _ batchedPlanNode = &deleteRangeNode{}
var _ mutationPlanNode = &deleteRangeNode{}

func (d *deleteRangeNode) BatchedNext(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(465956)
	return false, nil
}

func (d *deleteRangeNode) BatchedCount() int {
	__antithesis_instrumentation__.Notify(465957)
	return d.rowCount
}

func (d *deleteRangeNode) BatchedValues(rowIdx int) tree.Datums {
	__antithesis_instrumentation__.Notify(465958)
	panic("invalid")
}

func (d *deleteRangeNode) FastPathResults() (int, bool) {
	__antithesis_instrumentation__.Notify(465959)
	return d.rowCount, true
}

func (d *deleteRangeNode) rowsWritten() int64 {
	__antithesis_instrumentation__.Notify(465960)
	return int64(d.rowCount)
}

func (d *deleteRangeNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(465961)
	if err := params.p.cancelChecker.Check(); err != nil {
		__antithesis_instrumentation__.Notify(465966)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465967)
	}
	__antithesis_instrumentation__.Notify(465962)

	var spec descpb.IndexFetchSpec
	if err := rowenc.InitIndexFetchSpec(
		&spec, params.ExecCfg().Codec, d.desc, d.desc.GetPrimaryIndex(), nil,
	); err != nil {
		__antithesis_instrumentation__.Notify(465968)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465969)
	}
	__antithesis_instrumentation__.Notify(465963)
	if err := d.fetcher.Init(
		params.ctx,
		false,
		descpb.ScanLockingStrength_FOR_NONE,
		descpb.ScanLockingWaitPolicy_BLOCK,
		0,
		params.p.alloc,
		nil,
		&spec,
	); err != nil {
		__antithesis_instrumentation__.Notify(465970)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465971)
	}
	__antithesis_instrumentation__.Notify(465964)

	ctx := params.ctx
	log.VEvent(ctx, 2, "fast delete: skipping scan")
	spans := make([]roachpb.Span, len(d.spans))
	copy(spans, d.spans)
	if !d.autoCommitEnabled {
		__antithesis_instrumentation__.Notify(465972)

		for len(spans) != 0 {
			__antithesis_instrumentation__.Notify(465973)
			b := params.p.txn.NewBatch()
			b.Header.MaxSpanRequestKeys = row.TableTruncateChunkSize
			b.Header.LockTimeout = params.SessionData().LockTimeout
			d.deleteSpans(params, b, spans)
			if err := params.p.txn.Run(ctx, b); err != nil {
				__antithesis_instrumentation__.Notify(465975)
				return row.ConvertBatchError(ctx, d.desc, b)
			} else {
				__antithesis_instrumentation__.Notify(465976)
			}
			__antithesis_instrumentation__.Notify(465974)

			spans = spans[:0]
			var err error
			if spans, err = d.processResults(b.Results, spans); err != nil {
				__antithesis_instrumentation__.Notify(465977)
				return err
			} else {
				__antithesis_instrumentation__.Notify(465978)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(465979)
		log.Event(ctx, "autocommit enabled")

		b := params.p.txn.NewBatch()
		b.Header.LockTimeout = params.SessionData().LockTimeout
		d.deleteSpans(params, b, spans)
		if err := params.p.txn.CommitInBatch(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(465981)
			return row.ConvertBatchError(ctx, d.desc, b)
		} else {
			__antithesis_instrumentation__.Notify(465982)
		}
		__antithesis_instrumentation__.Notify(465980)
		if resumeSpans, err := d.processResults(b.Results, nil); err != nil {
			__antithesis_instrumentation__.Notify(465983)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465984)
			if len(resumeSpans) != 0 {
				__antithesis_instrumentation__.Notify(465985)

				return errors.AssertionFailedf("deleteRange without a limit unexpectedly returned resumeSpans")
			} else {
				__antithesis_instrumentation__.Notify(465986)
			}
		}
	}
	__antithesis_instrumentation__.Notify(465965)

	params.ExecCfg().StatsRefresher.NotifyMutation(d.desc, d.rowCount)

	return nil
}

func (d *deleteRangeNode) deleteSpans(params runParams, b *kv.Batch, spans roachpb.Spans) {
	__antithesis_instrumentation__.Notify(465987)
	ctx := params.ctx
	traceKV := params.p.ExtendedEvalContext().Tracing.KVTracingEnabled()
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(465988)
		if traceKV {
			__antithesis_instrumentation__.Notify(465990)
			log.VEventf(ctx, 2, "DelRange %s - %s", span.Key, span.EndKey)
		} else {
			__antithesis_instrumentation__.Notify(465991)
		}
		__antithesis_instrumentation__.Notify(465989)
		b.DelRange(span.Key, span.EndKey, true)
	}
}

func (d *deleteRangeNode) processResults(
	results []kv.Result, resumeSpans []roachpb.Span,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(465992)
	for _, r := range results {
		__antithesis_instrumentation__.Notify(465994)
		var prev []byte
		for _, keyBytes := range r.Keys {
			__antithesis_instrumentation__.Notify(465996)

			if len(prev) > 0 && func() bool {
				__antithesis_instrumentation__.Notify(465999)
				return bytes.HasPrefix(keyBytes, prev) == true
			}() == true {
				__antithesis_instrumentation__.Notify(466000)
				continue
			} else {
				__antithesis_instrumentation__.Notify(466001)
			}
			__antithesis_instrumentation__.Notify(465997)

			after, _, err := d.fetcher.DecodeIndexKey(keyBytes)
			if err != nil {
				__antithesis_instrumentation__.Notify(466002)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(466003)
			}
			__antithesis_instrumentation__.Notify(465998)
			k := keyBytes[:len(keyBytes)-len(after)]
			if !bytes.Equal(k, prev) {
				__antithesis_instrumentation__.Notify(466004)
				prev = k
				d.rowCount++
			} else {
				__antithesis_instrumentation__.Notify(466005)
			}
		}
		__antithesis_instrumentation__.Notify(465995)
		if r.ResumeSpan != nil && func() bool {
			__antithesis_instrumentation__.Notify(466006)
			return r.ResumeSpan.Valid() == true
		}() == true {
			__antithesis_instrumentation__.Notify(466007)
			resumeSpans = append(resumeSpans, *r.ResumeSpan)
		} else {
			__antithesis_instrumentation__.Notify(466008)
		}
	}
	__antithesis_instrumentation__.Notify(465993)
	return resumeSpans, nil
}

func (*deleteRangeNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(466009)

	return false, nil
}

func (*deleteRangeNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(466010)
	panic("invalid")
}

func (*deleteRangeNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(466011) }
