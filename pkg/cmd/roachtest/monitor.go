package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/roachprod"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

type monitorImpl struct {
	t interface {
		Fatal(...interface{})
		Failed() bool
		WorkerStatus(...interface{})
	}
	l         *logger.Logger
	nodes     string
	ctx       context.Context
	cancel    func()
	g         *errgroup.Group
	expDeaths int32
}

func newMonitor(
	ctx context.Context,
	t interface {
		Fatal(...interface{})
		Failed() bool
		WorkerStatus(...interface{})
		L() *logger.Logger
	},
	c cluster.Cluster,
	opts ...option.Option,
) *monitorImpl {
	__antithesis_instrumentation__.Notify(44216)
	m := &monitorImpl{
		t:     t,
		l:     t.L(),
		nodes: c.MakeNodes(opts...),
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.g, m.ctx = errgroup.WithContext(m.ctx)
	return m
}

func (m *monitorImpl) ExpectDeath() {
	__antithesis_instrumentation__.Notify(44217)
	m.ExpectDeaths(1)
}

func (m *monitorImpl) ExpectDeaths(count int32) {
	__antithesis_instrumentation__.Notify(44218)
	atomic.AddInt32(&m.expDeaths, count)
}

func (m *monitorImpl) ResetDeaths() {
	__antithesis_instrumentation__.Notify(44219)
	atomic.StoreInt32(&m.expDeaths, 0)
}

var errTestFatal = errors.New("t.Fatal() was called")

func (m *monitorImpl) Go(fn func(context.Context) error) {
	__antithesis_instrumentation__.Notify(44220)
	m.g.Go(func() (err error) {
		__antithesis_instrumentation__.Notify(44221)
		defer func() {
			__antithesis_instrumentation__.Notify(44223)
			r := recover()
			if r == nil {
				__antithesis_instrumentation__.Notify(44226)
				return
			} else {
				__antithesis_instrumentation__.Notify(44227)
			}
			__antithesis_instrumentation__.Notify(44224)
			rErr, ok := r.(error)
			if !ok {
				__antithesis_instrumentation__.Notify(44228)
				rErr = errors.Errorf("recovered panic: %v", r)
			} else {
				__antithesis_instrumentation__.Notify(44229)
			}
			__antithesis_instrumentation__.Notify(44225)

			err = rErr
		}()
		__antithesis_instrumentation__.Notify(44222)

		defer m.t.WorkerStatus()
		return fn(m.ctx)
	})
}

func (m *monitorImpl) WaitE() error {
	__antithesis_instrumentation__.Notify(44230)
	if m.t.Failed() {
		__antithesis_instrumentation__.Notify(44232)

		return errors.New("already failed")
	} else {
		__antithesis_instrumentation__.Notify(44233)
	}
	__antithesis_instrumentation__.Notify(44231)

	return errors.Wrap(m.wait(), "monitor failure")
}

func (m *monitorImpl) Wait() {
	__antithesis_instrumentation__.Notify(44234)
	if m.t.Failed() {
		__antithesis_instrumentation__.Notify(44236)

		return
	} else {
		__antithesis_instrumentation__.Notify(44237)
	}
	__antithesis_instrumentation__.Notify(44235)
	if err := m.WaitE(); err != nil {
		__antithesis_instrumentation__.Notify(44238)

		m.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(44239)
	}
}

func (m *monitorImpl) wait() error {
	__antithesis_instrumentation__.Notify(44240)

	var errOnce sync.Once
	var err error
	setErr := func(e error) {
		__antithesis_instrumentation__.Notify(44244)
		if e != nil {
			__antithesis_instrumentation__.Notify(44245)
			errOnce.Do(func() {
				__antithesis_instrumentation__.Notify(44246)
				err = e
			})
		} else {
			__antithesis_instrumentation__.Notify(44247)
		}
	}
	__antithesis_instrumentation__.Notify(44241)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		__antithesis_instrumentation__.Notify(44248)
		defer func() {
			__antithesis_instrumentation__.Notify(44250)
			m.cancel()
			wg.Done()
		}()
		__antithesis_instrumentation__.Notify(44249)
		setErr(errors.Wrap(m.g.Wait(), "monitor task failed"))
	}()
	__antithesis_instrumentation__.Notify(44242)

	wg.Add(1)
	go func() {
		__antithesis_instrumentation__.Notify(44251)
		defer func() {
			__antithesis_instrumentation__.Notify(44254)
			m.cancel()
			wg.Done()
		}()
		__antithesis_instrumentation__.Notify(44252)

		messagesChannel, err := roachprod.Monitor(m.ctx, m.l, m.nodes, install.MonitorOpts{})
		if err != nil {
			__antithesis_instrumentation__.Notify(44255)
			setErr(errors.Wrap(err, "monitor command failure"))
			return
		} else {
			__antithesis_instrumentation__.Notify(44256)
		}
		__antithesis_instrumentation__.Notify(44253)
		var monitorErr error
		for msg := range messagesChannel {
			__antithesis_instrumentation__.Notify(44257)
			if msg.Err != nil {
				__antithesis_instrumentation__.Notify(44260)
				msg.Msg += "error: " + msg.Err.Error()
			} else {
				__antithesis_instrumentation__.Notify(44261)
			}
			__antithesis_instrumentation__.Notify(44258)
			thisError := errors.Newf("%d: %s", msg.Node, msg.Msg)
			if msg.Err != nil || func() bool {
				__antithesis_instrumentation__.Notify(44262)
				return strings.Contains(msg.Msg, "dead") == true
			}() == true {
				__antithesis_instrumentation__.Notify(44263)
				monitorErr = errors.CombineErrors(monitorErr, thisError)
			} else {
				__antithesis_instrumentation__.Notify(44264)
			}
			__antithesis_instrumentation__.Notify(44259)
			var id int
			var s string
			newMsg := thisError.Error()
			if n, _ := fmt.Sscanf(newMsg, "%d: %s", &id, &s); n == 2 {
				__antithesis_instrumentation__.Notify(44265)
				if strings.Contains(s, "dead") && func() bool {
					__antithesis_instrumentation__.Notify(44266)
					return atomic.AddInt32(&m.expDeaths, -1) < 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(44267)
					setErr(errors.Wrap(fmt.Errorf("unexpected node event: %s", newMsg), "monitor command failure"))
					return
				} else {
					__antithesis_instrumentation__.Notify(44268)
				}
			} else {
				__antithesis_instrumentation__.Notify(44269)
			}
		}
	}()
	__antithesis_instrumentation__.Notify(44243)

	wg.Wait()
	return err
}
