package spanconfigsqlwatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedbuffer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var _ spanconfig.SQLWatcher = &SQLWatcher{}

type SQLWatcher struct {
	codec                keys.SQLCodec
	settings             *cluster.Settings
	stopper              *stop.Stopper
	knobs                *spanconfig.TestingKnobs
	rangeFeedFactory     *rangefeed.Factory
	bufferMemLimit       int64
	checkpointNoopsEvery time.Duration
}

func New(
	codec keys.SQLCodec,
	settings *cluster.Settings,
	rangeFeedFactory *rangefeed.Factory,
	bufferMemLimit int64,
	stopper *stop.Stopper,
	checkpointNoopsEvery time.Duration,
	knobs *spanconfig.TestingKnobs,
) *SQLWatcher {
	__antithesis_instrumentation__.Notify(241346)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(241349)
		knobs = &spanconfig.TestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(241350)
	}
	__antithesis_instrumentation__.Notify(241347)

	if override := knobs.SQLWatcherCheckpointNoopsEveryDurationOverride; override.Nanoseconds() != 0 {
		__antithesis_instrumentation__.Notify(241351)
		checkpointNoopsEvery = override
	} else {
		__antithesis_instrumentation__.Notify(241352)
	}
	__antithesis_instrumentation__.Notify(241348)

	return &SQLWatcher{
		codec:                codec,
		settings:             settings,
		rangeFeedFactory:     rangeFeedFactory,
		stopper:              stopper,
		bufferMemLimit:       bufferMemLimit,
		checkpointNoopsEvery: checkpointNoopsEvery,
		knobs:                knobs,
	}
}

const sqlWatcherBufferEntrySize = int64(unsafe.Sizeof(event{}) + unsafe.Sizeof(rangefeedbuffer.Event(nil)))

func (s *SQLWatcher) WatchForSQLUpdates(
	ctx context.Context, startTS hlc.Timestamp, handler spanconfig.SQLWatcherHandler,
) error {
	__antithesis_instrumentation__.Notify(241353)
	return s.watch(ctx, startTS, handler)
}

func (s *SQLWatcher) watch(
	ctx context.Context, startTS hlc.Timestamp, handler spanconfig.SQLWatcherHandler,
) error {
	__antithesis_instrumentation__.Notify(241354)

	errCh := make(chan error)
	frontierAdvanced := make(chan struct{})
	buf := newBuffer(int(s.bufferMemLimit/sqlWatcherBufferEntrySize), startTS)
	onFrontierAdvance := func(ctx context.Context, rangefeed rangefeedKind, timestamp hlc.Timestamp) {
		__antithesis_instrumentation__.Notify(241360)
		buf.advance(rangefeed, timestamp)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(241361)

		case frontierAdvanced <- struct{}{}:
			__antithesis_instrumentation__.Notify(241362)
		}
	}
	__antithesis_instrumentation__.Notify(241355)
	onEvent := func(ctx context.Context, event event) {
		__antithesis_instrumentation__.Notify(241363)
		err := func() error {
			__antithesis_instrumentation__.Notify(241365)
			if fn := s.knobs.SQLWatcherOnEventInterceptor; fn != nil {
				__antithesis_instrumentation__.Notify(241367)
				if err := fn(); err != nil {
					__antithesis_instrumentation__.Notify(241368)
					return err
				} else {
					__antithesis_instrumentation__.Notify(241369)
				}
			} else {
				__antithesis_instrumentation__.Notify(241370)
			}
			__antithesis_instrumentation__.Notify(241366)
			return buf.add(event)
		}()
		__antithesis_instrumentation__.Notify(241364)
		if err != nil {
			__antithesis_instrumentation__.Notify(241371)
			log.Warningf(ctx, "error adding event %v: %v", event, err)
			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(241372)

			case errCh <- err:
				__antithesis_instrumentation__.Notify(241373)
			}
		} else {
			__antithesis_instrumentation__.Notify(241374)
		}
	}
	__antithesis_instrumentation__.Notify(241356)

	descriptorsRF, err := s.watchForDescriptorUpdates(ctx, startTS, onEvent, onFrontierAdvance)
	if err != nil {
		__antithesis_instrumentation__.Notify(241375)
		return errors.Wrapf(err, "error establishing rangefeed over system.descriptors")
	} else {
		__antithesis_instrumentation__.Notify(241376)
	}
	__antithesis_instrumentation__.Notify(241357)
	defer descriptorsRF.Close()
	zonesRF, err := s.watchForZoneConfigUpdates(ctx, startTS, onEvent, onFrontierAdvance)
	if err != nil {
		__antithesis_instrumentation__.Notify(241377)
		return errors.Wrapf(err, "error establishing rangefeed over system.zones")
	} else {
		__antithesis_instrumentation__.Notify(241378)
	}
	__antithesis_instrumentation__.Notify(241358)
	defer zonesRF.Close()
	ptsRF, err := s.watchForProtectedTimestampUpdates(ctx, startTS, onEvent, onFrontierAdvance)
	if err != nil {
		__antithesis_instrumentation__.Notify(241379)
		return errors.Wrapf(err, "error establishing rangefeed over system.protected_ts_records")
	} else {
		__antithesis_instrumentation__.Notify(241380)
	}
	__antithesis_instrumentation__.Notify(241359)
	defer ptsRF.Close()

	checkpointNoops := util.Every(s.checkpointNoopsEvery)
	for {
		__antithesis_instrumentation__.Notify(241381)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(241382)
			return ctx.Err()
		case <-s.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(241383)
			return nil
		case err := <-errCh:
			__antithesis_instrumentation__.Notify(241384)
			return err
		case <-frontierAdvanced:
			__antithesis_instrumentation__.Notify(241385)
			sqlUpdates, combinedFrontierTS, err := buf.flush(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(241388)
				return err
			} else {
				__antithesis_instrumentation__.Notify(241389)
			}
			__antithesis_instrumentation__.Notify(241386)
			if len(sqlUpdates) == 0 && func() bool {
				__antithesis_instrumentation__.Notify(241390)
				return !checkpointNoops.ShouldProcess(timeutil.Now()) == true
			}() == true {
				__antithesis_instrumentation__.Notify(241391)
				continue
			} else {
				__antithesis_instrumentation__.Notify(241392)
			}
			__antithesis_instrumentation__.Notify(241387)
			if err := handler(ctx, sqlUpdates, combinedFrontierTS); err != nil {
				__antithesis_instrumentation__.Notify(241393)
				return err
			} else {
				__antithesis_instrumentation__.Notify(241394)
			}
		}
	}
}

