package reports

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

var ReporterInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.replication_reports.interval",
	"the frequency for generating the replication_constraint_stats, replication_stats_report and "+
		"replication_critical_localities reports (set to 0 to disable)",
	time.Minute,
	settings.NonNegativeDuration,
).WithPublic()

type Reporter struct {
	localStores *kvserver.Stores

	meta1LeaseHolder *kvserver.Store

	latestConfig *config.SystemConfig

	db        *kv.DB
	liveness  *liveness.NodeLiveness
	settings  *cluster.Settings
	storePool *kvserver.StorePool
	executor  sqlutil.InternalExecutor
	cfgs      config.SystemConfigProvider

	frequencyMu struct {
		syncutil.Mutex
		interval time.Duration
		changeCh chan struct{}
	}
}

func NewReporter(
	db *kv.DB,
	localStores *kvserver.Stores,
	storePool *kvserver.StorePool,
	st *cluster.Settings,
	liveness *liveness.NodeLiveness,
	executor sqlutil.InternalExecutor,
	provider config.SystemConfigProvider,
) *Reporter {
	__antithesis_instrumentation__.Notify(121894)
	r := Reporter{
		db:          db,
		localStores: localStores,
		storePool:   storePool,
		settings:    st,
		liveness:    liveness,
		executor:    executor,
		cfgs:        provider,
	}
	r.frequencyMu.changeCh = make(chan struct{})
	return &r
}

func (stats *Reporter) reportInterval() (time.Duration, <-chan struct{}) {
	__antithesis_instrumentation__.Notify(121895)
	stats.frequencyMu.Lock()
	defer stats.frequencyMu.Unlock()
	return ReporterInterval.Get(&stats.settings.SV), stats.frequencyMu.changeCh
}

func (stats *Reporter) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(121896)
	ReporterInterval.SetOnChange(&stats.settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(121898)
		stats.frequencyMu.Lock()
		defer stats.frequencyMu.Unlock()

		ch := stats.frequencyMu.changeCh
		close(ch)
		stats.frequencyMu.changeCh = make(chan struct{})
		stats.frequencyMu.interval = ReporterInterval.Get(&stats.settings.SV)
	})
	__antithesis_instrumentation__.Notify(121897)
	_ = stopper.RunAsyncTask(ctx, "stats-reporter", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(121899)
		var timer timeutil.Timer
		defer timer.Stop()
		ctx = logtags.AddTag(ctx, "replication-reporter", nil)

		replStatsSaver := makeReplicationStatsReportSaver()
		constraintsSaver := makeReplicationConstraintStatusReportSaver()
		criticalLocSaver := makeReplicationCriticalLocalitiesReportSaver()

		for {
			__antithesis_instrumentation__.Notify(121900)

			interval, changeCh := stats.reportInterval()

			var timerCh <-chan time.Time
			if interval != 0 {
				__antithesis_instrumentation__.Notify(121902)

				stats.meta1LeaseHolder = stats.meta1LeaseHolderStore(ctx)
				if stats.meta1LeaseHolder != nil {
					__antithesis_instrumentation__.Notify(121904)
					if err := stats.update(
						ctx, &constraintsSaver, &replStatsSaver, &criticalLocSaver,
					); err != nil {
						__antithesis_instrumentation__.Notify(121905)
						log.Errorf(ctx, "failed to generate replication reports: %s", err)
					} else {
						__antithesis_instrumentation__.Notify(121906)
					}
				} else {
					__antithesis_instrumentation__.Notify(121907)
				}
				__antithesis_instrumentation__.Notify(121903)
				timer.Reset(interval)
				timerCh = timer.C
			} else {
				__antithesis_instrumentation__.Notify(121908)
			}
			__antithesis_instrumentation__.Notify(121901)

			select {
			case <-timerCh:
				__antithesis_instrumentation__.Notify(121909)
				timer.Read = true
			case <-changeCh:
				__antithesis_instrumentation__.Notify(121910)
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(121911)
				return
			}
		}
	})
}

