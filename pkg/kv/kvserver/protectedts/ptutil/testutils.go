package ptutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigptsreader"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func TestingVerifyProtectionTimestampExistsOnSpans(
	ctx context.Context,
	t *testing.T,
	srv serverutils.TestServerInterface,
	ptsReader spanconfig.ProtectedTSReader,
	protectionTimestamp hlc.Timestamp,
	spans roachpb.Spans,
) error {
	__antithesis_instrumentation__.Notify(112193)
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(112195)
		if err := spanconfigptsreader.TestingRefreshPTSState(
			ctx, t, ptsReader, srv.Clock().Now(),
		); err != nil {
			__antithesis_instrumentation__.Notify(112198)
			return err
		} else {
			__antithesis_instrumentation__.Notify(112199)
		}
		__antithesis_instrumentation__.Notify(112196)
		for _, sp := range spans {
			__antithesis_instrumentation__.Notify(112200)
			timestamps, _, err := ptsReader.GetProtectionTimestamps(ctx, sp)
			if err != nil {
				__antithesis_instrumentation__.Notify(112203)
				return err
			} else {
				__antithesis_instrumentation__.Notify(112204)
			}
			__antithesis_instrumentation__.Notify(112201)
			found := false
			for _, ts := range timestamps {
				__antithesis_instrumentation__.Notify(112205)
				if ts.Equal(protectionTimestamp) {
					__antithesis_instrumentation__.Notify(112206)
					found = true
					break
				} else {
					__antithesis_instrumentation__.Notify(112207)
				}
			}
			__antithesis_instrumentation__.Notify(112202)
			if !found {
				__antithesis_instrumentation__.Notify(112208)
				return errors.Newf("protection timestamp %s does not exist on span %s", protectionTimestamp, sp)
			} else {
				__antithesis_instrumentation__.Notify(112209)
			}
		}
		__antithesis_instrumentation__.Notify(112197)
		return nil
	})
	__antithesis_instrumentation__.Notify(112194)
	return nil
}
