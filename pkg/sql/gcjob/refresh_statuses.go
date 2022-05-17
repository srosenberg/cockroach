package gcjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var maxDeadline = timeutil.Unix(0, math.MaxInt64)

func refreshTables(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	tableIDs []descpb.ID,
	tableDropTimes map[descpb.ID]int64,
	indexDropTimes map[descpb.IndexID]int64,
	jobID jobspb.JobID,
	progress *jobspb.SchemaChangeGCProgress,
) (expired bool, earliestDeadline time.Time) {
	__antithesis_instrumentation__.Notify(492503)
	earliestDeadline = maxDeadline
	var haveAnyMissing bool
	for _, tableID := range tableIDs {
		__antithesis_instrumentation__.Notify(492506)
		tableHasExpiredElem, tableIsMissing, deadline := updateStatusForGCElements(
			ctx,
			execCfg,
			jobID,
			tableID,
			tableDropTimes, indexDropTimes,
			progress,
		)
		expired = expired || func() bool {
			__antithesis_instrumentation__.Notify(492507)
			return tableHasExpiredElem == true
		}() == true
		haveAnyMissing = haveAnyMissing || func() bool {
			__antithesis_instrumentation__.Notify(492508)
			return tableIsMissing == true
		}() == true
		if deadline.Before(earliestDeadline) {
			__antithesis_instrumentation__.Notify(492509)
			earliestDeadline = deadline
		} else {
			__antithesis_instrumentation__.Notify(492510)
		}
	}
	__antithesis_instrumentation__.Notify(492504)

	if expired || func() bool {
		__antithesis_instrumentation__.Notify(492511)
		return haveAnyMissing == true
	}() == true {
		__antithesis_instrumentation__.Notify(492512)
		persistProgress(ctx, execCfg, jobID, progress, sql.RunningStatusWaitingGC)
	} else {
		__antithesis_instrumentation__.Notify(492513)
	}
	__antithesis_instrumentation__.Notify(492505)

	return expired, earliestDeadline
}

