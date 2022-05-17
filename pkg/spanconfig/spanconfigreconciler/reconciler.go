package spanconfigreconciler

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigsqltranslator"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigstore"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type Reconciler struct {
	sqlWatcher           spanconfig.SQLWatcher
	sqlTranslatorFactory *spanconfigsqltranslator.Factory
	kvAccessor           spanconfig.KVAccessor

	execCfg *sql.ExecutorConfig
	codec   keys.SQLCodec
	tenID   roachpb.TenantID
	knobs   *spanconfig.TestingKnobs

	mu struct {
		syncutil.RWMutex
		lastCheckpoint hlc.Timestamp
	}
}

var _ spanconfig.Reconciler = &Reconciler{}

func New(
	sqlWatcher spanconfig.SQLWatcher,
	sqlTranslatorFactory *spanconfigsqltranslator.Factory,
	kvAccessor spanconfig.KVAccessor,
	execCfg *sql.ExecutorConfig,
	codec keys.SQLCodec,
	tenID roachpb.TenantID,
	knobs *spanconfig.TestingKnobs,
) *Reconciler {
	__antithesis_instrumentation__.Notify(240788)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(240790)
		knobs = &spanconfig.TestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(240791)
	}
	__antithesis_instrumentation__.Notify(240789)
	return &Reconciler{
		sqlWatcher:           sqlWatcher,
		sqlTranslatorFactory: sqlTranslatorFactory,
		kvAccessor:           kvAccessor,

		execCfg: execCfg,
		codec:   codec,
		tenID:   tenID,
		knobs:   knobs,
	}
}

func (r *Reconciler) Reconcile(
	ctx context.Context,
	startTS hlc.Timestamp,
	session sqlliveness.Session,
	onCheckpoint func() error,
) error {
	__antithesis_instrumentation__.Notify(240792)

	if fn := r.knobs.ReconcilerInitialInterceptor; fn != nil {
		__antithesis_instrumentation__.Notify(240796)
		fn(startTS)
	} else {
		__antithesis_instrumentation__.Notify(240797)
	}
	__antithesis_instrumentation__.Notify(240793)

	full := fullReconciler{
		sqlTranslatorFactory: r.sqlTranslatorFactory,
		kvAccessor:           r.kvAccessor,
		session:              session,
		execCfg:              r.execCfg,
		codec:                r.codec,
		tenID:                r.tenID,
		knobs:                r.knobs,
	}
	latestStore, reconciledUpUntil, err := full.reconcile(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(240798)
		return err
	} else {
		__antithesis_instrumentation__.Notify(240799)
	}
	__antithesis_instrumentation__.Notify(240794)

	r.mu.Lock()
	r.mu.lastCheckpoint = reconciledUpUntil
	r.mu.Unlock()

	if err := onCheckpoint(); err != nil {
		__antithesis_instrumentation__.Notify(240800)
		return err
	} else {
		__antithesis_instrumentation__.Notify(240801)
	}
	__antithesis_instrumentation__.Notify(240795)

	incrementalStartTS := reconciledUpUntil
	incremental := incrementalReconciler{
		sqlTranslatorFactory: r.sqlTranslatorFactory,
		sqlWatcher:           r.sqlWatcher,
		kvAccessor:           r.kvAccessor,
		storeWithKVContents:  latestStore,
		session:              session,
		execCfg:              r.execCfg,
		codec:                r.codec,
		knobs:                r.knobs,
	}
	return incremental.reconcile(ctx, incrementalStartTS, func(reconciledUpUntil hlc.Timestamp) error {
		__antithesis_instrumentation__.Notify(240802)
		r.mu.Lock()
		r.mu.lastCheckpoint = reconciledUpUntil
		r.mu.Unlock()

		return onCheckpoint()
	})
}

func (r *Reconciler) Checkpoint() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(240803)
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.mu.lastCheckpoint
}

type fullReconciler struct {
	sqlTranslatorFactory *spanconfigsqltranslator.Factory
	kvAccessor           spanconfig.KVAccessor
	session              sqlliveness.Session

	execCfg *sql.ExecutorConfig
	codec   keys.SQLCodec
	tenID   roachpb.TenantID
	knobs   *spanconfig.TestingKnobs
}

