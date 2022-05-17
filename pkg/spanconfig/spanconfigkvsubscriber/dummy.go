package spanconfigkvsubscriber

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type noopKVSubscriber struct {
	clock *hlc.Clock
}

var _ spanconfig.KVSubscriber = &noopKVSubscriber{}

func NewNoopSubscriber(clock *hlc.Clock) spanconfig.KVSubscriber {
	__antithesis_instrumentation__.Notify(240586)
	return &noopKVSubscriber{
		clock: clock,
	}
}

func (n *noopKVSubscriber) Subscribe(func(context.Context, roachpb.Span)) {
	__antithesis_instrumentation__.Notify(240587)
}

func (n *noopKVSubscriber) LastUpdated() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(240588)
	return n.clock.Now()
}

func (n *noopKVSubscriber) NeedsSplit(context.Context, roachpb.RKey, roachpb.RKey) bool {
	__antithesis_instrumentation__.Notify(240589)
	return false
}

func (n *noopKVSubscriber) ComputeSplitKey(
	context.Context, roachpb.RKey, roachpb.RKey,
) roachpb.RKey {
	__antithesis_instrumentation__.Notify(240590)
	return roachpb.RKey{}
}

func (n *noopKVSubscriber) GetSpanConfigForKey(
	context.Context, roachpb.RKey,
) (roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(240591)
	return roachpb.SpanConfig{}, nil
}

func (n *noopKVSubscriber) GetProtectionTimestamps(
	context.Context, roachpb.Span,
) ([]hlc.Timestamp, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(240592)
	return nil, n.LastUpdated(), nil
}
