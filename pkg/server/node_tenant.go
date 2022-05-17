package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

const TraceRedactedMarker = redact.RedactableString("verbose trace message redacted")

func redactRecordingForTenant(tenID roachpb.TenantID, rec tracing.Recording) error {
	__antithesis_instrumentation__.Notify(194854)
	if tenID == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(194857)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(194858)
	}
	__antithesis_instrumentation__.Notify(194855)
	for i := range rec {
		__antithesis_instrumentation__.Notify(194859)
		sp := &rec[i]
		sp.Tags = nil
		for j := range sp.Logs {
			__antithesis_instrumentation__.Notify(194860)
			record := &sp.Logs[j]
			if record.Message != "" && func() bool {
				__antithesis_instrumentation__.Notify(194862)
				return !sp.RedactableLogs == true
			}() == true {
				__antithesis_instrumentation__.Notify(194863)

				return errors.AssertionFailedf(
					"recording has non-redactable span with the Message field set: %s", sp)
			} else {
				__antithesis_instrumentation__.Notify(194864)
			}
			__antithesis_instrumentation__.Notify(194861)
			record.Message = record.Message.Redact()

			for k := range record.DeprecatedFields {
				__antithesis_instrumentation__.Notify(194865)
				field := &record.DeprecatedFields[k]
				if field.Key != tracingpb.LogMessageField {
					__antithesis_instrumentation__.Notify(194868)

					field.Value = TraceRedactedMarker
					continue
				} else {
					__antithesis_instrumentation__.Notify(194869)
				}
				__antithesis_instrumentation__.Notify(194866)
				if !sp.RedactableLogs {
					__antithesis_instrumentation__.Notify(194870)

					field.Value = TraceRedactedMarker
					continue
				} else {
					__antithesis_instrumentation__.Notify(194871)
				}
				__antithesis_instrumentation__.Notify(194867)
				field.Value = field.Value.Redact()
			}
		}
	}
	__antithesis_instrumentation__.Notify(194856)
	return nil
}
