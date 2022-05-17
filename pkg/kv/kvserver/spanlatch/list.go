package spanlatch

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type latchList struct {
	root latch
	len  int
}

func (ll *latchList) front() *latch {
	__antithesis_instrumentation__.Notify(122721)
	if ll.len == 0 {
		__antithesis_instrumentation__.Notify(122723)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(122724)
	}
	__antithesis_instrumentation__.Notify(122722)
	return ll.root.next
}

func (ll *latchList) lazyInit() {
	__antithesis_instrumentation__.Notify(122725)
	if ll.root.next == nil {
		__antithesis_instrumentation__.Notify(122726)
		ll.root.next = &ll.root
		ll.root.prev = &ll.root
	} else {
		__antithesis_instrumentation__.Notify(122727)
	}
}

func (ll *latchList) pushBack(la *latch) {
	__antithesis_instrumentation__.Notify(122728)
	ll.lazyInit()
	at := ll.root.prev
	n := at.next
	at.next = la
	la.prev = at
	la.next = n
	n.prev = la
	ll.len++
}

func (ll *latchList) remove(la *latch) {
	__antithesis_instrumentation__.Notify(122729)
	la.prev.next = la.next
	la.next.prev = la.prev
	la.next = nil
	la.prev = nil
	ll.len--
}
