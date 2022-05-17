package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"
	"sync"

	"github.com/cockroachdb/cmux"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

type loopbackListener struct {
	stopper *stop.Stopper

	closeOnce sync.Once
	active    chan struct{}

	requests chan struct{}

	conns chan net.Conn
}

var _ net.Listener = (*loopbackListener)(nil)

var errLocalListenerClosed = errors.Wrap(cmux.ErrListenerClosed, "loopback listener")

func (l *loopbackListener) Accept() (conn net.Conn, err error) {
	__antithesis_instrumentation__.Notify(194248)
	select {
	case <-l.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(194251)
		return nil, errLocalListenerClosed
	case <-l.active:
		__antithesis_instrumentation__.Notify(194252)
		return nil, errLocalListenerClosed
	case <-l.requests:
		__antithesis_instrumentation__.Notify(194253)
	}
	__antithesis_instrumentation__.Notify(194249)
	c1, c2 := net.Pipe()
	select {
	case l.conns <- c1:
		__antithesis_instrumentation__.Notify(194254)
		return c2, nil
	case <-l.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(194255)
	case <-l.active:
		__antithesis_instrumentation__.Notify(194256)
	}
	__antithesis_instrumentation__.Notify(194250)
	err = errLocalListenerClosed
	err = errors.CombineErrors(err, c1.Close())
	err = errors.CombineErrors(err, c2.Close())
	return nil, err
}

func (l *loopbackListener) Close() error {
	__antithesis_instrumentation__.Notify(194257)
	l.closeOnce.Do(func() {
		__antithesis_instrumentation__.Notify(194259)
		close(l.active)
	})
	__antithesis_instrumentation__.Notify(194258)
	return nil
}

func (l *loopbackListener) Addr() net.Addr {
	__antithesis_instrumentation__.Notify(194260)
	return loopbackAddr{}
}

func (l *loopbackListener) Connect(ctx context.Context) (net.Conn, error) {
	__antithesis_instrumentation__.Notify(194261)

	select {
	case <-l.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(194263)
		return nil, errLocalListenerClosed
	case <-l.active:
		__antithesis_instrumentation__.Notify(194264)
		return nil, errLocalListenerClosed
	case l.requests <- struct{}{}:
		__antithesis_instrumentation__.Notify(194265)
	}
	__antithesis_instrumentation__.Notify(194262)

	select {
	case <-l.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(194266)
		return nil, errLocalListenerClosed
	case <-l.active:
		__antithesis_instrumentation__.Notify(194267)
		return nil, errLocalListenerClosed
	case conn := <-l.conns:
		__antithesis_instrumentation__.Notify(194268)
		return conn, nil
	}
}

func newLoopbackListener(ctx context.Context, stopper *stop.Stopper) *loopbackListener {
	__antithesis_instrumentation__.Notify(194269)
	return &loopbackListener{
		stopper:  stopper,
		active:   make(chan struct{}),
		requests: make(chan struct{}),
		conns:    make(chan net.Conn),
	}
}

type loopbackAddr struct{}

var _ net.Addr = loopbackAddr{}

func (loopbackAddr) Network() string { __antithesis_instrumentation__.Notify(194270); return "pipe" }
func (loopbackAddr) String() string  { __antithesis_instrumentation__.Notify(194271); return "loopback" }
