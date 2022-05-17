package kvevent

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
)

type errorWrapperEventBuffer struct {
	Buffer
}

func NewErrorWrapperEventBuffer(b Buffer) Buffer {
	__antithesis_instrumentation__.Notify(17106)
	return &errorWrapperEventBuffer{b}
}

func (e errorWrapperEventBuffer) Add(ctx context.Context, event Event) error {
	__antithesis_instrumentation__.Notify(17107)
	if err := e.Buffer.Add(ctx, event); err != nil {
		__antithesis_instrumentation__.Notify(17109)
		return changefeedbase.MarkRetryableError(err)
	} else {
		__antithesis_instrumentation__.Notify(17110)
	}
	__antithesis_instrumentation__.Notify(17108)
	return nil
}

var _ Buffer = (*errorWrapperEventBuffer)(nil)
