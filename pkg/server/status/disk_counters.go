//go:build !darwin
// +build !darwin

package status

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

func getDiskCounters(ctx context.Context) ([]diskStats, error) {
	__antithesis_instrumentation__.Notify(235424)
	driveStats, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(235427)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235428)
	}
	__antithesis_instrumentation__.Notify(235425)

	output := make([]diskStats, len(driveStats))
	i := 0
	for _, counters := range driveStats {
		__antithesis_instrumentation__.Notify(235429)
		output[i] = diskStats{
			readBytes:      int64(counters.ReadBytes),
			readCount:      int64(counters.ReadCount),
			readTime:       time.Duration(counters.ReadTime) * time.Millisecond,
			writeBytes:     int64(counters.WriteBytes),
			writeCount:     int64(counters.WriteCount),
			writeTime:      time.Duration(counters.WriteTime) * time.Millisecond,
			ioTime:         time.Duration(counters.IoTime) * time.Millisecond,
			weightedIOTime: time.Duration(counters.WeightedIO) * time.Millisecond,
			iopsInProgress: int64(counters.IopsInProgress),
		}
		i++
	}
	__antithesis_instrumentation__.Notify(235426)

	return output, nil
}
