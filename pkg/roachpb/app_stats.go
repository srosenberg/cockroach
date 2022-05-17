package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/util"
)

type StmtFingerprintID uint64

func (m *StatementStatisticsKey) FingerprintID() StmtFingerprintID {
	__antithesis_instrumentation__.Notify(154997)
	return ConstructStatementFingerprintID(
		m.Query,
		m.Failed,
		m.ImplicitTxn,
		m.Database,
	)
}

func ConstructStatementFingerprintID(
	anonymizedStmt string, failed bool, implicitTxn bool, database string,
) StmtFingerprintID {
	__antithesis_instrumentation__.Notify(154998)
	fnv := util.MakeFNV64()
	for _, c := range anonymizedStmt {
		__antithesis_instrumentation__.Notify(155003)
		fnv.Add(uint64(c))
	}
	__antithesis_instrumentation__.Notify(154999)
	for _, c := range database {
		__antithesis_instrumentation__.Notify(155004)
		fnv.Add(uint64(c))
	}
	__antithesis_instrumentation__.Notify(155000)
	if failed {
		__antithesis_instrumentation__.Notify(155005)
		fnv.Add('F')
	} else {
		__antithesis_instrumentation__.Notify(155006)
		fnv.Add('S')
	}
	__antithesis_instrumentation__.Notify(155001)
	if implicitTxn {
		__antithesis_instrumentation__.Notify(155007)
		fnv.Add('I')
	} else {
		__antithesis_instrumentation__.Notify(155008)
		fnv.Add('E')
	}
	__antithesis_instrumentation__.Notify(155002)
	return StmtFingerprintID(fnv.Sum())
}

type TransactionFingerprintID uint64

const InvalidTransactionFingerprintID = TransactionFingerprintID(0)

func (t TransactionFingerprintID) Size() int64 {
	__antithesis_instrumentation__.Notify(155009)
	return 8
}

func (l *NumericStat) GetVariance(count int64) float64 {
	__antithesis_instrumentation__.Notify(155010)
	return l.SquaredDiffs / (float64(count) - 1)
}

func (l *NumericStat) Record(count int64, val float64) {
	__antithesis_instrumentation__.Notify(155011)
	delta := val - l.Mean
	l.Mean += delta / float64(count)
	l.SquaredDiffs += delta * (val - l.Mean)
}

func (l *NumericStat) Add(b NumericStat, countA, countB int64) {
	__antithesis_instrumentation__.Notify(155012)
	*l = AddNumericStats(*l, b, countA, countB)
}

func (l *NumericStat) AlmostEqual(b NumericStat, eps float64) bool {
	__antithesis_instrumentation__.Notify(155013)
	return math.Abs(l.Mean-b.Mean) <= eps && func() bool {
		__antithesis_instrumentation__.Notify(155014)
		return math.Abs(l.SquaredDiffs-b.SquaredDiffs) <= eps == true
	}() == true
}

func AddNumericStats(a, b NumericStat, countA, countB int64) NumericStat {
	__antithesis_instrumentation__.Notify(155015)
	total := float64(countA + countB)
	delta := b.Mean - a.Mean

	return NumericStat{
		Mean: ((a.Mean * float64(countA)) + (b.Mean * float64(countB))) / total,
		SquaredDiffs: (a.SquaredDiffs + b.SquaredDiffs) +
			delta*delta*float64(countA)*float64(countB)/total,
	}
}

func (si SensitiveInfo) GetScrubbedCopy() SensitiveInfo {
	__antithesis_instrumentation__.Notify(155016)
	output := SensitiveInfo{}

	output.LastErr = "<redacted>"

	return output
}

func (s *TxnStats) Add(other TxnStats) {
	__antithesis_instrumentation__.Notify(155017)
	s.TxnTimeSec.Add(other.TxnTimeSec, s.TxnCount, other.TxnCount)
	s.TxnCount += other.TxnCount
	s.ImplicitCount += other.ImplicitCount
	s.CommittedCount += other.CommittedCount
}

func (t *TransactionStatistics) Add(other *TransactionStatistics) {
	__antithesis_instrumentation__.Notify(155018)
	if other.MaxRetries > t.MaxRetries {
		__antithesis_instrumentation__.Notify(155020)
		t.MaxRetries = other.MaxRetries
	} else {
		__antithesis_instrumentation__.Notify(155021)
	}
	__antithesis_instrumentation__.Notify(155019)

	t.CommitLat.Add(other.CommitLat, t.Count, other.Count)
	t.RetryLat.Add(other.RetryLat, t.Count, other.Count)
	t.ServiceLat.Add(other.ServiceLat, t.Count, other.Count)
	t.NumRows.Add(other.NumRows, t.Count, other.Count)
	t.RowsRead.Add(other.RowsRead, t.Count, other.Count)
	t.BytesRead.Add(other.BytesRead, t.Count, other.Count)
	t.RowsWritten.Add(other.RowsWritten, t.Count, other.Count)

	t.ExecStats.Add(other.ExecStats)

	t.Count += other.Count
}