func (stats *Reporter) update(
	ctx context.Context,
	constraintsSaver *replicationConstraintStatsReportSaver,
	replStatsSaver *replicationStatsReportSaver,
	locSaver *replicationCriticalLocalitiesReportSaver,
) error {
	__antithesis_instrumentation__.Notify(121912)
	start := timeutil.Now()
	log.VEventf(ctx, 2, "updating replication reports...")
	defer func() {
		__antithesis_instrumentation__.Notify(121922)
		log.VEventf(ctx, 2, "updating replication reports... done. Generation took: %s.",
			timeutil.Since(start))
	}()
	__antithesis_instrumentation__.Notify(121913)
	stats.updateLatestConfig()
	if stats.latestConfig == nil {
		__antithesis_instrumentation__.Notify(121923)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(121924)
	}
	__antithesis_instrumentation__.Notify(121914)

	allStores := stats.storePool.GetStores()
	var getStoresFromGossip StoreResolver = func(
		r *roachpb.RangeDescriptor,
	) []roachpb.StoreDescriptor {
		__antithesis_instrumentation__.Notify(121925)
		storeDescs := make([]roachpb.StoreDescriptor, len(r.Replicas().VoterDescriptors()))

		for i, repl := range r.Replicas().VoterDescriptors() {
			__antithesis_instrumentation__.Notify(121927)
			storeDescs[i] = allStores[repl.StoreID]
		}
		__antithesis_instrumentation__.Notify(121926)
		return storeDescs
	}
	__antithesis_instrumentation__.Notify(121915)

	isLiveMap := stats.liveness.GetIsLiveMap()
	isNodeLive := func(nodeID roachpb.NodeID) bool {
		__antithesis_instrumentation__.Notify(121928)
		return isLiveMap[nodeID].IsLive
	}
	__antithesis_instrumentation__.Notify(121916)

	nodeLocalities := make(map[roachpb.NodeID]roachpb.Locality, len(allStores))
	for _, storeDesc := range allStores {
		__antithesis_instrumentation__.Notify(121929)
		nodeDesc := storeDesc.Node

		nodeLocalities[nodeDesc.NodeID] = nodeDesc.Locality
	}
	__antithesis_instrumentation__.Notify(121917)

	constraintConfVisitor := makeConstraintConformanceVisitor(
		ctx, stats.latestConfig, getStoresFromGossip)
	localityStatsVisitor := makeCriticalLocalitiesVisitor(
		ctx, nodeLocalities, stats.latestConfig,
		getStoresFromGossip, isNodeLive)
	replicationStatsVisitor := makeReplicationStatsVisitor(ctx, stats.latestConfig, isNodeLive)

	const descriptorReadBatchSize = 10000
	rangeIter := makeMeta2RangeIter(stats.db, descriptorReadBatchSize)
	if err := visitRanges(
		ctx, &rangeIter, stats.latestConfig,
		&constraintConfVisitor, &localityStatsVisitor, &replicationStatsVisitor,
	); err != nil {
		__antithesis_instrumentation__.Notify(121930)
		if errors.HasType(err, (*visitorError)(nil)) {
			__antithesis_instrumentation__.Notify(121931)
			log.Errorf(ctx, "some reports have not been generated: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(121932)
			return errors.Wrap(err, "failed to compute constraint conformance report")
		}
	} else {
		__antithesis_instrumentation__.Notify(121933)
	}
	__antithesis_instrumentation__.Notify(121918)

	if !constraintConfVisitor.failed() {
		__antithesis_instrumentation__.Notify(121934)
		if err := constraintsSaver.Save(
			ctx, constraintConfVisitor.report, timeutil.Now(), stats.db, stats.executor,
		); err != nil {
			__antithesis_instrumentation__.Notify(121935)
			return errors.Wrap(err, "failed to save constraint report")
		} else {
			__antithesis_instrumentation__.Notify(121936)
		}
	} else {
		__antithesis_instrumentation__.Notify(121937)
	}
	__antithesis_instrumentation__.Notify(121919)
	if !localityStatsVisitor.failed() {
		__antithesis_instrumentation__.Notify(121938)
		if err := locSaver.Save(
			ctx, localityStatsVisitor.Report(), timeutil.Now(), stats.db, stats.executor,
		); err != nil {
			__antithesis_instrumentation__.Notify(121939)
			return errors.Wrap(err, "failed to save locality report")
		} else {
			__antithesis_instrumentation__.Notify(121940)
		}
	} else {
		__antithesis_instrumentation__.Notify(121941)
	}
	__antithesis_instrumentation__.Notify(121920)
	if !replicationStatsVisitor.failed() {
		__antithesis_instrumentation__.Notify(121942)
		if err := replStatsSaver.Save(
			ctx, replicationStatsVisitor.Report(),
			timeutil.Now(), stats.db, stats.executor,
		); err != nil {
			__antithesis_instrumentation__.Notify(121943)
			return errors.Wrap(err, "failed to save range status report")
		} else {
			__antithesis_instrumentation__.Notify(121944)
		}
	} else {
		__antithesis_instrumentation__.Notify(121945)
	}
	__antithesis_instrumentation__.Notify(121921)
	return nil
}

func (stats *Reporter) meta1LeaseHolderStore(ctx context.Context) *kvserver.Store {
	__antithesis_instrumentation__.Notify(121946)
	const meta1RangeID = roachpb.RangeID(1)
	repl, store, err := stats.localStores.GetReplicaForRangeID(ctx, meta1RangeID)
	if roachpb.IsRangeNotFoundError(err) {
		__antithesis_instrumentation__.Notify(121950)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(121951)
	}
	__antithesis_instrumentation__.Notify(121947)
	if err != nil {
		__antithesis_instrumentation__.Notify(121952)
		log.Fatalf(ctx, "unexpected error when visiting stores: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(121953)
	}
	__antithesis_instrumentation__.Notify(121948)
	if repl.OwnsValidLease(ctx, store.Clock().NowAsClockTimestamp()) {
		__antithesis_instrumentation__.Notify(121954)
		return store
	} else {
		__antithesis_instrumentation__.Notify(121955)
	}
	__antithesis_instrumentation__.Notify(121949)
	return nil
}

func (stats *Reporter) updateLatestConfig() {
	__antithesis_instrumentation__.Notify(121956)
	stats.latestConfig = stats.cfgs.GetSystemConfig()
}

type nodeChecker func(nodeID roachpb.NodeID) bool

type zoneResolver struct {
	init bool

	curObjectID config.ObjectID

	curRootZone *zonepb.ZoneConfig

	curZoneKey ZoneKey
}

func (c *zoneResolver) resolveRange(
	ctx context.Context, rng *roachpb.RangeDescriptor, cfg *config.SystemConfig,
) (ZoneKey, error) {
	__antithesis_instrumentation__.Notify(121957)
	if c.checkSameZone(ctx, rng) {
		__antithesis_instrumentation__.Notify(121959)
		return c.curZoneKey, nil
	} else {
		__antithesis_instrumentation__.Notify(121960)
	}
	__antithesis_instrumentation__.Notify(121958)
	return c.updateZone(ctx, rng, cfg)
}

func (c *zoneResolver) setZone(objectID config.ObjectID, key ZoneKey, rootZone *zonepb.ZoneConfig) {
	__antithesis_instrumentation__.Notify(121961)
	c.init = true
	c.curObjectID = objectID
	c.curRootZone = rootZone
	c.curZoneKey = key
}

func (c *zoneResolver) updateZone(
	ctx context.Context, rd *roachpb.RangeDescriptor, cfg *config.SystemConfig,
) (ZoneKey, error) {
	__antithesis_instrumentation__.Notify(121962)
	objectID, _ := config.DecodeKeyIntoZoneIDAndSuffix(keys.SystemSQLCodec, rd.StartKey)
	first := true
	var zoneKey ZoneKey
	var rootZone *zonepb.ZoneConfig

	found, err := visitZones(
		ctx, rd, cfg, includeSubzonePlaceholders,
		func(_ context.Context, zone *zonepb.ZoneConfig, key ZoneKey) bool {
			__antithesis_instrumentation__.Notify(121966)
			if first {
				__antithesis_instrumentation__.Notify(121969)
				first = false
				zoneKey = key
			} else {
				__antithesis_instrumentation__.Notify(121970)
			}
			__antithesis_instrumentation__.Notify(121967)
			if key.SubzoneID == NoSubzone {
				__antithesis_instrumentation__.Notify(121971)
				rootZone = zone
				return true
			} else {
				__antithesis_instrumentation__.Notify(121972)
			}
			__antithesis_instrumentation__.Notify(121968)
			return false
		})
	__antithesis_instrumentation__.Notify(121963)
	if err != nil {
		__antithesis_instrumentation__.Notify(121973)
		return ZoneKey{}, err
	} else {
		__antithesis_instrumentation__.Notify(121974)
	}
	__antithesis_instrumentation__.Notify(121964)
	if !found {
		__antithesis_instrumentation__.Notify(121975)
		return ZoneKey{}, errors.AssertionFailedf("failed to resolve zone for range: %s", rd)
	} else {
		__antithesis_instrumentation__.Notify(121976)
	}
	__antithesis_instrumentation__.Notify(121965)
	c.setZone(objectID, zoneKey, rootZone)
	return zoneKey, nil
}

func (c *zoneResolver) checkSameZone(ctx context.Context, rng *roachpb.RangeDescriptor) bool {
	__antithesis_instrumentation__.Notify(121977)
	if !c.init {
		__antithesis_instrumentation__.Notify(121980)
		return false
	} else {
		__antithesis_instrumentation__.Notify(121981)
	}
	__antithesis_instrumentation__.Notify(121978)

	objectID, keySuffix := config.DecodeKeyIntoZoneIDAndSuffix(keys.SystemSQLCodec, rng.StartKey)
	if objectID != c.curObjectID {
		__antithesis_instrumentation__.Notify(121982)
		return false
	} else {
		__antithesis_instrumentation__.Notify(121983)
	}
	__antithesis_instrumentation__.Notify(121979)
	_, subzoneIdx := c.curRootZone.GetSubzoneForKeySuffix(keySuffix)
	return subzoneIdx == c.curZoneKey.SubzoneID.ToSubzoneIndex()
}

type visitOpt bool

const (
	ignoreSubzonePlaceholders  visitOpt = false
	includeSubzonePlaceholders visitOpt = true
)

func visitZones(
	ctx context.Context,
	rng *roachpb.RangeDescriptor,
	cfg *config.SystemConfig,
	opt visitOpt,
	visitor func(context.Context, *zonepb.ZoneConfig, ZoneKey) bool,
) (bool, error) {
	__antithesis_instrumentation__.Notify(121984)
	id, keySuffix := config.DecodeKeyIntoZoneIDAndSuffix(keys.SystemSQLCodec, rng.StartKey)
	zone, err := getZoneByID(id, cfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(121987)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121988)
	}
	__antithesis_instrumentation__.Notify(121985)

	if zone != nil {
		__antithesis_instrumentation__.Notify(121989)

		subzone, subzoneIdx := zone.GetSubzoneForKeySuffix(keySuffix)
		if subzone != nil {
			__antithesis_instrumentation__.Notify(121991)
			if visitor(ctx, &subzone.Config, MakeZoneKey(id, base.SubzoneIDFromIndex(int(subzoneIdx)))) {
				__antithesis_instrumentation__.Notify(121992)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(121993)
			}
		} else {
			__antithesis_instrumentation__.Notify(121994)
		}
		__antithesis_instrumentation__.Notify(121990)

		if (opt == includeSubzonePlaceholders) || func() bool {
			__antithesis_instrumentation__.Notify(121995)
			return !zone.IsSubzonePlaceholder() == true
		}() == true {
			__antithesis_instrumentation__.Notify(121996)
			if visitor(ctx, zone, MakeZoneKey(id, 0)) {
				__antithesis_instrumentation__.Notify(121997)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(121998)
			}
		} else {
			__antithesis_instrumentation__.Notify(121999)
		}
	} else {
		__antithesis_instrumentation__.Notify(122000)
	}
	__antithesis_instrumentation__.Notify(121986)

	return visitAncestors(ctx, id, cfg, visitor)
}

func visitAncestors(
	ctx context.Context,
	id config.ObjectID,
	cfg *config.SystemConfig,
	visitor func(context.Context, *zonepb.ZoneConfig, ZoneKey) bool,
) (bool, error) {
	__antithesis_instrumentation__.Notify(122001)

	descVal := cfg.GetValue(catalogkeys.MakeDescMetadataKey(keys.TODOSQLCodec, descpb.ID(id)))
	if descVal == nil {
		__antithesis_instrumentation__.Notify(122007)

		return visitDefaultZone(ctx, cfg, visitor), nil
	} else {
		__antithesis_instrumentation__.Notify(122008)
	}
	__antithesis_instrumentation__.Notify(122002)

	var desc descpb.Descriptor
	if err := descVal.GetProto(&desc); err != nil {
		__antithesis_instrumentation__.Notify(122009)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(122010)
	}
	__antithesis_instrumentation__.Notify(122003)
	tableDesc, _, _, _ := descpb.FromDescriptorWithMVCCTimestamp(&desc, descVal.Timestamp)

	if tableDesc == nil {
		__antithesis_instrumentation__.Notify(122011)
		return visitDefaultZone(ctx, cfg, visitor), nil
	} else {
		__antithesis_instrumentation__.Notify(122012)
	}
	__antithesis_instrumentation__.Notify(122004)

	zone, err := getZoneByID(config.ObjectID(tableDesc.ParentID), cfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(122013)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(122014)
	}
	__antithesis_instrumentation__.Notify(122005)
	if zone != nil {
		__antithesis_instrumentation__.Notify(122015)
		if visitor(ctx, zone, MakeZoneKey(config.ObjectID(tableDesc.ParentID), NoSubzone)) {
			__antithesis_instrumentation__.Notify(122016)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(122017)
		}
	} else {
		__antithesis_instrumentation__.Notify(122018)
	}
	__antithesis_instrumentation__.Notify(122006)

	return visitDefaultZone(ctx, cfg, visitor), nil
}

func visitDefaultZone(
	ctx context.Context,
	cfg *config.SystemConfig,
	visitor func(context.Context, *zonepb.ZoneConfig, ZoneKey) bool,
) bool {
	__antithesis_instrumentation__.Notify(122019)
	zone, err := getZoneByID(keys.RootNamespaceID, cfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(122022)
		log.Fatalf(ctx, "failed to get default zone config: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(122023)
	}
	__antithesis_instrumentation__.Notify(122020)
	if zone == nil {
		__antithesis_instrumentation__.Notify(122024)
		log.Fatal(ctx, "default zone config missing unexpectedly")
	} else {
		__antithesis_instrumentation__.Notify(122025)
	}
	__antithesis_instrumentation__.Notify(122021)
	return visitor(ctx, zone, MakeZoneKey(keys.RootNamespaceID, NoSubzone))
}

func getZoneByID(id config.ObjectID, cfg *config.SystemConfig) (*zonepb.ZoneConfig, error) {
	__antithesis_instrumentation__.Notify(122026)
	zoneVal := cfg.GetValue(config.MakeZoneKey(keys.SystemSQLCodec, descpb.ID(id)))
	if zoneVal == nil {
		__antithesis_instrumentation__.Notify(122029)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(122030)
	}
	__antithesis_instrumentation__.Notify(122027)
	zone := new(zonepb.ZoneConfig)
	if err := zoneVal.GetProto(zone); err != nil {
		__antithesis_instrumentation__.Notify(122031)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(122032)
	}
	__antithesis_instrumentation__.Notify(122028)
	return zone, nil
}

type StoreResolver func(*roachpb.RangeDescriptor) []roachpb.StoreDescriptor

type rangeVisitor interface {
	visitNewZone(context.Context, *roachpb.RangeDescriptor) error
	visitSameZone(context.Context, *roachpb.RangeDescriptor)

	failed() bool

	reset(ctx context.Context)
}

type visitorError struct {
	errs []error
}

func (e *visitorError) Error() string {
	__antithesis_instrumentation__.Notify(122033)
	s := make([]string, len(e.errs))
	for i, err := range e.errs {
		__antithesis_instrumentation__.Notify(122035)
		s[i] = fmt.Sprintf("%d: %s", i, err)
	}
	__antithesis_instrumentation__.Notify(122034)
	return fmt.Sprintf("%d visitors encountered errors:\n%s", len(e.errs), strings.Join(s, "\n"))
}

func visitRanges(
	ctx context.Context, rangeStore RangeIterator, cfg *config.SystemConfig, visitors ...rangeVisitor,
) error {
	__antithesis_instrumentation__.Notify(122036)
	origVisitors := make([]rangeVisitor, len(visitors))
	copy(origVisitors, visitors)
	var visitorErrs []error
	var resolver zoneResolver

	var key ZoneKey
	first := true

	for {
		__antithesis_instrumentation__.Notify(122039)
		rd, err := rangeStore.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(122043)
			if errIsRetriable(err) {
				__antithesis_instrumentation__.Notify(122044)
				visitors = origVisitors
				for _, v := range visitors {
					__antithesis_instrumentation__.Notify(122046)
					v.reset(ctx)
				}
				__antithesis_instrumentation__.Notify(122045)

				continue
			} else {
				__antithesis_instrumentation__.Notify(122047)
				return err
			}
		} else {
			__antithesis_instrumentation__.Notify(122048)
		}
		__antithesis_instrumentation__.Notify(122040)
		if rd.RangeID == 0 {
			__antithesis_instrumentation__.Notify(122049)

			break
		} else {
			__antithesis_instrumentation__.Notify(122050)
		}
		__antithesis_instrumentation__.Notify(122041)

		newKey, err := resolver.resolveRange(ctx, &rd, cfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(122051)
			return err
		} else {
			__antithesis_instrumentation__.Notify(122052)
		}
		__antithesis_instrumentation__.Notify(122042)
		sameZoneAsPrevRange := !first && func() bool {
			__antithesis_instrumentation__.Notify(122053)
			return key == newKey == true
		}() == true
		key = newKey
		first = false

		for i, v := range visitors {
			__antithesis_instrumentation__.Notify(122054)
			var err error
			if sameZoneAsPrevRange {
				__antithesis_instrumentation__.Notify(122056)
				v.visitSameZone(ctx, &rd)
			} else {
				__antithesis_instrumentation__.Notify(122057)
				err = v.visitNewZone(ctx, &rd)
			}
			__antithesis_instrumentation__.Notify(122055)

			if err != nil {
				__antithesis_instrumentation__.Notify(122058)

				if !v.failed() {
					__antithesis_instrumentation__.Notify(122060)
					return errors.NewAssertionErrorWithWrappedErrf(err, "expected visitor %T to have failed() after error", v)
				} else {
					__antithesis_instrumentation__.Notify(122061)
				}
				__antithesis_instrumentation__.Notify(122059)

				visitors = append(visitors[:i], visitors[i+1:]...)
				visitorErrs = append(visitorErrs, err)
			} else {
				__antithesis_instrumentation__.Notify(122062)
			}
		}
	}
	__antithesis_instrumentation__.Notify(122037)
	if len(visitorErrs) > 0 {
		__antithesis_instrumentation__.Notify(122063)
		return &visitorError{errs: visitorErrs}
	} else {
		__antithesis_instrumentation__.Notify(122064)
	}
	__antithesis_instrumentation__.Notify(122038)
	return nil
}

type RangeIterator interface {
	Next(context.Context) (roachpb.RangeDescriptor, error)

	Close(context.Context)
}

type meta2RangeIter struct {
	db *kv.DB

	batchSize int

	txn *kv.Txn

	buffer []kv.KeyValue

	resumeSpan *roachpb.Span

	readingDone bool
}

func makeMeta2RangeIter(db *kv.DB, batchSize int) meta2RangeIter {
	__antithesis_instrumentation__.Notify(122065)
	return meta2RangeIter{db: db, batchSize: batchSize}
}

var _ RangeIterator = &meta2RangeIter{}

func (r *meta2RangeIter) Next(ctx context.Context) (_ roachpb.RangeDescriptor, retErr error) {
	__antithesis_instrumentation__.Notify(122066)
	defer func() { __antithesis_instrumentation__.Notify(122071); r.handleErr(ctx, retErr) }()
	__antithesis_instrumentation__.Notify(122067)

	rd, err := r.consumerBuffer()
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(122072)
		return rd.RangeID != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(122073)
		return rd, err
	} else {
		__antithesis_instrumentation__.Notify(122074)
	}
	__antithesis_instrumentation__.Notify(122068)

	if r.readingDone {
		__antithesis_instrumentation__.Notify(122075)

		return roachpb.RangeDescriptor{}, nil
	} else {
		__antithesis_instrumentation__.Notify(122076)
	}
	__antithesis_instrumentation__.Notify(122069)

	if err := r.readBatch(ctx); err != nil {
		__antithesis_instrumentation__.Notify(122077)
		return roachpb.RangeDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(122078)
	}
	__antithesis_instrumentation__.Notify(122070)
	return r.consumerBuffer()
}

