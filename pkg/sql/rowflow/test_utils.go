package rowflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func MakeTestRouter(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	spec *execinfrapb.OutputRouterSpec,
	streams []execinfra.RowReceiver,
	types []*types.T,
	wg *sync.WaitGroup,
) (execinfra.RowReceiver, error) {
	__antithesis_instrumentation__.Notify(576339)
	r, err := makeRouter(spec, streams)
	if err != nil {
		__antithesis_instrumentation__.Notify(576341)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(576342)
	}
	__antithesis_instrumentation__.Notify(576340)
	r.init(ctx, flowCtx, types)
	r.Start(ctx, wg, nil)
	return r, nil
}
