package idxusage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type usageType int8

const (
	readOp usageType = iota

	writeOp
)

type indexUse struct {
	key roachpb.IndexUsageKey

	usageTyp usageType
}

type LocalIndexUsageStats struct {
	st *cluster.Settings

	mu struct {
		syncutil.RWMutex

		usageStats map[roachpb.TableID]*tableIndexStats

		lastReset time.Time
	}
}

type tableIndexStats struct {
	syncutil.RWMutex

	tableID roachpb.TableID

	stats map[roachpb.IndexID]*indexStats
}

type indexStats struct {
	syncutil.RWMutex
	roachpb.IndexUsageStatistics
}

type Config struct {
	ChannelSize uint64

	Setting *cluster.Settings
}

type IteratorOptions struct {
	SortedTableID bool
	SortedIndexID bool
	Max           *uint64
}

type StatsVisitor func(key *roachpb.IndexUsageKey, value *roachpb.IndexUsageStatistics) error

const DefaultChannelSize = uint64(128)

var emptyIndexUsageStats roachpb.IndexUsageStatistics

func NewLocalIndexUsageStats(cfg *Config) *LocalIndexUsageStats {
	__antithesis_instrumentation__.Notify(492999)
	is := &LocalIndexUsageStats{
		st: cfg.Setting,
	}
	is.mu.usageStats = make(map[roachpb.TableID]*tableIndexStats)

	return is
}

func NewLocalIndexUsageStatsFromExistingStats(
	cfg *Config, stats []roachpb.CollectedIndexUsageStatistics,
) *LocalIndexUsageStats {
	__antithesis_instrumentation__.Notify(493000)
	s := NewLocalIndexUsageStats(cfg)

	s.batchInsertLocked(stats)
	return s
}

func (s *LocalIndexUsageStats) RecordRead(key roachpb.IndexUsageKey) {
	__antithesis_instrumentation__.Notify(493001)
	s.insertIndexUsage(key, readOp)
}

func (s *LocalIndexUsageStats) Get(
	tableID roachpb.TableID, indexID roachpb.IndexID,
) roachpb.IndexUsageStatistics {
	__antithesis_instrumentation__.Notify(493002)
	s.mu.RLock()
	defer s.mu.RUnlock()

	table, ok := s.mu.usageStats[tableID]
	if !ok {
		__antithesis_instrumentation__.Notify(493005)

		emptyStats := emptyIndexUsageStats
		return emptyStats
	} else {
		__antithesis_instrumentation__.Notify(493006)
	}
	__antithesis_instrumentation__.Notify(493003)

	table.RLock()
	defer table.RUnlock()

	indexStats, ok := table.stats[indexID]
	if !ok {
		__antithesis_instrumentation__.Notify(493007)
		emptyStats := emptyIndexUsageStats
		return emptyStats
	} else {
		__antithesis_instrumentation__.Notify(493008)
	}
	__antithesis_instrumentation__.Notify(493004)

	indexStats.RLock()
	defer indexStats.RUnlock()
	return indexStats.IndexUsageStatistics
}

func (s *LocalIndexUsageStats) ForEach(options IteratorOptions, visitor StatsVisitor) error {
	__antithesis_instrumentation__.Notify(493009)
	maxIterationLimit := uint64(math.MaxUint64)
	if options.Max != nil {
		__antithesis_instrumentation__.Notify(493014)
		maxIterationLimit = *options.Max
	} else {
		__antithesis_instrumentation__.Notify(493015)
	}
	__antithesis_instrumentation__.Notify(493010)

	s.mu.RLock()
	tableIDLists := make([]roachpb.TableID, 0, len(s.mu.usageStats))
	for tableID := range s.mu.usageStats {
		__antithesis_instrumentation__.Notify(493016)
		tableIDLists = append(tableIDLists, tableID)
	}
	__antithesis_instrumentation__.Notify(493011)
	s.mu.RUnlock()

	if options.SortedTableID {
		__antithesis_instrumentation__.Notify(493017)
		sort.Slice(tableIDLists, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(493018)
			return tableIDLists[i] < tableIDLists[j]
		})
	} else {
		__antithesis_instrumentation__.Notify(493019)
	}
	__antithesis_instrumentation__.Notify(493012)

	for _, tableID := range tableIDLists {
		__antithesis_instrumentation__.Notify(493020)
		tableIdxStats := s.getStatsForTableID(tableID, false)

		if tableIdxStats == nil {
			__antithesis_instrumentation__.Notify(493023)
			continue
		} else {
			__antithesis_instrumentation__.Notify(493024)
		}
		__antithesis_instrumentation__.Notify(493021)

		var err error
		maxIterationLimit, err = tableIdxStats.iterateIndexStats(options.SortedIndexID, maxIterationLimit, visitor)
		if err != nil {
			__antithesis_instrumentation__.Notify(493025)
			return errors.Wrap(err, "unexpected error encountered when iterating through index usage stats")
		} else {
			__antithesis_instrumentation__.Notify(493026)
		}
		__antithesis_instrumentation__.Notify(493022)

		if maxIterationLimit == 0 {
			__antithesis_instrumentation__.Notify(493027)
			break
		} else {
			__antithesis_instrumentation__.Notify(493028)
		}
	}
	__antithesis_instrumentation__.Notify(493013)

	return nil
}

