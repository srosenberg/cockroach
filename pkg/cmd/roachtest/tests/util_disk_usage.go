package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
)

func getDiskUsageInBytes(
	ctx context.Context, c cluster.Cluster, logger *logger.Logger, nodeIdx int,
) (int, error) {
	__antithesis_instrumentation__.Notify(52234)
	var result install.RunResultDetails
	for {
		__antithesis_instrumentation__.Notify(52238)
		var err error

		result, err = c.RunWithDetailsSingleNode(
			ctx,
			logger,
			c.Node(nodeIdx),
			"du -sk {store-dir} 2>/dev/null | grep -oE '^[0-9]+'",
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(52240)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(52242)
				return 0, ctx.Err()
			} else {
				__antithesis_instrumentation__.Notify(52243)
			}
			__antithesis_instrumentation__.Notify(52241)

			logger.Printf("retrying disk usage computation after spurious error: %s", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(52244)
		}
		__antithesis_instrumentation__.Notify(52239)

		break
	}
	__antithesis_instrumentation__.Notify(52235)

	if strings.Contains(result.Stdout, "Warning") {
		__antithesis_instrumentation__.Notify(52245)
		result.Stdout = strings.Split(result.Stdout, "\n")[1]
	} else {
		__antithesis_instrumentation__.Notify(52246)
	}
	__antithesis_instrumentation__.Notify(52236)

	size, err := strconv.Atoi(strings.TrimSpace(result.Stdout))
	if err != nil {
		__antithesis_instrumentation__.Notify(52247)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(52248)
	}
	__antithesis_instrumentation__.Notify(52237)

	return size * 1024, nil
}
