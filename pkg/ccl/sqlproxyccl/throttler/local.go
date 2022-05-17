package throttler

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

var errRequestDenied = errors.New("request denied")

type timeNow func() time.Time

type localService struct {
	clock        timeNow
	maxCacheSize int
	baseDelay    time.Duration
	maxDelay     time.Duration

	mu struct {
		syncutil.Mutex

		throttleCache *cache.UnorderedCache
	}
}

type LocalOption func(s *localService)

func WithBaseDelay(d time.Duration) LocalOption {
	__antithesis_instrumentation__.Notify(23304)
	return func(s *localService) {
		__antithesis_instrumentation__.Notify(23305)
		s.baseDelay = d
	}
}

func NewLocalService(opts ...LocalOption) Service {
	__antithesis_instrumentation__.Notify(23306)
	s := &localService{
		clock:        time.Now,
		maxCacheSize: 1e6,
		baseDelay:    time.Second,
		maxDelay:     time.Hour,
	}
	cacheConfig := cache.Config{
		Policy: cache.CacheLRU,
		ShouldEvict: func(size int, key, value interface{}) bool {
			__antithesis_instrumentation__.Notify(23309)
			return s.maxCacheSize < size
		},
	}
	__antithesis_instrumentation__.Notify(23307)
	s.mu.throttleCache = cache.NewUnorderedCache(cacheConfig)

	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(23310)
		opt(s)
	}
	__antithesis_instrumentation__.Notify(23308)

	return s
}

func (s *localService) lockedGetThrottle(connection ConnectionTags) *throttle {
	__antithesis_instrumentation__.Notify(23311)
	l, ok := s.mu.throttleCache.Get(connection)
	if ok && func() bool {
		__antithesis_instrumentation__.Notify(23313)
		return l != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(23314)
		return l.(*throttle)
	} else {
		__antithesis_instrumentation__.Notify(23315)
	}
	__antithesis_instrumentation__.Notify(23312)
	return nil
}

func (s *localService) lockedInsertThrottle(connection ConnectionTags) *throttle {
	__antithesis_instrumentation__.Notify(23316)
	l := newThrottle(s.baseDelay)
	s.mu.throttleCache.Add(connection, l)
	return l
}

func (s *localService) LoginCheck(connection ConnectionTags) (time.Time, error) {
	__antithesis_instrumentation__.Notify(23317)
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	throttle := s.lockedGetThrottle(connection)
	if throttle != nil && func() bool {
		__antithesis_instrumentation__.Notify(23319)
		return throttle.isThrottled(now) == true
	}() == true {
		__antithesis_instrumentation__.Notify(23320)
		return now, errRequestDenied
	} else {
		__antithesis_instrumentation__.Notify(23321)
	}
	__antithesis_instrumentation__.Notify(23318)
	return now, nil
}

func (s *localService) ReportAttempt(
	ctx context.Context, connection ConnectionTags, throttleTime time.Time, status AttemptStatus,
) error {
	__antithesis_instrumentation__.Notify(23322)
	s.mu.Lock()
	defer s.mu.Unlock()

	throttle := s.lockedGetThrottle(connection)
	if throttle == nil {
		__antithesis_instrumentation__.Notify(23326)
		throttle = s.lockedInsertThrottle(connection)
	} else {
		__antithesis_instrumentation__.Notify(23327)
	}
	__antithesis_instrumentation__.Notify(23323)

	if throttle.isThrottled(throttleTime) {
		__antithesis_instrumentation__.Notify(23328)
		return errRequestDenied
	} else {
		__antithesis_instrumentation__.Notify(23329)
	}
	__antithesis_instrumentation__.Notify(23324)

	switch {
	case status == AttemptInvalidCredentials:
		__antithesis_instrumentation__.Notify(23330)
		throttle.triggerThrottle(s.clock(), s.maxDelay)
		if throttle.nextBackoff == s.maxDelay {
			__antithesis_instrumentation__.Notify(23333)
			log.Warningf(ctx, "connection %v at max throttle delay %s", connection, s.maxDelay)
		} else {
			__antithesis_instrumentation__.Notify(23334)
		}
	case status == AttemptOK:
		__antithesis_instrumentation__.Notify(23331)
		throttle.disable()
	default:
		__antithesis_instrumentation__.Notify(23332)
	}
	__antithesis_instrumentation__.Notify(23325)

	return nil
}