func (r *meta2RangeIter) consumerBuffer() (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(122079)
	if len(r.buffer) == 0 {
		__antithesis_instrumentation__.Notify(122082)
		return roachpb.RangeDescriptor{}, nil
	} else {
		__antithesis_instrumentation__.Notify(122083)
	}
	__antithesis_instrumentation__.Notify(122080)
	first := r.buffer[0]
	var desc roachpb.RangeDescriptor
	if err := first.ValueProto(&desc); err != nil {
		__antithesis_instrumentation__.Notify(122084)
		return roachpb.RangeDescriptor{}, errors.NewAssertionErrorWithWrappedErrf(err,
			"%s: unable to unmarshal range descriptor", first.Key)
	} else {
		__antithesis_instrumentation__.Notify(122085)
	}
	__antithesis_instrumentation__.Notify(122081)
	r.buffer = r.buffer[1:]
	return desc, nil
}

func (r *meta2RangeIter) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(122086)
	if r.readingDone {
		__antithesis_instrumentation__.Notify(122088)
		return
	} else {
		__antithesis_instrumentation__.Notify(122089)
	}
	__antithesis_instrumentation__.Notify(122087)
	_ = r.txn.Rollback(ctx)
	r.txn = nil
	r.readingDone = true
}

func (r *meta2RangeIter) readBatch(ctx context.Context) (retErr error) {
	__antithesis_instrumentation__.Notify(122090)
	defer func() { __antithesis_instrumentation__.Notify(122097); r.handleErr(ctx, retErr) }()
	__antithesis_instrumentation__.Notify(122091)

	if len(r.buffer) > 0 {
		__antithesis_instrumentation__.Notify(122098)
		log.Fatalf(ctx, "buffer not exhausted: %d keys remaining", len(r.buffer))
	} else {
		__antithesis_instrumentation__.Notify(122099)
	}
	__antithesis_instrumentation__.Notify(122092)
	if r.txn == nil {
		__antithesis_instrumentation__.Notify(122100)
		r.txn = r.db.NewTxn(ctx, "rangeStoreImpl")
	} else {
		__antithesis_instrumentation__.Notify(122101)
	}
	__antithesis_instrumentation__.Notify(122093)

	b := r.txn.NewBatch()
	start := keys.Meta2Prefix
	if r.resumeSpan != nil {
		__antithesis_instrumentation__.Notify(122102)
		start = r.resumeSpan.Key
	} else {
		__antithesis_instrumentation__.Notify(122103)
	}
	__antithesis_instrumentation__.Notify(122094)
	b.Scan(start, keys.MetaMax)
	b.Header.MaxSpanRequestKeys = int64(r.batchSize)
	err := r.txn.Run(ctx, b)
	if err != nil {
		__antithesis_instrumentation__.Notify(122104)
		return err
	} else {
		__antithesis_instrumentation__.Notify(122105)
	}
	__antithesis_instrumentation__.Notify(122095)
	r.buffer = b.Results[0].Rows
	r.resumeSpan = b.Results[0].ResumeSpan
	if r.resumeSpan == nil {
		__antithesis_instrumentation__.Notify(122106)
		if err := r.txn.Commit(ctx); err != nil {
			__antithesis_instrumentation__.Notify(122108)
			return err
		} else {
			__antithesis_instrumentation__.Notify(122109)
		}
		__antithesis_instrumentation__.Notify(122107)
		r.txn = nil
		r.readingDone = true
	} else {
		__antithesis_instrumentation__.Notify(122110)
	}
	__antithesis_instrumentation__.Notify(122096)
	return nil
}

