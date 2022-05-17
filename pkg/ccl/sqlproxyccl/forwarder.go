package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/balancer"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/interceptor"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type forwarder struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	connector *connector

	metrics *metrics

	errCh chan error

	mu struct {
		syncutil.Mutex

		isTransferring bool

		clientConn *interceptor.PGConn
		serverConn *interceptor.PGConn

		request  *processor
		response *processor
	}
}

var _ balancer.ConnectionHandle = &forwarder{}

func forward(
	ctx context.Context,
	connector *connector,
	metrics *metrics,
	clientConn net.Conn,
	serverConn net.Conn,
) (*forwarder, error) {
	__antithesis_instrumentation__.Notify(21510)
	ctx, cancelFn := context.WithCancel(ctx)
	f := &forwarder{
		ctx:       ctx,
		ctxCancel: cancelFn,
		errCh:     make(chan error, 1),
		connector: connector,
		metrics:   metrics,
	}
	f.mu.clientConn = interceptor.NewPGConn(clientConn)
	f.mu.serverConn = interceptor.NewPGConn(serverConn)

	clockFn := makeLogicalClockFn()
	f.mu.request = newProcessor(clockFn, f.mu.clientConn, f.mu.serverConn)
	f.mu.response = newProcessor(clockFn, f.mu.serverConn, f.mu.clientConn)
	if err := f.resumeProcessors(); err != nil {
		__antithesis_instrumentation__.Notify(21512)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21513)
	}
	__antithesis_instrumentation__.Notify(21511)
	return f, nil
}

func (f *forwarder) Context() context.Context {
	__antithesis_instrumentation__.Notify(21514)
	return f.ctx
}

func (f *forwarder) Close() {
	__antithesis_instrumentation__.Notify(21515)
	f.ctxCancel()

	select {
	case f.errCh <- errors.New("forwarder closed"):
		__antithesis_instrumentation__.Notify(21517)
	default:
		__antithesis_instrumentation__.Notify(21518)
	}
	__antithesis_instrumentation__.Notify(21516)

	clientConn, serverConn := f.getConns()
	clientConn.Close()
	serverConn.Close()
}

func (f *forwarder) ServerRemoteAddr() string {
	__antithesis_instrumentation__.Notify(21519)
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.mu.serverConn.RemoteAddr().String()
}

func (f *forwarder) resumeProcessors() error {
	__antithesis_instrumentation__.Notify(21520)
	requestProc, responseProc := f.getProcessors()
	go func() {
		__antithesis_instrumentation__.Notify(21525)
		if err := requestProc.resume(f.ctx); err != nil {
			__antithesis_instrumentation__.Notify(21526)
			f.tryReportError(wrapClientToServerError(err))
		} else {
			__antithesis_instrumentation__.Notify(21527)
		}
	}()
	__antithesis_instrumentation__.Notify(21521)
	go func() {
		__antithesis_instrumentation__.Notify(21528)
		if err := responseProc.resume(f.ctx); err != nil {
			__antithesis_instrumentation__.Notify(21529)
			f.tryReportError(wrapServerToClientError(err))
		} else {
			__antithesis_instrumentation__.Notify(21530)
		}
	}()
	__antithesis_instrumentation__.Notify(21522)
	if err := requestProc.waitResumed(f.ctx); err != nil {
		__antithesis_instrumentation__.Notify(21531)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21532)
	}
	__antithesis_instrumentation__.Notify(21523)
	if err := responseProc.waitResumed(f.ctx); err != nil {
		__antithesis_instrumentation__.Notify(21533)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21534)
	}
	__antithesis_instrumentation__.Notify(21524)
	return nil
}

func (f *forwarder) tryReportError(err error) {
	__antithesis_instrumentation__.Notify(21535)
	select {
	case f.errCh <- err:
		__antithesis_instrumentation__.Notify(21536)

		f.Close()
	default:
		__antithesis_instrumentation__.Notify(21537)
	}
}

func (f *forwarder) getProcessors() (request, response *processor) {
	__antithesis_instrumentation__.Notify(21538)
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.mu.request, f.mu.response
}

func (f *forwarder) getConns() (client, server *interceptor.PGConn) {
	__antithesis_instrumentation__.Notify(21539)
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.mu.clientConn, f.mu.serverConn
}

