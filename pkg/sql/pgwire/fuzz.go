//go:build gofuzz
// +build gofuzz

package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

func FuzzServeConn(data []byte) int {
	__antithesis_instrumentation__.Notify(559825)
	s := MakeServer(
		log.AmbientContext{},
		&base.Config{},
		&cluster.Settings{},
		sql.MemoryMetrics{},
		&mon.BytesMonitor{},
		time.Minute,
		&sql.ExecutorConfig{
			Settings: &cluster.Settings{},
		},
	)

	srv, client := net.Pipe()
	go func() {
		__antithesis_instrumentation__.Notify(559829)

		_, _ = client.Write(data)
		_ = client.Close()
	}()
	__antithesis_instrumentation__.Notify(559826)
	go func() {
		__antithesis_instrumentation__.Notify(559830)

		_, _ = io.Copy(ioutil.Discard, client)
	}()
	__antithesis_instrumentation__.Notify(559827)
	err := s.ServeConn(context.Background(), srv)
	if err != nil {
		__antithesis_instrumentation__.Notify(559831)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(559832)
	}
	__antithesis_instrumentation__.Notify(559828)
	return 1
}
