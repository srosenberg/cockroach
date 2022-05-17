package stats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type TableStatistic struct {
	TableStatisticProto

	Histogram []cat.HistogramBucket
}

type TableStatisticsCache struct {
	mu struct {
		syncutil.Mutex
		cache *cache.UnorderedCache

		numInternalQueries int64
	}
	ClientDB    *kv.DB
	SQLExecutor sqlutil.InternalExecutor
	Codec       keys.SQLCodec
	Settings    *cluster.Settings

	collectionFactory *descs.CollectionFactory

	datumAlloc tree.DatumAlloc
}

type cacheEntry struct {
	mustWait bool
	waitCond sync.Cond

	lastRefreshTimestamp hlc.Timestamp

	refreshing bool

	stats []*TableStatistic

	err error
}

func NewTableStatisticsCache(
	ctx context.Context,
	cacheSize int,
	db *kv.DB,
	sqlExecutor sqlutil.InternalExecutor,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	rangeFeedFactory *rangefeed.Factory,
	cf *descs.CollectionFactory,
) *TableStatisticsCache {
	__antithesis_instrumentation__.Notify(626839)
	tableStatsCache := &TableStatisticsCache{
		ClientDB:          db,
		SQLExecutor:       sqlExecutor,
		Codec:             codec,
		Settings:          settings,
		collectionFactory: cf,
	}
	tableStatsCache.mu.cache = cache.NewUnorderedCache(cache.Config{
		Policy: cache.CacheLRU,
		ShouldEvict: func(s int, key, value interface{}) bool {
			__antithesis_instrumentation__.Notify(626842)
			return s > cacheSize
		},
	})
	__antithesis_instrumentation__.Notify(626840)

	statsTablePrefix := codec.TablePrefix(keys.TableStatisticsTableID)
	statsTableSpan := roachpb.Span{
		Key:    statsTablePrefix,
		EndKey: statsTablePrefix.PrefixEnd(),
	}

	var lastTableID descpb.ID
	var lastTS hlc.Timestamp

	handleEvent := func(ctx context.Context, kv *roachpb.RangeFeedValue) {
		__antithesis_instrumentation__.Notify(626843)
		tableID, err := decodeTableStatisticsKV(codec, kv, &tableStatsCache.datumAlloc)
		if err != nil {
			__antithesis_instrumentation__.Notify(626846)
			log.Warningf(ctx, "failed to decode table statistics row %v: %v", kv.Key, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(626847)
		}
		__antithesis_instrumentation__.Notify(626844)
		ts := kv.Value.Timestamp

		if tableID == lastTableID && func() bool {
			__antithesis_instrumentation__.Notify(626848)
			return ts == lastTS == true
		}() == true {
			__antithesis_instrumentation__.Notify(626849)
			return
		} else {
			__antithesis_instrumentation__.Notify(626850)
		}
		__antithesis_instrumentation__.Notify(626845)
		lastTableID = tableID
		lastTS = ts
		tableStatsCache.refreshTableStats(ctx, tableID, ts)
	}
	__antithesis_instrumentation__.Notify(626841)

	_, _ = rangeFeedFactory.RangeFeed(
		ctx,
		"table-stats-cache",
		[]roachpb.Span{statsTableSpan},
		db.Clock().Now(),
		handleEvent,
	)

	return tableStatsCache
}

func decodeTableStatisticsKV(
	codec keys.SQLCodec, kv *roachpb.RangeFeedValue, da *tree.DatumAlloc,
) (tableDesc descpb.ID, err error) {
	__antithesis_instrumentation__.Notify(626851)

	types := []*types.T{types.Int, types.Int}
	dirs := []descpb.IndexDescriptor_Direction{descpb.IndexDescriptor_ASC, descpb.IndexDescriptor_ASC}
	keyVals := make([]rowenc.EncDatum, 2)
	if _, _, err := rowenc.DecodeIndexKey(codec, types, keyVals, dirs, kv.Key); err != nil {
		__antithesis_instrumentation__.Notify(626855)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(626856)
	}
	__antithesis_instrumentation__.Notify(626852)

	if err := keyVals[0].EnsureDecoded(types[0], da); err != nil {
		__antithesis_instrumentation__.Notify(626857)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(626858)
	}
	__antithesis_instrumentation__.Notify(626853)

	tableID, ok := keyVals[0].Datum.(*tree.DInt)
	if !ok {
		__antithesis_instrumentation__.Notify(626859)
		return 0, errors.New("invalid tableID value")
	} else {
		__antithesis_instrumentation__.Notify(626860)
	}
	__antithesis_instrumentation__.Notify(626854)
	return descpb.ID(uint32(*tableID)), nil
}

func (sc *TableStatisticsCache) GetTableStats(
	ctx context.Context, table catalog.TableDescriptor,
) ([]*TableStatistic, error) {
	__antithesis_instrumentation__.Notify(626861)
	if !hasStatistics(table) {
		__antithesis_instrumentation__.Notify(626863)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(626864)
	}
	__antithesis_instrumentation__.Notify(626862)
	return sc.getTableStatsFromCache(ctx, table.GetID())
}

func hasStatistics(table catalog.TableDescriptor) bool {
	__antithesis_instrumentation__.Notify(626865)
	if catalog.IsSystemDescriptor(table) {
		__antithesis_instrumentation__.Notify(626869)

		return false
	} else {
		__antithesis_instrumentation__.Notify(626870)
	}
	__antithesis_instrumentation__.Notify(626866)
	if table.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(626871)

		return false
	} else {
		__antithesis_instrumentation__.Notify(626872)
	}
	__antithesis_instrumentation__.Notify(626867)
	if table.IsView() {
		__antithesis_instrumentation__.Notify(626873)

		return false
	} else {
		__antithesis_instrumentation__.Notify(626874)
	}
	__antithesis_instrumentation__.Notify(626868)
	return true
}

