package schemafeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type TableEvent struct {
	Before, After catalog.TableDescriptor
}

func (e TableEvent) Timestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(17712)
	return e.After.GetModificationTime()
}

type SchemaFeed interface {
	Run(ctx context.Context) error

	Peek(ctx context.Context, atOrBefore hlc.Timestamp) (events []TableEvent, err error)

	Pop(ctx context.Context, atOrBefore hlc.Timestamp) (events []TableEvent, err error)
}

func New(
	ctx context.Context,
	cfg *execinfra.ServerConfig,
	events changefeedbase.SchemaChangeEventClass,
	targets []jobspb.ChangefeedTargetSpecification,
	initialHighwater hlc.Timestamp,
	metrics *Metrics,
	changefeedOpts map[string]string,
) SchemaFeed {
	__antithesis_instrumentation__.Notify(17713)
	m := &schemaFeed{
		filter:            schemaChangeEventFilters[events],
		db:                cfg.DB,
		clock:             cfg.DB.Clock(),
		settings:          cfg.Settings,
		targets:           targets,
		leaseMgr:          cfg.LeaseManager.(*lease.Manager),
		ie:                cfg.SessionBoundInternalExecutorFactory(ctx, &sessiondata.SessionData{}),
		collectionFactory: cfg.CollectionFactory,
		metrics:           metrics,
		changefeedOpts:    changefeedOpts,
	}
	m.mu.previousTableVersion = make(map[descpb.ID]catalog.TableDescriptor)
	m.mu.highWater = initialHighwater
	m.mu.typeDeps = typeDependencyTracker{deps: make(map[descpb.ID][]descpb.ID)}
	return m
}

type schemaFeed struct {
	filter         tableEventFilter
	db             *kv.DB
	clock          *hlc.Clock
	settings       *cluster.Settings
	targets        []jobspb.ChangefeedTargetSpecification
	ie             sqlutil.InternalExecutor
	metrics        *Metrics
	changefeedOpts map[string]string

	leaseMgr          *lease.Manager
	collectionFactory *descs.CollectionFactory

	mu struct {
		syncutil.Mutex

		started bool

		highWater hlc.Timestamp

		errTS hlc.Timestamp

		err error

		waiters []tableHistoryWaiter

		events []TableEvent

		previousTableVersion map[descpb.ID]catalog.TableDescriptor

		typeDeps typeDependencyTracker
	}
}

type typeDependencyTracker struct {
	deps map[descpb.ID][]descpb.ID
}

func (t *typeDependencyTracker) addDependency(typeID, tableID descpb.ID) {
	__antithesis_instrumentation__.Notify(17714)
	deps, ok := t.deps[typeID]
	if !ok {
		__antithesis_instrumentation__.Notify(17715)
		t.deps[typeID] = []descpb.ID{tableID}
	} else {
		__antithesis_instrumentation__.Notify(17716)

		for _, dep := range deps {
			__antithesis_instrumentation__.Notify(17718)
			if dep == tableID {
				__antithesis_instrumentation__.Notify(17719)
				return
			} else {
				__antithesis_instrumentation__.Notify(17720)
			}
		}
		__antithesis_instrumentation__.Notify(17717)
		t.deps[typeID] = append(deps, tableID)
	}
}

func (t *typeDependencyTracker) removeDependency(typeID, tableID descpb.ID) {
	__antithesis_instrumentation__.Notify(17721)
	deps, ok := t.deps[typeID]
	if !ok {
		__antithesis_instrumentation__.Notify(17724)
		return
	} else {
		__antithesis_instrumentation__.Notify(17725)
	}
	__antithesis_instrumentation__.Notify(17722)
	for i := range deps {
		__antithesis_instrumentation__.Notify(17726)
		if deps[i] == tableID {
			__antithesis_instrumentation__.Notify(17727)
			deps = append(deps[:i], deps[i+1:]...)
			break
		} else {
			__antithesis_instrumentation__.Notify(17728)
		}
	}
	__antithesis_instrumentation__.Notify(17723)
	if len(deps) == 0 {
		__antithesis_instrumentation__.Notify(17729)
		delete(t.deps, typeID)
	} else {
		__antithesis_instrumentation__.Notify(17730)
		t.deps[typeID] = deps
	}
}