func (f *fullReconciler) reconcile(
	ctx context.Context,
) (storeWithLatestSpanConfigs *spanconfigstore.Store, reconciledUpUntil hlc.Timestamp, _ error) {
	__antithesis_instrumentation__.Notify(240804)
	storeWithExistingSpanConfigs, err := f.fetchExistingSpanConfigs(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(240812)
		return nil, hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(240813)
	}
	__antithesis_instrumentation__.Notify(240805)

	var records []spanconfig.Record

	if err := sql.DescsTxn(ctx, f.execCfg, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(240814)
		translator := f.sqlTranslatorFactory.NewSQLTranslator(txn, descsCol)
		records, reconciledUpUntil, err = spanconfig.FullTranslate(ctx, translator)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(240815)
		return nil, hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(240816)
	}
	__antithesis_instrumentation__.Notify(240806)

	updates := make([]spanconfig.Update, len(records))
	for i, record := range records {
		__antithesis_instrumentation__.Notify(240817)
		updates[i] = spanconfig.Update(record)
	}
	__antithesis_instrumentation__.Notify(240807)

	toDelete, toUpsert := storeWithExistingSpanConfigs.Apply(ctx, false, updates...)
	if len(toDelete) != 0 || func() bool {
		__antithesis_instrumentation__.Notify(240818)
		return len(toUpsert) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(240819)
		if err := updateSpanConfigRecords(
			ctx, f.kvAccessor, toDelete, toUpsert, f.session,
		); err != nil {
			__antithesis_instrumentation__.Notify(240820)
			return nil, hlc.Timestamp{}, err
		} else {
			__antithesis_instrumentation__.Notify(240821)
		}
	} else {
		__antithesis_instrumentation__.Notify(240822)
	}
	__antithesis_instrumentation__.Notify(240808)

	storeWithLatestSpanConfigs = storeWithExistingSpanConfigs.Copy(ctx)

	var storeWithExtraneousSpanConfigs *spanconfigstore.Store
	{
		__antithesis_instrumentation__.Notify(240823)
		for _, u := range updates {
			__antithesis_instrumentation__.Notify(240825)
			del, err := spanconfig.Deletion(u.GetTarget())
			if err != nil {
				__antithesis_instrumentation__.Notify(240827)
				return nil, hlc.Timestamp{}, err
			} else {
				__antithesis_instrumentation__.Notify(240828)
			}
			__antithesis_instrumentation__.Notify(240826)
			storeWithExistingSpanConfigs.Apply(ctx, false, del)
		}
		__antithesis_instrumentation__.Notify(240824)
		storeWithExtraneousSpanConfigs = storeWithExistingSpanConfigs
	}
	__antithesis_instrumentation__.Notify(240809)

	deletedSpans, err := f.deleteExtraneousSpanConfigs(ctx, storeWithExtraneousSpanConfigs)
	if err != nil {
		__antithesis_instrumentation__.Notify(240829)
		return nil, hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(240830)
	}
	__antithesis_instrumentation__.Notify(240810)

	for _, d := range deletedSpans {
		__antithesis_instrumentation__.Notify(240831)
		del, err := spanconfig.Deletion(d)
		if err != nil {
			__antithesis_instrumentation__.Notify(240833)
			return nil, hlc.Timestamp{}, err
		} else {
			__antithesis_instrumentation__.Notify(240834)
		}
		__antithesis_instrumentation__.Notify(240832)
		storeWithLatestSpanConfigs.Apply(ctx, false, del)
	}
	__antithesis_instrumentation__.Notify(240811)

	return storeWithLatestSpanConfigs, reconciledUpUntil, nil
}