func (sc *TableStatisticsCache) getTableStatsFromCache(
	ctx context.Context, tableID descpb.ID,
) ([]*TableStatistic, error) {
	__antithesis_instrumentation__.Notify(626875)
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if found, e := sc.lookupStatsLocked(ctx, tableID, false); found {
		__antithesis_instrumentation__.Notify(626877)
		return e.stats, e.err
	} else {
		__antithesis_instrumentation__.Notify(626878)
	}
	__antithesis_instrumentation__.Notify(626876)

	return sc.addCacheEntryLocked(ctx, tableID)
}

func (sc *TableStatisticsCache) lookupStatsLocked(
	ctx context.Context, tableID descpb.ID, stealthy bool,
) (found bool, e *cacheEntry) {
	__antithesis_instrumentation__.Notify(626879)
	var eUntyped interface{}
	var ok bool

	if !stealthy {
		__antithesis_instrumentation__.Notify(626883)
		eUntyped, ok = sc.mu.cache.Get(tableID)
	} else {
		__antithesis_instrumentation__.Notify(626884)
		eUntyped, ok = sc.mu.cache.StealthyGet(tableID)
	}
	__antithesis_instrumentation__.Notify(626880)
	if !ok {
		__antithesis_instrumentation__.Notify(626885)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(626886)
	}
	__antithesis_instrumentation__.Notify(626881)
	e = eUntyped.(*cacheEntry)

	if e.mustWait {
		__antithesis_instrumentation__.Notify(626887)

		log.VEventf(ctx, 1, "waiting for statistics for table %d", tableID)
		e.waitCond.Wait()
		log.VEventf(ctx, 1, "finished waiting for statistics for table %d", tableID)
	} else {
		__antithesis_instrumentation__.Notify(626888)

		if log.V(2) {
			__antithesis_instrumentation__.Notify(626889)
			log.Infof(ctx, "statistics for table %d found in cache", tableID)
		} else {
			__antithesis_instrumentation__.Notify(626890)
		}
	}
	__antithesis_instrumentation__.Notify(626882)
	return true, e
}

func (sc *TableStatisticsCache) addCacheEntryLocked(
	ctx context.Context, tableID descpb.ID,
) (stats []*TableStatistic, err error) {
	__antithesis_instrumentation__.Notify(626891)

	e := &cacheEntry{
		mustWait: true,
		waitCond: sync.Cond{L: &sc.mu},
	}
	sc.mu.cache.Add(tableID, e)
	sc.mu.numInternalQueries++

	func() {
		__antithesis_instrumentation__.Notify(626894)
		sc.mu.Unlock()
		defer sc.mu.Lock()

		log.VEventf(ctx, 1, "reading statistics for table %d", tableID)
		stats, err = sc.getTableStatsFromDB(ctx, tableID)
		log.VEventf(ctx, 1, "finished reading statistics for table %d", tableID)
	}()
	__antithesis_instrumentation__.Notify(626892)

	e.mustWait = false
	e.stats, e.err = stats, err

	e.waitCond.Broadcast()

	if err != nil {
		__antithesis_instrumentation__.Notify(626895)

		sc.mu.cache.Del(tableID)
	} else {
		__antithesis_instrumentation__.Notify(626896)
	}
	__antithesis_instrumentation__.Notify(626893)

	return stats, err
}

