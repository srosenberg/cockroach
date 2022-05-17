package instancestorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type Reader struct {
	storage         *Storage
	slReader        sqlliveness.Reader
	f               *rangefeed.Factory
	codec           keys.SQLCodec
	tableID         descpb.ID
	clock           *hlc.Clock
	stopper         *stop.Stopper
	rowcodec        rowCodec
	initialScanDone chan struct{}
	mu              struct {
		syncutil.Mutex
		instances  map[base.SQLInstanceID]instancerow
		startError error
		started    bool
	}
}

func NewTestingReader(
	storage *Storage,
	slReader sqlliveness.Reader,
	f *rangefeed.Factory,
	codec keys.SQLCodec,
	tableID descpb.ID,
	clock *hlc.Clock,
	stopper *stop.Stopper,
) *Reader {
	__antithesis_instrumentation__.Notify(623969)
	r := &Reader{
		storage:         storage,
		slReader:        slReader,
		f:               f,
		codec:           codec,
		tableID:         tableID,
		clock:           clock,
		rowcodec:        makeRowCodec(codec),
		initialScanDone: make(chan struct{}),
		stopper:         stopper,
	}
	r.mu.instances = make(map[base.SQLInstanceID]instancerow)
	return r
}

func NewReader(
	storage *Storage,
	slReader sqlliveness.Reader,
	f *rangefeed.Factory,
	codec keys.SQLCodec,
	clock *hlc.Clock,
	stopper *stop.Stopper,
) *Reader {
	__antithesis_instrumentation__.Notify(623970)
	return NewTestingReader(storage, slReader, f, codec, keys.SQLInstancesTableID, clock, stopper)
}