func (f *forwarder) replaceServerConn(newServerConn *interceptor.PGConn) {
	__antithesis_instrumentation__.Notify(21540)
	f.mu.Lock()
	defer f.mu.Unlock()
	clockFn := makeLogicalClockFn()
	f.mu.serverConn.Close()
	f.mu.serverConn = newServerConn
	f.mu.request = newProcessor(clockFn, f.mu.clientConn, f.mu.serverConn)
	f.mu.response = newProcessor(clockFn, f.mu.serverConn, f.mu.clientConn)
}

func wrapClientToServerError(err error) error {
	__antithesis_instrumentation__.Notify(21541)
	if err == nil || func() bool {
		__antithesis_instrumentation__.Notify(21543)
		return errors.IsAny(err, context.Canceled, context.DeadlineExceeded) == true
	}() == true {
		__antithesis_instrumentation__.Notify(21544)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(21545)
	}
	__antithesis_instrumentation__.Notify(21542)
	return newErrorf(codeClientDisconnected, "copying from client to target server: %v", err)
}

func wrapServerToClientError(err error) error {
	__antithesis_instrumentation__.Notify(21546)
	if err == nil || func() bool {
		__antithesis_instrumentation__.Notify(21548)
		return errors.IsAny(err, context.Canceled, context.DeadlineExceeded) == true
	}() == true {
		__antithesis_instrumentation__.Notify(21549)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(21550)
	}
	__antithesis_instrumentation__.Notify(21547)
	return newErrorf(codeBackendDisconnected, "copying from target server to client: %s", err)
}

func makeLogicalClockFn() func() uint64 {
	__antithesis_instrumentation__.Notify(21551)
	var counter uint64
	return func() uint64 {
		__antithesis_instrumentation__.Notify(21552)
		return atomic.AddUint64(&counter, 1)
	}
}

var aLongTimeAgo = timeutil.Unix(1, 0)

var errProcessorResumed = errors.New("processor has already been resumed")

type processor struct {
	src *interceptor.PGConn
	dst *interceptor.PGConn

	mu struct {
		syncutil.Mutex
		cond       *sync.Cond
		resumed    bool
		inPeek     bool
		suspendReq bool

		lastMessageTransferredAt uint64
		lastMessageType          byte
	}
	logicalClockFn func() uint64

	testingKnobs struct {
		beforeForwardMsg func()
	}
}

func newProcessor(logicalClockFn func() uint64, src, dst *interceptor.PGConn) *processor {
	__antithesis_instrumentation__.Notify(21553)
	p := &processor{logicalClockFn: logicalClockFn, src: src, dst: dst}
	p.mu.cond = sync.NewCond(&p.mu)
	return p
}

