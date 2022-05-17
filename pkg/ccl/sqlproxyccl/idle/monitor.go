package idle

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const maxIdleConns = 1000

type OnIdleFunc func()

type connMap map[*monitorConn]OnIdleFunc

type Monitor struct {
	timeout time.Duration

	mu struct {
		syncutil.Mutex

		conns map[string]connMap

		checks map[string]struct{}
	}
}

func NewMonitor(ctx context.Context, timeout time.Duration) *Monitor {
	__antithesis_instrumentation__.Notify(21640)
	if timeout == 0 {
		__antithesis_instrumentation__.Notify(21642)
		panic("monitor should never be constructed with a zero timeout")
	} else {
		__antithesis_instrumentation__.Notify(21643)
	}
	__antithesis_instrumentation__.Notify(21641)

	m := &Monitor{timeout: timeout}
	m.mu.conns = make(map[string]connMap)
	m.mu.checks = make(map[string]struct{})

	go m.start(ctx)

	return m
}

func (m *Monitor) SetIdleChecks(addr string) {
	__antithesis_instrumentation__.Notify(21644)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mu.checks[addr] = struct{}{}
}

func (m *Monitor) ClearIdleChecks(addr string) {
	__antithesis_instrumentation__.Notify(21645)
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.mu.checks, addr)
}

func (m *Monitor) DetectIdle(conn net.Conn, onIdle OnIdleFunc) net.Conn {
	__antithesis_instrumentation__.Notify(21646)
	m.mu.Lock()
	defer m.mu.Unlock()

	wrapper := newMonitorConn(m, conn)

	addr := conn.RemoteAddr().String()
	existing, ok := m.mu.conns[addr]
	if !ok {
		__antithesis_instrumentation__.Notify(21648)
		existing = make(map[*monitorConn]OnIdleFunc)
		m.mu.conns[addr] = existing
	} else {
		__antithesis_instrumentation__.Notify(21649)
	}
	__antithesis_instrumentation__.Notify(21647)

	existing[wrapper] = onIdle
	return wrapper
}

func (m *Monitor) start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(21650)
	var checkAddrs []string
	var idleFuncs []OnIdleFunc

	for {
		__antithesis_instrumentation__.Notify(21651)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(21653)
			return

		case <-time.After(m.timeout / 10):
			__antithesis_instrumentation__.Notify(21654)
		}
		__antithesis_instrumentation__.Notify(21652)

		checkAddrs = m.getAddrsToCheck(checkAddrs)
		for _, addr := range checkAddrs {
			__antithesis_instrumentation__.Notify(21655)

			idleFuncs = m.findIdleConns(addr, idleFuncs)

			for _, onIdle := range idleFuncs {
				__antithesis_instrumentation__.Notify(21656)
				onIdle()
			}
		}
	}
}

func (m *Monitor) getAddrsToCheck(checkAddrs []string) []string {
	__antithesis_instrumentation__.Notify(21657)
	checkAddrs = checkAddrs[:0]

	m.mu.Lock()
	defer m.mu.Unlock()
	for addr := range m.mu.checks {
		__antithesis_instrumentation__.Notify(21659)
		checkAddrs = append(checkAddrs, addr)
	}
	__antithesis_instrumentation__.Notify(21658)
	return checkAddrs
}

func (m *Monitor) findIdleConns(addr string, idleFuncs []OnIdleFunc) []OnIdleFunc {
	__antithesis_instrumentation__.Notify(21660)
	idleFuncs = idleFuncs[:0]
	now := timeutil.Now().UnixNano()

	deadline := now + int64(m.timeout)

	m.mu.Lock()
	defer m.mu.Unlock()

	connMap := m.mu.conns[addr]
	for conn, onIdle := range connMap {
		__antithesis_instrumentation__.Notify(21662)

		connDeadline := atomic.LoadInt64(&conn.deadline)
		if connDeadline != 0 {
			__antithesis_instrumentation__.Notify(21663)

			if now > connDeadline {
				__antithesis_instrumentation__.Notify(21664)
				idleFuncs = append(idleFuncs, onIdle)
				if len(idleFuncs) >= maxIdleConns {
					__antithesis_instrumentation__.Notify(21665)

					break
				} else {
					__antithesis_instrumentation__.Notify(21666)
				}
			} else {
				__antithesis_instrumentation__.Notify(21667)
			}
		} else {
			__antithesis_instrumentation__.Notify(21668)

			atomic.StoreInt64(&conn.deadline, deadline)
		}
	}
	__antithesis_instrumentation__.Notify(21661)

	return idleFuncs
}

func (m *Monitor) removeConn(conn *monitorConn) {
	__antithesis_instrumentation__.Notify(21669)
	m.mu.Lock()
	defer m.mu.Unlock()

	addr := conn.RemoteAddr().String()
	if connMap, ok := m.mu.conns[addr]; ok {
		__antithesis_instrumentation__.Notify(21670)
		delete(connMap, conn)
		if len(connMap) == 0 {
			__antithesis_instrumentation__.Notify(21671)

			delete(m.mu.conns, addr)
		} else {
			__antithesis_instrumentation__.Notify(21672)
		}
	} else {
		__antithesis_instrumentation__.Notify(21673)
	}
}

func (m *Monitor) countConnsToAddr(addr string) int {
	__antithesis_instrumentation__.Notify(21674)
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.mu.conns[addr])
}

func (m *Monitor) countAddrsToCheck() int {
	__antithesis_instrumentation__.Notify(21675)
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.mu.checks)
}