func (f *fullReconciler) fetchExistingSpanConfigs(
	ctx context.Context,
) (*spanconfigstore.Store, error) {
	__antithesis_instrumentation__.Notify(240835)
	var targets []spanconfig.Target
	if f.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(240838)

		targets = append(targets, spanconfig.MakeTargetFromSpan(roachpb.Span{
			Key:    keys.EverythingSpan.Key,
			EndKey: keys.SystemSpanConfigSpan.Key,
		}))
		targets = append(targets, spanconfig.MakeTargetFromSpan(roachpb.Span{
			Key:    keys.TableDataMin,
			EndKey: keys.TableDataMax,
		}))

		targets = append(targets,
			spanconfig.MakeTargetFromSystemTarget(spanconfig.MakeEntireKeyspaceTarget()))
		targets = append(targets,
			spanconfig.MakeTargetFromSystemTarget(spanconfig.MakeAllTenantKeyspaceTargetsSet(f.tenID)))
		if f.knobs.ConfigureScratchRange {
			__antithesis_instrumentation__.Notify(240839)
			sp := targets[1].GetSpan()
			targets[1] = spanconfig.MakeTargetFromSpan(roachpb.Span{Key: sp.Key, EndKey: keys.ScratchRangeMax})
		} else {
			__antithesis_instrumentation__.Notify(240840)
		}
	} else {
		__antithesis_instrumentation__.Notify(240841)

		tenPrefix := keys.MakeTenantPrefix(f.tenID)
		targets = append(targets, spanconfig.MakeTargetFromSpan(roachpb.Span{
			Key:    tenPrefix,
			EndKey: tenPrefix.PrefixEnd(),
		}))

		targets = append(targets,
			spanconfig.MakeTargetFromSystemTarget(spanconfig.MakeAllTenantKeyspaceTargetsSet(f.tenID)))
	}
	__antithesis_instrumentation__.Notify(240836)
	store := spanconfigstore.New(roachpb.SpanConfig{})
	{
		__antithesis_instrumentation__.Notify(240842)

		records, err := f.kvAccessor.GetSpanConfigRecords(ctx, targets)
		if err != nil {
			__antithesis_instrumentation__.Notify(240844)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(240845)
		}
		__antithesis_instrumentation__.Notify(240843)

		for _, record := range records {
			__antithesis_instrumentation__.Notify(240846)
			store.Apply(ctx, false, spanconfig.Update(record))
		}
	}
	__antithesis_instrumentation__.Notify(240837)
	return store, nil
}

func (f *fullReconciler) deleteExtraneousSpanConfigs(
	ctx context.Context, storeWithExtraneousSpanConfigs *spanconfigstore.Store,
) ([]spanconfig.Target, error) {
	__antithesis_instrumentation__.Notify(240847)
	var extraneousTargets []spanconfig.Target
	if err := storeWithExtraneousSpanConfigs.Iterate(func(record spanconfig.Record) error {
		__antithesis_instrumentation__.Notify(240850)
		extraneousTargets = append(extraneousTargets, record.GetTarget())
		return nil
	},
	); err != nil {
		__antithesis_instrumentation__.Notify(240851)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(240852)
	}
	__antithesis_instrumentation__.Notify(240848)

	if len(extraneousTargets) != 0 {
		__antithesis_instrumentation__.Notify(240853)
		if err := updateSpanConfigRecords(
			ctx, f.kvAccessor, extraneousTargets, nil, f.session,
		); err != nil {
			__antithesis_instrumentation__.Notify(240854)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(240855)
		}
	} else {
		__antithesis_instrumentation__.Notify(240856)
	}
	__antithesis_instrumentation__.Notify(240849)
	return extraneousTargets, nil
}

func updateSpanConfigRecords(
	ctx context.Context,
	kvAccessor spanconfig.KVAccessor,
	toDelete []spanconfig.Target,
	toUpsert []spanconfig.Record,
	session sqlliveness.Session,
) error {
	__antithesis_instrumentation__.Notify(240857)
	retryOpts := retry.Options{
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     5 * time.Second,
		MaxRetries:     5,
	}

	for retrier := retry.StartWithCtx(ctx, retryOpts); retrier.Next(); {
		__antithesis_instrumentation__.Notify(240859)
		sessionStart, sessionExpiration := session.Start(), session.Expiration()
		if sessionExpiration.IsEmpty() {
			__antithesis_instrumentation__.Notify(240862)
			return errors.Errorf("sqlliveness session has expired")
		} else {
			__antithesis_instrumentation__.Notify(240863)
		}
		__antithesis_instrumentation__.Notify(240860)

		err := kvAccessor.UpdateSpanConfigRecords(
			ctx, toDelete, toUpsert, sessionStart, sessionExpiration,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(240864)
			if spanconfig.IsCommitTimestampOutOfBoundsError(err) {
				__antithesis_instrumentation__.Notify(240866)

				log.Infof(ctx, "lease expired while updating span config records, retrying..")
				continue
			} else {
				__antithesis_instrumentation__.Notify(240867)
			}
			__antithesis_instrumentation__.Notify(240865)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240868)
		}
		__antithesis_instrumentation__.Notify(240861)
		return nil
	}
	__antithesis_instrumentation__.Notify(240858)
	return nil
}