func (p *processor) resume(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(21554)
	enterResume := func() error {
		__antithesis_instrumentation__.Notify(21560)
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.mu.resumed {
			__antithesis_instrumentation__.Notify(21562)
			return errProcessorResumed
		} else {
			__antithesis_instrumentation__.Notify(21563)
		}
		__antithesis_instrumentation__.Notify(21561)
		p.mu.resumed = true
		p.mu.cond.Broadcast()
		return nil
	}
	__antithesis_instrumentation__.Notify(21555)
	exitResume := func() {
		__antithesis_instrumentation__.Notify(21564)
		p.mu.Lock()
		defer p.mu.Unlock()
		p.mu.resumed = false
		p.mu.cond.Broadcast()
	}
	__antithesis_instrumentation__.Notify(21556)
	prepareNextMessage := func() (terminate bool, err error) {
		__antithesis_instrumentation__.Notify(21565)

		if terminate := func() bool {
			__antithesis_instrumentation__.Notify(21568)
			p.mu.Lock()
			defer p.mu.Unlock()

			if p.mu.suspendReq {
				__antithesis_instrumentation__.Notify(21570)
				return true
			} else {
				__antithesis_instrumentation__.Notify(21571)
			}
			__antithesis_instrumentation__.Notify(21569)
			p.mu.inPeek = true
			return false
		}(); terminate {
			__antithesis_instrumentation__.Notify(21572)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(21573)
		}
		__antithesis_instrumentation__.Notify(21566)

		typ, _, peekErr := p.src.PeekMsg()

		p.mu.Lock()
		defer p.mu.Unlock()
		p.mu.inPeek = false

		var netErr net.Error
		switch {
		case p.mu.suspendReq && func() bool {
			__antithesis_instrumentation__.Notify(21578)
			return peekErr == nil == true
		}() == true:
			__antithesis_instrumentation__.Notify(21574)
			return true, nil
		case p.mu.suspendReq && func() bool {
			__antithesis_instrumentation__.Notify(21579)
			return errors.As(peekErr, &netErr) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(21580)
			return netErr.Timeout() == true
		}() == true:
			__antithesis_instrumentation__.Notify(21575)
			return true, nil
		case peekErr != nil:
			__antithesis_instrumentation__.Notify(21576)
			return false, errors.Wrap(peekErr, "peeking message")
		default:
			__antithesis_instrumentation__.Notify(21577)
		}
		__antithesis_instrumentation__.Notify(21567)

		p.mu.lastMessageType = typ
		p.mu.lastMessageTransferredAt = p.logicalClockFn()
		return false, nil
	}
	__antithesis_instrumentation__.Notify(21557)

	if err := enterResume(); err != nil {
		__antithesis_instrumentation__.Notify(21581)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21582)
	}
	__antithesis_instrumentation__.Notify(21558)
	defer exitResume()

	for ctx.Err() == nil {
		__antithesis_instrumentation__.Notify(21583)
		if terminate, err := prepareNextMessage(); err != nil || func() bool {
			__antithesis_instrumentation__.Notify(21586)
			return terminate == true
		}() == true {
			__antithesis_instrumentation__.Notify(21587)
			return err
		} else {
			__antithesis_instrumentation__.Notify(21588)
		}
		__antithesis_instrumentation__.Notify(21584)
		if p.testingKnobs.beforeForwardMsg != nil {
			__antithesis_instrumentation__.Notify(21589)
			p.testingKnobs.beforeForwardMsg()
		} else {
			__antithesis_instrumentation__.Notify(21590)
		}
		__antithesis_instrumentation__.Notify(21585)
		if _, err := p.src.ForwardMsg(p.dst); err != nil {
			__antithesis_instrumentation__.Notify(21591)
			return errors.Wrap(err, "forwarding message")
		} else {
			__antithesis_instrumentation__.Notify(21592)
		}
	}
	__antithesis_instrumentation__.Notify(21559)
	return ctx.Err()
}

func (p *processor) waitResumed(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(21593)
	p.mu.Lock()
	defer p.mu.Unlock()

	for !p.mu.resumed {
		__antithesis_instrumentation__.Notify(21595)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(21597)
			return ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(21598)
		}
		__antithesis_instrumentation__.Notify(21596)
		p.mu.cond.Wait()
	}
	__antithesis_instrumentation__.Notify(21594)
	return nil
}

func (p *processor) suspend(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(21599)
	p.mu.Lock()
	defer p.mu.Unlock()

	defer func() {
		__antithesis_instrumentation__.Notify(21602)
		if p.mu.suspendReq {
			__antithesis_instrumentation__.Notify(21603)
			p.mu.suspendReq = false
			_ = p.src.SetReadDeadline(time.Time{})
		} else {
			__antithesis_instrumentation__.Notify(21604)
		}
	}()
	__antithesis_instrumentation__.Notify(21600)

	for p.mu.resumed {
		__antithesis_instrumentation__.Notify(21605)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(21608)
			return ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(21609)
		}
		__antithesis_instrumentation__.Notify(21606)
		p.mu.suspendReq = true
		if p.mu.inPeek {
			__antithesis_instrumentation__.Notify(21610)
			if err := p.src.SetReadDeadline(aLongTimeAgo); err != nil {
				__antithesis_instrumentation__.Notify(21611)
				return err
			} else {
				__antithesis_instrumentation__.Notify(21612)
			}
		} else {
			__antithesis_instrumentation__.Notify(21613)
		}
		__antithesis_instrumentation__.Notify(21607)
		p.mu.cond.Wait()
	}
	__antithesis_instrumentation__.Notify(21601)
	return nil
}