func (s *SQLWatcher) watchForDescriptorUpdates(
	ctx context.Context,
	startTS hlc.Timestamp,
	onEvent func(context.Context, event),
	onFrontierAdvance func(context.Context, rangefeedKind, hlc.Timestamp),
) (*rangefeed.RangeFeed, error) {
	__antithesis_instrumentation__.Notify(241395)
	descriptorTableStart := s.codec.TablePrefix(keys.DescriptorTableID)
	descriptorTableSpan := roachpb.Span{
		Key:    descriptorTableStart,
		EndKey: descriptorTableStart.PrefixEnd(),
	}
	handleEvent := func(ctx context.Context, ev *roachpb.RangeFeedValue) {
		__antithesis_instrumentation__.Notify(241399)
		if !ev.Value.IsPresent() && func() bool {
			__antithesis_instrumentation__.Notify(241405)
			return !ev.PrevValue.IsPresent() == true
		}() == true {
			__antithesis_instrumentation__.Notify(241406)

			return
		} else {
			__antithesis_instrumentation__.Notify(241407)
		}
		__antithesis_instrumentation__.Notify(241400)
		value := ev.Value
		if !ev.Value.IsPresent() {
			__antithesis_instrumentation__.Notify(241408)

			value = ev.PrevValue
		} else {
			__antithesis_instrumentation__.Notify(241409)
		}
		__antithesis_instrumentation__.Notify(241401)

		var descriptor descpb.Descriptor
		if err := value.GetProto(&descriptor); err != nil {
			__antithesis_instrumentation__.Notify(241410)
			logcrash.ReportOrPanic(
				ctx,
				&s.settings.SV,
				"%s: failed to unmarshal descriptor %v",
				ev.Key,
				value,
			)
			return
		} else {
			__antithesis_instrumentation__.Notify(241411)
		}
		__antithesis_instrumentation__.Notify(241402)
		if descriptor.Union == nil {
			__antithesis_instrumentation__.Notify(241412)
			return
		} else {
			__antithesis_instrumentation__.Notify(241413)
		}
		__antithesis_instrumentation__.Notify(241403)

		table, database, typ, schema := descpb.FromDescriptorWithMVCCTimestamp(&descriptor, ev.Value.Timestamp)

		var id descpb.ID
		var descType catalog.DescriptorType
		switch {
		case table != nil:
			__antithesis_instrumentation__.Notify(241414)
			id = table.GetID()
			descType = catalog.Table
		case database != nil:
			__antithesis_instrumentation__.Notify(241415)
			id = database.GetID()
			descType = catalog.Database
		case typ != nil:
			__antithesis_instrumentation__.Notify(241416)
			id = typ.GetID()
			descType = catalog.Type
		case schema != nil:
			__antithesis_instrumentation__.Notify(241417)
			id = schema.GetID()
			descType = catalog.Schema
		default:
			__antithesis_instrumentation__.Notify(241418)
			logcrash.ReportOrPanic(ctx, &s.settings.SV, "unknown descriptor unmarshalled %v", descriptor)
		}
		__antithesis_instrumentation__.Notify(241404)

		rangefeedEvent := event{
			timestamp: ev.Value.Timestamp,
			update:    spanconfig.MakeDescriptorSQLUpdate(id, descType),
		}
		onEvent(ctx, rangefeedEvent)
	}
	__antithesis_instrumentation__.Notify(241396)
	rf, err := s.rangeFeedFactory.RangeFeed(
		ctx,
		"sql-watcher-descriptor-rangefeed",
		[]roachpb.Span{descriptorTableSpan},
		startTS,
		handleEvent,
		rangefeed.WithDiff(true),
		rangefeed.WithOnFrontierAdvance(func(ctx context.Context, resolvedTS hlc.Timestamp) {
			__antithesis_instrumentation__.Notify(241419)
			onFrontierAdvance(ctx, descriptorsRangefeed, resolvedTS)
		}),
	)
	__antithesis_instrumentation__.Notify(241397)
	if err != nil {
		__antithesis_instrumentation__.Notify(241420)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241421)
	}
	__antithesis_instrumentation__.Notify(241398)

	log.Infof(ctx, "established range feed over system.descriptors starting at time %s", startTS)
	return rf, nil
}