type incrementalReconciler struct {
	sqlTranslatorFactory *spanconfigsqltranslator.Factory
	sqlWatcher           spanconfig.SQLWatcher
	kvAccessor           spanconfig.KVAccessor
	storeWithKVContents  *spanconfigstore.Store
	session              sqlliveness.Session

	execCfg *sql.ExecutorConfig
	codec   keys.SQLCodec
	knobs   *spanconfig.TestingKnobs
}

func (r *incrementalReconciler) reconcile(
	ctx context.Context, startTS hlc.Timestamp, callback func(reconciledUpUntil hlc.Timestamp) error,
) error {
	__antithesis_instrumentation__.Notify(240869)

	return r.sqlWatcher.WatchForSQLUpdates(ctx, startTS,
		func(ctx context.Context, sqlUpdates []spanconfig.SQLUpdate, checkpoint hlc.Timestamp) error {
			__antithesis_instrumentation__.Notify(240870)
			if len(sqlUpdates) == 0 {
				__antithesis_instrumentation__.Notify(240878)
				return callback(checkpoint)
			} else {
				__antithesis_instrumentation__.Notify(240879)
			}
			__antithesis_instrumentation__.Notify(240871)

			var generateSystemSpanConfigurations bool
			var allIDs descpb.IDs
			for _, update := range sqlUpdates {
				__antithesis_instrumentation__.Notify(240880)
				if update.IsDescriptorUpdate() {
					__antithesis_instrumentation__.Notify(240881)
					allIDs = append(allIDs, update.GetDescriptorUpdate().ID)
				} else {
					__antithesis_instrumentation__.Notify(240882)
					if update.IsProtectedTimestampUpdate() {
						__antithesis_instrumentation__.Notify(240883)
						generateSystemSpanConfigurations = true
					} else {
						__antithesis_instrumentation__.Notify(240884)
					}
				}
			}
			__antithesis_instrumentation__.Notify(240872)

			var missingTableIDs []descpb.ID
			var missingProtectedTimestampTargets []spanconfig.SystemTarget
			var records []spanconfig.Record

			if err := sql.DescsTxn(ctx, r.execCfg,
				func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
					__antithesis_instrumentation__.Notify(240885)
					var err error

					missingTableIDs, err = r.filterForMissingTableIDs(ctx, txn, descsCol, sqlUpdates)
					if err != nil {
						__antithesis_instrumentation__.Notify(240888)
						return err
					} else {
						__antithesis_instrumentation__.Notify(240889)
					}
					__antithesis_instrumentation__.Notify(240886)

					missingProtectedTimestampTargets, err = r.filterForMissingProtectedTimestampSystemTargets(
						ctx, txn, sqlUpdates,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(240890)
						return err
					} else {
						__antithesis_instrumentation__.Notify(240891)
					}
					__antithesis_instrumentation__.Notify(240887)

					translator := r.sqlTranslatorFactory.NewSQLTranslator(txn, descsCol)
					records, _, err = translator.Translate(ctx, allIDs, generateSystemSpanConfigurations)
					return err
				}); err != nil {
				__antithesis_instrumentation__.Notify(240892)
				return err
			} else {
				__antithesis_instrumentation__.Notify(240893)
			}
			__antithesis_instrumentation__.Notify(240873)

			updates := make([]spanconfig.Update, 0,
				len(missingTableIDs)+len(missingProtectedTimestampTargets)+len(records))
			for _, entry := range records {
				__antithesis_instrumentation__.Notify(240894)

				updates = append(updates, spanconfig.Update(entry))
			}
			__antithesis_instrumentation__.Notify(240874)
			for _, missingID := range missingTableIDs {
				__antithesis_instrumentation__.Notify(240895)

				tableSpan := roachpb.Span{
					Key:    r.codec.TablePrefix(uint32(missingID)),
					EndKey: r.codec.TablePrefix(uint32(missingID)).PrefixEnd(),
				}
				del, err := spanconfig.Deletion(spanconfig.MakeTargetFromSpan(tableSpan))
				if err != nil {
					__antithesis_instrumentation__.Notify(240897)
					return err
				} else {
					__antithesis_instrumentation__.Notify(240898)
				}
				__antithesis_instrumentation__.Notify(240896)
				updates = append(updates, del)
			}
			__antithesis_instrumentation__.Notify(240875)
			for _, missingSystemTarget := range missingProtectedTimestampTargets {
				__antithesis_instrumentation__.Notify(240899)
				del, err := spanconfig.Deletion(spanconfig.MakeTargetFromSystemTarget(missingSystemTarget))
				if err != nil {
					__antithesis_instrumentation__.Notify(240901)
					return err
				} else {
					__antithesis_instrumentation__.Notify(240902)
				}
				__antithesis_instrumentation__.Notify(240900)
				updates = append(updates, del)
			}
			__antithesis_instrumentation__.Notify(240876)

			toDelete, toUpsert := r.storeWithKVContents.Apply(ctx, false, updates...)
			if len(toDelete) != 0 || func() bool {
				__antithesis_instrumentation__.Notify(240903)
				return len(toUpsert) != 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(240904)
				if err := updateSpanConfigRecords(
					ctx, r.kvAccessor, toDelete, toUpsert, r.session,
				); err != nil {
					__antithesis_instrumentation__.Notify(240905)
					return err
				} else {
					__antithesis_instrumentation__.Notify(240906)
				}
			} else {
				__antithesis_instrumentation__.Notify(240907)
			}
			__antithesis_instrumentation__.Notify(240877)

			return callback(checkpoint)
		},
	)
}

