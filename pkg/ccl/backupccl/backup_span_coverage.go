package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/span"
	"github.com/cockroachdb/errors"
)

func checkCoverage(ctx context.Context, spans []roachpb.Span, backups []BackupManifest) error {
	__antithesis_instrumentation__.Notify(8961)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(8966)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(8967)
	}
	__antithesis_instrumentation__.Notify(8962)

	frontier, err := span.MakeFrontier(spans...)
	if err != nil {
		__antithesis_instrumentation__.Notify(8968)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8969)
	}
	__antithesis_instrumentation__.Notify(8963)

	for i := range backups {
		__antithesis_instrumentation__.Notify(8970)
		for _, sp := range backups[i].IntroducedSpans {
			__antithesis_instrumentation__.Notify(8971)
			if _, err := frontier.Forward(sp, backups[i].StartTime); err != nil {
				__antithesis_instrumentation__.Notify(8972)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8973)
			}
		}
	}
	__antithesis_instrumentation__.Notify(8964)

	for i := range backups {
		__antithesis_instrumentation__.Notify(8974)

		if start, required := frontier.Frontier(), backups[i].StartTime; start.Less(required) {
			__antithesis_instrumentation__.Notify(8977)
			s := frontier.PeekFrontierSpan()
			return errors.Errorf(
				"no backup covers time [%s,%s) for range [%s,%s) (or backups listed out of order)",
				start, required, s.Key, s.EndKey,
			)
		} else {
			__antithesis_instrumentation__.Notify(8978)
		}
		__antithesis_instrumentation__.Notify(8975)

		for _, s := range backups[i].Spans {
			__antithesis_instrumentation__.Notify(8979)
			if _, err := frontier.Forward(s, backups[i].EndTime); err != nil {
				__antithesis_instrumentation__.Notify(8980)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8981)
			}
		}
		__antithesis_instrumentation__.Notify(8976)

		if end, required := frontier.Frontier(), backups[i].EndTime; end.Less(required) {
			__antithesis_instrumentation__.Notify(8982)
			return errors.Errorf("expected previous backups to cover until time %v, got %v (e.g. span %v)",
				required, end, frontier.PeekFrontierSpan())
		} else {
			__antithesis_instrumentation__.Notify(8983)
		}
	}
	__antithesis_instrumentation__.Notify(8965)

	return nil
}
