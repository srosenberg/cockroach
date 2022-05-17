package spanconfigsplitter

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
)

var _ spanconfig.Splitter = &NoopSplitter{}

type NoopSplitter struct{}

func (i NoopSplitter) Splits(context.Context, catalog.TableDescriptor) (int, error) {
	__antithesis_instrumentation__.Notify(240960)
	return 0, nil
}