func (t *typeDependencyTracker) purgeTable(tbl catalog.TableDescriptor) error {
	__antithesis_instrumentation__.Notify(17731)
	for _, col := range tbl.UserDefinedTypeColumns() {
		__antithesis_instrumentation__.Notify(17733)
		id, err := typedesc.UserDefinedTypeOIDToID(col.GetType().Oid())
		if err != nil {
			__antithesis_instrumentation__.Notify(17735)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17736)
		}
		__antithesis_instrumentation__.Notify(17734)
		t.removeDependency(id, tbl.GetID())
	}
	__antithesis_instrumentation__.Notify(17732)

	return nil
}

func (t *typeDependencyTracker) ingestTable(tbl catalog.TableDescriptor) error {
	__antithesis_instrumentation__.Notify(17737)
	for _, col := range tbl.UserDefinedTypeColumns() {
		__antithesis_instrumentation__.Notify(17739)
		id, err := typedesc.UserDefinedTypeOIDToID(col.GetType().Oid())
		if err != nil {
			__antithesis_instrumentation__.Notify(17741)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17742)
		}
		__antithesis_instrumentation__.Notify(17740)
		t.addDependency(id, tbl.GetID())
	}
	__antithesis_instrumentation__.Notify(17738)
	return nil
}

func (t *typeDependencyTracker) containsType(id descpb.ID) bool {
	__antithesis_instrumentation__.Notify(17743)
	_, ok := t.deps[id]
	return ok
}

type tableHistoryWaiter struct {
	ts    hlc.Timestamp
	errCh chan error
}

func (tf *schemaFeed) markStarted() error {
	__antithesis_instrumentation__.Notify(17744)
	tf.mu.Lock()
	defer tf.mu.Unlock()
	if tf.mu.started {
		__antithesis_instrumentation__.Notify(17746)
		return errors.AssertionFailedf("SchemaFeed started more than once")
	} else {
		__antithesis_instrumentation__.Notify(17747)
	}
	__antithesis_instrumentation__.Notify(17745)
	tf.mu.started = true
	return nil
}

func (tf *schemaFeed) Run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17748)
	if err := tf.markStarted(); err != nil {
		__antithesis_instrumentation__.Notify(17751)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17752)
	}
	__antithesis_instrumentation__.Notify(17749)

	if err := tf.primeInitialTableDescs(ctx); err != nil {
		__antithesis_instrumentation__.Notify(17753)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17754)
	}
	__antithesis_instrumentation__.Notify(17750)

	return tf.pollTableHistory(ctx)
}

func (tf *schemaFeed) primeInitialTableDescs(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17755)
	tf.mu.Lock()
	initialTableDescTs := tf.mu.highWater
	tf.mu.Unlock()
	var initialDescs []catalog.Descriptor
	initialTableDescsFn := func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(17759)
		seen := make(map[descpb.ID]struct{}, len(tf.targets))
		initialDescs = initialDescs[:0]
		if err := txn.SetFixedTimestamp(ctx, initialTableDescTs); err != nil {
			__antithesis_instrumentation__.Notify(17762)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17763)
		}
		__antithesis_instrumentation__.Notify(17760)

		for _, table := range tf.targets {
			__antithesis_instrumentation__.Notify(17764)
			if _, dup := seen[table.TableID]; dup {
				__antithesis_instrumentation__.Notify(17767)
				continue
			} else {
				__antithesis_instrumentation__.Notify(17768)
			}
			__antithesis_instrumentation__.Notify(17765)
			seen[table.TableID] = struct{}{}
			flags := tree.ObjectLookupFlagsWithRequired()
			flags.AvoidLeased = true
			tableDesc, err := descriptors.GetImmutableTableByID(ctx, txn, table.TableID, flags)
			if err != nil {
				__antithesis_instrumentation__.Notify(17769)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17770)
			}
			__antithesis_instrumentation__.Notify(17766)
			initialDescs = append(initialDescs, tableDesc)
		}
		__antithesis_instrumentation__.Notify(17761)
		return nil
	}
	__antithesis_instrumentation__.Notify(17756)

	if err := tf.collectionFactory.Txn(
		ctx, tf.ie, tf.db, initialTableDescsFn,
	); err != nil {
		__antithesis_instrumentation__.Notify(17771)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17772)
	}
	__antithesis_instrumentation__.Notify(17757)

	tf.mu.Lock()

	for _, desc := range initialDescs {
		__antithesis_instrumentation__.Notify(17773)
		tbl := desc.(catalog.TableDescriptor)
		if err := tf.mu.typeDeps.ingestTable(tbl); err != nil {
			__antithesis_instrumentation__.Notify(17774)
			tf.mu.Unlock()
			return err
		} else {
			__antithesis_instrumentation__.Notify(17775)
		}
	}
	__antithesis_instrumentation__.Notify(17758)
	tf.mu.Unlock()

	return tf.ingestDescriptors(ctx, hlc.Timestamp{}, initialTableDescTs, initialDescs, tf.validateDescriptor)
}