func (sc *TableStatisticsCache) refreshCacheEntry(
	ctx context.Context, tableID descpb.ID, ts hlc.Timestamp,
) {
	__antithesis_instrumentation__.Notify(626897)
	sc.mu.Lock()
	defer sc.mu.Unlock()

	found, e := sc.lookupStatsLocked(ctx, tableID, true)
	if !found || func() bool {
		__antithesis_instrumentation__.Notify(626902)
		return e.err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(626903)
		return
	} else {
		__antithesis_instrumentation__.Notify(626904)
	}
	__antithesis_instrumentation__.Notify(626898)
	if ts.LessEq(e.lastRefreshTimestamp) {
		__antithesis_instrumentation__.Notify(626905)

		return
	} else {
		__antithesis_instrumentation__.Notify(626906)
	}
	__antithesis_instrumentation__.Notify(626899)
	e.lastRefreshTimestamp = ts

	if e.refreshing {
		__antithesis_instrumentation__.Notify(626907)
		return
	} else {
		__antithesis_instrumentation__.Notify(626908)
	}
	__antithesis_instrumentation__.Notify(626900)
	e.refreshing = true

	var stats []*TableStatistic
	var err error
	for {
		__antithesis_instrumentation__.Notify(626909)
		func() {
			__antithesis_instrumentation__.Notify(626912)
			sc.mu.numInternalQueries++
			sc.mu.Unlock()
			defer sc.mu.Lock()

			log.VEventf(ctx, 1, "refreshing statistics for table %d", tableID)

			stats, err = sc.getTableStatsFromDB(ctx, tableID)
			log.VEventf(ctx, 1, "done refreshing statistics for table %d", tableID)
		}()
		__antithesis_instrumentation__.Notify(626910)
		if e.lastRefreshTimestamp.Equal(ts) {
			__antithesis_instrumentation__.Notify(626913)
			break
		} else {
			__antithesis_instrumentation__.Notify(626914)
		}
		__antithesis_instrumentation__.Notify(626911)

		ts = e.lastRefreshTimestamp
	}
	__antithesis_instrumentation__.Notify(626901)

	e.stats, e.err = stats, err
	e.refreshing = false

	if err != nil {
		__antithesis_instrumentation__.Notify(626915)

		sc.mu.cache.Del(tableID)
	} else {
		__antithesis_instrumentation__.Notify(626916)
	}
}

func (sc *TableStatisticsCache) refreshTableStats(
	ctx context.Context, tableID descpb.ID, ts hlc.Timestamp,
) {
	__antithesis_instrumentation__.Notify(626917)
	log.VEventf(ctx, 1, "refreshing statistics for table %d", tableID)
	ctx, span := tracing.ForkSpan(ctx, "refresh-table-stats")

	go func() {
		__antithesis_instrumentation__.Notify(626918)
		defer span.Finish()
		sc.refreshCacheEntry(ctx, tableID, ts)
	}()
}

func (sc *TableStatisticsCache) InvalidateTableStats(ctx context.Context, tableID descpb.ID) {
	__antithesis_instrumentation__.Notify(626919)
	log.VEventf(ctx, 1, "evicting statistics for table %d", tableID)
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.mu.cache.Del(tableID)
}

const (
	tableIDIndex = iota
	statisticsIDIndex
	nameIndex
	columnIDsIndex
	createdAtIndex
	rowCountIndex
	distinctCountIndex
	nullCountIndex
	avgSizeIndex
	histogramIndex
	statsLen
)

