package slstorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
)

var GCInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"server.sqlliveness.gc_interval",
	"duration between attempts to delete extant sessions that have expired",
	20*time.Second,
	settings.NonNegativeDuration,
)

var GCJitter = settings.RegisterFloatSetting(
	settings.TenantWritable,
	"server.sqlliveness.gc_jitter",
	"jitter fraction on the duration between attempts to delete extant sessions that have expired",
	.15,
	func(f float64) error {
		__antithesis_instrumentation__.Notify(624268)
		if f < 0 || func() bool {
			__antithesis_instrumentation__.Notify(624270)
			return f > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(624271)
			return errors.Errorf("%f is not in [0, 1]", f)
		} else {
			__antithesis_instrumentation__.Notify(624272)
		}
		__antithesis_instrumentation__.Notify(624269)
		return nil
	},
)

var CacheSize = settings.RegisterIntSetting(
	settings.TenantWritable,
	"server.sqlliveness.storage_session_cache_size",
	"number of session entries to store in the LRU",
	1024)

type Storage struct {
	log.AmbientContext

	settings   *cluster.Settings
	stopper    *stop.Stopper
	clock      *hlc.Clock
	db         *kv.DB
	codec      keys.SQLCodec
	metrics    Metrics
	gcInterval func() time.Duration
	g          singleflight.Group
	newTimer   func() timeutil.TimerI
	tableID    descpb.ID

	mu struct {
		syncutil.Mutex
		started bool

		liveSessions *cache.UnorderedCache

		deadSessions *cache.UnorderedCache
	}
}

func NewTestingStorage(
	ambientCtx log.AmbientContext,
	stopper *stop.Stopper,
	clock *hlc.Clock,
	db *kv.DB,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	sqllivenessTableID descpb.ID,
	newTimer func() timeutil.TimerI,
) *Storage {
	__antithesis_instrumentation__.Notify(624273)
	s := &Storage{
		AmbientContext: ambientCtx,

		settings: settings,
		stopper:  stopper,
		clock:    clock,
		db:       db,
		codec:    codec,
		tableID:  sqllivenessTableID,
		newTimer: newTimer,
		gcInterval: func() time.Duration {
			__antithesis_instrumentation__.Notify(624276)
			baseInterval := GCInterval.Get(&settings.SV)
			jitter := GCJitter.Get(&settings.SV)
			frac := 1 + (2*rand.Float64()-1)*jitter
			return time.Duration(frac * float64(baseInterval.Nanoseconds()))
		},
		metrics: makeMetrics(),
	}
	__antithesis_instrumentation__.Notify(624274)
	cacheConfig := cache.Config{
		Policy: cache.CacheLRU,
		ShouldEvict: func(size int, key, value interface{}) bool {
			__antithesis_instrumentation__.Notify(624277)
			return size > int(CacheSize.Get(&settings.SV))
		},
	}
	__antithesis_instrumentation__.Notify(624275)
	s.mu.liveSessions = cache.NewUnorderedCache(cacheConfig)
	s.mu.deadSessions = cache.NewUnorderedCache(cacheConfig)
	return s
}

func NewStorage(
	ambientCtx log.AmbientContext,
	stopper *stop.Stopper,
	clock *hlc.Clock,
	db *kv.DB,
	codec keys.SQLCodec,
	settings *cluster.Settings,
) *Storage {
	__antithesis_instrumentation__.Notify(624278)
	return NewTestingStorage(ambientCtx, stopper, clock, db, codec, settings, keys.SqllivenessID,
		timeutil.DefaultTimeSource{}.NewTimer)
}

func (s *Storage) Metrics() *Metrics {
	__antithesis_instrumentation__.Notify(624279)
	return &s.metrics
}

func (s *Storage) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624280)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mu.started {
		__antithesis_instrumentation__.Notify(624282)
		return
	} else {
		__antithesis_instrumentation__.Notify(624283)
	}
	__antithesis_instrumentation__.Notify(624281)
	_ = s.stopper.RunAsyncTask(ctx, "slstorage", s.deleteSessionsLoop)
	s.mu.started = true
}

func (s *Storage) IsAlive(ctx context.Context, sid sqlliveness.SessionID) (alive bool, err error) {
	__antithesis_instrumentation__.Notify(624284)
	return s.isAlive(ctx, sid, sync)
}

type readType byte

