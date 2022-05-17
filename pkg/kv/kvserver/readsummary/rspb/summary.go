package rspb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func FromTimestamp(ts hlc.Timestamp) ReadSummary {
	__antithesis_instrumentation__.Notify(114268)
	seg := Segment{LowWater: ts}
	return ReadSummary{
		Local:  seg,
		Global: seg,
	}
}

func (c ReadSummary) Clone() *ReadSummary {
	__antithesis_instrumentation__.Notify(114269)

	return &c
}

func (c *ReadSummary) Merge(o ReadSummary) {
	__antithesis_instrumentation__.Notify(114270)
	c.Local.merge(o.Local)
	c.Global.merge(o.Global)
}

func (c *Segment) merge(o Segment) {
	__antithesis_instrumentation__.Notify(114271)
	c.LowWater.Forward(o.LowWater)
}

func (c *ReadSummary) AssertNoRegression(ctx context.Context, o ReadSummary) {
	__antithesis_instrumentation__.Notify(114272)
	c.Local.assertNoRegression(ctx, o.Local, "local")
	c.Global.assertNoRegression(ctx, o.Global, "global")
}

func (c *Segment) assertNoRegression(ctx context.Context, o Segment, name string) {
	__antithesis_instrumentation__.Notify(114273)
	if c.LowWater.Less(o.LowWater) {
		__antithesis_instrumentation__.Notify(114274)
		log.Fatalf(ctx, "read summary regression in %s segment, was %s, now %s",
			name, o.LowWater, c.LowWater)
	} else {
		__antithesis_instrumentation__.Notify(114275)
	}
}

var _ = (*ReadSummary).AssertNoRegression
