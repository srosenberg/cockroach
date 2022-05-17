package balancer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/logtags"
)

type ConnTracker struct {
	mu struct {
		syncutil.Mutex
		tenants map[roachpb.TenantID]*tenantEntry
	}
}

func NewConnTracker() *ConnTracker {
	__antithesis_instrumentation__.Notify(21070)
	t := &ConnTracker{}
	t.mu.tenants = make(map[roachpb.TenantID]*tenantEntry)
	return t
}

func (t *ConnTracker) OnConnect(tenantID roachpb.TenantID, handle ConnectionHandle) bool {
	__antithesis_instrumentation__.Notify(21071)
	e := t.ensureTenantEntry(tenantID)
	success := e.addHandle(handle)
	if success {
		__antithesis_instrumentation__.Notify(21073)
		logTrackerEvent("OnConnect", handle)
	} else {
		__antithesis_instrumentation__.Notify(21074)
	}
	__antithesis_instrumentation__.Notify(21072)
	return success
}

func (t *ConnTracker) OnDisconnect(tenantID roachpb.TenantID, handle ConnectionHandle) bool {
	__antithesis_instrumentation__.Notify(21075)
	e := t.ensureTenantEntry(tenantID)
	success := e.removeHandle(handle)
	if success {
		__antithesis_instrumentation__.Notify(21077)
		logTrackerEvent("OnDisconnect", handle)
	} else {
		__antithesis_instrumentation__.Notify(21078)
	}
	__antithesis_instrumentation__.Notify(21076)
	return success
}

func (t *ConnTracker) GetConns(tenantID roachpb.TenantID) []ConnectionHandle {
	__antithesis_instrumentation__.Notify(21079)
	e := t.ensureTenantEntry(tenantID)
	return e.getConns()
}

func (t *ConnTracker) GetAllConns() map[roachpb.TenantID][]ConnectionHandle {
	__antithesis_instrumentation__.Notify(21080)

	snapshotTenantEntries := func() map[roachpb.TenantID]*tenantEntry {
		__antithesis_instrumentation__.Notify(21083)
		t.mu.Lock()
		defer t.mu.Unlock()

		m := make(map[roachpb.TenantID]*tenantEntry)
		for tenantID, entry := range t.mu.tenants {
			__antithesis_instrumentation__.Notify(21085)
			m[tenantID] = entry
		}
		__antithesis_instrumentation__.Notify(21084)
		return m
	}
	__antithesis_instrumentation__.Notify(21081)

	m := make(map[roachpb.TenantID][]ConnectionHandle)
	tenantsCopy := snapshotTenantEntries()
	for tenantID, entry := range tenantsCopy {
		__antithesis_instrumentation__.Notify(21086)
		conns := entry.getConns()
		if len(conns) == 0 {
			__antithesis_instrumentation__.Notify(21088)
			continue
		} else {
			__antithesis_instrumentation__.Notify(21089)
		}
		__antithesis_instrumentation__.Notify(21087)
		m[tenantID] = conns
	}
	__antithesis_instrumentation__.Notify(21082)
	return m
}

func (t *ConnTracker) ensureTenantEntry(tenantID roachpb.TenantID) *tenantEntry {
	__antithesis_instrumentation__.Notify(21090)
	t.mu.Lock()
	defer t.mu.Unlock()

	entry, ok := t.mu.tenants[tenantID]
	if !ok {
		__antithesis_instrumentation__.Notify(21092)
		entry = newTenantEntry()
		t.mu.tenants[tenantID] = entry
	} else {
		__antithesis_instrumentation__.Notify(21093)
	}
	__antithesis_instrumentation__.Notify(21091)
	return entry
}

func logTrackerEvent(event string, handle ConnectionHandle) {
	__antithesis_instrumentation__.Notify(21094)

	logCtx := logtags.WithTags(context.Background(), logtags.FromContext(handle.Context()))
	log.Infof(logCtx, "%s: %s", event, handle.ServerRemoteAddr())
}

type tenantEntry struct {
	mu struct {
		syncutil.Mutex
		conns map[ConnectionHandle]struct{}
	}
}

func newTenantEntry() *tenantEntry {
	__antithesis_instrumentation__.Notify(21095)
	e := &tenantEntry{}
	e.mu.conns = make(map[ConnectionHandle]struct{})
	return e
}

func (e *tenantEntry) addHandle(handle ConnectionHandle) bool {
	__antithesis_instrumentation__.Notify(21096)
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.mu.conns[handle]; ok {
		__antithesis_instrumentation__.Notify(21098)
		return false
	} else {
		__antithesis_instrumentation__.Notify(21099)
	}
	__antithesis_instrumentation__.Notify(21097)
	e.mu.conns[handle] = struct{}{}
	return true
}

func (e *tenantEntry) removeHandle(handle ConnectionHandle) bool {
	__antithesis_instrumentation__.Notify(21100)
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.mu.conns[handle]; !ok {
		__antithesis_instrumentation__.Notify(21102)
		return false
	} else {
		__antithesis_instrumentation__.Notify(21103)
	}
	__antithesis_instrumentation__.Notify(21101)
	delete(e.mu.conns, handle)
	return true
}

func (e *tenantEntry) getConns() []ConnectionHandle {
	__antithesis_instrumentation__.Notify(21104)
	e.mu.Lock()
	defer e.mu.Unlock()

	conns := make([]ConnectionHandle, 0, len(e.mu.conns))
	for handle := range e.mu.conns {
		__antithesis_instrumentation__.Notify(21106)
		conns = append(conns, handle)
	}
	__antithesis_instrumentation__.Notify(21105)
	return conns
}
