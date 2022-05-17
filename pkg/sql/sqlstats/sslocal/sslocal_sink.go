package sslocal

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/ssmemstorage"
)

type Sink interface {
	AddAppStats(ctx context.Context, appName string, other *ssmemstorage.Container) error
}

var _ Sink = &SQLStats{}

func (s *SQLStats) AddAppStats(
	ctx context.Context, appName string, other *ssmemstorage.Container,
) error {
	__antithesis_instrumentation__.Notify(625445)
	stats := s.getStatsForApplication(appName)

	return stats.Add(ctx, other)
}