func (tf *schemaFeed) pollTableHistory(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17776)
	for {
		__antithesis_instrumentation__.Notify(17777)
		if err := tf.updateTableHistory(ctx, tf.clock.Now()); err != nil {
			__antithesis_instrumentation__.Notify(17779)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17780)
		}
		__antithesis_instrumentation__.Notify(17778)

		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(17781)
			return nil
		case <-time.After(changefeedbase.TableDescriptorPollInterval.Get(&tf.settings.SV)):
			__antithesis_instrumentation__.Notify(17782)
		}
	}
}

func (tf *schemaFeed) updateTableHistory(ctx context.Context, endTS hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(17783)
	startTS := tf.highWater()
	if endTS.LessEq(startTS) {
		__antithesis_instrumentation__.Notify(17786)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(17787)
	}
	__antithesis_instrumentation__.Notify(17784)
	descs, err := tf.fetchDescriptorVersions(ctx, tf.leaseMgr.Codec(), tf.db, startTS, endTS)
	if err != nil {
		__antithesis_instrumentation__.Notify(17788)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17789)
	}
	__antithesis_instrumentation__.Notify(17785)
	return tf.ingestDescriptors(ctx, startTS, endTS, descs, tf.validateDescriptor)
}

func (tf *schemaFeed) Peek(
	ctx context.Context, atOrBefore hlc.Timestamp,
) (events []TableEvent, err error) {
	__antithesis_instrumentation__.Notify(17790)

	return tf.peekOrPop(ctx, atOrBefore, false)
}

func (tf *schemaFeed) Pop(
	ctx context.Context, atOrBefore hlc.Timestamp,
) (events []TableEvent, err error) {
	__antithesis_instrumentation__.Notify(17791)
	return tf.peekOrPop(ctx, atOrBefore, true)
}

func (tf *schemaFeed) peekOrPop(
	ctx context.Context, atOrBefore hlc.Timestamp, pop bool,
) (events []TableEvent, err error) {
	__antithesis_instrumentation__.Notify(17792)
	if err = tf.waitForTS(ctx, atOrBefore); err != nil {
		__antithesis_instrumentation__.Notify(17797)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(17798)
	}
	__antithesis_instrumentation__.Notify(17793)
	tf.mu.Lock()
	defer tf.mu.Unlock()
	i := sort.Search(len(tf.mu.events), func(i int) bool {
		__antithesis_instrumentation__.Notify(17799)
		return !tf.mu.events[i].Timestamp().LessEq(atOrBefore)
	})
	__antithesis_instrumentation__.Notify(17794)
	if i == -1 {
		__antithesis_instrumentation__.Notify(17800)
		i = 0
	} else {
		__antithesis_instrumentation__.Notify(17801)
	}
	__antithesis_instrumentation__.Notify(17795)
	events = tf.mu.events[:i]
	if pop {
		__antithesis_instrumentation__.Notify(17802)
		tf.mu.events = tf.mu.events[i:]
	} else {
		__antithesis_instrumentation__.Notify(17803)
	}
	__antithesis_instrumentation__.Notify(17796)
	return events, nil
}

func (tf *schemaFeed) highWater() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(17804)
	tf.mu.Lock()
	highWater := tf.mu.highWater
	tf.mu.Unlock()
	return highWater
}

