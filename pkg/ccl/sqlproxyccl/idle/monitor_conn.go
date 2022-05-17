package idle

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net"
	"sync/atomic"
)

type monitorConn struct {
	net.Conn
	monitor  *Monitor
	deadline int64
}

func newMonitorConn(monitor *Monitor, conn net.Conn) *monitorConn {
	__antithesis_instrumentation__.Notify(21676)
	return &monitorConn{Conn: conn, monitor: monitor}
}

func (c *monitorConn) Read(b []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(21677)
	c.clearDeadline()
	return c.Conn.Read(b)
}

func (c *monitorConn) Write(b []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(21678)
	c.clearDeadline()
	return c.Conn.Write(b)
}

func (c *monitorConn) Close() error {
	__antithesis_instrumentation__.Notify(21679)
	c.monitor.removeConn(c)
	return c.Conn.Close()
}

func (c *monitorConn) clearDeadline() {
	__antithesis_instrumentation__.Notify(21680)
	deadline := atomic.LoadInt64(&c.deadline)
	if deadline == 0 {
		__antithesis_instrumentation__.Notify(21682)
		return
	} else {
		__antithesis_instrumentation__.Notify(21683)
	}
	__antithesis_instrumentation__.Notify(21681)
	atomic.StoreInt64(&c.deadline, 0)
}