func (s *SQLWatcher) watchForZoneConfigUpdates(
	ctx context.Context,
	startTS hlc.Timestamp,
	onEvent func(context.Context, event),
	onFrontierAdvance func(context.Context, rangefeedKind, hlc.Timestamp),
) (*rangefeed.RangeFeed, error) {
	__antithesis_instrumentation__.Notify(241422)
	zoneTableStart := s.codec.TablePrefix(keys.ZonesTableID)
	zoneTableSpan := roachpb.Span{
		Key:    zoneTableStart,
		EndKey: zoneTableStart.PrefixEnd(),
	}

	decoder := newZonesDecoder(s.codec)
	handleEvent := func(ctx context.Context, ev *roachpb.RangeFeedValue) {
		__antithesis_instrumentation__.Notify(241426)
		descID, err := decoder.DecodePrimaryKey(ev.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(241428)
			logcrash.ReportOrPanic(
				ctx,
				&s.settings.SV,
				"sql watcher zones range feed error: %v",
				err,
			)
			return
		} else {
			__antithesis_instrumentation__.Notify(241429)
		}
		__antithesis_instrumentation__.Notify(241427)

		rangefeedEvent := event{
			timestamp: ev.Value.Timestamp,
			update:    spanconfig.MakeDescriptorSQLUpdate(descID, catalog.Any),
		}
		onEvent(ctx, rangefeedEvent)
	}
	__antithesis_instrumentation__.Notify(241423)
	rf, err := s.rangeFeedFactory.RangeFeed(
		ctx,
		"sql-watcher-zones-rangefeed",
		[]roachpb.Span{zoneTableSpan},
		startTS,
		handleEvent,
		rangefeed.WithOnFrontierAdvance(func(ctx context.Context, resolvedTS hlc.Timestamp) {
			__antithesis_instrumentation__.Notify(241430)
			onFrontierAdvance(ctx, zonesRangefeed, resolvedTS)
		}),
	)
	__antithesis_instrumentation__.Notify(241424)
	if err != nil {
		__antithesis_instrumentation__.Notify(241431)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241432)
	}
	__antithesis_instrumentation__.Notify(241425)

	log.Infof(ctx, "established range feed over system.zones starting at time %s", startTS)
	return rf, nil
}