func (tf *schemaFeed) waitForTS(ctx context.Context, ts hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(17805)
	var errCh chan error

	tf.mu.Lock()
	highWater := tf.mu.highWater
	var err error
	if !tf.mu.errTS.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(17810)
		return tf.mu.errTS.LessEq(ts) == true
	}() == true {
		__antithesis_instrumentation__.Notify(17811)
		err = tf.mu.err
	} else {
		__antithesis_instrumentation__.Notify(17812)
	}
	__antithesis_instrumentation__.Notify(17806)
	fastPath := err != nil || func() bool {
		__antithesis_instrumentation__.Notify(17813)
		return ts.LessEq(highWater) == true
	}() == true
	if !fastPath {
		__antithesis_instrumentation__.Notify(17814)
		errCh = make(chan error, 1)
		tf.mu.waiters = append(tf.mu.waiters, tableHistoryWaiter{ts: ts, errCh: errCh})
	} else {
		__antithesis_instrumentation__.Notify(17815)
	}
	__antithesis_instrumentation__.Notify(17807)
	tf.mu.Unlock()
	if fastPath {
		__antithesis_instrumentation__.Notify(17816)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(17818)
			log.Infof(ctx, "fastpath for %s: %v", ts, err)
		} else {
			__antithesis_instrumentation__.Notify(17819)
		}
		__antithesis_instrumentation__.Notify(17817)
		return err
	} else {
		__antithesis_instrumentation__.Notify(17820)
	}
	__antithesis_instrumentation__.Notify(17808)

	if log.V(1) {
		__antithesis_instrumentation__.Notify(17821)
		log.Infof(ctx, "waiting for %s highwater", ts)
	} else {
		__antithesis_instrumentation__.Notify(17822)
	}
	__antithesis_instrumentation__.Notify(17809)
	start := timeutil.Now()
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(17823)
		return ctx.Err()
	case err := <-errCh:
		__antithesis_instrumentation__.Notify(17824)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(17827)
			log.Infof(ctx, "waited %s for %s highwater: %v", timeutil.Since(start), ts, err)
		} else {
			__antithesis_instrumentation__.Notify(17828)
		}
		__antithesis_instrumentation__.Notify(17825)
		if tf.metrics != nil {
			__antithesis_instrumentation__.Notify(17829)
			tf.metrics.TableMetadataNanos.Inc(timeutil.Since(start).Nanoseconds())
		} else {
			__antithesis_instrumentation__.Notify(17830)
		}
		__antithesis_instrumentation__.Notify(17826)
		return err
	}
}

func descLess(a, b catalog.Descriptor) bool {
	__antithesis_instrumentation__.Notify(17831)
	aTime, bTime := a.GetModificationTime(), b.GetModificationTime()
	if aTime.Equal(bTime) {
		__antithesis_instrumentation__.Notify(17833)
		return a.GetID() < b.GetID()
	} else {
		__antithesis_instrumentation__.Notify(17834)
	}
	__antithesis_instrumentation__.Notify(17832)
	return aTime.Less(bTime)
}

func (tf *schemaFeed) ingestDescriptors(
	ctx context.Context,
	startTS, endTS hlc.Timestamp,
	descs []catalog.Descriptor,
	validateFn func(ctx context.Context, earliestTsBeingIngested hlc.Timestamp, desc catalog.Descriptor) error,
) error {
	__antithesis_instrumentation__.Notify(17835)
	sort.Slice(descs, func(i, j int) bool { __antithesis_instrumentation__.Notify(17838); return descLess(descs[i], descs[j]) })
	__antithesis_instrumentation__.Notify(17836)
	var validateErr error
	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(17839)
		if err := validateFn(ctx, startTS, desc); validateErr == nil {
			__antithesis_instrumentation__.Notify(17840)
			validateErr = err
		} else {
			__antithesis_instrumentation__.Notify(17841)
		}
	}
	__antithesis_instrumentation__.Notify(17837)
	return tf.adjustTimestamps(startTS, endTS, validateErr)
}