func (sc *TableStatisticsCache) parseStats(
	ctx context.Context, datums tree.Datums, avgSizeColVerActive bool,
) (*TableStatistic, error) {
	__antithesis_instrumentation__.Notify(626920)
	if datums == nil || func() bool {
		__antithesis_instrumentation__.Notify(626930)
		return datums.Len() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(626931)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(626932)
	}
	__antithesis_instrumentation__.Notify(626921)

	hgIndex := histogramIndex
	numStats := statsLen
	if !avgSizeColVerActive {
		__antithesis_instrumentation__.Notify(626933)
		hgIndex = histogramIndex - 1
		numStats = statsLen - 1
	} else {
		__antithesis_instrumentation__.Notify(626934)
	}
	__antithesis_instrumentation__.Notify(626922)

	if datums.Len() != numStats {
		__antithesis_instrumentation__.Notify(626935)
		return nil, errors.Errorf("%d values returned from table statistics lookup. Expected %d", datums.Len(), numStats)
	} else {
		__antithesis_instrumentation__.Notify(626936)
	}
	__antithesis_instrumentation__.Notify(626923)

	expectedTypes := []struct {
		fieldName    string
		fieldIndex   int
		expectedType *types.T
		nullable     bool
	}{
		{"tableID", tableIDIndex, types.Int, false},
		{"statisticsID", statisticsIDIndex, types.Int, false},
		{"name", nameIndex, types.String, true},
		{"columnIDs", columnIDsIndex, types.IntArray, false},
		{"createdAt", createdAtIndex, types.Timestamp, false},
		{"rowCount", rowCountIndex, types.Int, false},
		{"distinctCount", distinctCountIndex, types.Int, false},
		{"nullCount", nullCountIndex, types.Int, false},
		{"histogram", hgIndex, types.Bytes, true},
	}

	if avgSizeColVerActive {
		__antithesis_instrumentation__.Notify(626937)
		expectedTypes = append(expectedTypes,
			struct {
				fieldName    string
				fieldIndex   int
				expectedType *types.T
				nullable     bool
			}{
				"avgSize", avgSizeIndex, types.Int, false,
			},
		)
	} else {
		__antithesis_instrumentation__.Notify(626938)
	}
	__antithesis_instrumentation__.Notify(626924)
	for _, v := range expectedTypes {
		__antithesis_instrumentation__.Notify(626939)
		if !datums[v.fieldIndex].ResolvedType().Equivalent(v.expectedType) && func() bool {
			__antithesis_instrumentation__.Notify(626940)
			return (!v.nullable || func() bool {
				__antithesis_instrumentation__.Notify(626941)
				return datums[v.fieldIndex].ResolvedType().Family() != types.UnknownFamily == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(626942)
			return nil, errors.Errorf("%s returned from table statistics lookup has type %s. Expected %s",
				v.fieldName, datums[v.fieldIndex].ResolvedType(), v.expectedType)
		} else {
			__antithesis_instrumentation__.Notify(626943)
		}
	}
	__antithesis_instrumentation__.Notify(626925)

	res := &TableStatistic{
		TableStatisticProto: TableStatisticProto{
			TableID:       descpb.ID((int32)(*datums[tableIDIndex].(*tree.DInt))),
			StatisticID:   (uint64)(*datums[statisticsIDIndex].(*tree.DInt)),
			CreatedAt:     datums[createdAtIndex].(*tree.DTimestamp).Time,
			RowCount:      (uint64)(*datums[rowCountIndex].(*tree.DInt)),
			DistinctCount: (uint64)(*datums[distinctCountIndex].(*tree.DInt)),
			NullCount:     (uint64)(*datums[nullCountIndex].(*tree.DInt)),
		},
	}
	if avgSizeColVerActive {
		__antithesis_instrumentation__.Notify(626944)
		res.AvgSize = (uint64)(*datums[avgSizeIndex].(*tree.DInt))
	} else {
		__antithesis_instrumentation__.Notify(626945)
	}
	__antithesis_instrumentation__.Notify(626926)
	columnIDs := datums[columnIDsIndex].(*tree.DArray)
	res.ColumnIDs = make([]descpb.ColumnID, len(columnIDs.Array))
	for i, d := range columnIDs.Array {
		__antithesis_instrumentation__.Notify(626946)
		res.ColumnIDs[i] = descpb.ColumnID((int32)(*d.(*tree.DInt)))
	}
	__antithesis_instrumentation__.Notify(626927)
	if datums[nameIndex] != tree.DNull {
		__antithesis_instrumentation__.Notify(626947)
		res.Name = string(*datums[nameIndex].(*tree.DString))
	} else {
		__antithesis_instrumentation__.Notify(626948)
	}
	__antithesis_instrumentation__.Notify(626928)
	if datums[hgIndex] != tree.DNull {
		__antithesis_instrumentation__.Notify(626949)
		res.HistogramData = &HistogramData{}
		if err := protoutil.Unmarshal(
			[]byte(*datums[hgIndex].(*tree.DBytes)),
			res.HistogramData,
		); err != nil {
			__antithesis_instrumentation__.Notify(626952)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(626953)
		}
		__antithesis_instrumentation__.Notify(626950)

		if typ := res.HistogramData.ColumnType; typ != nil && func() bool {
			__antithesis_instrumentation__.Notify(626954)
			return typ.UserDefined() == true
		}() == true {
			__antithesis_instrumentation__.Notify(626955)

			if err := sc.collectionFactory.Txn(ctx, sc.SQLExecutor, sc.ClientDB, func(
				ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
			) error {
				__antithesis_instrumentation__.Notify(626956)
				resolver := descs.NewDistSQLTypeResolver(descriptors, txn)
				var err error
				res.HistogramData.ColumnType, err = resolver.ResolveTypeByOID(ctx, typ.Oid())
				return err
			}); err != nil {
				__antithesis_instrumentation__.Notify(626957)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(626958)
			}
		} else {
			__antithesis_instrumentation__.Notify(626959)
		}
		__antithesis_instrumentation__.Notify(626951)
		if err := DecodeHistogramBuckets(res); err != nil {
			__antithesis_instrumentation__.Notify(626960)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(626961)
		}
	} else {
		__antithesis_instrumentation__.Notify(626962)
	}
	__antithesis_instrumentation__.Notify(626929)

	return res, nil
}

func DecodeHistogramBuckets(tabStat *TableStatistic) error {
	__antithesis_instrumentation__.Notify(626963)
	var offset int
	if tabStat.NullCount > 0 {
		__antithesis_instrumentation__.Notify(626966)

		tabStat.Histogram = make([]cat.HistogramBucket, len(tabStat.HistogramData.Buckets)+1)
		tabStat.Histogram[0] = cat.HistogramBucket{
			NumEq:         float64(tabStat.NullCount),
			NumRange:      0,
			DistinctRange: 0,
			UpperBound:    tree.DNull,
		}
		offset = 1
	} else {
		__antithesis_instrumentation__.Notify(626967)
		tabStat.Histogram = make([]cat.HistogramBucket, len(tabStat.HistogramData.Buckets))
		offset = 0
	}
	__antithesis_instrumentation__.Notify(626964)

	var a tree.DatumAlloc
	for i := offset; i < len(tabStat.Histogram); i++ {
		__antithesis_instrumentation__.Notify(626968)
		bucket := &tabStat.HistogramData.Buckets[i-offset]
		datum, _, err := keyside.Decode(&a, tabStat.HistogramData.ColumnType, bucket.UpperBound, encoding.Ascending)
		if err != nil {
			__antithesis_instrumentation__.Notify(626970)
			return err
		} else {
			__antithesis_instrumentation__.Notify(626971)
		}
		__antithesis_instrumentation__.Notify(626969)
		tabStat.Histogram[i] = cat.HistogramBucket{
			NumEq:         float64(bucket.NumEq),
			NumRange:      float64(bucket.NumRange),
			DistinctRange: bucket.DistinctRange,
			UpperBound:    datum,
		}
	}
	__antithesis_instrumentation__.Notify(626965)
	return nil
}

func (sc *TableStatisticsCache) getTableStatsFromDB(
	ctx context.Context, tableID descpb.ID,
) ([]*TableStatistic, error) {
	__antithesis_instrumentation__.Notify(626972)
	avgSizeColVerActive := sc.Settings.Version.IsActive(ctx, clusterversion.AlterSystemTableStatisticsAddAvgSizeCol)
	var avgSize string
	if avgSizeColVerActive {
		__antithesis_instrumentation__.Notify(626977)
		avgSize = `
					"avgSize",`
	} else {
		__antithesis_instrumentation__.Notify(626978)
	}
	__antithesis_instrumentation__.Notify(626973)
	getTableStatisticsStmt := fmt.Sprintf(`
SELECT
  "tableID",
	"statisticID",
	name,
	"columnIDs",
	"createdAt",
	"rowCount",
	"distinctCount",
	"nullCount",
	%s
	histogram
FROM system.table_statistics
WHERE "tableID" = $1
ORDER BY "createdAt" DESC
`, avgSize)

	it, err := sc.SQLExecutor.QueryIterator(
		ctx, "get-table-statistics", nil, getTableStatisticsStmt, tableID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(626979)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(626980)
	}
	__antithesis_instrumentation__.Notify(626974)

	var statsList []*TableStatistic
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(626981)
		stats, err := sc.parseStats(ctx, it.Cur(), avgSizeColVerActive)
		if err != nil {
			__antithesis_instrumentation__.Notify(626983)
			log.Warningf(ctx, "could not decode statistic for table %d: %v", tableID, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(626984)
		}
		__antithesis_instrumentation__.Notify(626982)
		statsList = append(statsList, stats)
	}
	__antithesis_instrumentation__.Notify(626975)
	if err != nil {
		__antithesis_instrumentation__.Notify(626985)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(626986)
	}
	__antithesis_instrumentation__.Notify(626976)

	return statsList, nil
}
