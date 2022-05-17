package ptcache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type Cache struct {
	db       *kv.DB
	storage  protectedts.Storage
	stopper  *stop.Stopper
	settings *cluster.Settings
	sf       singleflight.Group
	mu       struct {
		syncutil.RWMutex

		started bool

		lastUpdate hlc.Timestamp
		state      ptpb.State

		recordsByID map[uuid.UUID]*ptpb.Record
	}
}

type Config struct {
	DB       *kv.DB
	Storage  protectedts.Storage
	Settings *cluster.Settings
}

func New(config Config) *Cache {
	__antithesis_instrumentation__.Notify(110406)
	c := &Cache{
		db:       config.DB,
		storage:  config.Storage,
		settings: config.Settings,
	}
	c.mu.recordsByID = make(map[uuid.UUID]*ptpb.Record)
	return c
}

var _ protectedts.Cache = (*Cache)(nil)

func (c *Cache) Iterate(
	_ context.Context, from, to roachpb.Key, it protectedts.Iterator,
) (asOf hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(110407)
	c.mu.RLock()
	state, lastUpdate := c.mu.state, c.mu.lastUpdate
	c.mu.RUnlock()

	sp := roachpb.Span{
		Key:    from,
		EndKey: to,
	}
	for i := range state.Records {
		__antithesis_instrumentation__.Notify(110409)
		r := &state.Records[i]
		if !overlaps(r, sp) {
			__antithesis_instrumentation__.Notify(110411)
			continue
		} else {
			__antithesis_instrumentation__.Notify(110412)
		}
		__antithesis_instrumentation__.Notify(110410)
		if wantMore := it(r); !wantMore {
			__antithesis_instrumentation__.Notify(110413)
			break
		} else {
			__antithesis_instrumentation__.Notify(110414)
		}
	}
	__antithesis_instrumentation__.Notify(110408)
	return lastUpdate
}

func (c *Cache) QueryRecord(_ context.Context, id uuid.UUID) (exists bool, asOf hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(110415)
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists = c.mu.recordsByID[id]
	return exists, c.mu.lastUpdate
}

const refreshKey = ""

func (c *Cache) Refresh(ctx context.Context, asOf hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(110416)
	for !c.upToDate(asOf) {
		__antithesis_instrumentation__.Notify(110418)
		ch, _ := c.sf.DoChan(refreshKey, c.doSingleFlightUpdate)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(110419)
			return ctx.Err()
		case res := <-ch:
			__antithesis_instrumentation__.Notify(110420)
			if res.Err != nil {
				__antithesis_instrumentation__.Notify(110421)
				return res.Err
			} else {
				__antithesis_instrumentation__.Notify(110422)
			}
		}
	}
	__antithesis_instrumentation__.Notify(110417)
	return nil
}

func (c *Cache) GetProtectionTimestamps(
	ctx context.Context, sp roachpb.Span,
) (protectionTimestamps []hlc.Timestamp, asOf hlc.Timestamp, err error) {
	__antithesis_instrumentation__.Notify(110423)
	readAt := c.Iterate(ctx,
		sp.Key,
		sp.EndKey,
		func(rec *ptpb.Record) (wantMore bool) {
			__antithesis_instrumentation__.Notify(110425)
			protectionTimestamps = append(protectionTimestamps, rec.Timestamp)
			return true
		})
	__antithesis_instrumentation__.Notify(110424)
	return protectionTimestamps, readAt, nil
}

func (c *Cache) Start(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(110426)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mu.started {
		__antithesis_instrumentation__.Notify(110428)
		return errors.New("cannot start a Cache more than once")
	} else {
		__antithesis_instrumentation__.Notify(110429)
	}
	__antithesis_instrumentation__.Notify(110427)
	c.mu.started = true
	c.stopper = stopper
	return c.stopper.RunAsyncTask(ctx, "periodically-refresh-protectedts-cache",
		c.periodicallyRefreshProtectedtsCache)
}

