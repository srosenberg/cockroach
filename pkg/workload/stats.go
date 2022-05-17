package workload

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const AutoStatsName = "__auto__"

type JSONStatistic struct {
	Name          string   `json:"name,omitempty"`
	CreatedAt     string   `json:"created_at"`
	Columns       []string `json:"columns"`
	RowCount      uint64   `json:"row_count"`
	DistinctCount uint64   `json:"distinct_count"`
	NullCount     uint64   `json:"null_count"`
}

func MakeStat(columns []string, rowCount, distinctCount, nullCount uint64) JSONStatistic {
	__antithesis_instrumentation__.Notify(697438)
	return JSONStatistic{
		Name:          AutoStatsName,
		CreatedAt:     timeutil.Now().Round(time.Microsecond).UTC().Format(timestampOutputFormat),
		Columns:       columns,
		RowCount:      rowCount,
		DistinctCount: distinctCount,
		NullCount:     nullCount,
	}
}

func DistinctCount(rowCount, maxDistinctCount uint64) uint64 {
	__antithesis_instrumentation__.Notify(697439)
	n := float64(maxDistinctCount)
	k := float64(rowCount)

	count := n * (1 - math.Pow((n-1)/n, k))
	return uint64(int64(math.Round(count)))
}
