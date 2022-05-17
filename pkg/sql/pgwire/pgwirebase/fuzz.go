//go:build gofuzz
// +build gofuzz

package pgwirebase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

var (
	typs = func() []*types.T {
		__antithesis_instrumentation__.Notify(561386)
		var ret []*types.T
		for _, typ := range types.OidToType {
			__antithesis_instrumentation__.Notify(561388)
			ret = append(ret, typ)
		}
		__antithesis_instrumentation__.Notify(561387)
		return ret
	}()
)

func FuzzDecodeDatum(data []byte) int {
	__antithesis_instrumentation__.Notify(561389)
	if len(data) < 2 {
		__antithesis_instrumentation__.Notify(561392)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(561393)
	}
	__antithesis_instrumentation__.Notify(561390)

	typ := typs[int(data[1])%len(typs)]
	code := FormatCode(data[0]) % (FormatBinary + 1)
	b := data[2:]

	evalCtx := tree.NewTestingEvalContext(cluster.MakeTestingClusterSettings())
	defer evalCtx.Stop(context.Background())

	_, err := DecodeDatum(evalCtx, typ, code, b)
	if err != nil {
		__antithesis_instrumentation__.Notify(561394)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(561395)
	}
	__antithesis_instrumentation__.Notify(561391)
	return 1
}
