//go:build darwin
// +build darwin

package status

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/lufia/iostat"
)

func getDiskCounters(context.Context) ([]diskStats, error) {
	__antithesis_instrumentation__.Notify(235430)
	driveStats, err := iostat.ReadDriveStats()
	if err != nil {
		__antithesis_instrumentation__.Notify(235433)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235434)
	}
	__antithesis_instrumentation__.Notify(235431)

	output := make([]diskStats, len(driveStats))
	for i, counters := range driveStats {
		__antithesis_instrumentation__.Notify(235435)
		output[i] = diskStats{
			readBytes:      counters.BytesRead,
			readCount:      counters.NumRead,
			readTime:       counters.TotalReadTime,
			writeBytes:     counters.BytesWritten,
			writeCount:     counters.NumWrite,
			writeTime:      counters.TotalWriteTime,
			ioTime:         0,
			weightedIOTime: 0,
			iopsInProgress: 0,
		}
	}
	__antithesis_instrumentation__.Notify(235432)

	return output, nil
}