func (c *Cache) periodicallyRefreshProtectedtsCache(ctx context.Context) {
	__antithesis_instrumentation__.Notify(110430)
	settingChanged := make(chan struct{}, 1)
	protectedts.PollInterval.SetOnChange(&c.settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(110432)
		select {
		case settingChanged <- struct{}{}:
			__antithesis_instrumentation__.Notify(110433)
		default:
			__antithesis_instrumentation__.Notify(110434)
		}
	})
	__antithesis_instrumentation__.Notify(110431)
	timer := timeutil.NewTimer()
	defer timer.Stop()
	timer.Reset(0)
	var lastReset time.Time
	var doneCh <-chan singleflight.Result

	for {
		__antithesis_instrumentation__.Notify(110435)
		select {
		case <-timer.C:
			__antithesis_instrumentation__.Notify(110436)

			timer.Read = true
			doneCh, _ = c.sf.DoChan(refreshKey, c.doSingleFlightUpdate)
		case <-settingChanged:
			__antithesis_instrumentation__.Notify(110437)
			if timer.Read {
				__antithesis_instrumentation__.Notify(110442)
				continue
			} else {
				__antithesis_instrumentation__.Notify(110443)
			}
			__antithesis_instrumentation__.Notify(110438)
			interval := protectedts.PollInterval.Get(&c.settings.SV)

			nextUpdate := interval - timeutil.Since(lastReset)
			timer.Reset(nextUpdate)
			lastReset = timeutil.Now()
		case res := <-doneCh:
			__antithesis_instrumentation__.Notify(110439)
			if res.Err != nil {
				__antithesis_instrumentation__.Notify(110444)
				if ctx.Err() == nil {
					__antithesis_instrumentation__.Notify(110445)
					log.Errorf(ctx, "failed to refresh protected timestamps: %v", res.Err)
				} else {
					__antithesis_instrumentation__.Notify(110446)
				}
			} else {
				__antithesis_instrumentation__.Notify(110447)
			}
			__antithesis_instrumentation__.Notify(110440)
			timer.Reset(protectedts.PollInterval.Get(&c.settings.SV))
			lastReset = timeutil.Now()
		case <-c.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(110441)
			return
		}
	}
}

func (c *Cache) doSingleFlightUpdate() (interface{}, error) {
	__antithesis_instrumentation__.Notify(110448)

	ctx, cancel := c.stopper.WithCancelOnQuiesce(context.Background())
	defer cancel()
	return nil, c.stopper.RunTaskWithErr(ctx,
		"refresh-protectedts-cache", c.doUpdate)
}

func (c *Cache) getMetadata() ptpb.Metadata {
	__antithesis_instrumentation__.Notify(110449)
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mu.state.Metadata
}

func (c *Cache) doUpdate(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(110450)

	prev := c.getMetadata()
	var (
		versionChanged bool
		state          ptpb.State
		ts             hlc.Timestamp
	)
	err := c.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(110454)

		defer func() {
			__antithesis_instrumentation__.Notify(110459)
			if err == nil {
				__antithesis_instrumentation__.Notify(110460)
				ts = txn.ReadTimestamp()
			} else {
				__antithesis_instrumentation__.Notify(110461)
			}
		}()
		__antithesis_instrumentation__.Notify(110455)
		md, err := c.storage.GetMetadata(ctx, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(110462)
			return errors.Wrap(err, "failed to fetch protectedts metadata")
		} else {
			__antithesis_instrumentation__.Notify(110463)
		}
		__antithesis_instrumentation__.Notify(110456)
		if versionChanged = md.Version != prev.Version; !versionChanged {
			__antithesis_instrumentation__.Notify(110464)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(110465)
		}
		__antithesis_instrumentation__.Notify(110457)
		if state, err = c.storage.GetState(ctx, txn); err != nil {
			__antithesis_instrumentation__.Notify(110466)
			return errors.Wrap(err, "failed to fetch protectedts state")
		} else {
			__antithesis_instrumentation__.Notify(110467)
		}
		__antithesis_instrumentation__.Notify(110458)
		return nil
	})
	__antithesis_instrumentation__.Notify(110451)
	if err != nil {
		__antithesis_instrumentation__.Notify(110468)
		return err
	} else {
		__antithesis_instrumentation__.Notify(110469)
	}
	__antithesis_instrumentation__.Notify(110452)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mu.lastUpdate = ts
	if versionChanged {
		__antithesis_instrumentation__.Notify(110470)
		c.mu.state = state
		for id := range c.mu.recordsByID {
			__antithesis_instrumentation__.Notify(110472)
			delete(c.mu.recordsByID, id)
		}
		__antithesis_instrumentation__.Notify(110471)
		for i := range state.Records {
			__antithesis_instrumentation__.Notify(110473)
			r := &state.Records[i]
			c.mu.recordsByID[r.ID.GetUUID()] = r
		}
	} else {
		__antithesis_instrumentation__.Notify(110474)
	}
	__antithesis_instrumentation__.Notify(110453)
	return nil
}

func (c *Cache) upToDate(asOf hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(110475)
	c.mu.RLock()
	defer c.mu.RUnlock()
	return asOf.LessEq(c.mu.lastUpdate)
}

func overlaps(r *ptpb.Record, sp roachpb.Span) bool {
	__antithesis_instrumentation__.Notify(110476)
	for i := range r.DeprecatedSpans {
		__antithesis_instrumentation__.Notify(110478)
		if r.DeprecatedSpans[i].Overlaps(sp) {
			__antithesis_instrumentation__.Notify(110479)
			return true
		} else {
			__antithesis_instrumentation__.Notify(110480)
		}
	}
	__antithesis_instrumentation__.Notify(110477)
	return false
}