func (tf *schemaFeed) adjustTimestamps(startTS, endTS hlc.Timestamp, validateErr error) error {
	__antithesis_instrumentation__.Notify(17842)
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if validateErr != nil {
		__antithesis_instrumentation__.Notify(17846)

		if tf.mu.errTS.IsEmpty() || func() bool {
			__antithesis_instrumentation__.Notify(17848)
			return endTS.Less(tf.mu.errTS) == true
		}() == true {
			__antithesis_instrumentation__.Notify(17849)
			tf.mu.errTS = endTS
			tf.mu.err = validateErr
			newWaiters := make([]tableHistoryWaiter, 0, len(tf.mu.waiters))
			for _, w := range tf.mu.waiters {
				__antithesis_instrumentation__.Notify(17851)
				if w.ts.Less(tf.mu.errTS) {
					__antithesis_instrumentation__.Notify(17853)
					newWaiters = append(newWaiters, w)
					continue
				} else {
					__antithesis_instrumentation__.Notify(17854)
				}
				__antithesis_instrumentation__.Notify(17852)
				w.errCh <- validateErr
			}
			__antithesis_instrumentation__.Notify(17850)
			tf.mu.waiters = newWaiters
		} else {
			__antithesis_instrumentation__.Notify(17855)
		}
		__antithesis_instrumentation__.Notify(17847)
		return validateErr
	} else {
		__antithesis_instrumentation__.Notify(17856)
	}
	__antithesis_instrumentation__.Notify(17843)

	if tf.mu.highWater.Less(startTS) {
		__antithesis_instrumentation__.Notify(17857)
		return errors.Errorf(`gap between %s and %s`, tf.mu.highWater, startTS)
	} else {
		__antithesis_instrumentation__.Notify(17858)
	}
	__antithesis_instrumentation__.Notify(17844)
	if tf.mu.highWater.Less(endTS) {
		__antithesis_instrumentation__.Notify(17859)
		tf.mu.highWater = endTS
		newWaiters := make([]tableHistoryWaiter, 0, len(tf.mu.waiters))
		for _, w := range tf.mu.waiters {
			__antithesis_instrumentation__.Notify(17861)
			if tf.mu.highWater.Less(w.ts) {
				__antithesis_instrumentation__.Notify(17863)
				newWaiters = append(newWaiters, w)
				continue
			} else {
				__antithesis_instrumentation__.Notify(17864)
			}
			__antithesis_instrumentation__.Notify(17862)
			w.errCh <- nil
		}
		__antithesis_instrumentation__.Notify(17860)
		tf.mu.waiters = newWaiters
	} else {
		__antithesis_instrumentation__.Notify(17865)
	}
	__antithesis_instrumentation__.Notify(17845)
	return nil
}
func (e TableEvent) String() string {
	__antithesis_instrumentation__.Notify(17866)
	return formatEvent(e)
}

func formatDesc(desc catalog.TableDescriptor) string {
	__antithesis_instrumentation__.Notify(17867)
	return fmt.Sprintf("%d:%d@%v", desc.GetID(), desc.GetVersion(), desc.GetModificationTime())
}

func formatEvent(e TableEvent) string {
	__antithesis_instrumentation__.Notify(17868)
	return fmt.Sprintf("%v->%v", formatDesc(e.Before), formatDesc(e.After))
}

