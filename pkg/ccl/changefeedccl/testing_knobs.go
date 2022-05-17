package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeeddist"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvfeed"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

type TestingKnobs struct {
	BeforeEmitRow func(context.Context) error

	MemMonitor *mon.BytesMonitor

	HandleDistChangefeedError func(error) error

	WrapSink func(s Sink, jobID jobspb.JobID) Sink

	ShouldSkipResolved func(resolved *jobspb.ResolvedSpan) bool

	FeedKnobs     kvfeed.TestingKnobs
	DistflowKnobs changefeeddist.TestingKnobs
}

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(18994) }
