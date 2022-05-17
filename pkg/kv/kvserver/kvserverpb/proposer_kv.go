package kvserverpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

var maxRaftCommandFooterSize = (&RaftCommandFooter{
	MaxLeaseIndex: math.MaxUint64,
	ClosedTimestamp: hlc.Timestamp{
		WallTime:  math.MaxInt64,
		Logical:   math.MaxInt32,
		Synthetic: true,
	},
}).Size()

func MaxRaftCommandFooterSize() int {
	__antithesis_instrumentation__.Notify(102148)
	return maxRaftCommandFooterSize
}

func (r ReplicatedEvalResult) IsZero() bool {
	__antithesis_instrumentation__.Notify(102149)
	return r == ReplicatedEvalResult{}
}