func (s *LocalIndexUsageStats) batchInsertLocked(
	otherStats []roachpb.CollectedIndexUsageStatistics,
) {
	__antithesis_instrumentation__.Notify(493029)
	for _, newStats := range otherStats {
		__antithesis_instrumentation__.Notify(493030)
		tableIndexStats := s.getStatsForTableIDLocked(newStats.Key.TableID, true)
		stats := tableIndexStats.getStatsForIndexIDLocked(newStats.Key.IndexID, true)
		stats.Add(&newStats.Stats)
	}
}

func (s *LocalIndexUsageStats) clear() {
	__antithesis_instrumentation__.Notify(493031)
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, tableStats := range s.mu.usageStats {
		__antithesis_instrumentation__.Notify(493033)
		tableStats.clear()
	}
	__antithesis_instrumentation__.Notify(493032)
	s.mu.lastReset = timeutil.Now()
}

func (s *LocalIndexUsageStats) Reset() {
	__antithesis_instrumentation__.Notify(493034)
	s.clear()
}

func (s *LocalIndexUsageStats) insertIndexUsage(key roachpb.IndexUsageKey, usageTyp usageType) {
	__antithesis_instrumentation__.Notify(493035)

	if !Enable.Get(&s.st.SV) {
		__antithesis_instrumentation__.Notify(493037)
		return
	} else {
		__antithesis_instrumentation__.Notify(493038)
	}
	__antithesis_instrumentation__.Notify(493036)

	tableStats := s.getStatsForTableID(key.TableID, true)
	indexStats := tableStats.getStatsForIndexID(key.IndexID, true)
	indexStats.Lock()
	defer indexStats.Unlock()
	switch usageTyp {

	case readOp:
		__antithesis_instrumentation__.Notify(493039)
		indexStats.TotalReadCount++
		indexStats.LastRead = timeutil.Now()

	case writeOp:
		__antithesis_instrumentation__.Notify(493040)
		indexStats.TotalWriteCount++
		indexStats.LastWrite = timeutil.Now()
	default:
		__antithesis_instrumentation__.Notify(493041)
	}
}

func (s *LocalIndexUsageStats) getStatsForTableID(
	id roachpb.TableID, createIfNotExists bool,
) *tableIndexStats {
	__antithesis_instrumentation__.Notify(493042)

	s.mu.RLock()

	if tableIndexStats, ok := s.mu.usageStats[id]; ok || func() bool {
		__antithesis_instrumentation__.Notify(493044)
		return !createIfNotExists == true
	}() == true {
		__antithesis_instrumentation__.Notify(493045)
		defer s.mu.RUnlock()
		return tableIndexStats
	} else {
		__antithesis_instrumentation__.Notify(493046)
	}
	__antithesis_instrumentation__.Notify(493043)

	s.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.getStatsForTableIDLocked(id, createIfNotExists)
}

func (s *LocalIndexUsageStats) getStatsForTableIDLocked(
	id roachpb.TableID, createIfNotExists bool,
) *tableIndexStats {
	__antithesis_instrumentation__.Notify(493047)
	if tableIndexStats, ok := s.mu.usageStats[id]; ok {
		__antithesis_instrumentation__.Notify(493050)
		return tableIndexStats
	} else {
		__antithesis_instrumentation__.Notify(493051)
	}
	__antithesis_instrumentation__.Notify(493048)

	if createIfNotExists {
		__antithesis_instrumentation__.Notify(493052)
		newTableIndexStats := &tableIndexStats{
			tableID: id,
			stats:   make(map[roachpb.IndexID]*indexStats),
		}
		s.mu.usageStats[id] = newTableIndexStats
		return newTableIndexStats
	} else {
		__antithesis_instrumentation__.Notify(493053)
	}
	__antithesis_instrumentation__.Notify(493049)

	return nil
}

