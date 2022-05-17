package colfetcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type cFetcherTableArgs struct {
	spec descpb.IndexFetchSpec

	ColIdxMap catalog.TableColMap

	typs []*types.T
}

var cFetcherTableArgsPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(455543)
		return &cFetcherTableArgs{}
	},
}

func (a *cFetcherTableArgs) Release() {
	__antithesis_instrumentation__.Notify(455544)
	*a = cFetcherTableArgs{

		typs: a.typs[:0],
	}
	cFetcherTableArgsPool.Put(a)
}

func (a *cFetcherTableArgs) populateTypes(cols []descpb.IndexFetchSpec_Column) {
	__antithesis_instrumentation__.Notify(455545)
	if cap(a.typs) < len(cols) {
		__antithesis_instrumentation__.Notify(455547)
		a.typs = make([]*types.T, len(cols))
	} else {
		__antithesis_instrumentation__.Notify(455548)
		a.typs = a.typs[:len(cols)]
	}
	__antithesis_instrumentation__.Notify(455546)
	for i := range cols {
		__antithesis_instrumentation__.Notify(455549)
		a.typs[i] = cols[i].Type
	}
}

func populateTableArgs(
	ctx context.Context, flowCtx *execinfra.FlowCtx, fetchSpec *descpb.IndexFetchSpec,
) (_ *cFetcherTableArgs, _ error) {
	__antithesis_instrumentation__.Notify(455550)
	args := cFetcherTableArgsPool.Get().(*cFetcherTableArgs)

	*args = cFetcherTableArgs{
		spec: *fetchSpec,
		typs: args.typs,
	}

	resolver := flowCtx.NewTypeResolver(flowCtx.Txn)
	for i := range args.spec.FetchedColumns {
		__antithesis_instrumentation__.Notify(455554)
		if err := typedesc.EnsureTypeIsHydrated(ctx, args.spec.FetchedColumns[i].Type, &resolver); err != nil {
			__antithesis_instrumentation__.Notify(455555)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(455556)
		}
	}
	__antithesis_instrumentation__.Notify(455551)
	for i := range args.spec.KeyAndSuffixColumns {
		__antithesis_instrumentation__.Notify(455557)
		if err := typedesc.EnsureTypeIsHydrated(ctx, args.spec.KeyAndSuffixColumns[i].Type, &resolver); err != nil {
			__antithesis_instrumentation__.Notify(455558)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(455559)
		}
	}
	__antithesis_instrumentation__.Notify(455552)
	args.populateTypes(args.spec.FetchedColumns)
	for i := range args.spec.FetchedColumns {
		__antithesis_instrumentation__.Notify(455560)
		args.ColIdxMap.Set(args.spec.FetchedColumns[i].ColumnID, i)
	}
	__antithesis_instrumentation__.Notify(455553)

	return args, nil
}
