// Package instanceprovider provides an implementation of the sqlinstance.provider interface.
package instanceprovider

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance/instancestorage"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type writer interface {
	CreateInstance(ctx context.Context, sessionID sqlliveness.SessionID, sessionExpiration hlc.Timestamp, instanceAddr string) (base.SQLInstanceID, error)
	ReleaseInstanceID(ctx context.Context, instanceID base.SQLInstanceID) error
}

type provider struct {
	*instancestorage.Reader
	storage      writer
	stopper      *stop.Stopper
	instanceAddr string
	session      sqlliveness.Instance
	initOnce     sync.Once
	initialized  chan struct{}
	instanceID   base.SQLInstanceID
	sessionID    sqlliveness.SessionID
	initError    error
	mu           struct {
		syncutil.Mutex
		started bool
	}
}

func New(
	stopper *stop.Stopper,
	db *kv.DB,
	codec keys.SQLCodec,
	slProvider sqlliveness.Provider,
	addr string,
	f *rangefeed.Factory,
	clock *hlc.Clock,
) sqlinstance.Provider {
	__antithesis_instrumentation__.Notify(623916)
	storage := instancestorage.NewStorage(db, codec, slProvider)
	reader := instancestorage.NewReader(storage, slProvider.CachedReader(), f, codec, clock, stopper)
	p := &provider{
		storage:      storage,
		stopper:      stopper,
		Reader:       reader,
		session:      slProvider,
		instanceAddr: addr,
		initialized:  make(chan struct{}),
	}
	return p
}

func (p *provider) Start(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(623917)
	if p.started() {
		__antithesis_instrumentation__.Notify(623920)
		return p.initError
	} else {
		__antithesis_instrumentation__.Notify(623921)
	}
	__antithesis_instrumentation__.Notify(623918)
	if err := p.Reader.Start(ctx); err != nil {
		__antithesis_instrumentation__.Notify(623922)
		p.initOnce.Do(func() {
			__antithesis_instrumentation__.Notify(623923)
			p.initError = err
			close(p.initialized)
		})
	} else {
		__antithesis_instrumentation__.Notify(623924)
	}
	__antithesis_instrumentation__.Notify(623919)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mu.started = true
	return p.initError
}

func (p *provider) started() bool {
	__antithesis_instrumentation__.Notify(623925)
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mu.started
}

func (p *provider) Instance(
	ctx context.Context,
) (_ base.SQLInstanceID, _ sqlliveness.SessionID, err error) {
	__antithesis_instrumentation__.Notify(623926)
	if !p.started() {
		__antithesis_instrumentation__.Notify(623928)
		return base.SQLInstanceID(0), "", sqlinstance.NotStartedError
	} else {
		__antithesis_instrumentation__.Notify(623929)
	}
	__antithesis_instrumentation__.Notify(623927)

	p.maybeInitialize()
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(623930)
		return base.SQLInstanceID(0), "", ctx.Err()
	case <-p.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(623931)
		return base.SQLInstanceID(0), "", stop.ErrUnavailable
	case <-p.initialized:
		__antithesis_instrumentation__.Notify(623932)
		if p.initError == nil {
			__antithesis_instrumentation__.Notify(623934)
			log.Ops.Infof(ctx, "created SQL instance %d", p.instanceID)
		} else {
			__antithesis_instrumentation__.Notify(623935)
			log.Ops.Warningf(ctx, "error creating SQL instance: %s", p.initError)
		}
		__antithesis_instrumentation__.Notify(623933)
		return p.instanceID, p.sessionID, p.initError
	}
}

func (p *provider) maybeInitialize() {
	__antithesis_instrumentation__.Notify(623936)
	p.initOnce.Do(func() {
		__antithesis_instrumentation__.Notify(623937)
		ctx := context.Background()
		if err := p.stopper.RunAsyncTask(ctx, "initialize-instance", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(623938)
			ctx = logtags.AddTag(ctx, "initialize-instance", nil)
			p.initError = p.initialize(ctx)
			close(p.initialized)
		}); err != nil {
			__antithesis_instrumentation__.Notify(623939)
			p.initError = err
			close(p.initialized)
		} else {
			__antithesis_instrumentation__.Notify(623940)
		}
	})
}

func (p *provider) initialize(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(623941)
	session, err := p.session.Session(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(623945)
		return errors.Wrap(err, "constructing session")
	} else {
		__antithesis_instrumentation__.Notify(623946)
	}
	__antithesis_instrumentation__.Notify(623942)
	instanceID, err := p.storage.CreateInstance(ctx, session.ID(), session.Expiration(), p.instanceAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(623947)
		return err
	} else {
		__antithesis_instrumentation__.Notify(623948)
	}
	__antithesis_instrumentation__.Notify(623943)
	p.sessionID = session.ID()
	p.instanceID = instanceID

	session.RegisterCallbackForSessionExpiry(func(_ context.Context) {
		__antithesis_instrumentation__.Notify(623949)

		go func() {
			__antithesis_instrumentation__.Notify(623950)
			ctx, sp := p.stopper.Tracer().StartSpanCtx(context.Background(), "instance shutdown")
			defer sp.Finish()
			p.shutdownSQLInstance(ctx)
		}()
	})
	__antithesis_instrumentation__.Notify(623944)
	return nil
}

func (p *provider) shutdownSQLInstance(ctx context.Context) {
	__antithesis_instrumentation__.Notify(623951)
	if !p.started() {
		__antithesis_instrumentation__.Notify(623957)
		return
	} else {
		__antithesis_instrumentation__.Notify(623958)
	}
	__antithesis_instrumentation__.Notify(623952)

	go func() {
		__antithesis_instrumentation__.Notify(623959)
		p.initOnce.Do(func() {
			__antithesis_instrumentation__.Notify(623960)
			p.initError = errors.New("instance never initialized")
			close(p.initialized)
		})
	}()
	__antithesis_instrumentation__.Notify(623953)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(623961)
		return
	case <-p.initialized:
		__antithesis_instrumentation__.Notify(623962)
	}
	__antithesis_instrumentation__.Notify(623954)

	if p.initError != nil {
		__antithesis_instrumentation__.Notify(623963)
		return
	} else {
		__antithesis_instrumentation__.Notify(623964)
	}
	__antithesis_instrumentation__.Notify(623955)
	err := p.storage.ReleaseInstanceID(ctx, p.instanceID)
	if err != nil {
		__antithesis_instrumentation__.Notify(623965)
		log.Ops.Warningf(ctx, "could not release instance id %d", p.instanceID)
	} else {
		__antithesis_instrumentation__.Notify(623966)
	}
	__antithesis_instrumentation__.Notify(623956)
	p.stopper.Stop(ctx)
}