const (
	_ readType = iota
	sync
	async
)

func (s *Storage) isAlive(
	ctx context.Context, sid sqlliveness.SessionID, syncOrAsync readType,
) (alive bool, _ error) {
	__antithesis_instrumentation__.Notify(624285)
	s.mu.Lock()
	if !s.mu.started {
		__antithesis_instrumentation__.Notify(624291)
		s.mu.Unlock()
		return false, sqlliveness.NotStartedError
	} else {
		__antithesis_instrumentation__.Notify(624292)
	}
	__antithesis_instrumentation__.Notify(624286)
	if _, ok := s.mu.deadSessions.Get(sid); ok {
		__antithesis_instrumentation__.Notify(624293)
		s.mu.Unlock()
		s.metrics.IsAliveCacheHits.Inc(1)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(624294)
	}
	__antithesis_instrumentation__.Notify(624287)
	var prevExpiration hlc.Timestamp
	if expiration, ok := s.mu.liveSessions.Get(sid); ok {
		__antithesis_instrumentation__.Notify(624295)
		expiration := expiration.(hlc.Timestamp)

		if s.clock.Now().Less(expiration) {
			__antithesis_instrumentation__.Notify(624297)
			s.mu.Unlock()
			s.metrics.IsAliveCacheHits.Inc(1)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(624298)
		}
		__antithesis_instrumentation__.Notify(624296)

		prevExpiration = expiration
	} else {
		__antithesis_instrumentation__.Notify(624299)
	}
	__antithesis_instrumentation__.Notify(624288)

	resChan, _ := s.g.DoChan(string(sid), func() (interface{}, error) {
		__antithesis_instrumentation__.Notify(624300)

		bgCtx := s.AnnotateCtx(context.Background())
		bgCtx = logtags.AddTags(bgCtx, logtags.FromContext(ctx))
		newCtx, cancel := s.stopper.WithCancelOnQuiesce(bgCtx)
		defer cancel()

		live, expiration, err := s.deleteOrFetchSession(newCtx, sid, prevExpiration)
		if err != nil {
			__antithesis_instrumentation__.Notify(624303)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(624304)
		}
		__antithesis_instrumentation__.Notify(624301)
		s.mu.Lock()
		defer s.mu.Unlock()
		if live {
			__antithesis_instrumentation__.Notify(624305)
			s.mu.liveSessions.Add(sid, expiration)
		} else {
			__antithesis_instrumentation__.Notify(624306)
			s.mu.deadSessions.Del(sid)
			s.mu.deadSessions.Add(sid, nil)
		}
		__antithesis_instrumentation__.Notify(624302)
		return live, nil
	})
	__antithesis_instrumentation__.Notify(624289)
	s.mu.Unlock()
	s.metrics.IsAliveCacheMisses.Inc(1)

	if syncOrAsync == async {
		__antithesis_instrumentation__.Notify(624307)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(624308)
	}
	__antithesis_instrumentation__.Notify(624290)
	select {
	case res := <-resChan:
		__antithesis_instrumentation__.Notify(624309)
		if res.Err != nil {
			__antithesis_instrumentation__.Notify(624312)
			return false, res.Err
		} else {
			__antithesis_instrumentation__.Notify(624313)
		}
		__antithesis_instrumentation__.Notify(624310)
		return res.Val.(bool), nil
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(624311)
		return false, ctx.Err()
	}
}

