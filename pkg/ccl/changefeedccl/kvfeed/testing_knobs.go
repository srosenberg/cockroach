package kvfeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type TestingKnobs struct {
	BeforeScanRequest func(b *kv.Batch) error
	OnRangeFeedValue  func(kv roachpb.KeyValue) error
}

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(17479) }
