package types

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

var (
	OneIntCol = []*T{Int}

	TwoIntCols = []*T{Int, Int}

	ThreeIntCols = []*T{Int, Int, Int}

	FourIntCols = []*T{Int, Int, Int, Int}
)

func MakeIntCols(numCols int) []*T {
	__antithesis_instrumentation__.Notify(629569)
	ret := make([]*T, numCols)
	for i := 0; i < numCols; i++ {
		__antithesis_instrumentation__.Notify(629571)
		ret[i] = Int
	}
	__antithesis_instrumentation__.Notify(629570)
	return ret
}