func (r *Reader) Start(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(623971)
	rf := r.maybeStartRangeFeed(ctx)
	select {
	case <-r.initialScanDone:
		__antithesis_instrumentation__.Notify(623972)

		if rf != nil {
			__antithesis_instrumentation__.Notify(623976)

			r.stopper.AddCloser(rf)
		} else {
			__antithesis_instrumentation__.Notify(623977)
		}
		__antithesis_instrumentation__.Notify(623973)
		return r.checkStarted()
	case <-r.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(623974)
		return errors.Wrap(stop.ErrUnavailable,
			"failed to retrieve initial instance data")
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(623975)
		return errors.Wrap(ctx.Err(),
			"failed to retrieve initial instance data")
	}
}
func (r *Reader) maybeStartRangeFeed(ctx context.Context) *rangefeed.RangeFeed {
	__antithesis_instrumentation__.Notify(623978)
	if r.started() {
		__antithesis_instrumentation__.Notify(623984)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(623985)
	}
	__antithesis_instrumentation__.Notify(623979)
	updateCacheFn := func(
		ctx context.Context, keyVal *roachpb.RangeFeedValue,
	) {
		__antithesis_instrumentation__.Notify(623986)
		instanceID, addr, sessionID, timestamp, tombstone, err := r.rowcodec.decodeRow(kv.KeyValue{
			Key:   keyVal.Key,
			Value: &keyVal.Value,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(623988)
			log.Ops.Warningf(ctx, "failed to decode settings row %v: %v", keyVal.Key, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(623989)
		}
		__antithesis_instrumentation__.Notify(623987)
		instance := instancerow{
			instanceID: instanceID,
			addr:       addr,
			sessionID:  sessionID,
			timestamp:  timestamp,
		}
		r.updateInstanceMap(instance, tombstone)
	}
	__antithesis_instrumentation__.Notify(623980)
	initialScanDoneFn := func(_ context.Context) {
		__antithesis_instrumentation__.Notify(623990)
		close(r.initialScanDone)
	}
	__antithesis_instrumentation__.Notify(623981)
	initialScanErrFn := func(_ context.Context, err error) (shouldFail bool) {
		__antithesis_instrumentation__.Notify(623991)
		if grpcutil.IsAuthError(err) || func() bool {
			__antithesis_instrumentation__.Notify(623993)
			return strings.Contains(err.Error(), "rpc error: code = Unauthenticated") == true
		}() == true {
			__antithesis_instrumentation__.Notify(623994)
			shouldFail = true
			r.setStartError(err)
			close(r.initialScanDone)
		} else {
			__antithesis_instrumentation__.Notify(623995)
		}
		__antithesis_instrumentation__.Notify(623992)
		return shouldFail
	}
	__antithesis_instrumentation__.Notify(623982)

	instancesTablePrefix := r.codec.TablePrefix(uint32(r.tableID))
	instancesTableSpan := roachpb.Span{
		Key:    instancesTablePrefix,
		EndKey: instancesTablePrefix.PrefixEnd(),
	}
	rf, err := r.f.RangeFeed(ctx,
		"sql_instances",
		[]roachpb.Span{instancesTableSpan},
		r.clock.Now(),
		updateCacheFn,
		rangefeed.WithInitialScan(initialScanDoneFn),
		rangefeed.WithOnInitialScanError(initialScanErrFn),
	)
	r.setStarted()
	if err != nil {
		__antithesis_instrumentation__.Notify(623996)
		r.setStartError(err)
		close(r.initialScanDone)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(623997)
	}
	__antithesis_instrumentation__.Notify(623983)
	return rf
}

func (r *Reader) GetInstance(
	ctx context.Context, instanceID base.SQLInstanceID,
) (sqlinstance.InstanceInfo, error) {
	__antithesis_instrumentation__.Notify(623998)
	if err := r.checkStarted(); err != nil {
		__antithesis_instrumentation__.Notify(624003)
		return sqlinstance.InstanceInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(624004)
	}
	__antithesis_instrumentation__.Notify(623999)
	r.mu.Lock()
	instance, ok := r.mu.instances[instanceID]
	r.mu.Unlock()
	if !ok {
		__antithesis_instrumentation__.Notify(624005)
		return sqlinstance.InstanceInfo{}, sqlinstance.NonExistentInstanceError
	} else {
		__antithesis_instrumentation__.Notify(624006)
	}
	__antithesis_instrumentation__.Notify(624000)
	alive, err := r.slReader.IsAlive(ctx, instance.sessionID)
	if err != nil {
		__antithesis_instrumentation__.Notify(624007)
		return sqlinstance.InstanceInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(624008)
	}
	__antithesis_instrumentation__.Notify(624001)
	if !alive {
		__antithesis_instrumentation__.Notify(624009)
		return sqlinstance.InstanceInfo{}, sqlinstance.NonExistentInstanceError
	} else {
		__antithesis_instrumentation__.Notify(624010)
	}
	__antithesis_instrumentation__.Notify(624002)
	instanceInfo := sqlinstance.InstanceInfo{
		InstanceID:   instance.instanceID,
		InstanceAddr: instance.addr,
		SessionID:    instance.sessionID,
	}
	return instanceInfo, nil
}

func (r *Reader) GetAllInstances(
	ctx context.Context,
) (sqlInstances []sqlinstance.InstanceInfo, _ error) {
	__antithesis_instrumentation__.Notify(624011)
	if err := r.checkStarted(); err != nil {
		__antithesis_instrumentation__.Notify(624015)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624016)
	}
	__antithesis_instrumentation__.Notify(624012)
	liveInstances, err := r.getAllLiveInstances(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(624017)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624018)
	}
	__antithesis_instrumentation__.Notify(624013)
	for _, liveInstance := range liveInstances {
		__antithesis_instrumentation__.Notify(624019)
		instanceInfo := sqlinstance.InstanceInfo{
			InstanceID:   liveInstance.instanceID,
			InstanceAddr: liveInstance.addr,
			SessionID:    liveInstance.sessionID,
		}
		sqlInstances = append(sqlInstances, instanceInfo)
	}
	__antithesis_instrumentation__.Notify(624014)
	return sqlInstances, nil
}