func (tf *schemaFeed) validateDescriptor(
	ctx context.Context, earliestTsBeingIngested hlc.Timestamp, desc catalog.Descriptor,
) error {
	__antithesis_instrumentation__.Notify(17869)
	tf.mu.Lock()
	defer tf.mu.Unlock()
	switch desc := desc.(type) {
	case catalog.TypeDescriptor:
		__antithesis_instrumentation__.Notify(17870)
		if !tf.mu.typeDeps.containsType(desc.GetID()) {
			__antithesis_instrumentation__.Notify(17877)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(17878)
		}
		__antithesis_instrumentation__.Notify(17871)

		return tf.leaseMgr.AcquireFreshestFromStore(ctx, desc.GetID())
	case catalog.TableDescriptor:
		__antithesis_instrumentation__.Notify(17872)
		if err := changefeedbase.ValidateTable(tf.targets, desc, tf.changefeedOpts); err != nil {
			__antithesis_instrumentation__.Notify(17879)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17880)
		}
		__antithesis_instrumentation__.Notify(17873)
		log.VEventf(ctx, 1, "validate %v", formatDesc(desc))
		if lastVersion, ok := tf.mu.previousTableVersion[desc.GetID()]; ok {
			__antithesis_instrumentation__.Notify(17881)

			if desc.GetModificationTime().LessEq(lastVersion.GetModificationTime()) {
				__antithesis_instrumentation__.Notify(17886)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(17887)
			}
			__antithesis_instrumentation__.Notify(17882)

			if err := tf.leaseMgr.AcquireFreshestFromStore(ctx, desc.GetID()); err != nil {
				__antithesis_instrumentation__.Notify(17888)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17889)
			}
			__antithesis_instrumentation__.Notify(17883)

			if err := tf.mu.typeDeps.purgeTable(lastVersion); err != nil {
				__antithesis_instrumentation__.Notify(17890)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17891)
			}
			__antithesis_instrumentation__.Notify(17884)

			e := TableEvent{
				Before: lastVersion,
				After:  desc,
			}
			shouldFilter, err := tf.filter.shouldFilter(ctx, e)
			log.VEventf(ctx, 1, "validate shouldFilter %v %v", formatEvent(e), shouldFilter)
			if err != nil {
				__antithesis_instrumentation__.Notify(17892)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17893)
			}
			__antithesis_instrumentation__.Notify(17885)
			if !shouldFilter {
				__antithesis_instrumentation__.Notify(17894)

				idxToSort := sort.Search(len(tf.mu.events), func(i int) bool {
					__antithesis_instrumentation__.Notify(17896)
					return !tf.mu.events[i].After.GetModificationTime().Less(earliestTsBeingIngested)
				})
				__antithesis_instrumentation__.Notify(17895)
				tf.mu.events = append(tf.mu.events, e)
				toSort := tf.mu.events[idxToSort:]
				sort.Slice(toSort, func(i, j int) bool {
					__antithesis_instrumentation__.Notify(17897)
					return descLess(toSort[i].After, toSort[j].After)
				})
			} else {
				__antithesis_instrumentation__.Notify(17898)
			}
		} else {
			__antithesis_instrumentation__.Notify(17899)
		}
		__antithesis_instrumentation__.Notify(17874)

		if err := tf.mu.typeDeps.ingestTable(desc); err != nil {
			__antithesis_instrumentation__.Notify(17900)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17901)
		}
		__antithesis_instrumentation__.Notify(17875)
		tf.mu.previousTableVersion[desc.GetID()] = desc
		return nil
	default:
		__antithesis_instrumentation__.Notify(17876)
		return errors.AssertionFailedf("unexpected descriptor type %T", desc)
	}
}