func (s *StatementStatistics) Add(other *StatementStatistics) {
	__antithesis_instrumentation__.Notify(155022)
	s.FirstAttemptCount += other.FirstAttemptCount
	if other.MaxRetries > s.MaxRetries {
		__antithesis_instrumentation__.Notify(155027)
		s.MaxRetries = other.MaxRetries
	} else {
		__antithesis_instrumentation__.Notify(155028)
	}
	__antithesis_instrumentation__.Notify(155023)
	s.SQLType = other.SQLType
	s.NumRows.Add(other.NumRows, s.Count, other.Count)
	s.ParseLat.Add(other.ParseLat, s.Count, other.Count)
	s.PlanLat.Add(other.PlanLat, s.Count, other.Count)
	s.RunLat.Add(other.RunLat, s.Count, other.Count)
	s.ServiceLat.Add(other.ServiceLat, s.Count, other.Count)
	s.OverheadLat.Add(other.OverheadLat, s.Count, other.Count)
	s.BytesRead.Add(other.BytesRead, s.Count, other.Count)
	s.RowsRead.Add(other.RowsRead, s.Count, other.Count)
	s.RowsWritten.Add(other.RowsWritten, s.Count, other.Count)
	s.Nodes = util.CombineUniqueInt64(s.Nodes, other.Nodes)
	s.PlanGists = util.CombineUniqueString(s.PlanGists, other.PlanGists)

	s.ExecStats.Add(other.ExecStats)

	if other.SensitiveInfo.LastErr != "" {
		__antithesis_instrumentation__.Notify(155029)
		s.SensitiveInfo.LastErr = other.SensitiveInfo.LastErr
	} else {
		__antithesis_instrumentation__.Notify(155030)
	}
	__antithesis_instrumentation__.Notify(155024)

	if s.SensitiveInfo.MostRecentPlanTimestamp.Before(other.SensitiveInfo.MostRecentPlanTimestamp) {
		__antithesis_instrumentation__.Notify(155031)
		s.SensitiveInfo = other.SensitiveInfo
	} else {
		__antithesis_instrumentation__.Notify(155032)
	}
	__antithesis_instrumentation__.Notify(155025)

	if s.LastExecTimestamp.Before(other.LastExecTimestamp) {
		__antithesis_instrumentation__.Notify(155033)
		s.LastExecTimestamp = other.LastExecTimestamp
	} else {
		__antithesis_instrumentation__.Notify(155034)
	}
	__antithesis_instrumentation__.Notify(155026)

	s.Count += other.Count
}

func (s *StatementStatistics) AlmostEqual(other *StatementStatistics, eps float64) bool {
	__antithesis_instrumentation__.Notify(155035)
	return s.Count == other.Count && func() bool {
		__antithesis_instrumentation__.Notify(155036)
		return s.FirstAttemptCount == other.FirstAttemptCount == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155037)
		return s.MaxRetries == other.MaxRetries == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155038)
		return s.NumRows.AlmostEqual(other.NumRows, eps) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155039)
		return s.ParseLat.AlmostEqual(other.ParseLat, eps) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155040)
		return s.PlanLat.AlmostEqual(other.PlanLat, eps) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155041)
		return s.RunLat.AlmostEqual(other.RunLat, eps) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155042)
		return s.ServiceLat.AlmostEqual(other.ServiceLat, eps) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155043)
		return s.OverheadLat.AlmostEqual(other.OverheadLat, eps) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155044)
		return s.SensitiveInfo.Equal(other.SensitiveInfo) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155045)
		return s.BytesRead.AlmostEqual(other.BytesRead, eps) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155046)
		return s.RowsRead.AlmostEqual(other.RowsRead, eps) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(155047)
		return s.RowsWritten.AlmostEqual(other.RowsWritten, eps) == true
	}() == true

}

func (s *ExecStats) Add(other ExecStats) {
	__antithesis_instrumentation__.Notify(155048)

	execStatCollectionCount := s.Count
	if execStatCollectionCount == 0 && func() bool {
		__antithesis_instrumentation__.Notify(155050)
		return other.Count == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(155051)

		execStatCollectionCount = 1
	} else {
		__antithesis_instrumentation__.Notify(155052)
	}
	__antithesis_instrumentation__.Notify(155049)
	s.NetworkBytes.Add(other.NetworkBytes, execStatCollectionCount, other.Count)
	s.MaxMemUsage.Add(other.MaxMemUsage, execStatCollectionCount, other.Count)
	s.ContentionTime.Add(other.ContentionTime, execStatCollectionCount, other.Count)
	s.NetworkMessages.Add(other.NetworkMessages, execStatCollectionCount, other.Count)
	s.MaxDiskUsage.Add(other.MaxDiskUsage, execStatCollectionCount, other.Count)

	s.Count += other.Count
}
