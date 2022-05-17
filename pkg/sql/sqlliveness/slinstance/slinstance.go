// Package slinstance provides functionality for acquiring sqlliveness leases
// via sessions that have a unique id and expiration. The creation and
// maintenance of session liveness is entrusted to the Instance. Each SQL
// server will have a handle to an instance.
package slinstance

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

var (
	DefaultTTL = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.sqlliveness.ttl",
		"default sqlliveness session ttl",
		40*time.Second,
		settings.NonNegativeDuration,
	)

	DefaultHeartBeat = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.sqlliveness.heartbeat",
		"duration heart beats to push session expiration further out in time",
		5*time.Second,
		settings.NonNegativeDuration,
	)
)

type Writer interface {
	Insert(ctx context.Context, id sqlliveness.SessionID, expiration hlc.Timestamp) error

	Update(ctx context.Context, id sqlliveness.SessionID, expiration hlc.Timestamp) (bool, error)
}

type session struct {
	id    sqlliveness.SessionID
	start hlc.Timestamp

	mu struct {
		syncutil.RWMutex
		exp hlc.Timestamp

		sessionExpiryCallbacks []func(ctx context.Context)
	}
}

func (s *session) ID() sqlliveness.SessionID {
	__antithesis_instrumentation__.Notify(624168)
	return s.id
}

func (s *session) Expiration() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(624169)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mu.exp
}

func (s *session) Start() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(624170)
	return s.start
}

func (s *session) RegisterCallbackForSessionExpiry(sExp func(context.Context)) {
	__antithesis_instrumentation__.Notify(624171)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.sessionExpiryCallbacks = append(s.mu.sessionExpiryCallbacks, sExp)
}

func (s *session) invokeSessionExpiryCallbacks(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624172)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, callback := range s.mu.sessionExpiryCallbacks {
		__antithesis_instrumentation__.Notify(624173)
		callback(ctx)
	}
}

func (s *session) setExpiration(exp hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(624174)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.exp = exp
}

type Instance struct {
	clock     *hlc.Clock
	settings  *cluster.Settings
	stopper   *stop.Stopper
	storage   Writer
	ttl       func() time.Duration
	hb        func() time.Duration
	testKnobs sqlliveness.TestingKnobs
	mu        struct {
		started bool
		syncutil.Mutex
		blockCh chan struct{}
		s       *session
	}
}

func (l *Instance) getSessionOrBlockCh() (*session, chan struct{}) {
	__antithesis_instrumentation__.Notify(624175)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.mu.s != nil {
		__antithesis_instrumentation__.Notify(624177)
		return l.mu.s, nil
	} else {
		__antithesis_instrumentation__.Notify(624178)
	}
	__antithesis_instrumentation__.Notify(624176)
	return nil, l.mu.blockCh
}

func (l *Instance) setSession(s *session) {
	__antithesis_instrumentation__.Notify(624179)
	l.mu.Lock()
	l.mu.s = s

	close(l.mu.blockCh)
	l.mu.Unlock()
}

func (l *Instance) clearSession(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624180)
	l.mu.Lock()
	defer l.mu.Unlock()
	if expiration := l.mu.s.Expiration(); expiration.Less(l.clock.Now()) {
		__antithesis_instrumentation__.Notify(624182)

		l.mu.s.invokeSessionExpiryCallbacks(ctx)
	} else {
		__antithesis_instrumentation__.Notify(624183)
	}
	__antithesis_instrumentation__.Notify(624181)
	l.mu.s = nil
	l.mu.blockCh = make(chan struct{})
}

