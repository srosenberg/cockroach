package spanconfiglimiter

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
)

var _ spanconfig.Limiter = &NoopLimiter{}

type NoopLimiter struct{}

func (n NoopLimiter) ShouldLimit(context.Context, *kv.Txn, int) (bool, error) {
	__antithesis_instrumentation__.Notify(240706)
	return false, nil
}
