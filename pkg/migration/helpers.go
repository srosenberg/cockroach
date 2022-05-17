package migration

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func FenceVersionFor(
	ctx context.Context, cv clusterversion.ClusterVersion,
) clusterversion.ClusterVersion {
	__antithesis_instrumentation__.Notify(128142)
	if (cv.Internal % 2) != 0 {
		__antithesis_instrumentation__.Notify(128144)
		log.Fatalf(ctx, "only even numbered internal versions allowed, found %s", cv.Version)
	} else {
		__antithesis_instrumentation__.Notify(128145)
	}
	__antithesis_instrumentation__.Notify(128143)

	fenceCV := cv
	fenceCV.Internal--
	return fenceCV
}