func (l *Instance) createSession(ctx context.Context) (*session, error) {
	__antithesis_instrumentation__.Notify(624184)
	id := sqlliveness.SessionID(uuid.MakeV4().GetBytes())
	start := l.clock.Now()
	exp := start.Add(l.ttl().Nanoseconds(), 0)
	s := &session{
		id:    id,
		start: start,
	}
	s.mu.exp = exp

	opts := retry.Options{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     2 * time.Second,
		Multiplier:     1.5,
	}
	everySecond := log.Every(time.Second)
	var err error
	for i, r := 0, retry.StartWithCtx(ctx, opts); r.Next(); {
		__antithesis_instrumentation__.Notify(624187)
		i++
		if err = l.storage.Insert(ctx, s.id, s.Expiration()); err != nil {
			__antithesis_instrumentation__.Notify(624189)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(624192)
				break
			} else {
				__antithesis_instrumentation__.Notify(624193)
			}
			__antithesis_instrumentation__.Notify(624190)
			if everySecond.ShouldLog() {
				__antithesis_instrumentation__.Notify(624194)
				log.Errorf(ctx, "failed to create a session at %d-th attempt: %v", i, err)
			} else {
				__antithesis_instrumentation__.Notify(624195)
			}
			__antithesis_instrumentation__.Notify(624191)
			continue
		} else {
			__antithesis_instrumentation__.Notify(624196)
		}
		__antithesis_instrumentation__.Notify(624188)
		break
	}
	__antithesis_instrumentation__.Notify(624185)
	if err != nil {
		__antithesis_instrumentation__.Notify(624197)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624198)
	}
	__antithesis_instrumentation__.Notify(624186)
	log.Infof(ctx, "created new SQL liveness session %s", s.ID())
	return s, nil
}

func (l *Instance) extendSession(ctx context.Context, s *session) (bool, error) {
	__antithesis_instrumentation__.Notify(624199)
	exp := l.clock.Now().Add(l.ttl().Nanoseconds(), 0)

	opts := retry.Options{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     2 * time.Second,
		Multiplier:     1.5,
	}
	var err error
	var found bool
	for r := retry.StartWithCtx(ctx, opts); r.Next(); {
		__antithesis_instrumentation__.Notify(624203)
		if found, err = l.storage.Update(ctx, s.ID(), exp); err != nil {
			__antithesis_instrumentation__.Notify(624205)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(624207)
				break
			} else {
				__antithesis_instrumentation__.Notify(624208)
			}
			__antithesis_instrumentation__.Notify(624206)
			continue
		} else {
			__antithesis_instrumentation__.Notify(624209)
		}
		__antithesis_instrumentation__.Notify(624204)
		break
	}
	__antithesis_instrumentation__.Notify(624200)
	if err != nil {
		__antithesis_instrumentation__.Notify(624210)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(624211)
	}
	__antithesis_instrumentation__.Notify(624201)

	if !found {
		__antithesis_instrumentation__.Notify(624212)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(624213)
	}
	__antithesis_instrumentation__.Notify(624202)

	s.setExpiration(exp)
	return true, nil
}

func (l *Instance) heartbeatLoop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624214)
	defer func() {
		__antithesis_instrumentation__.Notify(624216)
		log.Warning(ctx, "exiting heartbeat loop")
	}()
	__antithesis_instrumentation__.Notify(624215)
	ctx, cancel := l.stopper.WithCancelOnQuiesce(ctx)
	defer cancel()
	t := timeutil.NewTimer()
	t.Reset(0)
	for {
		__antithesis_instrumentation__.Notify(624217)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(624218)
			return
		case <-t.C:
			__antithesis_instrumentation__.Notify(624219)
			t.Read = true
			s, _ := l.getSessionOrBlockCh()
			if s == nil {
				__antithesis_instrumentation__.Notify(624224)
				newSession, err := l.createSession(ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(624226)
					return
				} else {
					__antithesis_instrumentation__.Notify(624227)
				}
				__antithesis_instrumentation__.Notify(624225)
				l.setSession(newSession)
				t.Reset(l.hb())
				continue
			} else {
				__antithesis_instrumentation__.Notify(624228)
			}
			__antithesis_instrumentation__.Notify(624220)
			found, err := l.extendSession(ctx, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(624229)
				l.clearSession(ctx)
				return
			} else {
				__antithesis_instrumentation__.Notify(624230)
			}
			__antithesis_instrumentation__.Notify(624221)
			if !found {
				__antithesis_instrumentation__.Notify(624231)
				l.clearSession(ctx)

				t.Reset(0)
				continue
			} else {
				__antithesis_instrumentation__.Notify(624232)
			}
			__antithesis_instrumentation__.Notify(624222)
			if log.V(2) {
				__antithesis_instrumentation__.Notify(624233)
				log.Infof(ctx, "extended SQL liveness session %s", s.ID())
			} else {
				__antithesis_instrumentation__.Notify(624234)
			}
			__antithesis_instrumentation__.Notify(624223)
			t.Reset(l.hb())
		}
	}
}

