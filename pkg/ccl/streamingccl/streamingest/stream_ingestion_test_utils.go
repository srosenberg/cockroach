package streamingest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/cdctest"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type streamClientValidator struct {
	cdctest.StreamValidator

	mu syncutil.Mutex
}

func newStreamClientValidator() *streamClientValidator {
	__antithesis_instrumentation__.Notify(25614)
	ov := cdctest.NewStreamOrderValidator()
	return &streamClientValidator{
		StreamValidator: ov,
	}
}

func (sv *streamClientValidator) noteRow(
	partition string, key, value string, updated hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(25615)
	sv.mu.Lock()
	defer sv.mu.Unlock()
	return sv.NoteRow(partition, key, value, updated)
}

func (sv *streamClientValidator) noteResolved(partition string, resolved hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(25616)
	sv.mu.Lock()
	defer sv.mu.Unlock()
	return sv.NoteResolved(partition, resolved)
}

func (sv *streamClientValidator) failures() []string {
	__antithesis_instrumentation__.Notify(25617)
	sv.mu.Lock()
	defer sv.mu.Unlock()
	return sv.Failures()
}

func (sv *streamClientValidator) getValuesForKeyBelowTimestamp(
	key string, timestamp hlc.Timestamp,
) ([]roachpb.KeyValue, error) {
	__antithesis_instrumentation__.Notify(25618)
	sv.mu.Lock()
	defer sv.mu.Unlock()
	return sv.GetValuesForKeyBelowTimestamp(key, timestamp)
}