func updateStatusForGCElements(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	jobID jobspb.JobID,
	tableID descpb.ID,
	tableDropTimes map[descpb.ID]int64,
	indexDropTimes map[descpb.IndexID]int64,
	progress *jobspb.SchemaChangeGCProgress,
) (expired, missing bool, timeToNextTrigger time.Time) {
	__antithesis_instrumentation__.Notify(492514)
	defTTL := execCfg.DefaultZoneConfig.GC.TTLSeconds
	cfg := execCfg.SystemConfig.GetSystemConfig()
	protectedtsCache := execCfg.ProtectedTimestampProvider

	earliestDeadline := timeutil.Unix(0, int64(math.MaxInt64))

	if err := sql.DescsTxn(ctx, execCfg, func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error {
		__antithesis_instrumentation__.Notify(492516)
		table, err := col.Direct().MustGetTableDescByID(ctx, txn, tableID)
		if err != nil {
			__antithesis_instrumentation__.Notify(492522)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492523)
		}
		__antithesis_instrumentation__.Notify(492517)
		v := execCfg.Settings.Version.ActiveVersionOrEmpty(ctx)
		zoneCfg, err := cfg.GetZoneConfigForObject(execCfg.Codec, v, config.ObjectID(tableID))
		if err != nil {
			__antithesis_instrumentation__.Notify(492524)
			log.Errorf(ctx, "zone config for desc: %d, err = %+v", tableID, err)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(492525)
		}
		__antithesis_instrumentation__.Notify(492518)
		tableTTL := getTableTTL(defTTL, zoneCfg)

		if table.Dropped() {
			__antithesis_instrumentation__.Notify(492526)
			deadline := updateTableStatus(ctx, execCfg, jobID, int64(tableTTL), table, tableDropTimes, progress)
			if timeutil.Until(deadline) < 0 {
				__antithesis_instrumentation__.Notify(492527)
				expired = true
			} else {
				__antithesis_instrumentation__.Notify(492528)
				if deadline.Before(earliestDeadline) {
					__antithesis_instrumentation__.Notify(492529)
					earliestDeadline = deadline
				} else {
					__antithesis_instrumentation__.Notify(492530)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(492531)
		}
		__antithesis_instrumentation__.Notify(492519)

		indexesExpired, deadline := updateIndexesStatus(
			ctx, execCfg, jobID, tableTTL, table, protectedtsCache, zoneCfg, indexDropTimes, progress,
		)
		if indexesExpired {
			__antithesis_instrumentation__.Notify(492532)
			expired = true
		} else {
			__antithesis_instrumentation__.Notify(492533)
		}
		__antithesis_instrumentation__.Notify(492520)
		if deadline.Before(earliestDeadline) {
			__antithesis_instrumentation__.Notify(492534)
			earliestDeadline = deadline
		} else {
			__antithesis_instrumentation__.Notify(492535)
		}
		__antithesis_instrumentation__.Notify(492521)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(492536)
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(492538)
			log.Warningf(ctx, "table %d not found, marking as GC'd", tableID)
			markTableGCed(ctx, tableID, progress)
			return false, true, maxDeadline
		} else {
			__antithesis_instrumentation__.Notify(492539)
		}
		__antithesis_instrumentation__.Notify(492537)
		log.Warningf(ctx, "error while calculating GC time for table %d, err: %+v", tableID, err)
		return false, false, maxDeadline
	} else {
		__antithesis_instrumentation__.Notify(492540)
	}
	__antithesis_instrumentation__.Notify(492515)

	return expired, false, earliestDeadline
}

func updateTableStatus(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	jobID jobspb.JobID,
	ttlSeconds int64,
	table catalog.TableDescriptor,
	tableDropTimes map[descpb.ID]int64,
	progress *jobspb.SchemaChangeGCProgress,
) time.Time {
	__antithesis_instrumentation__.Notify(492541)
	deadline := timeutil.Unix(0, int64(math.MaxInt64))
	sp := table.TableSpan(execCfg.Codec)

	for i, t := range progress.Tables {
		__antithesis_instrumentation__.Notify(492543)
		droppedTable := &progress.Tables[i]
		if droppedTable.ID != table.GetID() || func() bool {
			__antithesis_instrumentation__.Notify(492548)
			return droppedTable.Status == jobspb.SchemaChangeGCProgress_DELETED == true
		}() == true {
			__antithesis_instrumentation__.Notify(492549)
			continue
		} else {
			__antithesis_instrumentation__.Notify(492550)
		}
		__antithesis_instrumentation__.Notify(492544)

		deadlineNanos := tableDropTimes[t.ID] + ttlSeconds*time.Second.Nanoseconds()
		deadline = timeutil.Unix(0, deadlineNanos)
		isProtected, err := isProtected(
			ctx,
			jobID,
			tableDropTimes[t.ID],
			execCfg,
			execCfg.SpanConfigKVAccessor,
			execCfg.ProtectedTimestampProvider,
			sp,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(492551)
			log.Errorf(ctx, "error checking protection status %v", err)

			return maxDeadline
		} else {
			__antithesis_instrumentation__.Notify(492552)
		}
		__antithesis_instrumentation__.Notify(492545)
		if isProtected {
			__antithesis_instrumentation__.Notify(492553)
			log.Infof(ctx, "a timestamp protection delayed GC of table %d", t.ID)
			return maxDeadline
		} else {
			__antithesis_instrumentation__.Notify(492554)
		}
		__antithesis_instrumentation__.Notify(492546)

		lifetime := timeutil.Until(deadline)
		if lifetime < 0 {
			__antithesis_instrumentation__.Notify(492555)
			if log.V(2) {
				__antithesis_instrumentation__.Notify(492557)
				log.Infof(ctx, "detected expired table %d", t.ID)
			} else {
				__antithesis_instrumentation__.Notify(492558)
			}
			__antithesis_instrumentation__.Notify(492556)
			droppedTable.Status = jobspb.SchemaChangeGCProgress_DELETING
		} else {
			__antithesis_instrumentation__.Notify(492559)
			if log.V(2) {
				__antithesis_instrumentation__.Notify(492560)
				log.Infof(ctx, "table %d still has %+v until GC", t.ID, lifetime)
			} else {
				__antithesis_instrumentation__.Notify(492561)
			}
		}
		__antithesis_instrumentation__.Notify(492547)
		break
	}
	__antithesis_instrumentation__.Notify(492542)

	return deadline
}

func updateIndexesStatus(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	jobID jobspb.JobID,
	tableTTL int32,
	table catalog.TableDescriptor,
	protectedtsCache protectedts.Cache,
	zoneCfg *zonepb.ZoneConfig,
	indexDropTimes map[descpb.IndexID]int64,
	progress *jobspb.SchemaChangeGCProgress,
) (expired bool, soonestDeadline time.Time) {
	__antithesis_instrumentation__.Notify(492562)

	soonestDeadline = timeutil.Unix(0, int64(math.MaxInt64))
	for i := 0; i < len(progress.Indexes); i++ {
		__antithesis_instrumentation__.Notify(492564)
		idxProgress := &progress.Indexes[i]
		if idxProgress.Status == jobspb.SchemaChangeGCProgress_DELETED {
			__antithesis_instrumentation__.Notify(492569)
			continue
		} else {
			__antithesis_instrumentation__.Notify(492570)
		}
		__antithesis_instrumentation__.Notify(492565)

		sp := table.IndexSpan(execCfg.Codec, idxProgress.IndexID)

		ttlSeconds := getIndexTTL(tableTTL, zoneCfg, idxProgress.IndexID)

		deadlineNanos := indexDropTimes[idxProgress.IndexID] + int64(ttlSeconds)*time.Second.Nanoseconds()
		deadline := timeutil.Unix(0, deadlineNanos)
		isProtected, err := isProtected(
			ctx,
			jobID,
			indexDropTimes[idxProgress.IndexID],
			execCfg,
			execCfg.SpanConfigKVAccessor,
			protectedtsCache,
			sp,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(492571)
			log.Errorf(ctx, "error checking protection status %v", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(492572)
		}
		__antithesis_instrumentation__.Notify(492566)
		if isProtected {
			__antithesis_instrumentation__.Notify(492573)
			log.Infof(ctx, "a timestamp protection delayed GC of index %d from table %d", idxProgress.IndexID, table.GetID())
			continue
		} else {
			__antithesis_instrumentation__.Notify(492574)
		}
		__antithesis_instrumentation__.Notify(492567)
		lifetime := time.Until(deadline)
		if lifetime > 0 {
			__antithesis_instrumentation__.Notify(492575)
			if log.V(2) {
				__antithesis_instrumentation__.Notify(492576)
				log.Infof(ctx, "index %d from table %d still has %+v until GC", idxProgress.IndexID, table.GetID(), lifetime)
			} else {
				__antithesis_instrumentation__.Notify(492577)
			}
		} else {
			__antithesis_instrumentation__.Notify(492578)
		}
		__antithesis_instrumentation__.Notify(492568)
		if lifetime < 0 {
			__antithesis_instrumentation__.Notify(492579)
			expired = true
			if log.V(2) {
				__antithesis_instrumentation__.Notify(492581)
				log.Infof(ctx, "detected expired index %d from table %d", idxProgress.IndexID, table.GetID())
			} else {
				__antithesis_instrumentation__.Notify(492582)
			}
			__antithesis_instrumentation__.Notify(492580)
			idxProgress.Status = jobspb.SchemaChangeGCProgress_DELETING
		} else {
			__antithesis_instrumentation__.Notify(492583)
			if deadline.Before(soonestDeadline) {
				__antithesis_instrumentation__.Notify(492584)
				soonestDeadline = deadline
			} else {
				__antithesis_instrumentation__.Notify(492585)
			}
		}
	}
	__antithesis_instrumentation__.Notify(492563)
	return expired, soonestDeadline
}

func getIndexTTL(tableTTL int32, placeholder *zonepb.ZoneConfig, indexID descpb.IndexID) int32 {
	__antithesis_instrumentation__.Notify(492586)
	ttlSeconds := tableTTL
	if placeholder != nil {
		__antithesis_instrumentation__.Notify(492588)
		if subzone := placeholder.GetSubzone(
			uint32(indexID), ""); subzone != nil && func() bool {
			__antithesis_instrumentation__.Notify(492589)
			return subzone.Config.GC != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(492590)
			ttlSeconds = subzone.Config.GC.TTLSeconds
		} else {
			__antithesis_instrumentation__.Notify(492591)
		}
	} else {
		__antithesis_instrumentation__.Notify(492592)
	}
	__antithesis_instrumentation__.Notify(492587)
	return ttlSeconds
}

func getTableTTL(defTTL int32, zoneCfg *zonepb.ZoneConfig) int32 {
	__antithesis_instrumentation__.Notify(492593)
	ttlSeconds := defTTL
	if zoneCfg != nil {
		__antithesis_instrumentation__.Notify(492595)
		ttlSeconds = zoneCfg.GC.TTLSeconds
	} else {
		__antithesis_instrumentation__.Notify(492596)
	}
	__antithesis_instrumentation__.Notify(492594)
	return ttlSeconds
}

func isProtected(
	ctx context.Context,
	jobID jobspb.JobID,
	droppedAtTime int64,
	execCfg *sql.ExecutorConfig,
	kvAccessor spanconfig.KVAccessor,
	ptsCache protectedts.Cache,
	sp roachpb.Span,
) (bool, error) {
	__antithesis_instrumentation__.Notify(492597)

	isProtected, err := func() (bool, error) {
		__antithesis_instrumentation__.Notify(492601)

		if execCfg.Codec.ForSystemTenant() && func() bool {
			__antithesis_instrumentation__.Notify(492609)
			return deprecatedIsProtected(ctx, ptsCache, droppedAtTime, sp) == true
		}() == true {
			__antithesis_instrumentation__.Notify(492610)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(492611)
		}
		__antithesis_instrumentation__.Notify(492602)

		spanConfigRecords, err := kvAccessor.GetSpanConfigRecords(ctx, spanconfig.Targets{
			spanconfig.MakeTargetFromSpan(sp),
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(492612)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(492613)
		}
		__antithesis_instrumentation__.Notify(492603)

		_, tenID, err := keys.DecodeTenantPrefix(execCfg.Codec.TenantPrefix())
		if err != nil {
			__antithesis_instrumentation__.Notify(492614)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(492615)
		}
		__antithesis_instrumentation__.Notify(492604)
		systemSpanConfigs, err := kvAccessor.GetAllSystemSpanConfigsThatApply(ctx, tenID)
		if err != nil {
			__antithesis_instrumentation__.Notify(492616)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(492617)
		}
		__antithesis_instrumentation__.Notify(492605)

		var protectedTimestamps []hlc.Timestamp
		collectProtectedTimestamps := func(configs ...roachpb.SpanConfig) {
			__antithesis_instrumentation__.Notify(492618)
			for _, config := range configs {
				__antithesis_instrumentation__.Notify(492619)
				for _, protectionPolicy := range config.GCPolicy.ProtectionPolicies {
					__antithesis_instrumentation__.Notify(492620)

					if config.ExcludeDataFromBackup && func() bool {
						__antithesis_instrumentation__.Notify(492622)
						return protectionPolicy.IgnoreIfExcludedFromBackup == true
					}() == true {
						__antithesis_instrumentation__.Notify(492623)
						continue
					} else {
						__antithesis_instrumentation__.Notify(492624)
					}
					__antithesis_instrumentation__.Notify(492621)
					protectedTimestamps = append(protectedTimestamps, protectionPolicy.ProtectedTimestamp)
				}
			}
		}
		__antithesis_instrumentation__.Notify(492606)
		for _, record := range spanConfigRecords {
			__antithesis_instrumentation__.Notify(492625)
			collectProtectedTimestamps(record.GetConfig())
		}
		__antithesis_instrumentation__.Notify(492607)
		collectProtectedTimestamps(systemSpanConfigs...)

		for _, protectedTimestamp := range protectedTimestamps {
			__antithesis_instrumentation__.Notify(492626)
			if protectedTimestamp.WallTime < droppedAtTime {
				__antithesis_instrumentation__.Notify(492627)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(492628)
			}
		}
		__antithesis_instrumentation__.Notify(492608)

		return false, nil
	}()
	__antithesis_instrumentation__.Notify(492598)
	if err != nil {
		__antithesis_instrumentation__.Notify(492629)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(492630)
	}
	__antithesis_instrumentation__.Notify(492599)

	if fn := execCfg.GCJobTestingKnobs.RunAfterIsProtectedCheck; fn != nil {
		__antithesis_instrumentation__.Notify(492631)
		fn(jobID, isProtected)
	} else {
		__antithesis_instrumentation__.Notify(492632)
	}
	__antithesis_instrumentation__.Notify(492600)

	return isProtected, nil
}

func deprecatedIsProtected(
	ctx context.Context, protectedtsCache protectedts.Cache, atTime int64, sp roachpb.Span,
) bool {
	__antithesis_instrumentation__.Notify(492633)
	protected := false
	protectedtsCache.Iterate(ctx,
		sp.Key, sp.EndKey,
		func(r *ptpb.Record) (wantMore bool) {
			__antithesis_instrumentation__.Notify(492635)

			if r.Timestamp.WallTime < atTime {
				__antithesis_instrumentation__.Notify(492637)
				protected = true
				return false
			} else {
				__antithesis_instrumentation__.Notify(492638)
			}
			__antithesis_instrumentation__.Notify(492636)
			return true
		})
	__antithesis_instrumentation__.Notify(492634)
	return protected
}

func isTenantProtected(
	ctx context.Context, atTime hlc.Timestamp, tenantID roachpb.TenantID, execCfg *sql.ExecutorConfig,
) (bool, error) {
	__antithesis_instrumentation__.Notify(492639)
	if !execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(492642)
		return false, errors.AssertionFailedf("isTenantProtected incorrectly invoked by secondary tenant")
	} else {
		__antithesis_instrumentation__.Notify(492643)
	}
	__antithesis_instrumentation__.Notify(492640)

	isProtected := false
	ptsProvider := execCfg.ProtectedTimestampProvider
	if err := execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(492644)
		ptsState, err := ptsProvider.GetState(ctx, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(492648)
			return errors.Wrap(err, "failed to get protectedts State")
		} else {
			__antithesis_instrumentation__.Notify(492649)
		}
		__antithesis_instrumentation__.Notify(492645)
		ptsStateReader := spanconfig.NewProtectedTimestampStateReader(ctx, ptsState)

		clusterProtections := ptsStateReader.GetProtectionPoliciesForCluster()
		for _, p := range clusterProtections {
			__antithesis_instrumentation__.Notify(492650)
			if p.ProtectedTimestamp.Less(atTime) {
				__antithesis_instrumentation__.Notify(492651)
				isProtected = true
				return nil
			} else {
				__antithesis_instrumentation__.Notify(492652)
			}
		}
		__antithesis_instrumentation__.Notify(492646)

		protectionsOnTenant := ptsStateReader.GetProtectionsForTenant(tenantID)
		for _, p := range protectionsOnTenant {
			__antithesis_instrumentation__.Notify(492653)
			if p.ProtectedTimestamp.Less(atTime) {
				__antithesis_instrumentation__.Notify(492654)
				isProtected = true
				return nil
			} else {
				__antithesis_instrumentation__.Notify(492655)
			}
		}
		__antithesis_instrumentation__.Notify(492647)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(492656)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(492657)
	}
	__antithesis_instrumentation__.Notify(492641)
	return isProtected, nil
}

func refreshTenant(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	dropTime int64,
	details *jobspb.SchemaChangeGCDetails,
	progress *jobspb.SchemaChangeGCProgress,
) (expired bool, _ time.Time, _ error) {
	__antithesis_instrumentation__.Notify(492658)
	if progress.Tenant.Status != jobspb.SchemaChangeGCProgress_WAITING_FOR_GC {
		__antithesis_instrumentation__.Notify(492662)
		return true, time.Time{}, nil
	} else {
		__antithesis_instrumentation__.Notify(492663)
	}
	__antithesis_instrumentation__.Notify(492659)

	tenID := details.Tenant.ID
	cfg := execCfg.SystemConfig.GetSystemConfig()
	tenantTTLSeconds := execCfg.DefaultZoneConfig.GC.TTLSeconds
	v := execCfg.Settings.Version.ActiveVersionOrEmpty(ctx)
	zoneCfg, err := cfg.GetZoneConfigForObject(keys.SystemSQLCodec, v, keys.TenantsRangesID)
	if err == nil {
		__antithesis_instrumentation__.Notify(492664)
		tenantTTLSeconds = zoneCfg.GC.TTLSeconds
	} else {
		__antithesis_instrumentation__.Notify(492665)
		log.Errorf(ctx, "zone config for tenants range: err = %+v", err)
	}
	__antithesis_instrumentation__.Notify(492660)

	deadlineNanos := dropTime + int64(tenantTTLSeconds)*time.Second.Nanoseconds()
	deadlineUnix := timeutil.Unix(0, deadlineNanos)
	if timeutil.Now().UnixNano() >= deadlineNanos {
		__antithesis_instrumentation__.Notify(492666)

		atTime := hlc.Timestamp{WallTime: dropTime}
		isProtected, err := isTenantProtected(ctx, atTime, roachpb.MakeTenantID(tenID), execCfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(492669)
			return false, time.Time{}, err
		} else {
			__antithesis_instrumentation__.Notify(492670)
		}
		__antithesis_instrumentation__.Notify(492667)

		if isProtected {
			__antithesis_instrumentation__.Notify(492671)
			log.Infof(ctx, "GC TTL for dropped tenant %d has expired, but protected timestamp "+
				"record(s) on the tenant keyspace are preventing GC", tenID)
			return false, deadlineUnix, nil
		} else {
			__antithesis_instrumentation__.Notify(492672)
		}
		__antithesis_instrumentation__.Notify(492668)

		progress.Tenant.Status = jobspb.SchemaChangeGCProgress_DELETING
		return true, deadlineUnix, nil
	} else {
		__antithesis_instrumentation__.Notify(492673)
	}
	__antithesis_instrumentation__.Notify(492661)
	return false, deadlineUnix, nil
}