func (tf *schemaFeed) fetchDescriptorVersions(
	ctx context.Context, codec keys.SQLCodec, db *kv.DB, startTS, endTS hlc.Timestamp,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(17902)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(17907)
		log.Infof(ctx, `fetching table descs (%s,%s]`, startTS, endTS)
	} else {
		__antithesis_instrumentation__.Notify(17908)
	}
	__antithesis_instrumentation__.Notify(17903)
	start := timeutil.Now()
	span := roachpb.Span{Key: codec.TablePrefix(keys.DescriptorTableID)}
	span.EndKey = span.Key.PrefixEnd()
	header := roachpb.Header{Timestamp: endTS}
	req := &roachpb.ExportRequest{
		RequestHeader: roachpb.RequestHeaderFromSpan(span),
		StartTime:     startTS,
		MVCCFilter:    roachpb.MVCCFilter_All,
		ReturnSST:     true,
	}
	res, pErr := kv.SendWrappedWith(ctx, db.NonTransactionalSender(), header, req)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(17909)
		log.Infof(ctx, `fetched table descs (%s,%s] took %s`, startTS, endTS, timeutil.Since(start))
	} else {
		__antithesis_instrumentation__.Notify(17910)
	}
	__antithesis_instrumentation__.Notify(17904)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(17911)
		err := pErr.GoError()
		return nil, errors.Wrapf(err, `fetching changes for %s`, span)
	} else {
		__antithesis_instrumentation__.Notify(17912)
	}
	__antithesis_instrumentation__.Notify(17905)

	tf.mu.Lock()
	defer tf.mu.Unlock()

	var descriptors []catalog.Descriptor
	for _, file := range res.(*roachpb.ExportResponse).Files {
		__antithesis_instrumentation__.Notify(17913)
		if err := func() error {
			__antithesis_instrumentation__.Notify(17914)
			it, err := storage.NewMemSSTIterator(file.SST, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(17916)
				return err
			} else {
				__antithesis_instrumentation__.Notify(17917)
			}
			__antithesis_instrumentation__.Notify(17915)
			defer it.Close()
			for it.SeekGE(storage.NilKey); ; it.Next() {
				__antithesis_instrumentation__.Notify(17918)
				if ok, err := it.Valid(); err != nil {
					__antithesis_instrumentation__.Notify(17926)
					return err
				} else {
					__antithesis_instrumentation__.Notify(17927)
					if !ok {
						__antithesis_instrumentation__.Notify(17928)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(17929)
					}
				}
				__antithesis_instrumentation__.Notify(17919)
				k := it.UnsafeKey()
				remaining, _, _, err := codec.DecodeIndexPrefix(k.Key)
				if err != nil {
					__antithesis_instrumentation__.Notify(17930)
					return err
				} else {
					__antithesis_instrumentation__.Notify(17931)
				}
				__antithesis_instrumentation__.Notify(17920)
				_, id, err := encoding.DecodeUvarintAscending(remaining)
				if err != nil {
					__antithesis_instrumentation__.Notify(17932)
					return err
				} else {
					__antithesis_instrumentation__.Notify(17933)
				}
				__antithesis_instrumentation__.Notify(17921)
				var origName string
				var isTable bool
				for _, cts := range tf.targets {
					__antithesis_instrumentation__.Notify(17934)
					if cts.TableID == descpb.ID(id) {
						__antithesis_instrumentation__.Notify(17935)
						origName = cts.StatementTimeName
						isTable = true
						break
					} else {
						__antithesis_instrumentation__.Notify(17936)
					}
				}
				__antithesis_instrumentation__.Notify(17922)
				isType := tf.mu.typeDeps.containsType(descpb.ID(id))

				if !(isTable || func() bool {
					__antithesis_instrumentation__.Notify(17937)
					return isType == true
				}() == true) {
					__antithesis_instrumentation__.Notify(17938)

					continue
				} else {
					__antithesis_instrumentation__.Notify(17939)
				}
				__antithesis_instrumentation__.Notify(17923)

				unsafeValue := it.UnsafeValue()
				if unsafeValue == nil {
					__antithesis_instrumentation__.Notify(17940)
					name := origName
					if name == "" {
						__antithesis_instrumentation__.Notify(17942)
						name = fmt.Sprintf("desc(%d)", id)
					} else {
						__antithesis_instrumentation__.Notify(17943)
					}
					__antithesis_instrumentation__.Notify(17941)
					return errors.Errorf(`"%v" was dropped or truncated`, name)
				} else {
					__antithesis_instrumentation__.Notify(17944)
				}
				__antithesis_instrumentation__.Notify(17924)

				value := roachpb.Value{RawBytes: unsafeValue}
				var desc descpb.Descriptor
				if err := value.GetProto(&desc); err != nil {
					__antithesis_instrumentation__.Notify(17945)
					return err
				} else {
					__antithesis_instrumentation__.Notify(17946)
				}
				__antithesis_instrumentation__.Notify(17925)

				b := descbuilder.NewBuilderWithMVCCTimestamp(&desc, k.Timestamp)
				if b != nil && func() bool {
					__antithesis_instrumentation__.Notify(17947)
					return (b.DescriptorType() == catalog.Table || func() bool {
						__antithesis_instrumentation__.Notify(17948)
						return b.DescriptorType() == catalog.Type == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(17949)
					descriptors = append(descriptors, b.BuildImmutable())
				} else {
					__antithesis_instrumentation__.Notify(17950)
				}
			}
		}(); err != nil {
			__antithesis_instrumentation__.Notify(17951)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(17952)
		}
	}
	__antithesis_instrumentation__.Notify(17906)
	return descriptors, nil
}

type doNothingSchemaFeed struct{}

func (f doNothingSchemaFeed) Run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17953)
	<-ctx.Done()
	return ctx.Err()
}

func (f doNothingSchemaFeed) Peek(
	ctx context.Context, atOrBefore hlc.Timestamp,
) (events []TableEvent, err error) {
	__antithesis_instrumentation__.Notify(17954)
	return nil, nil
}

func (f doNothingSchemaFeed) Pop(
	ctx context.Context, atOrBefore hlc.Timestamp,
) (events []TableEvent, err error) {
	__antithesis_instrumentation__.Notify(17955)
	return nil, nil
}

var DoNothingSchemaFeed SchemaFeed = &doNothingSchemaFeed{}