func (r *incrementalReconciler) filterForMissingProtectedTimestampSystemTargets(
	ctx context.Context, txn *kv.Txn, updates []spanconfig.SQLUpdate,
) ([]spanconfig.SystemTarget, error) {
	__antithesis_instrumentation__.Notify(240908)
	seen := make(map[spanconfig.SystemTarget]struct{})
	var missingSystemTargets []spanconfig.SystemTarget
	tenantPrefix := r.codec.TenantPrefix()
	_, sourceTenantID, err := keys.DecodeTenantPrefix(tenantPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(240912)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(240913)
	}
	__antithesis_instrumentation__.Notify(240909)

	ptsState, err := r.execCfg.ProtectedTimestampProvider.GetState(ctx, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(240914)
		return nil, errors.Wrap(err, "failed to get protected timestamp state")
	} else {
		__antithesis_instrumentation__.Notify(240915)
	}
	__antithesis_instrumentation__.Notify(240910)
	ptsStateReader := spanconfig.NewProtectedTimestampStateReader(ctx, ptsState)
	clusterProtections := ptsStateReader.GetProtectionPoliciesForCluster()
	missingClusterProtection := len(clusterProtections) == 0
	for _, update := range updates {
		__antithesis_instrumentation__.Notify(240916)
		if update.IsDescriptorUpdate() {
			__antithesis_instrumentation__.Notify(240920)
			continue
		} else {
			__antithesis_instrumentation__.Notify(240921)
		}
		__antithesis_instrumentation__.Notify(240917)

		ptsUpdate := update.GetProtectedTimestampUpdate()
		missingSystemTarget := spanconfig.SystemTarget{}
		if ptsUpdate.IsClusterUpdate() && func() bool {
			__antithesis_instrumentation__.Notify(240922)
			return missingClusterProtection == true
		}() == true {
			__antithesis_instrumentation__.Notify(240923)

			if r.codec.ForSystemTenant() {
				__antithesis_instrumentation__.Notify(240924)
				missingSystemTarget = spanconfig.MakeEntireKeyspaceTarget()
			} else {
				__antithesis_instrumentation__.Notify(240925)

				missingSystemTarget, err = spanconfig.MakeTenantKeyspaceTarget(sourceTenantID,
					sourceTenantID)
				if err != nil {
					__antithesis_instrumentation__.Notify(240926)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(240927)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(240928)
		}
		__antithesis_instrumentation__.Notify(240918)

		if ptsUpdate.IsTenantsUpdate() {
			__antithesis_instrumentation__.Notify(240929)
			noProtectionsOnTenant :=
				len(ptsStateReader.GetProtectionsForTenant(ptsUpdate.TenantTarget)) == 0
			if noProtectionsOnTenant {
				__antithesis_instrumentation__.Notify(240930)
				missingSystemTarget, err = spanconfig.MakeTenantKeyspaceTarget(sourceTenantID,
					ptsUpdate.TenantTarget)
				if err != nil {
					__antithesis_instrumentation__.Notify(240931)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(240932)
				}
			} else {
				__antithesis_instrumentation__.Notify(240933)
			}
		} else {
			__antithesis_instrumentation__.Notify(240934)
		}
		__antithesis_instrumentation__.Notify(240919)

		if !missingSystemTarget.IsEmpty() {
			__antithesis_instrumentation__.Notify(240935)
			if _, found := seen[missingSystemTarget]; !found {
				__antithesis_instrumentation__.Notify(240936)
				seen[missingSystemTarget] = struct{}{}
				missingSystemTargets = append(missingSystemTargets, missingSystemTarget)
			} else {
				__antithesis_instrumentation__.Notify(240937)
			}
		} else {
			__antithesis_instrumentation__.Notify(240938)
		}
	}
	__antithesis_instrumentation__.Notify(240911)

	return missingSystemTargets, nil
}

func (r *incrementalReconciler) filterForMissingTableIDs(
	ctx context.Context, txn *kv.Txn, descsCol *descs.Collection, updates []spanconfig.SQLUpdate,
) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(240939)
	seen := make(map[descpb.ID]struct{})
	var missingIDs descpb.IDs

	for _, update := range updates {
		__antithesis_instrumentation__.Notify(240941)
		if update.IsProtectedTimestampUpdate() {
			__antithesis_instrumentation__.Notify(240945)
			continue
		} else {
			__antithesis_instrumentation__.Notify(240946)
		}
		__antithesis_instrumentation__.Notify(240942)
		descriptorUpdate := update.GetDescriptorUpdate()
		if descriptorUpdate.Type != catalog.Table {
			__antithesis_instrumentation__.Notify(240947)
			continue
		} else {
			__antithesis_instrumentation__.Notify(240948)
		}
		__antithesis_instrumentation__.Notify(240943)

		desc, err := descsCol.GetImmutableDescriptorByID(ctx, txn, descriptorUpdate.ID, tree.CommonLookupFlags{
			Required:       true,
			IncludeDropped: true,
			IncludeOffline: true,
			AvoidLeased:    true,
		})

		considerAsMissing := false
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(240949)
			considerAsMissing = true
		} else {
			__antithesis_instrumentation__.Notify(240950)
			if err != nil {
				__antithesis_instrumentation__.Notify(240951)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(240952)
				if r.knobs.ExcludeDroppedDescriptorsFromLookup && func() bool {
					__antithesis_instrumentation__.Notify(240953)
					return desc.Dropped() == true
				}() == true {
					__antithesis_instrumentation__.Notify(240954)
					considerAsMissing = true
				} else {
					__antithesis_instrumentation__.Notify(240955)
				}
			}
		}
		__antithesis_instrumentation__.Notify(240944)

		if considerAsMissing {
			__antithesis_instrumentation__.Notify(240956)
			if _, found := seen[descriptorUpdate.ID]; !found {
				__antithesis_instrumentation__.Notify(240957)
				seen[descriptorUpdate.ID] = struct{}{}
				missingIDs = append(missingIDs, descriptorUpdate.ID)
			} else {
				__antithesis_instrumentation__.Notify(240958)
			}
		} else {
			__antithesis_instrumentation__.Notify(240959)
		}
	}
	__antithesis_instrumentation__.Notify(240940)

	return missingIDs, nil
}
