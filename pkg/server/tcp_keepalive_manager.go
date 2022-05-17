package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cmux"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type tcpKeepAliveManager struct {
	tcpKeepAlive time.Duration

	loggedKeepAliveStatus int32
}

func makeTCPKeepAliveManager() tcpKeepAliveManager {
	__antithesis_instrumentation__.Notify(238182)
	return tcpKeepAliveManager{
		tcpKeepAlive: envutil.EnvOrDefaultDuration("COCKROACH_SQL_TCP_KEEP_ALIVE", time.Minute),
	}
}

func (k *tcpKeepAliveManager) configure(ctx context.Context, conn net.Conn) {
	__antithesis_instrumentation__.Notify(238183)
	if k.tcpKeepAlive == 0 {
		__antithesis_instrumentation__.Notify(238189)
		return
	} else {
		__antithesis_instrumentation__.Notify(238190)
	}
	__antithesis_instrumentation__.Notify(238184)

	muxConn, ok := conn.(*cmux.MuxConn)
	if !ok {
		__antithesis_instrumentation__.Notify(238191)
		return
	} else {
		__antithesis_instrumentation__.Notify(238192)
	}
	__antithesis_instrumentation__.Notify(238185)
	tcpConn, ok := muxConn.Conn.(*net.TCPConn)
	if !ok {
		__antithesis_instrumentation__.Notify(238193)
		return
	} else {
		__antithesis_instrumentation__.Notify(238194)
	}
	__antithesis_instrumentation__.Notify(238186)

	doLog := atomic.CompareAndSwapInt32(&k.loggedKeepAliveStatus, 0, 1)
	if err := tcpConn.SetKeepAlive(true); err != nil {
		__antithesis_instrumentation__.Notify(238195)
		if doLog {
			__antithesis_instrumentation__.Notify(238197)
			log.Ops.Warningf(ctx, "failed to enable TCP keep-alive for pgwire: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(238198)
		}
		__antithesis_instrumentation__.Notify(238196)
		return

	} else {
		__antithesis_instrumentation__.Notify(238199)
	}
	__antithesis_instrumentation__.Notify(238187)
	if err := tcpConn.SetKeepAlivePeriod(k.tcpKeepAlive); err != nil {
		__antithesis_instrumentation__.Notify(238200)
		if doLog {
			__antithesis_instrumentation__.Notify(238202)
			log.Ops.Warningf(ctx, "failed to set TCP keep-alive duration for pgwire: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(238203)
		}
		__antithesis_instrumentation__.Notify(238201)
		return
	} else {
		__antithesis_instrumentation__.Notify(238204)
	}
	__antithesis_instrumentation__.Notify(238188)

	if doLog {
		__antithesis_instrumentation__.Notify(238205)
		log.VEventf(ctx, 2, "setting TCP keep-alive to %s for pgwire", k.tcpKeepAlive)
	} else {
		__antithesis_instrumentation__.Notify(238206)
	}
}