func (s *SQLWatcher) watchForProtectedTimestampUpdates(
	ctx context.Context,
	startTS hlc.Timestamp,
	onEvent func(context.Context, event),
	onFrontierAdvance func(context.Context, rangefeedKind, hlc.Timestamp),
) (*rangefeed.RangeFeed, error) {
	__antithesis_instrumentation__.Notify(241433)
	ptsRecordsTableStart := s.codec.TablePrefix(keys.ProtectedTimestampsRecordsTableID)
	ptsRecordsTableSpan := roachpb.Span{
		Key:    ptsRecordsTableStart,
		EndKey: ptsRecordsTableStart.PrefixEnd(),
	}

	decoder := newProtectedTimestampDecoder()
	handleEvent := func(ctx context.Context, ev *roachpb.RangeFeedValue) {
		__antithesis_instrumentation__.Notify(241437)
		if !ev.Value.IsPresent() && func() bool {
			__antithesis_instrumentation__.Notify(241442)
			return !ev.PrevValue.IsPresent() == true
		}() == true {
			__antithesis_instrumentation__.Notify(241443)

			return
		} else {
			__antithesis_instrumentation__.Notify(241444)
		}
		__antithesis_instrumentation__.Notify(241438)
		value := ev.Value
		if !ev.Value.IsPresent() {
			__antithesis_instrumentation__.Notify(241445)

			value = ev.PrevValue
		} else {
			__antithesis_instrumentation__.Notify(241446)
		}
		__antithesis_instrumentation__.Notify(241439)
		target, err := decoder.decode(roachpb.KeyValue{Value: value})
		if err != nil {
			__antithesis_instrumentation__.Notify(241447)
			logcrash.ReportOrPanic(
				ctx,
				&s.settings.SV,
				"sql watcher protected timestamp range feed error: %v",
				err,
			)
			return
		} else {
			__antithesis_instrumentation__.Notify(241448)
		}
		__antithesis_instrumentation__.Notify(241440)
		if target.Union == nil {
			__antithesis_instrumentation__.Notify(241449)
			return
		} else {
			__antithesis_instrumentation__.Notify(241450)
		}
		__antithesis_instrumentation__.Notify(241441)

		ts := ev.Value.Timestamp
		switch t := target.Union.(type) {
		case *ptpb.Target_Cluster:
			__antithesis_instrumentation__.Notify(241451)
			rangefeedEvent := event{
				timestamp: ts,
				update:    spanconfig.MakeClusterProtectedTimestampSQLUpdate(),
			}
			onEvent(ctx, rangefeedEvent)
		case *ptpb.Target_Tenants:
			__antithesis_instrumentation__.Notify(241452)

			for _, tenID := range t.Tenants.IDs {
				__antithesis_instrumentation__.Notify(241455)
				rangefeedEvent := event{
					timestamp: ts,
					update:    spanconfig.MakeTenantProtectedTimestampSQLUpdate(tenID),
				}
				onEvent(ctx, rangefeedEvent)
			}
		case *ptpb.Target_SchemaObjects:
			__antithesis_instrumentation__.Notify(241453)

			for _, id := range t.SchemaObjects.IDs {
				__antithesis_instrumentation__.Notify(241456)
				rangefeedEvent := event{
					timestamp: ts,
					update:    spanconfig.MakeDescriptorSQLUpdate(id, catalog.Any),
				}
				onEvent(ctx, rangefeedEvent)
			}
		default:
			__antithesis_instrumentation__.Notify(241454)
			logcrash.ReportOrPanic(ctx, &s.settings.SV,
				"unknown protected timestamp target %v", target)
		}
	}
	__antithesis_instrumentation__.Notify(241434)
	rf, err := s.rangeFeedFactory.RangeFeed(
		ctx,
		"sql-watcher-protected-ts-records-rangefeed",
		[]roachpb.Span{ptsRecordsTableSpan},
		startTS,
		handleEvent,
		rangefeed.WithOnFrontierAdvance(func(ctx context.Context, resolvedTS hlc.Timestamp) {
			__antithesis_instrumentation__.Notify(241457)
			onFrontierAdvance(ctx, protectedTimestampRangefeed, resolvedTS)
		}),
		rangefeed.WithDiff(true))
	__antithesis_instrumentation__.Notify(241435)
	if err != nil {
		__antithesis_instrumentation__.Notify(241458)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(241459)
	}
	__antithesis_instrumentation__.Notify(241436)

	log.Infof(ctx, "established range feed over system.protected_ts_records starting at time %s", startTS)
	return rf, nil
}