func (s *Storage) deleteOrFetchSession(
	ctx context.Context, sid sqlliveness.SessionID, prevExpiration hlc.Timestamp,
) (alive bool, expiration hlc.Timestamp, err error) {
	__antithesis_instrumentation__.Notify(624314)
	var deleted bool
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	if err := s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(624317)
		deleted = false
		k := s.makeSessionKey(sid)
		kv, err := txn.Get(ctx, k)
		if err != nil {
			__antithesis_instrumentation__.Notify(624322)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624323)
		}
		__antithesis_instrumentation__.Notify(624318)

		if kv.Value == nil {
			__antithesis_instrumentation__.Notify(624324)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(624325)
		}
		__antithesis_instrumentation__.Notify(624319)
		expiration, err = decodeValue(kv)
		if err != nil {
			__antithesis_instrumentation__.Notify(624326)
			return errors.Wrapf(err, "failed to decode expiration for %s",
				redact.SafeString(sid.String()))
		} else {
			__antithesis_instrumentation__.Notify(624327)
		}
		__antithesis_instrumentation__.Notify(624320)
		if !expiration.Equal(prevExpiration) {
			__antithesis_instrumentation__.Notify(624328)
			alive = true
			return nil
		} else {
			__antithesis_instrumentation__.Notify(624329)
		}
		__antithesis_instrumentation__.Notify(624321)

		deleted = true
		expiration = hlc.Timestamp{}
		return txn.Del(ctx, k)
	}); err != nil {
		__antithesis_instrumentation__.Notify(624330)
		return false, hlc.Timestamp{}, errors.Wrapf(err,
			"could not query session id: %s", sid)
	} else {
		__antithesis_instrumentation__.Notify(624331)
	}
	__antithesis_instrumentation__.Notify(624315)
	if deleted {
		__antithesis_instrumentation__.Notify(624332)
		s.metrics.SessionsDeleted.Inc(1)
		log.Infof(ctx, "deleted session %s which expired at %s", sid, prevExpiration)
	} else {
		__antithesis_instrumentation__.Notify(624333)
	}
	__antithesis_instrumentation__.Notify(624316)
	return alive, expiration, nil
}

func (s *Storage) deleteSessionsLoop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624334)
	ctx, cancel := s.stopper.WithCancelOnQuiesce(ctx)
	defer cancel()
	t := s.newTimer()
	t.Reset(s.gcInterval())
	for {
		__antithesis_instrumentation__.Notify(624335)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(624336)
			return
		case <-t.Ch():
			__antithesis_instrumentation__.Notify(624337)
			t.MarkRead()
			s.deleteExpiredSessions(ctx)
			t.Reset(s.gcInterval())
		}
	}
}

func (s *Storage) deleteExpiredSessions(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624338)
	now := s.clock.Now()
	var deleted int64
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	if err := s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(624340)
		deleted = 0
		start := s.makeTablePrefix()
		end := start.PrefixEnd()
		const maxRows = 1024
		for {
			__antithesis_instrumentation__.Notify(624341)
			rows, err := txn.Scan(ctx, start, end, maxRows)
			if err != nil {
				__antithesis_instrumentation__.Notify(624346)
				return err
			} else {
				__antithesis_instrumentation__.Notify(624347)
			}
			__antithesis_instrumentation__.Notify(624342)
			if len(rows) == 0 {
				__antithesis_instrumentation__.Notify(624348)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(624349)
			}
			__antithesis_instrumentation__.Notify(624343)
			var toDel []interface{}
			for i := range rows {
				__antithesis_instrumentation__.Notify(624350)
				exp, err := decodeValue(rows[i])
				if err != nil {
					__antithesis_instrumentation__.Notify(624352)
					log.Warningf(ctx, "failed to decode row %s: %v", rows[i].Key.String(), err)
				} else {
					__antithesis_instrumentation__.Notify(624353)
				}
				__antithesis_instrumentation__.Notify(624351)
				if exp.Less(now) {
					__antithesis_instrumentation__.Notify(624354)
					toDel = append(toDel, rows[i].Key)
					deleted++
				} else {
					__antithesis_instrumentation__.Notify(624355)
				}
			}
			__antithesis_instrumentation__.Notify(624344)
			if err := txn.Del(ctx, toDel...); err != nil {
				__antithesis_instrumentation__.Notify(624356)
				return err
			} else {
				__antithesis_instrumentation__.Notify(624357)
			}
			__antithesis_instrumentation__.Notify(624345)
			start = rows[len(rows)-1].Key.Next()
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(624358)
		if ctx.Err() == nil {
			__antithesis_instrumentation__.Notify(624360)
			log.Errorf(ctx, "could not delete expired sessions: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(624361)
		}
		__antithesis_instrumentation__.Notify(624359)
		return
	} else {
		__antithesis_instrumentation__.Notify(624362)
	}
	__antithesis_instrumentation__.Notify(624339)

	s.metrics.SessionDeletionsRuns.Inc(1)
	s.metrics.SessionsDeleted.Inc(deleted)
	if log.V(2) || func() bool {
		__antithesis_instrumentation__.Notify(624363)
		return deleted > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(624364)
		log.Infof(ctx, "deleted %d expired SQL liveness sessions", deleted)
	} else {
		__antithesis_instrumentation__.Notify(624365)
	}
}