func (r *Reader) getAllLiveInstances(ctx context.Context) ([]instancerow, error) {
	__antithesis_instrumentation__.Notify(624020)
	rows := r.getAllInstanceRows()

	{
		__antithesis_instrumentation__.Notify(624023)
		truncated := rows[:0]
		for _, row := range rows {
			__antithesis_instrumentation__.Notify(624025)
			isAlive, err := r.slReader.IsAlive(ctx, row.sessionID)
			if err != nil {
				__antithesis_instrumentation__.Notify(624027)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(624028)
			}
			__antithesis_instrumentation__.Notify(624026)
			if isAlive {
				__antithesis_instrumentation__.Notify(624029)
				truncated = append(truncated, row)
			} else {
				__antithesis_instrumentation__.Notify(624030)
			}
		}
		__antithesis_instrumentation__.Notify(624024)
		rows = truncated
	}
	__antithesis_instrumentation__.Notify(624021)
	sort.Slice(rows, func(idx1, idx2 int) bool {
		__antithesis_instrumentation__.Notify(624031)
		if rows[idx1].addr == rows[idx2].addr {
			__antithesis_instrumentation__.Notify(624033)
			return !rows[idx1].timestamp.Less(rows[idx2].timestamp)
		} else {
			__antithesis_instrumentation__.Notify(624034)
		}
		__antithesis_instrumentation__.Notify(624032)
		return rows[idx1].addr < rows[idx2].addr
	})

	{
		__antithesis_instrumentation__.Notify(624035)
		truncated := rows[:0]
		for i := 0; i < len(rows); i++ {
			__antithesis_instrumentation__.Notify(624037)
			if i == 0 || func() bool {
				__antithesis_instrumentation__.Notify(624038)
				return rows[i].addr != rows[i-1].addr == true
			}() == true {
				__antithesis_instrumentation__.Notify(624039)
				truncated = append(truncated, rows[i])
			} else {
				__antithesis_instrumentation__.Notify(624040)
			}
		}
		__antithesis_instrumentation__.Notify(624036)
		rows = truncated
	}
	__antithesis_instrumentation__.Notify(624022)
	return rows, nil
}

func (r *Reader) getAllInstanceRows() (instances []instancerow) {
	__antithesis_instrumentation__.Notify(624041)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, instance := range r.mu.instances {
		__antithesis_instrumentation__.Notify(624043)
		instances = append(instances, instance)
	}
	__antithesis_instrumentation__.Notify(624042)
	return instances
}

func (r *Reader) updateInstanceMap(instance instancerow, deletionEvent bool) {
	__antithesis_instrumentation__.Notify(624044)
	r.mu.Lock()
	defer r.mu.Unlock()
	if deletionEvent {
		__antithesis_instrumentation__.Notify(624046)
		delete(r.mu.instances, instance.instanceID)
		return
	} else {
		__antithesis_instrumentation__.Notify(624047)
	}
	__antithesis_instrumentation__.Notify(624045)
	r.mu.instances[instance.instanceID] = instance
}

func (r *Reader) setStarted() {
	__antithesis_instrumentation__.Notify(624048)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mu.started = true
}

func (r *Reader) started() bool {
	__antithesis_instrumentation__.Notify(624049)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.mu.started
}

func (r *Reader) checkStarted() error {
	__antithesis_instrumentation__.Notify(624050)
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.mu.started {
		__antithesis_instrumentation__.Notify(624052)
		return sqlinstance.NotStartedError
	} else {
		__antithesis_instrumentation__.Notify(624053)
	}
	__antithesis_instrumentation__.Notify(624051)
	return r.mu.startError
}

func (r *Reader) setStartError(err error) {
	__antithesis_instrumentation__.Notify(624054)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mu.startError = err
}
