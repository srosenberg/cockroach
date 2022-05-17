package mutations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util"
)

const productionMaxBatchSize = 10000

var maxBatchSize = defaultMaxBatchSize

var defaultMaxBatchSize = int64(util.ConstantWithMetamorphicTestRange(
	"max-batch-size",
	productionMaxBatchSize,
	1,
	productionMaxBatchSize,
))

var testingMaxBatchByteSize = util.ConstantWithMetamorphicTestRange(
	"max-batch-byte-size",
	0,
	1,
	32<<20,
)

func MaxBatchSize(forceProductionMaxBatchSize bool) int {
	__antithesis_instrumentation__.Notify(501771)
	if forceProductionMaxBatchSize {
		__antithesis_instrumentation__.Notify(501773)
		return productionMaxBatchSize
	} else {
		__antithesis_instrumentation__.Notify(501774)
	}
	__antithesis_instrumentation__.Notify(501772)
	return int(atomic.LoadInt64(&maxBatchSize))
}

func SetMaxBatchSizeForTests(newMaxBatchSize int) {
	__antithesis_instrumentation__.Notify(501775)
	atomic.SwapInt64(&maxBatchSize, int64(newMaxBatchSize))
}

func ResetMaxBatchSizeForTests() {
	__antithesis_instrumentation__.Notify(501776)
	atomic.SwapInt64(&maxBatchSize, defaultMaxBatchSize)
}

func MaxBatchByteSize(clusterSetting int, forceProductionBatchSizes bool) int {
	__antithesis_instrumentation__.Notify(501777)
	if forceProductionBatchSizes || func() bool {
		__antithesis_instrumentation__.Notify(501779)
		return testingMaxBatchByteSize == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(501780)
		return clusterSetting
	} else {
		__antithesis_instrumentation__.Notify(501781)
	}
	__antithesis_instrumentation__.Notify(501778)
	return testingMaxBatchByteSize
}