func (t *tableIndexStats) getStatsForIndexID(
	id roachpb.IndexID, createIfNotExists bool,
) *indexStats {
	__antithesis_instrumentation__.Notify(493054)
	t.RLock()

	if stats, ok := t.stats[id]; ok || func() bool {
		__antithesis_instrumentation__.Notify(493056)
		return !createIfNotExists == true
	}() == true {
		__antithesis_instrumentation__.Notify(493057)
		t.RUnlock()
		return stats
	} else {
		__antithesis_instrumentation__.Notify(493058)
	}
	__antithesis_instrumentation__.Notify(493055)

	t.RUnlock()
	t.Lock()
	defer t.Unlock()

	return t.getStatsForIndexIDLocked(id, createIfNotExists)
}

func (t *tableIndexStats) getStatsForIndexIDLocked(
	id roachpb.IndexID, createIfNotExists bool,
) *indexStats {
	__antithesis_instrumentation__.Notify(493059)
	if stats, ok := t.stats[id]; ok {
		__antithesis_instrumentation__.Notify(493062)
		return stats
	} else {
		__antithesis_instrumentation__.Notify(493063)
	}
	__antithesis_instrumentation__.Notify(493060)
	if createIfNotExists {
		__antithesis_instrumentation__.Notify(493064)
		newUsageEntry := &indexStats{}
		t.stats[id] = newUsageEntry
		return newUsageEntry
	} else {
		__antithesis_instrumentation__.Notify(493065)
	}
	__antithesis_instrumentation__.Notify(493061)
	return nil
}

func (s *LocalIndexUsageStats) GetLastReset() time.Time {
	__antithesis_instrumentation__.Notify(493066)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.lastReset
}

func (t *tableIndexStats) iterateIndexStats(
	orderedIndexID bool, iterLimit uint64, visitor StatsVisitor,
) (newIterLimit uint64, err error) {
	__antithesis_instrumentation__.Notify(493067)
	t.RLock()
	indexIDs := make([]roachpb.IndexID, 0, len(t.stats))
	for indexID := range t.stats {
		__antithesis_instrumentation__.Notify(493071)
		if iterLimit == 0 {
			__antithesis_instrumentation__.Notify(493073)
			break
		} else {
			__antithesis_instrumentation__.Notify(493074)
		}
		__antithesis_instrumentation__.Notify(493072)
		indexIDs = append(indexIDs, indexID)
		iterLimit--
	}
	__antithesis_instrumentation__.Notify(493068)
	t.RUnlock()

	if orderedIndexID {
		__antithesis_instrumentation__.Notify(493075)
		sort.Slice(indexIDs, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(493076)
			return indexIDs[i] < indexIDs[j]
		})
	} else {
		__antithesis_instrumentation__.Notify(493077)
	}
	__antithesis_instrumentation__.Notify(493069)

	for _, indexID := range indexIDs {
		__antithesis_instrumentation__.Notify(493078)
		indexStats := t.getStatsForIndexID(indexID, false)

		if indexStats == nil {
			__antithesis_instrumentation__.Notify(493080)
			continue
		} else {
			__antithesis_instrumentation__.Notify(493081)
		}
		__antithesis_instrumentation__.Notify(493079)

		indexStats.RLock()

		statsCopy := indexStats.IndexUsageStatistics
		indexStats.RUnlock()

		if err := visitor(&roachpb.IndexUsageKey{
			TableID: t.tableID,
			IndexID: indexID,
		}, &statsCopy); err != nil {
			__antithesis_instrumentation__.Notify(493082)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(493083)
		}
	}
	__antithesis_instrumentation__.Notify(493070)
	return iterLimit, nil
}

func (t *tableIndexStats) clear() {
	__antithesis_instrumentation__.Notify(493084)
	t.Lock()
	defer t.Unlock()

	for k := range t.stats {
		__antithesis_instrumentation__.Notify(493085)
		delete(t.stats, k)
	}
}
