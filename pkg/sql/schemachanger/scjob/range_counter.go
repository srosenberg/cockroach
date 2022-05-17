package scjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps"
)

type rangeCounter struct {
	db  *kv.DB
	dsp *sql.DistSQLPlanner
}

func NewRangeCounter(db *kv.DB, dsp *sql.DistSQLPlanner) scdeps.RangeCounter {
	__antithesis_instrumentation__.Notify(582379)
	return &rangeCounter{
		db:  db,
		dsp: dsp,
	}
}

var _ scdeps.RangeCounter = (*rangeCounter)(nil)

func (r rangeCounter) NumRangesInSpanContainedBy(
	ctx context.Context, span roachpb.Span, containedBy []roachpb.Span,
) (total, inContainedBy int, _ error) {
	__antithesis_instrumentation__.Notify(582380)
	return sql.NumRangesInSpanContainedBy(ctx, r.db, r.dsp, span, containedBy)
}
