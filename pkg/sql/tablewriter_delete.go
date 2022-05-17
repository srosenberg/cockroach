package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type tableDeleter struct {
	tableWriterBase

	rd    row.Deleter
	alloc *tree.DatumAlloc
}

var _ tableWriter = &tableDeleter{}

func (*tableDeleter) desc() string { __antithesis_instrumentation__.Notify(627748); return "deleter" }

func (td *tableDeleter) walkExprs(_ func(desc string, index int, expr tree.TypedExpr)) {
	__antithesis_instrumentation__.Notify(627749)
}

func (td *tableDeleter) init(
	_ context.Context, txn *kv.Txn, evalCtx *tree.EvalContext, sv *settings.Values,
) error {
	td.tableWriterBase.init(txn, td.tableDesc(), evalCtx, sv)
	return nil
}

func (td *tableDeleter) row(
	ctx context.Context, values tree.Datums, pm row.PartialIndexUpdateHelper, traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(627750)
	td.currentBatchSize++
	return td.rd.DeleteRow(ctx, td.b, values, pm, traceKV)
}

func (td *tableDeleter) deleteIndex(
	ctx context.Context, idx catalog.Index, resume roachpb.Span, limit int64, traceKV bool,
) (roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(627751)
	if resume.Key == nil {
		__antithesis_instrumentation__.Notify(627756)
		resume = td.tableDesc().IndexSpan(td.rd.Helper.Codec, idx.GetID())
	} else {
		__antithesis_instrumentation__.Notify(627757)
	}
	__antithesis_instrumentation__.Notify(627752)

	if traceKV {
		__antithesis_instrumentation__.Notify(627758)
		log.VEventf(ctx, 2, "DelRange %s - %s", resume.Key, resume.EndKey)
	} else {
		__antithesis_instrumentation__.Notify(627759)
	}
	__antithesis_instrumentation__.Notify(627753)
	td.b.DelRange(resume.Key, resume.EndKey, false)
	td.b.Header.MaxSpanRequestKeys = limit
	if err := td.finalize(ctx); err != nil {
		__antithesis_instrumentation__.Notify(627760)
		return resume, err
	} else {
		__antithesis_instrumentation__.Notify(627761)
	}
	__antithesis_instrumentation__.Notify(627754)
	if l := len(td.b.Results); l != 1 {
		__antithesis_instrumentation__.Notify(627762)
		panic(errors.AssertionFailedf("%d results returned, expected 1", l))
	} else {
		__antithesis_instrumentation__.Notify(627763)
	}
	__antithesis_instrumentation__.Notify(627755)
	return td.b.Results[0].ResumeSpanAsValue(), nil
}

func (td *tableDeleter) tableDesc() catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(627764)
	return td.rd.Helper.TableDesc
}