func errIsRetriable(err error) bool {
	__antithesis_instrumentation__.Notify(122111)
	return errors.HasType(err, (*roachpb.TransactionRetryWithProtoRefreshError)(nil))
}

func (r *meta2RangeIter) handleErr(ctx context.Context, err error) {
	__antithesis_instrumentation__.Notify(122112)
	if err == nil {
		__antithesis_instrumentation__.Notify(122114)
		return
	} else {
		__antithesis_instrumentation__.Notify(122115)
	}
	__antithesis_instrumentation__.Notify(122113)
	if !errIsRetriable(err) {
		__antithesis_instrumentation__.Notify(122116)
		if r.txn != nil {
			__antithesis_instrumentation__.Notify(122118)

			r.txn.CleanupOnError(ctx, err)
			r.txn = nil
		} else {
			__antithesis_instrumentation__.Notify(122119)
		}
		__antithesis_instrumentation__.Notify(122117)
		r.reset()
		r.readingDone = true
	} else {
		__antithesis_instrumentation__.Notify(122120)
		r.reset()
	}
}

func (r *meta2RangeIter) reset() {
	__antithesis_instrumentation__.Notify(122121)
	r.buffer = nil
	r.resumeSpan = nil
	r.readingDone = false
}

type reportID int

func getReportGenerationTime(
	ctx context.Context, rid reportID, ex sqlutil.InternalExecutor, txn *kv.Txn,
) (time.Time, error) {
	__antithesis_instrumentation__.Notify(122122)
	row, err := ex.QueryRowEx(
		ctx,
		"get-previous-timestamp",
		txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		"select generated from system.reports_meta where id = $1",
		rid,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(122127)
		return time.Time{}, err
	} else {
		__antithesis_instrumentation__.Notify(122128)
	}
	__antithesis_instrumentation__.Notify(122123)

	if row == nil {
		__antithesis_instrumentation__.Notify(122129)
		return time.Time{}, nil
	} else {
		__antithesis_instrumentation__.Notify(122130)
	}
	__antithesis_instrumentation__.Notify(122124)

	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(122131)
		return time.Time{}, errors.AssertionFailedf(
			"expected 1 column from intenal query, got: %d", len(row))
	} else {
		__antithesis_instrumentation__.Notify(122132)
	}
	__antithesis_instrumentation__.Notify(122125)
	generated, ok := row[0].(*tree.DTimestampTZ)
	if !ok {
		__antithesis_instrumentation__.Notify(122133)
		return time.Time{}, errors.AssertionFailedf("expected to get timestamptz from "+
			"system.reports_meta got %+v (%T)", row[0], row[0])
	} else {
		__antithesis_instrumentation__.Notify(122134)
	}
	__antithesis_instrumentation__.Notify(122126)
	return generated.Time, nil
}