func (s *Storage) Insert(
	ctx context.Context, sid sqlliveness.SessionID, expiration hlc.Timestamp,
) (err error) {
	__antithesis_instrumentation__.Notify(624366)
	k := s.makeSessionKey(sid)
	v := encodeValue(expiration)
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	if err := s.db.InitPut(ctx, k, &v, true); err != nil {
		__antithesis_instrumentation__.Notify(624368)
		s.metrics.WriteFailures.Inc(1)
		return errors.Wrapf(err, "could not insert session %s", sid)
	} else {
		__antithesis_instrumentation__.Notify(624369)
	}
	__antithesis_instrumentation__.Notify(624367)
	log.Infof(ctx, "inserted sqlliveness session %s", sid)
	s.metrics.WriteSuccesses.Inc(1)
	return nil
}

func (s *Storage) Update(
	ctx context.Context, sid sqlliveness.SessionID, expiration hlc.Timestamp,
) (sessionExists bool, err error) {
	__antithesis_instrumentation__.Notify(624370)
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	err = s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(624374)
		k := s.makeSessionKey(sid)
		kv, err := txn.Get(ctx, k)
		if err != nil {
			__antithesis_instrumentation__.Notify(624377)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624378)
		}
		__antithesis_instrumentation__.Notify(624375)
		if sessionExists = kv.Value != nil; !sessionExists {
			__antithesis_instrumentation__.Notify(624379)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(624380)
		}
		__antithesis_instrumentation__.Notify(624376)
		v := encodeValue(expiration)
		return txn.Put(ctx, k, &v)
	})
	__antithesis_instrumentation__.Notify(624371)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(624381)
		return !sessionExists == true
	}() == true {
		__antithesis_instrumentation__.Notify(624382)
		s.metrics.WriteFailures.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(624383)
	}
	__antithesis_instrumentation__.Notify(624372)
	if err != nil {
		__antithesis_instrumentation__.Notify(624384)
		return false, errors.Wrapf(err, "could not update session %s", sid)
	} else {
		__antithesis_instrumentation__.Notify(624385)
	}
	__antithesis_instrumentation__.Notify(624373)
	s.metrics.WriteSuccesses.Inc(1)
	return sessionExists, nil
}

func (s *Storage) CachedReader() sqlliveness.Reader {
	__antithesis_instrumentation__.Notify(624386)
	return (*cachedStorage)(s)
}

type cachedStorage Storage

func (s *cachedStorage) IsAlive(
	ctx context.Context, sid sqlliveness.SessionID,
) (alive bool, err error) {
	__antithesis_instrumentation__.Notify(624387)
	return (*Storage)(s).isAlive(ctx, sid, async)
}

func (s *Storage) makeTablePrefix() roachpb.Key {
	__antithesis_instrumentation__.Notify(624388)
	return s.codec.IndexPrefix(uint32(s.tableID), 1)
}

func (s *Storage) makeSessionKey(id sqlliveness.SessionID) roachpb.Key {
	__antithesis_instrumentation__.Notify(624389)
	return keys.MakeFamilyKey(encoding.EncodeBytesAscending(s.makeTablePrefix(), id.UnsafeBytes()), 0)
}

func decodeValue(kv kv.KeyValue) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(624390)
	tup, err := kv.Value.GetTuple()
	if err != nil {
		__antithesis_instrumentation__.Notify(624393)
		return hlc.Timestamp{},
			errors.Wrapf(err, "failed to decode tuple from key %v", kv.Key)
	} else {
		__antithesis_instrumentation__.Notify(624394)
	}
	__antithesis_instrumentation__.Notify(624391)
	_, dec, err := encoding.DecodeDecimalValue(tup)
	if err != nil {
		__antithesis_instrumentation__.Notify(624395)
		return hlc.Timestamp{},
			errors.Wrapf(err, "failed to decode decimal from key %v", kv.Key)
	} else {
		__antithesis_instrumentation__.Notify(624396)
	}
	__antithesis_instrumentation__.Notify(624392)
	return tree.DecimalToHLC(&dec)
}

func encodeValue(expiration hlc.Timestamp) roachpb.Value {
	__antithesis_instrumentation__.Notify(624397)
	var v roachpb.Value
	dec := tree.TimestampToDecimal(expiration)
	v.SetTuple(encoding.EncodeDecimalValue(nil, 2, &dec))
	return v
}
