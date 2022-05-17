package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type TableID uint32

type IndexID uint32

func (m *IndexUsageStatistics) Add(other *IndexUsageStatistics) {
	__antithesis_instrumentation__.Notify(170488)
	m.TotalRowsRead += other.TotalRowsRead
	m.TotalRowsWritten += other.TotalRowsWritten

	m.TotalReadCount += other.TotalReadCount
	m.TotalWriteCount += other.TotalWriteCount

	if m.LastWrite.Before(other.LastWrite) {
		__antithesis_instrumentation__.Notify(170490)
		m.LastWrite = other.LastWrite
	} else {
		__antithesis_instrumentation__.Notify(170491)
	}
	__antithesis_instrumentation__.Notify(170489)

	if m.LastRead.Before(other.LastRead) {
		__antithesis_instrumentation__.Notify(170492)
		m.LastRead = other.LastRead
	} else {
		__antithesis_instrumentation__.Notify(170493)
	}
}
