package spanlatch

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync/atomic"
	"unsafe"
)

const (
	noSig int32 = iota

	sig

	sigClosed
)

type signal struct {
	a int32
	c unsafe.Pointer
}

func (s *signal) signal() {
	__antithesis_instrumentation__.Notify(122884)
	s.signalWithChoice(false)
}

func (s *signal) signalWithChoice(idempotent bool) {
	__antithesis_instrumentation__.Notify(122885)
	if !atomic.CompareAndSwapInt32(&s.a, noSig, sig) {
		__antithesis_instrumentation__.Notify(122887)
		if idempotent {
			__antithesis_instrumentation__.Notify(122889)
			return
		} else {
			__antithesis_instrumentation__.Notify(122890)
		}
		__antithesis_instrumentation__.Notify(122888)
		panic("signaled twice")
	} else {
		__antithesis_instrumentation__.Notify(122891)
	}
	__antithesis_instrumentation__.Notify(122886)

	if cPtr := atomic.LoadPointer(&s.c); cPtr != nil {
		__antithesis_instrumentation__.Notify(122892)

		if atomic.CompareAndSwapInt32(&s.a, sig, sigClosed) {
			__antithesis_instrumentation__.Notify(122893)
			close(ptrToChan(cPtr))
		} else {
			__antithesis_instrumentation__.Notify(122894)
		}
	} else {
		__antithesis_instrumentation__.Notify(122895)
	}
}

func (s *signal) signaled() bool {
	__antithesis_instrumentation__.Notify(122896)
	return atomic.LoadInt32(&s.a) > noSig
}

func (s *signal) signalChan() <-chan struct{} {
	__antithesis_instrumentation__.Notify(122897)

	if s.signaled() {
		__antithesis_instrumentation__.Notify(122902)
		return closedC
	} else {
		__antithesis_instrumentation__.Notify(122903)
	}
	__antithesis_instrumentation__.Notify(122898)

	if cPtr := atomic.LoadPointer(&s.c); cPtr != nil {
		__antithesis_instrumentation__.Notify(122904)
		return ptrToChan(cPtr)
	} else {
		__antithesis_instrumentation__.Notify(122905)
	}
	__antithesis_instrumentation__.Notify(122899)

	c := make(chan struct{})
	if !atomic.CompareAndSwapPointer(&s.c, nil, chanToPtr(c)) {
		__antithesis_instrumentation__.Notify(122906)

		return ptrToChan(atomic.LoadPointer(&s.c))
	} else {
		__antithesis_instrumentation__.Notify(122907)
	}
	__antithesis_instrumentation__.Notify(122900)

	if atomic.CompareAndSwapInt32(&s.a, sig, sigClosed) {
		__antithesis_instrumentation__.Notify(122908)
		close(c)
	} else {
		__antithesis_instrumentation__.Notify(122909)
	}
	__antithesis_instrumentation__.Notify(122901)
	return c
}

type idempotentSignal struct {
	sig signal
}

func (s *idempotentSignal) signal() {
	__antithesis_instrumentation__.Notify(122910)
	s.sig.signalWithChoice(true)
}

func (s *idempotentSignal) signalChan() <-chan struct{} {
	__antithesis_instrumentation__.Notify(122911)
	return s.sig.signalChan()
}

func chanToPtr(c chan struct{}) unsafe.Pointer {
	__antithesis_instrumentation__.Notify(122912)
	return *(*unsafe.Pointer)(unsafe.Pointer(&c))
}

func ptrToChan(p unsafe.Pointer) chan struct{} {
	__antithesis_instrumentation__.Notify(122913)
	return *(*chan struct{})(unsafe.Pointer(&p))
}

var closedC chan struct{}

func init() {
	closedC = make(chan struct{})
	close(closedC)
}