func NewSQLInstance(
	stopper *stop.Stopper,
	clock *hlc.Clock,
	storage Writer,
	settings *cluster.Settings,
	testKnobs *sqlliveness.TestingKnobs,
) *Instance {
	__antithesis_instrumentation__.Notify(624235)
	l := &Instance{
		clock:    clock,
		settings: settings,
		storage:  storage,
		stopper:  stopper,
		ttl: func() time.Duration {
			__antithesis_instrumentation__.Notify(624238)
			return DefaultTTL.Get(&settings.SV)
		},
		hb: func() time.Duration {
			__antithesis_instrumentation__.Notify(624239)
			return DefaultHeartBeat.Get(&settings.SV)
		},
	}
	__antithesis_instrumentation__.Notify(624236)
	if testKnobs != nil {
		__antithesis_instrumentation__.Notify(624240)
		l.testKnobs = *testKnobs
	} else {
		__antithesis_instrumentation__.Notify(624241)
	}
	__antithesis_instrumentation__.Notify(624237)
	l.mu.blockCh = make(chan struct{})
	return l
}

func (l *Instance) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624242)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.mu.started {
		__antithesis_instrumentation__.Notify(624244)
		return
	} else {
		__antithesis_instrumentation__.Notify(624245)
	}
	__antithesis_instrumentation__.Notify(624243)
	log.Infof(ctx, "starting SQL liveness instance")
	_ = l.stopper.RunAsyncTask(ctx, "slinstance", l.heartbeatLoop)
	l.mu.started = true
}

func (l *Instance) Session(ctx context.Context) (sqlliveness.Session, error) {
	__antithesis_instrumentation__.Notify(624246)
	if l.testKnobs.SessionOverride != nil {
		__antithesis_instrumentation__.Notify(624249)
		if s, err := l.testKnobs.SessionOverride(ctx); s != nil || func() bool {
			__antithesis_instrumentation__.Notify(624250)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(624251)
			return s, err
		} else {
			__antithesis_instrumentation__.Notify(624252)
		}
	} else {
		__antithesis_instrumentation__.Notify(624253)
	}
	__antithesis_instrumentation__.Notify(624247)
	l.mu.Lock()
	if !l.mu.started {
		__antithesis_instrumentation__.Notify(624254)
		l.mu.Unlock()
		return nil, sqlliveness.NotStartedError
	} else {
		__antithesis_instrumentation__.Notify(624255)
	}
	__antithesis_instrumentation__.Notify(624248)
	l.mu.Unlock()

	for {
		__antithesis_instrumentation__.Notify(624256)
		s, ch := l.getSessionOrBlockCh()
		if s != nil {
			__antithesis_instrumentation__.Notify(624258)
			return s, nil
		} else {
			__antithesis_instrumentation__.Notify(624259)
		}
		__antithesis_instrumentation__.Notify(624257)

		select {
		case <-l.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(624260)
			return nil, stop.ErrUnavailable
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(624261)
			return nil, ctx.Err()
		case <-ch:
			__antithesis_instrumentation__.Notify(624262)
		}
	}
}
